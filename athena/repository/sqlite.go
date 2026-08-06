package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	database *sql.DB
}

func OpenSQLite(path string) (*SQLiteStore, error) {
	database, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open Athena database: %w", err)
	}
	database.SetMaxOpenConns(1)
	store := &SQLiteStore{database: database}
	if err := store.initialize(context.Background()); err != nil {
		database.Close()
		return nil, err
	}
	return store, nil
}

func (s *SQLiteStore) Close() error {
	return s.database.Close()
}

func (s *SQLiteStore) Latest(ctx context.Context, repositoryID string) (Snapshot, []File, bool, error) {
	row := s.database.QueryRowContext(ctx, `
		SELECT s.id, s.repository_id, s.revision, s.fingerprint, s.created_at
		FROM repositories r
		JOIN snapshots s ON s.id = r.latest_snapshot_id
		WHERE r.id = ?`, repositoryID)
	var snapshot Snapshot
	var createdAt string
	if err := row.Scan(&snapshot.ID, &snapshot.RepositoryID, &snapshot.Revision, &snapshot.Fingerprint, &createdAt); err != nil {
		if err == sql.ErrNoRows {
			return Snapshot{}, nil, false, nil
		}
		return Snapshot{}, nil, false, fmt.Errorf("load latest snapshot: %w", err)
	}
	parsed, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return Snapshot{}, nil, false, fmt.Errorf("parse snapshot time: %w", err)
	}
	snapshot.CreatedAt = parsed
	rows, err := s.database.QueryContext(ctx, `
		SELECT path, sha256, size, language
		FROM files
		WHERE snapshot_id = ?
		ORDER BY path`, snapshot.ID)
	if err != nil {
		return Snapshot{}, nil, false, fmt.Errorf("load snapshot files: %w", err)
	}
	defer rows.Close()
	files := []File{}
	for rows.Next() {
		var file File
		if err := rows.Scan(&file.Path, &file.SHA256, &file.Size, &file.Language); err != nil {
			return Snapshot{}, nil, false, fmt.Errorf("scan snapshot file: %w", err)
		}
		files = append(files, file)
	}
	if err := rows.Err(); err != nil {
		return Snapshot{}, nil, false, fmt.Errorf("iterate snapshot files: %w", err)
	}
	return snapshot, files, true, nil
}

