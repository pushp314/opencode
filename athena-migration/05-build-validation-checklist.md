# 5. Build Validation Checklist — OpenCode → Athena

**Status:** FOR APPROVAL. This is the Phase 6 gate. Run the full checklist after **every** cleanup stage (Stage 1–5 in deliverable 2). If any item fails → rollback (`git checkout` / restore snapshot) before proceeding.

## 5.0 Preconditions

- [ ] Fresh snapshot/branch of the current single commit (`git status` clean, or tag `pre-cleanup`).
- [ ] `bun@1.3.14` available (root `packageManager`).
- [ ] Clean install possible: `rm -rf node_modules` not required unless lockfile changed.

## 5.1 Baseline (before any change) — must pass once

| # | Command | Working dir | Expect |
|---|---|---|---|
| 1 | `bun install` | root | success, lockfile consistent |
| 2 | `bun run typecheck` (turbo typecheck) | root | all packages pass tsgo |
| 3 | `bun run lint` (oxlint) | root | 0 errors (fix `.oxlintrc.json` malformed keys first) |
| 4 | `bun test` | `packages/opencode` | 247 files pass |
| 5 | `bun test` | `packages/core` | 142 files pass |
| 6 | `bun test` | `packages/tui` | 45 files pass |
| 7 | `bun test` | `packages/llm` | 30 files pass |
| 8 | `bun test` | `packages/schema` | 6 files pass |
| 9 | `bun test` | `packages/codemode` | 7 files pass |
| 10 | `bun test` | `packages/client` | 4 files pass (incl. import-boundary) |
| 11 | `bun test` | `packages/sdk-next` | 2 files pass |
| 12 | `bun test` | `packages/sdk` | 1 file passes |
| 13 | `bun test` | `packages/app` | 116 unit + 14 browser pass |
| 14 | `bun test` | `packages/ui`, `packages/session-ui` | 4 + 14 pass |
| 15 | `bun test` | `packages/effect-drizzle-sqlite`, `packages/httpapi-codegen`, `packages/http-recorder` | 1 + 2 + 1 pass |
| 16 | `bun test` | `packages/desktop` | 14 pass |
| 17 | `bun test` | `packages/protocol` | 1 passes |
| 18 | `bun run --cwd packages/opencode build` | `packages/opencode` | standalone binary produced |
| 19 | Smoke: run built binary `--help` / `opencode tui --headless` | — | boots without crash |

## 5.2 Post-stage gates (repeat for each stage)

### Stage 1 — Dead code + examples removal
- [ ] `bun run typecheck` (root)
- [ ] `bun run lint`
- [ ] `bun test` in `packages/core` (142), `packages/opencode` (247) — confirm no import of removed examples/dfdf
- [ ] grep `dfdf|effect-sqlite-node|layer-map.example|packages/identity` → **0 hits in packages/**

### Stage 2 — Leaf product removal (web, enterprise, function, slack, github, sdks)
- [ ] `bun install` (lockfile/workspaces updated)
- [ ] `bun run typecheck` (root)
- [ ] `bun test` in `packages/opencode` (247) — `@opencode-ai/sdk` legacy root import sites unchanged
- [ ] `bun test` in `packages/app`, `packages/session-ui` — vendored client tarball untouched
- [ ] `bun run --cwd packages/opencode build` + binary smoke test
- [ ] grep `@opencode-ai/web|@opencode-ai/enterprise|@opencode-ai/slack|opencode-ai/function` → **0 hits in packages/**

### Stage 3 — SaaS removal (console, stats, infra, sst)
- [ ] `bun install` after workspace glob removal
- [ ] `bun run typecheck` (root) — confirm no `sst` imports remain in retained packages (grep `from "sst"` in packages/core, server, opencode, tui, cli, sdk, app, ui, session-ui, desktop, plugin, llm, schema, protocol, client, codemode → 0 hits)
- [ ] `bun test` in `packages/opencode`, `packages/core`
- [ ] `bun run --cwd packages/opencode build` + binary smoke test
- [ ] `bun test` in `packages/app` (unit only) — e2e deferred (needs backend)
- [ ] Remove 26 `sst-env.d.ts` files, confirm nothing imports them

### Stage 4 — CI normalization
- [ ] `bun run typecheck` (root)
- [ ] `bun turbo test` equivalent in a clean checkout (simulate CI): `GITHUB_ACTIONS=false bun turbo test`
- [ ] `bun run --cwd packages/opencode test:httpapi` (HTTP API coverage suite) if applicable
- [ ] App e2e smoke (if backend available): `bun --cwd packages/app test:e2e:local`

### Stage 5 — Docs / rebrand
- [ ] `bun run typecheck` (root) — docs changes must not touch code
- [ ] README/CONTEXT/AGENTS render and reflect Athena
- [ ] Full re-run of §5.1 #2–#19 (final acceptance)

## 5.3 Application-launch smoke test (each stage)

| Test | Command | Pass criteria |
|---|---|---|
| TUI boots | `bun run --cwd packages/opencode src/index.ts tui --help` | exits 0, no import errors |
| Run mode | `bun run --cwd packages/opencode src/index.ts run "say hi" --model <any-catalog-model> --dry-run` (or offline) | streams without crash; graceful offline error |
| Serve | `bun run --cwd packages/opencode src/index.ts serve --port 4096` (timeout 10s) | HTTP server starts, `/health` responds |
| Desktop dev | `bun --cwd packages/desktop dev` (if desktop kept) | Electron window loads |
| Web dev | `bun --cwd packages/app dev` | Vite serves, app loads |

## 5.4 Rollback procedure

1. Single-commit repo → **create a tag/branch before cleanup**: `git tag athena-pre-cleanup` or `git stash`/`git checkout`.
2. On failure: `git checkout -- .` (restores working tree) or `git reset --hard <tag>`.
3. Re-run §5.1 #2–#3 to confirm rollback restored baseline.
4. Do NOT proceed to the next stage until the failing gate passes.

## 5.5 Known test caveats (documented, not failures)

- Tests **cannot run from repo root** (guard `do-not-run-tests-from-root`); always run from package dirs.
- `app` e2e requires a live backend at `localhost:4096`; unit/browser tests do not.
- `packages/opencode` has `test:httpapi`, `bench:test`, `profile:test` suites — run `test:httpapi` for contract confidence.
- Removing `storybook` drops the `storybook.yml` CI check; stories remain in `ui/session-ui/app` src and are still typechecked.
