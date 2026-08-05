# Dependency Graph

## Package Dependencies

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                          │
│  opencode (CLI)  │  server  │  sdk  │  sdk-next  │  tui   │
├─────────────────────────────────────────────────────────────┤
│                    Core Layer                                 │
│  core (runtime)  │  llm  │  schema  │  protocol  │  plugin │
├─────────────────────────────────────────────────────────────┤
│                    Infrastructure Layer                        │
│  database  │  effect  │  ai-sdk  │  hono  │  drizzle-orm  │
└─────────────────────────────────────────────────────────────┘
```

## Dependency Rules

### Allowed Dependencies
```
Schema → Core → Protocol → Server
Client → Schema + Protocol
sdk-next → Client + Core + Server
opencode → Core + LLM + Schema + Protocol + Plugin + Server + SDK
```

### Forbidden Dependencies
- Client runtime code must NOT depend on Core or Server
- Protocol must NOT depend on Core
- Schema must NOT depend on anything

## Runtime Dependency Graph

```
index.ts (CLI entry)
  ├── yargs (CLI parsing)
  ├── RunCommand
  │   ├── Config.Service
  │   ├── Plugin.Service
  │   ├── Auth.Service
  │   ├── Provider.Service
  │   ├── Agent.Service
  │   ├── ToolRegistry.Service
  │   ├── Permission.Service
  │   ├── Session.Service
  │   ├── LLM.Service
  │   ├── Database.Service
  │   ├── EventV2Bridge.Service
  │   ├── MCP.Service
  │   └── LSP.Service
  ├── Server (HTTP server)
  │   ├── Hono router
  │   ├── Handlers (session, message, model, provider, etc.)
  │   └── Auth middleware
  └── TUI (terminal UI)
      ├── OpenTUI components
      ├── Session renderer
      ├── Footer (prompt, tools, permissions)
      └── Stream transport
```

## Service Dependency Graph

```
Session.Service
  ├── Database.Service
  ├── EventV2Bridge.Service
  ├── BackgroundJob.Service
  ├── RuntimeFlags.Service
  └── Snapshot.Service

LLM.Service
  ├── Auth.Service
  ├── Config.Service
  ├── Provider.Service
  ├── Plugin.Service
  ├── Permission.Service
  ├── EventV2Bridge.Service
  ├── LLMClient.Service
  ├── RuntimeFlags.Service
  └── EffectBridge.Service

ToolRegistry.Service
  ├── Config.Service
  ├── Plugin.Service
  ├── Agent.Service
  ├── Truncate.Service
  ├── RuntimeFlags.Service
  ├── MCP.Service
  ├── InvalidTool
  ├── TaskTool
  ├── ReadTool
  ├── QuestionTool
  ├── TodoWriteTool
  ├── LspTool
  ├── PlanExitTool
  ├── WebFetchTool
  ├── WebSearchTool
  ├── ShellTool
  ├── GlobTool
  ├── WriteTool
  ├── EditTool
  ├── GrepTool
  ├── ApplyPatchTool
  └── SkillTool

Permission.Service
  ├── EventV2Bridge.Service
  └── InstanceState

Plugin.Service
  ├── Config.Service
  ├── Auth.Service
  ├── Session.Service
  ├── PluginLoader
  └── internal plugins (Codex, Copilot, Modal, etc.)

Provider.Service
  ├── Config.Service
  ├── Auth.Service
  ├── Env.Service
  ├── Npm.Service
  ├── Hash.Service
  ├── FSUtil.Service
  └── Installation.Service

Config.Service
  ├── FSUtil.Service
  ├── HttpClient.Service
  ├── InstanceState
  ├── ConfigParse
  ├── ConfigPaths
  ├── ConfigAgent
  ├── ConfigCommand
  ├── ConfigPlugin
  ├── ConfigVariable
  └── ConfigManaged
```

## Import Graph (Key Files)

```
packages/opencode/src/index.ts
  → packages/opencode/src/cli/cmd/run.ts
  → packages/opencode/src/cli/effect-cmd.ts
  → packages/opencode/src/cli/ui.ts
  → packages/opencode/src/cli/error.ts

packages/opencode/src/cli/cmd/run.ts
  → packages/opencode/src/session/session.ts
  → packages/opencode/src/session/processor.ts
  → packages/opencode/src/session/llm.ts
  → packages/opencode/src/session/tools.ts
  → packages/opencode/src/agent/agent.ts
  → packages/opencode/src/provider/provider.ts
  → packages/opencode/src/config/config.ts
  → packages/opencode/src/plugin/index.ts
  → packages/opencode/src/permission/index.ts
  → packages/opencode/src/auth/index.ts
  → packages/opencode/src/server/server.ts
  → packages/opencode/src/mcp/index.ts

packages/core/src/session/execution/local.ts
  → packages/core/src/session/store.ts
  → packages/core/src/session/runner/index.ts
  → packages/core/src/session/runner/model.ts
  → packages/core/src/session/history.ts
  → packages/core/src/session/schema.ts
  → packages/core/src/database/database.ts
  → packages/core/src/location-services.ts
  → packages/core/src/effect/app-node.ts
```