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

# 2M Code — AI Session Context
**Saved:** 2026-05-24  
**Purpose:** Allows AI to resume work after context restart without losing state.

---

## Session 3 Summary (Polish & Config)

This session **polished everything for GitHub push**:

### Infrastructure
- **`base_url` in team YAML** — `Agent.BaseURL` field added end-to-end: Go Agent struct → bridge → Python server → provider. Per-agent endpoint config for `openai_compatible`, overriding `OPENAI_COMPATIBLE_BASE_URL` env var.
- **`**kwargs` on all providers** — Future provider-specific configs pass through without breaking other providers.
- **Pure-Go SQLite** — Migrated to `modernc.org/sqlite`, no CGO/GCC needed on Windows.
- **`2mcode` Windows launcher** — `bin/2mcode.cmd` on PATH allows typing `2mcode` from any terminal.

### UX Fixes
- **Streaming buffer** — `PrintAgentText` accumulates chunks and flushes on newlines; empty responses no longer produce stray `│` lines.
- **Instant CLI** — Engine startup deferred for `--help`, `--version`, `new-team`, `team`, `config`, `completion`, and bare invocation.
- **ASCII logo centered** — All 8 `.md` files use `<div align="center">` for the 2M CODE banner.

### Documentation
- All `.md` files updated with centered 2M CODE ASCII logo.
- `agent.md` Bugs Fixed table expanded with all recent fixes + features.
- `context-for-ai.md` updated with full live state.
- `issue.md` updated with base_url entry.
- `SETUP.md` updated with `base_url` YAML config note.

### Previous Sessions
- **Session 2 (V3 P0):** Plugin/extension system — `plugin_base.py`, `plugin_loader.py`, `2m plugin list` CLI, example plugins.
- **Session 1:** 8 bugs fixed, `openai_compatible` provider added, all docs updated.

---

## Project State

| Aspect | Status |
|--------|--------|
| Go build (`go build ./cmd/2m`) | ✅ Passes (pure-Go SQLite, no CGO needed) |
| Go vet (`go vet ./...`) | ✅ Passes |
| Python syntax (all .py files) | ✅ Passes |
| Branch | `main` |
| Remote | `origin/main` — `https://github.com/ArafatAhmed-2M/2M-Code` |
| V2 gaps closed | Streaming renderer ✅ — remaining: tests, history, web_fetch, chat budget |

---

## V3 Feature Progress

| Priority | Feature | Status |
|----------|---------|--------|
| P0 | Plugin/extension system | ✅ Complete |
| P1 | GitHub PR & CI/CD integration | 🔲 Not started |
| P2 | Agent self-improvement loops | 🔲 Not started |
| P3 | Web dashboard (read-only) | 🔲 Not started |
| P4 | V2 gap closure (tests, history, web_fetch, streaming, chat budget) | 🔶 In progress (streaming ✅) |

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
4. **V2 gap closure** — tests, `2m history`, `web_fetch` fix, chat budget (streaming ✅ done)

### V2 gaps (still open)
- **Tests** — No test files exist in Go or Python
- **`2m history` command** — Stub only in `internal/cli/team.go:173-186`
- **`web_fetch` tool** — Go-side returns stub string instead of fetching URL
- **Streaming renderer** — ✅ **FIXED** — buffers chunks and flushes on newlines
- **Chat token budget** — `RunTask` enforces `MaxTokensPerRun` but `RunChatTurn` does not

## Recent Fixes (Session: 2026-05-24)

### Bugs Fixed (in chronological order)
| # | File | Bug | Fix |
|---|------|-----|-----|
| 1 | `internal/cli/newteam.go:91` | `openai_compatible` missing from new-team wizard | Added to options list |
| 2 | `internal/team/config.go:180` | Error message missing `openai_compatible` | Added to error message |
| 3 | `agent_engine/server.py:119` | Error listed only 4 providers instead of 9 | Updated to all 9 |
| 4 | `scripts/install.sh:150-158` | `OPENAI_COMPATIBLE_API_KEY` missing from next steps | Added env var + base URL note |
| 5 | `internal/bus/schema.go` | Binary requires CGO (`go-sqlite3`), fails without GCC | Migrated to `modernc.org/sqlite` |
| 6 | `internal/cli/renderer.go` | Streaming chunks printed per-chunk; empty responses show stray `│` | Buffered flushes on newlines |
| 7 | `cmd/2m/main.go` | Engine startup blocks help/version/bare invocation | Added `needsEngine()` skip |

### Features Added
| Feature | Details |
|---------|---------|
| **`base_url` in team YAML** | `Agent.BaseURL` overrides `OPENAI_COMPATIBLE_BASE_URL` env var; per-agent endpoint config |
| **`**kwargs` on all providers** | Future provider-specific configs pass through without breaking others |
| **`2mcode` Windows launcher** | `bin/2mcode.cmd` on PATH — `2mcode` works from any terminal |
| **Instant CLI** | Engine deferred for help/version/config wizards |
| **Pure-Go SQLite** | `modernc.org/sqlite` — no GOP/CC needed |
| **ASCII logo centered** | All 8 `.md` files — `<div align="center">` wrapper |
| **Test team `test-openrouter`** | `config/teams/test-openrouter.yaml` using OpenRouter free models |
| **Verified working** | OpenRouter key valid, MiniMax model responded, memory system functional |
