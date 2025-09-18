package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"agentflow/internal/agents"
	"agentflow/internal/config"
)

type DesignOptions struct {
	ConfigPath string
	SourceDir  string // where to read prior docs (srs/stories/acceptance_criteria). If empty, use cfg.IO.OutputDir
	OutputDir  string // where to write architecture.md and uml.md
	Role       string
	DryRun     bool
}

func Design(opts DesignOptions) error {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	cfg.ApplyEnv()
	if opts.OutputDir != "" {
		cfg.IO.OutputDir = opts.OutputDir
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	if err := config.EnsureDirs(opts.ConfigPath, cfg); err != nil {
		return err
	}

	sourceDir := opts.SourceDir
	if strings.TrimSpace(sourceDir) == "" {
		sourceDir = cfg.IO.OutputDir
	}

	// Gather context files if present
	var ctxParts []string
	for _, name := range []string{"requirements.md", "srs.md", "stories.md", "acceptance_criteria.md"} {
		p := filepath.Join(sourceDir, name)
		if b, err := os.ReadFile(p); err == nil {
			ctxParts = append(ctxParts, fmt.Sprintf("# File: %s\n\n%s", name, string(b)))
		}
	}
	ctxContent := strings.TrimSpace(strings.Join(ctxParts, "\n\n"))

	role := opts.Role
	if role == "" {
		role = "sa"
	}
	tpl := cfg.Roles[role]
	if strings.TrimSpace(tpl) == "" {
		tpl = "You are a Solution Architect. Produce an architecture overview with a Project Structure section and PlantUML component and deployment diagrams, and a separate UML document with sequence, class, and activity diagrams."
	}

	prompt := strings.TrimSpace(strings.Join([]string{
		"SYSTEM:\n" + strings.TrimSpace(tpl),
		"CONTEXT:\n" + ctxContent,
		"EXTRA:\nProduce two markdown documents. Delimit each with exact markers on their own lines:\n--- ARCH START ---\n...\n--- UML START ---\n...\nMake sure architecture.md contains a '## Project Structure' section and PlantUML component/deployment diagrams using ```plantuml fences. UML doc must include at least sequence, class, and activity diagrams using ```plantuml fences.",
	}, "\n\n"))

	var content string
	var runID string
	if opts.DryRun {
		content = scaffoldDesignOutput()
	} else {
		resp, err := agents.SA.Run(context.Background(), prompt)
		if err != nil {
			content = scaffoldDesignOutput() + fmt.Sprintf("\n\n> Note: OpenAI call failed, wrote scaffold instead. Error: %v\n", err)
		} else {
			// runID = resp.RunID
			content = resp
		}
	}

	arch, uml := splitDesignContent(content)
	arch = ensureArchitecture(arch)
	uml = ensureUML(uml)

	date := time.Now().Format("2006-01-02")
	archBody := "# AgentFlow — Architecture\n\n**Date:** " + date + "\n\n" + arch
	umlBody := "# AgentFlow — UML\n\n**Date:** " + date + "\n\n" + uml

	// Source path string for metadata
	sourcePath := filepath.Join(sourceDir, "{srs,stories,ac}")
	if err := writeFileWithHeader(cfg, role, runID, sourcePath, filepath.Join(cfg.IO.OutputDir, "architecture.md"), archBody); err != nil {
		return err
	}
	if err := writeFileWithHeader(cfg, role, runID, sourcePath, filepath.Join(cfg.IO.OutputDir, "uml.md"), umlBody); err != nil {
		return err
	}
	return nil
}

func scaffoldDesignOutput() string {
	return strings.Join([]string{
		"--- ARCH START ---",
		defaultArchitectureSection(),
		"--- UML START ---",
		defaultUMLSection(),
	}, "\n")
}

func splitDesignContent(s string) (string, string) {
	// Similar to plan split: find markers
	const archMark = "--- ARCH START ---"
	const umlMark = "--- UML START ---"
	idxA := strings.Index(s, archMark)
	idxU := strings.Index(s, umlMark)
	if idxA == -1 && idxU == -1 {
		return s, ""
	}
	var arch, uml string
	if idxA != -1 && idxU != -1 {
		if idxA < idxU {
			arch = strings.TrimSpace(s[idxA+len(archMark) : idxU])
			uml = strings.TrimSpace(s[idxU+len(umlMark):])
			return strings.TrimSpace(arch), strings.TrimSpace(uml)
		}
		// UML first (unexpected) but handle
		uml = strings.TrimSpace(s[idxU+len(umlMark) : idxA])
		arch = strings.TrimSpace(s[idxA+len(archMark):])
		return strings.TrimSpace(arch), strings.TrimSpace(uml)
	}
	if idxA != -1 {
		arch = strings.TrimSpace(s[idxA+len(archMark):])
		return strings.TrimSpace(arch), ""
	}
	uml = strings.TrimSpace(s[idxU+len(umlMark):])
	return "", strings.TrimSpace(uml)
}

func ensureArchitecture(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return defaultArchitectureSection()
	}
	// Ensure Project Structure section exists
	lower := strings.ToLower(s)
	if !strings.Contains(lower, "project structure") {
		s = s + "\n\n" + projectStructureSection()
	}
	// Ensure at least one PlantUML fence exists for component/deployment
	if !strings.Contains(lower, "plantuml") {
		s = s + "\n\n" + plantUMLComponentDeployment()
	}
	return s
}

