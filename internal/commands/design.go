package commands

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"path/filepath"
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
