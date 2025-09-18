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

	// Gather context files if present
	var ctxParts []string
	for _, name := range []string{"srs.md", "stories.md", "acceptance_criteria.md", "requirements.md"} {
		p := filepath.Join(sourceDir, name)
		if b, err := os.ReadFile(p); err == nil {
			ctxParts = append(ctxParts, fmt.Sprintf("# File: %s\n\n%s", name, string(b)))
		}
	}
	ctxContent := strings.TrimSpace(strings.Join(ctxParts, "\n\n"))

	role := strings.TrimSpace(opts.Role)
	if role == "" {
		role = "qa"
	}
	tpl := cfg.Roles[role]
	if strings.TrimSpace(tpl) == "" {
		tpl = "You are a QA Lead. Produce a concise but thorough test plan that aligns to the SRS and Acceptance Criteria."
	}

	prompt := strings.TrimSpace(strings.Join([]string{
		"SYSTEM:\n" + strings.TrimSpace(tpl),
		"CONTEXT:\n" + ctxContent,
		"EXTRA:\nProduce one markdown document delimited by the exact marker on its own line:\n--- TESTPLAN START ---\n...\nThe test plan must include at minimum these sections as markdown headings: 1) Test Strategy, 2) Scope, 3) Test Types (unit/integration/e2e), 4) Mapping to Acceptance Criteria, 5) Test Environments & Data, 6) Entry/Exit Criteria, 7) Risks & Mitigations, 8) Execution Plan & Responsibilities. Keep it practical and actionable.",
	}, "\n\n"))

	var content string
	var runID string
	if opts.DryRun {
		content = scaffoldQATestPlan()
	} else {
		resp, err := agents.LQ.Run(context.Background(), prompt)
		if err != nil {
			content = scaffoldQATestPlan() + fmt.Sprintf("\n\n> Note: OpenAI call failed, wrote scaffold instead. Error: %v\n", err)
		} else {
			// runID = resp.RunID
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
	if err := writeFileWithHeader(cfg, role, runID, sourcePath, outPath, body); err != nil {
		return err
	}
	return nil
}

func scaffoldQATestPlan() string {
	return strings.Join([]string{
		"--- TESTPLAN START ---",
		"## Test Strategy\nProvide overall test approach emphasizing risk-based testing and alignment with business goals.",
		"## Scope\nIn-scope and out-of-scope areas.",
		"## Test Types\n- Unit\n- Integration\n- End-to-End\n- Non-functional (Performance, Security, Usability)",
		"## Mapping to Acceptance Criteria\nFor each AC, outline test ideas or scenarios.",
		"## Test Environments & Data\nDescribe environments, dependencies, seed data, and data management.",
		"## Entry/Exit Criteria\nDefine readiness and completion criteria.",
		"## Risks & Mitigations\nIdentify key risks and how to mitigate them.",
		"## Execution Plan & Responsibilities\nWho does what, when, with tooling and reporting cadence.",
	}, "\n\n")
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
