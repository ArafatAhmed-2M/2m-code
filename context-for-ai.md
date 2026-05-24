# 2M Code — AI Session Context
**Saved:** 2026-05-24  
**Purpose:** Allows AI to resume work after context restart without losing state.

---

## Session 2 Summary (V3 Begins)

This session **started V3** by building the **plugin/extension system (P0)**:
- `agent_engine/plugin_base.py` — Plugin base class with lifecycle hooks
- `agent_engine/plugin_loader.py` — Scans `~/.2mcode/plugins/` and `.2mcode/plugins/` for .py files
- `agent.py` — Added `init_plugins()`, `shutdown_plugins()`, hook integration in `run_agent()` and `run_agent_stream()`
- `server.py` — Startup/shutdown event hooks, `/plugins` endpoint
- `internal/cli/plugin.go` — `2m plugin list` CLI command
- `internal/bridge/bridge.go` — Added `StylePath()` helper
- `.2mcode/plugins/turn_logger.py` — Example plugin that logs turns to a file
- `.2mcode/plugins/context_injector.py` — Example plugin that injects coding guidelines

Previous work (Session 1): 8 bugs fixed, openai_compatible provider added, all docs updated.

---

## Project State

| Aspect | Status |
|--------|--------|
| Go build (`go build ./cmd/2m`) | ✅ Passes |
| Go vet (`go vet ./...`) | ✅ Passes |
| Python syntax (all .py files) | ✅ Passes |
| Last commit | `1c88de5` — placeholder (will be replaced) |
| Branch | `main` |

---

## V3 Feature Progress

| Priority | Feature | Status |
|----------|---------|--------|
| P0 | Plugin/extension system | ✅ Complete |
| P1 | GitHub PR & CI/CD integration | 🔲 Not started |
| P2 | Agent self-improvement loops | 🔲 Not started |
| P3 | Web dashboard (read-only) | 🔲 Not started |
| P4 | V2 gap closure (tests, history, web_fetch, streaming, chat budget) | 🔲 Not started |

---

## Plugin System Architecture

### User-facing
1. Place a `.py` file in `~/.2mcode/plugins/` (global) or `.2mcode/plugins/` (project)
2. Subclass `Plugin` from `plugin_base.Plugin`
3. Override any hooks: `on_startup`, `on_shutdown`, `on_agent_turn_start`, `on_agent_turn_end`, `on_tool_exec`
4. Run `2m plugin list` to verify it loaded

### Implementation
- `plugin_loader.discover_plugins()` iterates directories, imports each `.py` file, finds `Plugin` subclasses, instantiates them
- `agent.init_plugins(server_app)` called on server startup via FastAPI `on_startup`
- `agent.shutdown_plugins()` called on server shutdown via FastAPI `on_shutdown`
- `_run_plugin_turn_start_hooks(req)` chains request through each plugin's `on_agent_turn_start`
- `_run_plugin_turn_end_hooks(response)` chains response through each plugin's `on_agent_turn_end`
- `server.py` GET `/plugins` endpoint returns loaded plugins with their hook list
- Go `2m plugin list` scans both dirs for `.py` files + queries `/plugins` for loaded plugin info

### Plugin directories checked (in order)
1. `~/.2mcode/plugins/` — global, user-wide
2. `$CWD/.2mcode/plugins/` — project-local (if CWD is project root)
3. `$CWD/../.2mcode/plugins/` — project root (if CWD is agent_engine/)
4. `$CWD/../../.2mcode/plugins/` — further up (fallback)

---

## V3/V4/V5 Roadmap (from PRD.md)

### V3 — Extensibility & Integration
| Milestone | Scope | Status |
|-----------|-------|--------|
| M11 — Plugin System | Python-based plugins with lifecycle hooks | ✅ Complete |
| M12 — GitHub Integration | Auto-review PRs, run on push via webhooks | 🔲 |
| M13 — Feedback Loops | Agents review each other's work | 🔲 |
| M14 — Web Dashboard | Read-only web UI for monitoring | 🔲 |
| M15 — V2 Gap Closure | Tests, history, web_fetch, streaming, budget | 🔲 |

### V4 — Enterprise & Collaboration
| Milestone | Scope |
|-----------|-------|
| M16 — Multi-User | Team members share the same team channel |
| M17 — Team Management | Invite users, roles, access control for teams |
| M18 — Audit Logs | Every agent action logged with timestamp and user |
| M19 — Agent Personas | Agents persist history across projects |
| M20 — Analytics | Dashboards, cost breakdowns, performance metrics |

