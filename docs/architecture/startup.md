# Startup Sequence

## Program Entry Point

The entry point is `packages/opencode/src/index.ts`. It:
1. Parses CLI arguments using yargs
2. Sets environment variables (`OPENCODE`, `AGENT`, `OPENCODE_PID`)
3. Starts the heap profiler
4. Dispatches to the appropriate command handler

## Initialization Flow

### 1. CLI Parsing
```
yargs → args parsed → middleware sets env vars
```

### 2. Command Dispatch
Each command is an `effectCmd` that wraps a yargs command with Effect integration:
- `run` - Main execution command
- `mini` - Interactive split-footer mode
- `attach` - Connect to remote server
- `generate` - Code generation
- `agent` - Agent management
- `provider` - Provider configuration
- `mcp` - MCP server management
- `session` - Session management
- `stats` - Usage statistics
- `export`/`import` - Session data exchange
- `plug` - Plugin management
- `db` - Database operations
- `debug` - Debug commands
- `serve` - Start server
- `web` - Web interface
- `pr` - PR-related commands
- `github` - GitHub integration
- `account` - Account management
- `models` - Model management
- `upgrade`/`uninstall` - Maintenance commands

### 3. Effect Command Setup
Each `effectCmd` creates an Effect layer graph that includes:
- `InstanceRef` - Instance reference for the current project
- `RuntimeFlags` - Feature flags
- `Config.Service` - Configuration
- `Auth.Service` - Authentication
- `Provider.Service` - Provider catalog
- `Agent.Service` - Agent configuration
- `ToolRegistry.Service` - Tool registry
- `Plugin.Service` - Plugin management
- `Permission.Service` - Permission system
- `Session.Service` - Session management
- `LLM.Service` - LLM streaming
- `Database.Service` - Database connection
- `EventV2Bridge.Service` - Event bridge
- `MCP.Service` - MCP client
- `LSP.Service` - LSP integration

### 4. Session Creation/Resumption
- If `--session` or `--continue` is specified, resume existing session
- If `--fork` is specified, fork the session before continuing
- Otherwise, create a new session

### 5. LLM Stream Initiation
The LLM stream is initiated with:
- System prompt (built from environment, references, MCP tools)
- Conversation history
- Tool definitions
- Model configuration
- Provider authentication

### 6. Tool Execution Loop
As the LLM produces tool calls:
1. ToolRegistry resolves the tool definition
2. Permission check is performed
3. Tool.execute() runs with context
4. Result is returned to the LLM
5. Process continues until LLM produces final response

### 7. Shutdown
- Background jobs cancelled
- Session state persisted
- Resources released via Effect Scope
- Process exits with appropriate code

## Key Files

| File | Purpose |
|------|---------|
| `packages/opencode/src/index.ts` | CLI entry point |
| `packages/opencode/src/cli/cmd/run.ts` | Run command handler |
| `packages/opencode/src/cli/effect-cmd.ts` | Effect command wrapper |
| `packages/opencode/src/cli/ui.ts` | CLI UI utilities |
| `packages/opencode/src/cli/error.ts` | Error formatting |