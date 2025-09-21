# AGENTS.md

## Overview

AgentFlow is a CLI tool that orchestrates specialized AI agents to transform project requirements through a structured pipeline. The system uses role-based agents that work together to produce comprehensive project documentation, from initial requirements to development tasks.

## Agent Architecture

### Core Agent Framework

AgentFlow uses the `openai-agents-go` library to create specialized agents that operate within defined roles. Each agent is configured with:

- **Role-specific instructions** - Defines the agent's purpose and behavior
- **LLM Model** - Currently configured for GPT-5 by default
- **Built-in tools** - File creation and reading capabilities
- **Model settings** - Temperature, token limits, and other parameters

### Agent Implementation

```go
// Core agent structure
type Agent struct {
    Agent *agents.Agent
}

// Global agent instances
var (
    PO *Agent  // Product Owner
    SA *Agent  // Solution Architect
    LD *Agent  // Lead Developer
    LQ *Agent  // Lead QA
)
```

## Available Agents

### 1. Product Owner (PO)
**Role**: `po_pm`
**Purpose**: Transforms raw input into formal requirements
**Model**: `gpt-5`
**Temperature**: `1.0`

**Responsibilities**:
- Convert input context into formal requirements
- Structure requirements with sections: Goals, Scope, FR, NFR, Assumptions, Open Questions
- Aggregate multiple input sources into coherent requirements

**Usage**:
- Used by the `intake` command
- Processes markdown files from input directory
- Generates `requirements.md`

**Configuration**:
```json
{
  "roles": {
    "po_pm": "You are a PO/PM. Convert input context into formal requirements with sections: Goals, Scope, FR, NFR, Assumptions, Open Questions."
  }
}
```

### 2. Solution Architect (SA)
**Role**: `sa`
**Purpose**: Transforms requirements into technical specifications
**Model**: `gpt-5`
**Temperature**: `1.0`

**Responsibilities**:
- Convert requirements into Software Requirements Specification (SRS)
- Generate user stories following INVEST principles
- Create acceptance criteria for each story
- Design system architecture and component relationships
- Generate UML diagrams and technical documentation

**Usage**:
- Used by `plan`, `design`, `uml`, and `devplan` commands
- Processes requirements to create technical specifications
- Generates multiple output files: `srs.md`, `stories.md`, `acceptance_criteria.md`, `architecture.md`

**Configuration**:
```json
{
  "roles": {
    "sa": "You are a Solution Architect. Transform requirements into SRS/Stories/AC."
  }
}
```

### 3. Lead Developer (LD)
**Role**: `dev`
**Purpose**: Creates development plans and task breakdowns
**Model**: `gpt-5`
**Temperature**: `1.0`

**Responsibilities**:
- Analyze technical specifications and architecture
- Break down features into development tasks
- Create detailed task context for implementation
- Generate development timeline and dependencies

**Usage**:
- Used by the `devplan` command
- Processes all prior documentation (requirements, SRS, stories, architecture)
- Generates `task_list.md` and individual task files in `tasks/` directory

**Configuration**:
```json
{
  "roles": {
    "dev": "You are a Tech Lead. Produce dev task list and per-task context."
  }
}
```

### 4. Lead QA (LQ)
**Role**: `qa`
**Purpose**: Creates comprehensive testing strategies
**Model**: `gpt-5`
**Temperature**: `1.0`

**Responsibilities**:
- Analyze requirements and acceptance criteria
- Design test strategies and test cases
- Create quality assurance plans
- Define testing methodologies and coverage

**Usage**:
- Used by the `qa` command
- Processes SRS, stories, and acceptance criteria
- Generates `test-plan.md`

**Configuration**:
```json
{
  "roles": {
    "qa": "You are a QA Lead. Produce a concise test plan."
  }
}
```

## Agent Tools

### File Creator Tool
**Function**: `file_creator`
**Purpose**: Create files with specified content
**Parameters**:
- `Path`: Target file path
- `Content`: File content to write

### File Reader Tool
**Function**: `file_reader`
**Purpose**: Read contents of existing files
**Parameters**:
- `Path`: Source file path to read

## Agent Workflow Pipeline

The agents work in a structured pipeline where each stage builds upon the previous:

```
1. [PO Agent] Input Files → requirements.md
2. [SA Agent] requirements.md → srs.md, stories.md, acceptance_criteria.md
3. [SA Agent] Prior docs → architecture.md, uml.md
4. [LQ Agent] Prior docs → test-plan.md
5. [LD Agent] All docs → task_list.md, tasks/*.md
```

## Configuration

### Agent Configuration File
Location: `.agentflow/config.json`

```json
{
  "schemaVersion": "0.1",
  "projectName": "MyProject",
  "llm": {
    "model": "gpt-5",
    "temperature": 0.2,
    "maxTokens": 4000
  },
  "roles": {
    "po_pm": "You are a PO/PM. Convert input context into formal requirements with sections: Goals, Scope, FR, NFR, Assumptions, Open Questions.",
    "sa": "You are a Solution Architect. Transform requirements into SRS/Stories/AC.",
    "qa": "You are a QA Lead. Produce a concise test plan.",
    "dev": "You are a Tech Lead. Produce dev task list and per-task context."
  },
  "io": {
    "inputDir": ".agentflow/input",
    "outputDir": ".agentflow/output"
  },
  "security": {
    "envKeys": ["OPENAI_API_KEY"]
  },
  "devplan": {
    "maxContextCharsPerTask": 4000
  }
}
```

