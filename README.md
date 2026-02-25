# DigiStock 📈

An AI-powered NSE stock analysis tool driven by a **multi-agent pipeline**, exposed entirely through a **Go CLI** — no browser, no server, just your terminal.

The CLI orchestrates a team of specialized AI agents (Stock Finder → Market Data Analyst → News Analyst → Trading Advisor) that together produce structured, real-time **Buy/Sell/Hold** recommendations for Indian NSE stocks.

---

## Architecture

```
digistock analyze "query"
        │
        ▼
  Go CLI (cli/digistock)
        │  spawns subprocess
        ▼
  runner.py  ──▶  agent_logic.py
                       │
              LangGraph Supervisor
                       │
        ┌──────────────┼──────────────┐
        ▼              ▼              ▼              ▼
  Stock_Finder  Market_Data  News_Analyst  Trading_Advisor
        │              │              │              │
        └──────────────┴──────────────┴──────────────┘
                       │
                  Bright Data MCP
               (live NSE web data)
```

**No web UI or HTTP server required.** The CLI talks directly to Python via subprocess.

---

## Prerequisites

| Requirement | Version |
|---|---|
| Go | ≥ 1.22 |
| Python | ≥ 3.10 |
| Node.js + npx | For Bright Data MCP |
| `uv` (recommended) | Python package manager |

---

## Setup

### 1. Clone the repo

```bash
git clone <repo-url>
cd DigiStock
```

### 2. Install Python dependencies

```bash
# Using uv (recommended)
uv sync

# Or with pip
pip install -r requirements.txt
```

### 3. Create your `.env` file

```bash
cp .env.example .env   # or create it manually
```

Edit `.env` with your API keys:

```env
# Required
OPENAI_API_KEY=sk-...
BRIGHT_DATA_API_TOKEN=your_bright_data_token

# Optional (these have defaults)
WEB_UNLOCKER_ZONE=unblocker
BROWSER_ZONE=scraping_browser
```

> **Where to get keys:**
> - `OPENAI_API_KEY` → [platform.openai.com](https://platform.openai.com/api-keys)
> - `BRIGHT_DATA_API_TOKEN` → [brightdata.com](https://brightdata.com)

### 4. Build the CLI

```bash
cd cli
go build -o digistock .
```

This produces the `digistock` binary inside `cli/`.

### 5. (Optional) Add to PATH

```bash
# Add to your ~/.zshrc or ~/.bashrc
export PATH="$PATH:/path/to/DigiStock/cli"

source ~/.zshrc
```

---

## Usage

### Check your environment

Always run this first to confirm everything is configured correctly:

```bash
./cli/digistock health
```

Expected output when ready:
```
🔍 DigiStock Environment Check
────────────────────────────────────────
  ✔ Project root found
  ✔ runner.py
  ✔ agent_logic.py
  ✔ Python interpreter           Python 3.x.x
  ✔ .env file
  ✔ OPENAI_API_KEY               set
  ✔ BRIGHT_DATA_API_TOKEN        set
────────────────────────────────────────
✔ All checks passed — run: digistock analyze
```

---

### Run a stock analysis

```bash
# Default analysis (picks 2 NSE stocks automatically)
./cli/digistock analyze

# Ask a specific question
./cli/digistock analyze "Should I buy RELIANCE or INFY today?"
./cli/digistock analyze "What are the top momentum NSE stocks this week?"
./cli/digistock analyze "Give me a short-term trade for tomorrow"
```

The pipeline streams live output as each agent completes its step:

```
╔══════════════════════════════════════════════════╗
║          DigiStock Multi-Agent Analysis          ║
╚══════════════════════════════════════════════════╝

📊 Query: Should I buy RELIANCE or INFY today?

⠋  Initialising agent pipeline...
✔ Pipeline started — streaming output...

--- 🧑‍💻 Calling Sub-Agent: Stock_Finder ---
...
--- 🧑‍💻 Calling Sub-Agent: Market_Data_Analyst ---
...
--- 🧑‍💻 Calling Sub-Agent: News_Analyst ---
...
--- 🧑‍💻 Calling Sub-Agent: Trading_Advisor ---

**RELIANCE (RELIANCE)**
**Recommendation:** Buy
**Target Price:** INR 1520
**Reason:** Strong volume breakout with bullish MACD crossover...
...
✔ Analysis complete.
```

---

### Flags

| Flag | Short | Default | Description |
|---|---|---|---|
| `--dir` | `-d` | auto-detected | Path to DigiStock project root |
| `--help` | `-h` | — | Show help |

```bash
# If the CLI can't auto-detect the project root
./cli/digistock analyze --dir /path/to/DigiStock "your query"
./cli/digistock health  --dir /path/to/DigiStock
```

---

## Project Structure

```
DigiStock/
├── agent_logic.py     # Multi-agent pipeline (LangGraph)
├── runner.py          # Python entry point called by the CLI
├── backend.py         # (Legacy) FastAPI server — not needed for CLI
├── index.html         # (Legacy) Web frontend — not needed for CLI
├── pyproject.toml     # Python project config
├── .env               # Your API keys (create this)
└── cli/
    ├── main.go
    ├── go.mod
    └── cmd/
        ├── root.go    # Root command + --dir flag
        ├── analyze.go # analyze subcommand
        └── health.go  # health subcommand
```

---

## Agent Pipeline

| Agent | Role |
|---|---|
| **Stock_Finder** | Picks 2 actively traded NSE stocks based on momentum/news/volume |
| **Market_Data_Analyst** | Fetches price, volume, RSI, MACD, moving averages for chosen stocks |
| **News_Analyst** | Summarizes recent headlines and sentiment for each stock |
| **Trading_Advisor** | Produces the final structured Buy/Sell/Hold recommendation |

All agents use **Bright Data via MCP** to access live, unrestricted financial web data.

---

## Troubleshooting

**`zsh: command not found: digistock`**
→ Either use the full path `./cli/digistock` or [add the `cli/` folder to your PATH](#5-optional-add-to-path).

**`.env file NOT SET` in health check**
→ Create a `.env` file in the project root with your API keys (see [Setup](#3-create-your-env-file)).

**`Failed to start Python`**
→ Make sure your Python virtualenv is active, or install dependencies with `uv sync`.

**Long wait before first output**
→ The Bright Data MCP server (`npx @brightdata/mcp`) starts on the first run — this is normal. Subsequent steps stream faster.