func (s *SQLiteStore) Save(ctx context.Context, repositoryID string, root string, snapshot Snapshot, files []File, facts []ParseFact) error {
	transaction, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin snapshot transaction: %w", err)
	}
	defer transaction.Rollback()
	createdAt := snapshot.CreatedAt.Format(time.RFC3339Nano)
	if _, err := transaction.ExecContext(ctx, `
		INSERT INTO repositories (id, root_path, latest_snapshot_id, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			root_path = excluded.root_path,
			latest_snapshot_id = excluded.latest_snapshot_id,
			updated_at = excluded.updated_at`, repositoryID, root, snapshot.ID, createdAt); err != nil {
		return fmt.Errorf("save repository: %w", err)
	}
	if _, err := transaction.ExecContext(ctx, `
		INSERT OR IGNORE INTO snapshots (id, repository_id, revision, fingerprint, created_at)
		VALUES (?, ?, ?, ?, ?)`, snapshot.ID, snapshot.RepositoryID, snapshot.Revision, snapshot.Fingerprint, createdAt); err != nil {
		return fmt.Errorf("save snapshot: %w", err)
	}
	statement, err := transaction.PrepareContext(ctx, `
		INSERT OR IGNORE INTO files (snapshot_id, path, sha256, size, language)
		VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare file insert: %w", err)
	}
	defer statement.Close()
	for _, file := range files {
		if _, err := statement.ExecContext(ctx, snapshot.ID, file.Path, file.SHA256, file.Size, file.Language); err != nil {
			return fmt.Errorf("save file %s: %w", file.Path, err)
		}
	}
	if err := saveParseFacts(ctx, transaction, repositoryID, facts); err != nil {
		return err
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit snapshot transaction: %w", err)
	}
	return nil
}

// ParsedSet returns the set of file contents that already have durable parse
// facts for the repository, keyed by path and source hash.
func (s *SQLiteStore) ParsedSet(ctx context.Context, repositoryID string) (map[ParseKey]struct{}, error) {
	rows, err := s.database.QueryContext(ctx, `
		SELECT path, sha256
		FROM parses
		WHERE repository_id = ?`, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("load parse cache: %w", err)
	}
	defer rows.Close()
	parsed := make(map[ParseKey]struct{})
	for rows.Next() {
		var key ParseKey
		if err := rows.Scan(&key.Path, &key.SHA256); err != nil {
			return nil, fmt.Errorf("scan parse cache: %w", err)
		}
		parsed[key] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate parse cache: %w", err)
	}
	return parsed, nil
}

// SaveParse persists parse facts that were produced outside a snapshot write,
// for example when an unchanged snapshot is re-indexed after an upgrade.
func (s *SQLiteStore) SaveParse(ctx context.Context, repositoryID string, facts []ParseFact) error {
	transaction, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin parse transaction: %w", err)
	}
	defer transaction.Rollback()
	if err := saveParseFacts(ctx, transaction, repositoryID, facts); err != nil {
		return err
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit parse transaction: %w", err)
	}
	return nil
}

func saveParseFacts(ctx context.Context, transaction *sql.Tx, repositoryID string, facts []ParseFact) error {
	parses, err := transaction.PrepareContext(ctx, `
		INSERT OR REPLACE INTO parses (repository_id, path, sha256, language, has_errors, parsed_at)
		VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare parse insert: %w", err)
	}
	defer parses.Close()
	symbols, err := transaction.PrepareContext(ctx, `
		INSERT OR REPLACE INTO symbols (repository_id, path, sha256, name, kind, start_byte, end_byte, start_line, end_line)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare symbol insert: %w", err)
	}
	defer symbols.Close()
	imports, err := transaction.PrepareContext(ctx, `
		INSERT OR REPLACE INTO imports (repository_id, path, sha256, kind, spec, start_byte, end_byte, start_line, end_line)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare import insert: %w", err)
	}
	defer imports.Close()
	parsedAt := time.Now().UTC().Format(time.RFC3339Nano)
	for _, fact := range facts {
		if fact.Skipped {
			continue
		}
		if _, err := parses.ExecContext(ctx, repositoryID, fact.Path, fact.SHA256, fact.Language, boolInt(fact.HasErrors), parsedAt); err != nil {
			return fmt.Errorf("save parse %s: %w", fact.Path, err)
		}
		for _, symbol := range fact.Symbols {
			if _, err := symbols.ExecContext(ctx, repositoryID, fact.Path, fact.SHA256, symbol.Name, string(symbol.Kind),
				symbol.StartByte, symbol.EndByte, symbol.StartLine, symbol.EndLine); err != nil {
				return fmt.Errorf("save symbol %s in %s: %w", symbol.Name, fact.Path, err)
			}
		}
		for _, imp := range fact.Imports {
			if _, err := imports.ExecContext(ctx, repositoryID, fact.Path, fact.SHA256, string(imp.Kind), imp.Spec,
				imp.StartByte, imp.EndByte, imp.StartLine, imp.EndLine); err != nil {
				return fmt.Errorf("save import %s in %s: %w", imp.Spec, fact.Path, err)
			}
		}
	}
	return nil
}

// Symbols returns the symbols of a snapshot, ordered by path and byte offset.
// The join against the snapshot's files scopes symbols to the content that the
// snapshot actually captured.
func (s *SQLiteStore) Symbols(ctx context.Context, snapshotID string, filter SymbolQuery) ([]Symbol, error) {
	query := `
		SELECT f.path, f.sha256, sym.name, sym.kind, sym.start_byte, sym.end_byte, sym.start_line, sym.end_line
		FROM snapshots snap
		JOIN files f ON f.snapshot_id = snap.id
		JOIN symbols sym ON sym.repository_id = snap.repository_id AND sym.path = f.path AND sym.sha256 = f.sha256
		WHERE snap.id = ?`
	args := []any{snapshotID}
	if filter.Path != "" {
		query += " AND f.path = ?"
		args = append(args, filter.Path)
	}
	if filter.Name != "" {
		query += " AND sym.name = ?"
		args = append(args, filter.Name)
	}
	if filter.Kind != "" {
		query += " AND sym.kind = ?"
		args = append(args, filter.Kind)
	}
	query += " ORDER BY f.path, sym.start_byte, sym.name"
	rows, err := s.database.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query symbols: %w", err)
	}
	defer rows.Close()
	symbols := []Symbol{}
	for rows.Next() {
		var symbol Symbol
		if err := rows.Scan(&symbol.Path, &symbol.SHA256, &symbol.Name, &symbol.Kind, &symbol.StartByte, &symbol.EndByte, &symbol.StartLine, &symbol.EndLine); err != nil {
			return nil, fmt.Errorf("scan symbol: %w", err)
		}
		symbols = append(symbols, symbol)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate symbols: %w", err)
	}
	return symbols, nil
}

