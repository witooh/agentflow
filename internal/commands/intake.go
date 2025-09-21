package commands

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"text/template"

	_ "embed"

	"agentflow/internal/agents"
	"agentflow/internal/config"
)

type IntakeOptions struct {
	ConfigPath string
	InputsDir  string
	OutputDir  string
	Role       string
	DryRun     bool
}

var ErrNoInputs = errors.New("no input files found")

//go:embed intake_prompt.md
var intakePromptTemplate string

func Intake(opts IntakeOptions) error {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	cfg.ApplyEnv()
	// Ensure IO dirs from opts override config if provided
	if opts.InputsDir != "" {
		cfg.IO.InputDir = opts.InputsDir
	}
	if opts.OutputDir != "" {
		cfg.IO.OutputDir = opts.OutputDir
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	if err := config.EnsureDirs(opts.ConfigPath, cfg); err != nil {
		return err
	}

	systemMessages, err := buildIntakeSystemMessage(cfg.IO.InputDir, cfg.IO.OutputDir)
	if err != nil {
		return err
	}

	if opts.DryRun {
		return nil
	} else {
		_, err := agents.PO.RunInputs(context.Background(), systemMessages)
		if err != nil {
			fmt.Printf("\n\n> Note: OpenAI call failed, wrote scaffold instead. Error: %v\n", err)
		}
	}

	return nil
}

func buildIntakeSystemMessage(inputDir, outputDir string) ([]agents.TResponseInputItem, error) {
	data := struct {
		InputPath        string
		RequirementsPath string
	}{
		InputPath:        filepath.Clean(inputDir),
		RequirementsPath: filepath.Join(outputDir, "requirements.md"),
	}
	tmpl, err := template.New("intake").Parse(intakePromptTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse intake template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("render intake template: %w", err)
	}
	return agents.InputList(
		agents.SystemMessage(buf.String()),
	), nil
}
