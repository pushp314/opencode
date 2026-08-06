package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pushp314/opencode/athena/repository/parse"
)

// maxParseWorkers bounds the number of files parsed concurrently so that a
// repository index never consumes unbounded CPU.
const maxParseWorkers = 8

type Ref struct {
	Root string
}

type File struct {
	Path     string
	SHA256   string
	Size     int64
	Language string
}

type Snapshot struct {
	ID           string
	RepositoryID string
	Revision     string
	Fingerprint  string
	CreatedAt    time.Time
}

type Report struct {
	SnapshotID   string    `json:"snapshot_id"`
	RepositoryID string    `json:"repository_id"`
	Root         string    `json:"root"`
	Revision     string    `json:"revision"`
	Files        int       `json:"files"`
	Changed      int       `json:"changed"`
	Skipped      int       `json:"skipped"`
	Reused       bool      `json:"reused"`
	Parsed       int       `json:"parsed"`
	ParseErrors  int       `json:"parse_errors"`
	Symbols      int       `json:"symbols"`
	Imports      int       `json:"imports"`
	Edges        int       `json:"edges"`
	CompletedAt  time.Time `json:"completed_at"`
}

// IndexOptions bounds an index run. Parse enables durable parse and symbol
// facts; Workers bounds parse concurrency and defaults to a fixed bound.
type IndexOptions struct {
	Parse   bool
	Workers int
}

// ParseKey identifies one file content in the parse cache.
type ParseKey struct {
	Path   string
	SHA256 string
}

// ParseFact is the durable result of parsing one file. Skipped facts are the
// result of a file changing after hashing and are never persisted.
type ParseFact struct {
	Path      string
	SHA256    string
	Language  string
	HasErrors bool
	Symbols   []parse.Symbol
	Imports   []parse.Import
	Skipped   bool
}

type Store interface {
	Latest(context.Context, string) (Snapshot, []File, bool, error)
	Save(context.Context, string, string, Snapshot, []File, []ParseFact) error
	ParsedSet(context.Context, string) (map[ParseKey]struct{}, error)
	SaveParse(context.Context, string, []ParseFact) error
	Symbols(context.Context, string, SymbolQuery) ([]Symbol, error)
	Imports(context.Context, string, ImportQuery) ([]Import, error)
	SaveEdges(context.Context, string, string, []Edge) error
	Edges(context.Context, string, EdgeQuery) ([]Edge, error)
}

type Indexer interface {
	Index(context.Context, Ref, IndexOptions) (Report, error)
}

type indexer struct {
	store Store
}

func NewIndexer(store Store) Indexer {
	return indexer{store: store}
}

func (i indexer) Index(ctx context.Context, ref Ref, options IndexOptions) (Report, error) {
	root, err := normalizeRoot(ref.Root)
	if err != nil {
		return Report{}, err
	}
	paths, err := trackedPaths(ctx, root)
	if err != nil {
		return Report{}, err
	}
	files := make([]File, 0, len(paths))
	skipped := 0
	for _, relative := range paths {
		if err := ctx.Err(); err != nil {
			return Report{}, err
		}
		file, include, err := inspect(root, relative)
		if err != nil {
			return Report{}, err
		}
		if !include {
			skipped++
			continue
		}
		files = append(files, file)
	}
	sort.Slice(files, func(left, right int) bool {
		return files[left].Path < files[right].Path
	})

	repositoryID := digest(root)
	revision, err := gitRevision(ctx, root)
	if err != nil {
		return Report{}, err
	}
	fingerprint := fileFingerprint(files)
	latest, previousFiles, found, err := i.store.Latest(ctx, repositoryID)
	if err != nil {
		return Report{}, err
	}
	changed := changedCount(previousFiles, files)
	completedAt := time.Now().UTC()

	report := Report{
		RepositoryID: repositoryID,
		Root:         root,
		Files:        len(files),
		Changed:      changed,
		Skipped:      skipped,
		CompletedAt:  completedAt,
	}
	var facts []ParseFact
	if options.Parse {
		facts, err = i.reconcileParses(ctx, root, repositoryID, files, options.Workers)
		if err != nil {
			return Report{}, err
		}
		report.Parsed, report.ParseErrors, report.Symbols, report.Imports = summarizeFacts(facts)
	}
	if found && changed == 0 {
		if options.Parse {
			if err := i.store.SaveParse(ctx, repositoryID, facts); err != nil {
				return Report{}, err
			}
			if report.Edges, err = i.resolveEdges(ctx, repositoryID, latest.ID, files); err != nil {
				return Report{}, err
			}
		}
		report.SnapshotID = latest.ID
		report.Revision = latest.Revision
		report.Reused = true
		return report, nil
	}
	snapshot := Snapshot{
		ID:           digest(repositoryID + "\n" + fingerprint),
		RepositoryID: repositoryID,
		Revision:     revision,
		Fingerprint:  fingerprint,
		CreatedAt:    completedAt,
	}
	if err := i.store.Save(ctx, repositoryID, root, snapshot, files, facts); err != nil {
		return Report{}, err
	}
	report.SnapshotID = snapshot.ID
	report.Revision = revision
	if options.Parse {
		if report.Edges, err = i.resolveEdges(ctx, repositoryID, snapshot.ID, files); err != nil {
			return Report{}, err
		}
	}
	return report, nil
}

