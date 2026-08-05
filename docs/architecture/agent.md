# Agent Architecture

## Overview

Agents are the primary unit of execution in OpenCode. Each agent has a name, description, mode, model configuration, prompt, tools, and permissions. Agents can be primary agents, subagents, or generated agents.

## Agent Info Schema

```typescript
interface Agent.Info {
  name: string
  description?: string
  mode: "subagent" | "primary" | "all"
  native?: boolean
  hidden?: boolean
  topP?: number
  temperature?: number
  color?: string
  permission: PermissionV1.Ruleset
  model?: { modelID: ModelV2.ID; providerID: ProviderV2.ID }
  variant?: string
  prompt?: string
  options: Record<string, unknown>
  steps?: number
}
```

## Agent Modes

| Mode | Description | Permissions |
|------|-------------|-------------|
| `primary` | Main agent for user tasks | Full access (subject to ruleset) |
| `subagent` | Lightweight agent for sub-tasks | Restricted permissions |
| `all` | Agent that can be used in any context | Depends on context |

## Agent Resolution

1. **Config-defined agents**: Agents defined in `opencode.json`
2. **Default agent**: The `default` agent from config or a built-in default
3. **Generated agents**: AI-generated agents via `Agent.generate()`
4. **CLI-specified agents**: Agents specified via `--agent` flag

## Agent Service

The `Agent.Service` (`packages/opencode/src/agent/agent.ts`) provides:

- **`get(name)`** - Get an agent by name
- **`list()`** - List all configured agents
- **`defaultInfo()`** - Get the default agent configuration
- **`defaultAgent()`** - Get the default agent name
- **`generate(input)`** - Generate an agent from a description

Agent generation uses a dedicated prompt template (`generate.txt`) and the LLM to produce:
- `identifier` - Unique agent name
- `whenToUse` - Description of when to use this agent
- `systemPrompt` - The agent's system prompt

## Agent Prompt System

Agents have a prompt system that builds the system prompt:

1. **Environment info**: Working directory, git status, platform, date
2. **References**: Project references from `.opencode/references/`
3. **MCP tools**: Available MCP server tools
4. **Agent-specific prompt**: Custom prompt from agent config
5. **Instruction files**: AGENTS.md, CLAUDE.md from project directory

The system prompt is built by `SystemPrompt.Service` (`packages/opencode/src/session/system.ts`).

## Agent-Model Binding

Each agent can have a specific model bound to it:
- If no model is specified, the default model is used
- Model selection considers provider, model ID, and variant
- The model is resolved from the provider catalog

## Subagent Permissions

Subagents have restricted permissions:
- Default ruleset denies `question`, `plan_enter`, `plan_exit`
- Subagents can only use explicitly allowed tools
- Permission evaluation uses wildcard matching

## Key Files

| File | Purpose |
|------|---------|
| `packages/opencode/src/agent/agent.ts` | Agent service and schema |
| `packages/opencode/src/agent/generate.txt` | Agent generation prompt |
| `packages/opencode/src/session/system.ts` | System prompt builder |
| `packages/opencode/src/session/instruction.ts` | Instruction file loading |