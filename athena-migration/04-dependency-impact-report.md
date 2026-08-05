# 4. Dependency Impact Report — OpenCode → Athena

**Status:** FOR APPROVAL. Evidence-based dependency verification for every planned removal. Rule applied: *if anything depends on it, do not delete.*

## 4.1 Workspace dependency graph (runtime, verified by grep of imports)

```
(leaf, no workspace deps)
schema ──────────────────────────────────────────────────────────┐
effect-drizzle-sqlite ──► core                                   │
llm ──(2 refs)──► schema                                         │
plugin ──(16 refs)──► sdk                                        │
codemode (leaf)                                                  │
httpapi-codegen (leaf)                                           │
http-recorder (leaf, test-only devDep for core/llm/opencode)     │
effect-sqlite-node (leaf, UNUSED)                                │
identity (leaf, UNUSED except sdks/vscode symlinks)              │

core ──► schema(60), llm(28), plugin(13), effect-drizzle-sqlite(3), sdk(3)
server ──► core(52), protocol(20)
protocol ──► schema(52)
client ──► schema, protocol, httpapi-codegen(dev), server(dev), core(dev)
sdk-next ──► client/effect, core, server
opencode ──► core(581!), llm(27), schema(19), plugin(17), server(13),
             sdk, tui, codemode, protocol(1), http-recorder(dev)
tui ──► core, plugin, sdk(v2,34), ui
cli ──► tui, core, server, sdk/v2, script(dev)
ui (leaf) ──► (nothing)
session-ui ──► ui(132), sdk(17), core(13), client(tarball)
app ──► ui(563), sdk(98), core(65), session-ui(48), client(tarball), schema(1)
desktop ──► app, ui

CONSOLE/STATS/WEB/ENTERPRISE/FUNCTION/SLACK (all leaf for Athena):
console/* ──► console-* siblings + ui only
stats/*   ──► stats-* siblings + ui only
web       ──► plugin(161), sdk(108)  [one-way; nothing depends on web]
enterprise ──► ui, session-ui, core, sdk  [one-way]
function  ──► nothing (deployed via infra/app.ts)
slack     ──► sdk  [one-way]
storybook ──► ui, session-ui, app  [dev-only host]
```

## 4.2 Impact matrix — if a package is removed, what breaks?

| Removed | Breaks (direct) | Broken by transit | Impact after planned co-removals |
|---|---|---|---|
| `effect-sqlite-node` | none (0 imports in core/src) | none | **No impact** — also remove its line in `core/package.json` deps |
| `identity` | `sdks/vscode/images` symlinks | none | Removed together with `sdks/` → **no impact** |
| `core/src/effect/dfdf` | none | none | **No impact** |
| `storybook` | root `dev:storybook`, `storybook.yml`, `@storybook` story globs | stories in ui/session-ui/app lose a host (stories remain in-repo) | Dev-only loss |
| `console/*` | `dev:console` root script, root `workspaces` glob `packages/console/*`, `infra/console.ts`, SST resources, console routes in `pr-standards.yml` | `@opencode-ai/ui` keeps working (console is only a consumer) | Remove workspaces glob + script + infra + workflow refs → **no impact** on product |
| `stats/*` | `dev:stats` root script, root `workspaces` glob `packages/stats/*`, `infra/lake.ts` + `infra/stats.ts`, `stats.yml`, `sst` scripts in stats package.jsons | `@opencode-ai/ui` unaffected | Same as above → **no impact** |
| `web` | `docs-update.yml`, `docs-locale-sync.yml` (web/src/content/docs), `@astrojs/*` deps | none | Remove workflows → **no impact** |
| `enterprise` | `infra/enterprise.ts` (SST Teams SolidStart) | none | **No impact** |
| `function` | `infra/app.ts` (Worker `Api` + SyncServer DO) | none | **No impact** |
| `slack` | none in-repo (legacy `@opencode-ai/sdk` root import is one-way) | none | **No impact** |
| `github/` | `publish-github-action.yml`, `release-github-action.yml`, `opencode.yml`/`review.yml` (invoke action) | none | Remove workflows → **no impact** |
| `sdks/vscode` | `publish-vscode.yml`, `packages/identity` symlinks | none | Remove together → **no impact** |
| `infra/`, `sst.config.ts`, `sst-env.d.ts` | root `sst` devDep, per-package `sst-env.d.ts` (26 files), `sst.config.ts` imports all `infra/*` | cloudflare/AWS/PlanetScale/Stripe/Honeycomb resources | Remove AFTER console/stats/web/enterprise/function; keep core packages that do NOT use SST (`sst` only appears in stats package.jsons + root). Verify no remaining `sst` import. → **no impact** |

