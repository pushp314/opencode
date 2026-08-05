# 7. New Repository Tree — Athena (target state)

**Status:** PROPOSED. This is the *future* Athena layout after cleanup + renames. Renames map old → new; `MERGE`/`REPLACE` targets are the Athena components from deliverable 3 (integration points only — no Athena logic implemented yet).

## Rename map (old → new)

| Old | New | Action |
|---|---|---|
| `opencode/` (root) | `athena/` | RENAME |
| `@opencode-ai/schema` | `@athena/schema` | RENAME (contract leaf) |
| `@opencode-ai/protocol` | `@athena/protocol` | RENAME |
| `@opencode-ai/core` | `@athena/core` | RENAME (runtime engine) |
| `@opencode-ai/server` | `@athena/server` | RENAME |
| `@opencode-ai/llm` | `@athena/model` → seam to **Athena Model Orchestrator** | RENAME + REPLACE (Phase 4+, not now) |
| `@opencode-ai/plugin` | `@athena/plugin` | RENAME |
| `@opencode-ai/codemode` | `@athena/codemode` | RENAME |
| `packages/opencode` (`opencode`) | `packages/athena` (`athena`) | RENAME (composition root / CLI) |
| `@opencode-ai/tui` | `@athena/tui` → **Athena Terminal** | RENAME |
| `@opencode-ai/cli` | `@athena/cli` | RENAME |
| `@opencode-ai/client` | `@athena/client` | RENAME |
| `@opencode-ai/sdk` / `sdk-next` | `@athena/sdk` (sdk-next assumes the name) | MERGE (after legacy migration) |
| `@opencode-ai/httpapi-codegen` | `@athena/httpapi-codegen` | RENAME |
| `@opencode-ai/effect-drizzle-sqlite` | `@athena/db-sqlite` | RENAME |
| `@opencode-ai/script` | `@athena/script` | RENAME |
| `@opencode-ai/ui` | `@athena/ui` | RENAME |
| `@opencode-ai/session-ui` | `@athena/session-ui` | RENAME |
| `@opencode-ai/app` | `@athena/app` | RENAME |
| `@opencode-ai/desktop` | `@athena/desktop` | RENAME |
| `@opencode-ai/effect-sqlite-node` | — | REMOVED (dead) |
| `@opencode-ai/identity` | — | REMOVED |
| `@opencode-ai/console|stats|web|enterprise|function|slack|storybook` | — | REMOVED |

## Target tree

```
athena/
├── package.json                     # name: athena, workspaces: packages/*
├── bun.lock
├── bunfig.toml
├── turbo.json
├── tsconfig.json
├── .oxlintrc.json  .prettierignore  .editorconfig  .gitattributes  .gitignore  .gitleaksignore  .dockerignore
├── .husky/  .vscode/  .zed/
├── patches/
├── install                           # athena installer
├── AGENTS.md  CONTRIBUTING.md  CONTEXT.md  SECURITY.md  LICENSE  README.md
├── docs/                             # Athena architecture + ADRs (from docs/architecture)
├── specs/                            # V2 runtime specs (Athena design source)
├── .github/
│   └── workflows/  test.yml  typecheck.yml  generate.yml  publish.yml  containers.yml
│   └── actions/    setup-bun/  setup-git-committer/
└── packages/
    ├── schema/          @athena/schema           # contract leaf
    ├── protocol/        @athena/protocol         # HttpApi contract
    ├── core/            @athena/core             # runtime engine (sessions, system-context)
    │   └── src/
    │       ├── session/                          # → Athena Session Core (kept)
    │       ├── system-context/                   # → seam: Athena Context Engine
    │       ├── session/runner/                   # → seam: Athena Context Assembly
    │       ├── repository*.ts  ripgrep/  glob/   # → seam: Athena Repository Intelligence
    │       └── tool/  todowrite.ts  apply-patch.ts  # → seam: Athena Verification Pipeline
    ├── server/          @athena/server           # HTTP runtime
    ├── model/           @athena/model            # → Athena Model Orchestrator (Phase 4+)
    ├── plugin/          @athena/plugin           # plugin SDK (extension point for all seams)
    ├── codemode/        @athena/codemode         # confined code execution
    ├── athena/          athena                   # CLI composition root
    ├── tui/             @athena/tui              # → Athena Terminal
    ├── cli/             @athena/cli              # companion CLI
    ├── client/          @athena/client           # generated clients
    ├── sdk/             @athena/sdk              # (sdk-next merged in later)
    ├── httpapi-codegen/ @athena/httpapi-codegen  # codegen
    ├── db-sqlite/       @athena/db-sqlite        # Drizzle+Effect SQLite
    ├── script/          @athena/script           # release helper
    ├── ui/              @athena/ui               # design system
    ├── session-ui/      @athena/session-ui       # session rendering
    ├── app/             @athena/app              # product web app
    └── desktop/         @athena/desktop          # Electron shell
```

## Athena extension points (prepared by this cleanup, not implemented)

1. **Model Orchestrator seam** — `packages/model` + `core/src/catalog.ts` + `core/src/plugin/provider/`; single registry for all provider adapters (AI-SDK + native `llm` routes).
2. **Context Engine seam** — `core/src/system-context/` `SystemContext` algebra + `SystemContextRegistry`; plugin-defined Context Sources land here.
3. **Working Memory seam** — `core/src/session/` durable inbox/history + `snapshot.ts`; compaction + Context Epoch.
4. **Repository Intelligence seam** — `core/src/repository*.ts`, `ripgrep/`, `glob/`; index-backed search replacing ad-hoc tools.
5. **Context Assembly seam** — `core/src/session/runner/` single `llm.stream(request)` per turn.
6. **Verification Pipeline seam** — tool registry settlement + `todowrite`/`apply-patch` + file-mutation; model-executed gates at settlement boundary.

## Explicit non-goals for this phase

- No `@athena/*` renames executed (require separate approval — high blast radius across ~4,400 imports).
- No Athena component logic created.
- No provider catalog, session, or system-context changes.
