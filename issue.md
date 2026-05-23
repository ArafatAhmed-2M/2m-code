# 2M Code — Issue Log

> Every bug, error, and fix encountered during development.
> When an AI agent fixes a bug, it must document it here with the fix.

---

## 1. Wrong GitHub repo URLs in README & install script

**File(s):** `README.md`, `scripts/install.sh`, `agent.md`, `internal/cli/root.go`, `agent_engine/providers/openrouter_provider.py`

**Problem:** All URLs pointed to `github.com/2mcode/2mcode` but the actual remote is `github.com/ArafatAhmed-2M/2M-Code`. The quick-start curl command returned a 404.

**Fix:** Updated 29 references across 12 files to use the correct repo URL. The Go module path (`go.mod`) was left as `github.com/2mcode/2mcode` since that's an import path, not a web URL.

**Commit:** `5a830a1`

---

## 2. `2m chat` command not found ("unknown command")

**File(s):** `internal/cli/chat.go`

**Problem:** Running `2m chat <team>` returned `Error: unknown command "chat" for "2m"`. The `chat.go` file defined `chatCmd` and `runChat` but was missing the `init()` function to register itself with Cobra's `rootCmd.AddCommand(chatCmd)`. All other commands (`run`, `new-team`, `team`, `models`) had this registration.

**Fix:** Added `func init() { rootCmd.AddCommand(chatCmd) }` to `chat.go`.

**Commit:** `aad22ba`

---

## 3. `2m new-team` reviewer prompt accepts free-text instead of numbered selection

**File(s):** `internal/cli/newteam.go`

**Problem:** The reviewer prompt used `prompt()` (free-text input) while the leader prompt used `promptWithOptions()` (numbered selection). User typed "1" expecting numbered selection but it was treated as an agent name, causing `workflow reviewer '1' is not a defined agent`.

**Fix:** Changed reviewer prompt to use `promptWithOptions()` with agent names + a `(skip)` option as the first choice.

**Commit:** `4e9c5b1`

---

## 4. Provider registration incomplete — missing env vars & validation

**File(s):** `internal/team/config.go`, `internal/team/team.go`, `internal/cli/newteam.go`

**Problem:** 8 provider Python files existed (`anthropic`, `google`, `openai`, `mistral`, `cohere`, `groq`, `ollama`, `openrouter`) but the Go backend only recognized 4:
- `GetProviderAPIKey()` only mapped 4 providers' env vars — cohere, groq, openrouter missing; ollama (no key) not handled
- Team validation error message only listed 4 providers
- `new-team` wizard only offered 4 provider choices

**Fix:** Added all 8 providers to env var map, validation error, and wizard options. Special-cased ollama (no API key needed).

**Commit:** `5a830a1`

---

## 5. Install script missing permissions on binary

**File(s):** `scripts/install.sh`

**Problem:** The install script copies the binary to `/usr/local/bin/2m` but doesn't set executable permissions. Running `2m` gives `Permission denied`.

**Workaround:** Run `sudo chmod +x /usr/local/bin/2m` after installation.

**Status:** Needs fix — add `chmod +x` after the copy in the install script.

---

## 6. Agent engine port conflict on re-run

**File(s):** `cmd/2m/main.go`

**Problem:** When `2m` is run again while a previous instance is still running, the Python agent engine fails to bind to port 8765 with `[Errno 98] address already in use`.

**Workaround:** Run `pkill -f agent_engine` before re-running.

**Status:** Needs fix — could kill previous Python process on startup or use a PID file.

---

## 7. Bash tool timeout blocks agent loop

**File(s):** `agent_engine/tools/__init__.py`

**Problem:** Commands like `python3 -m http.server 8000` that run indefinitely (servers, watchers) hit the 30-second timeout and get killed. The tool returns a timeout error instead of starting the server in the background.

**Workaround:** Agents can use `command &` (background) but the tool still waits and times out.

**Status:** Needs fix — background commands should return immediately with the PID.
