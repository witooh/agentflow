package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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

	role := opts.Role
	if role == "" {
		role = "po_pm"
	}
	tpl := cfg.Roles[role]
	if strings.TrimSpace(tpl) == "" {
		// fallback
		tpl = "You are a PO/PM. Convert input context into formal requirements."
	}

	// Build a timeline summary from dated input files for context and guaranteed output sections
	timeline, _, terr := prompt.BuildTimelineSummary(cfg.IO.InputDir)
	if terr != nil {
		// non-fatal; continue without timeline
		timeline = ""
	}

	p, files, err := prompt.BuildForRole(prompt.BuildOptions{
		RoleTemplate: tpl,
		InputsDir:    cfg.IO.InputDir,
		ExtraContext: "Please produce markdown with sections: Goals, Scope, FR, NFR, Assumptions, Constraints, Timeline Summary, Questions to Human. Use the TIMELINE SUMMARY to narrate idea evolution chronologically.\n\nTIMELINE SUMMARY (derived from filenames):\n" + timeline,
	})
	if err != nil {
		return err
	}

	var baseContent string
	var runID string
	if len(files) == 0 {
		// still proceed but signal no inputs
		baseContent = "_No inputs found. This is a scaffold file._\n" + scaffoldRequirements(p, timeline)
	} else if opts.DryRun {
		baseContent = scaffoldRequirements(p, timeline)
	} else {
		resp, err := agents.PO.Run(context.Background(), p) // log prompt via agent
		if err != nil {
			// Fallback to scaffold on error
			baseContent = scaffoldRequirements(p, timeline) + fmt.Sprintf("\n\n> Note: OpenAI call failed, wrote scaffold instead. Error: %v\n", err)
		} else {
			// runID = resp.RunID
			// baseContent = ensureRequirementsSections(resp, timeline)
			baseContent = resp
		}
	}

	final := withMetadataHeader(cfg, role, files, runID, baseContent)
	if err := writeRequirements(cfg.IO.OutputDir, final); err != nil {
		return err
	}
	if len(files) == 0 {
		return ErrNoInputs
	}
	return nil
}

func writeRequirements(outputDir string, content string) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(outputDir, "requirements.md")
	return os.WriteFile(path, []byte(content), 0o644)
}

func scaffoldRequirements(p string, timeline string) string {
	parts := []string{
		"# requirements",
		fmt.Sprintf("> Generated at %s (scaffold)\n", time.Now().Format(time.RFC3339)),
		"## Goals\n\n- ...",
		"## Scope\n\n- ...",
		"## Functional Requirements (FR)\n\n- ...",
		"## Non-Functional Requirements (NFR)\n\n- ...",
		"## Assumptions\n\n- ...",
		"## Constraints\n\n- ...",
		"## Timeline Summary\n\n" + strings.TrimSpace(timeline),
		"\n## Open Questions\n\n- ...",
		"\n## Questions to Human\n\n- ...",
		"\n<!-- Prompt Context (debug) -->\n",
		"```\n" + p + "\n```\n",
	}
	return strings.Join(parts, "\n")
}

func ensureRequirementsSections(s string) string {
	return strings.TrimSpace(s)
}

func withMetadataHeader(cfg *config.Config, role string, files []string, runID string, body string) string {
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
	header.WriteString(fmt.Sprintf("Role: %s\n", role))
	header.WriteString(fmt.Sprintf("Model: %s\n", cfg.LLM.Model))
	header.WriteString(fmt.Sprintf("Temperature: %.2f\n", cfg.LLM.Temperature))
	header.WriteString(fmt.Sprintf("MaxTokens: %d\n", cfg.LLM.MaxTokens))
	if runID != "" {
		header.WriteString(fmt.Sprintf("RunID: %s\n", runID))
	}
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
