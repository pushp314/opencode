# Configuration System

## Overview

OpenCode's configuration system loads settings from multiple sources with a clear precedence order. Configuration is managed through the `Config.Service` and supports both V1 (legacy) and V2 (current) formats.

## Configuration Sources (Precedence Order)

1. **Environment variables** - Highest precedence
2. **`opencode.json` in project root** - Project-specific config
3. **`~/.config/opencode/config.json`** - Global config
4. **Defaults** - Lowest precedence

## Configuration Structure

### V1 Config (Legacy)
```typescript
interface ConfigV1.Info {
  provider?: Record<string, ProviderConfig>
  agent?: Record<string, AgentConfig>
  tool?: Record<string, ToolConfig>
  permission?: PermissionConfig
  mcp?: Record<string, MCPConfig>
  experimental?: ExperimentalConfig
}
```

### V2 Config (Current)
The V2 config uses Effect Schema for validation and type inference.

## Config Loading

### Config.Parse
`Config.Parse` (`packages/opencode/src/config/parse.ts`) handles:
- Reading config files from disk
- Parsing JSON with JSONC support (comments, trailing commas)
- Normalizing loaded config (removing legacy TUI/theme/keybind fields)
- Substituting well-known remote config variables

### Config.Merge
Config merging uses `mergeDeep` from Remeda with custom array concatenation for instructions:
- Standard merge for most fields
- Array fields (like `instructions`) are concatenated and deduplicated
- Later sources override earlier sources

### Config Variable Substitution
Config values can reference environment variables and other config values:
```json
{
  "provider": {
    "openai": {
      "apiKey": "${OPENAI_API_KEY}"
    }
  }
}
```

## Key Configuration Sections

### Provider Config
```json
{
  "provider": {
    "openai": {
      "apiKey": "sk-...",
      "models": {
        "gpt-4": { "cost": { "input": 0.03, "output": 0.06 } }
      }
    }
  }
}
```

### Agent Config
```json
{
  "agent": {
    "reviewer": {
      "description": "Code reviewer",
      "mode": "subagent",
      "model": { "providerID": "anthropic", "modelID": "claude-3-5-sonnet" },
      "prompt": "Review the code changes...",
      "permission": [{ "permission": "read", "action": "allow", "pattern": "*" }]
    }
  }
}
```

### Permission Config
```json
{
  "permission": {
    "rules": [
      { "permission": "shell", "action": "allow", "pattern": "echo*" },
      { "permission": "shell", "action": "deny", "pattern": "rm*" }
    ]
  }
}
```

### MCP Config
```json
{
  "mcp": {
    "github": {
      "command": "npx",
      "args": ["-y", "@github/mcp-server"],
      "env": { "GITHUB_TOKEN": "${GITHUB_TOKEN}" }
    }
  }
}
```

## Config Directories

Config is loaded from multiple directories:
1. Project directory (`opencode.json`)
2. User home directory (`~/.config/opencode/config.json`)
3. Global config directory

The `config.directories()` method returns all directories to search for config files.

## Feature Flags

Feature flags are managed through `RuntimeFlags`:
- `experimentalNativeLlm` - Use native LLM runtime
- `experimentalBackgroundSubagents` - Enable background subagents
- `experimentalWebSockets` - Enable WebSockets for auth
- `experimentalCodeMode` - Enable code mode tools
- `experimentalWorkspaces` - Enable workspace support
- `disableClaudeCodePrompt` - Disable CLAUDE.md loading

## Key Files

| File | Purpose |
|------|---------|
| `packages/opencode/src/config/config.ts` | Config service and loading |
| `packages/opencode/src/config/parse.ts` | Config parsing |
| `packages/opencode/src/config/paths.ts` | Config path resolution |
| `packages/opencode/src/config/agent.ts` | Agent config schema |
| `packages/opencode/src/config/command.ts` | Command config schema |
| `packages/opencode/src/config/plugin.ts` | Plugin config schema |
| `packages/opencode/src/config/variable.ts` | Variable substitution |
| `packages/opencode/src/config/managed.ts` | Managed config |
| `packages/opencode/src/config/tui.ts` | TUI config |