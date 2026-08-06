package repository

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIndexParseExtractsAndReusesSymbols(t *testing.T) {
	root := parseFixtureRepository(t)
	store, err := OpenSQLite(filepath.Join(t.TempDir(), "athena.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})
	indexer := NewIndexer(store)
	first, err := indexer.Index(context.Background(), Ref{Root: root}, IndexOptions{Parse: true, Workers: 2})
	if err != nil {
		t.Fatal(err)
	}
	if first.Reused || first.Changed != 4 {
		t.Fatalf("unexpected first report: %+v", first)
	}
	if first.Parsed != 3 || first.ParseErrors != 0 || first.Symbols != 5 {
		t.Fatalf("unexpected parse report: %+v", first)
	}

	query := NewQuery(store)
	repositoryID, err := RepositoryID(root)
	if err != nil {
		t.Fatal(err)
	}
	symbols, err := query.Symbols(context.Background(), repositoryID, SymbolQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if len(symbols) != first.Symbols {
		t.Fatalf("symbols query returned %d, report says %d", len(symbols), first.Symbols)
	}
	for _, symbol := range symbols {
		if symbol.SHA256 == "" || symbol.Path == "" {
			t.Fatalf("symbol without evidence: %+v", symbol)
		}
	}

	second, err := indexer.Index(context.Background(), Ref{Root: root}, IndexOptions{Parse: true})
	if err != nil {
		t.Fatal(err)
	}
	if !second.Reused || second.Parsed != 0 {
		t.Fatalf("unchanged repository must parse zero files: %+v", second)
	}
}

func TestIndexParseOnlyReParsesChangedFiles(t *testing.T) {
	root := parseFixtureRepository(t)
	store, err := OpenSQLite(filepath.Join(t.TempDir(), "athena.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})
	indexer := NewIndexer(store)
	if _, err := indexer.Index(context.Background(), Ref{Root: root}, IndexOptions{Parse: true}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\n\nfunc Changed() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	report, err := indexer.Index(context.Background(), Ref{Root: root}, IndexOptions{Parse: true})
	if err != nil {
		t.Fatal(err)
	}
	if report.Reused || report.Changed != 1 || report.Parsed != 1 {
		t.Fatalf("changed repository must parse only the changed file: %+v", report)
	}

	repositoryID, err := RepositoryID(root)
	if err != nil {
		t.Fatal(err)
	}
	symbols, err := NewQuery(store).Symbols(context.Background(), repositoryID, SymbolQuery{})
	if err != nil {
		t.Fatal(err)
	}
	paths := map[string]bool{}
	for _, symbol := range symbols {
		paths[symbol.Path] = true
	}
	if !paths["main.go"] || !paths["app.ts"] || !paths["lib.rs"] {
		t.Fatalf("snapshot symbols missing files: %v", paths)
	}
}

func TestSymbolsQueryFilters(t *testing.T) {
	root := parseFixtureRepository(t)
	store, err := OpenSQLite(filepath.Join(t.TempDir(), "athena.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})
	if _, err := NewIndexer(store).Index(context.Background(), Ref{Root: root}, IndexOptions{Parse: true}); err != nil {
		t.Fatal(err)
	}
	repositoryID, err := RepositoryID(root)
	if err != nil {
		t.Fatal(err)
	}
	query := NewQuery(store)

	byName, err := query.Symbols(context.Background(), repositoryID, SymbolQuery{Name: "Foo"})
	if err != nil {
		t.Fatal(err)
	}
	if len(byName) != 1 || byName[0].Kind != "function" {
		t.Fatalf("unexpected name filter result: %+v", byName)
	}
	byKind, err := query.Symbols(context.Background(), repositoryID, SymbolQuery{Kind: "class"})
	if err != nil {
		t.Fatal(err)
	}
	if len(byKind) != 1 || byKind[0].Name != "App" {
		t.Fatalf("unexpected kind filter result: %+v", byKind)
	}
	byPath, err := query.Symbols(context.Background(), repositoryID, SymbolQuery{Path: "lib.rs"})
	if err != nil {
		t.Fatal(err)
	}
	if len(byPath) != 1 || byPath[0].Name != "root" {
		t.Fatalf("unexpected path filter result: %+v", byPath)
	}
}

func TestIndexParseBackfillsOnReuse(t *testing.T) {
	root := parseFixtureRepository(t)
	store, err := OpenSQLite(filepath.Join(t.TempDir(), "athena.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})
	indexer := NewIndexer(store)
	if _, err := indexer.Index(context.Background(), Ref{Root: root}, IndexOptions{}); err != nil {
		t.Fatal(err)
	}
	backfill, err := indexer.Index(context.Background(), Ref{Root: root}, IndexOptions{Parse: true})
	if err != nil {
		t.Fatal(err)
	}
	if !backfill.Reused || backfill.Parsed != 3 {
		t.Fatalf("upgrade re-index must reuse the snapshot and parse all eligible files: %+v", backfill)
	}
	steady, err := indexer.Index(context.Background(), Ref{Root: root}, IndexOptions{Parse: true})
	if err != nil {
		t.Fatal(err)
	}
	if !steady.Reused || steady.Parsed != 0 {
		t.Fatalf("steady state must parse zero files: %+v", steady)
	}
}

func TestIndexParseCountsParseErrors(t *testing.T) {
	root := t.TempDir()
	runGit(t, root, "init")
	if err := os.WriteFile(filepath.Join(root, "broken.go"), []byte("package main\n\nfunc broken( {\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "add", "broken.go")
	store, err := OpenSQLite(filepath.Join(t.TempDir(), "athena.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})
	report, err := NewIndexer(store).Index(context.Background(), Ref{Root: root}, IndexOptions{Parse: true})
	if err != nil {
		t.Fatal(err)
	}
	if report.Parsed != 1 || report.ParseErrors != 1 {
		t.Fatalf("unexpected parse error report: %+v", report)
	}
}

func TestIndexParseSkipsUnsupportedLanguages(t *testing.T) {
	root := t.TempDir()
	runGit(t, root, "init")
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# Readme\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\n\nfunc Foo() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "add", "README.md", "main.go")
	store, err := OpenSQLite(filepath.Join(t.TempDir(), "athena.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})
	indexer := NewIndexer(store)
	report, err := indexer.Index(context.Background(), Ref{Root: root}, IndexOptions{Parse: true})
	if err != nil {
		t.Fatal(err)
	}
	if report.Files != 2 || report.Parsed != 1 {
		t.Fatalf("unsupported language must not be counted as parsed: %+v", report)
	}
	second, err := indexer.Index(context.Background(), Ref{Root: root}, IndexOptions{Parse: true})
	if err != nil {
		t.Fatal(err)
	}
	if second.Parsed != 0 {
		t.Fatalf("unsupported files must not accumulate parse attempts: %+v", second)
	}
}

func TestParseFileSkipsStaleContent(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "main.go")
	if err := os.WriteFile(path, []byte("package main\n\nfunc Fresh() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	indexer := indexer{}
	fact, err := indexer.parseFile(root, File{Path: "main.go", SHA256: strings.Repeat("0", 64)})
	if err != nil {
		t.Fatal(err)
	}
	if !fact.Skipped {
		t.Fatal("content that does not match its hash must be skipped")
	}
}

func TestIndexParseIsDeterministicAcrossWorkers(t *testing.T) {
	root := parseFixtureRepository(t)
	snapshots := map[int]string{}
	counts := map[int]int{}
	for _, workers := range []int{1, 8} {
		store, err := OpenSQLite(filepath.Join(t.TempDir(), "athena.db"))
		if err != nil {
			t.Fatal(err)
		}
		report, err := NewIndexer(store).Index(context.Background(), Ref{Root: root}, IndexOptions{Parse: true, Workers: workers})
		if err := store.Close(); err != nil {
			t.Fatal(err)
		}
		if err != nil {
			t.Fatal(err)
		}
		snapshots[workers] = report.SnapshotID
		counts[workers] = report.Symbols
	}
	if snapshots[1] != snapshots[8] || counts[1] != counts[8] {
		t.Fatalf("index must be identical across worker counts: %v %v", snapshots, counts)
	}
}

func TestIndexImportsPersistAndQuery(t *testing.T) {
	root := importFixtureRepository(t)
	store, err := OpenSQLite(filepath.Join(t.TempDir(), "athena.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})
	indexer := NewIndexer(store)
	first, err := indexer.Index(context.Background(), Ref{Root: root}, IndexOptions{Parse: true, Workers: 2})
	if err != nil {
		t.Fatal(err)
	}
	if first.Parsed != 3 || first.Imports != 7 {
		t.Fatalf("unexpected import report: %+v", first)
	}

	repositoryID, err := RepositoryID(root)
	if err != nil {
		t.Fatal(err)
	}
	query := NewQuery(store)

	all, err := query.Imports(context.Background(), repositoryID, ImportQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 7 {
		t.Fatalf("imports query returned %d, report says %d", len(all), first.Imports)
	}
	for _, imp := range all {
		if imp.Path == "" || imp.SHA256 == "" || imp.Spec == "" {
			t.Fatalf("import without evidence: %+v", imp)
		}
	}

	bySpec, err := query.Imports(context.Background(), repositoryID, ImportQuery{Spec: "./a"})
	if err != nil {
		t.Fatal(err)
	}
	if len(bySpec) != 1 || bySpec[0].Kind != "import" {
		t.Fatalf("unexpected spec filter result: %+v", bySpec)
	}

	byKind, err := query.Imports(context.Background(), repositoryID, ImportQuery{Kind: "include"})
	if err != nil {
		t.Fatal(err)
	}
	if len(byKind) != 2 {
		t.Fatalf("unexpected kind filter result: %+v", byKind)
	}

	byPath, err := query.Imports(context.Background(), repositoryID, ImportQuery{Path: "app.ts"})
	if err != nil {
		t.Fatal(err)
	}
	if len(byPath) != 4 {
		t.Fatalf("unexpected path filter result: %+v", byPath)
	}

	second, err := indexer.Index(context.Background(), Ref{Root: root}, IndexOptions{Parse: true})
	if err != nil {
		t.Fatal(err)
	}
	if !second.Reused || second.Parsed != 0 {
		t.Fatalf("unchanged repository must parse zero files: %+v", second)
	}
}

func TestIndexEdgesPersistAndQuery(t *testing.T) {
	root := edgesFixtureRepository(t)
	store, err := OpenSQLite(filepath.Join(t.TempDir(), "athena.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})
	indexer := NewIndexer(store)
	first, err := indexer.Index(context.Background(), Ref{Root: root}, IndexOptions{Parse: true, Workers: 2})
	if err != nil {
		t.Fatal(err)
	}
	if first.Parsed != 5 || first.Edges != 2 {
		t.Fatalf("unexpected edge report: %+v", first)
	}

	repositoryID, err := RepositoryID(root)
	if err != nil {
		t.Fatal(err)
	}
	query := NewQuery(store)

	all, err := query.Edges(context.Background(), repositoryID, EdgeQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("edges query returned %d, report says %d", len(all), first.Edges)
	}
	for _, edge := range all {
		if edge.SourcePath == "" || edge.TargetPath == "" || edge.SourceSHA == "" || edge.TargetSHA == "" || edge.Spec == "" {
			t.Fatalf("edge without evidence: %+v", edge)
		}
	}

	bySource, err := query.Edges(context.Background(), repositoryID, EdgeQuery{Path: "src/app.ts"})
	if err != nil {
		t.Fatal(err)
	}
	if len(bySource) != 1 || bySource[0].TargetPath != "src/lib/util.ts" {
		t.Fatalf("unexpected source filter result: %+v", bySource)
	}

	// An incremental change to an unchanged importer must still produce edges
	// for it by recomputing from the snapshot's complete import set.
	if err := os.WriteFile(filepath.Join(root, "src/lib/util.ts"), []byte("export function extra() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	second, err := indexer.Index(context.Background(), Ref{Root: root}, IndexOptions{Parse: true})
	if err != nil {
		t.Fatal(err)
	}
	if second.Reused || second.Changed != 1 || second.Edges != 2 {
		t.Fatalf("incremental change must rebuild the full edge set: %+v", second)
	}

	// A steady-state reuse run keeps edges and resolves nothing new.
	third, err := indexer.Index(context.Background(), Ref{Root: root}, IndexOptions{Parse: true})
	if err != nil {
		t.Fatal(err)
	}
	if !third.Reused || third.Parsed != 0 || third.Edges != 2 {
		t.Fatalf("steady state must keep the edge set: %+v", third)
	}
}

func edgesFixtureRepository(t testing.TB) string {
	root := t.TempDir()
	runGit(t, root, "init")
	files := map[string]string{
		"src/app.ts":       "import L from \"./lib/util\"\n",
		"src/lib/util.ts":  "export const x = 1\n",
		"src/plain.go":     "package main\n\nimport \"fmt\"\n",
		"include/util.hpp": "#pragma once\nint helper();\n",
		"src/consumer.ts":  "import { x } from \"react\"\nimport H from \"../include/util.hpp\"\n",
	}
	for name, content := range files {
		target := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	runGit(t, root, "add", "src/app.ts", "src/lib/util.ts", "src/plain.go", "include/util.hpp", "src/consumer.ts")
	return root
}

func importFixtureRepository(t testing.TB) string {
	root := t.TempDir()
	runGit(t, root, "init")
	files := map[string]string{
		"main.go": "package main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(\"hi\") }\n",
		"app.ts": `import { x } from "./a"
import { Y } from "./b"
export { z } from "./c"
const d = require("./d")
`,
		"util.hpp": `#include <vector>
#include "local.hpp"
`,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	runGit(t, root, "add", "main.go", "app.ts", "util.hpp")
	return root
}

func parseFixtureRepository(t testing.TB) string {
	root := t.TempDir()
	runGit(t, root, "init")
	files := map[string]string{
		"main.go":  "package main\n\nfunc Foo() int { return 1 }\n\ntype Point struct{ X int }\n",
		"app.ts":   "export class App { m() {} }\n",
		"lib.rs":   "fn root() {}\n",
		"notes.md": "# Notes\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	runGit(t, root, "add", "main.go", "app.ts", "lib.rs", "notes.md")
	return root
}
