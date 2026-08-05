# 6. Remaining Repository Structure — after approved cleanup

**Status:** PROPOSED. Represents the repository after Phase 5 stages 1–4 (deletion) and before Athena renames (deliverable 7). No deletions have been performed.

## Retained top-level tree

```
opencode/  (root — will become athena/)
├── package.json                  # workspaces trimmed to retained packages
├── bun.lock
├── bunfig.toml
├── turbo.json                    # extended with tui#test, client#test, etc.
├── tsconfig.json
├── .oxlintrc.json                # malformed duplicate keys fixed
├── .prettierignore  .editorconfig  .gitattributes  .gitignore  .gitleaksignore  .dockerignore
├── .husky/                       # pre-push typecheck
├── .vscode/  .zed/               # editor configs (trim to taste)
├── patches/                      # 15 bun patchedDependencies (install-critical)
├── install                       # installer (rebranded)
├── AGENTS.md  CONTRIBUTING.md  CONTEXT.md  SECURITY.md  LICENSE
├── README.md                     # single English README (rewritten for Athena)
├── docs/                         # architecture docs (trimmed; internal/ + marketing removed)
├── specs/                        # V2 specs (kept as Athena design source)
├── .github/                      # trimmed: test, typecheck, generate, publish, containers (+actions/)
├── .opencode/                    # trimmed: opencode.jsonc, skills/; team bots removed
└── packages/
    ├── schema/                   # @opencode-ai/schema        (contract leaf)
    ├── protocol/                 # @opencode-ai/protocol      (HttpApi contract)
    ├── core/                     # @opencode-ai/core          (runtime engine)
    ├── server/                   # @opencode-ai/server        (HTTP runtime)
    ├── llm/                      # @opencode-ai/llm           (native model runtime)
    ├── plugin/                   # @opencode-ai/plugin        (plugin SDK)
    ├── codemode/                 # @opencode-ai/codemode      (confined code exec)
    ├── opencode/                 # opencode                    (CLI/app composition root)
    ├── tui/                      # @opencode-ai/tui           (terminal UI)
    ├── cli/                      # @opencode-ai/cli           (companion CLI)
    ├── client/                   # @opencode-ai/client        (generated clients)
    ├── sdk/                      # @opencode-ai/sdk           (legacy + v2 SDK)
    ├── sdk-next/                 # @opencode-ai/sdk-next      (embedded host)
    ├── httpapi-codegen/          # @opencode-ai/httpapi-codegen (codegen)
    ├── effect-drizzle-sqlite/    # @opencode-ai/effect-drizzle-sqlite (DB adapter)
    ├── script/                   # @opencode-ai/script        (release helper)
    ├── ui/                       # @opencode-ai/ui            (design system)
    ├── session-ui/               # @opencode-ai/session-ui    (session rendering)
    ├── app/                      # @opencode-ai/app           (product web app)
    └── desktop/                  # @opencode-ai/desktop       (Electron shell)
```

## Removed top-level items (summary — see deliverable 2 for per-item evidence)

| Removed | Count/type |
|---|---|
| `packages/console/*`, `packages/stats/*` | 9 packages (SaaS) |
| `packages/web`, `packages/enterprise`, `packages/function`, `packages/slack` | 4 packages (standalone surfaces) |
| `packages/storybook`, `packages/identity`, `packages/effect-sqlite-node` | 3 packages (dev-tool / dead) |
| `github/`, `sdks/`, `infra/`, `artifacts/`, `perf/`, `nix/` (optional) | 6 top-level dirs |
| `sst.config.ts`, `sst-env.d.ts` (26), `STATS.md`, `screenshot-uk.png`, localized READMEs (23), `.DS_Store` | misc files |
| `.github/workflows` | trimmed 25 → ~6 |
| in-file deletions | `core/src/effect/dfdf`, `layer-map.example.ts`, `plugin/src/example*.ts`, `patches/install-korean-ime-fix.sh`, `storybook/debug-storybook.log` |

## Retained counts

| Metric | Value |
|---|---|
| Workspace packages | 20 |
| src files (approx) | ~4,400 |
| Test files retained | ~620 across 17 packages |
| CI workflows | ~6 |
