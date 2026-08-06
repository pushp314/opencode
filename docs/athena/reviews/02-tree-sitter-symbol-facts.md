# Engineering Review: Tree-sitter Symbol Facts

## Scope

This review covers `athena/repository/parse` and the parse extension of `athena/repository/indexer.go`, `sqlite.go`, and `query.go`, plus the `athena symbols` command. The capability parses supported source files with official Tree-sitter grammars, extracts named declarations, and persists them as durable parse/symbol facts keyed by content SHA-256. It does not build qualified types, import/call/reference graphs, embeddings, or replace OpenCode behavior. It is read-only with respect to the target repository.

## Correctness

Parsing is incremental: only files whose `(repository_id, path, content SHA-256)` parse fact is absent are parsed, so unchanged content is never re-parsed and unsupported-language files are skipped so future grammars can backfill. Snapshot identity hashes only `(repository ID, fingerprint)`, so parse facts never change the deterministic snapshot ID. Extraction uses compile-time-verified Tree-sitter query patterns with named captures; an extractor panic on a static query error is a programming error, not runtime behavior. Symbol line numbers are 1-based and byte offsets are stable. A real Git-and-SQLite test suite verifies extraction, reuse, changed-file-only reparse, snapshot backfill on reuse, query filters, parse-error counting, unsupported-language exclusion, stale-content skip, and worker-count determinism. Malformed input cannot panic the parser, covered by a table test and a fuzz target that completed over 3 million executions without a crash.

## Maintainability and Architecture

The parse domain is a small self-contained package (`parse`) with explicit `Lang`, `Symbol`, `Result`, and `Parse` types; grammars are declared per language in `grammars.go` and extraction queries per language in `extract.go`. The `Store` interface is the single persistence seam, extended with `ParsedSet`, `SaveParse`, and `Symbols`. The CLI remains a thin presentation adapter (`index --parse`, `symbols`). No Brain, execution, or UI subsystem imports repository internals.

## Performance and Memory

Parse work is bounded to a worker pool capped at 8 (defaulting to `GOMAXPROCS`), with deterministic output ordering and early error abort. Cold parse on this checkout of 1,292 eligible files produced 1,135 parsed files, 5 parse errors, and 27,229 symbols in about 0.84 s wall at ~286% CPU; a reuse run parsed zero files and reused the snapshot. Benchmarks on Apple M4 (10 iterations, real fixture):

```text
BenchmarkIndexerUnchanged-10         10.56 ms/op
BenchmarkIndexParseUnchanged-10      10.70 ms/op
BenchmarkIndexParseCold-10           13.83 ms/op
```

SQLite stores parse and symbol rows, not source text or full ASTs.

## Security and Safety

The capability inherits the inventory command's read-only posture and Git path trust boundaries, and additionally bounds concurrency and cancels on SIGINT/SIGTERM through the shared context. Parser/tree/query/cursor objects are always `Close()`d, avoiding known CGO Finalizer issues. The CGO dependency means `athena` no longer builds with `CGO_ENABLED=0`; the official Tree-sitter bindings and grammars are pinned in `go.mod`/`go.sum`. No hidden network, cloud, model, or telemetry behavior exists.

## Error Handling and Logging

Parse failures are counted per file and surfaced in the report (`parse_errors`) without aborting the index; fatal errors abort with a non-zero exit. A repository with no snapshot returns a typed error from `symbols`. JSON output uses the stable report and symbol contracts.

## Verification Evidence

```text
go vet ./...
go test -race ./...
go test -run '^$' -fuzz FuzzParse -fuzztime 15s ./repository/parse
go test -run '^$' -bench 'BenchmarkIndexerUnchanged|BenchmarkIndexParseUnchanged|BenchmarkIndexParseCold' -benchtime=10x ./repository
```

All commands passed on macOS arm64. End-to-end verification of this checkout: cold `index --parse --json` reported 1,292 files, 1,135 parsed, 5 parse errors, 27,229 symbols; the reuse run reported `changed=0 parsed=0 reused=true`; `symbols` queries verified name, kind, exact-path, and JSON output behavior.

## Decision

Accepted as the second Athena capability. The next migration may consume these facts only after its own dossier, implementation, verification, benchmark, and review are complete. The recorded next increment is qualified-type resolution and durable import/edge facts.
