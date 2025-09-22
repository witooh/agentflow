# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

AgentFlow is a Go-based CLI that orchestrates specialized AI agents to transform project requirements through a structured pipeline. It uses the `openai-agents-go` library to create role-based agents (Product Owner, Solution Architect, Lead Developer, Lead QA) that work together to produce comprehensive project documentation.

## Development Commands

### Building
```bash
# Build for current platform
./scripts/build_cli.sh --current

# Build for all platforms
./scripts/build_cli.sh

# Cross-platform build with specific targets
./scripts/build_cli.sh --target linux/386 --target freebsd/amd64
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with race detection
go test -race ./...

# Run tests with coverage
go test -cover ./...
```

### Code Quality
```bash
# Format code
go fmt ./...

# Vet code
go vet ./...

# Format and vet (typical pre-commit)
go fmt ./... && go vet ./...
```

### Running from Source
```bash
# Run CLI directly
go run ./cmd/agentflow --help

# Example workflow
go run ./cmd/agentflow init --project-name MyApp
```

## Architecture

### Agent System
The core architecture revolves around four specialized AI agents defined in `internal/agents/`:

- **Product Owner (PO)**: Converts raw inputs into formal requirements (`intake` command)
- **Solution Architect (SA)**: Creates technical specifications, architecture, and UML (`plan`, `design`, `uml` commands)
- **Lead Developer (LD)**: Breaks down features into development tasks (`devplan` command)
- **Lead QA (LQ)**: Creates comprehensive testing strategies (`qa` command)

Each agent uses role-specific prompts embedded as Markdown files in `internal/commands/` and operates with configurable LLM settings.

### Configuration System
- `.agentflow/config.json` - Main configuration file with agent roles, LLM settings, and I/O paths
- Environment variable overrides supported via `AGENTFLOW_*` variables
- `OPENAI_API_KEY` required for agent operation

### Command Pipeline
Commands are designed to work in sequence, each building on previous outputs:
```
Input Files → [PO] → requirements.md → [SA] → srs.md, stories.md, acceptance_criteria.md → [SA] → architecture.md, uml.md → [LQ] → test-plan.md → [LD] → task_list.md, tasks/*.md
```

### Project Structure
```
cmd/agentflow/          # CLI entrypoint and main()
internal/commands/      # Command implementations and embedded prompts
internal/config/        # Configuration loading and validation
internal/agents/        # Agent initialization and management
internal/langgraph/     # HTTP client for LLM backends (legacy)
internal/prompt/        # Prompt template utilities
scripts/               # Build and tooling scripts
docs/                  # Generated documentation outputs
.agentflow/            # Project-specific config and I/O directories
```

## Key Implementation Details

### Agent Creation
Agents are created using `newAgent(role, description, model)` and configured with:
- File creation/reading tools
- Temperature, token limits, and model settings from config
- Role-specific instructions from the `roles` configuration

### Dry Run Mode
All commands support `--dry-run` to scaffold outputs without API calls, useful for:
- Testing configuration
- Generating template structures
- Development without consuming API credits

### Error Handling
- Commands provide fallback scaffold content when agent calls fail
- Configuration validation ensures required fields and ranges
- Environment variable validation for required API keys

### Template System
Commands use Go template syntax in embedded Markdown prompts with variables like:
- `{{.RequirementsPath}}` - Input file paths
- `{{.ProjectName}}` - From configuration
- `{{.SrsPath}}`, `{{.ArchitecturePath}}` - Output paths

## Dependencies

- Go 1.24.3+ (with toolchain 1.24.6)
- `github.com/nlpodyssey/openai-agents-go` - Core agent framework
- `github.com/openai/openai-go/v2` - OpenAI client
- Standard library for CLI, JSON, file operations

## Environment Variables

Required:
- `OPENAI_API_KEY` - OpenAI API access

Optional overrides:
- `AGENTFLOW_MODEL` - Override LLM model
- `AGENTFLOW_TEMPERATURE` - Override temperature setting
- `AGENTFLOW_MAX_TOKENS` - Override token limit
- `AGENTFLOW_INPUT_DIR` - Override input directory
- `AGENTFLOW_OUTPUT_DIR` - Override output directory

## Testing Notes

- Unit tests exist for core components (`*_test.go` files)
- Integration tests can use `--dry-run` mode to avoid API calls
- Configuration validation has comprehensive test coverage
- Agent initialization and basic functionality are tested