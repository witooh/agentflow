package commands

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	_ "embed"

	"agentflow/internal/agents"
	"agentflow/internal/config"
	"agentflow/internal/prompt"
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

	files, err := prompt.ListInputFiles(cfg.IO.InputDir)
	if err != nil {
		return err
	}
	userPrompts, err := prompt.GetPromptFromFiles(files)
	if err != nil {
		return err
	}
	systemMessages, err := buildIntakeSystemMessage(cfg.IO.OutputDir)
	if err != nil {
		return err
	}
	prompts := agents.InputList(userPrompts, systemMessages)

	if len(files) == 0 {
		return ErrNoInputs
	} else if opts.DryRun {
		return nil
	} else {
		_, err := agents.PO.RunInputs(context.Background(), prompts)
		if err != nil {
			fmt.Printf("\n\n> Note: OpenAI call failed, wrote scaffold instead. Error: %v\n", err)
		}
	}

	return nil
}

func buildIntakeSystemMessage(outputDir string) ([]agents.TResponseInputItem, error) {
	data := struct {
		RequirementsPath string
	}{
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

func ensureRequirementsSections(s string) string {
	return strings.TrimSpace(s)
}

func withMetadataHeader(cfg *config.Config, files []string, body string) string {
	date := time.Now().Format("2006-01-02")
	header := strings.Builder{}
	header.WriteString("# AgentFlow — Requirements\n\n")
	header.WriteString(fmt.Sprintf("**Version:** %s  \n", cfg.SchemaVersion))
	header.WriteString(fmt.Sprintf("**Date:** %s  \n", date))
	owner := cfg.Metadata.Owner
	if strings.TrimSpace(owner) == "" {
		owner = ""
	}
	if owner != "" {
		header.WriteString(fmt.Sprintf("**Owner:** %s\n\n", owner))
	} else {
		header.WriteString("\n")
	}
	// Run metadata (comment block)
	header.WriteString("<!-- Run Metadata\n")
	header.WriteString(fmt.Sprintf("Project: %s\n", cfg.ProjectName))
	header.WriteString(fmt.Sprintf("Model: %s\n", cfg.LLM.Model))
	header.WriteString(fmt.Sprintf("Temperature: %.2f\n", cfg.LLM.Temperature))
	header.WriteString(fmt.Sprintf("MaxTokens: %d\n", cfg.LLM.MaxTokens))
	header.WriteString(fmt.Sprintf("Timestamp: %s\n", time.Now().Format(time.RFC3339)))
	if len(files) > 0 {
		header.WriteString("Inputs:\n")
		for _, f := range files {
			header.WriteString("- " + f + "\n")
		}
	}
	header.WriteString("-->\n\n")

	content := body
	// Avoid duplicate title if model already produced a full doc
	if strings.HasPrefix(strings.TrimSpace(strings.ToLower(content)), strings.ToLower("# AgentFlow — Requirements")) {
		return content
	}
	return header.String() + ensureRequirementsSections(content)
}
