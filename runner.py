"""
runner.py — CLI entry point for the DigiStock multi-agent pipeline.
Invoked by the Go CLI as a subprocess:
    python runner.py "Your query here"
Streams agent output line-by-line to stdout.
"""

import sys
import asyncio
from agent_logic import stream_agent_run


async def main():
    if len(sys.argv) < 2:
        query = "Analyze current NSE market and recommend 2 stocks to trade today."
    else:
        query = " ".join(sys.argv[1:])

    async for chunk in stream_agent_run(query):
        print(chunk, end="", flush=True)


if __name__ == "__main__":
    asyncio.run(main())
