<div align="center">

```
██████╗ ███╗   ███╗     ██████╗ ██████╗ ██████╗ ███████╗
╚════██╗████╗ ████║    ██╔════╝██╔═══██╗██╔══██╗██╔════╝
 █████╔╝██╔████╔██║    ██║     ██║   ██║██║  ██║█████╗  
██╔═══╝ ██║╚██╔╝██║    ██║     ██║   ██║██║  ██║██╔══╝  
███████╗██║ ╚═╝ ██║    ╚██████╗╚██████╔╝██████╔╝███████╗
╚══════╝╚═╝     ╚═╝     ╚═════╝ ╚═════╝ ╚═════╝ ╚══════╝
```

</div>

# 2M Code — Full Foundation Build

Build the **entire** project foundation so that any AI (including less powerful models) can pick up where this leaves off. Every file described in `agent.md` will be created with production-quality code, proper error handling, and comprehensive doc comments.

## Proposed Changes

The build follows the exact order from `agent.md` §Build Order, creating all ~30 files across Go and Python.

---

### Phase 1 — Dependency Files

#### [NEW] [go.mod](file:///c:/Users/pc/Desktop/2m-code/go.mod)
Go module definition with Cobra, go-sqlite3, yaml.v3, lipgloss.

#### [NEW] [go.sum](file:///c:/Users/pc/Desktop/2m-code/go.sum)
Auto-generated after `go mod tidy`.

#### [NEW] [requirements.txt](file:///c:/Users/pc/Desktop/2m-code/requirements.txt)
Python deps: anthropic, openai, google-generativeai, mistralai, fastapi, uvicorn, pydantic.

---

### Phase 2 — Python Agent Engine (Core)

#### [NEW] [server.py](file:///c:/Users/pc/Desktop/2m-code/agent_engine/server.py)
FastAPI app with `GET /health` and `POST /call`. Binds to `127.0.0.1:8765` (localhost only per security rules).

#### [NEW] [agent.py](file:///c:/Users/pc/Desktop/2m-code/agent_engine/agent.py)
Routes requests to correct provider. Handles tool definitions injection.

#### [NEW] [__init__.py](file:///c:/Users/pc/Desktop/2m-code/agent_engine/__init__.py)
Empty init for package.

#### [NEW] [providers/__init__.py](file:///c:/Users/pc/Desktop/2m-code/agent_engine/providers/__init__.py)
Provider registry — all 8 providers registered.

#### [NEW] [anthropic_provider.py](file:///c:/Users/pc/Desktop/2m-code/agent_engine/providers/anthropic_provider.py)
Anthropic SDK adapter — normalizes to `{content, tool_calls, input_tokens, output_tokens}`.

#### [NEW] [google_provider.py](file:///c:/Users/pc/Desktop/2m-code/agent_engine/providers/google_provider.py)
Google Gemini SDK adapter — same normalized shape.

#### [NEW] [openai_provider.py](file:///c:/Users/pc/Desktop/2m-code/agent_engine/providers/openai_provider.py)
OpenAI SDK adapter — same normalized shape.

#### [NEW] [mistral_provider.py](file:///c:/Users/pc/Desktop/2m-code/agent_engine/providers/mistral_provider.py)
Mistral SDK adapter — same normalized shape.

#### [NEW] [cohere_provider.py](file:///c:/Users/pc/Desktop/2m-code/agent_engine/providers/cohere_provider.py)
Cohere SDK adapter — same normalized shape.

#### [NEW] [groq_provider.py](file:///c:/Users/pc/Desktop/2m-code/agent_engine/providers/groq_provider.py)
Groq SDK adapter (OpenAI-compatible, ultra-fast LPU inference) — same normalized shape.

#### [NEW] [ollama_provider.py](file:///c:/Users/pc/Desktop/2m-code/agent_engine/providers/ollama_provider.py)
Ollama local inference adapter — same normalized shape. No API key needed.

#### [NEW] [openrouter_provider.py](file:///c:/Users/pc/Desktop/2m-code/agent_engine/providers/openrouter_provider.py)
OpenRouter unified API adapter (200+ models via OpenAI-compatible SDK) — same normalized shape.

#### [NEW] [tools/__init__.py](file:///c:/Users/pc/Desktop/2m-code/agent_engine/tools/__init__.py)
Tool definitions (bash, read_file, write_file) and `execute_tool()` function.

#### [NEW] [bash_tool.py](file:///c:/Users/pc/Desktop/2m-code/agent_engine/tools/bash_tool.py)
Bash execution with 30s timeout, stdout+stderr capture.

#### [NEW] [file_tool.py](file:///c:/Users/pc/Desktop/2m-code/agent_engine/tools/file_tool.py)
File read (max 100KB) and write operations with path validation.

#### [NEW] [web_tool.py](file:///c:/Users/pc/Desktop/2m-code/agent_engine/tools/web_tool.py)
Web fetch tool — GET URL, return text content (max 50KB).

---

### Phase 3 — Go Event Bus (SQLite)

