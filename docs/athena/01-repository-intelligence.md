# Repository Intelligence Migration Dossier

## Current Architecture

OpenCode has project, git, filesystem, LSP, and tool modules in `packages/opencode/src/project`, `git`, `lsp`, and `tool`. They serve sessions and tools but do not provide a durable, versioned repository fact model, incremental AST index, import graph, call graph, or architecture detector.

## Athena Architecture

`athena/repository` owns repository discovery, content hashing, Tree-sitter parsing, symbol extraction, and durable facts. It emits immutable snapshots keyed by repository ID, Git revision, and index generation. Facts include files, languages, imports, symbols, references, dependencies, and detected architectural concepts.

## Comparison and Migration Risks

OpenCode exposes live project and tool views; Athena makes repository evidence durable and snapshot-scoped. Main risks are parser correctness across languages, index time on large worktrees, path escape through symlinks, and stale facts during concurrent edits. The migration mitigates them with read-only indexing, content hashes, explicit snapshot IDs, bounded concurrency, and parity fixtures before any context or execution consumer depends on the index.

## Public Interfaces

```text
RepositoryIndexer.Index(ctx, RepositoryRef, IndexOptions) (IndexReport, error)
RepositoryReader.Snapshot(ctx, RepositoryRef) (RepositorySnapshot, error)
RepositoryQuery.Symbols(ctx, SnapshotID, SymbolQuery) ([]Symbol, error)
RepositoryQuery.Impact(ctx, SnapshotID, ChangeSet) (ImpactReport, error)
```

`IndexOptions` bounds roots, ignored paths, concurrency, and incremental mode. Results always identify their snapshot and evidence locations.

## Internal Components

- `discover`: repository root, Git state, ignore rules, and file inventory.
- `parse`: Tree-sitter grammars and per-file AST extraction.
- `extract`: symbols, imports, exports, calls, tests, and architectural concepts.
- `graph`: import, dependency, and call graph construction.
- `store`: transactional SQLite persistence and content-addressed parse cache.

## Migration Strategy

First capability: an explicit CLI command that inventories a selected repository and persists file hashes only. It is read-only, independently useful, and establishes repository identity and incremental invalidation. Tree-sitter symbols follow only after hash-index correctness is verified.

## Dependency Graph

```mermaid
flowchart LR
  CLI --> Indexer
  Indexer --> Discover
  Indexer --> Parser[Tree-sitter]
  Parser --> Extract
  Extract --> Graph
  Discover --> SQLite
  Extract --> SQLite
  Graph --> SQLite
```

## Sequence Diagram

```mermaid
sequenceDiagram
  participant U as CLI
  participant I as Indexer
  participant G as Git/filesystem
  participant P as Parser
  participant S as SQLite
  U->>I: Index(repository, options)
  I->>G: discover and hash files
  I->>P: parse changed files only
  P-->>I: AST facts
  I->>S: commit snapshot and graph facts
  I-->>U: IndexReport with evidence
```

## Data Flow

Repository bytes → normalized path and hash → parse result → extracted facts → graph edges → SQLite snapshot. No model is involved in fact creation.

## Acceptance Criteria

- Re-indexing an unchanged repository parses zero files.
- Changed, renamed, deleted, ignored, symlink, and binary files produce deterministic reports.
- Every symbol and edge points to a snapshot, path, byte range, and source hash.

## Testing and Benchmark Strategy

Unit-test ignore and hash behavior; integration-test fixture repositories and Git transitions; fuzz parsers with malformed text. Benchmark cold and incremental indexing separately, reporting files/sec, bytes/sec, parse time, SQLite commit time, and peak memory on macOS.

## Documentation

Document supported languages, snapshot schema, ignored-path policy, invalidation semantics, and every query evidence field before exposing the CLI command.
