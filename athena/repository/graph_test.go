package repository

import (
	"testing"

	"github.com/pushp314/opencode/athena/repository/parse"
)

func TestResolveImports(t *testing.T) {
	files := []File{
		{Path: "src/app.ts", SHA256: "a"},
		{Path: "src/lib/util.ts", SHA256: "b"},
		{Path: "src/dir/index.ts", SHA256: "d"},
		{Path: "include/util.hpp", SHA256: "f"},
	}
	refs := []Import{
		{Import: parse.Import{Kind: parse.KindImport, Spec: "./lib/util", StartByte: 1, EndByte: 2}, Path: "src/app.ts", SHA256: "a"},
		{Import: parse.Import{Kind: parse.KindInclude, Spec: "../include/util.hpp", StartByte: 3, EndByte: 4}, Path: "src/app.ts", SHA256: "a"},
		{Import: parse.Import{Kind: parse.KindImport, Spec: "./dir", StartByte: 5, EndByte: 6}, Path: "src/app.ts", SHA256: "a"},
		{Import: parse.Import{Kind: parse.KindImport, Spec: "react", StartByte: 9, EndByte: 10}, Path: "src/app.ts", SHA256: "a"},
		{Import: parse.Import{Kind: parse.KindInclude, Spec: "vector", StartByte: 11, EndByte: 12}, Path: "src/app.ts", SHA256: "a"},
	}
	edges := resolveImports(files, refs)

	expect := []Edge{
		{SourcePath: "src/app.ts", SourceSHA: "a", Kind: parse.KindImport, Spec: "./lib/util", TargetPath: "src/lib/util.ts", TargetSHA: "b"},
		{SourcePath: "src/app.ts", SourceSHA: "a", Kind: parse.KindInclude, Spec: "../include/util.hpp", TargetPath: "include/util.hpp", TargetSHA: "f"},
		{SourcePath: "src/app.ts", SourceSHA: "a", Kind: parse.KindImport, Spec: "./dir", TargetPath: "src/dir/index.ts", TargetSHA: "d"},
	}
	if len(edges) != len(expect) {
		t.Fatalf("got %d edges %+v, want %d", len(edges), edges, len(expect))
	}
	for i, want := range expect {
		got := edges[i]
		if got.SourcePath != want.SourcePath || got.TargetPath != want.TargetPath || got.Spec != want.Spec ||
			got.Kind != want.Kind || got.SourceSHA != want.SourceSHA || got.TargetSHA != want.TargetSHA {
			t.Errorf("edge %d = %+v, want %+v", i, got, want)
		}
	}
}

func TestResolvePythonPackageEdges(t *testing.T) {
	files := []File{
		{Path: "pkg/__init__.py", SHA256: "i"},
		{Path: "pkg/mod.py", SHA256: "a"},
		{Path: "pkg/util.py", SHA256: "b"},
		{Path: "pkg/sub/__init__.py", SHA256: "j"},
		{Path: "pkg/sub/impl.py", SHA256: "c"},
		{Path: "pkg/other/helper.py", SHA256: "d"},
	}
	refs := []Import{
		{Import: parse.Import{Kind: parse.KindImport, Spec: ".util", StartByte: 1, EndByte: 2}, Path: "pkg/mod.py", SHA256: "a"},
		{Import: parse.Import{Kind: parse.KindImport, Spec: ".", StartByte: 3, EndByte: 4}, Path: "pkg/sub/impl.py", SHA256: "c"},
		{Import: parse.Import{Kind: parse.KindImport, Spec: "..other.helper", StartByte: 5, EndByte: 6}, Path: "pkg/sub/impl.py", SHA256: "c"},
		{Import: parse.Import{Kind: parse.KindImport, Spec: "..", StartByte: 7, EndByte: 8}, Path: "pkg/sub/impl.py", SHA256: "c"},
		{Import: parse.Import{Kind: parse.KindImport, Spec: "os.path", StartByte: 9, EndByte: 10}, Path: "pkg/mod.py", SHA256: "a"},
	}
	edges := resolveImports(files, refs)
	expect := []Edge{
		// .util from pkg/mod.py -> pkg/util.py (same package)
		{SourcePath: "pkg/mod.py", SourceSHA: "a", Kind: parse.KindImport, Spec: ".util", TargetPath: "pkg/util.py", TargetSHA: "b"},
		// . from pkg/sub/impl.py -> own package __init__
		{SourcePath: "pkg/sub/impl.py", SourceSHA: "c", Kind: parse.KindImport, Spec: ".", TargetPath: "pkg/sub/__init__.py", TargetSHA: "j"},
		// ..other.helper from pkg/sub/impl.py -> one package up -> pkg/other/helper.py
		{SourcePath: "pkg/sub/impl.py", SourceSHA: "c", Kind: parse.KindImport, Spec: "..other.helper", TargetPath: "pkg/other/helper.py", TargetSHA: "d"},
		// .. from pkg/sub/impl.py -> parent package __init__
		{SourcePath: "pkg/sub/impl.py", SourceSHA: "c", Kind: parse.KindImport, Spec: "..", TargetPath: "pkg/__init__.py", TargetSHA: "i"},
	}
	if len(edges) != len(expect) {
		t.Fatalf("got %d edges %+v, want %d", len(edges), edges, len(expect))
	}
	for i, want := range expect {
		got := edges[i]
		if got.SourcePath != want.SourcePath || got.TargetPath != want.TargetPath || got.Spec != want.Spec || got.Kind != want.Kind {
			t.Errorf("edge %d = %+v, want %+v", i, got, want)
		}
	}
}

func TestResolveEdgesOrderingIsDeterministic(t *testing.T) {
	files := []File{{Path: "src/app.ts", SHA256: "a"}, {Path: "src/util.ts", SHA256: "b"}}
	refs := []Import{
		{Import: parse.Import{Kind: parse.KindImport, Spec: "./util", StartByte: 5, EndByte: 6}, Path: "src/app.ts", SHA256: "a"},
		{Import: parse.Import{Kind: parse.KindImport, Spec: "./util", StartByte: 1, EndByte: 2}, Path: "src/app.ts", SHA256: "a"},
	}
	first := resolveImports(files, refs)
	second := resolveImports(files, refs)
	if len(first) != len(second) {
		t.Fatalf("edge sets differ: %d vs %d", len(first), len(second))
	}
	for i := range first {
		if first[i] != second[i] {
			t.Errorf("edge %d differs: %+v vs %+v", i, first[i], second[i])
		}
	}
}
