# Architecture Overview

OpenCode is an AI-powered development tool built as a TypeScript monorepo using the Effect v4 functional programming framework. It provides an intelligent coding assistant with support for multiple LLM providers, tool execution, session management, and extensibility through plugins.

## System Architecture

OpenCode follows a layered architecture with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                    Presentation Layer                        │
│  CLI (yargs)  │  TUI (OpenTUI)  │  Web App  │  Desktop App │
├─────────────────────────────────────────────────────────────┤
│                    Application Layer                         │
│  RunCommand  │  SessionProcessor  │  LLM Service  │  Server │
├─────────────────────────────────────────────────────────────┤
│                    Core Runtime Layer                        │
│  Effect DI  │  Database  │  Config  │  Auth  │  Permission │
├─────────────────────────────────────────────────────────────┤
│                    Infrastructure Layer                      │
│  SQLite  │  HTTP  │  WebSocket  │  File System  │  Process │
└─────────────────────────────────────────────────────────────┘
```

## Key Design Decisions

### 1. Effect v4 for Runtime
All services use Effect's `Context.Service` and `Layer` system for dependency injection. This provides:
- Type-safe service resolution
- Composable service graphs
- Built-in error handling
- Resource management (Scope)
- Concurrency primitives (Fiber)

### 2. Schema-first Data Modeling
Data structures are defined using Effect Schema for runtime validation and type inference. This ensures:
- Data integrity at boundaries
- Automatic JSON serialization/deserialization
- Type safety without manual type annotations

### 3. Stream-based LLM Communication
LLM responses are processed as Effect Streams, enabling:
- Real-time streaming to the UI
- Backpressure handling
- Composable stream transformations
- Resource-safe stream lifecycle

### 4. SQLite for Persistence
Sessions and their data are persisted using SQLite via Drizzle ORM. This provides:
- Zero-configuration persistence
- Portable single-file database
- ACID transactions
- SQL query capabilities

### 5. Plugin Architecture
Plugins extend OpenCode's capabilities through:
- Hook system for lifecycle events
- Tool registration
- Auth provider integration
- Custom command execution

## Package Dependency Rules

The architecture enforces strict dependency rules:

```
Schema → Core → Protocol → Server
Client → Schema + Protocol
sdk-next → Client + Core + Server
```

Client runtime code may depend on Schema and Protocol but never Core or Server.

## Module Categories

| Category | Description | Examples |
|----------|-------------|----------|
| Core | Runtime infrastructure | Database, Config, Auth, Permission |
| Infrastructure | Cross-cutting concerns | Logger, Tracing, HTTP Client |
| Domain | Business logic | Session, Agent, Provider, Tool |
| Application | User-facing features | CLI, TUI, Server, SDK |
| UI | Interface layer | OpenTUI, Web App, Desktop |
| Plugin | Extensibility | Plugin loader, hooks, tools |
| Provider | LLM integration | AI SDK adapters, protocols |
| Utility | Helper functions | FS, Path, String utils |

## Concurrency Model

OpenCode uses Effect's fiber-based concurrency:
- Each LLM stream runs in its own fiber
- Tool execution is scoped to the session fiber
- Background jobs use the BackgroundJob service
- Session execution is process-local with advisory wake coordination
- The `SessionRunCoordinator` coalesces prompt wakeups for the same session

## Ownership Model

- **Session**: Owned by the process that created it
- **SessionExecution**: Process-global, Session-ID based
- **SessionRunner**: Scoped to the session's Location
- **Tool execution**: Owned by the session fiber
- **Background jobs**: Owned by the session, cancelled on session removal