**Verified bottom line:** no runtime product package (`schema/core/protocol/server/llm/plugin/codemode/opencode/tui/cli/client/sdk/sdk-next/effect-drizzle-sqlite/script/ui/session-ui/app/desktop`) imports or executes anything in the removal list. Every removal is a leaf removal.

## 4.3 Files that reference removed items (must be edited in the same change)

| Reference site | Referenced removed item | Edit |
|---|---|---|
| `package.json` root | workspaces globs `packages/console/*`, `packages/stats/*`; scripts `dev:console`, `dev:stats`, `dev:storybook`, `dev:web` | Remove globs/scripts; keep `dev`, `dev:desktop` |
| `package.json` root | `sst` devDependency, `@aws-sdk/client-s3`, `heap-snapshot-toolkit` | Trim after infra removal (verify usage) |
| `packages/core/package.json` | `@opencode-ai/effect-sqlite-node` dep | Remove line |
| `packages/opencode/package.json` | `@opencode-ai/http-recorder` devDep (used by 1 recorded test) | Keep (test infra), or keep only if test kept |
| `turbo.json` | `opencode#test`, `@opencode-ai/core#test`, `@opencode-ai/app#test`, `@opencode-ai/ui#test`, `@opencode-ai/session-ui#test` | Keep; add `tui#test`, `client#test`, etc. if desired |
| `.github/workflows/*` | 21 workflows listed in deliverable 2 Groups D/E | Remove/trim |
| `docs/README.md` | `packages/identity` ("Identity management") | Fix tree listing |
| `sst-env.d.ts` (26 package dirs) | SST resource types | Remove after infra removal |
| `.opencode/tool/github-*.ts`, `.opencode/agent/triage.md` | `@opencode-ai/plugin` + team roster | Trim personal config |
| `script/stats.ts`, `script/beta.ts`, `script/changelog.ts`, `script/duplicate-pr.ts`, `script/publish.ts`, `script/github/*` | community/release automation | Trim to Athena needs |
| `patches/install-korean-ime-fix.sh` | none (stray) | Remove |
| `packages/desktop/package.json` | `native:build` → missing `native/` | Fix |

## 4.4 Dependencies that are NOT removable (evidence)

| Dependency | Why it stays |
|---|---|
| `@opentui/{core,keymap,solid}` + `solid-js` | tui, opencode, cli, app, desktop, ui — terminal + web UI runtime |
| `effect` (+ `@effect/platform-node`, `@effect/sql-sqlite-bun`, `@effect/opentelemetry`) | Every runtime package; DI/runtime backbone |
| `drizzle-orm` (+ `drizzle-kit`) | core, opencode, server, desktop — DB layer |
| `@ai-sdk/*` (~24 providers) + `ai` | core/opencode provider catalog (until Athena Model Orchestrator lands) |
| `@opencode-ai/*` workspace self-deps | Runtime graph in §4.1 |
| `patches/*` (15 bun patches) | `bun install` integrity (trustedDependencies + patchedDependencies) |
| `node-pty` / `@lydell/node-pty` / `bun-pty` | PTY terminal runtime |
| `tree-sitter*`, `web-tree-sitter` | parsing/tooling in opencode |
| `@modelcontextprotocol/sdk` | MCP server/client in opencode |
| `zod`, `semver`, `glob`, `ignore`, `minimatch`, `remeda`, `luxon` etc. | shared utility deps used across packages |

## 4.5 Retained package count after cleanup

| Metric | Before | After |
|---|---|---|
| Workspace packages | 35 | 20 (16 runtime + 4 UI) |
| Packages with tests retained | 14 | 14 (schema, protocol, core, llm, codemode, opencode, tui, cli, client, sdk, sdk-next, ui, session-ui, app, effect-drizzle-sqlite, httpapi-codegen, http-recorder, desktop) |
| CI workflows | 25 | ~6 (test, typecheck, generate, publish, containers) |
