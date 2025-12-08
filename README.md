📈 Multi-Agent Stock Analysis Chatbot

Real-time, ML-powered NSE Stock Research System

This project is a real-time streaming AI chatbot that performs multi-agent financial research for India's NSE (National Stock Exchange). It uses LangGraph, FastAPI, custom MCP servers, and Bright Data tools to deliver:

📊 Real-time market data

📰 Latest news sentiment

📉 Technical indicators

🤖 Machine learning–based price predictions

🧠 Structured Buy/Sell/Hold recommendations

All results are coordinated in a powerful multi-agent workflow, producing a complete analyst report.

✨ Features
🧩 Multi-Agent Research Workflow

A LangGraph-powered "assembly line" of agents:

Stock_Finder → identifies promising NSE stocks

Market_Data_Analyst → gathers live data (price, trends, indicators)

News_Analyst → extracts sentiment from recent headlines

Trading_Advisor → generates final recommendation

🌐 Live Web Data via Bright Data MCP

The system connects to Bright Data's Web Unlocker / Scraping Browser through:

MultiServerMCPClient

Fully MCP-compliant tool interface

This enables real-time market data scraping, bypassing site restrictions.

⚙️ Custom Stock MCP Server (New!)

A dedicated MCP server provides:

Historical market data

Technical indicators (RSI, MACD, SMA/EMA)

ML-powered price predictions (XGBoost)

Strategy backtesting (SMA crossover)

Tools available:
get_data
indicators
predict
backtest

🤖 Machine Learning Prediction Engine

A trained XGBoost model forecasts future stock prices using:

Close

Volume

RSI

SMA-50 / SMA-200

EMA-20

MACD

The ML model and scaler are stored as:
models/xgb_model.json
models/scaler.pkl

⚡ Real-Time Streaming

Responses stream token-by-token to the web UI using FastAPI’s StreamingResponse.

🖥️ Simple Frontend UI

A clean HTML/CSS/JS interface for chatting with the AI system.

🧱 Modern Tech Stack

FastAPI (backend)

LangGraph (agent orchestration)

LangChain (tool + LLM abstraction)

OpenAI GPT-4 Turbo

Bright Data MCP tools

Custom Stock MCP tools

Vanilla JS frontend

🏗️ System Architecture
   

            ┌────────────────────────────┐
               │        Frontend UI          │
               │   (HTML / CSS / JS)         │
               └─────────────┬──────────────┘
                             │ HTTP (Stream)
                             ▼
               ┌────────────────────────────┐
               │         FastAPI             │
               │      (backend.py)           │
               └─────────────┬──────────────┘
                             │
                             ▼
               ┌────────────────────────────┐
               │      Multi-Agent System     │
               │     (agent_logic.py, LG)    │
               ├────────────────────────────┤
               │ Stock_Finder Agent          │
               │ Market_Data_Analyst Agent   │
               │ News_Analyst Agent          │
               │ Trading_Advisor Agent       │
               └─────────────┬──────────────┘
      ┌──────────────────────┼────────────────────────┐
      ▼                      ▼                        ▼
┌───────────────┐   ┌─────────────────┐     ┌──────────────────────┐
│ Bright Data    │   │ Custom Stock MCP│     │ OpenAI GPT-4 Turbo   │
│ MCP Tools      │   │ (XGBoost, TA,   │     │ LLM Reasoning Engine │
│ (Scraping)     │   │  Backtesting)   │     └──────────────────────┘
└───────────────┘   └─────────────────┘


📂 Project Structure

├── frontend/
│   └── index.html
│
├── backend/
│   └── backend.py
│
├── agent/
│   └── agent_logic.py
│
├── multi_agent.ipynb       # Example notebook for dev/testing
│
├── digi_mcp/               # NEW — Custom Stock MCP Server
│   ├── server.py
│   ├── mcp.json
│   ├── tools/
│   │   ├── market_data.py
│   │   ├── indicators.py
│   │   ├── predictions.py
│   │   ├── backtesting.py
│   │   ├── utils.py
│   └── models/
│       ├── xgb_model.json
│       ├── scaler.pkl
│
└── train_xgb_model.py      # NEW — ML Training Script


🤖 Multi-Agent Workflow

User Query → sent to /chat

FastAPI streams to LangGraph

Supervisor activates agents:

Stock_Finder → chooses stocks

Market_Data_Analyst → fetches market data

News_Analyst → processes news

Trading_Advisor → final recommendation

Bright Data MCP + Custom Stock MCP provide tools

Final structured report streams back to UI