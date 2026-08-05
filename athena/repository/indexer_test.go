package repository

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestIndexerPersistsAndReusesSnapshots(t *testing.T) {
	root := testRepository(t)
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
	first, err := indexer.Index(context.Background(), Ref{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	if first.Files != 2 || first.Changed != 2 || first.Skipped != 1 || first.Reused {
		t.Fatalf("unexpected first report: %+v", first)
	}
	second, err := indexer.Index(context.Background(), Ref{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	if second.SnapshotID != first.SnapshotID || second.Changed != 0 || !second.Reused {
		t.Fatalf("unexpected reused report: %+v", second)
	}
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	third, err := indexer.Index(context.Background(), Ref{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	if third.SnapshotID == first.SnapshotID || third.Changed != 1 || third.Reused {
		t.Fatalf("unexpected changed report: %+v", third)
	}
}

func BenchmarkIndexerUnchanged(b *testing.B) {
	root := testRepository(b)
	store, err := OpenSQLite(filepath.Join(b.TempDir(), "athena.db"))
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() {
		if err := store.Close(); err != nil {
			b.Error(err)
		}
	})
	indexer := NewIndexer(store)
	if _, err := indexer.Index(context.Background(), Ref{Root: root}); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for range b.N {
		if _, err := indexer.Index(context.Background(), Ref{Root: root}); err != nil {
			b.Fatal(err)
		}
	}
}

func testRepository(t testing.TB) string {
	root := t.TempDir()
	runGit(t, root, "init")
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("ignored.txt\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "ignored.txt"), []byte("ignore me\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "binary.bin"), []byte{0, 1, 2}, 0o600); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "add", ".gitignore", "main.go")
	return root
}

func runGit(t testing.TB, root string, args ...string) {
	command := exec.Command("git", append([]string{"-C", root}, args...)...)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatal(string(output), err)
	}
}
