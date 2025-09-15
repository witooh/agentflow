# Repository Guidelines

## Project Structure & Modules
- `cmd/agentflow/`: CLI entrypoint (`main.go`).
- `internal/commands/`: subcommands (`init`, `intake`, `plan`, `devplan`).
- `internal/config/`, `internal/langgraph/`, `internal/prompt/`: config, HTTP client, and prompt builder.
- `docs/`: generated and reference docs; `output/` holds current run outputs.
- `scripts/`: helper scripts (e.g., `build_cli.sh`).
- `langgraph/server/app.py`: FastAPI mock for the LangGraph service.
- Docker: `docker-compose.yml`, `Dockerfile` for the mock server.

## Build, Test, and Run
- Build CLI (current platform): `./scripts/build_cli.sh --current` → `dist/agentflow`.
- Run from source: `go run ./cmd/agentflow --help`
- Manual build: `go build -o dist/agentflow ./cmd/agentflow`
- Start mock LangGraph: `docker-compose up --build -d` (health: `curl localhost:8123/healthz`).
- Tests: `go test ./...` (add tests as described below).

## Coding Style & Naming
- Go 1.22; format and vet before pushing: `go fmt ./... && go vet ./...`.
- Package names lower-case; exported identifiers use CamelCase.
- Options pattern: define `XOptions` structs; return `error` (no panics in libs).
- Errors: sentinel variables (e.g., `var ErrNoInputs = errors.New("...")`).

## Testing Guidelines
- Place `*_test.go` alongside code under `internal/...`.
- Prefer table-driven tests; include edge cases and error paths.
- Useful invocations: `go test -race ./...` and `go test -cover ./...`.
- For HTTP calls, test with the mock server or stub the client.

## Commit & PR Guidelines
- Use Conventional Commits (e.g., `feat(cli): add intake command`, `fix(langgraph): retry jitter`).
- PRs must include: purpose, linked issues, test notes (commands/output), and risk/rollback.
- Keep diffs focused; update docs if behavior changes (`docs/`, usage text).

## Security & Configuration
- Initialize config: `agentflow init --project-name MyApp` → `.agentflow/config.json`.
- Do not commit secrets; `LANGGRAPH_API_KEY` is read from the environment.
- Consider ignoring generated artifacts (`output/`, `dist/`); `.agentflow/.gitignore` is scaffolded.

## LangGraph Client
- Location: `internal/langgraph/client.go`
- Endpoints: `POST /agents/run` (RunRequest → RunResponse), `POST /agents/questions` (QuestionsRequest → QuestionsResponse), `GET /healthz`.
- Behavior: HTTP client with context/timeout, Authorization header if `LANGGRAPH_API_KEY` is set, and retry with exponential backoff + jitter for transient and non-2xx responses.
- Testing: see `internal/langgraph/client_test.go` for retry, auth header, and healthcheck tests using `httptest`.
