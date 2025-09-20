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

//go:embed qa_prompt.md
var qaPromptTemplate string

type QAOptions struct {
	ConfigPath string
	SourceDir  string // where to read prior docs (srs/stories/acceptance_criteria). If empty, use cfg.IO.OutputDir
	OutputDir  string // where to write test-plan.md
	Role       string
	DryRun     bool
}

// QA generates a test-plan.md using SRS/Stories/Acceptance Criteria as context.
func QA(opts QAOptions) error {
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

	sourceDir := strings.TrimSpace(opts.SourceDir)
	if sourceDir == "" {
		sourceDir = cfg.IO.OutputDir
	}

	prompts, err := buildQASystemMessage(sourceDir, cfg.IO.OutputDir)
	if err != nil {
		return err
	}

	if opts.DryRun {
		return nil
	} else {
		_, err := agents.LQ.RunInputs(context.Background(), prompts)
		if err != nil {
			fmt.Printf("OpenAI call failed, wrote scaffold instead. Error: %v\n", err)
		}
	}

	return nil
}

func buildQASystemMessage(sourceDir, outputDir string) ([]agents.TResponseInputItem, error) {
	data := struct {
		RequirementsPath       string
		SrsPath                string
		StoriesPath            string
		AcceptanceCriteriaPath string
		TestPlanPath           string
	}{
		RequirementsPath:       filepath.Join(sourceDir, "requirements.md"),
		SrsPath:                filepath.Join(sourceDir, "srs.md"),
		StoriesPath:            filepath.Join(sourceDir, "stories.md"),
		AcceptanceCriteriaPath: filepath.Join(sourceDir, "acceptance_criteria.md"),
		TestPlanPath:           filepath.Join(outputDir, "test-plan.md"),
	}

	tmpl, err := template.New("qa").Parse(qaPromptTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse qa template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("render qa template: %w", err)
	}

	return agents.InputList(
		agents.SystemMessage(buf.String()),
	), nil
}
