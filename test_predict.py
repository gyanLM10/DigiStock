from tools.predictions import predict_price
import asyncio

async def main():
    result = await predict_price("RELIANCE.NS", horizon=5)
    print(result)

asyncio.run(main())