### Environment Variables
- `OPENAI_API_KEY`: Required for agent operation
- `AGENTFLOW_MODEL`: Override default LLM model
- `AGENTFLOW_TEMPERATURE`: Override temperature setting
- `AGENTFLOW_MAX_TOKENS`: Override token limit
- `AGENTFLOW_INPUT_DIR`: Override input directory
- `AGENTFLOW_OUTPUT_DIR`: Override output directory

## Usage Examples

### Initialize Project with Agents
```bash
# Initialize configuration
agentflow init --project-name "MyApp" --model "gpt-5"

# Verify agent configuration
cat .agentflow/config.json
```

### Run Agent Pipeline
```bash
# 1. PO Agent: Convert inputs to requirements
agentflow intake --input .agentflow/input --output .agentflow/output

# 2. SA Agent: Generate technical specifications
agentflow plan --requirements .agentflow/output/requirements.md

# 3. SA Agent: Create architecture documents
agentflow design --source .agentflow/output

# 4. LQ Agent: Generate test plan
agentflow qa --source .agentflow/output

# 5. LD Agent: Create development tasks
agentflow devplan --source .agentflow/output
```

### Dry Run Mode
All commands support `--dry-run` to test agent configuration without making API calls:

```bash
agentflow plan --dry-run --requirements requirements.md
agentflow design --dry-run --source docs/
```

## Agent Prompt Engineering

### Prompt Templates
Each command uses embedded Markdown templates that define agent behavior:

- `plan_prompt.md`: SA agent instructions for planning phase
- `design_prompt.md`: SA agent instructions for architecture design
- `devplan_prompt.md`: LD agent instructions for development planning
- `uml_prompt.md`: SA agent instructions for UML generation

### Template Variables
Prompt templates use Go template syntax with variables like:
- `{{.RequirementsPath}}`: Path to requirements file
- `{{.SrsPath}}`: Path to SRS output
- `{{.ArchitecturePath}}`: Path to architecture output
- `{{.ProjectName}}`: Project name from config

## Best Practices

### Agent Role Consistency
- Maintain clear role boundaries between agents
- Ensure each agent focuses on its domain expertise
- Avoid overlapping responsibilities across agents

### Configuration Management
- Use environment variables for sensitive data (API keys)
- Version control your `.agentflow/config.json`
- Test configuration changes with `--dry-run`

### File Organization
```
.agentflow/
├── config.json          # Agent configuration
├── input/               # Input files for PO agent
│   ├── feature1.md
│   └── requirements.md
└── output/              # Agent-generated outputs
    ├── requirements.md      # PO agent output
    ├── srs.md              # SA agent output
    ├── stories.md          # SA agent output
    ├── acceptance_criteria.md # SA agent output
    ├── architecture.md     # SA agent output
    ├── uml.md             # SA agent output
    ├── test-plan.md       # LQ agent output
    ├── task_list.md       # LD agent output
    └── tasks/             # LD agent task details
        ├── task-001.md
        └── task-002.md
```

### Error Handling
- Agents provide fallback scaffold content when API calls fail
- Use appropriate error handling for missing input files
- Monitor agent outputs for quality and completeness

## Troubleshooting

### Common Issues

**Agent API Failures**:
- Verify `OPENAI_API_KEY` is set correctly
- Check model availability (gpt-5 may require access)
- Review rate limits and quotas

**Missing Input Files**:
- Ensure required input files exist before running agents
- Use `--dry-run` to test without API calls
- Check file paths and permissions

**Configuration Errors**:
- Validate JSON syntax in config file
- Ensure all required fields are present
- Test with minimal configuration first

**Output Quality Issues**:
- Adjust temperature settings for more deterministic output
- Modify role descriptions for better behavior
- Increase token limits for longer outputs

### Debugging Commands
```bash
# Check configuration validity
agentflow plan --dry-run

# Verify file paths
ls -la .agentflow/input/
ls -la .agentflow/output/

# Test with minimal input
echo "# Test Requirements" > .agentflow/input/test.md
agentflow intake --dry-run
```

## Extension and Customization

### Adding New Agents
To add a new agent role:

1. Define the role in `config.json`:
```json
{
  "roles": {
    "custom_role": "You are a Custom Agent. Perform specific tasks..."
  }
}
```

2. Create agent instance in code:
```go
var CustomAgent = newAgent("Custom Agent", "", "gpt-5")
```

3. Implement command logic to use the agent:
```go
result, err := agents.CustomAgent.RunInputs(context.Background(), prompts)
```

### Customizing Prompt Templates
- Modify embedded `.md` files in `internal/commands/`
- Update template variables as needed
- Test changes with `--dry-run` mode

### Model Configuration
Agents support different models and settings:
```json
{
  "llm": {
    "model": "gpt-4",
    "temperature": 0.1,
    "maxTokens": 8000
  }
}
```

## Integration with Development Workflow

### CI/CD Integration
```yaml
# Example GitHub Actions workflow
name: AgentFlow Documentation
on:
  push:
    paths: ['.agentflow/input/**']
jobs:
  generate-docs:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run AgentFlow Pipeline
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
        run: |
          agentflow intake
          agentflow plan
          agentflow design
          agentflow qa
          agentflow devplan
```

### Version Control
- Track agent inputs and outputs
- Use semantic versioning for configuration changes
- Document agent behavior changes in commit messages

## Monitoring and Analytics

### Agent Performance
- Monitor API call success rates
- Track output quality and consistency
- Measure documentation generation time

### Usage Metrics
- Count of documents generated per agent
- Most frequently used agent roles
- Common failure patterns and resolutions

---

*This documentation describes AgentFlow v0.1.0. For the latest updates, see the project repository.*