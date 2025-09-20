package commands

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"agentflow/internal/agents"
	"agentflow/internal/config"
)

type DevPlanOptions struct {
	ConfigPath string
	// SourceDir is where we read prior generated docs to build context (defaults to cfg.IO.OutputDir)
	SourceDir string
	// OutputDir is where we expect task_list.md and tasks/*.md to be generated (defaults to cfg.IO.OutputDir)
	OutputDir string
	Role      string // usually "dev"
	DryRun    bool
}

//go:embed devplan_prompt.md
var devPlanPromptTemplate string

func DevPlan(opts DevPlanOptions) error {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	cfg.ApplyEnv()
	if strings.TrimSpace(opts.OutputDir) != "" {
		cfg.IO.OutputDir = strings.TrimSpace(opts.OutputDir)
	}
	if strings.TrimSpace(opts.SourceDir) == "" {
		opts.SourceDir = cfg.IO.OutputDir
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	if err := config.EnsureDirs(opts.ConfigPath, cfg); err != nil {
		return err
	}

	prompts, err := buildDevPlanSystemMessage(opts.SourceDir, cfg)
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

func buildDevPlanSystemMessage(sourceDir string, cfg *config.Config) ([]agents.TResponseInputItem, error) {
	outputDir := cfg.IO.OutputDir
	data := struct {
		RequirementsPath       string
		SrsPath                string
		StoriesPath            string
		AcceptanceCriteriaPath string
		ArchitecturePath       string
		UmlPath                string
		TaskListPath           string
		TasksDir               string
		MaxContextChars        int
		ProjectName            string
		Model                  string
		Temperature            float64
		MaxTokens              int
		RunTimestamp           string
	}{
		RequirementsPath:       filepath.Join(sourceDir, "requirements.md"),
		SrsPath:                filepath.Join(sourceDir, "srs.md"),
		StoriesPath:            filepath.Join(sourceDir, "stories.md"),
		AcceptanceCriteriaPath: filepath.Join(sourceDir, "acceptance_criteria.md"),
		ArchitecturePath:       filepath.Join(sourceDir, "architecture.md"),
		UmlPath:                filepath.Join(sourceDir, "uml.md"),
		TaskListPath:           filepath.Join(outputDir, "task_list.md"),
		TasksDir:               filepath.Join(outputDir, "tasks"),
		MaxContextChars:        cfg.DevPlan.MaxContextCharsPerTask,
		ProjectName:            cfg.ProjectName,
		Model:                  cfg.LLM.Model,
		Temperature:            cfg.LLM.Temperature,
		MaxTokens:              cfg.LLM.MaxTokens,
		RunTimestamp:           time.Now().Format(time.RFC3339),
	}

	tmpl, err := template.New("devplan").Parse(devPlanPromptTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse devplan template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("render devplan template: %w", err)
	}

	return agents.InputList(
		agents.SystemMessage(buf.String()),
	), nil
}
