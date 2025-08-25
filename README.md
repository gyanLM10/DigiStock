Multi-Agent Stock Analysis Chatbot 📈
This project is a real-time, streaming chatbot that uses a multi-agent AI system to provide detailed stock analysis for India's National Stock Exchange (NSE). Users can interact with a simple web interface, and the backend orchestrates a team of specialized AI agents to research and deliver structured stock recommendations.

The system is built with FastAPI for the backend, LangGraph for multi-agent orchestration, and a vanilla HTML/CSS/JS frontend. It leverages the Multi-Server Command Protocol (MCP) to connect with Bright Data, enabling agents to access live, unrestricted web data.

✨ Features
Multi-Agent Workflow: A sophisticated "assembly line" of AI agents, each with a specialized role (finding stocks, fetching data, analyzing news, and creating a final recommendation).

Live Web Data via Bright Data & MCP: Utilizes the MultiServerMCPClient to connect with Bright Data's powerful web scraping tools. This gives agents unrestricted, real-time access to market data and news, bypassing common web blocks.

Real-Time Streaming: Responses are streamed token-by-token to the user interface, providing a dynamic and responsive experience.

Structured Output: The final output is a well-formatted and easy-to-read analyst note, perfect for quick decision-making.

Modern Tech Stack: Built with FastAPI, LangGraph, LangChain, and OpenAI's GPT-4 Turbo.

🏗️ System Architecture
The application is composed of three main parts: a frontend, a backend, and the core agent logic.

1. Frontend (index.html)
A clean, single-page web interface that allows users to send queries and receive streamed responses.

Technology: Vanilla HTML, CSS, and JavaScript.

Functionality: Captures user input, sends it to the backend's /chat endpoint, and dynamically displays the streamed response.

2. Backend (backend.py)
A FastAPI server that acts as the bridge between the frontend and the AI agent system.

Technology: FastAPI.

Endpoints:

GET /: Serves the main index.html chat interface.

POST /chat: The main endpoint that accepts a user query and returns a StreamingResponse by calling the agent logic.

GET /health: A simple health check endpoint.

3. Agent Logic (agent_logic.py)
The core of the application, where the multi-agent system is defined and orchestrated using LangGraph.

The Supervisor & Agents: A central supervisor manages a team of agents (Stock_Finder, Market_Data_Analyst, News_Analyst, Trading_Advisor) to execute the research workflow step-by-step.

Tool Integration: Bright Data via MCP
The agents' ability to access live web data is the most critical component of this system, enabled by:

MultiServerMCPClient: This client from the langchain-mcp-adapters library implements the Multi-Server Command Protocol. It acts as a standardized bridge, allowing the LangGraph agents to seamlessly communicate with and command external tool servers.

Bright Data: The external tool server in this architecture. It provides enterprise-grade web scraping capabilities, including the Web Unlocker and Scraping Browser. This is essential for accessing financial websites that might otherwise block automated requests, ensuring the data fed to the agents is timely and accurate.



All results are generated in multi_agent.ipynb


