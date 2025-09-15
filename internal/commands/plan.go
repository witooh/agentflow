package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"agentflow/internal/config"
	"agentflow/internal/langgraph"
)

type PlanOptions struct {
	ConfigPath   string
	Requirements string
	OutputDir    string
	Role         string
	DryRun       bool
}

var ErrNoRequirements = errors.New("requirements.md not found")

func Plan(opts PlanOptions) error {
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

	reqPath := opts.Requirements
	if strings.TrimSpace(reqPath) == "" {
		reqPath = filepath.Join(cfg.IO.OutputDir, "requirements.md")
	}
	b, err := os.ReadFile(reqPath)
	if err != nil {
		return ErrNoRequirements
	}
	role := opts.Role
	if role == "" {
		role = "sa"
	}
	tpl := cfg.Roles[role]
	if strings.TrimSpace(tpl) == "" {
		tpl = "You are a Solution Architect. Transform requirements into SRS, Stories (INVEST), and Acceptance Criteria. Ensure Use Cases, Interfaces, Constraints in SRS; INVEST in stories; and each story maps to AC. Output markdown clearly separated with headings: --- SRS START ---, --- STORIES START ---, --- AC START ---"
	}

	// Build prompt inline from requirements content to avoid scanning whole input dir.
	prompt := strings.TrimSpace(strings.Join([]string{
		"SYSTEM:\n" + strings.TrimSpace(tpl),
		"CONTEXT:\n# File: requirements.md\n\n" + string(b),
		"EXTRA:\nPlease produce three markdown documents. Delimit each section with exact markers on their own lines: \n--- SRS START ---\n...\n--- STORIES START ---\n...\n--- AC START ---\n...\nMake content concise and complete per acceptance criteria.",
	}, "\n\n"))

	var content string
	var runID string
	if opts.DryRun {
		content = scaffoldPlanOutput(prompt)
	} else {
		client := langgraph.NewClient()
		resp, err := client.RunAgent(langgraph.RunRequest{
			Role:   role,
			Prompt: prompt,
			Params: map[string]interface{}{
				"model":       cfg.LLM.Model,
				"temperature": cfg.LLM.Temperature,
				"max_tokens":  cfg.LLM.MaxTokens,
			},
		})
		if err != nil {
			content = scaffoldPlanOutput(prompt) + fmt.Sprintf("\n\n> Note: OpenAI call failed, wrote scaffold instead. Error: %v\n", err)
		} else {
			runID = resp.RunID
			content = resp.Content
		}
	}

	srs, stories, ac := splitPlanContent(content)
	// Ensure minimal structures
	srs = ensureSRS(srs)
	stories = ensureStories(stories)
	ac = ensureAC(ac)

	date := time.Now().Format("2006-01-02")
	// Add headers and write files
	if err := writeFileWithHeader(cfg, role, runID, reqPath, filepath.Join(cfg.IO.OutputDir, "srs.md"),
		"# AgentFlow — Software Requirements Specification (SRS)\n\n**Version:** "+cfg.SchemaVersion+"  \n**Date:** "+date+"\n\n"+srs); err != nil {
		return err
	}
	if err := writeFileWithHeader(cfg, role, runID, reqPath, filepath.Join(cfg.IO.OutputDir, "stories.md"),
		"# AgentFlow — User Stories (INVEST)\n\n**Date:** "+date+"\n\n"+stories); err != nil {
		return err
	}
	if err := writeFileWithHeader(cfg, role, runID, reqPath, filepath.Join(cfg.IO.OutputDir, "acceptance_criteria.md"),
		"# AgentFlow — Acceptance Criteria\n\n**Date:** "+date+"\n\n"+ac); err != nil {
		return err
	}
	return nil
}

