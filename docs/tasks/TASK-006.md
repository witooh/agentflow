# TASK-006 — LangGraph server (Python/FastAPI): implement real server under langgraph/server/ and update Dockerfile
**Date:** 2025-09-15

## ContextEngineering
```xml
<task>LangGraph server (Python/FastAPI): implement real server under langgraph/server/ and update Dockerfile</task>
<context>
  <requirement># AgentFlow — Requirements (v2)

- REST endpoints: POST /agents/run → { runId, content }, POST /agents/questions → { status }, GET /healthz → ok
- Optional Authorization: Bearer ${LANGGRAPH_API_KEY}
- Server runs by default at :8123; configurable via PORT env
- Prefer real implementation with optional LLM call; fallback to deterministic mock when no API key
  </requirement>
  <srs># Integration/Architecture Notes
- Component: langgraph/server/app.py (FastAPI)
- Contracts stable for Go client (internal/langgraph): healthz, run, questions
- Security: if LANGGRAPH_API_KEY is set in env, server must require Authorization header matching `Bearer ${LANGGRAPH_API_KEY}` for non-health endpoints
- Extensibility: if OPENAI_API_KEY is set, server may call OpenAI to synthesize content; otherwise return deterministic scaffold
  </srs>
  <stories># Stories
- STORY-LGS-1: As a developer, I can run a real Python server that serves the LangGraph REST expected by the CLI.
- STORY-LGS-2: As a developer, I can protect the endpoints with a simple Bearer token when configured.
- STORY-LGS-3: As a developer, I can optionally enable LLM-backed generation via OPENAI_API_KEY without changing the Go client.
  </stories>
</context>
```

## Implement
- Update langgraph/server/app.py
  - Remove "mock" wording; title becomes "LangGraph Server (Python)"
  - Add `verify_auth(request)` that enforces Bearer token when LANGGRAPH_API_KEY is set
  - Add `generate_content(role, prompt)` that:
    - If `OPENAI_API_KEY` available, attempts to call OpenAI Chat Completions (model from env `OPENAI_MODEL` or default `gpt-4o-mini`) and returns content
    - Falls back to deterministic `synthesize_markdown(...)` on error or when no key
  - Keep endpoints: GET /healthz (public), POST /agents/run (auth), POST /agents/questions (auth)
- Update Dockerfile
  - Install `openai` alongside FastAPI/uvicorn
  - Keep PORT env and uvicorn entrypoint

## Subtasks
1. Implement auth guard and optional OpenAI-backed generator in app.py
2. Update Dockerfile to include openai dependency
3. Verify docker-compose up brings server at :8123 and healthz returns ok
4. Ensure Go client contracts unchanged; run `go test ./...`

## Definition of Done (DoD)
- [ ] FastAPI server enforces Bearer token when LANGGRAPH_API_KEY is set, but keeps /healthz public
- [ ] Endpoints return correct JSON shapes: RunResponse {runId, content}, QuestionsResponse {status}
- [ ] Docker image builds successfully and runs with `docker-compose up -d`
- [ ] Go tests still pass and CLI dry-run unaffected

## Risk
- OpenAI dependency may increase image size; mitigated by optional use and slim base
- External API failures: mitigated by fallback to deterministic content

## Notes
- Keep behavior deterministic when no OPENAI_API_KEY for easier local dev
- Do not change Go client code/contracts in this task

## Traceability
- Linked: TASK-007 (LangGraph integration in Go client), TASK-002 (REST client)
