# Athena CLI Foundation

This repository is the CLI-only runtime extracted from OpenCode. It retains interactive and non-interactive command execution, configuration, providers, streaming, tools, durable sessions, the local HTTP control surface, and the OpenTUI terminal renderer. Web, desktop, hosted-console, documentation-site, extension, infrastructure, benchmark, test, and example products have been removed.

Athena is not implemented here. The current runtime remains functional and is deliberately kept as the migration baseline.

## Build and development

Install Bun 1.3.14, then run:

```sh
bun install
bun run typecheck
bun run dev -- --help
bun run build
```

`bun run build` builds the current-platform CLI binary without embedding a browser UI. Run `bun run --cwd packages/opencode dev -- --help` for direct command discovery.

## Repository structure

```text
packages/
  opencode/                 CLI entrypoint, commands, runtime composition, sessions, tools
  tui/                      OpenTUI terminal renderer and interaction runtime
  core/                     durable state, configuration, providers, filesystem and tool primitives
  server/                   local CLI HTTP control surface
  protocol/                 HTTP API contracts
  schema/                   shared runtime schemas
  llm/                      provider transport and stream normalization
  plugin/                   plugin and terminal-extension contracts
  codemode/                 confined code-execution tool runtime
  sdk/js/                   local API client used by the CLI and TUI
  effect-drizzle-sqlite/    SQLite Effect/Drizzle adapter
  effect-sqlite-node/       Node SQLite adapter
  script/                   build metadata helpers
patches/                    required runtime dependency patches
```

## CLI architecture and dependency graph

```text
opencode CLI
 ├─ terminal UI ─────────────── tui ──────── plugin, sdk
 ├─ command/runtime/session ─── core ─────── schema, llm, plugin, SQLite adapters
 ├─ streaming/model execution ─ llm ──────── schema
 ├─ local control API ───────── server ───── core, protocol
 ├─ API contracts ───────────── protocol ─── schema
 ├─ sandboxed code tool ─────── codemode
 └─ local API client ────────── sdk
```

All edges above are runtime imports or package dependencies reachable from `packages/opencode/src/index.ts`. Provider SDKs remain because provider selection is dynamic configuration, so their reachability cannot be safely inferred from a single startup path.

## Startup sequence

1. `packages/opencode/src/index.ts` parses CLI flags and selects a command.
2. Command bootstrap creates a location-scoped runtime and loads configuration.
3. The CLI starts its local server and connects the SDK client.
4. TUI or run commands create/resume durable sessions.
5. The session processor resolves an agent, tools, and provider, then consumes one normalized LLM event stream.
6. Events update session state and render through OpenTUI.

## Development workflow

Keep changes inside the runtime dependency graph. Run `bun run typecheck` from the affected package and build the CLI after entrypoint, provider, session, server, or terminal changes. The repository intentionally has no in-tree test suite or browser build; validation is typecheck, binary build, and CLI smoke execution.

## Athena extension points

These seams are retained, functional, and intentionally not replaced:

| Future Athena system | Current seam |
| --- | --- |
| Context Engine | `packages/core/src/system-context` and `packages/opencode/src/session/instruction.ts` |
| Memory | `packages/opencode/src/session` durable state and prompts |
| Planner | `packages/opencode/src/agent` and `packages/opencode/src/session/processor.ts` |
| Repository Intelligence | `packages/core/src/project`, `packages/core/src/git`, and `packages/opencode/src/lsp` |
| Verification | `packages/opencode/src/tool` and session continuation handling |
| Model Routing | `packages/opencode/src/session/llm.ts`, `packages/opencode/src/provider`, and `packages/llm/src/route` |

## Extraction report

The retained package graph above is the dependency report. Removed material includes all web/desktop/console/stats/documentation/SDK-extension/deployment packages, hosted infrastructure, Nix and CI configuration, examples, benchmarks, tests, snapshots, screenshots, and auxiliary release scripts. The only source assets carried across packages are the five notification sounds now owned by `packages/tui/src/assets/audio`.

Build and runtime verification require Bun. This environment did not provide Bun on `PATH`, so `bun run --cwd packages/opencode src/index.ts --help` and `bun run --cwd packages/opencode typecheck` could not be executed here. No successful build or runtime claim is made until the commands in the build section pass on a Bun-equipped environment.
