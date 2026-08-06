package repository

import (
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pushp314/opencode/athena/repository/parse"
)

// Edge is a resolved cross-file module reference within a single repository
// snapshot. Both ends are files captured by the snapshot, so an edge always
// carries source and target evidence.
type Edge struct {
	SourcePath string           `json:"source_path"`
	SourceSHA  string           `json:"source_sha256"`
	Kind       parse.ImportKind `json:"kind"`
	Spec       string           `json:"spec"`
	StartByte  uint32           `json:"start_byte"`
	EndByte    uint32           `json:"end_byte"`
	TargetPath string           `json:"target_path"`
	TargetSHA  string           `json:"target_sha256"`
}

// EdgeQuery filters edges of a snapshot. Empty fields match anything.
type EdgeQuery struct {
	Path string
	Kind string
	To   string
}

// resolveImports produces the cross-file edges whose specifier resolves to a
// file already in the snapshot. External references (stdlib, packages, system
// headers not present in the repo) and unsupported module systems produce no
// edge: they remain import facts but are not graph edges. Resolution is
// deterministic: candidate targets are tried in a fixed order and the first
// existing file wins.
func resolveImports(files []File, imports []Import) []Edge {
	byPath := make(map[string]string, len(files))
	for _, file := range files {
		byPath[file.Path] = file.SHA256
	}
	var edges []Edge
	for _, imp := range imports {
		target, ok := resolveTarget(imp.Path, imp.Kind, imp.Spec, byPath)
		if !ok {
			continue
		}
		sourceSHA, ok := byPath[imp.Path]
		if !ok {
			continue
		}
		edges = append(edges, Edge{
			SourcePath: imp.Path,
			SourceSHA:  sourceSHA,
			Kind:       imp.Kind,
			Spec:       imp.Spec,
			StartByte:  imp.StartByte,
			EndByte:    imp.EndByte,
			TargetPath: target,
			TargetSHA:  byPath[target],
		})
	}
	sort.Slice(edges, func(left, right int) bool {
		a, b := edges[left], edges[right]
		if a.SourcePath != b.SourcePath {
			return a.SourcePath < b.SourcePath
		}
		if a.StartByte != b.StartByte {
			return a.StartByte < b.StartByte
		}
		if a.TargetPath != b.TargetPath {
			return a.TargetPath < b.TargetPath
		}
		return a.Kind < b.Kind
	})
	return edges
}

// resolveTarget maps one import reference to a snapshot file, when possible.
// Only relative module specifiers (./ or ../), Python package-relative
// references, and C-family includes are file-resolvable; absolute package and
// module names are not.
func resolveTarget(source string, kind parse.ImportKind, spec string, byPath map[string]string) (string, bool) {
	var rel string
	switch kind {
	case parse.KindInclude:
		rel = cleanRelPath(spec)
	default:
		if !strings.HasPrefix(spec, ".") {
			return "", false
		}
		if !strings.Contains(spec, "/") {
			return pythonPackageTarget(source, spec, byPath)
		}
		rel = cleanRelPath(spec)
	}
	base := path.Join(path.Dir(source), rel)
	if _, ok := byPath[base]; ok {
		return base, true
	}
	suffixes := suffixesFor(source)
	for _, suffix := range suffixes {
		candidate := base + suffix
		if _, ok := byPath[candidate]; ok {
			return candidate, true
		}
	}
	if kind != parse.KindInclude && suffixesForDirectory(source) {
		for _, suffix := range suffixes {
			candidate := path.Join(base, "index") + suffix
			if _, ok := byPath[candidate]; ok {
				return candidate, true
			}
		}
	}
	return "", false
}

// pythonPackageTarget resolves a package-relative Python reference such as
// `.util`, `..other.helper`, or `.` using the repository's __init__.py package
// tree rather than raw directory hops. The source file is assigned to the
// nearest ancestor directory that is a package, then the leading dots (one
// package level each beyond the first) drop matching package dirs before the
// module remainder is appended.
func pythonPackageTarget(source string, spec string, byPath map[string]string) (string, bool) {
	rest := spec
	dots := 0
	for strings.HasPrefix(rest, ".") {
		rest = rest[1:]
		dots++
	}
	if dots == 0 {
		return "", false
	}
	base := path.Dir(source)
	for base != "." && base != "" {
		if _, ok := byPath[path.Join(base, "__init__.py")]; ok {
			break
		}
		base = path.Dir(base)
	}
	steps := dots - 1
	for steps > 0 && base != "." && base != "" {
		base = path.Dir(base)
		steps--
	}
	if steps > 0 {
		return "", false
	}
	if rel := strings.ReplaceAll(rest, ".", "/"); rel != "" {
		if _, ok := byPath[path.Join(base, rel+".py")]; ok {
			return path.Join(base, rel+".py"), true
		}
		if _, ok := byPath[path.Join(base, rel, "__init__.py")]; ok {
			return path.Join(base, rel, "__init__.py"), true
		}
		return "", false
	}
	target := path.Join(base, "__init__.py")
	if _, ok := byPath[target]; ok {
		return target, true
	}
	return "", false
}

// cleanRelPath normalizes a relative or include specifier to a POSIX path that
// path.Join can consume, preserving ../ traversal.
func cleanRelPath(rel string) string {
	if rel == "." {
		return "."
	}
	return path.Clean(filepath.ToSlash(rel))
}

// suffixesFor returns candidate file suffixes, ordered, for the language of the
// importing file.
func suffixesFor(source string) []string {
	switch strings.ToLower(filepath.Ext(source)) {
	case ".ts", ".tsx":
		return []string{".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs"}
	case ".js", ".jsx", ".mjs", ".cjs":
		return []string{".js", ".jsx", ".mjs", ".cjs", ".ts", ".tsx"}
	case ".c", ".h":
		return []string{".h", ".c"}
	case ".cc", ".cpp", ".cxx", ".hpp":
		return []string{".hpp", ".h", ".cpp", ".cxx", ".cc", ".c"}
	case ".py":
		return []string{".py"}
	default:
		return nil
	}
}

// suffixesForDirectory reports whether +index-style resolution applies to the
// importing language.
func suffixesForDirectory(source string) bool {
	switch strings.ToLower(filepath.Ext(source)) {
	case ".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs":
		return true
	default:
		return false
	}
}
