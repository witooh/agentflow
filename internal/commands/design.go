package commands

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

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

//go:embed design_prompt.md
var designPromptTemplate string

func Design(opts DesignOptions) error {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	cfg.ApplyEnv()
	if opts.OutputDir != "" {
		cfg.IO.OutputDir = opts.OutputDir
	}
	if opts.SourceDir == "" {
		opts.SourceDir = cfg.IO.OutputDir
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	if err := config.EnsureDirs(opts.ConfigPath, cfg); err != nil {
		return err
	}

	systemMessages, err := buildDesignSystemMessage(opts.SourceDir, cfg.IO.OutputDir)
	if err != nil {
		return err
	}

	if opts.DryRun {
		// In dry run mode, write scaffold files without making API calls
		return writeDesignScaffold(cfg.IO.OutputDir)
	}

	_, err = agents.SA.RunInputs(context.Background(), systemMessages)
	if err != nil {
		fmt.Printf("\n\n> Note: OpenAI call failed, wrote scaffold instead. Error: %v\n", err)
		// Write scaffold as fallback
		if scaffoldErr := writeDesignScaffold(cfg.IO.OutputDir); scaffoldErr != nil {
			return fmt.Errorf("API call failed and scaffold write failed: %v (original: %v)", scaffoldErr, err)
		}
	}

	return err
}

func buildDesignSystemMessage(sourceDir, outputDir string) ([]agents.TResponseInputItem, error) {
	data := struct {
		RequirementsPath string
		SrsPath          string
		StoriesPath      string
		ArchitecturePath string
	}{
		RequirementsPath: filepath.Join(sourceDir, "requirements.md."),
		SrsPath:          filepath.Join(sourceDir, "srs.md"),
		StoriesPath:      filepath.Join(sourceDir, "stories.md"),
		ArchitecturePath: filepath.Join(outputDir, "architecture.md"),
	}

	tmpl, err := template.New("design").Parse(designPromptTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse design template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("render design template: %w", err)
	}

	return agents.InputList(
		agents.SystemMessage(buf.String()),
	), nil
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
		// Provide fallback content when empty
		return `# Architecture

## Project Structure

This section describes the overall project structure and organization.

## Components

Key components and their relationships.

## PlantUML Diagrams

` + "```plantuml\n@startuml\n!theme plain\ntitle System Architecture\n@enduml\n```"
	}

	lower := strings.ToLower(s)
	var additions []string

	// Check if Project Structure section exists
	if !strings.Contains(lower, "project structure") {
		additions = append(additions, "\n## Project Structure\n\nThis section describes the overall project structure and organization.")
	}

	// Check if PlantUML section exists
	if !strings.Contains(lower, "plantuml") {
		additions = append(additions, "\n## PlantUML Diagrams\n\n```plantuml\n@startuml\n!theme plain\ntitle System Architecture\n@enduml\n```")
	}

	if len(additions) > 0 {
		return s + strings.Join(additions, "")
	}

	return s
}

func ensureUML(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		// Provide fallback content when empty
		return `# UML Diagrams

## Sequence: User Interactions

` + "```plantuml\n@startuml\nactor User\nUser -> System: Request\nSystem -> User: Response\n@enduml\n```" + `

## Class: Domain Models

` + "```plantuml\n@startuml\nclass Entity {\n  +id: string\n  +method()\n}\n@enduml\n```" + `

## Activity: Process Flow

` + "```plantuml\n@startuml\nstart\n:Process Request;\n:Generate Response;\nstop\n@enduml\n```"
	}

	lower := strings.ToLower(s)
	var additions []string

	// Check for required diagram types
	if !strings.Contains(lower, "sequence") {
		additions = append(additions, "\n## Sequence: User Interactions\n\n```plantuml\n@startuml\nactor User\nUser -> System: Request\nSystem -> User: Response\n@enduml\n```")
	}

	if !strings.Contains(lower, "class") {
		additions = append(additions, "\n## Class: Domain Models\n\n```plantuml\n@startuml\nclass Entity {\n  +id: string\n  +method()\n}\n@enduml\n```")
	}

	if !strings.Contains(lower, "activity") {
		additions = append(additions, "\n## Activity: Process Flow\n\n```plantuml\n@startuml\nstart\n:Process Request;\n:Generate Response;\nstop\n@enduml\n```")
	}

	if len(additions) > 0 {
		return s + strings.Join(additions, "")
	}

	return s
}
func writeDesignScaffold(outputDir string) error {
	// Write architecture.md with scaffold content
	archContent := ensureArchitecture("")
	archPath := filepath.Join(outputDir, "architecture.md")
	if err := os.WriteFile(archPath, []byte(archContent), 0644); err != nil {
		return fmt.Errorf("write architecture.md: %w", err)
	}

	// Write uml.md with scaffold content
	umlContent := ensureUML("")
	umlPath := filepath.Join(outputDir, "uml.md")
	if err := os.WriteFile(umlPath, []byte(umlContent), 0644); err != nil {
		return fmt.Errorf("write uml.md: %w", err)
	}

	return nil
}
