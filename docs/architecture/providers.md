# Provider Architecture

## Overview

The provider system abstracts LLM providers through a unified interface. OpenCode supports multiple providers including OpenAI, Anthropic, Google, Azure, AWS Bedrock, Cloudflare, GitHub Copilot, OpenRouter, X.AI, and more.

## Provider Types

### AI SDK Providers
Most providers use the AI SDK (`@ai-sdk/*` packages) which provides:
- Unified `LanguageModelV3` interface
- Streaming support
- Tool calling
- Structured output

### Provider Routing
The `Provider.Service` (`packages/opencode/src/provider/provider.ts`) handles:
- Model catalog management
- Provider resolution
- Authentication lookup
- Model selection

## Provider Configuration

Providers are configured in `opencode.json`:

```json
{
  "provider": {
    "openai": {
      "apiKey": "sk-...",
      "models": {
        "gpt-4": { "cost": { "input": 0.03, "output": 0.06 } }
      }
    }
  }
}
```

## Authentication

Auth credentials are managed by `Auth.Service` (`packages/opencode/src/auth/index.ts`):

### Auth Types
- **API Key**: `{ type: "api", key: string, metadata?: Record<string, string> }`
- **OAuth**: `{ type: "oauth", refresh: string, access: string, expires: number }`
- **Well-known**: `{ type: "wellknown", key: string, token: string }`

### Auth Storage
- Stored in `~/.config/opencode/auth.json`
- File permissions: `0o600` (owner read/write only)
- Environment variable override: `OPENCODE_AUTH_CONTENT`

## Provider Resolution Flow

```
1. Agent specifies model (provider/model)
2. Provider.Service.getProvider(providerID) resolves the provider
3. Provider.Service.getLanguage(model) gets the AI SDK language model
4. Auth.Service.get(providerID) gets authentication
5. ProviderTransform prepares the request
6. LLM.Service.stream() initiates the stream
```

## Native vs AI SDK Runtime

OpenCode supports two LLM execution runtimes:

### AI SDK Runtime (Default)
- Uses `streamText()` from the AI SDK
- Provider execution and tool dispatch handled by AI SDK
- Broader provider support
- Falls back to this if native runtime is unavailable

### Native Runtime (Experimental)
- Uses `@opencode-ai/llm` directly
- Opt-in via `experimentalNativeLlm` flag
- Returns `LLMEvent` stream directly
- Better performance for supported providers

The runtime selection happens in `LLM.Service` (`packages/opencode/src/session/llm.ts`).

## Provider Transform

`ProviderTransform` (`packages/opencode/src/provider/transform.ts`) handles:
- Message format transformation
- Tool schema transformation
- Provider-specific options
- Cost calculation

## Key Files

| File | Purpose |
|------|---------|
| `packages/opencode/src/provider/provider.ts` | Provider service and model resolution |
| `packages/opencode/src/provider/auth.ts` | Auth methods and prompts |
| `packages/opencode/src/provider/transform.ts` | Request transformation |
| `packages/opencode/src/provider/error.ts` | Provider error types |
| `packages/llm/src/index.ts` | LLM package public API |
| `packages/llm/src/llm.ts` | LLM request/response types |
| `packages/llm/src/tool.ts` | Tool schema definitions |
| `packages/llm/src/route/client.ts` | LLM client for native runtime |