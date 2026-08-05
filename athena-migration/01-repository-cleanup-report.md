# 1. Repository Cleanup Report — OpenCode → Athena

**Status:** FOR APPROVAL — no files deleted.
**Repo analyzed:** `/Users/pushp/Desktop/opencode` (single commit `e637f2d`, 6,417 files, ~280 MB, `packages/` = 127 MB).
**Scope:** Every top-level directory and every workspace package, plus notable individual files.

## Classification Legend

| Category | Meaning |
|---|---|
| **A** | Runtime Critical — required for compile, runtime, DI, execution, configuration, providers, tools, sessions, build system. Never delete. |
| **B** | Development Critical — package mgmt, lint, format, test, CI, release automation. Keep unless intentionally replaced. |
| **C** | Documentation — READMEs, CONTRIBUTING, SECURITY, architecture/spec/migration docs. Removable after confirmation. |
| **D** | Community / optional product surface — `.github` community bits, marketing, standalone products, commercial SaaS. |
| **E** | Demonstration — examples, demos, samples, unused fixtures, media. |
| **F** | Dead Code — unused packages/modules/assets/config with evidence. |

---

## Part 1 — Root Level

| Path | Purpose | Class | Dependents | Safe to Remove | Reason / Risk | Athena Action |
|---|---|---|---|---|---|---|
| `package.json` | Workspace root, scripts, catalog, patches | A | all | NO | Root of build system | KEEP (rename to athena) |
| `bun.lock` | Lockfile (852 KB) | A | install | NO | Regenerated on install | KEEP (regenerate) |
| `bunfig.toml` | Bun config, version pin, test guard | B | install/test | NO | Package manager config | KEEP |
| `turbo.json` | Task orchestration (typecheck/build/test) | A | CI, all packages | NO | Build system | KEEP (extend for Athena) |
| `tsconfig.json` | TS config | B | typecheck | NO | Needed for tsgo | KEEP |
| `.oxlintrc.json` | Lint config | B | lint | NO | Lint. NOTE: malformed (duplicate `options` keys) — repair | KEEP (fix) |
| `.prettierignore` | Format exclusions | B | format | NO | Formatting | KEEP |
| `.editorconfig` | Editor basics | B | all | NO | Trivial, keep | KEEP |
| `.gitattributes` | linguist-generated marks | B | git | NO | Repo hygiene | KEEP |
| `.gitignore` | Ignore rules | B | git | NO | Repo hygiene | KEEP (trim for Athena) |
| `.gitleaksignore` | Secret-scan ignore list | B | CI | NO | Security scanning | KEEP |
| `.dockerignore` | Container exclusions | B | containers | NO (only if containers kept) | Used by container builds | KEEP |
| `.husky/` | pre-push hook (typecheck) | B | git | NO | Dev gate | KEEP |
| `.vscode/` | Editor launch/settings examples | B | editors | optional | Only `.example` files | KEEP or trim |
| `.zed/` | Editor settings | B | editors | optional | Format-on-save | KEEP or trim |
| `patches/` | 15 bun `patchedDependencies` | A | `bun install` | NO | Required for install integrity | KEEP |
| `install` | Bash installer (13.7 KB) | A | release/CI | NO | Distribution entrypoint | KEEP (rebrand) |
| `flake.nix`, `flake.lock`, `nix/` | Nix packaging (opencode, desktop, node_modules) | B | release | optional | Distribution alternative; CI `nix-eval`/`nix-hashes` | KEEP or REMOVE |
| `AGENTS.md` | Agent/dev instructions | C | devs, AI agents | NO | Coding conventions; still valid | KEEP (rename refs) |
| `CONTRIBUTING.md` | Contributor guide | C | devs | optional | Outdated in places | KEEP (rewrite) |
| `CONTEXT.md` | Session-runtime architecture doc | C | devs | NO | Canonical V2 vocabulary for Athena | KEEP → Athena CONTEXT |
| `SECURITY.md` | Threat model | C | devs | optional | Policy doc | KEEP |
| `STATS.md` | Marketing download stats | D | `script/stats.ts`, `stats.yml` | **YES** | Generated marketing telemetry | REMOVE (w/ script + workflow) |
| `README.md` | Main README (English) | C | everyone | NO | Rebrand for Athena | KEEP (rewrite) |
| `README.{ar,bn,br,bs,da,de,es,fr,gr,it,ja,ko,no,pl,ru,th,tr,uk,vi,zh,zht}.md` | 23 localized READMEs | C | none | **YES** | Translation matrix; not maintained | REMOVE (single EN README) |
| `screenshot-uk.png` | Marketing screenshot (83 KB) | E | none | **YES** | Stale demo asset | REMOVE |
| `docs/` | Architecture + API docs | C | devs, docs site | NO (trim) | Valuable ADRs; some internal-only | KEEP → trim to Athena |
| `specs/` | V2 architecture specs | C | devs | NO | Source of truth for V2 runtime | KEEP |
| `script/` | Release/CI/chore scripts | B | CI, publish | partial | See `script/` section below | TRIM |
| `infra/` | SST production SaaS infra (Cloudflare/PlanetScale/AWS/Stripe/Honeycomb) | D | console/stats/enterprise/function/web | **YES** | Commercial backend not needed for Athena | REMOVE |
| `sst.config.ts` | SST app wiring | D | infra | **YES** | SaaS provisioning | REMOVE |
| `sst-env.d.ts` | Generated SST resource types | D | infra | **YES** | Regenerated; remove with infra | REMOVE |
| `github/` | Standalone "opencode GitHub Action" product | D | publish workflows | **YES** | Separate product, own publish | REMOVE |
| `sdks/` | VS Code extension (`sdks/vscode`) | D | `publish-vscode.yml`, `packages/identity` | **YES** | Separate product | REMOVE |
| `artifacts/` | Remotion marketing videos (~7.8 MB incl. rendered MP4s) | E | none | **YES** | Stale demo of "GLM-52" release | REMOVE |
| `perf/` | Test-suite benchmark log | F | none | **YES** | Finished investigation log | REMOVE |
| `.opencode/` | Repo-local agent config, tools, triage agents, themes | B | AI workflows, `.opencode/tool` | trim | Triage bots embed team roster; personal config | TRIM (keep jsonc/skills) |
| `.github/` | 25 workflows + templates + CODEOWNERS | B/D | CI/CD | partial | Keep test/typecheck; remove community bots (see Part 3) | TRIM |
| `.DS_Store` | macOS metadata | F | none | **YES** | Junk | REMOVE |

