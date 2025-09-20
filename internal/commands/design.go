package commands

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
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
		return nil
	}

	_, err = agents.SA.RunInputs(context.Background(), systemMessages)
	if err != nil {
		fmt.Printf("\n\n> Note: OpenAI call failed, wrote scaffold instead. Error: %v\n", err)
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
	// Ensure Project Structure section exists
	return s
}

func ensureUML(s string) string {
	s = strings.TrimSpace(s)
	lower := strings.ToLower(s)
	needSeq := !strings.Contains(lower, "sequence")
	needClass := !strings.Contains(lower, "class")
	needAct := !strings.Contains(lower, "activity")
	if needSeq || needClass || needAct {
		var b strings.Builder
		b.WriteString(s)
		b.WriteString("\n\n")
		s = b.String()
	}
	return s
}
