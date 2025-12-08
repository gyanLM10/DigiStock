# server.py
import asyncio
from fastmcp import MCP, mcp_tool

# Import tools
from tools.market_data import get_historical_data
from tools.indicators import compute_indicators
from tools.predictions import predict_price
from tools.backtesting import backtest_strategy

mcp = MCP("digi-stock-mcp", version="1.0.0")


# -------------------------
# Register Tools
# -------------------------
@mcp.tool()
async def get_data(ticker: str, period: str = "6mo", interval: str = "1d"):
    """
    Fetch historical stock data from Yahoo Finance.
    """
    return await get_historical_data(ticker, period, interval)


@mcp.tool()
async def indicators(ticker: str, period: str = "6mo"):
    """
    Compute technical indicators such as RSI, MACD, Moving Averages.
    """
    return await compute_indicators(ticker, period)


@mcp.tool()
async def predict(ticker: str, horizon: int = 5):
    """
    Predict next N-day price using a trained ML model (XGBoost or LSTM).
    """
    return await predict_price(ticker, horizon)


@mcp.tool()
async def backtest(ticker: str, strategy: str = "sma_cross"):
    """
    Run backtests on a selected strategy.
    """
    return await backtest_strategy(ticker, strategy)



if __name__ == "__main__":
    asyncio.run(mcp.run())
