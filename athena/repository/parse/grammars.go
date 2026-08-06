package parse

import (
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_c "github.com/tree-sitter/tree-sitter-c/bindings/go"
	tree_sitter_cpp "github.com/tree-sitter/tree-sitter-cpp/bindings/go"
	tree_sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
	tree_sitter_java "github.com/tree-sitter/tree-sitter-java/bindings/go"
	tree_sitter_javascript "github.com/tree-sitter/tree-sitter-javascript/bindings/go"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
	tree_sitter_rust "github.com/tree-sitter/tree-sitter-rust/bindings/go"
	ts_typescript "github.com/tree-sitter/tree-sitter-typescript/bindings/go"
)

// grammars maps the file extensions handled by each grammar to a Lang. The
// Name of each Lang is the canonical repository language label and matches the
// labels produced by the repository indexer for the same extensions.
var grammars = map[string]Lang{
	".go":   newLang("go", tree_sitter.NewLanguage(tree_sitter_go.Language()), goExtractor, goImportExtractor),
	".ts":   newLang("typescript", tree_sitter.NewLanguage(ts_typescript.LanguageTypescript()), typescriptExtractor, typescriptImportExtractor),
	".tsx":  newLang("typescript", tree_sitter.NewLanguage(ts_typescript.LanguageTSX()), typescriptExtractor, typescriptImportExtractor),
	".js":   newLang("javascript", tree_sitter.NewLanguage(tree_sitter_javascript.Language()), javascriptExtractor, javascriptImportExtractor),
	".jsx":  newLang("javascript", tree_sitter.NewLanguage(tree_sitter_javascript.Language()), javascriptExtractor, javascriptImportExtractor),
	".mjs":  newLang("javascript", tree_sitter.NewLanguage(tree_sitter_javascript.Language()), javascriptExtractor, javascriptImportExtractor),
	".cjs":  newLang("javascript", tree_sitter.NewLanguage(tree_sitter_javascript.Language()), javascriptExtractor, javascriptImportExtractor),
	".py":   newLang("python", tree_sitter.NewLanguage(tree_sitter_python.Language()), pythonExtractor, pythonImportExtractor),
	".rs":   newLang("rust", tree_sitter.NewLanguage(tree_sitter_rust.Language()), rustExtractor, rustImportExtractor),
	".java": newLang("java", tree_sitter.NewLanguage(tree_sitter_java.Language()), javaExtractor, javaImportExtractor),
	".c":    newLang("c", tree_sitter.NewLanguage(tree_sitter_c.Language()), cExtractor, cIncludeExtractor),
	".h":    newLang("c", tree_sitter.NewLanguage(tree_sitter_c.Language()), cExtractor, cIncludeExtractor),
	".cc":   newLang("cpp", tree_sitter.NewLanguage(tree_sitter_cpp.Language()), cppExtractor, cppIncludeExtractor),
	".cpp":  newLang("cpp", tree_sitter.NewLanguage(tree_sitter_cpp.Language()), cppExtractor, cppIncludeExtractor),
	".cxx":  newLang("cpp", tree_sitter.NewLanguage(tree_sitter_cpp.Language()), cppExtractor, cppIncludeExtractor),
	".hpp":  newLang("cpp", tree_sitter.NewLanguage(tree_sitter_cpp.Language()), cppExtractor, cppIncludeExtractor),
}

func newLang(name string, language *tree_sitter.Language, build func(*tree_sitter.Language) *extractor, buildImports func(*tree_sitter.Language) *importExtractor) Lang {
	return Lang{name: name, language: language, extract: build(language), imports: buildImports(language)}
}