#### [NEW] [schema.go](file:///c:/Users/pc/Desktop/2m-code/internal/bus/schema.go)
SQLite schema creation — `sessions` and `messages` tables with indexes. Uses parameterized queries throughout.

#### [NEW] [bus.go](file:///c:/Users/pc/Desktop/2m-code/internal/bus/bus.go)
`Post()` and `GetHistory()` methods. All SQL uses prepared statements (no string concatenation).

---

### Phase 4 — Team Config

#### [NEW] [team.go](file:///c:/Users/pc/Desktop/2m-code/internal/team/team.go)
Team/Agent/Workflow structs. YAML loader with search order: project-local → global → bundled.

#### [NEW] [config.go](file:///c:/Users/pc/Desktop/2m-code/internal/team/config.go)
Global config management (`~/.2mcode/config.yaml`). API key resolution from env vars.

---

### Phase 5 — Bridge (Go → Python IPC)

#### [NEW] [bridge.go](file:///c:/Users/pc/Desktop/2m-code/internal/bridge/bridge.go)
HTTP client posting to `http://127.0.0.1:8765/call`. 30s timeout. JSON request/response.

---

### Phase 6 — Orchestrator

#### [NEW] [orchestrator.go](file:///c:/Users/pc/Desktop/2m-code/internal/orchestrator/orchestrator.go)
Core engine: `RunTask()` with leader-first and round-robin orchestration. Tool-use loop.

#### [NEW] [scheduler.go](file:///c:/Users/pc/Desktop/2m-code/internal/orchestrator/scheduler.go)
Turn order calculation — determines agent sequence based on workflow config.

#### [NEW] [tools.go](file:///c:/Users/pc/Desktop/2m-code/internal/orchestrator/tools.go)
Go-side tool execution dispatcher — delegates to Python engine or handles locally.

---

### Phase 7 — CLI (Cobra)

#### [NEW] [renderer.go](file:///c:/Users/pc/Desktop/2m-code/internal/cli/renderer.go)
Lipgloss-based terminal rendering with color-coded agent badges, tool call display, completion summary.

#### [NEW] [root.go](file:///c:/Users/pc/Desktop/2m-code/internal/cli/root.go)
Cobra root command with version, help, persistent flags.

#### [NEW] [run.go](file:///c:/Users/pc/Desktop/2m-code/internal/cli/run.go)
`2m run <team> "<task>"` — one-shot task execution.

#### [NEW] [chat.go](file:///c:/Users/pc/Desktop/2m-code/internal/cli/chat.go)
`2m chat <team>` — interactive REPL.

#### [NEW] [team.go](file:///c:/Users/pc/Desktop/2m-code/internal/cli/team.go)
`2m team list` and `2m team show <name>` subcommands.

#### [NEW] [newteam.go](file:///c:/Users/pc/Desktop/2m-code/internal/cli/newteam.go)
`2m new-team` interactive wizard.

---

### Phase 8 — Main Entrypoint

#### [NEW] [main.go](file:///c:/Users/pc/Desktop/2m-code/cmd/2m/main.go)
Entrypoint: spawns Python agent engine, health-checks it, runs Cobra CLI, kills Python on exit.

---

### Phase 9 — Example Teams + Docs

#### [NEW] [fullstack.yaml](file:///c:/Users/pc/Desktop/2m-code/config/teams/fullstack.yaml)
Aria (Tech Lead/Anthropic) + Dev (Engineer/Gemini) + Quinn (QA/OpenAI). Full 150-300 word system prompts.

#### [NEW] [code-review.yaml](file:///c:/Users/pc/Desktop/2m-code/config/teams/code-review.yaml)
Alex (Security) + Sam (Performance) + Jordan (Style).

#### [NEW] [data-science.yaml](file:///c:/Users/pc/Desktop/2m-code/config/teams/data-science.yaml)
Nova (Data Lead) + Sage (ML Engineer) + River (Data Engineer).

#### [NEW] [README.md](file:///c:/Users/pc/Desktop/2m-code/README.md)
Full README per spec: headline, what/why, install, quick start, config example, providers table, how-it-works, roadmap, contributing, MIT license.

#### [NEW] [Makefile](file:///c:/Users/pc/Desktop/2m-code/Makefile)
build, install, test, run-dev, clean targets.

#### [NEW] [install.sh](file:///c:/Users/pc/Desktop/2m-code/scripts/install.sh)
Installer script.

---

## Security Considerations

- **FastAPI binds to `127.0.0.1` only** — never `0.0.0.0`
- **All SQLite queries use parameterized statements** — no string concatenation
- **API keys read from env vars only** — never hardcoded, never logged
- **File tool validates paths** with `os.path.realpath()` to prevent directory traversal
- **Bash tool has 30s timeout** to prevent runaway processes
- **File read capped at 100KB** to prevent memory exhaustion
- **No secrets in team channel DB** — keys are resolved at call time only

## Verification Plan

### Automated
1. `go vet ./...` — no warnings
2. `go build ./cmd/2m` — produces working binary
3. Python agent engine starts and responds to `/health`

### Manual
- Verify all 30+ files exist with proper structure
- Verify YAML team configs parse correctly
- Verify Go code compiles with all dependencies resolved
