# DigiStock 📈

An AI-powered NSE stock analysis tool driven by a **multi-agent pipeline**, exposed entirely through a **Go CLI** — no browser, no server, just your terminal.

The CLI orchestrates a team of specialized AI agents (Stock Finder → Market Data Analyst → News Analyst → Trading Advisor) and also exposes direct technical tools (indicators, prediction, backtesting) — all from one binary.

---

## Architecture

```
digistock <command>
        │
        ▼
  Go CLI (cli/digistock)
        │  spawns subprocess
        ├──▶  runner.py  ──▶  agent_logic.py  (AI multi-agent pipeline)
        │
        └──▶  runner_tools.py  ──▶  tools/    (yfinance + XGBoost tools)
                                      ├── indicators.py
                                      ├── predictions.py
                                      └── backtesting.py
```

**No web UI or HTTP server.** The CLI talks directly to Python via subprocess.

---

## Prerequisites

| Requirement | Version |
|---|---|
| Go | ≥ 1.22 |
| Python | ≥ 3.10 |
| Node.js + npx | For Bright Data MCP (analyze command only) |
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
# Required for: digistock analyze
OPENAI_API_KEY=sk-...
BRIGHT_DATA_API_TOKEN=your_bright_data_token

# Optional (these have defaults)
WEB_UNLOCKER_ZONE=unblocker
BROWSER_ZONE=scraping_browser
```

> **Note:** `digistock indicators`, `predict`, and `backtest` use only Yahoo Finance — **no API keys required** for those commands.

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

Always run this first:

```bash
./cli/digistock health
```

Expected output when ready:
```
🔍 DigiStock Environment Check
────────────────────────────────────────
  ✔ Project root found
  ✔ runner.py
  ✔ runner_tools.py
  ✔ agent_logic.py
  ✔ Python interpreter           Python 3.x.x
  ✔ .env file
  ✔ OPENAI_API_KEY               set
  ✔ BRIGHT_DATA_API_TOKEN        set
────────────────────────────────────────
✔ All checks passed — run: digistock analyze
```

---

### Run a stock analysis (requires API keys)

```bash
# Default analysis (picks 2 NSE stocks automatically)
./cli/digistock analyze

# Ask a specific question
./cli/digistock analyze "Should I buy RELIANCE or INFY today?"
./cli/digistock analyze "What are the top momentum NSE stocks this week?"
```

---

### Fetch technical indicators (no API key needed)

```bash
./cli/digistock indicators TCS
./cli/digistock indicators RELIANCE
```

Output:
```
📊  Indicators for TCS.NS
──────────────────────────────────────────
  SMA 50          : 3842.15
  SMA 200         : 3712.44
  EMA 20          : 3891.02
  RSI (14)        : 58.34
  MACD            : 42.1837
  MACD Signal     : 35.7203
  Trend           : Bullish
──────────────────────────────────────────
```

---

### Predict next-N-day price (requires models/)

```bash
./cli/digistock predict TCS
./cli/digistock predict RELIANCE --days 10
```

---

### Backtest an SMA crossover strategy (no API key needed)

```bash
./cli/digistock backtest TCS
./cli/digistock backtest RELIANCE --strategy sma_cross
```

Output:
```
📈  Backtest Results for TCS.NS
──────────────────────────────────────────
  Strategy Return : +18.42%
  Buy & Hold      : +14.87%
  Outperformance  : +3.55%
──────────────────────────────────────────
```

---

### Global Flags

| Flag | Short | Default | Description |
|---|---|---|---|
| `--dir` | `-d` | auto-detected | Path to DigiStock project root |
| `--help` | `-h` | — | Show help |

---

## Project Structure

```
DigiStock/
├── agent_logic.py      # Multi-agent pipeline (LangGraph)
├── runner.py           # Python entry point for: digistock analyze
├── runner_tools.py     # Python entry point for: indicators / predict / backtest
├── tools/
│   ├── market_data.py  # Yahoo Finance data fetching
│   ├── indicators.py   # RSI, MACD, SMA, EMA
│   ├── predictions.py  # XGBoost price prediction
│   ├── backtesting.py  # SMA crossover backtest
│   └── utils.py        # Ticker normalization
├── models/
│   ├── xgb_model.json  # Trained XGBoost model
│   └── scaler.pkl      # Feature scaler
├── pyproject.toml
├── .env                # Your API keys (create this)
└── cli/
    ├── main.go
    ├── go.mod
    └── cmd/
        ├── root.go        # Root command + --dir flag
        ├── analyze.go     # digistock analyze
        ├── health.go      # digistock health
        └── tools.go       # digistock indicators / predict / backtest
```

---

## Agent Pipeline (analyze command)

| Agent | Role |
|---|---|
| **Stock_Finder** | Picks 2 actively traded NSE stocks based on momentum/news/volume |
| **Market_Data_Analyst** | Fetches price, volume, RSI, MACD, moving averages |
| **News_Analyst** | Summarizes recent headlines and sentiment |
| **Trading_Advisor** | Produces the final structured Buy/Sell/Hold recommendation |

All agents use **Bright Data via MCP** to access live, unrestricted financial web data.

---

## Troubleshooting

**`zsh: command not found: digistock`**
→ Use the full path `./cli/digistock` or [add `cli/` to your PATH](#5-optional-add-to-path).

**`.env file NOT SET` in health check**
→ Create a `.env` file in the project root with your API keys (see [Setup](#3-create-your-env-file)).

**`Failed to start Python`**
→ Make sure your Python virtualenv is active, or install dependencies with `uv sync`.

**`predict` returns an error**
→ Check that `models/xgb_model.json` and `models/scaler.pkl` exist in the project root.

**scikit-learn version warning on `predict` / `backtest`**
→ The model scaler was pickled on a different scikit-learn version. Retrain to fix it:
```bash
uv run python xgb_model.py
```
