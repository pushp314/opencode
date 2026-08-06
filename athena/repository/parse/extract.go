package parse

import (
	"fmt"
	"strings"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

// pattern describes one symbol query. Captures are the @name node(s) for the
// declaration; name overrides the default capture text and refine adjusts the
// symbol kind from surrounding node context.
type pattern struct {
	kind   Kind
	query  string
	name   func(node *tree_sitter.Node, source []byte) string
	refine func(kind Kind, node *tree_sitter.Node, source []byte) Kind
}

type extractor struct {
	query    *tree_sitter.Query
	patterns []pattern
}

// newExtractor compiles a single query from all patterns of a language. It
// panics on a static query error because patterns are fixed at build time and
// are covered by compile tests.
func newExtractor(language *tree_sitter.Language, name string, patterns []pattern) *extractor {
	source := make([]string, 0, len(patterns))
	for _, p := range patterns {
		source = append(source, p.query)
	}
	query, qerr := tree_sitter.NewQuery(language, strings.Join(source, "\n"))
	if qerr != nil {
		panic(fmt.Sprintf("%s symbol query: %s", name, qerr.Message))
	}
	return &extractor{query: query, patterns: patterns}
}

// run extracts symbols in document order.
func (e *extractor) run(tree *tree_sitter.Tree, source []byte) []Symbol {
	cursor := tree_sitter.NewQueryCursor()
	defer cursor.Close()
	matches := cursor.Matches(e.query, tree.RootNode(), source)
	symbols := []Symbol{}
	for {
		match := matches.Next()
		if match == nil {
			break
		}
		pattern := e.patterns[match.PatternIndex]
		for _, capture := range match.Captures {
			kind := pattern.kind
			if pattern.refine != nil {
				kind = pattern.refine(kind, &capture.Node, source)
			}
			if kind == "" {
				continue
			}
			name := capture.Node.Utf8Text(source)
			if pattern.name != nil {
				name = pattern.name(&capture.Node, source)
			}
			if name == "" {
				continue
			}
			start, end := capture.Node.ByteRange()
			startPoint := capture.Node.StartPosition()
			endPoint := capture.Node.EndPosition()
			symbols = append(symbols, Symbol{
				Name:      name,
				Kind:      kind,
				StartByte: uint32(start),
				EndByte:   uint32(end),
				StartLine: uint32(startPoint.Row + 1),
				EndLine:   uint32(endPoint.Row + 1),
			})
		}
	}
	return symbols
}

// goRefineKind classifies a Go type declaration by its underlying type.
func goRefineKind(kind Kind, node *tree_sitter.Node, source []byte) Kind {
	_ = source
	spec := node.Parent()
	if spec == nil || spec.Kind() != "type_spec" {
		return kind
	}
	underlying := spec.ChildByFieldName("type")
	switch {
	case underlying == nil:
		return kind
	case underlying.Kind() == "struct_type":
		return Struct
	case underlying.Kind() == "interface_type":
		return Interface
	}
	return kind
}

// declaratorName returns the innermost identifier of a C-family declarator
// chain, for example the function name of a function_declarator.
func declaratorName(node *tree_sitter.Node, source []byte) string {
	for node != nil {
		switch node.Kind() {
		case "function_declarator", "pointer_declarator", "array_declarator", "parenthesized_declarator":
			node = node.ChildByFieldName("declarator")
		case "identifier", "field_identifier", "type_identifier":
			return node.Utf8Text(source)
		case "qualified_identifier":
			if node.NamedChildCount() == 0 {
				return ""
			}
			return node.NamedChild(node.NamedChildCount() - 1).Utf8Text(source)
		default:
			return ""
		}
	}
	return ""
}

func goExtractor(language *tree_sitter.Language) *extractor {
	return newExtractor(language, "go", []pattern{
		{kind: Function, query: `(function_declaration name: (identifier) @name)`},
		{kind: Method, query: `(method_declaration name: (field_identifier) @name)`},
		{kind: Type, query: `(type_declaration (type_spec name: (type_identifier) @name))`, refine: goRefineKind},
		{kind: Type, query: `(type_alias name: (type_identifier) @name)`},
		{kind: Const, query: `(const_declaration (const_spec name: (identifier) @name))`},
		{kind: Var, query: `(var_declaration (var_spec name: (identifier) @name))`},
	})
}

func typescriptExtractor(language *tree_sitter.Language) *extractor {
	return newExtractor(language, "typescript", []pattern{
		{kind: Function, query: `(function_declaration name: (identifier) @name)`},
		{kind: Method, query: `(method_definition name: (property_identifier) @name)`},
		{kind: Class, query: `(class_declaration name: (type_identifier) @name)`},
		{kind: Interface, query: `(interface_declaration name: (type_identifier) @name)`},
		{kind: Enum, query: `(enum_declaration name: (identifier) @name)`},
		{kind: Type, query: `(type_alias_declaration name: (type_identifier) @name)`},
		{kind: Method, query: `(method_signature name: (property_identifier) @name)`},
		{kind: Method, query: `(function_signature name: (identifier) @name)`},
		{kind: Var, query: `(lexical_declaration (variable_declarator name: (identifier) @name))`},
	})
}

func javascriptExtractor(language *tree_sitter.Language) *extractor {
	return newExtractor(language, "javascript", []pattern{
		{kind: Function, query: `(function_declaration name: (identifier) @name)`},
		{kind: Method, query: `(method_definition name: (property_identifier) @name)`},
		{kind: Class, query: `(class_declaration name: (identifier) @name)`},
		{kind: Var, query: `(lexical_declaration (variable_declarator name: (identifier) @name))`},
		{kind: Var, query: `(variable_declaration (variable_declarator name: (identifier) @name))`},
	})
}

func pythonExtractor(language *tree_sitter.Language) *extractor {
	return newExtractor(language, "python", []pattern{
		{kind: Function, query: `(function_definition name: (identifier) @name)`},
		{kind: Class, query: `(class_definition name: (identifier) @name)`},
	})
}

func rustExtractor(language *tree_sitter.Language) *extractor {
	return newExtractor(language, "rust", []pattern{
		{kind: Function, query: `(function_item name: (identifier) @name)`},
		{kind: Struct, query: `(struct_item name: (type_identifier) @name)`},
		{kind: Enum, query: `(enum_item name: (type_identifier) @name)`},
		{kind: Interface, query: `(trait_item name: (type_identifier) @name)`},
		{kind: Type, query: `(type_item name: (type_identifier) @name)`},
		{kind: Const, query: `(const_item name: (identifier) @name)`},
		{kind: Var, query: `(static_item name: (identifier) @name)`},
	})
}

func javaExtractor(language *tree_sitter.Language) *extractor {
	return newExtractor(language, "java", []pattern{
		{kind: Class, query: `(class_declaration name: (identifier) @name)`},
		{kind: Interface, query: `(interface_declaration name: (identifier) @name)`},
		{kind: Enum, query: `(enum_declaration name: (identifier) @name)`},
		{kind: Class, query: `(record_declaration name: (identifier) @name)`},
		{kind: Method, query: `(method_declaration name: (identifier) @name)`},
		{kind: Method, query: `(constructor_declaration name: (identifier) @name)`},
	})
}

func cExtractor(language *tree_sitter.Language) *extractor {
	return newExtractor(language, "c", []pattern{
		{kind: Function, query: `(function_definition declarator: (_) @name)`, name: declaratorName},
		{kind: Struct, query: `(struct_specifier name: (type_identifier) @name body: (field_declaration_list))`},
		{kind: Enum, query: `(enum_specifier name: (type_identifier) @name body: (enumerator_list))`},
		{kind: Type, query: `(union_specifier name: (type_identifier) @name body: (field_declaration_list))`},
		{kind: Type, query: `(type_definition declarator: (type_identifier) @name)`},
	})
}

func cppExtractor(language *tree_sitter.Language) *extractor {
	return newExtractor(language, "cpp", []pattern{
		{kind: Function, query: `(function_definition declarator: (_) @name)`, name: declaratorName},
		{kind: Class, query: `(class_specifier name: (type_identifier) @name body: (field_declaration_list))`},
		{kind: Struct, query: `(struct_specifier name: (type_identifier) @name body: (field_declaration_list))`},
		{kind: Enum, query: `(enum_specifier name: (type_identifier) @name body: (enumerator_list))`},
		{kind: Type, query: `(union_specifier name: (type_identifier) @name body: (field_declaration_list))`},
		{kind: Type, query: `(type_definition declarator: (type_identifier) @name)`},
	})
}

// importPattern describes one import query. @spec is the single named capture
// holding the module/path reference; spec normalizes and validates it. Captures
// that are not the module reference (for example a require callee used only for
// a predicate) return ok=false from spec.
type importPattern struct {
	kind  ImportKind
	query string
	spec  func(node *tree_sitter.Node, source []byte) (string, bool)
}

// importExtractor compiles the import queries of a language and projects them
// into durable import facts.
type importExtractor struct {
	query    *tree_sitter.Query
	patterns []importPattern
}

func newImportExtractor(language *tree_sitter.Language, name string, patterns []importPattern) *importExtractor {
	source := make([]string, 0, len(patterns))
	for _, p := range patterns {
		source = append(source, p.query)
	}
	query, qerr := tree_sitter.NewQuery(language, strings.Join(source, "\n"))
	if qerr != nil {
		panic(fmt.Sprintf("%s import query: %s", name, qerr.Message))
	}
	return &importExtractor{query: query, patterns: patterns}
}

// run extracts imports in document order.
func (e *importExtractor) run(tree *tree_sitter.Tree, source []byte) []Import {
	cursor := tree_sitter.NewQueryCursor()
	defer cursor.Close()
	matches := cursor.Matches(e.query, tree.RootNode(), source)
	imports := []Import{}
	for {
		match := matches.Next()
		if match == nil {
			break
		}
		pattern := e.patterns[match.PatternIndex]
		for _, capture := range match.Captures {
			spec, ok := pattern.spec(&capture.Node, source)
			if !ok || spec == "" {
				continue
			}
			start, end := capture.Node.ByteRange()
			startPoint := capture.Node.StartPosition()
			endPoint := capture.Node.EndPosition()
			imports = append(imports, Import{
				Kind:      pattern.kind,
				Spec:      spec,
				StartByte: uint32(start),
				EndByte:   uint32(end),
				StartLine: uint32(startPoint.Row + 1),
				EndLine:   uint32(endPoint.Row + 1),
			})
		}
	}
	return imports
}

// stripQuotedSpec normalizes a quoted module reference, for example an
// interpreted string literal or JavaScript string, to its bare path.
func stripQuotedSpec(node *tree_sitter.Node, source []byte) (string, bool) {
	if node.Kind() != "string" && node.Kind() != "interpreted_string_literal" {
		return "", false
	}
	text := node.Utf8Text(source)
	return strings.Trim(text, `"`), true
}

// includeSpec normalizes a C-family preprocessor include: "header.h" from a
// string literal or <header.h> from a system include.
func includeSpec(node *tree_sitter.Node, source []byte) (string, bool) {
	text := node.Utf8Text(source)
	switch node.Kind() {
	case "string_literal":
		return strings.Trim(text, `"`), true
	case "system_lib_string":
		return strings.NewReplacer("<", "", ">", "").Replace(text), true
	default:
		return "", false
	}
}

// requireSpec normalizes a CommonJS require call's string argument. The @spec
// node is the string argument; its enclosing call expression must have a
// callee that resolves to require.
func requireSpec(node *tree_sitter.Node, source []byte) (string, bool) {
	if node.Kind() != "string" {
		return "", false
	}
	call := node
	for call != nil && call.Kind() != "call_expression" {
		call = call.Parent()
	}
	if call == nil {
		return "", false
	}
	callee := call.ChildByFieldName("function")
	if callee == nil || callee.Kind() != "identifier" || callee.Utf8Text(source) != "require" {
		return "", false
	}
	return strings.Trim(node.Utf8Text(source), `"`), true
}

// goImport tracks Go import specs.
func goImportExtractor(language *tree_sitter.Language) *importExtractor {
	return newImportExtractor(language, "go", []importPattern{
		{kind: KindImport, query: `(import_spec path: (interpreted_string_literal) @spec)`, spec: stripQuotedSpec},
	})
}

// moduleScriptImport extracts static imports, re-exports, and require calls.
func typescriptImportExtractor(language *tree_sitter.Language) *importExtractor {
	return newImportExtractor(language, "typescript", []importPattern{
		{kind: KindImport, query: `(import_statement source: (string) @spec)`, spec: stripQuotedSpec},
		{kind: KindExport, query: `(export_statement source: (string) @spec)`, spec: stripQuotedSpec},
		{kind: KindRequire, query: `(call_expression arguments: (arguments (string) @string))`, spec: requireSpec},
	})
}

func javascriptImportExtractor(language *tree_sitter.Language) *importExtractor {
	return newImportExtractor(language, "javascript", []importPattern{
		{kind: KindImport, query: `(import_statement source: (string) @spec)`, spec: stripQuotedSpec},
		{kind: KindExport, query: `(export_statement source: (string) @spec)`, spec: stripQuotedSpec},
		{kind: KindRequire, query: `(call_expression arguments: (arguments (string) @string))`, spec: requireSpec},
	})
}

func pythonImportExtractor(language *tree_sitter.Language) *importExtractor {
	return newImportExtractor(language, "python", []importPattern{
		{kind: KindImport, query: `(import_statement name: (dotted_name) @spec)`, spec: plainSpec},
		{kind: KindImport, query: `(import_from_statement module_name: (dotted_name) @spec)`, spec: plainSpec},
		{kind: KindImport, query: `(import_from_statement (relative_import) @spec)`, spec: relativeSpec},
	})
}

func rustImportExtractor(language *tree_sitter.Language) *importExtractor {
	return newImportExtractor(language, "rust", []importPattern{
		{kind: KindUse, query: `(use_declaration) @spec`, spec: useSpec},
	})
}

func javaImportExtractor(language *tree_sitter.Language) *importExtractor {
	return newImportExtractor(language, "java", []importPattern{
		{kind: KindImport, query: `(import_declaration (scoped_identifier) @spec)`, spec: plainSpec},
		{kind: KindImport, query: `(import_declaration (identifier) @spec)`, spec: plainSpec},
	})
}

func cIncludeExtractor(language *tree_sitter.Language) *importExtractor {
	return newImportExtractor(language, "c", []importPattern{
		{kind: KindInclude, query: `(preproc_include path: (string_literal) @spec)`, spec: includeSpec},
		{kind: KindInclude, query: `(preproc_include path: (system_lib_string) @spec)`, spec: includeSpec},
	})
}

func cppIncludeExtractor(language *tree_sitter.Language) *importExtractor {
	return newImportExtractor(language, "cpp", []importPattern{
		{kind: KindInclude, query: `(preproc_include path: (string_literal) @spec)`, spec: includeSpec},
		{kind: KindInclude, query: `(preproc_include path: (system_lib_string) @spec)`, spec: includeSpec},
	})
}

// plainSpec accepts any dotted or scoped identifier as-is.
func plainSpec(node *tree_sitter.Node, source []byte) (string, bool) {
	text := node.Utf8Text(source)
	return strings.TrimPrefix(text, "."), text != ""
}

// relativeSpec preserves a Python `from X import` relative module reference,
// including its dot prefix (for example ".util" or "..pkg.util"). A bare
// relative import with no module (from statements earlier in the file) yields
// no spec because there is nothing file-resolvable.
func relativeSpec(node *tree_sitter.Node, source []byte) (string, bool) {
	if node.Kind() != "relative_import" {
		return "", false
	}
	text := node.Utf8Text(source)
	if strings.ReplaceAll(text, ".", "") == "" {
		return "", false
	}
	return text, true
}

// useSpec normalizes a Rust use tree to its full path text.
func useSpec(node *tree_sitter.Node, source []byte) (string, bool) {
	text := strings.TrimSpace(node.Utf8Text(source))
	text = strings.TrimPrefix(text, "use")
	text = strings.Trim(text, " ;")
	text = strings.TrimPrefix(text, "::")
	return text, text != ""
}