// Imports returns the module references of a snapshot, ordered by path and byte
// offset. The join against the snapshot's files scopes imports to the content
// that the snapshot actually captured.
func (s *SQLiteStore) Imports(ctx context.Context, snapshotID string, filter ImportQuery) ([]Import, error) {
	query := `
		SELECT f.path, f.sha256, imp.kind, imp.spec, imp.start_byte, imp.end_byte, imp.start_line, imp.end_line
		FROM snapshots snap
		JOIN files f ON f.snapshot_id = snap.id
		JOIN imports imp ON imp.repository_id = snap.repository_id AND imp.path = f.path AND imp.sha256 = f.sha256
		WHERE snap.id = ?`
	args := []any{snapshotID}
	if filter.Path != "" {
		query += " AND f.path = ?"
		args = append(args, filter.Path)
	}
	if filter.Kind != "" {
		query += " AND imp.kind = ?"
		args = append(args, filter.Kind)
	}
	if filter.Spec != "" {
		query += " AND imp.spec = ?"
		args = append(args, filter.Spec)
	}
	query += " ORDER BY f.path, imp.start_byte, imp.spec"
	rows, err := s.database.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query imports: %w", err)
	}
	defer rows.Close()
	imports := []Import{}
	for rows.Next() {
		var imp Import
		if err := rows.Scan(&imp.Path, &imp.SHA256, &imp.Kind, &imp.Spec, &imp.StartByte, &imp.EndByte, &imp.StartLine, &imp.EndLine); err != nil {
			return nil, fmt.Errorf("scan import: %w", err)
		}
		imports = append(imports, imp)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate imports: %w", err)
	}
	return imports, nil
}

