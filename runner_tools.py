"""
runner_tools.py — Python dispatch entry point for the DigiStock tool subcommands.
Invoked by the Go CLI as a subprocess:

    python runner_tools.py indicators <TICKER>
    python runner_tools.py predict <TICKER> [<HORIZON_DAYS>]
    python runner_tools.py backtest <TICKER> [<STRATEGY>]

Prints structured output to stdout. Exits with code 1 on error.
"""

import sys
import json
import asyncio

from tools.indicators import compute_indicators
from tools.predictions import predict_price
from tools.backtesting import backtest_strategy
from tools.utils import normalize_ticker


def _fmt_indicators(data: dict) -> str:
    t = data["ticker"]
    lines = [
        f"📊  Indicators for {t}",
        "─" * 42,
        f"  SMA 50          : {data['SMA_50']:.2f}",
        f"  SMA 200         : {data['SMA_200']:.2f}",
        f"  EMA 20          : {data['EMA_20']:.2f}",
        f"  RSI (14)        : {data['RSI']:.2f}",
        f"  MACD            : {data['MACD']:.4f}",
        f"  MACD Signal     : {data['MACD_SIGNAL']:.4f}",
        f"  Trend           : {data['trend']}",
        "─" * 42,
    ]
    return "\n".join(lines)


def _fmt_predict(data: dict) -> str:
    lines = [
        f"🔮  Price Prediction for {data['ticker']}",
        "─" * 42,
        f"  Current Price   : ₹{data['current_price']:.2f}",
        f"  Predicted (+{data['horizon_days']}d): ₹{data['predicted_price']:.2f}",
        f"  Δ               : {((data['predicted_price'] - data['current_price']) / data['current_price']) * 100:+.2f}%",
        "─" * 42,
    ]
    return "\n".join(lines)


def _fmt_backtest(data: dict) -> str:
    strat = data["strategy_total_return"] * 100
    bh = data["buy_hold_return"] * 100
    imp = data["improvement_over_buy_hold"] * 100
    lines = [
        f"📈  Backtest Results for {data['ticker']}",
        "─" * 42,
        f"  Strategy Return : {strat:+.2f}%",
        f"  Buy & Hold      : {bh:+.2f}%",
        f"  Outperformance  : {imp:+.2f}%",
        "─" * 42,
    ]
    return "\n".join(lines)


async def main():
    if len(sys.argv) < 3:
        print("Usage: runner_tools.py <indicators|predict|backtest> <TICKER> [extra...]", file=sys.stderr)
        sys.exit(1)

    command = sys.argv[1].lower()
    ticker = normalize_ticker(sys.argv[2].upper())

    if command == "indicators":
        result = await compute_indicators(ticker)
        if "error" in result:
            print(f"Error: {result['error']}", file=sys.stderr)
            sys.exit(1)
        print(_fmt_indicators(result))

    elif command == "predict":
        horizon = int(sys.argv[3]) if len(sys.argv) > 3 else 5
        result = await predict_price(ticker, horizon)
        if "error" in result:
            print(f"Error: {result['error']}", file=sys.stderr)
            sys.exit(1)
        print(_fmt_predict(result))

    elif command == "backtest":
        strategy = sys.argv[3] if len(sys.argv) > 3 else "sma_cross"
        result = await backtest_strategy(ticker, strategy)
        if "error" in result:
            print(f"Error: {result['error']}", file=sys.stderr)
            sys.exit(1)
        print(_fmt_backtest(result))

    else:
        print(f"Unknown command: {command}. Use indicators | predict | backtest", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    asyncio.run(main())
