# Engineering Review: Durable Import Facts

## Scope

This review covers the durable import/export reference facts added to `athena/repository/parse` and the persistence/query/CLI extensions in `athena/repository/indexer.go`, `sqlite.go`, and `query.go`, plus the `athena imports` command. For each parsed file, the capability extracts module, package, and header references (imports, exports, includes, requires, and Rust `use` trees) and persists them as durable `imports` rows keyed by content SHA-256. It does not add qualified-type resolution, import-graph resolution across files, call/reference graphs, or embeddings. It is read-only with respect to the target repository.

This is the durable "edges" slice of the recorded next increment ("qualified-type resolution and durable import/edge facts"). Qualified-type resolution remains un-implemented and is a distinct future increment.

## Correctness

Import extraction is incremental in the same model as symbols: only files whose `(repository_id, path, content SHA-256)` parse fact is absent are parsed, so unchanged content is never re-parsed and the `imports` rows are written only for files in the current parse batch. Snapshot identity is unchanged because it still hashes only `(repository ID, fingerprint)`; import facts never alter the deterministic snapshot ID. Extraction uses compile-time-verified Tree-sitter queries with a single named `@spec` capture that a per-language normalizer validates and strips. Impossible-to-attach predicate groups (which the Go binding counts as extra pattern indices) are deliberately avoided; the CommonJS `require` reference is instead recognized by walking from the string argument to its enclosing `call_expression` and verifying its `function` field. Queries are per language:

- **Go**: `interpreted_string_literal` under `import_spec` (kind `import`).
- **TypeScript/TypeScript+JS/JavaScript**: `source` of `import_statement` and `export_statement` (kinds `import`/`export`), plus `require(...)` (kind `require`).
- **Python**: `name` of `import_statement` and `module_name` of `import_from_statement` (kind `import`); relative `/` references are intentionally omitted.
- **Rust**: whole `use_declaration` normalized to its path text (kind `use`).
- **Java**: `scoped_identifier`/`identifier` of `import_declaration`, including static imports (kind `import`).
- **C / C++**: `string_literal` and `system_lib_string` of `preproc_include` (kind `include`), normalized by stripping quotes/angle brackets.

Line numbers are 1-based and byte offsets are stable. Fixture tests cover all eight language groups and reject non-`require` calls. A real Git-and-SQLite suite verifies counts, reuse parses zero files, and path/spec/kind query filters. Malformed input cannot panic the parser (inherits the shared table test and fuzz target).

## Maintainability and Architecture

The import projection is a self-contained sibling of the symbol extractor: `importExtractor` and `importPattern` in `extract.go`, per-language extractors beside their symbol extractors, and `Lang.imports` built in `grammars.go`. The `Store` interface gains one method (`Imports`), and the SQLite store adds an `imports` table plus a snapshot-scoped query that joins through `files` so evidence is always scoped to captured content. The CLI gains a thin `athena imports` adapter. No Brain, execution, or UI subsystem imports repository internals.

## Performance and Memory

Import extraction reuses the same Tree-sitter parse tree as symbol extraction, one parse per file per content hash, bounded to the existing worker pool (capped at 8). Cold parse on this checkout of 1,296 eligible files produced 1,135 parsed files, 5 parse errors, 27,263 symbols, and 7,348 import references; a reuse run parsed zero files and reused the snapshot. SQLite stores `imports` rows, not source text or ASTs.

## Security and Safety

The capability inherits the inventory and symbol-facts read-only posture and Git path trust boundaries, bounds concurrency, and cancels on SIGINT/SIGTERM through the shared context. Parser/tree/cursor/query objects are always `Close()`d. No hidden network, cloud, model, or telemetry behavior exists. The CGO dependency remains as documented for the official Tree-sitter bindings.

## Error Handling and Logging

Parse failures are counted per file and surfaced in `parse_errors` without aborting the index; fatal errors abort with a non-zero exit. A repository with no snapshot returns a typed error from `imports`. JSON output uses the stable symbol report contract plus the new `imports` count and the import record contract.

## Verification Evidence

```text
go vet ./...
go test ./...
go test -race ./...
```

All passed on macOS arm64. End-to-end verification of this checkout: cold `index --parse --json` reported 1,296 files, 1,135 parsed, 5 parse errors, 27,263 symbols, 7,348 `imports`; the reuse run reported `changed=0 parsed=0 reused=true`; `athena imports` verified kind, spec, and exact-path filters plus JSON output with path/SHA-256 evidence.

## Decision

Accepted as the durable module-reference slice of the repository-intelligence "qualified-type resolution and durable import/edge facts" increment. Qualified-type resolution and cross-file import-graph resolution remain future increments; they may consume these facts only after their own dossier, implementation, verification, benchmark, and review.