---

## Part 2 — Workspace Packages (35 packages)

### 2a. Runtime Critical — KEEP (Category A)

| Package | Purpose | Workspace deps | Tests | Risk if removed |
|---|---|---|---|---|
| `packages/schema` | Browser-safe wire/storage Effect contracts (the "leaf") | none | 6 | **HIGH** — everything above it breaks |
| `packages/protocol` | Effect `HttpApi` contract surface | schema | 1 | **HIGH** — server/client depend on it |
| `packages/core` | Domain/runtime engine: sessions V2, persistence, plugins, providers, tools, pty, permissions, config, filesystem | schema, llm, plugin, effect-drizzle-sqlite, sdk | 142 | **HIGH** — the product engine |
| `packages/server` | Concrete HTTP server binding protocol → core | core, protocol | 0 | **HIGH** — HTTP runtime |
| `packages/llm` | Native Effect LLM runtime (protocols/providers/tools) | schema | 30 | **HIGH** — model execution path |
| `packages/plugin` | Public plugin SDK (v1 hooks + v2 effect/promise) | sdk | 0 | **HIGH** — extension API |
| `packages/codemode` | Confined JS code execution (interpreter + stdlib + openapi adapter) | none | 7 | **MED** — used once by opencode tool |
| `packages/opencode` | Main app/CLI composition root (yargs, TUI boot, serve/run, generate) | core, server, sdk, tui, schema, protocol, plugin, llm, codemode | 247 | **HIGH** — the binary |
| `packages/tui` | Terminal UI library (Solid + opentui) | core, plugin, sdk, ui | 45 | **HIGH** — interactive surface |
| `packages/cli` | Companion CLI "lildax" (service mgmt + remote TUI) | tui, core, server, sdk/v2 | 1 | **LOW/MED** — separable product |
| `packages/client` | Generated Promise + Effect network clients (codegen target) | schema, protocol | 4 | **HIGH** — SDK emission + import-boundary tests |
| `packages/sdk` (`js/`) | Published SDK: legacy `createOpencode*` + regenerated `/v2` | none (cross-spawn only) | 1 | **HIGH** — consumed by plugin, tui, cli, opencode, app, session-ui |
| `packages/sdk-next` | Embedded in-process OpenCode host (Effect layer) | client, core, server | 2 | **MED** — future `@opencode-ai/sdk` |
| `packages/httpapi-codegen` | Code generator for client (Promise/Effect emitters) | none | 2 | **MED** — build-time only |
| `packages/effect-drizzle-sqlite` | Vendored Drizzle+Effect SQLite adapter | none | 1 | **HIGH** — core DB layer |
| `packages/script` | Release/version helper for build scripts | none | 0 | **MED** — build tooling |

