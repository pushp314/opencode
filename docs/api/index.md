# API Reference

## Core APIs

### Session API

#### Session.Service
The primary interface for session management.

**Methods:**
- `create(input?)` - Create a new session
- `get(id)` - Get session by ID
- `list(input?)` - List sessions
- `fork(input)` - Fork a session
- `remove(sessionID)` - Remove a session
- `messages(input)` - Get session messages
- `setTitle(input)` - Set session title
- `setArchived(input)` - Archive a session
- `setMetadata(input)` - Set session metadata
- `setAgentModel(input)` - Set agent and model
- `setPermission(input)` - Set permission ruleset
- `setRevert(input)` - Set revert point
- `setShare(input)` - Set share URL
- `setWorkspace(input)` - Set workspace
- `diff(sessionID)` - Get session diff
- `children(parentID)` - Get child sessions
- `updateMessage(msg)` - Update a message
- `removeMessage(input)` - Remove a message
- `removePart(input)` - Remove a message part
- `updatePart(part)` - Update a message part
- `updatePartDelta(input)` - Update a part delta
- `findMessage(sessionID, predicate)` - Find message by predicate

### LLM API

#### LLM.Service
The interface for LLM streaming.

**Methods:**
- `stream(input)` - Stream LLM response

**StreamInput:**
```typescript
interface StreamInput {
  user: SessionV1.User
  sessionID: string
  parentSessionID?: string
  model: Provider.Model
  agent: Agent.Info
  permission?: PermissionV1.Ruleset
  system: string[]
  messages: ModelMessage[]
  small?: boolean
  tools: Record<string, Tool>
  retries?: number
  toolChoice?: "auto" | "required" | "none"
}
```

### Tool Registry API

#### ToolRegistry.Service
The interface for tool discovery and registration.

**Methods:**
- `tools(model)` - Get tools for a model/agent
- `ids()` - Get all tool IDs
- `all()` - Get all tool definitions

### Agent API

#### Agent.Service
The interface for agent management.

**Methods:**
- `get(agent)` - Get agent by name
- `list()` - List all agents
- `defaultInfo()` - Get default agent info
- `defaultAgent()` - Get default agent name
- `generate(input)` - Generate agent from description

### Config API

#### Config.Service
The interface for configuration.

**Methods:**
- `get()` - Get current configuration
- `directories()` - Get config directories
- `waitForDependencies()` - Wait for config dependencies

### Auth API

#### Auth.Service
The interface for authentication.

**Methods:**
- `get(providerID)` - Get auth for a provider
- `all()` - Get all auth credentials
- `set(key, info)` - Set auth credential
- `remove(key)` - Remove auth credential

### Permission API

#### Permission.Service
The interface for permission management.

**Methods:**
- `ask(input)` - Ask for permission
- `reply(input)` - Reply to a permission request
- `list()` - List pending permissions

### Plugin API

#### Plugin.Service
The interface for plugin management.

**Methods:**
- `trigger(name, input, output)` - Trigger a hook
- `list()` - List loaded plugins
- `init()` - Initialize plugins

### SessionProcessor API

#### SessionProcessor.Service
The interface for session processing.

**Methods:**
- `create(input)` - Create a session processor handle

**Handle:**
```typescript
interface Handle {
  message: SessionV1.Assistant
  updateToolCall(toolCallID, update)
  completeToolCall(toolCallID, output)
  process(streamInput)
}
```