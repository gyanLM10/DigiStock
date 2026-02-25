# ─────────────────────────────────────────────────────────────
# Stage 1: Build the Go CLI binary
# ─────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS go-builder

WORKDIR /build
COPY cli/go.mod cli/go.sum ./
RUN go mod download

COPY cli/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o digistock .

# ─────────────────────────────────────────────────────────────
# Stage 2: Runtime — Python 3.11 + Node.js + Go binary
# ─────────────────────────────────────────────────────────────
FROM python:3.11-slim-bookworm

# Install Node.js 20 (for npx @brightdata/mcp used by 'analyze')
RUN apt-get update && apt-get install -y --no-install-recommends curl ca-certificates && \
    curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get install -y --no-install-recommends nodejs && \
    rm -rf /var/lib/apt/lists/*

# Install uv for fast Python dependency management
RUN pip install --no-cache-dir uv

WORKDIR /app

# Copy Python source and project config
COPY pyproject.toml uv.lock ./
COPY agent_logic.py runner.py runner_tools.py xgb_model.py ./
COPY tools/ tools/
COPY models/ models/

# Install Python dependencies into /app/.venv
# resolvePython in the Go CLI will auto-detect .venv/bin/python
RUN uv sync --frozen --no-dev

# Copy the compiled Go binary from the builder stage
COPY --from=go-builder /build/digistock /usr/local/bin/digistock

# Unbuffered Python output for real-time streaming
ENV PYTHONUNBUFFERED=1

# API keys — override at runtime via --env-file or -e flags
ENV OPENAI_API_KEY=""
ENV BRIGHT_DATA_API_TOKEN=""
ENV WEB_UNLOCKER_ZONE="unblocker"
ENV BROWSER_ZONE="scraping_browser"

# Default: launch the TUI (requires: docker run -it ...)
ENTRYPOINT ["digistock"]
CMD ["tui"]
