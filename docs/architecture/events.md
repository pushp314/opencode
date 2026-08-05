# Event System

## Overview

OpenCode uses an event system for cross-service communication. The event system is built on `EventV2Bridge` and uses the V2 event manifest for type safety.

## EventV2Bridge

The `EventV2Bridge` (`packages/opencode/src/event-v2-bridge.ts`) provides:
- Cross-service event publishing and subscribing
- Type-safe event handling
- Promise-based event waiting

### Key Methods
- `publish(event, data)` - Publish an event
- `listen(handler)` - Subscribe to events
- `promise(handler)` - Wait for a specific event
- `remove(handler)` - Unsubscribe from events

## Event Types

### Session Events
- `Session.Created` - Session created
- `Session.Updated` - Session updated
- `Session.Deleted` - Session deleted
- `Session.Diff` - Session diff computed
- `Session.Error` - Session error occurred
- `Session.Status` - Session status changed (idle/busy)

### Message Events
- `Message.Updated` - Message updated
- `Message.Removed` - Message removed
- `Part.Updated` - Message part updated
- `Part.Delta` - Message part delta update
- `Part.Removed` - Message part removed

### Permission Events
- `Permission.Asked` - Permission requested
- `Permission.Replied` - Permission response received

### Tool Events
- `Tool.Execute.Before` - Before tool execution
- `Tool.Execute.After` - After tool execution

### Plugin Events
- `Plugin.Hook` - Plugin hook triggered

## Event Manifest

The event manifest (`packages/opencode/src/event-manifest.ts`) defines the type-safe event schema. Each event type has a defined structure for its data payload.

## Event Flow

```
Service A publishes event
    ↓
EventV2Bridge distributes to subscribers
    ↓
Service B receives event
    ↓
Service B processes event
    ↓
Service B may publish response event
```

## Key Files

| File | Purpose |
|------|---------|
| `packages/opencode/src/event-v2-bridge.ts` | Event bridge implementation |
| `packages/opencode/src/event-manifest.ts` | Event manifest definitions |
| `packages/core/src/event.ts` | Core event types |
| `packages/schema/src/event-manifest.ts` | Schema event manifest |