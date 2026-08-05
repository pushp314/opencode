# Engineering Review: Repository Inventory

## Scope

This review covers only `athena/cmd/athena index` and `athena/repository`. The capability reads Git-visible text files, computes SHA-256 hashes, and persists immutable snapshots in local SQLite. It does not parse source, call Ollama, generate embeddings, modify a repository, or replace OpenCode behavior.

## Correctness

The implementation canonicalizes the root, requires a Git worktree, lists tracked and non-ignored files through Git, rejects path escapes, skips symbolic links and binary files, and computes deterministic snapshots from sorted path/hash/size records. A real Git-and-SQLite test verifies initial persistence, unchanged snapshot reuse, modified-file invalidation, ignored-file exclusion, and binary-file exclusion.

## Maintainability and Architecture

The public surface is restricted to repository reference, snapshot, report, store, and indexer contracts. SQLite is the only persistence implementation; the `Store` interface is the single test and future-adapter seam. The CLI is a thin presentation adapter. No Brain, execution, or UI subsystem imports repository internals.

## Performance and Memory

The first version reads files sequentially to preserve deterministic error reporting and bounded memory. On Apple M4, the unchanged-fixture benchmark is 7.21 ms/op for 10 iterations. It stores metadata and hashes, not source text or ASTs. Large-repository concurrency is deferred until the repository-intelligence migration explicitly introduces bounded parallel parsing and benchmarks it.

## Security and Safety

The command is read-only with respect to its target repository. It follows Git's ignore rules, skips symlinks, validates relative paths, uses SHA-256 only for identity, writes application state outside the target by default, and permits a user-selected SQLite path. SIGINT and SIGTERM cancel the CLI context.

## Error Handling and Logging

Errors carry the failed operation and are returned to the CLI with a non-zero exit. JSON output contains a stable report contract; human output contains snapshot, count, change, skip, and reuse status. The capability has no hidden network, cloud, model, or telemetry behavior.

## Verification Evidence

```text
go vet ./...
go test -race ./...
go test -run '^$' -bench BenchmarkIndexerUnchanged -benchtime=10x ./repository
```

All commands passed on macOS arm64. The benchmark was `6,918,262 ns/op` on the focused real Git fixture. An end-to-end index of this checkout found 1,284 eligible text files, skipped 5 binary files, then reused the same snapshot with zero changes.

## Decision

Accepted as the first Athena capability. The next migration may consume these snapshots only after its own dossier, implementation, verification, benchmark, and review are complete.
