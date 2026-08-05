# Architecture Decision Records (ADRs)

## ADR-001: Effect v4 for Runtime
**Status**: Accepted
**Date**: 2025
**Context**: Need a functional programming framework for type-safe dependency injection, error handling, and concurrency.
**Decision**: Use Effect v4 for all runtime services.
**Consequences**:
- All services are `Context.Service` based
- Services are composable via `Layer` graphs
- Errors are typed error classes, not exceptions
- Resource management via `Scope`
- Concurrency via fibers

## ADR-002: SQLite for Persistence
**Status**: Accepted
**Date**: 2025
**Context**: Need a lightweight, file-based database for session storage.
**Decision**: Use SQLite via Drizzle ORM.
**Consequences**:
- Single-file database in `.opencode/` directory
- No database server needed
- ACID transactions
- SQL query capabilities via Drizzle
- Portable across environments

## ADR-003: AI SDK for LLM Integration
**Status**: Accepted
**Date**: 2025
**Context**: Need unified interface for multiple LLM providers.
**Decision**: Use AI SDK v6 for provider abstraction and streaming.
**Consequences**:
- Provider-agnostic LLM calls
- Streaming support via `streamText()`
- Tool calling built-in
- Native runtime as opt-in for performance

## ADR-004: Monorepo Structure
**Status**: Accepted
**Date**: 2025
**Context**: Multiple packages with shared types and runtime.
**Decision**: Use Bun workspaces with Turborepo for build orchestration.
**Consequences**:
- Shared types across packages
- Independent versioning per package
- Coordinated builds via Turborepo
- Strict dependency rules between packages

## ADR-005: V2 Session Architecture
**Status**: Accepted
**Date**: 2025
**Context**: Need durable session management with process-local execution.
**Decision**: Separate prompt admission from model execution, use `SessionExecution` for process-local ownership.
**Consequences**:
- Durable sessions survive process restarts
- Process-local execution for performance
- Advisory wake pattern for efficiency
- `SessionRunCoordinator` for same-Session resume coalescing

## ADR-006: Schema-first Data Modeling
**Status**: Accepted
**Date**: 2025
**Context**: Need runtime data validation and type safety.
**Decision**: Use Effect Schema for all data structures.
**Consequences**:
- Runtime validation at boundaries
- Automatic JSON serialization/deserialization
- Type inference without manual annotations
- Self-documenting data structures

## ADR-007: Plugin Architecture
**Status**: Accepted
**Date**: 2025
**Context**: Need extensibility for tools, auth, and hooks.
**Decision**: Plugin system with hooks, tools, and auth providers.
**Consequences**:
- Plugins are functions returning hooks
- Tool registration via plugin `tool` hook
- Auth providers via plugin auth methods
- Lifecycle hooks for tool execution

## ADR-008: Hono for HTTP Server
**Status**: Accepted
**Date**: 2025
**Context**: Need a lightweight HTTP server for SDK and attach mode.
**Decision**: Use Hono for the HTTP server.
**Consequences**:
- Lightweight and fast
- OpenAPI support via `hono-openapi`
- Middleware composition
- WebSocket support

## ADR-009: OpenTUI for Terminal UI
**Status**: Accepted
**Date**: 2025
**Context**: Need a modern terminal UI for interactive mode.
**Decision**: Use OpenTUI (SolidJS-based) for the terminal UI.
**Consequences**:
- Reactive UI in the terminal
- Split-footer mode for interactive sessions
- SolidJS reactivity model
- Rich terminal rendering

## ADR-010: ULID for Session IDs
**Status**: Accepted
**Date**: 2025
**Context**: Need globally unique, sortable session identifiers.
**Decision**: Use ULID (Universally Unique Lexicographically Sortable Identifier).
**Consequences**:
- Sortable by time
- Globally unique
- URL-safe
- No collision risk