// SaveEdges persists resolved cross-file import edges for a snapshot, replacing
// any previously derived edge set for that snapshot.
func (s *SQLiteStore) SaveEdges(ctx context.Context, repositoryID string, snapshotID string, edges []Edge) error {
	transaction, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin edge transaction: %w", err)
	}
	defer transaction.Rollback()
	if _, err := transaction.ExecContext(ctx, `DELETE FROM import_edges WHERE snapshot_id = ?`, snapshotID); err != nil {
		return fmt.Errorf("clear edges: %w", err)
	}
	statement, err := transaction.PrepareContext(ctx, `
		INSERT OR REPLACE INTO import_edges (snapshot_id, source_path, source_sha256, kind, spec, start_byte, end_byte, target_path, target_sha256)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare edge insert: %w", err)
	}
	defer statement.Close()
	for _, edge := range edges {
		if _, err := statement.ExecContext(ctx, snapshotID, edge.SourcePath, edge.SourceSHA, string(edge.Kind), edge.Spec,
			edge.StartByte, edge.EndByte, edge.TargetPath, edge.TargetSHA); err != nil {
			return fmt.Errorf("save edge %s -> %s in %s: %w", edge.SourcePath, edge.TargetPath, repositoryID, err)
		}
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit edge transaction: %w", err)
	}
	return nil
}

// Edges returns the resolved import edges of a snapshot, ordered by source
// path and byte offset.
func (s *SQLiteStore) Edges(ctx context.Context, snapshotID string, filter EdgeQuery) ([]Edge, error) {
	query := `
		SELECT source_path, source_sha256, kind, spec, start_byte, end_byte, target_path, target_sha256
		FROM import_edges
		WHERE snapshot_id = ?`
	args := []any{snapshotID}
	if filter.Path != "" {
		query += " AND source_path = ?"
		args = append(args, filter.Path)
	}
	if filter.Kind != "" {
		query += " AND kind = ?"
		args = append(args, filter.Kind)
	}
	if filter.To != "" {
		query += " AND target_path = ?"
		args = append(args, filter.To)
	}
	query += " ORDER BY source_path, start_byte, target_path"
	rows, err := s.database.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query edges: %w", err)
	}
	defer rows.Close()
	edges := []Edge{}
	for rows.Next() {
		var edge Edge
		if err := rows.Scan(&edge.SourcePath, &edge.SourceSHA, &edge.Kind, &edge.Spec, &edge.StartByte, &edge.EndByte, &edge.TargetPath, &edge.TargetSHA); err != nil {
			return nil, fmt.Errorf("scan edge: %w", err)
		}
		edges = append(edges, edge)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate edges: %w", err)
	}
	return edges, nil
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func (s *SQLiteStore) initialize(ctx context.Context) error {
	if _, err := s.database.ExecContext(ctx, `PRAGMA foreign_keys = ON; PRAGMA journal_mode = WAL; PRAGMA busy_timeout = 5000;`); err != nil {
		return fmt.Errorf("configure Athena database: %w", err)
	}
	if _, err := s.database.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS repositories (
			id TEXT PRIMARY KEY,
			root_path TEXT NOT NULL UNIQUE,
			latest_snapshot_id TEXT,
			updated_at TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS snapshots (
			id TEXT PRIMARY KEY,
			repository_id TEXT NOT NULL REFERENCES repositories(id),
			revision TEXT NOT NULL,
			fingerprint TEXT NOT NULL,
			created_at TEXT NOT NULL,
			UNIQUE(repository_id, fingerprint)
		);
		CREATE TABLE IF NOT EXISTS files (
			snapshot_id TEXT NOT NULL REFERENCES snapshots(id),
			path TEXT NOT NULL,
			sha256 TEXT NOT NULL,
			size INTEGER NOT NULL,
			language TEXT NOT NULL,
			PRIMARY KEY(snapshot_id, path)
		);
		CREATE TABLE IF NOT EXISTS parses (
			repository_id TEXT NOT NULL REFERENCES repositories(id),
			path TEXT NOT NULL,
			sha256 TEXT NOT NULL,
			language TEXT NOT NULL,
			has_errors INTEGER NOT NULL,
			parsed_at TEXT NOT NULL,
			PRIMARY KEY(repository_id, path, sha256)
		);
		CREATE TABLE IF NOT EXISTS symbols (
			repository_id TEXT NOT NULL REFERENCES repositories(id),
			path TEXT NOT NULL,
			sha256 TEXT NOT NULL,
			name TEXT NOT NULL,
			kind TEXT NOT NULL,
			start_byte INTEGER NOT NULL,
			end_byte INTEGER NOT NULL,
			start_line INTEGER NOT NULL,
			end_line INTEGER NOT NULL,
			PRIMARY KEY(repository_id, path, sha256, start_byte, name)
		);
		CREATE TABLE IF NOT EXISTS imports (
			repository_id TEXT NOT NULL REFERENCES repositories(id),
			path TEXT NOT NULL,
			sha256 TEXT NOT NULL,
			kind TEXT NOT NULL,
			spec TEXT NOT NULL,
			start_byte INTEGER NOT NULL,
			end_byte INTEGER NOT NULL,
			start_line INTEGER NOT NULL,
			end_line INTEGER NOT NULL,
			PRIMARY KEY(repository_id, path, sha256, start_byte, kind, spec)
		);
		CREATE INDEX IF NOT EXISTS files_snapshot_path ON files(snapshot_id, path);
		CREATE INDEX IF NOT EXISTS symbols_repository ON symbols(repository_id, path, name);
		CREATE INDEX IF NOT EXISTS imports_repository ON imports(repository_id, path, spec);
		CREATE TABLE IF NOT EXISTS import_edges (
			snapshot_id TEXT NOT NULL REFERENCES snapshots(id),
			source_path TEXT NOT NULL,
			source_sha256 TEXT NOT NULL,
			kind TEXT NOT NULL,
			spec TEXT NOT NULL,
			start_byte INTEGER NOT NULL,
			end_byte INTEGER NOT NULL,
			target_path TEXT NOT NULL,
			target_sha256 TEXT NOT NULL,
			PRIMARY KEY(snapshot_id, source_path, source_sha256, start_byte, kind, target_path)
		);
		CREATE INDEX IF NOT EXISTS import_edges_target ON import_edges(snapshot_id, target_path);`); err != nil {
		return fmt.Errorf("initialize Athena schema: %w", err)
	}
	return nil
}
