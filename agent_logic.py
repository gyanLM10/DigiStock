import os
import asyncio
from dotenv import load_dotenv
from langchain_mcp_adapters.client import MultiServerMCPClient
from langgraph.prebuilt import create_react_agent
from langchain_openai import ChatOpenAI
from langgraph_supervisor import create_supervisor
from langchain_core.messages import convert_to_messages

load_dotenv()


def pretty_format_message(message, indent=False):
    """Formats a message into a plain text string instead of printing."""
    pretty_message = message.pretty_repr(html=False)  # Use plain text for backend streaming
    if not indent:
        return pretty_message
    return "\n".join("\t" + c for c in pretty_message.split("\n"))


def pretty_format_messages(update, last_message=False):
    """Formats a full update chunk from the graph into a single string for streaming."""
    output_lines = []
    is_subgraph = False
    if isinstance(update, tuple):
        ns, update = update
        if len(ns) == 0:
            return ""  # Skip empty updates
        graph_id = ns[-1].split(":")[0]
        output_lines.append(f"--- 🧑‍💻 Calling Sub-Agent: {graph_id} ---\n")
        is_subgraph = True

    for node_name, node_update in update.items():
        # Only show updates from the main supervisor or sub-agent entrypoints
        if 'messages' not in node_update or not node_update['messages']:
            continue

        update_label = f"--- ✅ Responding as Supervisor: {node_name} ---\n"
        if is_subgraph:
            update_label = f"\t--- 🗣️ Response from {node_name} ---\n"
        output_lines.append(update_label)

        messages = convert_to_messages(node_update["messages"])
        if last_message:
            messages = messages[-1:]
        for m in messages:
            output_lines.append(pretty_format_message(m, indent=is_subgraph))
        output_lines.append("\n")

    return "\n".join(output_lines)


async def stream_agent_run(query: str):
    """
    Runs the multi-agent supervisor for NSE stock analysis and yields updates for streaming.
    """
    async with MultiServerMCPClient(
            {
                "bright_data": {
                    "command": "npx",
                    "args": ["@brightdata/mcp"],
                    "env": {
                        "API_TOKEN": os.getenv("BRIGHT_DATA_API_TOKEN"),
                        "WEB_UNLOCKER_ZONE": os.getenv("WEB_UNLOCKER_ZONE", "unblocker"),
                        "BROWSER_ZONE": os.getenv("BROWSER_ZONE", "scraping_browser"),
                    },
                    "transport": "stdio",
                },
            }
    ) as client:
        tools = await client.get_tools()

        # --- Base model ---
        model = ChatOpenAI(model="gpt-4-turbo", temperature=0)

        # --- Agents ---
        stock_finder_agent = create_react_agent(
            model,
            tools,
            messages_modifier=(
                "You are a stock research analyst specializing in the Indian Stock Market (NSE). "
                "Select EXACTLY 2 actively traded NSE-listed stocks for short-term trading (buy/sell). "
                "Use recent performance, news buzz, volume, or technical strength. Avoid penny/illiquid stocks. "
                "Output ONLY a compact list with: Stock Name and NSE ticker (e.g., TCS, RELIANCE) and 1-line rationale."
            ),
        )

        market_data_agent = create_react_agent(
            model,
            tools,
            messages_modifier=(
                "You are a market data analyst for NSE stocks. Given a list of NSE tickers (e.g., RELIANCE, INFY), "
                "gather the following for EACH stock (currency INR):\n"
                "- Current price\n- Previous close\n- Today's volume\n- 7-day and 30-day price trend (up/down/flat with % approx)\n"
                "- Technical indicators: RSI(14), 50/200-day moving averages level (above/below), MACD signal (bullish/bearish/flat)\n"
                "- Any notable SPIKES in volume or volatility (mention ratio vs 30-day average if possible)\n"
                "Return data in a concise, structured bullet list per stock. If data is unavailable, write 'N/A'."
            ),
        )

        news_analyst_agent = create_react_agent(
            model,
            tools,
            messages_modifier=(
                "You are a financial news analyst for NSE stocks. For each given ticker/name, in the PAST 3–5 DAYS:\n"
                "- Find the most relevant headlines\n- Summarize key updates in 2–4 bullets\n"
                "- Label each stock's overall news sentiment: Positive / Negative / Neutral\n"
                "- Briefly note the short-term price impact (bullish/bearish/unclear)\n"
                "Keep it concise, factual, and clearly attributed per stock. If no fresh news, state 'No major updates'."
            ),
        )

        price_recommender_agent = create_react_agent(
            model,
            tools,
            messages_modifier=(
                "You are a trading strategy advisor for Indian NSE stocks. You are given:\n"
                "- Market data (price, volume, trends, RSI/MAs/MACD)\n"
                "- News summaries and sentiment\n\n"
                "For EACH stock, produce a concise analyst note in EXACTLY this format:\n\n"
                "**{Stock Name} ({Ticker})**\n\n"
                "**Recommendation:** <Buy|Sell|Hold>\n"
                "**Target Price:** INR <number>  # if Hold, you may omit or give a range\n"
                "**Reason:** <1–2 lines summarizing core rationale>\n\n"
                "--- Analysis ---\n"
                "**Technical:** <50/200-day MA position, RSI number + state (overbought/oversold/neutral), MACD signal>\n"
                "**Risk Level:** <Low|Medium|High> (1 short phrase on why)\n"
                "**Time Horizon:** <e.g., 1-3 Weeks>\n"
                "**Volume & Sentiment:** <volume vs avg, overall news sentiment>\n"
                "**Valuation:** <P/E vs sector avg if available, dividend yield>\n"
                "**Event Triggers:** <upcoming earnings, policy, sector catalysts>\n\n"
                "RULES:\n"
                "- Currency must be INR.\n"
                "- If a field is unavailable, write 'N/A'.\n"
                "- Keep each section to a single line.\n"
                "- Be practical for the NEXT trading day; keep target sensible.\n"
            ),
        )

        # --- Supervisor ---
        supervisor = create_supervisor(
            model=model,
            agents={
                "Stock_Finder": stock_finder_agent,
                "Market_Data_Analyst": market_data_agent,
                "News_Analyst": news_analyst_agent,
                "Trading_Advisor": price_recommender_agent,
            },
            system_message=(
                "You are a supervisor managing a team of financial agents for the Indian stock market (NSE).\n"
                "Follow this exact workflow:\n"
                "1. Call `Stock_Finder` to get two promising stocks.\n"
                "2. Call `Market_Data_Analyst` with the chosen stocks to get technical and price data.\n"
                "3. Call `News_Analyst` with the same stocks to get recent news and sentiment.\n"
                "4. Collate all information from the previous steps and pass it to the `Trading_Advisor`.\n"
                "5. The `Trading_Advisor` will produce the final, structured report for the user. This is the final step.\n"
                "Assign work to ONE agent at a time. The user's query is the starting signal; the plan is fixed. Do not deviate."
            ),
        ).compile()

        # --- Stream results ---
        async for chunk in supervisor.astream(
                {"messages": [{"role": "user", "content": query}]}
        ):
            formatted_chunk = pretty_format_messages(chunk, last_message=True)
            if formatted_chunk.strip():
                yield formatted_chunk