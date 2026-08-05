# Tool Architecture

## Overview

Tools are functions that the LLM can call to perform actions. OpenCode provides a rich set of built-in tools and supports custom tools through plugins and configuration.

## Tool Definition

```typescript
interface Tool.Def {
  id: string
  description: string
  parameters: Schema.Decoder<unknown>
  jsonSchema?: JSONSchema7
  execute(args: Schema.Schema.Type<Parameters>, ctx: Tool.Context): Effect.Effect<Tool.ExecuteResult>
  formatValidationError?(error: unknown): string
}
```

### Tool Context

```typescript
interface Tool.Context {
  sessionID: SessionID
  abort: AbortSignal
  messageID: MessageID
  callID?: string
  extra?: Record<string, unknown>
  messages: SessionV1.WithParts[]
  metadata(input: { title?: string; metadata?: Record<string, any> }): Effect.Effect<void>
  ask(input: Omit<PermissionV1.Request, "id" | "sessionID" | "tool">): Effect.Effect<void>
}
```

## Built-in Tools

| Tool | Description |
|------|-------------|
| `read` | Read a file from the filesystem |
| `write` | Write a file to the filesystem |
| `edit` | Edit a file using string replacement |
| `apply_patch` | Apply a patch to a file |
| `shell` | Execute a shell command |
| `glob` | Find files matching a pattern |
| `grep` | Search for text in files |
| `task` | Create/manage tasks |
| `todo` | Manage todo items |
| `plan` | Plan mode entry/exit |
| `skill` | Execute a skill |
| `lsp` | Language Server Protocol operations |
| `webfetch` | Fetch content from a URL |
| `websearch` | Search the web |
| `mcp` | Model Context Protocol tools |
| `question` | Ask the user a question |

## Tool Registry

The `ToolRegistry.Service` (`packages/opencode/src/tool/registry.ts`) manages:

1. **Builtin tools**: All built-in tools are registered at startup
2. **Plugin tools**: Tools from loaded plugins
3. **Custom tools**: Tools from `{tool,tools}/*.ts` files in project directories
4. **MCP tools**: Tools from MCP servers (list_resources, read_resource)

### Tool Resolution

When the LLM calls a tool:
1. `ToolRegistry.tools()` returns the available tools for the model/agent
2. Each tool is wrapped with the AI SDK `tool()` function
3. The tool's `execute()` is called with the parsed arguments
4. The result is returned to the LLM

### Plugin Tool Bridging

Plugin tools are bridged from Promise-based to Effect-based:
- `EffectBridge.make()` creates a bridge for Promise-based operations
- Plugin `ask()` is wrapped to work with Effect's permission system
- Plugin `metadata()` is bridged to the session's metadata system

## Tool Execution Flow

```
1. LLM produces tool call
2. SessionProcessor receives the tool call event
3. ToolRegistry.resolve() finds the tool definition
4. Permission check via Permission.ask()
5. Tool.execute() runs with context
6. Result is returned to LLM
7. Tool output is truncated if too large
8. Session is updated with tool result
```

## Tool Output Truncation

Tool output is truncated to prevent context overflow:
- Default max: 2000 characters
- Protected tools (e.g., `skill`) are not truncated
- Truncated output includes a note about the truncation

## Custom Tools

### Project-Level Custom Tools
Place tool files in `{tool,tools}/*.ts` in your project directory:

```typescript
// tool/my-tool.ts
export const myTool = {
  description: "My custom tool",
  parameters: z.object({ input: z.string() }),
  execute: async (args, ctx) => {
    return { title: "Done", output: "Result" }
  },
}
```

### Plugin Tools
Plugins can register tools via the `tool` hook:

```typescript
tool: {
  myTool: {
    description: "My custom tool",
    parameters: { /* Zod schema */ },
    execute: async (args, ctx) => {
      return { title: "Done", output: "Result" }
    },
  },
}
```

## Key Files

| File | Purpose |
|------|---------|
| `packages/opencode/src/tool/registry.ts` | Tool registry and discovery |
| `packages/opencode/src/tool/tool.ts` | Tool type definitions |
| `packages/opencode/src/tool/shell.ts` | Shell tool implementation |
| `packages/opencode/src/tool/read.ts` | Read tool implementation |
| `packages/opencode/src/tool/write.ts` | Write tool implementation |
| `packages/opencode/src/tool/edit.ts` | Edit tool implementation |
| `packages/opencode/src/tool/skill.ts` | Skill tool implementation |
| `packages/opencode/src/tool/task.ts` | Task tool implementation |
| `packages/opencode/src/tool/plan.ts` | Plan tool implementation |
| `packages/opencode/src/tool/webfetch.ts` | Web fetch tool implementation |
| `packages/opencode/src/tool/websearch.ts` | Web search tool implementation |
| `packages/opencode/src/tool/apply_patch.ts` | Apply patch tool implementation |
| `packages/opencode/src/tool/lsp.ts` | LSP tool implementation |
| `packages/opencode/src/tool/mcp.ts` | MCP tool integration |