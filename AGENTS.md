# Repository Guidelines

## Project Structure & Module Organization
Core entrypoint lives in `cmd/agentflow/`. Reusable commands sit under `internal/commands/` (`init`, `intake`, `plan`, `devplan`). Shared utilities are in `internal/config/`, `internal/langgraph/`, and `internal/prompt/`. Generated docs live in `docs/`, with `docs/output/` capturing the latest runs; avoid editing generated files directly. Helper scripts are in `scripts/`, and the mock LangGraph FastAPI server resides at `langgraph/server/app.py` with Docker assets at the repo root.

## Build, Test, and Development Commands
- `./scripts/build_cli.sh --current` builds the CLI binary into `dist/agentflow` for the host platform.
- `go run ./cmd/agentflow --help` exercises the CLI from source; swap `--help` for any command during active development.
- `go build -o dist/agentflow ./cmd/agentflow` performs a manual build when iterating on the entrypoint.
- `docker-compose up --build -d` starts the mock LangGraph service; confirm readiness via `curl localhost:8123/healthz`.
- `go test ./...` runs the full Go test suite; append `-race` or `-cover` for deeper validation.

## Coding Style & Naming Conventions
Target Go 1.22, format code with `go fmt ./...`, and lint using `go vet ./...`. Package names stay lowercase; exported identifiers use CamelCase. Favor the options pattern (`XOptions`) for constructors and return `error` rather than panicking. Keep comments purposeful and close to non-obvious logic.

## Testing Guidelines
Place `*_test.go` alongside the code under `internal/...`. Prefer table-driven cases, covering happy paths, edge conditions, and failure retries. Use `httptest` or the Docker mock server for LangGraph interactions, asserting headers, retries, and timeouts. Run `go test -race ./...` before landing changes that touch concurrency.

## Commit & Pull Request Guidelines
Adopt Conventional Commits (e.g., `feat(cli): add intake command`, `fix(langgraph): retry jitter`). PRs should state purpose, link relevant issues, document test runs (command and outcome), and highlight risk plus rollback strategy. Keep diffs focused and update `docs/` when behaviour changes.

## Security & Configuration Tips
Initialize projects with `agentflow init --project-name MyApp`, which seeds `.agentflow/config.json`. Never commit secrets; supply `LANGGRAPH_API_KEY` via the environment. Ignore generated artifacts such as `dist/` and `docs/output/` to keep the repo clean.
