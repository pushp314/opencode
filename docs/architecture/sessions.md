# Session Architecture

## Overview

Sessions are the primary unit of conversation in OpenCode. Each session has a unique ID, contains messages and their parts, and tracks metadata like cost, tokens, and permissions.

## Session Schema

```typescript
interface Session.Info {
  id: SessionID
  slug: string
  projectID: ProjectV2.ID
  workspaceID?: WorkspaceV2.ID
  directory: string
  path?: string
  parentID?: SessionID
  title: string
  agent?: string
  model?: { id: ModelV2.ID; providerID: ProviderV2.ID; variant: string }
  version: string
  summary?: Summary
  cost?: number
  tokens?: Tokens
  share?: { url: string }
  metadata?: Record<string, any>
  revert?: Revert
  permission?: PermissionV1.Ruleset
  time: {
    created: number
    updated: number
    compacting?: number
    archived?: number
  }
}
```

## Session V2 Architecture

The current session architecture (V2) is based on the Effect framework and follows these principles:

### Durable Prompt Admission
- `SessionV2.prompt()` admits one durable `session_input` row before scheduling advisory `SessionExecution.wake(sessionID)`
- Unless `resume: false` requests admit-only behavior
- The serialized runner promotes admitted inputs into visible user messages at safe boundaries

### Process-Local Execution
- `SessionExecution` is process-global and Session-ID based
- Its local implementation owns the process-local Session coordinator
- Discovers placement through `SessionStore` plus `LocationServiceMap.get(session.location)` only when a drain starts
- No layer takes a Session ID directly

### Session Run Coordinator
- `SessionRunCoordinator` joins explicit same-Session resumes
- Coalesces prompt wakeups
- Allows different Sessions to run concurrently
- Advisory wakes drain eligible durable inbox rows only
- Post-crash continuation recovery requires separate design

### Drain Model
- A drain has no durable identity or transcript boundary
- Keeps local Session drains process-local until clustering is implemented

## Session Operations

### Creation
```typescript
Session.create({ title, agent, model, permission, workspaceID })
```
- Creates a new session with a unique ID (descending ULID)
- Sets slug, version, projectID, directory, path
- Initializes cost and tokens to zero
- Publishes `Session.Created` event

### Forking
```typescript
Session.fork({ sessionID, messageID })
```
- Creates a new session with the same history up to `messageID`
- Clones all messages and parts with new IDs
- Updates parentID reference
- Publishes `Session.Created` event for the forked session

### Message Management
- Messages are stored with sequence numbers for ordering
- Each message has an ID, role, parts, and metadata
- Parts can be text, tool, reasoning, compaction, or subtask
- Messages are paginated for efficient retrieval

### Compaction
- Automatic when context exceeds token limits
- Preserves recent turns (configurable)
- Summarizes older turns
- Creates compaction parts in the message history
- Protected tools (e.g., `skill`) are not truncated during compaction

### Revert
- Sessions support reverting to a previous state
- Captures a snapshot before the revert point
- Applies patches in reverse
- Computes diff of the reverted changes
- Stores revert metadata in the session

### Archiving
- Sessions can be archived (soft delete)
- Archived sessions are excluded from normal listings
- Archive timestamp is stored in session metadata

## Session History

Session history is loaded from the database:
1. `SessionHistory.load()` loads all messages for a session
2. Respects compaction boundaries (loads from compaction seq onward)
3. Respects context epoch baselines
4. Decodes messages from database rows

## Session Persistence

Sessions are persisted in SQLite via Drizzle ORM:
- `SessionTable` - Session metadata
- `SessionMessageTable` - Messages with parts
- `PartTable` - Individual message parts
- `SessionContextEpochTable` - Context epoch tracking

## Key Files

| File | Purpose |
|------|---------|
| `packages/opencode/src/session/session.ts` | Session service and schema |
| `packages/opencode/src/session/processor.ts` | Session processor (orchestrator) |
| `packages/opencode/src/session/llm.ts` | LLM streaming for sessions |
| `packages/opencode/src/session/tools.ts` | Tool resolution for sessions |
| `packages/opencode/src/session/compaction.ts` | Session compaction logic |
| `packages/opencode/src/session/message-v2.ts` | Message V2 handling |
| `packages/opencode/src/session/revert.ts` | Session revert logic |
| `packages/opencode/src/session/run-state.ts` | Session run state management |
| `packages/opencode/src/session/summary.ts` | Session summary generation |
| `packages/opencode/src/session/retry.ts` | Retry logic for LLM errors |
| `packages/core/src/session/store.ts` | V2 session store |
| `packages/core/src/session/execution/local.ts` | V2 session execution |
| `packages/core/src/session/runner/model.ts` | Model resolution for sessions |