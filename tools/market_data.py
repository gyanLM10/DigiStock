import yfinance as yf
import pandas as pd

async def get_historical_data(ticker: str, period="6mo", interval="1d"):
    """
    Async wrapper for yfinance market data.
    """
    try:
        stock = yf.Ticker(ticker)
        df = stock.history(period=period, interval=interval)

        if df.empty:
            return {"error": f"No data for ticker {ticker}"}

        df.reset_index(inplace=True)
        return df.to_dict(orient="records")

    except Exception as e:
        return {"error": str(e)}