### V5 — Autonomous & Intelligent
| Milestone | Scope |
|-----------|-------|
| M21 — Autonomous Mode | Agents proactively suggest tasks |
| M22 — Cross-Project Memory | Agents transfer learning between projects |
| M23 — NL Workflow Builder | Describe team in plain English, auto-generate YAML |
| M24 — Self-Hosted Models | Deep integration with local model serving |
| M25 — Real-Time Collab | Multiple users simultaneously interacting with same team |

---

## Current Architecture

### Go CLI (`cmd/2m/main.go`)
- Starts Python agent engine subprocess, health-checks it, runs Cobra CLI, kills Python on exit
- Searches for `server.py` in: `2M_ENGINE_PATH` → `~/.2mcode/agent_engine/` → relative to binary → relative to CWD
- On Windows, uses `taskkill` to free port 8765

### Go Orchestrator (`internal/orchestrator/orchestrator.go`)
- `RunTask()` — creates session, posts task, runs agents in schedule order, prints summary with per-agent cost, saves memory
- `RunChatTurn()` — same but interactive, no budget enforcement
- `runAgentTurn()` — gets history, builds request with memory context, streams via bridge, tool use loop (max 5 iterations), posts response to bus

### Go Bridge (`internal/bridge/bridge.go`)
- `Call()` — HTTP POST to `/call` (non-streaming)
- `CallStream()` — HTTP POST to `/call` with SSE reading, `onEvent` callback for each chunk
- `ListModels()` — GET `/models`
- `WaitForReady()` — polls `/health` every 200ms

### Python Agent Engine (`agent_engine/server.py`)
- FastAPI on `127.0.0.1:8765`
- `POST /call` — non-streaming returns JSON, streaming returns SSE
- `GET /health` — returns `{"status": "ok"}`
- `GET /models` — returns `{provider: [model, ...]}`
- `GET /plugins` — returns loaded plugins with hooks
- Startup: runs `init_plugins()`, Shutdown: runs `shutdown_plugins()`

### Python Agent Router (`agent_engine/agent.py`)
- `_resolve_provider()` — returns provider module, falls back to OpenRouter if provider's env var is missing but `OPENROUTER_API_KEY` is set
- `init_plugins(server_app)` — discovers & initializes plugins, runs `on_startup`
- `shutdown_plugins()` — runs each plugin's `on_shutdown`
- `run_agent()` — runs plugin turn-start hooks → provider call → plugin turn-end hooks
- `run_agent_stream()` — same hook chain with streaming support

### Team Loading (`internal/team/team.go`)
- Search order: `./.2mcode/teams/` → `~/.2mcode/teams/` → `~/.2mcode/config/teams/` → relative to binary → `config/teams/` (relative to CWD)
- `Validate()` checks agents, tools, workflow, sets defaults

### Memory (`internal/memory/`)
- `FileStore` — JSONL files at `~/.2mcode/memory/<team>.jsonl`, thread-safe (RWMutex)
- `Summarizer` — calls LLM via bridge to summarize session, saves entry
- `BuildContext()` — loads last 5 entries, formats as `[PAST SESSION MEMORY]` block

---

## Key File Locations (V3 additions marked **NEW**)

```
agent_engine/
├── server.py                        ← FastAPI server (+ plugins endpoint)
├── agent.py                         ← Agent router (+ plugin hooks)
├── plugin_base.py                    ← ** NEW ** Plugin base class
├── plugin_loader.py                  ← ** NEW ** Plugin discovery/loading
├── providers/...
├── tools/...

internal/cli/
├── plugin.go                         ← ** NEW ** `2m plugin list` command

.2mcode/plugins/
├── turn_logger.py                    ← ** NEW ** Example plugin
├── context_injector.py               ← ** NEW ** Example plugin
```

---

## What's Still Needed (for next agent)

### V3 P1-P4 (see priority order in agent.md)
1. **GitHub PR Integration** — `2m github review <pr-url>`, webhook server
2. **Agent self-improvement loops** — agents review each other's work
3. **Web dashboard** — read-only session monitoring (FastAPI + Jinja2 + HTMX)
4. **V2 gap closure** — tests, `2m history`, `web_fetch` fix, streaming fix, chat budget

### V2 gaps (still open)
- **Tests** — No test files exist in Go or Python
- **`2m history` command** — Stub only in `internal/cli/team.go:173-186`
- **`web_fetch` tool** — Go-side returns stub string instead of fetching URL
- **Streaming renderer** — `PrintAgentText` prints every SSE chunk on new line
- **Chat token budget** — `RunTask` enforces `MaxTokensPerRun` but `RunChatTurn` does not
