// Package parse extracts durable symbol facts from source files using
// Tree-sitter grammars. It owns the grammar registry and the AST-to-symbol
// projection; it performs no I/O and no persistence.
package parse

import (
	"path/filepath"
	"strings"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

// Kind classifies a declared symbol.
type Kind string

const (
	Function  Kind = "function"
	Method    Kind = "method"
	Class     Kind = "class"
	Struct    Kind = "struct"
	Interface Kind = "interface"
	Enum      Kind = "enum"
	Type      Kind = "type"
	Const     Kind = "const"
	Var       Kind = "var"
)

// Symbol is a declared identifier together with its source range. Byte ranges
// are offsets into the source; lines are 1-based.
type Symbol struct {
	Name      string
	Kind      Kind
	StartByte uint32
	EndByte   uint32
	StartLine uint32
	EndLine   uint32
}

// ImportKind classifies how a module or header is referenced.
type ImportKind string

const (
	KindImport  ImportKind = "import"
	KindExport  ImportKind = "export"
	KindInclude ImportKind = "include"
	KindRequire ImportKind = "require"
	KindUse     ImportKind = "use"
)

// Import is a reference to an external module, package, or header together with
// its source range. Byte ranges are offsets into the source; lines are 1-based.
type Import struct {
	Kind      ImportKind
	Spec      string
	StartByte uint32
	EndByte   uint32
	StartLine uint32
	EndLine   uint32
}

// Result is the outcome of parsing a single file.
type Result struct {
	Language  string
	HasErrors bool
	Symbols   []Symbol
	Imports   []Import
}

// Lang is a supported language grammar and its symbol and import extractors.
type Lang struct {
	name     string
	language *tree_sitter.Language
	extract  *extractor
	imports  *importExtractor
}

// Name returns the canonical name of the language label for the grammar.
func (l Lang) Name() string { return l.name }

// Parse parses source and returns the durable symbol and import facts for the
// file. Sub-graph facts are extracted from the same AST. Parsing performs no
// I/O and no persistence.
func (l Lang) Parse(source []byte) Result {
	parser := tree_sitter.NewParser()
	defer parser.Close()
	parser.SetLanguage(l.language)
	tree := parser.Parse(source, nil)
	if tree == nil {
		return Result{Language: l.name, HasErrors: true}
	}
	defer tree.Close()
	return Result{
		Language:  l.name,
		HasErrors: tree.RootNode().HasError(),
		Symbols:   l.extract.run(tree, source),
		Imports:   l.imports.run(tree, source),
	}
}

// Detect resolves a path to its language. The second result is false when no
// supported grammar applies to the path.
func Detect(path string) (Lang, bool) {
	lang, ok := grammars[strings.ToLower(filepath.Ext(path))]
	return lang, ok
}
