package commands

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"agentflow/internal/agents"
	"agentflow/internal/config"
	"agentflow/internal/prompt"
)

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

	files, err := prompt.GetInputFiles(sourceDir, []string{"srs.md", "stories.md", "acceptance_criteria.md", "requirements.md"})
	if err != nil {
		return err
	}
	systemMessages, err := prompt.GetPromptFromFiles(files)
	if err != nil {
		return err
	}
	userMessage := agents.UserMessage(strings.Join([]string{
		"You are a QA Lead. Produce a concise but thorough test plan that aligns to the SRS and Acceptance Criteria.",
	}, "\n\n"))
	prompts := agents.InputList(systemMessages, userMessage)

	var content string
	if opts.DryRun {
		content = ""
	} else {
		resp, err := agents.LQ.RunInputs(context.Background(), prompts)
		if err != nil {
			content = fmt.Sprintf("\n\n> Note: OpenAI call failed, wrote scaffold instead. Error: %v\n", err)
		} else {
			content = resp
		}
	}

	plan := extractTestPlan(content)
	plan = ensureTestPlan(plan)

	date := time.Now().Format("2006-01-02")
	body := "# AgentFlow — Test Plan\n\n**Date:** " + date + "\n\n" + plan

	// Source path string for metadata
	sourcePath := filepath.Join(sourceDir, "{srs,stories,ac}")
	outPath := filepath.Join(cfg.IO.OutputDir, "test-plan.md")
	if err := writeFileWithHeader(cfg, sourcePath, outPath, body); err != nil {
		return err
	}
	return nil
}

func extractTestPlan(s string) string {
	const mark = "--- TESTPLAN START ---"
	idx := strings.Index(s, mark)
	if idx == -1 {
		return strings.TrimSpace(s)
	}
	return strings.TrimSpace(s[idx+len(mark):])
}

func ensureTestPlan(s string) string {
	lower := strings.ToLower(s)
	ensure := func(h string) bool { return strings.Contains(lower, strings.ToLower("## "+h)) }
	parts := []string{strings.TrimSpace(s)}
	// Append missing sections in order
	sections := map[string]string{
		"Test Strategy":                     "## Test Strategy\nTBD",
		"Scope":                             "## Scope\nTBD",
		"Test Types":                        "## Test Types\n- Unit\n- Integration\n- E2E",
		"Mapping to Acceptance Criteria":    "## Mapping to Acceptance Criteria\nTBD",
		"Test Environments & Data":          "## Test Environments & Data\nTBD",
		"Entry/Exit Criteria":               "## Entry/Exit Criteria\nTBD",
		"Risks & Mitigations":               "## Risks & Mitigations\nTBD",
		"Execution Plan & Responsibilities": "## Execution Plan & Responsibilities\nTBD",
	}
	for _, order := range []string{
		"Test Strategy", "Scope", "Test Types", "Mapping to Acceptance Criteria", "Test Environments & Data", "Entry/Exit Criteria", "Risks & Mitigations", "Execution Plan & Responsibilities",
	} {
		if !ensure(order) {
			parts = append(parts, sections[order])
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n"))
}
