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

func (s *SQLiteStore) Save(ctx context.Context, repositoryID string, root string, snapshot Snapshot, files []File) error {
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
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit snapshot transaction: %w", err)
	}
	return nil
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
		CREATE INDEX IF NOT EXISTS files_snapshot_path ON files(snapshot_id, path);`); err != nil {
		return fmt.Errorf("initialize Athena schema: %w", err)
	}
	return nil
}