// resolveEdges derives the cross-file import graph for a snapshot from its
// complete import set and persists it. The snapshot's imports are loaded after
// parse facts are saved so that edges cover unchanged as well as recently
// parsed content.
func (i indexer) resolveEdges(ctx context.Context, repositoryID string, snapshotID string, files []File) (int, error) {
	imports, err := i.store.Imports(ctx, snapshotID, ImportQuery{})
	if err != nil {
		return 0, err
	}
	edges := resolveImports(files, imports)
	if err := i.store.SaveEdges(ctx, repositoryID, snapshotID, edges); err != nil {
		return 0, err
	}
	return len(edges), nil
}

// reconcileParses parses every file in the snapshot whose content has no
// durable parse fact yet. Unsupported languages are skipped and are retried
// on later runs so that new grammars can backfill without migration.
func (i indexer) reconcileParses(ctx context.Context, root string, repositoryID string, files []File, workers int) ([]ParseFact, error) {
	parsed, err := i.store.ParsedSet(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	candidates := make([]File, 0, len(files))
	for _, file := range files {
		if _, ok := parsed[ParseKey{Path: file.Path, SHA256: file.SHA256}]; ok {
			continue
		}
		if _, ok := parse.Detect(file.Path); !ok {
			continue
		}
		candidates = append(candidates, file)
	}
	return i.parseFiles(ctx, root, candidates, workers)
}

func (i indexer) parseFiles(ctx context.Context, root string, files []File, workers int) ([]ParseFact, error) {
	if len(files) == 0 {
		return nil, nil
	}
	if workers <= 0 {
		workers = runtime.GOMAXPROCS(0)
	}
	workers = min(workers, len(files), maxParseWorkers)
	jobs := make(chan File)
	results := make(chan ParseFact)
	var group sync.WaitGroup
	var errored atomic.Bool
	var mutex sync.Mutex
	var firstErr error
	for range workers {
		group.Add(1)
		go func() {
			defer group.Done()
			for file := range jobs {
				if ctx.Err() != nil {
					return
				}
				if errored.Load() {
					continue
				}
				fact, err := i.parseFile(root, file)
				if err != nil {
					mutex.Lock()
					if !errored.Load() {
						errored.Store(true)
						firstErr = err
					}
					mutex.Unlock()
					continue
				}
				results <- fact
			}
		}()
	}
	go func() {
		defer close(jobs)
		for _, file := range files {
			select {
			case jobs <- file:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		group.Wait()
		close(results)
	}()
	facts := []ParseFact{}
	for fact := range results {
		facts = append(facts, fact)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if firstErr != nil {
		return nil, firstErr
	}
	sort.Slice(facts, func(left, right int) bool {
		return facts[left].Path < facts[right].Path
	})
	return facts, nil
}

func (i indexer) parseFile(root string, file File) (ParseFact, error) {
	path := filepath.Join(root, file.Path)
	content, err := os.ReadFile(path)
	if err != nil {
		return ParseFact{}, fmt.Errorf("read %s for parsing: %w", file.Path, err)
	}
	if hashBytes(content) != file.SHA256 {
		return ParseFact{Path: file.Path, SHA256: file.SHA256, Skipped: true}, nil
	}
	lang, _ := parse.Detect(file.Path)
	result := lang.Parse(content)
	return ParseFact{
		Path:      file.Path,
		SHA256:    file.SHA256,
		Language:  result.Language,
		HasErrors: result.HasErrors,
		Symbols:   result.Symbols,
		Imports:   result.Imports,
	}, nil
}

func summarizeFacts(facts []ParseFact) (parsed int, parseErrors int, symbols int, imports int) {
	for _, fact := range facts {
		if fact.Skipped {
			continue
		}
		parsed++
		if fact.HasErrors {
			parseErrors++
		}
		symbols += len(fact.Symbols)
		imports += len(fact.Imports)
	}
	return parsed, parseErrors, symbols, imports
}

// RepositoryID canonicalizes a repository root and returns its stable ID.
func RepositoryID(input string) (string, error) {
	root, err := normalizeRoot(input)
	if err != nil {
		return "", err
	}
	return digest(root), nil
}

func normalizeRoot(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("repository root is required")
	}
	abs, err := filepath.Abs(input)
	if err != nil {
		return "", fmt.Errorf("resolve repository root: %w", err)
	}
	root, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("resolve repository symlinks: %w", err)
	}
	info, err := os.Stat(root)
	if err != nil {
		return "", fmt.Errorf("stat repository root: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("repository root is not a directory: %s", root)
	}
	return root, nil
}

func trackedPaths(ctx context.Context, root string) ([]string, error) {
	command := exec.CommandContext(ctx, "git", "-C", root, "ls-files", "-co", "--exclude-standard", "-z")
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("list Git repository files: %w", err)
	}
	paths := strings.FieldsFunc(string(output), func(r rune) bool { return r == 0 })
	sort.Strings(paths)
	return paths, nil
}

func gitRevision(ctx context.Context, root string) (string, error) {
	command := exec.CommandContext(ctx, "git", "-C", root, "rev-parse", "--verify", "HEAD")
	output, err := command.Output()
	if err != nil {
		return "unborn", nil
	}
	return strings.TrimSpace(string(output)), nil
}

func inspect(root string, relative string) (File, bool, error) {
	path := filepath.Join(root, relative)
	resolved, err := filepath.Rel(root, path)
	if err != nil {
		return File{}, false, fmt.Errorf("resolve repository path: %w", err)
	}
	if resolved == ".." || strings.HasPrefix(resolved, ".."+string(filepath.Separator)) || filepath.IsAbs(resolved) {
		return File{}, false, fmt.Errorf("refusing path outside repository: %s", relative)
	}
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return File{}, false, nil
		}
		return File{}, false, fmt.Errorf("stat %s: %w", relative, err)
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return File{}, false, nil
	}
	file, err := os.Open(path)
	if err != nil {
		return File{}, false, fmt.Errorf("open %s: %w", relative, err)
	}
	defer file.Close()
	probe := make([]byte, 8192)
	count, err := file.Read(probe)
	if err != nil && err != io.EOF {
		return File{}, false, fmt.Errorf("inspect %s: %w", relative, err)
	}
	if strings.IndexByte(string(probe[:count]), 0) >= 0 {
		return File{}, false, nil
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return File{}, false, fmt.Errorf("rewind %s: %w", relative, err)
	}
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return File{}, false, fmt.Errorf("hash %s: %w", relative, err)
	}
	return File{
		Path:     filepath.ToSlash(relative),
		SHA256:   hex.EncodeToString(hash.Sum(nil)),
		Size:     info.Size(),
		Language: language(relative),
	}, true, nil
}

