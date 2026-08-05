# Architecture Diagrams

## System Architecture

```mermaid
graph TD
    subgraph "Presentation Layer"
        CLI["CLI (yargs)"]
        TUI["TUI (OpenTUI)"]
        WEB["Web App"]
        DESKTOP["Desktop App"]
        SDK["SDK Client"]
    end

    subgraph "Application Layer"
        RUN["RunCommand"]
        PROC["SessionProcessor"]
        LLM_SVC["LLM Service"]
        SERVER["HTTP Server"]
    end

    subgraph "Core Runtime"
        SESSION["Session Service"]
        PROVIDER["Provider Service"]
        TOOL_REG["Tool Registry"]
        AGENT["Agent Service"]
        CONFIG["Config Service"]
        AUTH["Auth Service"]
        PERM["Permission Service"]
        PLUGIN["Plugin Service"]
        EVENT["EventV2Bridge"]
    end

    subgraph "Infrastructure"
        DB[(SQLite)]
        FS[(File System)]
        HTTP[(HTTP Client)]
        WS[(WebSocket)]
    end

    CLI --> RUN
    TUI --> RUN
    SDK --> SERVER
    WEB --> SERVER
    DESKTOP --> SERVER

    RUN --> PROC
    PROC --> LLM_SVC
    PROC --> SESSION
    PROC --> TOOL_REG
    PROC --> AGENT
    PROC --> CONFIG
    PROC --> AUTH
    PROC --> PERM
    PROC --> PLUGIN

    LLM_SVC --> PROVIDER
    LLM_SVC --> EVENT
    LLM_SVC --> DB

    TOOL_REG --> PERM
    TOOL_REG --> PLUGIN
    TOOL_REG --> DB

    SERVER --> SESSION
    SERVER --> LLM_SVC
    SERVER --> TOOL_REG

    SESSION --> DB
    CONFIG --> FS
    AUTH --> FS
    PLUGIN --> FS
```

## Startup Sequence

```mermaid
sequenceDiagram
    participant User
    participant CLI as CLI (yargs)
    participant Run as RunCommand
    participant Config as Config Service
    participant Plugin as Plugin Service
    participant Session as Session Service
    participant LLM as LLM Service
    participant Provider as Provider Service
    participant Tool as Tool Registry
    participant Agent as Agent Service
    participant Auth as Auth Service
    participant Perm as Permission Service

    User->>CLI: Runs `opencode run [message]`
    CLI->>CLI: Parse arguments, set env vars
    CLI->>Run: Execute handler
    Run->>Config: Load configuration
    Run->>Plugin: Load plugins
    Run->>Auth: Initialize auth
    Run->>Provider: Initialize provider catalog
    Run->>Agent: Initialize agent service
    Run->>Tool: Initialize tool registry
    Run->>Perm: Initialize permission service
    Run->>Session: Create or resume session
    Run->>LLM: Stream LLM request
    LLM->>Provider: Get language model
    Provider->>Auth: Get credentials
    Auth-->>Provider: Auth info
    Provider-->>LLM: Language model
    LLM->>Tool: Execute tools as needed
    Tool->>Perm: Check permissions
    Perm-->>Tool: Allow/deny
    Tool-->>LLM: Tool results
    LLM-->>Session: Stream events
    Session-->>Run: Session updates
    Run-->>CLI: Output results
    CLI-->>User: Display output
```

## Runtime Flow

```mermaid
flowchart TD
    A["User Input"] --> B["CLI Parsing"]
    B --> C["Config Loading"]
    C --> D["Plugin Initialization"]
    D --> E["Session Creation/Resumption"]
    E --> F["Agent Resolution"]
    F --> G["Model Resolution"]
    G --> H["System Prompt Construction"]
    H --> I["Conversation History"]
    I --> J["Tool Definitions"]
    J --> K["LLM Stream Initiation"]
    K --> L{"LLM Produces Tool Call?"}
    L -->|Yes| M["Tool Execution"]
    M --> N["Permission Check"]
    N --> O{"Approved?"}
    O -->|Yes| P["Execute Tool"]
    O -->|No| Q["Reject/Deny"]
    P --> R["Tool Result"]
    R --> K
    L -->|No| S["Final Response"]
    S --> T["Session Update"]
    T --> U["Output Display"]
    U --> V["Session Persistence"]
```

## Module Relationships

