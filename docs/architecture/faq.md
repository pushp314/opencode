# Frequently Asked Questions

## General

### What is OpenCode?
OpenCode is an AI-powered development tool that provides an intelligent coding assistant with support for multiple LLM providers, tool execution, session management, and extensibility through plugins.

### What is the difference between `--mini` and normal mode?
`--mini` runs in split-footer interactive mode with a TUI, while normal mode streams output to the terminal. `--mini` is the recommended way to use OpenCode interactively.

### How do I attach to a running OpenCode server?
Use `opencode run --attach http://localhost:4096 "your message"`. This connects to a running OpenCode server and runs interactive mode against it.

### How are sessions persisted?
Sessions are stored in SQLite via Drizzle ORM in the `.opencode/` directory in your project or global data directory.

### Can I use custom tools?
Yes, via plugins or by placing tool files in `{tool,tools}/*.ts` in your project directory.

### How does the permission system work?
Permissions are evaluated against a ruleset. Each tool call checks the ruleset and either allows, denies, or asks the user for approval.

## Configuration

### Where is the configuration file?
- Project-level: `opencode.json` in the project root
- Global: `~/.config/opencode/config.json`
- Environment variables override both

### How do I configure a provider?
Add the provider configuration to `opencode.json`:
```json
{
  "provider": {
    "openai": {
      "apiKey": "sk-..."
    }
  }
}
```

### How do I set a default model?
Set the default model in the provider configuration or use the `--model` flag when running.

## Agents

### What is the difference between a primary agent and a subagent?
A primary agent has full access (subject to permission rules). A subagent has restricted permissions and is typically used for sub-tasks.

### How do I create a custom agent?
Add an agent definition to `opencode.json`:
```json
{
  "agent": {
    "myagent": {
      "description": "My custom agent",
      "mode": "primary",
      "model": { "providerID": "anthropic", "modelID": "claude-3-5-sonnet" }
    }
  }
}
```

### Can I generate an agent from a description?
Yes, use `opencode run --agent <description>` to generate an agent from a natural language description.

## Tools

### What tools are available?
OpenCode provides built-in tools: read, write, edit, apply_patch, shell, glob, grep, task, todo, plan, skill, lsp, webfetch, websearch, and MCP tools.

### Can I disable a tool?
Yes, by configuring the permission ruleset to deny the tool.

### How are tool outputs truncated?
Tool outputs are truncated to 2000 characters by default. Protected tools (e.g., `skill`) are not truncated.

## Sessions

### How do I resume a session?
Use `--continue` to resume the last session or `--session <id>` to resume a specific session.

### How do I fork a session?
Use `--fork` with `--continue` or `--session` to create a fork of an existing session.

### Can I share a session?
Yes, use `--share` to share a session and get a shareable URL.

### How do I revert a session?
Sessions support reverting to a previous state. The revert captures a snapshot before the revert point and applies patches in reverse.

## Plugins

### How do I install a plugin?
Plugins are installed via npm and configured in `opencode.json`:
```json
{
  "plugin": {
    "my-plugin": {}
  }
}
```

### How do I write a plugin?
See the [Extension Guide](#extension-guide) in the main README.

### Can I write custom plugins?
Yes, plugins are TypeScript functions that return hooks for tool registration, auth, and lifecycle events.

## Performance

### How can I improve response times?
- Use a faster model
- Reduce conversation history (compaction helps)
- Use the native LLM runtime (experimental)
- Limit tool calls

### How does compaction work?
Compaction summarizes older messages to reduce token usage while preserving recent messages in full.

## Troubleshooting

### "Session not found"
Check that the session ID is correct and the session exists in the database.

### "Model unavailable"
Check that the provider is configured, the model exists in the catalog, and authentication credentials are valid.

### Permission denied
Check the permission ruleset for the agent. Use `--auto` to auto-approve permissions (dangerous).

### "Provider not configured"
Add the provider configuration to `opencode.json` or set the API key as an environment variable.

### Slow responses
Try a faster model, reduce conversation history, or enable the native LLM runtime.