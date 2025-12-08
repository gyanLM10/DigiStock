# tools/backtesting.py
import pandas as pd
from .market_data import get_historical_data

async def backtest_strategy(ticker: str, strategy="sma_cross"):
    """
    Run simple SMA crossover backtest.
    """
    data = await get_historical_data(ticker, "1y")
    if "error" in data:
        return data

    df = pd.DataFrame(data)
    df["SMA_50"] = df["Close"].rolling(50).mean()
    df["SMA_200"] = df["Close"].rolling(200).mean()

    df["signal"] = 0
    df.loc[df["SMA_50"] > df["SMA_200"], "signal"] = 1

    df["strategy_return"] = df["signal"].shift(1) * df["Close"].pct_change()
    df["buy_hold_return"] = df["Close"].pct_change()

    return {
        "ticker": ticker,
        "strategy_total_return": float(df["strategy_return"].sum()),
        "buy_hold_return": float(df["buy_hold_return"].sum()),
        "improvement_over_buy_hold": float(df["strategy_return"].sum() - df["buy_hold_return"].sum())
    }
