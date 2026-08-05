# Cleanup Analysis

## Resource Cleanup

OpenCode uses Effect's Scope system for resource cleanup. All resources are acquired and released within scopes, ensuring proper cleanup even in error conditions.

## Cleanup Patterns

### 1. Effect.addFinalizer
Used to register cleanup actions that run when a scope is closed:

```typescript
yield* Effect.addFinalizer(Effect.gen(function* () {
  // Cleanup code here
}))
```

### 2. Effect.acquireRelease
Used for resource acquisition with automatic release:

```typescript
yield* Effect.acquireRelease(
  Effect.sync(() => new AbortController()),
  (ctrl) => Effect.sync(() => ctrl.abort()),
)
```

### 3. Latch
Used for synchronization in shell execution:

```typescript
const latch = new Latch()
// Signal when ready
latch.release()
// Wait for completion
await latch.await
```

## Session Cleanup

When a session is removed:
1. Background jobs for the session are cancelled
2. Child sessions are recursively removed
3. Session events are published
4. Event listeners are removed
5. Database rows are deleted

## Background Job Cleanup

Background jobs are managed by `BackgroundJob.Service`:
- Jobs are associated with sessions
- Jobs are cancelled when sessions are removed
- Job cancellation propagates to child processes

## Tool Execution Cleanup

Tool execution is scoped to the session fiber:
- Abort signals are propagated
- Child processes are terminated
- File handles are closed
- Memory is released when the fiber completes

## LLM Stream Cleanup

LLM streams are managed with proper cleanup:
- AbortController is created for each stream
- Stream is cancelled when the session ends
- Resources are released via Effect Scope

## Memory Management

- Effect fibers are lightweight and managed by the runtime
- No manual memory management needed
- Large tool outputs are truncated
- Session history is paginated
- Compaction reduces memory usage for long sessions

## Key Files

| File | Purpose |
|------|---------|
| `packages/opencode/src/session/run-state.ts` | Session run state and cleanup |
| `packages/opencode/src/background/job.ts` | Background job management |
| `packages/core/src/session/execution/local.ts` | V2 session execution cleanup |