func ensureUML(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return defaultUMLSection()
	}
	lower := strings.ToLower(s)
	needSeq := !strings.Contains(lower, "sequence")
	needClass := !strings.Contains(lower, "class")
	needAct := !strings.Contains(lower, "activity")
	if needSeq || needClass || needAct {
		var b strings.Builder
		b.WriteString(s)
		b.WriteString("\n\n")
		if needSeq {
			b.WriteString(sampleSequence())
			b.WriteString("\n\n")
		}
		if needClass {
			b.WriteString(sampleClass())
			b.WriteString("\n\n")
		}
		if needAct {
			b.WriteString(sampleActivity())
		}
		s = b.String()
	}
	return s
}

func defaultArchitectureSection() string {
	return strings.Join([]string{
		"## Overview",
		"CLI (Go) ↔ LangGraph (REST) ↔ LLM Provider. Documents are written under output/ by default.",
		projectStructureSection(),
		plantUMLComponentDeployment(),
	}, "\n\n")
}

func projectStructureSection() string {
	return "## Project Structure\n\n" + "```text\n" + strings.TrimSpace(projectStructureTree()) + "\n```"
}

func plantUMLComponentDeployment() string {
	return strings.Join([]string{
		"### Component Diagram",
		"```plantuml",
		"@startuml",
		"package AgentFlow {",
		"  [CLI] --> (LangGraph API)",
		"}",
		"(LangGraph API) --> (LLM Provider)",
		"@enduml",
		"```",
		"",
		"### Deployment Diagram",
		"```plantuml",
		"@startuml",
		"node DevMachine {",
		"  artifact agentflow.exe",
		"}",
		"node DockerHost {",
		"  node LangGraph {",
		"    artifact fastapi",
		"  }",
		"}",
		"agentflow.exe --> fastapi",
		"@enduml",
		"```",
	}, "\n")
}

func defaultUMLSection() string {
	return strings.Join([]string{
		sampleSequence(),
		sampleClass(),
		sampleActivity(),
	}, "\n\n")
}

func sampleSequence() string {
	return strings.Join([]string{
		"## Sequence: Intake",
		"```plantuml",
		"@startuml",
		"actor Dev",
		"Dev -> CLI: intake --dry-run",
		"CLI -> LangGraph: POST /agents/run",
		"LangGraph --> CLI: content",
		"@enduml",
		"```",
	}, "\n")
}

func sampleClass() string {
	return strings.Join([]string{
		"## Class: Core",
		"```plantuml",
		"@startuml",
		"class Config {\n  +ProjectName\n  +LLM\n}",
		"class Client {\n  +RunAgent()\n}",
		"Config <.. Client",
		"@enduml",
		"```",
	}, "\n")
}

func sampleActivity() string {
	return strings.Join([]string{
		"## Activity: DevPlan",
		"```plantuml",
		"@startuml",
		"start",
		":Read docs;",
		":Generate tasks;",
		"stop",
		"@enduml",
		"```",
	}, "\n")
}

// projectStructureTree returns a static tree suitable for docs; keep deterministic
func projectStructureTree() string {
	lines := []string{
		".",
		"├── cmd/agentflow",
		"├── internal/commands",
		"├── internal/config",
		"├── internal/langgraph",
		"├── internal/prompt",
		"├── docs",
		"│   └── tasks",
		"├── langgraph/server",
		"└── scripts",
	}
	return strings.Join(lines, "\n")
}
