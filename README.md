# DigiStock 📈

An AI-powered NSE stock analysis tool driven by a **multi-agent pipeline**, exposed entirely through a **Go CLI** — no browser, no server, just your terminal.

Run everything through the interactive **TUI** or use individual subcommands directly.

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
        └──▶  runner_tools.py  ──▶  tools/    (yfinance + XGBoost)
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
| Node.js + npx | For Bright Data MCP (`analyze` only) |
| `uv` (recommended) | Python package manager |

---

## Docker

No local Python or Go installation needed — run everything in a container.

### Build the image

```bash
docker build -t digistock .
```

### Run the TUI (interactive)

```bash
docker run -it --env-file .env digistock
```

### Run a direct command

```bash
# No API key needed
docker run --rm -it --env-file .env digistock indicators TCS
docker run --rm -it --env-file .env digistock backtest TCS

# Requires API keys in .env
docker run --rm -it --env-file .env digistock analyze "Should I buy RELIANCE today?"
```

### Using docker compose

```bash
# Launch TUI
docker compose up

# Run a specific command
docker compose run --rm digistock indicators TCS
docker compose run --rm digistock health
```

> **Note:** The TUI requires `-it` (interactive + TTY). Streaming commands also need `-it` for live output.

---

## Manual Setup

### 1. Clone the repo

```bash
git clone <repo-url>
cd DigiStock
```

### 2. Install Python dependencies

```bash
uv sync          # recommended
# or: pip install -r requirements.txt
```

### 3. Create your `.env` file

```env
# Required for: digistock analyze
OPENAI_API_KEY=sk-...
BRIGHT_DATA_API_TOKEN=your_bright_data_token

# Optional (these have defaults)
WEB_UNLOCKER_ZONE=unblocker
BROWSER_ZONE=scraping_browser
```

> `indicators`, `predict`, and `backtest` use only Yahoo Finance — **no API keys needed**.

### 4. Build the CLI

```bash
cd cli
go build -o digistock .
```

### 5. (Optional) Add to PATH

```bash
export PATH="$PATH:/path/to/DigiStock/cli"
```

---

## Usage

### Interactive TUI ✨

The easiest way to use DigiStock — a fullscreen keyboard-driven interface:

```bash
./cli/digistock tui
```

**Screen flow:**
```
🏠 Menu  →  📝 Form  →  ⟳ Streaming  →  ✔ Results
                ↑_______ q / Esc back ________↑
```

**Form screen** — each command shows:
- A one-line context note (API key requirements, what it does)
- A `▶` marker on the active field you're typing in
- A `↳` hint explaining the expected format
- Example values so you're never guessing

```
  ✦ 📊 Indicators
  Fetches live technical indicator data from Yahoo Finance. No API key required.

  ▶ NSE ticker:
    TCS█
    ↳  Just the symbol — .NS suffix is added for you automatically
    eg: TCS  ·  RELIANCE  ·  INFY  ·  HDFCBANK  ·  ICICIBANK

  Tab  next field   ↵  run   Esc  back
```

**Keybindings:**

| Screen | Key | Action |
|---|---|---|
| Menu | `↑/↓` or `j/k` | Navigate |
| Menu | `↵` | Select command |
| Menu | `q` / `Ctrl+C` | Quit |
| Form | `Tab` | Next field |
| Form | `↵` | Run |
| Form | `Esc` | Back to menu |
| Results | `j/k` | Scroll |
| Results | `g / G` | Top / Bottom |
| Results | `q` | Back to menu |

---

### Direct Commands

#### Check your environment

```bash
./cli/digistock health
```

#### AI stock analysis (requires API keys)

```bash
./cli/digistock analyze
./cli/digistock analyze "Should I buy RELIANCE or INFY today?"
```

#### Technical indicators (no API key)

```bash
./cli/digistock indicators TCS
```

```
📊  Indicators for TCS.NS
──────────────────────────────────────────
  SMA 50          : 3065.00
  EMA 20          : 2954.12
  RSI (14)        : 13.23
  MACD            : -133.01  →  Bearish
──────────────────────────────────────────
```

#### Price prediction (no API key)

```bash
./cli/digistock predict TCS
./cli/digistock predict RELIANCE --days 10
```

#### Backtest SMA strategy (no API key)

```bash
./cli/digistock backtest TCS
./cli/digistock backtest RELIANCE --strategy sma_cross
```

```
📈  Backtest Results for TCS.NS
──────────────────────────────────────────
  Strategy Return : -13.75%
  Buy & Hold      : -29.97%
  Outperformance  : +16.23%
──────────────────────────────────────────
```

---

### Global Flags

| Flag | Default | Description |
|---|---|---|
| `--dir / -d` | auto-detected | Path to DigiStock project root |
| `--help / -h` | — | Show help |

---

## Project Structure

```
DigiStock/
├── agent_logic.py      # Multi-agent pipeline (LangGraph)
├── runner.py           # Entry point for: digistock analyze
├── runner_tools.py     # Entry point for: indicators / predict / backtest
├── xgb_model.py        # Retrain XGBoost model + scaler
├── tools/
│   ├── market_data.py  # Yahoo Finance fetching
│   ├── indicators.py   # RSI, MACD, SMA, EMA
│   ├── predictions.py  # XGBoost price prediction
│   ├── backtesting.py  # SMA crossover backtest
│   └── utils.py        # Ticker normalization (.NS suffix)
├── models/
│   ├── xgb_model.json
│   └── scaler.pkl
├── pyproject.toml
├── .env
└── cli/
    ├── main.go
    ├── go.mod
    └── cmd/
        ├── root.go        # Root command + --dir flag
        ├── tui.go         # digistock tui
        ├── analyze.go     # digistock analyze
        ├── health.go      # digistock health
        └── tools.go       # digistock indicators / predict / backtest
    └── tui/
        ├── model.go       # Bubble Tea model (menu → form → stream → result)
        └── styles.go      # Lip Gloss styles
```

---

## Agent Pipeline (`analyze`)

| Agent | Role |
|---|---|
| **Stock_Finder** | Picks 2 actively traded NSE stocks |
| **Market_Data_Analyst** | Fetches price, volume, RSI, MACD, moving averages |
| **News_Analyst** | Summarizes recent headlines and sentiment |
| **Trading_Advisor** | Produces structured Buy/Sell/Hold recommendation |

All agents use **Bright Data via MCP** for live, unrestricted financial web data.

---

## Troubleshooting

**`zsh: command not found: digistock`**
→ Use `./cli/digistock` or add `cli/` to your PATH.

**`.env file NOT SET` in health check**
→ Create `.env` with your API keys (see [Setup](#3-create-your-env-file)).

**`Failed to start Python`**
→ Run `uv sync` to install dependencies.

**`predict` returns an error**
→ Check that `models/xgb_model.json` and `models/scaler.pkl` exist.

**scikit-learn version warning on `predict` / `backtest`**
→ Retrain the model to regenerate the scaler:
```bash
uv run python xgb_model.py
```