### 2b. Product UI — KEEP (Category A, product surface)

| Package | Purpose | Workspace deps | Tests | Notes |
|---|---|---|---|---|
| `packages/ui` | SolidJS design system (v1+v2 components, theme, i18n, markdown) | none | 4 | Most-referenced package (9 dependents) |
| `packages/session-ui` | Session/message rendering domain lib | ui, sdk, core, client | 14 | Used by app, enterprise, storybook |
| `packages/app` | Main product web app (SolidStart-free Vite SPA) | ui, session-ui, core, sdk, client | 116 unit + 84 e2e | Depends on vendored client tarball (`vendor/*.tgz`) |
| `packages/desktop` | Electron shell (sidecar server, updater, WSL) | app, ui | 14 | Depends on `app/vite` + `opencode/dist/node` |

### 2c. Dead Code — REMOVE (Category F, with evidence)

| Package | Purpose | Evidence it is unused | Risk |
|---|---|---|---|
| `packages/effect-sqlite-node` | node:sqlite Effect client | Declared in `core/package.json` deps but **never imported in `core/src`** (grep: only self-reference in `src/index.ts`); `core` ships its own superset in `src/database/sqlite.node.ts` / `sqlite.bun.ts`. Nothing else references it. | LOW |
| `packages/identity` | Brand mark logos (no package.json) | **Zero references** repo-wide except `sdks/vscode/images` symlinks (also being removed). `docs/README.md` mislabels it as "Identity management". | LOW |
| `packages/core/src/effect/dfdf` | Stray file | Plain-text junk containing `"File to save in: ~/.local/share/opencode/worktree/.../src/effect/"` committed accidentally | LOW (trivial) |
| `packages/storybook` | Storybook host (90 stories, real) | Dev-only; nothing depends on it. Optional removal — it IS dev-critical tooling, keep if stories matter to Athena. | MED (dev loss only) |

### 2d. Commercial / Standalone Product Surface — REMOVE (Category D)

| Package | Purpose | Dependents | Why removable | Risk |
|---|---|---|---|---|
| `packages/console/*` (6 pkgs) | OpenCode Zen/Go/Black paid SaaS (SolidStart, Cloudflare, Stripe, PlanetScale, Upstash) | none in-product (only `@opencode-ai/ui` + each other) | Separate commercial product, deployed by `infra/console.ts` | LOW (infra removal bundled) |
| `packages/stats/*` (3 pkgs) | Internal analytics (AWS Athena/Firehose/S3 Tables, PlanetScale) | none in-product (only `@opencode-ai/ui`) | Separate product, deployed by `infra/lake.ts`+`stats.ts` | LOW |
| `packages/web` | Marketing + docs site + share pages (Astro/Starlight, 614 pages) | none (dev-depends on `opencode` CLI) | Standalone deployable | LOW |
| `packages/enterprise` | "Teams" SolidStart share/embed app + Hono API proxy | none | Paid product surface; uses ui/session-ui/core/sdk | LOW |
| `packages/function` | Cloudflare Worker: share sync + Feishu→Discord + GitHub token broker | none (deployed via `infra/app.ts`) | Infra glue for opencode.ai backend | LOW |
| `packages/slack` | Slack bot integration (Socket Mode) | none | Standalone integration | LOW |

### 2e. Supporting / Config (mixed)

| Package | Purpose | Class | Decision |
|---|---|---|---|
| `packages/script` | Version/release helper | B | KEEP (used by build scripts) |
| `.github/actions/setup-bun`, `setup-git-committer` | CI composite actions | B | KEEP if CI kept |
| `.opencode/` | Repo agent config | B | TRIM |

---

## Part 3 — `.github/` Workflow Classification

