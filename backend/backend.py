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
# Logging setup
# --------------------------------------------------------------------
logger = logging.getLogger(__name__)
logging.basicConfig(level=logging.INFO)

# --------------------------------------------------------------------
# FastAPI app
# --------------------------------------------------------------------
app = FastAPI(
    title="Stock Analysis Chatbot API",
    description="An API for a multi-agent system that provides stock recommendations.",
    version="1.0.0",
)

# --------------------------------------------------------------------
# CORS (adjust allowed origins as needed)
# --------------------------------------------------------------------
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Replace "*" with ["http://localhost:3000", "https://your-frontend.com"] in prod
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# --------------------------------------------------------------------
# Templates
# --------------------------------------------------------------------
templates = Jinja2Templates(directory="templates")

# --------------------------------------------------------------------
# Request models
# --------------------------------------------------------------------
class ChatRequest(BaseModel):
    query: str

# --------------------------------------------------------------------
# Routes
# --------------------------------------------------------------------

# Root endpoint to serve the chatbot's HTML user interface
@app.get("/", response_class=HTMLResponse)
async def get_chat_ui(request: Request):
    """Serves the main chat interface."""
    return templates.TemplateResponse("index.html", {"request": request})


# Streaming chat endpoint
@app.post("/chat")
async def stream_chat_response(chat_request: ChatRequest):
    """
    Accepts a user query and streams the multi-agent system's response back to the client.
    Streams plain text chunks; front-end can consume via Fetch/ReadableStream.
    """

    async def response_generator() -> AsyncGenerator[str, None]:
        """A generator function that yields text chunks from the agent run."""
        try:
            async for chunk in stream_agent_run(chat_request.query):
                if not chunk:
                    continue
                # Each yield is one chunk of text
                yield chunk
                # Tiny sleep to make chunking visually distinct on the client (optional)
                await asyncio.sleep(0.01)
        except asyncio.CancelledError:
            # Happens when client disconnects mid-stream
            logger.info("Client disconnected while streaming response.")
            raise
        except Exception as e:
            # Log full stack trace server-side
            logger.exception("Error during agent execution")
            # Send a user-visible error message as final chunk
            yield "\n[ERROR] An internal error occurred while generating the response."

    # `text/plain` so it's easy to consume; include charset
    return StreamingResponse(
        response_generator(),
        media_type="text/plain; charset=utf-8",
    )


# Health check endpoint
@app.get("/health")
async def health_check():
    """A simple endpoint to confirm the server is running."""
    return {"status": "ok"}

