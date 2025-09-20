# AgentFlow

AgentFlow is a Go-based CLI that streamlines generating and validating software planning artifacts with help from large language models. It orchestrates intake, planning, design, QA, and development task generation using LangGraph-backed prompts.

## Features
- Opinionated workflow commands (`init`, `intake`, `plan`, `design`, `uml`, `qa`, `devplan`) for end-to-end project scoping.
- Configurable LangGraph client with retry, auth header, and timeout handling.
- Scriptable builds for multi-platform binaries.

## Getting Started
### Prerequisites
- Go 1.22+
- `OPENAI_API_KEY` environment variable when talking to a live backend

### Build the CLI
```bash
./scripts/build_cli.sh --current
```
Binaries are emitted in `dist/agentflow` (platform-specific name when cross-compiling).

### Run from Source
```bash
go run ./cmd/agentflow --help
```

### Initialize Configuration
```bash
./dist/agentflow init --project-name MyApp
```
This creates `.agentflow/config.json`. Customize the project name, default model (`--model`), and config path if desired. Generated docs and inputs live under `.agentflow/` by default.

## Typical Workflow
1. **Collect inputs**: place project notes as Markdown inside `.agentflow/input/`.
2. **Aggregate requirements**: `agentflow intake --input .agentflow/input` → generates `requirements.md`.
3. **Produce planning docs**: `agentflow plan` → emits `srs.md`, `stories.md`, `acceptance_criteria.md`.
4. **Design deliverables**: `agentflow design` and `agentflow uml` create `architecture.md` and `uml.md`.
5. **Quality plan**: `agentflow qa` writes `test-plan.md`.
6. **Dev tasking**: `agentflow devplan` creates task lists with supporting context.
7. Use `--dry-run` on any command to scaffold output without contacting the LLM backend.

## Testing & QA
- Run unit tests: `go test ./...`
- Suggested extras: `go test -race ./...` or `go test -cover ./...`
- Format & vet before committing: `go fmt ./... && go vet ./...`

## Project Layout
- `cmd/agentflow/` – CLI entrypoint and flag wiring.
- `internal/commands/` – command implementations (`init`, `intake`, `plan`, `devplan`, etc.).
- `internal/config/`, `internal/langgraph/`, `internal/prompt/` – configuration loader, HTTP client, and prompt builders.
- `docs/` – generated/reference docs; `docs/output/` contains the latest run artifacts.
- `scripts/` – helper scripts (build CLI binaries, tooling helpers).

## Contributing
Follow Conventional Commits (e.g., `feat(cli): add intake command`). Keep diffs focused and update docs if behavior changes. Run tests and formatting commands before opening a PR.

## License
Licensed under the MIT License. See `LICENSE` for details.
