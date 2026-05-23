# 2M Code — Setup Guide

> Run AI agent teams on your machine. One setup, all platforms.

---

## Quick Start (30 seconds)

```bash
# 1. Set your API key (pick one provider)
export OPENROUTER_API_KEY="sk-or-..."   # OpenRouter (works with any model)

# 2. Start chatting
2m chat fullstack
```

---

## Prerequisites

| Requirement | Version | Check with |
|-------------|---------|------------|
| Go | 1.22+ | `go version` |
| Python | 3.11+ | `python --version` |

---

## Installation

### Step 1: Clone the repo

```bash
git clone https://github.com/ArafatAhmed-2M/2M-Code.git
cd 2M-Code
```

### Step 2: Install Python dependencies

```bash
pip install -r requirements.txt
```

### Step 3: Build the binary

**macOS / Linux:**
```bash
go build -o bin/2m ./cmd/2m
sudo cp bin/2m /usr/local/bin/2m   # Optional: add to PATH
```

**Windows:**
```powershell
go build -o bin\2m.exe .\cmd\2m
# Add bin\ folder to your PATH, or run from the project directory
```

---

## API Key Setup

You need at least **one** API key. 2M Code auto-detects which keys are available.

### Option A: OpenRouter (recommended — one key for any model)

Get a key at [openrouter.ai/keys](https://openrouter.ai/keys), then:

```bash
export OPENROUTER_API_KEY="sk-or-..."
```

Free models available: `qwen/qwen3-coder:free`, `google/gemini-2.0-flash-lite-preview-02-05:free`, `meta-llama/llama-3.2-3b-instruct:free`

### Option B: Provider-specific keys

Set whichever you have:

```bash
export ANTHROPIC_API_KEY="sk-ant-..."   # Claude models
export OPENAI_API_KEY="sk-proj-..."     # GPT models
export GOOGLE_API_KEY="AIza..."         # Gemini models
export MISTRAL_API_KEY="..."            # Mistral models
export GROQ_API_KEY="gsk_..."           # Groq (fast, free tier)
export COHERE_API_KEY="..."             # Command models
```

**Windows (PowerShell):**
```powershell
$env:OPENROUTER_API_KEY = "sk-or-..."
```

**Windows (CMD):**
```cmd
set OPENROUTER_API_KEY=sk-or-...
```

**Persist across terminals** — add the export line to your shell profile:
- **macOS/Linux:** `~/.bashrc`, `~/.zshrc`, or `~/.profile`
- **Windows:** Set via System Environment Variables GUI

---

## Run the Engine

2M Code runs a lightweight Python server on `localhost:8765`. Start it before your first command:

```bash
# Start the agent engine (keep this terminal open)
python -m uvicorn agent_engine.server:app --host 127.0.0.1 --port 8765
```

Or let the 2M binary manage it automatically:

```bash
# The binary starts the engine for you
2m run fullstack "Hello world"
```

---

## Your First Chat

```bash
# List available teams
ls config/teams/

# Start a chat with the full-stack team
2m chat fullstack

# Run a one-shot task
2m run fullstack "Build a REST API with Go"
```

### Chat Commands

| Command | What it does |
|---------|-------------|
| Type any message | Chat with your agent team |
| `/help` | Show available commands |
| `/info` | Show team configuration |
| `/exit` or `exit` | End the session |

---

## Platform-Specific Notes

### macOS

```bash
# Install Go
brew install go

# Install Python
brew install python@3.12

# Set up PATH
echo 'export PATH="$PATH:$HOME/go/bin"' >> ~/.zshrc
source ~/.zshrc

# Build and install
go build -o bin/2m ./cmd/2m
sudo cp bin/2m /usr/local/bin/2m
```

### Linux (Ubuntu/Debian)

```bash
# Install Go
sudo apt install golang-go

# Install Python
sudo apt install python3 python3-pip python3-venv

# Build
go build -o bin/2m ./cmd/2m
sudo cp bin/2m /usr/local/bin/2m
```

### Windows

```powershell
# Install Go — download from https://go.dev/dl/
# Install Python — download from https://python.org/downloads/

# Build
go build -o bin\2m.exe .\cmd\2m

# Add to PATH (PowerShell Admin):
$env:Path += ";$PWD\bin"
[Environment]::SetEnvironmentVariable("Path", $env:Path, "User")

# For repeated use, add bin\ to your system PATH
```

---

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| `Agent engine is not running` | Start the engine: `python -m uvicorn agent_engine.server:app --host 127.0.0.1 --port 8765` |
| `API key not set` | Set one of the API key env vars above |
| Provider not responding | Check your API key is valid and has credits |
| Port 8765 in use | Kill the old process: `kill $(lsof -t -i:8765)` (macOS/Linux) or `Stop-Process -Id (Get-NetTCPConnection -LocalPort 8765).OwningProcess` (Windows) |
| `command not found: 2m` | Run from project dir: `./bin/2m` or add to PATH |

---

## Clean Uninstall

```bash
# Remove the binary
rm /usr/local/bin/2m          # macOS/Linux
del C:\path\to\2m.exe         # Windows

# Remove config and data
rm -rf ~/.2mcode              # All teams, sessions, memory

# Remove repo
rm -rf 2M-Code/
```

---

## Next Steps

- Read the full [README.md](README.md)
- Create your own team: `2m new-team`
- Explore example teams in `config/teams/`
- Set `OPENROUTER_API_KEY` for persistent memory across sessions