func writeFileWithHeader(cfg *config.Config, role, runID, sourcePath, outPath, body string) error {
	var meta strings.Builder
	meta.WriteString("<!-- Run Metadata\n")
	meta.WriteString(fmt.Sprintf("Project: %s\n", cfg.ProjectName))
	meta.WriteString(fmt.Sprintf("Role: %s\n", role))
	meta.WriteString(fmt.Sprintf("Model: %s\n", cfg.LLM.Model))
	meta.WriteString(fmt.Sprintf("Temperature: %.2f\n", cfg.LLM.Temperature))
	meta.WriteString(fmt.Sprintf("MaxTokens: %d\n", cfg.LLM.MaxTokens))
	if runID != "" {
		meta.WriteString(fmt.Sprintf("RunID: %s\n", runID))
	}
	meta.WriteString(fmt.Sprintf("SourceRequirements: %s\n", sourcePath))
	meta.WriteString(fmt.Sprintf("Timestamp: %s\n", time.Now().Format(time.RFC3339)))
	meta.WriteString("-->\n\n")

	content := body
	// Avoid duplicating header if model already provided a full document with same title
	lower := strings.ToLower(strings.TrimSpace(content))
	if strings.HasPrefix(lower, strings.ToLower("# agentflow")) {
		content = content + "\n\n" + meta.String()
	} else {
		content = body + "\n\n" + meta.String()
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(outPath, []byte(content), 0o644)
}

func scaffoldPlanOutput(prompt string) string {
	return strings.Join([]string{
		"--- SRS START ---",
		"## 1. บทนำ\n- ...",
		"## 3. Use Cases\n- UC-01 ...",
		"## Interfaces\n- ...\n## Constraints\n- ...",
		"--- STORIES START ---",
		"## EPIC-1\n- STORY-1.1: ...\n  - AC: ...",
		"--- AC START ---",
		"## STORY-1.1\n- [ ] ...",
		"\n<!-- Prompt Context (debug) -->\n```\n" + prompt + "\n```",
	}, "\n\n")
}

func splitPlanContent(s string) (string, string, string) {
	// naive splitting by markers; if missing, treat whole as SRS
	low := strings.ToLower(s)
	iSRS := strings.Index(low, "--- srs start ---")
	iStories := strings.Index(low, "--- stories start ---")
	iAC := strings.Index(low, "--- ac start ---")
	if iSRS == -1 && iStories == -1 && iAC == -1 {
		return s, "", ""
	}
	// compute slices
	var srs, stories, ac string
	end := len(s)
	if iSRS >= 0 {
		start := iSRS + len("--- srs start ---")
		if iStories >= 0 {
			endSRS := iStories
			if iAC >= 0 && iAC < iStories {
				endSRS = iAC
			}
			srs = strings.TrimSpace(s[start:endSRS])
		} else if iAC >= 0 {
			endSRS := iAC
			srs = strings.TrimSpace(s[start:endSRS])
		} else {
			srs = strings.TrimSpace(s[start:end])
		}
	}
	if iStories >= 0 {
		start := iStories + len("--- stories start ---")
		if iAC >= 0 {
			stories = strings.TrimSpace(s[start:iAC])
		} else {
			stories = strings.TrimSpace(s[start:end])
		}
	}
	if iAC >= 0 {
		start := iAC + len("--- ac start ---")
		ac = strings.TrimSpace(s[start:end])
	}
	return srs, stories, ac
}

func ensureSRS(s string) string {
	if strings.TrimSpace(s) == "" {
		return "## 1. บทนำ\n- ...\n\n## 3. Use Cases\n- UC-01 ...\n\n## Interfaces\n- ...\n\n## Constraints\n- ..."
	}
	low := strings.ToLower(s)
	need := []string{"use cases", "interfaces", "constraints"}
	missing := 0
	for _, n := range need {
		if !strings.Contains(low, n) {
			missing++
		}
	}
	if missing >= 2 { // likely not SRS shaped
		return "## 1. บทนำ\n- ...\n\n## 3. Use Cases\n- UC-01 ...\n\n## Interfaces\n- ...\n\n## Constraints\n- ..."
	}
	return s
}

func ensureStories(s string) string {
	if strings.TrimSpace(s) == "" {
		return "## EPIC-1\n- STORY-1.1: ...\n  - AC: ..."
	}
	return s
}

func ensureAC(s string) string {
	if strings.TrimSpace(s) == "" {
		return "## STORY-1.1\n- [ ] ..."
	}
	return s
}