```mermaid
graph TD
    subgraph "packages/opencode"
        OPENCODE["Main CLI"]
        CONFIG["Config"]
        SESSION["Session"]
        LLM["LLM"]
        TOOLS["Tools"]
        AGENT["Agent"]
        PROVIDER["Provider"]
        AUTH["Auth"]
        PERM["Permission"]
        PLUGIN["Plugin"]
        MCP["MCP"]
        LSP["LSP"]
    end

    subgraph "packages/core"
        CORE["Core Runtime"]
        DB["Database"]
        EVENT["Event"]
        SCHEMA["Schema"]
        SESSION_V2["Session V2"]
        EXEC["Session Execution"]
        RUNNER["Session Runner"]
    end

    subgraph "packages/llm"
        LLM_PKG["LLM Package"]
        PROVIDERS["Providers"]
        PROTOCOLS["Protocols"]
        ROUTE["Route"]
    end

    subgraph "packages/server"
        SERVER["HTTP Server"]
        HANDLERS["Handlers"]
    end

    subgraph "packages/plugin"
        PLUGIN_PKG["Plugin Package"]
    end

    subgraph "packages/schema"
        SCHEMA_PKG["Schema Package"]
    end

    OPENCODE --> CONFIG
    OPENCODE --> SESSION
    OPENCODE --> LLM
    OPENCODE --> TOOLS
    OPENCODE --> AGENT
    OPENCODE --> PROVIDER
    OPENCODE --> AUTH
    OPENCODE --> PERM
    OPENCODE --> PLUGIN
    OPENCODE --> MCP
    OPENCODE --> LSP
    OPENCODE --> SERVER

    CONFIG --> CORE
    SESSION --> CORE
    LLM --> CORE
    TOOLS --> CORE
    AGENT --> CORE
    PROVIDER --> CORE
    AUTH --> CORE
    PERM --> CORE
    PLUGIN --> CORE
    MCP --> CORE
    LSP --> CORE

    LLM --> LLM_PKG
    LLM_PKG --> PROVIDERS
    LLM_PKG --> PROTOCOLS
    LLM_PKG --> ROUTE

    SERVER --> HANDLERS
    HANDLERS --> SESSION_V2
    HANDLERS --> EXEC
    HANDLERS --> RUNNER

    PLUGIN_PKG --> SCHEMA_PKG
```

## Provider Selection Flow

```mermaid
flowchart TD
    A["Agent specifies model"] --> B{"Model in catalog?"}
    B -->|Yes| C["Get provider from catalog"]
    B -->|No| D["Use default model"]
    C --> E["Get auth credentials"]
    D --> E
    E --> F{"Auth available?"}
    F -->|Yes| G["Resolve language model"]
    F -->|No| H["Use public API key"]
    G --> I{"Provider type?"}
    H --> I
    I -->|OpenAI| J["Use OpenAI Responses API"]
    I -->|Anthropic| K["Use Anthropic Messages API"]
    I -->|OpenAI Compatible| L["Use OpenAI Compatible Chat"]
    I -->|Other| M["Use AI SDK default"]
    J --> N["Stream response"]
    K --> N
    L --> N
    M --> N
```

## Tool Execution Flow

```mermaid
sequenceDiagram
    participant LLM as LLM
    participant Proc as SessionProcessor
    participant Reg as ToolRegistry
    participant Perm as Permission
    participant Tool as Tool
    participant Plugin as Plugin

    LLM->>Proc: Tool call event
    Proc->>Reg: Resolve tool definition
    Reg-->>Proc: Tool definition
    Proc->>Perm: Check permission
    Perm-->>Proc: Allow/Deny/Ask
    alt Permission denied
        Proc-->>LLM: Error response
    else Permission asked
        Proc->>Plugin: Trigger permission prompt
        Plugin-->>User: Show permission prompt
        User-->>Plugin: Approve/Deny
        Plugin-->>Proc: Permission result
        alt Approved
            Proc->>Tool: Execute tool
            Tool-->>Proc: Tool result
            Proc-->>LLM: Tool result
        else Denied
            Proc-->>LLM: Permission denied
        end
    else Permission allowed
        Proc->>Tool: Execute tool
        Tool-->>Proc: Tool result
        Proc-->>LLM: Tool result
    end
```

## Session Lifecycle

```mermaid
stateDiagram-v2
    [*] --> Created : Session.create()
    Created --> Active : First prompt
    Active --> Active : LLM stream + tool calls
    Active --> Compacting : Context exceeds limit
    Compacting --> Active : Compaction complete
    Active --> Idle : LLM response complete
    Idle --> Active : New prompt
    Active --> Forked : Session.fork()
    Active --> Archived : Session.setArchived()
    Archived --> [*] : Cleanup
    Forked --> Active : New session starts
    Created --> [*] : Session.remove()
```

## Configuration Flow

```mermaid
flowchart TD
    A["Start"] --> B["Load env vars"]
    B --> C["Load opencode.json"]
    C --> D["Load ~/.config/opencode/config.json"]
    D --> E["Apply defaults"]
    E --> F["Normalize config"]
    F --> G["Substitute variables"]
    G --> H["Validate config"]
    H --> I["Merge configs"]
    I --> J["Config ready"]
```

## Plugin Loading Flow

```mermaid
flowchart TD
    A["Start"] --> B["Load internal plugins"]
    B --> C["Load config plugins"]
    C --> D["Resolve plugin spec"]
    D --> E{"Plugin source?"}
    E -->|npm| F["Install plugin"]
    E -->|file| G["Read plugin file"]
    F --> H["Import plugin module"]
    G --> H
    H --> I{"Plugin type?"}
    I -->|server| J["Register server hooks"]
    I -->|tui| K["Register TUI hooks"]
    I -->|both| L["Register both hooks"]
    J --> M["Plugin loaded"]
    K --> M
    L --> M
```