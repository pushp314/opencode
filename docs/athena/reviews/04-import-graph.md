# Engineering Review: Import Graph Edges

## 1. Scope

This review covers the resolved cross-file import graph added to `athena/repository/graph.go` and its persistence/query/CLI integration across `sqlite.go`, `indexer.go`, `query.go`, and `cmd/athena/main.go` (`athena edges`). For each captured import reference, a deterministic resolver maps a module or include specifier to a file already in the same snapshot and persists a typed `import_edges` row carrying both source and target evidence (path + SHA-256). It does not add qualified-type resolution, call/reference graphs, or the `athena/knowledge` ontology graph; it remains a repository-intelligence graph slice contributing facts only.

## 2. Correctness

Resolution is deterministic and evidence-scoped. Only references whose specifier maps to a file in the snapshot produce an edge; every edge carries `source_path`, `source_sha256`, `target_path`, `target_sha256`, specifier, import kind, and byte offset. Supported resolvable forms:

- **TypeScript/JavaScript** `import`/`export`/`require` specifiers starting with `./` or `../`, resolved relative to the importing file, with deterministic extension fallback (`.ts`→`.tsx`→`.js`→`.jsx`→`.mjs`→`.cjs` for TS sources; the JS order mirrors the importing extension) and `dir/index` fallback.
- **C/C++** `include` paths resolved relative to the source file, with header/implementation extension fallback.
- **Python** package-relative `from X import` module references resolved through the repository's `__init__.py` package tree: a source file belongs to the nearest ancestor directory that is a package (a directory containing `__init__.py`), each leading dot beyond the first drops one package level, and the module remainder resolves to a sibling module file (`module.py`) or a nested package's `__init__.py`. A bare `.`/`..` reference resolves to that package's own `__init__.py`. Directories without `__init__.py` are left transparent, matching Python package semantics rather than raw directory nesting. Relative references naming an absent module simply produce no edge.
- Absolute package names (`react`, Go stdlib), system includes not present in the repo (`<vector>`), and Python absolute module names produce no edge; they remain queryable import facts.

Snapshots identity is unchanged: edge creation never alters `(repository ID, fingerprint)`. The edge set for a snapshot is derived by reading the snapshot's complete import set after parse facts are saved, so edges cover unchanged and recently parsed content, and are recomputed (delete-and-replace per snapshot) on every parse run, including steady-state reuse. This makes incremental edits and post-upgrade backfill deterministic. `athena edges` supports path/kind/target-source filters and JSON output.

## 3. Architecture

`graph.go` is a pure, read-only resolver (`resolveImports`) with no I/O, adjacent to the dictionary in `query.go`; `parse` remains dependency-free. The `Store` interface grows by `Edges` and `SaveEdges`, SQLite adds an `import_edges` table plus a snapshot-scoped query, and the CLI remains a thin adapter. No Brain/execution/UI subsystem depends on repository internals.

## 3. Performance and Safety

Edge resolution is O(resolutions) against a directory map built once per snapshot, plus one delete/re-insert per snapshot. On this checkout a cold pass produced 1,137 parsed files, 7,355 imports, and 3,400 edges; a reuse run resolved zero files and recomputed the same 3,400 edges. No network, cloud, model, or telemetry behavior exists; parser/cursor/query objects remain `Close()`d.

## 5. Correctness Limits / Future

- Ambiguity: when multiple suffix candidates exist (e.g. `x.ts` and `x.tsx`), the deterministic first-match wins; a larger repertoire or explicit resolver config would order by module resolution rules.
- Python relative `from ..` references are resolved against the snapshot's `__init__.py` package tree (namespaced dirs without `__init__.py` stay transparent, so multi-level namespace packages may under-resolve); Rust `use super::`/`crate::` and Go absolute package references (which need `go.mod` module mapping) are not yet captured or resolved. These are the documented next increments (qualified-type resolution, per-language module semantics).
- Self-edges: a re-export `export * from "./account"` in `account.ts` resolves to `account.ts`; valid structurally but should be treated as re-export, not dependency, in later consumers.

## 6. Verification

```text
go vet ./...
go test ./...
go test -race ./...
```

All accepted on macOS arm64. Unit tests cover resolution, rejection of external/unresolved refs, determinism, and ordering; integration tests cover persistence, filters, incremental changes, and steady-state reuse. End-to-end: a cold `index --parse` produced 7,355 imports and 3,400 edges; `reused=true edges=3400` on rerun; `athena edges` printed resolved edges with both-end evidence.

## 7. Decision

Accepted as the durable cross-file import-graph slice of the repository-intelligence roadmap. Qualified-type resolution and language-specific module resolution (Python/Rust/Go) are future increments that may consume these facts.