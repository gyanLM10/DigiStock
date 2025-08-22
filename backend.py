from fastapi import FastAPI, Request
from fastapi.responses import HTMLResponse, StreamingResponse
from fastapi.templating import Jinja2Templates
from pydantic import BaseModel
import asyncio

from agent_logic import stream_agent_run

# Initialize the FastAPI application
app = FastAPI(
    title="Stock Analysis Chatbot API",
    description="An API for a multi-agent system that provides stock recommendations.",
    version="1.0.0"
)

# Setup templates for serving the HTML frontend
templates = Jinja2Templates(directory="templates")

# Pydantic model for validating the request body
class ChatRequest(BaseModel):
    query: str

# Root endpoint to serve the chatbot's HTML user interface
@app.get("/", response_class=HTMLResponse)
async def get_chat_ui(request: Request):
    """Serves the main chat interface."""
    return templates.TemplateResponse("index.html", {"request": request})

# The streaming chat endpoint
@app.post("/chat")
async def stream_chat_response(request: ChatRequest):
    """
    Accepts a user query and streams the multi-agent system's response back to the client.
    """
    async def response_generator():
        """A generator function that yields text chunks from the agent run."""
        try:
            async for chunk in stream_agent_run(request.query):
                yield chunk
                await asyncio.sleep(0.05) # Small delay to ensure chunks are sent distinctly
        except Exception as e:
            error_message = f"An error occurred during agent execution: {str(e)}"
            print(error_message) # Log the error to the server console
            yield error_message

    return StreamingResponse(response_generator(), media_type="text/plain")

# Health check endpoint
@app.get("/health")
async def health_check():
    """A simple endpoint to confirm the server is running."""
    return {"status": "ok"}