func changedCount(previous []File, current []File) int {
	before := make(map[string]File, len(previous))
	for _, file := range previous {
		before[file.Path] = file
	}
	after := make(map[string]File, len(current))
	for _, file := range current {
		after[file.Path] = file
	}
	changed := 0
	for path, file := range after {
		if previous, ok := before[path]; !ok || previous.SHA256 != file.SHA256 || previous.Size != file.Size {
			changed++
		}
	}
	for path := range before {
		if _, ok := after[path]; !ok {
			changed++
		}
	}
	return changed
}

func fileFingerprint(files []File) string {
	lines := make([]string, 0, len(files))
	for _, file := range files {
		lines = append(lines, file.Path+"\t"+file.SHA256+"\t"+fmt.Sprint(file.Size))
	}
	return digest(strings.Join(lines, "\n"))
}

func digest(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}

func hashBytes(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}

func language(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go":
		return "go"
	case ".ts", ".tsx":
		return "typescript"
	case ".js", ".jsx", ".mjs", ".cjs":
		return "javascript"
	case ".py":
		return "python"
	case ".rs":
		return "rust"
	case ".java":
		return "java"
	case ".swift":
		return "swift"
	case ".c", ".h":
		return "c"
	case ".cc", ".cpp", ".cxx", ".hpp":
		return "cpp"
	case ".md":
		return "markdown"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".toml":
		return "toml"
	case ".sql":
		return "sql"
	default:
		return ""
	}
}
