import pickle
import numpy as np
import xgboost as xgb
import pandas as pd
from .market_data import get_historical_data
from .indicators import compute_indicators

# Load ML model + scaler
MODEL_PATH = "models/xgb_model.json"
SCALER_PATH = "models/scaler.pkl"

model = xgb.XGBRegressor()
model.load_model(MODEL_PATH)

with open(SCALER_PATH, "rb") as f:
    scaler = pickle.load(f)

async def predict_price(ticker: str, horizon: int = 5):
    """
    Predict the next N-day price using XGBoost model.
    """
    try:
        hist = await get_historical_data(ticker, "6mo")
        if "error" in hist:
            return hist

        df = pd.DataFrame(hist)
        indicators = await compute_indicators(ticker)

        features = np.array([
            df["Close"].iloc[-1],
            df["Volume"].iloc[-1],
            indicators["RSI"],
            indicators["SMA_50"],
            indicators["SMA_200"],
            indicators["EMA_20"],
            indicators["MACD"],
        ]).reshape(1, -1)

        scaled = scaler.transform(features)
        pred = model.predict(scaled)[0]

        return {
            "ticker": ticker,
            "horizon_days": horizon,
            "predicted_price": float(pred),
            "current_price": float(df["Close"].iloc[-1])
        }

    except Exception as e:
        return {"error": str(e)}
