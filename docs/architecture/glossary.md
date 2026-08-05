# Glossary

## A

- **Agent**: A configured AI assistant with specific permissions, model, and prompt. Agents can be primary, subagent, or generated.
- **AI SDK**: The Vercel AI SDK (`ai` package) that provides a unified interface for LLM providers.
- **Apply Patch**: A tool that applies a diff/patch to a file.
- **Architecture Decision Record (ADR)**: A document that captures an important architectural decision along with its context and consequences.

## B

- **Background Job**: A long-running task associated with a session that runs in the background.
- **Builtin Tool**: A tool that is built into OpenCode (read, write, edit, shell, etc.).

## C

- **CLI**: Command Line Interface - the terminal-based interface for OpenCode.
- **Compaction**: The process of compressing conversation history to reduce token usage.
- **Context Epoch**: A baseline sequence number used to track which messages are part of the current context window.
- **Core Layer**: The `@opencode-ai/core` package that provides runtime infrastructure.
- **Cross-Spawn**: A utility for spawning child processes across platforms.
- **Cursor**: A pagination cursor for loading messages in batches.

## D

- **Database Layer**: The `@opencode-ai/core/database` module that provides SQLite persistence via Drizzle ORM.
- **Dependency Injection**: A design pattern where services receive their dependencies from an external source (Effect's Layer system).
- **Durable Session**: A session that persists across process restarts and can be resumed.
- **Durable Event Manifest**: A type-safe event schema definition for durable events.

## E

- **Effect**: A TypeScript functional programming framework used throughout OpenCode for dependency injection, error handling, and concurrency.
- **Effect Layer**: A composable unit of dependency injection in the Effect framework.
- **Effect Service**: A `Context.Service` that defines a service interface and its implementation.
- **EventV2Bridge**: The cross-service event communication bridge.
- **Execution**: The process of running a session's LLM stream and tool calls.

## F

- **Fiber**: A lightweight concurrent unit of execution in the Effect framework.
- **Fork**: Creating a new session from an existing session with its history.
- **FSUtil**: File system utility service for reading/writing files.

## G

- **Generate**: The agent generation feature that creates an agent from a description using the LLM.
- **Global Config**: Configuration stored in `~/.config/opencode/config.json`.
- **Glob**: A tool for finding files matching a pattern.

## H

- **Hook**: A plugin extension point that allows plugins to intercept and modify behavior.
- **HTTP API**: The Hono-based HTTP server for SDK and attach mode.
- **HTTP Client**: The Effect HTTP client for making HTTP requests.

## I

- **Instance**: A running OpenCode process with its own configuration and state.
- **Instance Context**: The runtime context for an OpenCode instance (directory, project, workspace).
- **Instance State**: Process-local mutable state managed by `InstanceState`.
- **Instruction Files**: AGENTS.md and CLAUDE.md files that provide project instructions to the LLM.

## L

- **Layer**: A composable unit of dependency injection in the Effect framework.
- **LLM**: Large Language Model - the AI model that generates responses.
- **LLMEvent**: An event emitted during LLM streaming (start, text, toolCall, etc.).
- **Location**: A logical location (directory + workspace) that determines which provider catalog and credentials to use.

## M

- **MCP**: Model Context Protocol - a protocol for external tool and resource servers.
- **Message**: A unit of conversation (user, assistant, system) with parts.
- **MessageV2**: The V2 message format with enhanced part types.
- **Model**: An LLM model instance with provider, ID, and configuration.
- **Model Catalog**: The dynamic model discovery system from `models.dev`.

## P

- **Part**: A component of a message (text, tool, reasoning, compaction, subtask).
- **Permission**: A rule that controls what actions the LLM can take.
- **Permission Ruleset**: An array of permission rules that evaluate against tool calls.
- **Plugin**: An extension mechanism for adding tools, auth, and hooks to OpenCode.
- **Provider**: An LLM provider (OpenAI, Anthropic, etc.) with its own API and authentication.
- **Provider Catalog**: The dynamic model discovery system.
- **Provider Transform**: Request transformation for different LLM providers.

## R

- **Read Tool**: A tool that reads a file from the filesystem.
- **Retry**: The mechanism for retrying failed LLM requests with exponential backoff.
- **Revert**: The ability to revert a session to a previous state.
- **Run Command**: The `opencode run` CLI command that starts a session.

## S

- **Schema**: Effect Schema for runtime data validation and type inference.
- **SDK**: The JavaScript/TypeScript client SDK for OpenCode.
- **Session**: A conversation unit with a unique ID, containing messages and metadata.
- **Session ID**: A unique identifier for a session (ULID-based).
- **Session Processor**: The orchestrator that manages the LLM stream and tool execution.
- **Session V2**: The current session architecture with durable prompt admission and process-local execution.
- **Shell Tool**: A tool that executes shell commands.
- **Skill**: A reusable prompt template that can be executed as a tool.
- **Snapshot**: A file system snapshot for tracking changes and enabling revert.
- **Stream**: Effect's streaming abstraction for LLM responses.
- **Subagent**: A lightweight agent with restricted permissions.
- **Summary**: A summary of a session's changes (additions, deletions, files).

## T

- **TUI**: Terminal User Interface - the interactive terminal interface for OpenCode.
- **Tool**: A function the LLM can call to perform actions.
- **Tool Context**: The context provided to tool execution (sessionID, abort, ask, metadata).
- **Tool Registry**: The service that manages tool discovery and registration.
- **Truncate**: The tool output truncation mechanism to prevent context overflow.

## W

- **Workspace**: A logical workspace that groups sessions and projects.
- **Write Tool**: A tool that writes content to a file.