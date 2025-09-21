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

//go:embed uml_prompt.md
var umlPromptTemplate string

type UmlOptions struct {
	ConfigPath string
	SourceDir  string // where to read prior docs (requirements/srs/stories). If empty, use cfg.IO.OutputDir
	OutputDir  string // where to write uml.md
	Role       string
	DryRun     bool
}

func Uml(opts UmlOptions) error {
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

	prompts, err := buildUmlSystemMessage(opts.SourceDir, cfg.IO.OutputDir)
	if err != nil {
		return err
	}

	if opts.DryRun {
		return nil
	}

	_, err = agents.SA.RunInputs(context.Background(), prompts)
	if err != nil {
		fmt.Printf("\n\n> Note: OpenAI call failed, wrote scaffold instead. Error: %v\n", err)
	}
	return err
}

func buildUmlSystemMessage(sourceDir, outputDir string) ([]agents.TResponseInputItem, error) {
	data := struct {
		RequirementsPath string
		SrsPath          string
		StoriesPath      string
		UmlPath          string
	}{
		RequirementsPath: filepath.Join(sourceDir, "requirements.md"),
		SrsPath:          filepath.Join(sourceDir, "srs.md"),
		StoriesPath:      filepath.Join(sourceDir, "stories.md"),
		UmlPath:          filepath.Join(outputDir, "uml.md"),
	}

	tmpl, err := template.New("uml").Parse(umlPromptTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse uml template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("render uml template: %w", err)
	}

	return agents.InputList(
		agents.SystemMessage(buf.String()),
	), nil
}
