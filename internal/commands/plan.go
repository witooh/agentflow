package commands

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	_ "embed"

	"agentflow/internal/agents"
	"agentflow/internal/config"
)

type PlanOptions struct {
	ConfigPath   string
	Requirements string
	OutputDir    string
	Role         string
	DryRun       bool
}

var ErrNoRequirements = errors.New("requirements.md not found")

//go:embed plan_prompt.md
var planPromptTemplate string

func Plan(opts PlanOptions) error {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	cfg.ApplyEnv()
	if opts.OutputDir != "" {
		cfg.IO.OutputDir = opts.OutputDir
	}
	if strings.TrimSpace(opts.Requirements) == "" {
		opts.Requirements = filepath.Join(cfg.IO.OutputDir, "requirements.md")
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	if err := config.EnsureDirs(opts.ConfigPath, cfg); err != nil {
		return err
	}
	if _, err := os.Stat(opts.Requirements); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrNoRequirements
		}
		return err
	}

	prompts, err := buildPlanSystemMessage(opts.Requirements, cfg.IO.OutputDir)
	if err != nil {
		return err
	}

	if opts.DryRun {
		return nil
	}

	_, err = agents.SA.RunInputs(context.Background(), prompts)
	if err != nil {
		fmt.Printf("OpenAI call failed, wrote scaffold instead. Error: %v\n", err)
	}

	return err
}

func buildPlanSystemMessage(requirementsPath, outputDir string) ([]agents.TResponseInputItem, error) {
	data := struct {
		RequirementsPath       string
		SrsPath                string
		StoriesPath            string
		AcceptanceCriteriaPath string
	}{
		RequirementsPath:       requirementsPath,
		SrsPath:                filepath.Join(outputDir, "srs.md"),
		StoriesPath:            filepath.Join(outputDir, "stories.md"),
		AcceptanceCriteriaPath: filepath.Join(outputDir, "acceptance_criteria.md"),
	}

	tmpl, err := template.New("plan").Parse(planPromptTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse plan template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("render plan template: %w", err)
	}

	return agents.InputList(
		agents.SystemMessage(buf.String()),
	), nil
}

func writeFileWithHeader(cfg *config.Config, sourcePath, outPath, body string) error {
	var meta strings.Builder
	meta.WriteString("<!-- Run Metadata\n")
	meta.WriteString(fmt.Sprintf("Project: %s\n", cfg.ProjectName))
	meta.WriteString(fmt.Sprintf("Model: %s\n", cfg.LLM.Model))
	meta.WriteString(fmt.Sprintf("Temperature: %.2f\n", cfg.LLM.Temperature))
	meta.WriteString(fmt.Sprintf("MaxTokens: %d\n", cfg.LLM.MaxTokens))
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
