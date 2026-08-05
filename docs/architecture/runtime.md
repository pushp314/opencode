# Runtime Architecture

## Runtime Lifecycle

The OpenCode runtime is built on the Effect v4 framework and follows a layered lifecycle:

### Phase 1: Layer Construction
All services are defined as `Context.Service` instances and composed into `Layer` graphs. The layer graph is constructed at startup and provides all dependencies.

### Phase 2: Service Acquisition
Services are acquired via `Effect.service` or `yield*` in Effect generators. The runtime resolves dependencies automatically through the Layer graph.

### Phase 3: Execution
The main execution flow runs within an Effect runtime:
- CLI commands execute as Effect programs
- LLM streams are processed as Effect Streams
- Tool execution is scoped to the session fiber
- Background jobs run in their own fibers

### Phase 4: Cleanup
Resources are released via Effect's Scope system:
- Finalizers run when scopes are closed
- Database connections are released
- File handles are closed
- Background jobs are cancelled

## Major Services

### Session.Service
Manages session CRUD operations, message history, and session metadata.

**Key methods:**
- `create()` - Create a new session
- `get()` - Get session by ID
- `list()` - List sessions
- `fork()` - Fork a session
- `messages()` - Get session messages
- `remove()` - Remove a session

### LLM.Service
Handles LLM streaming with support for both native and AI SDK runtimes.

**Key methods:**
- `stream()` - Stream LLM response

**Runtime selection:**
- Native runtime (`@opencode-ai/llm`) for providers that support it
- AI SDK runtime as fallback

### ToolRegistry.Service
Manages tool discovery, registration, and execution.

**Key methods:**
- `tools()` - Get tools for a model/agent
- `ids()` - Get all tool IDs
- `all()` - Get all tool definitions

### Agent.Service
Manages agent configuration and generation.

**Key methods:**
- `get()` - Get agent by name
- `list()` - List all agents
- `defaultInfo()` - Get default agent info
- `generate()` - Generate agent from description

### Config.Service
Loads and merges configuration from multiple sources.

**Key methods:**
- `get()` - Get current configuration
- `directories()` - Get config directories
- `waitForDependencies()` - Wait for config dependencies

### Auth.Service
Manages authentication credentials for providers.

**Key methods:**
- `get()` - Get auth for a provider
- `all()` - Get all auth credentials
- `set()` - Set auth credential
- `remove()` - Remove auth credential

### Permission.Service
Evaluates permissions and prompts for approval.

**Key methods:**
- `ask()` - Ask for permission
- `reply()` - Reply to a permission request
- `list()` - List pending permissions

### Plugin.Service
Manages plugin loading and hook execution.

**Key methods:**
- `trigger()` - Trigger a hook
- `list()` - List loaded plugins
- `init()` - Initialize plugins

### EventV2Bridge.Service
Cross-service event communication bridge.

**Key methods:**
- `publish()` - Publish an event
- `listen()` - Subscribe to events
- `remove()` - Remove event listeners

## Communication Flow

### Internal Communication
Services communicate through:
1. **Effect Context** - Direct service dependency injection
2. **EventV2Bridge** - Cross-service event publishing/subscribing
3. **Database** - Shared persistence layer
4. **InstanceState** - Process-local mutable state

### External Communication
- **CLI** - stdin/stdout for non-interactive mode
- **TUI** - Terminal UI for interactive mode
- **HTTP** - Hono server for SDK/attach mode
- **WebSocket** - For real-time event streaming

## Event Flow

```
User Input
    ↓
CLI Parsing (yargs)
    ↓
RunCommand Handler
    ↓
Config Loading
    ↓
Plugin Initialization
    ↓
Session Creation/Resumption
    ↓
LLM Stream Initiation
    ↓
Tool Execution (as needed)
    ↓
Permission Checks
    ↓
Result Streaming
    ↓
Session Persistence
    ↓
Output Display
```

## State Management

### Process-Local State
- `InstanceState` - Generic state container for process-local mutable state
- Used by Agent, ToolRegistry, Permission, SessionRunState, Instruction

### Database State
- SQLite via Drizzle ORM
- Sessions, messages, parts, compactions, context epochs

### Session State
- Session metadata (title, model, agent, permissions)
- Message history (ordered by sequence number)
- Tool call state (pending, running, completed, error)
- Compaction state (baseline sequence, tail turns)

## Concurrency Model

### Fiber-Based Concurrency
- Each LLM stream runs in its own Effect fiber
- Tool execution is scoped to the session fiber
- Background jobs run in separate fibers
- Session execution uses `SessionRunCoordinator` for coalescing

### Session Execution Ownership
- `SessionExecution` is process-global
- Session-ID based routing
- `SessionRunCoordinator` joins same-Session resumes
- Advisory wakes drain eligible durable inbox rows
- Different Sessions can run concurrently

### Concurrency Primitives
- `Effect.all()` - Parallel execution
- `Effect.fork()` - Fork a new fiber
- `Effect.race()` - Race between operations
- `Effect.scoped()` - Scoped resource management
- `Latch` - Synchronization primitive for shell execution