import yfinance as yf
import pandas as pd
import numpy as np
import xgboost as xgb
import pickle
from sklearn.preprocessing import StandardScaler

# You can expand this list later
TICKERS = ["RELIANCE.NS", "TCS.NS", "INFY.NS", "HDFCBANK.NS", "ICICIBANK.NS"]

# ---------------------------------------------
# Feature Engineering: MUST MATCH MCP FEATURES
# ---------------------------------------------
def compute_features(df):
    # Simple Moving Averages
    df["SMA_50"] = df["Close"].rolling(50).mean()
    df["SMA_200"] = df["Close"].rolling(200).mean()

    # Exponential Moving Average
    df["EMA_20"] = df["Close"].ewm(span=20).mean()

    # RSI
    delta = df["Close"].diff()
    gain = (delta.where(delta > 0, 0)).rolling(14).mean()
    loss = (-delta.where(delta < 0, 0)).rolling(14).mean()
    rs = gain / loss
    df["RSI"] = 100 - (100 / (1 + rs))

    # MACD (matches your MCP indicator)
    exp1 = df["Close"].ewm(span=12, adjust=False).mean()
    exp2 = df["Close"].ewm(span=26, adjust=False).mean()
    df["MACD"] = exp1 - exp2

    df = df.dropna()
    return df

# ---------------------------------------------
# Dataset Construction
# ---------------------------------------------
X, y = [], []

print("Downloading and building dataset...")

for ticker in TICKERS:
    print(f"Fetching {ticker}...")
    df = yf.Ticker(ticker).history("2y")

    if df.empty:
        print(f"⚠️ Skipping {ticker}, no data found.")
        continue

    df = compute_features(df)

    # Predict next-day close price
    df["Target"] = df["Close"].shift(-1)
    df = df.dropna()

    # IMPORTANT — MUST MATCH MCP PREDICTOR INPUT ORDER
    features = df[[
        "Close",
        "Volume",
        "RSI",
        "SMA_50",
        "SMA_200",
        "EMA_20",
        "MACD"
    ]]

    X.extend(features.values)
    y.extend(df["Target"].values)

# Convert to arrays
X = np.array(X)
y = np.array(y)

print(f"Dataset built: {X.shape[0]} samples, {X.shape[1]} features")

# ---------------------------------------------
# Scaling
# ---------------------------------------------
scaler = StandardScaler()
X_scaled = scaler.fit_transform(X)

# ---------------------------------------------
# Train Model
# ---------------------------------------------
print("Training XGBoost model...")

model = xgb.XGBRegressor(
    n_estimators=300,
    max_depth=6,
    learning_rate=0.05,
    subsample=0.9,
    colsample_bytree=0.8,
    objective="reg:squarederror"
)

model.fit(X_scaled, y)

# ---------------------------------------------
# Save artifacts
# ---------------------------------------------
MODEL_PATH = "models/xgb_model.json"
SCALER_PATH = "models/scaler.pkl"

model.save_model(MODEL_PATH)

with open(SCALER_PATH, "wb") as f:
    pickle.dump(scaler, f)

print("\n🎉 Training complete!")
print(f"Model saved to: {MODEL_PATH}")
print(f"Scaler saved to: {SCALER_PATH}")
