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
	"sort"
	"strings"
	"time"
)

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
	CompletedAt  time.Time `json:"completed_at"`
}

type Store interface {
	Latest(context.Context, string) (Snapshot, []File, bool, error)
	Save(context.Context, string, string, Snapshot, []File) error
}

type Indexer interface {
	Index(context.Context, Ref) (Report, error)
}

type indexer struct {
	store Store
}

func NewIndexer(store Store) Indexer {
	return indexer{store: store}
}

func (i indexer) Index(ctx context.Context, ref Ref) (Report, error) {
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
	if found && changed == 0 {
		return Report{
			SnapshotID:   latest.ID,
			RepositoryID: repositoryID,
			Root:         root,
			Revision:     latest.Revision,
			Files:        len(files),
			Changed:      0,
			Skipped:      skipped,
			Reused:       true,
			CompletedAt:  completedAt,
		}, nil
	}

	snapshot := Snapshot{
		ID:           digest(repositoryID + "\n" + fingerprint),
		RepositoryID: repositoryID,
		Revision:     revision,
		Fingerprint:  fingerprint,
		CreatedAt:    completedAt,
	}
	if err := i.store.Save(ctx, repositoryID, root, snapshot, files); err != nil {
		return Report{}, err
	}
	return Report{
		SnapshotID:   snapshot.ID,
		RepositoryID: repositoryID,
		Root:         root,
		Revision:     revision,
		Files:        len(files),
		Changed:      changed,
		Skipped:      skipped,
		CompletedAt:  completedAt,
	}, nil
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
