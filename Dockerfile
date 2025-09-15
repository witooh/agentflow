# Dockerfile for LangGraph Python server (FastAPI)
FROM python:3.11-slim

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

# Install dependencies
RUN pip install --no-cache-dir fastapi uvicorn[standard] openai

WORKDIR /app
COPY langgraph/server/app.py /app/app.py

EXPOSE 8123

# Default command; use PORT env if provided
ENV PORT=8123
CMD ["bash", "-lc", "exec uvicorn app:app --host 0.0.0.0 --port ${PORT}"]
