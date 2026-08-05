# Internal Architecture

## Overview

This document covers internal implementation details of OpenCode that are important for maintainers and contributors.

## Effect Layer System

### How Layers Work

OpenCode uses Effect's Layer system for dependency injection. Each service is defined as a `Context.Service` and provided via a `Layer`.

```typescript
// Define the service interface
export interface Interface {
  readonly method: (input: Input) => Effect.Effect<Output>
}

// Create the service
export class Service extends Context.Service<Service, Interface>()("@opencode/ServiceName") {}

// Define the layer
const layer = Layer.effect(
  Service,
  Effect.gen(function* () {
    const dep1 = yield* Dep1.Service
    const dep2 = yield* Dep2.Service
    return Service.of({
      method: Effect.fn("method")(function* (input) {
        // Implementation
      }),
    })
  }),
)

// Export the layer
export const node = LayerNode.make({ service: Service, layer: layer, deps: [Dep1.node, Dep2.node] })
```

### Layer Composition

Layers are composed using `Layer.provide` and `Layer.mergeAll`:

```typescript
// Merge multiple layers
const appLayer = Layer.mergeAll(
  Database.node,
  Config.node,
  Auth.node,
  Provider.node,
  Agent.node,
  ToolRegistry.node,
  Plugin.node,
  Permission.node,
  Session.node,
  LLM.node,
)

// Provide a layer to an effect
const result = yield* SomeEffect.pipe(Effect.provide(appLayer))
```

## InstanceState

`InstanceState` is a generic state container for process-local mutable state:

```typescript
const state = yield* InstanceState.make<State>(
  Effect.fn("state")(function* (ctx) {
    return {
      runners: new Map<SessionID, Runner>(),
      claims: new Map<MessageID, Set<string>>(),
      pending: new Map<PermissionID, PendingEntry>(),
    }
  }),
)

// Access state
const current = yield* InstanceState.get(state)
// Modify state
yield* InstanceState.set(state, { ...current, runners: newRunners })
```

## EffectBridge

`EffectBridge` bridges Promise-based APIs to Effect:

```typescript
const bridge = yield* EffectBridge.make()

// Convert a Promise to an Effect
const result = yield* bridge.promise(somePromise)

// Fork an effect (fire-and-forget)
bridge.fork(someEffect)
```

## EventV2Bridge

The `EventV2Bridge` provides cross-service event communication:

```typescript
// Publish an event
yield* events.publish(Session.Event.Created, { sessionID, info })

// Subscribe to events
const unsub = await events.listen((event) => {
  if (event.type === Session.Event.Updated.type) {
    // Handle event
  }
  return Effect.void
})

// Wait for a specific event
const result = await bridge.promise(
  events.listen((event) => {
    if (event.type === Permission.Event.Replied.type) {
      return Effect.void
    }
  }),
)
```

## Database Patterns

### Schema Definitions

Database schemas use Drizzle ORM with snake_case field names:

```typescript
const SessionTable = sqliteTable("session", {
  id: text().primaryKey(),
  project_id: text().notNull(),
  workspace_id: text(),
  parent_id: text(),
  slug: text().notNull(),
  directory: text().notNull(),
  title: text().notNull(),
  // ... more fields
})
```

### Query Patterns

Queries use Drizzle's query builder with Effect wrapping:

```typescript
const row = yield* db
  .select()
  .from(SessionTable)
  .where(eq(SessionTable.id, sessionID))
  .get()
  .pipe(Effect.orDie)
```

### Transaction Patterns

Transactions are handled via Drizzle's transaction API:

```typescript
yield* db.transaction(async (tx) => {
  const result = await tx.insert(SessionTable).values(data).returning().get()
  // ... more operations
})
```

## Error Handling Patterns

### Typed Errors

Errors are defined as `Schema.TaggedErrorClass`:

```typescript
export class ModelUnavailableError extends Schema.TaggedErrorClass<ModelUnavailableError>()(
  "SessionRunnerModel.ModelUnavailableError",
  {
    providerID: ProviderV2.ID,
    modelID: ModelV2.ID,
  },
)
```

### Error Propagation

Errors propagate through Effect's error channel:

```typescript
const result = yield* someEffect.pipe(
  Effect.orDie,  // Convert Option to Effect with error
  Effect.catchAll((error) => Effect.logError("Error occurred", { error })),
)
```

### Error Recovery

Retry logic uses Effect's retry primitives:

```typescript
const result = yield* someEffect.pipe(
  Effect.retry({
    schedule: Schedule.exponential("200ms"),
    while: (error) => error instanceof RetryableError,
  }),
)
```

## Runtime Flags

Feature flags are managed through `RuntimeFlags`:

```typescript
interface Info {
  experimentalNativeLlm: boolean
  experimentalBackgroundSubagents: boolean
  experimentalWebSockets: boolean
  experimentalCodeMode: boolean
  experimentalWorkspaces: boolean
  disableClaudeCodePrompt: boolean
  // ... more flags
}
```

Flags are checked at runtime to enable/disable features:

```typescript
if (flags.experimentalNativeLlm) {
  // Use native LLM runtime
} else {
  // Use AI SDK runtime
}
```

## Key Files

| File | Purpose |
|------|---------|
| `packages/core/src/effect/layer-node.ts` | Layer node base class |
| `packages/core/src/effect/app-node.ts` | App node for global services |
| `packages/core/src/effect/app-node-platform.ts` | Platform-specific app node |
| `packages/core/src/effect/app-node-builder.ts` | App node builder |
| `packages/opencode/src/effect/instance-state.ts` | Instance state container |
| `packages/opencode/src/effect/bridge.ts` | Effect bridge for Promise APIs |
| `packages/opencode/src/effect/runtime-flags.ts` | Runtime flags service |
| `packages/opencode/src/event-v2-bridge.ts` | Event bridge implementation |