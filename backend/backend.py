from typing import AsyncGenerator
import asyncio
import logging

from fastapi import FastAPI, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import HTMLResponse, StreamingResponse
from fastapi.templating import Jinja2Templates
from pydantic import BaseModel

from agent_logic import stream_agent_run

# --------------------------------------------------------------------
# Logging Setup
# --------------------------------------------------------------------
logger = logging.getLogger("StockChatbot")
logging.basicConfig(
    level=logging.INFO,
    format="[%(asctime)s] [%(levelname)s] %(message)s",
)

logger.info("Starting Stock Analysis Chatbot Backend...")

# --------------------------------------------------------------------
# FastAPI App
# --------------------------------------------------------------------
app = FastAPI(
    title="Multi-Agent Stock Analysis API",
    description=(
        "Real-time NSE stock analysis powered by LangGraph multi-agent workflows, "
        "Bright Data MCP tools, and a custom Stock MCP server with ML predictions."
    ),
    version="2.0.0",
)

# --------------------------------------------------------------------
# CORS (configure for frontend deployment later)
# --------------------------------------------------------------------
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# --------------------------------------------------------------------
# Jinja Templates
# --------------------------------------------------------------------
templates = Jinja2Templates(directory="templates")

# --------------------------------------------------------------------
# Request Models
# --------------------------------------------------------------------
class ChatRequest(BaseModel):
    query: str


# --------------------------------------------------------------------
# Routes
# --------------------------------------------------------------------

@app.get("/", response_class=HTMLResponse)
async def get_chat_ui(request: Request):
    """
    Serves the main HTML interface for chatting with the multi-agent system.
    """
    logger.info("Serving chat UI.")
    return templates.TemplateResponse("index.html", {"request": request})


@app.post("/chat")
async def stream_chat_response(chat_request: ChatRequest):
    """
    POST /chat
    Accepts a query and streams multi-agent reasoning output in real time.
    """

    async def response_generator() -> AsyncGenerator[str, None]:
        logger.info(f"Received query: {chat_request.query}")

        try:
            # Stream agent output chunk-by-chunk
            async for chunk in stream_agent_run(chat_request.query):
                if chunk:
                    yield chunk
                await asyncio.sleep(0.005)  # smoother streaming

        except asyncio.CancelledError:
            logger.warning("Client disconnected during stream.")
            raise

        except Exception as e:
            logger.exception("Critical backend error during stream")
            yield "\n[ERROR] A server-side error occurred while generating the response."

    return StreamingResponse(
        response_generator(),
        media_type="text/plain; charset=utf-8",
    )


@app.get("/health")
async def health_check():
    """
    Simple health endpoint for deployment environments.
    Can be expanded to check:
    - MCP connectivity
    - Model existence
    - Agent system readiness
    """
    return {
        "status": "ok",
        "agents": "ready",
        "mcp_servers": ["BrightData", "StockMCP"],
        "model": "xgboost",
    }

