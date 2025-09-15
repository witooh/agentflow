AgentFlow – Project‑Specific Development Guidelines
Date: 2025-09-15

Scope
These notes capture project‑specific knowledge needed to build, configure, test, and extend AgentFlow. They assume proficiency with Go, Docker, and general CLI development.

1) Build and Configuration
Go toolchain
- Language: Go 1.22.x (see go.mod). Modules only; no vendoring committed.
- Quick sanity build: go build ./...

CLI build script
- Preferred for reproducible builds: scripts/build_cli.sh
  - Current platform only: ./scripts/build_cli.sh --current
    - Output: ./dist/agentflow (agentflow.exe on Windows)
  - Cross‑compile default matrix: ./scripts/build_cli.sh
    - Targets: darwin/linux (amd64, arm64) and windows/amd64
  - Custom targets: ./scripts/build_cli.sh --target linux/386 --target freebsd/amd64
  - Clean output: ./scripts/build_cli.sh --clean --current
- The script sets CGO_ENABLED=0 and strips symbols (-s -w) to minimize binary size.

Runtime layout and config
- Config path: .agentflow/config.json created by agentflow init
- Defaults (internal/config.Config):
  - schemaVersion: "0.1"
  - projectName: from init flag --project-name (default: MyProject)
  - langgraph.baseUrl: from init flag --base-url (default: http://localhost:8123)
  - llm: model (default gpt-4o-mini), temperature 0.2, maxTokens 4000
  - roles: po_pm, sa, qa, dev templates
  - io: inputDir "input", outputDir "output"
  - security.envKeys: ["OPENAI_API_KEY"] (informational; not used directly yet)
  - devplan.maxContextCharsPerTask: 4000
  - askHuman.mode: "interactive"
  - metadata: owner/repo/tags optional
- Directory creation: EnsureDirs() makes .agentflow/, input/, and output/ as needed. The init command also writes .agentflow/.gitignore suggesting ignoring generated outputs.

Services and external dependencies
- LangGraph service: default http://localhost:8123 (see internal/langgraph/client.go)
  - HTTP API endpoints used:
    - POST /agents/run → { runId, content }
    - POST /agents/questions → { status }
    - GET  /healthz → 200 OK indicates healthy
  - Client behavior:
    - Retries: default 3 (maxRetries) with exponential backoff + jitter
    - Timeout: default 30s (override with WithTimeout)
    - Authorization: optional Bearer <LANGGRAPH_API_KEY>
- Local mock server (recommended for dev):
  - docker-compose up -d brings up a FastAPI mock at :8123
    - Dockerfile builds a tiny server from langgraph/server/app.py
  - Health: curl http://localhost:8123/healthz → ok

CLI entry points and flags
- agentflow help prints available commands and flags.
- Commands:
  - init: Initialize .agentflow/config.json
    - Flags: --project-name, --base-url, --model, --config
  - intake: Aggregate input/*.md into output/requirements.md
    - Flags: --config, --input, --output, --role (default po_pm), --dry-run
    - Behavior: reads markdown and txt files recursively, builds a prompt, calls LangGraph unless --dry-run; with no inputs, writes scaffold and returns ErrNoInputs (handled as warning in CLI).
  - plan: Generate srs.md, stories.md, acceptance_criteria.md from requirements.md
    - Flags: --config, --requirements, --output, --role (default sa), --dry-run
    - Behavior: reads a single requirements.md; builds inline prompt; splits model output using explicit markers (--- SRS START --- etc.); falls back to scaffold on errors.
  - devplan: Generate docs/task_list.md and docs/tasks/*.md from prior docs
    - Flags: --config, --source (default docs), --output (default docs), --role (default dev), --dry-run
    - Behavior: parses a checklist from model output, ensures a scaffold first task, assigns IDs, and generates per‑task files embedding compact XML/Markdown context (see internal/commands/devplan.go).
  - design, qa: currently stubbed (print not implemented yet).

2) Testing
How tests are intended to run
- Standard Go tests: go test ./...
- No network calls should be made in unit tests by default. Prefer:
  - Using --dry-run flags when testing command orchestration through small wrappers, or
  - Isolating pure functions (prompt/builder.go, config operations, string processing), or
  - Mocking HTTP endpoints with net/http/httptest if you need to exercise internal/langgraph Client.

Adding tests
- Package‑level guidance:
  - For unexported helpers in a package, keep the test in the same package name (not package xyz_test) if you must access unexported symbols (e.g., trimTrailingSlash in internal/langgraph).
  - Favor table‑driven tests and avoid timing‑fragile assertions (especially around backoffSleep; if needed, inject a sleeper or refactor to return durations).
- Examples of good test targets:
  - internal/prompt.BuildForRole: feed a temp input dir with small .md files and assert prompt structure and file order is stable (sorted path order).
  - internal/config.DefaultConfig + EnsureDirs + Save/Load cycle.
  - internal/langgraph.trimTrailingSlash and transient.

Worked example (validated locally)
- Example: testing internal/langgraph trimTrailingSlash and transient classification
  - File suggestion: internal/langgraph/client_test.go (package langgraph)
  - Contents:
    
    package langgraph
    
    import (
      "testing"
    )
    
    func TestTrimTrailingSlash(t *testing.T) {
      cases := map[string]string{
        "http://localhost:8123/": "http://localhost:8123",
        "http://localhost:8123":  "http://localhost:8123",
        "":                        "",
        "/api/v1///":             "/api/v1",
      }
      for in, want := range cases {
        if got := trimTrailingSlash(in); got != want {
          t.Fatalf("trimTrailingSlash(%q) = %q, want %q", in, got, want)
        }
      }
    }
    
    type dummyNetErr struct{}
    func (dummyNetErr) Error() string   { return "dummy" }
    func (dummyNetErr) Timeout() bool   { return true }
    func (dummyNetErr) Temporary() bool { return true }
    
    func TestTransientNetError(t *testing.T) {
      if !transient(dummyNetErr{}) {
        t.Fatalf("expected transient to treat net.Error with Timeout()=true as transient")
      }
    }
    
- Run: go test ./...
- Expected: internal/langgraph package tests pass quickly without network.

Integration and E2E notes
- For future E2E, spin up the mock LangGraph via docker-compose and run the CLI against it:
  - docker-compose up -d
  - ./dist/agentflow init --project-name Demo --base-url http://localhost:8123 --config .agentflow/config.json
  - echo "# Input\nHello" > input/demo.md
  - ./dist/agentflow intake --dry-run
  - ./dist/agentflow plan --dry-run
  - ./dist/agentflow devplan --dry-run --source docs --output docs
- When the real LangGraph backend is introduced, ensure LANGGRAPH_API_KEY is set if required by the service; client will forward it as Authorization: Bearer.

3) Additional Development Information
Code structure
- cmd/agentflow: CLI entry (flag parsing, command dispatch, user‑facing messages and errors)
- internal/commands: core orchestration logic per subcommand
  - intake.go: scans input dir, builds prompt via prompt.BuildForRole, writes output/requirements.md with a metadata header
  - plan.go: consumes output/requirements.md, composes architect prompt, splits model response into SRS/Stories/AC, writes distinct files with metadata
  - devplan.go: parses checklist into tasks, ensures SC-001 scaffold, assigns IDs, writes docs/task_list.md and docs/tasks/TASK-xxx.md with compact XML/MD context
- internal/config: schema, default values, validation, Save/Load, EnsureDirs
- internal/prompt: deterministic prompt construction; stable order of files; supports .md and .txt
- internal/langgraph: HTTP client with retry/timeout and a minimal healthcheck
- langgraph/server: lightweight Python FastAPI mock of LangGraph for local development

Conventions and style
- Go formatting: go fmt ./...; linters may be added later but are not required by this repo.
- Errors: sentinel errors exported from commands (e.g., ErrNoInputs, ErrNoRequirements) are surfaced for CLI‑level user messages; keep business logic returning wrapped errors with context.
- Logging: stdlib log used only in CLI boundary; most functions return errors upward.
- Filesystem: all generated artifacts are written under configurable io.outputDir (default output) or docs for devplan; avoid committing generated outputs unless explicitly required.
- i18n: Some scaffold text in plan.go is intentionally mixed (e.g., Thai headings) to signal placeholder content—do not rely on that phrasing for parsing.

Operational tips
- Dry‑run first: use --dry-run to validate flow without any network dependency.
- Determinism: prompt.BuildForRole sorts input file paths; if tests assert exact prompts, control the filenames and order.
- Resilience: client retries transient network errors (timeouts, DNS, connection issues). For tests, avoid sleeping; if backoff needs to be verified, refactor to inject a sleeper function.
- Security: Default config lists OPENAI_API_KEY in security.envKeys as informational. The only env key actually consumed by the Go client is LANGGRAPH_API_KEY (optional) for Authorization. Redaction settings exist for future use.

Maintenance checklist
- When adding a new subcommand:
  - Wire in cmd/agentflow/main.go switch + usage()
  - Implement internal/commands/<name>.go
  - Add a stub or mock path using --dry-run for offline development
  - Update docs and consider adding unit tests around pure logic
- When changing config schema:
  - Update config.DefaultConfig, Validate, and any docs that mirror the fields
  - Ensure init command still produces a usable .agentflow/config.json and that EnsureDirs behavior is preserved

Appendix: Known ports and paths
- Mock server: http://localhost:8123
- Config: .agentflow/config.json and .agentflow/.gitignore (suggested ignore for outputs)
- I/O: input/, output/ (generated), docs/ (devplan outputs)
