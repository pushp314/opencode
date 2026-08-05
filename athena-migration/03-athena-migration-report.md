# 3. Athena Migration Report — Integration Points (Phase 4)

**Status:** ANALYSIS ONLY. No Athena logic is implemented. This maps the existing OpenCode architecture onto Athena's target architecture and identifies integration points (seams) for replacement. Nothing here is executable yet.

## 3.1 Athena target architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│                       Athena (replaces OpenCode)                      │
│                                                                      │
│  ┌────────────┐   ┌──────────────────┐   ┌─────────────────────────┐ │
│  │ Terminal   │   │ App / Desktop    │   │ Athena CLI / Server     │ │
│  │ (tui)      │   │ (ui, session-ui, │   │ (runtime, HTTP, PTY)    │ │
│  │            │   │  app)            │   │                         │ │
│  └─────┬──────┘   └───────┬──────────┘   └───────────┬─────────────┘ │
│        └──────────────────┴──────┬───────────────────┘               │
│                                  ▼                                  │
│        ┌──────────────────────────────────────────────────────┐     │
│        │  Athena Session Core (sessions, history, context)    │     │
│        │  = existing core/src/session + system-context        │     │
│        └──────────┬───────────────────────────────────────────┘     │
│                   ▼                                                  │
│  ┌─────────────────────┐  ┌──────────────────────┐  ┌─────────────┐  │
│  │ Athena Model        │  │ Athena Repository    │  │ Athena      │  │
│  │ Orchestrator        │  │ Intelligence         │  │ Verification│  │
│  │ (REPLACES provider  │  │ (REPLACES repo       │  │ Pipeline    │  │
│  │  layer + llm)       │  │  index)              │  │ (REPLACES   │  │
│  └─────────────────────┘  └──────────────────────┘  │  tool/verify)│  │
│  ┌─────────────────────┐  ┌──────────────────────┐  └─────────────┘  │
│  │ Athena Context      │  │ Athena Working       │                   │
│  │ Engine (REPLACES    │  │ Memory (REPLACES     │                   │
│  │  context builder)   │  │  session memory)     │                   │
│  └─────────────────────┘  └──────────────────────┘                   │
└──────────────────────────────────────────────────────────────────────┘
```

## 3.2 Component mapping & integration points

| Existing component (module path) | Athena replacement | Integration point (seam to keep clean) |
|---|---|---|
| **Provider Layer** — `packages/core/src/catalog.ts`, `core/src/provider.ts`, `core/src/plugin/provider/` (32 provider plugins), `core/src/aisdk.ts`, `core/src/github-copilot/`, `packages/llm/src/providers/`, `packages/opencode/src/provider/` | **Athena Model Orchestrator** | Seam: `catalog.ts` registry + `llm` `LLMRequest`/`LLMResponse`/`route` algebra. Athena orchestration should register providers through one registry interface so the 32 AI-SDK provider plugins and the native `llm` route runtime collapse behind one adapter. **Do not let new code touch AI-SDK providers directly — go through the orchestrator seam.** |
| **Context Builder** — `packages/core/src/system-context/` (registry, built-ins, context sources), `core/src/instruction-context.ts`, `opencode` mid-conversation system messages, Context Epoch | **Athena Context Engine** | Seam: `SystemContext` algebra (`make`, `combine`, `initialize`, `reconcile`, `replace`) + `SystemContextRegistry` + stable contribution keys. Athena Context Engine owns all baseline/update/removal rendering and epoch lifecycle. Extension point: plugin-defined Context Sources (currently a follow-up). |
| **Memory** — `core/src/session/` (history, projector, prompt, compaction, context-epoch, input/delivery), `core/src/snapshot.ts`, `packages/opencode/src/session/prompt/` (prompt display/history persistence) | **Athena Working Memory** | Seam: durable `session_input` → Session History projection (prompt admission → promotion), compaction events, Context Snapshot, managed tool-output store. Athena Working Memory should own the durable inbox/history split so prompts/queues/steers are single-source-of-truth. |
| **Repository Index** — `core/src/repository-cache.ts`, `core/src/repository.ts`, `core/src/project/`, `core/src/ripgrep/`, `core/src/glob/` (`glob.ts`, `grep.ts` tools), filesystem watchers | **Athena Repository Intelligence** | Seam: `Repository` API + ripgrep-based search + `@parcel/watcher` events + location-scoped filesystem authority. Athena Repository Intelligence replaces ad-hoc glob/grep with a queryable index; the existing `tool/glob`, `tool/grep`, `tool/read-filesystem`, `tool/webfetch`, `tool/websearch` registry entries are the consumption boundary. |
| **Prompt Builder** — `core/src/session/runner/` (request assembly), `opencode/src/session/llm.ts` (AI-SDK vs native llm branch), Context Epoch baseline + Session History selection, `core/src/config/` agent/model/compaction settings | **Athena Context Assembly** | Seam: `SessionRunner` builds one `llm.stream(request)` per provider turn from Baseline System Context + projected Session History + Model Request Options + Generation Controls. Athena Context Assembly owns this single-request assembly so provider-turn boundaries, continuation metadata, and native-continuation keys stay in one place. |
| **Verification** — `core/src/tool/todowrite.ts`, `core/src/session/todo.ts`, `core/src/tool/apply-patch.ts`, `core/src/file-mutation.ts`, `core/src/patch.ts`, tool-output bounding, session todo state | **Athena Verification Pipeline** | Seam: tool registry settlement + durable todo/list state + patch application. Athena Verification Pipeline should add model-executed checks (build/lint/test gates) at the tool-settlement boundary without bypassing the Tool Registry's final size limits. |
| **Session V2 Core** — `core/src/session/` (execution, run-coordinator, runner, store, sql) | **Athena Session Core** (retained largely as-is) | This is the durable spine (AGENTS.md V2 invariants). Athena extends it; does NOT replace the durable-input admission / drain / delivery vocabulary. |
| **Model Catalog & Config** — `core/src/catalog.ts`, `core/src/models-dev.ts`, `core/src/config/` | Athena Model Orchestrator + Athena Config | Seam: `ConfigProvider` self-export pattern (`src/config`). Athena keeps the plugin/config system; provider catalog becomes a facet of the Model Orchestrator. |
| **Plugin API** — `packages/plugin` (v1 hooks, v2 effect/promise) | Athena Plugin SDK | Retained as-is; the plugin host (`core/src/plugin/`) is the extension point for all Athena components. |
| **Protocol/Client** — `packages/protocol`, `packages/client`, `packages/httpapi-codegen` | Athena API contract + Athena Client | The `HttpApi`-as-authoritative-source pipeline is the contract seam: protocol owns group construction, server supplies concrete middleware keys, client emitters generate from IR. Athena extends the contract, never hand-edits generated code. |

## 3.3 Packages kept as-is (no replacement in Phase 4)

- `packages/schema` (contract leaf), `packages/protocol`, `packages/server` (HTTP runtime), `packages/opencode` (composition root), `packages/tui`, `packages/cli`, `packages/client`, `packages/httpapi-codegen`, `packages/codemode`, `packages/effect-drizzle-sqlite`, `packages/script`, `packages/ui`, `packages/session-ui`, `packages/app`, `packages/desktop`.

## 3.4 Known integration gaps / open questions (for Athena design)

1. **AI-SDK dual path**: `opencode/src/session/llm.ts` branches between AI-SDK providers and the native `llm` package route runtime. Athena Model Orchestrator must settle this branch (single provider-turn path) — flagged, not implemented.
2. **Vendored client tarball**: `app`/`session-ui` pin to `vendor/opencode-ai-client-1.17.13-v2.tgz`. Before Athena stabilizes the client contract, decide publish-vs-vendor.
3. **Plugin-defined Context Sources**: the `SystemContextRegistry` seam exists but plugin-defined registration + hot-reload is a documented follow-up. Athena Context Engine should land this first.
4. **`experimental.chat.system.transform`** legacy hook (CONTEXT.md flagged ambiguity): decide whether to port to Context Sources or drop.
5. **Legacy `@opencode-ai/sdk` root vs `/v2`**: `slack`, `plugin`, `opencode` still import legacy root. Athena should finish the `/v2` migration (per CONTEXT.md, `sdk-next` assumes the `@opencode-ai/sdk` name) before dropping the legacy emitters.
6. **Location/workspace placement**: `Location.workspaceID` is reserved for future placement semantics; Athena Repository Intelligence/Working Memory must not assume implicit-local placement permanently.
7. **github-copilot provider** is self-described temporary; fold into Model Orchestrator or delete.

## 3.5 Explicitly NOT in scope for this phase

- No Athena code, packages, or services have been created.
- No provider/model catalog changes.
- No renames performed (`@opencode-ai/*` → `athena/*` left to Phase 5 approval).