| Workflow | Purpose | Class | Action |
|---|---|---|---|
| `test.yml` | Unit + e2e test matrix | B | **KEEP** (core gate) |
| `typecheck.yml` | `bun typecheck` | B | **KEEP** |
| `generate.yml` | Regenerate SDK/codegen | B | **KEEP** |
| `publish.yml` | Main release pipeline | B | KEEP (retarget) |
| `containers.yml` | Docker builds | B | KEEP (retarget) |
| `beta.yml`, `publish-python-sdk.yml` | Beta automation / commented-out placeholder | D | REMOVE |
| `docs-update.yml`, `docs-locale-sync.yml` | Docs regen / locale sync (disabled) | D | REMOVE (if web removed) |
| `stats.yml` | STATS.md updater | D | REMOVE (w/ STATS.md) |
| `deploy.yml`, `nix-eval.yml`, `nix-hashes.yml` | SaaS deploy / Nix CI | D | REMOVE (if infra/nix removed) |
| `publish-vscode.yml`, `publish-github-action.yml`, `release-github-action.yml` | Standalone product publishes | D | REMOVE (sdks/github removed) |
| `storybook.yml` | Storybook build | B | KEEP or REMOVE (with storybook) |
| `opencode.yml`, `review.yml`, `triage.yml`, `close-issues.yml`, `close-prs.yml`, `compliance-close.yml`, `duplicate-issues.yml`, `pr-management.yml`, `pr-standards.yml`, `notify-discord.yml` | Community/automation bots | D | REMOVE (team/community automation) |
| `ISSUE_TEMPLATE/*`, `pull_request_template.md` | Community templates | D | KEEP or REMOVE |
| `CODEOWNERS`, `TEAM_MEMBERS` | Team ownership | D | REMOVE or replace |
| `actions/` | setup-bun, setup-git-committer | B | KEEP |

---

## Part 4 — Notable Individual Files (non-package)

| Path | Purpose | Class | Action |
|---|---|---|---|
| `packages/app/vendor/opencode-ai-client-1.17.13-v2.tgz` | Vendored prebuilt client tarball | A | **KEEP** — both `app` and `session-ui` pin to it via `file:` deps. Migration target: publish `@opencode-ai/client` and drop the tarball. |
| `packages/app/create-effect-simplification-spec.md`, `packages/app/V1_API_MIGRATION.md` | Internal specs | C | KEEP or move to `specs/` |
| `packages/app/public/` (favicon symlinks, `oc-theme-preload.js`) | App public assets | A | KEEP |
| `packages/opencode/migration/` | Drizzle migration snapshots | A | KEEP (DB migrations) |
| `packages/opencode/specs/`, `packages/session-ui/...` etc. | Package-local specs | C | Consolidate into root `specs/` optionally |
| `packages/stats/README.md`, `packages/stats/core/migrations/` | Stats docs/migrations | D | REMOVE with stats |
| `packages/storybook/debug-storybook.log` (29 KB) | Stale log | F | REMOVE |
| `packages/desktop/native:build` script | References non-existent `native/` dir | F | Fix or remove script |
| `patches/install-korean-ime-fix.sh` | Unreferenced stray script (not in patchedDependencies) | F | REMOVE |
| `packages/llm/example/`, `packages/sdk/js/example/`, `packages/effect-drizzle-sqlite/examples/`, `packages/plugin/src/example*.ts`, `packages/core/src/plugin/layer-map.example.ts` | Example/demo files | E | REMOVE (unreferenced) |
| `packages/schema/src/v1/` | Temporary V1 contract subtree | A (compat) | KEEP until legacy runtime retired (per AGENTS.md) |
| `packages/core/src/github-copilot/` | Self-described temporary provider shim | A | KEEP until V2 provider path replaces it (flagged in Athena report) |

---

## Part 5 — Summary Counts

| Action | Items |
|---|---|
| **KEEP** (runtime + UI + build) | 16 core pkgs + 4 UI pkgs + root build files + patches + specs |
| **REMOVE** (dead code) | `effect-sqlite-node`, `identity`, `core/src/effect/dfdf`, storybook (optional), 2 stray files |
| **REMOVE** (commercial) | `console/*` (6), `stats/*` (3), `web`, `enterprise`, `function`, `slack`, `infra/`, `sst.config.ts`, `sst-env.d.ts` |
| **REMOVE** (standalone products) | `github/`, `sdks/`, `nix/` (optional) |
| **REMOVE** (demo/community) | `artifacts/`, `perf/`, `STATS.md`, `screenshot-uk.png`, 23 localized READMEs, community workflows, `.DS_Store` files |

**Estimated cleaned repo size:** ~280 MB → ~95 MB (before `bun install`), 6,417 files → ~4,400 files.
