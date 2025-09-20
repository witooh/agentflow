package commands

import (
	"context"
	"errors"
	"fmt"
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

	files, err := prompt.ListInputFiles(cfg.IO.InputDir)
	if err != nil {
		return err
	}
	userPrompts, err := prompt.GetPromptFromFiles(files)
	if err != nil {
		return err
	}
	systemMessage := createIntakeSystemMessage(cfg.IO.OutputDir)
	prompts := agents.InputList(userPrompts, systemMessage)

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

func createIntakeSystemMessage(path string) []agents.TResponseInputItem {
	systemMessage := agents.SystemMessage(`### 🎯 Output format (Markdown)

- **Business Goals & Success KPIs**  
  - Describe business drivers (compliance, UX, marketing agility, cost savings).  
  - Define measurable KPIs (e.g., opt-in rate target, consent sync SLA, regulator reporting turnaround).  

- **User Personas & Journeys**  
  - Customer (mobile/web) → manage consent, banner UX.  
  - Marketing/Analytics team → use dashboard, reporting.  
  - Regulator/Audit → compliance log, proof of consent.  
  - Backend/System Integrator → consume consent via API/SDK.  

- **Scope (MVP vs Future Phases)**  
  - Clearly separate **MVP features** vs **future expansion**.  
  - Use MoSCoW or phased roadmap (MVP → Phase 2 → Mature state).  

- **Functional Requirements (FR)**  
  - Detail user-facing and system-facing capabilities.  
  - Link each FR back to persona & business goal.  

- **Non-Functional Requirements (NFR)**  
  - Scale, latency, retention, compliance, UX accessibility, availability.  
  - Prioritize what is critical at MVP vs later.  

- **Dependencies & Risks**  
  - Dependencies on other teams (e.g., Data Lake, Security, Compliance).  
  - Risks (regulatory, adoption, tech feasibility).  

- **Constraints**  
  - Jurisdiction: Thailand only (PDPA).  
  - Data residency: PDPA compliant.  
  - Migration: cutover from OneTrust (big bang).  
  - Certifications: not required at MVP.  

- **Deliverables to Solution Architect (SA)**  
  - Consent use cases & flows (opt-in, revoke, merge, reporting).  
  - High-level data model (consent record, audit log, mapping to customer/device).  
  - Integration points (mobile, web, backend, data lake).  
  - Prioritized features (MVP vs future).  
  - Reporting requirements (dimensions, regulator templates).  

- **Timeline Summary (Product Roadmap)**  
  - Narrate evolution chronologically:  
    - MVP (core consent, banner, reporting baseline).  
    - Phase 2 (advanced analytics, audience targeting, cookie discovery).  
    - Future (scalability, certifications, multi-region compliance).  

- **Questions to Human (Stakeholders)**  
  - Business-side clarifications (regulator reporting expectation, marketing KPIs, branding rules).  
  - Technical-side clarifications (DB choice, API standards, realtime infra).  
`)

	writeFileMessage := agents.SystemMessage(fmt.Sprintf("เอา output มาสร้างไฟล์ %s/requirements.md", path))
	return agents.InputList(systemMessage, writeFileMessage)
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
