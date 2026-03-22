# Go Chatbot (Fyne + LLM)

Starter desktop chatbot built with Go and Fyne, wired to an OpenAI-compatible Chat Completions API.

## Requirements

- Go 1.23+
- A valid API key for your model provider
- On Windows, a C toolchain for CGO (for example, MinGW-w64)

If you see an error like "build constraints exclude all Go files ... github.com/go-gl/gl", install MinGW-w64 and make sure `gcc` is on your `PATH`, then reopen your terminal.

## Setup

1. Copy environment template:

   ```powershell
   Copy-Item .env.example .env
   ```

2. Set env vars in your shell (or system env):

   ```powershell
   $env:LLM_API_KEY="your_api_key_here"
   $env:LLM_MODEL="gpt-4o-mini"
   $env:LLM_BASE_URL="https://api.openai.com/v1"
   $env:CHAT_SYSTEM_PROMPT="You are a helpful assistant for a desktop chat app."
   ```

   Notes:
   - `LLM_API_KEY` is required.
   - `LLM_MODEL` defaults to `gpt-4o-mini`.
   - `LLM_BASE_URL` defaults to `https://api.openai.com/v1`.
   - `CHAT_SYSTEM_PROMPT` is optional.

## Run

```powershell
go mod tidy
go run .
```

## Project Layout

- `main.go`: Fyne UI shell and event handlers
- `internal/config/config.go`: environment-driven app configuration
- `internal/chat/service.go`: chat orchestration and conversation state
- `internal/llm/client.go`: OpenAI-compatible HTTP client
- `.env.example`: environment variable template

## Quality Checks

Run unit tests:

```powershell
go test ./internal/...
```

Run lints (after installing golangci-lint):

```powershell
golangci-lint run
```

## Next Improvements

- Persist conversation history to disk.
- Stream tokens into the UI.
- Add provider profiles (OpenAI, Azure OpenAI, local Ollama endpoint).
