import pandas as pd
import numpy as np
from .market_data import get_historical_data

def _compute_rsi(series, period=14):
    delta = series.diff()
    gain = (delta.where(delta > 0, 0)).rolling(period).mean()
    loss = (-delta.where(delta < 0, 0)).rolling(period).mean()
    rs = gain / loss
    return 100 - (100 / (1 + rs))

def _compute_macd(close):
    exp1 = close.ewm(span=12, adjust=False).mean()
    exp2 = close.ewm(span=26, adjust=False).mean()
    macd = exp1 - exp2
    signal = macd.ewm(span=9, adjust=False).mean()
    return macd, signal

async def compute_indicators(ticker: str, period="6mo"):
    """
    Compute RSI, MACD, SMA, EMA and return last values.
    """
    data = await get_historical_data(ticker, period)

    if "error" in data:
        return data

    df = pd.DataFrame(data)

    df["SMA_50"] = df["Close"].rolling(50).mean()
    df["SMA_200"] = df["Close"].rolling(200).mean()
    df["EMA_20"] = df["Close"].ewm(20).mean()

    df["RSI"] = _compute_rsi(df["Close"])

    macd, signal = _compute_macd(df["Close"])
    df["MACD"] = macd
    df["MACD_SIGNAL"] = signal

    last = df.iloc[-1]

    return {
        "ticker": ticker,
        "SMA_50": float(last["SMA_50"]),
        "SMA_200": float(last["SMA_200"]),
        "EMA_20": float(last["EMA_20"]),
        "RSI": float(last["RSI"]),
        "MACD": float(last["MACD"]),
        "MACD_SIGNAL": float(last["MACD_SIGNAL"]),
        "trend": "Bullish" if last["MACD"] > last["MACD_SIGNAL"] else "Bearish"
    }
