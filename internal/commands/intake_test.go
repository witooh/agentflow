package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"agentflow/internal/config"
)

func TestEnsureRequirementsSections(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "trims leading and trailing whitespace",
			input: "\n\ncontent\n\n",
			want:  "content",
		},
		{
			name:  "preserves interior formatting",
			input: "## Title\n\n- item\n",
			want:  "## Title\n\n- item",
		},
		{
			name:  "handles empty string",
			input: "",
			want:  "",
		},
		{
			name:  "handles only whitespace",
			input: "   \n\t  \n  ",
			want:  "",
		},
		{
			name:  "preserves single line content",
			input: "single line",
			want:  "single line",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ensureRequirementsSections(tt.input)
			if got != tt.want {
				t.Errorf("ensureRequirementsSections() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestScaffoldRequirements(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		prompt       string
		timeline     string
		wantContains []string
	}{
		{
			name:     "basic scaffold with timeline",
			prompt:   "test prompt",
			timeline: "- 2025-01-01 — First idea\n- 2025-01-02 — More details",
			wantContains: []string{
				"# requirements",
				"## Goals",
				"## Scope",
				"## Functional Requirements (FR)",
				"## Non-Functional Requirements (NFR)",
				"## Assumptions",
				"## Constraints",
				"## Timeline Summary",
				"## Open Questions",
				"## Questions to Human",
				"2025-01-01 — First idea",
				"2025-01-02 — More details",
				"```\ntest prompt\n```",
			},
		},
		{
			name:     "empty timeline",
			prompt:   "empty prompt",
			timeline: "",
			wantContains: []string{
				"# requirements",
				"## Timeline Summary\n\n",
				"```\nempty prompt\n```",
			},
		},
		{
			name:     "whitespace timeline",
			prompt:   "whitespace prompt",
			timeline: "   \n\t  ",
			wantContains: []string{
				"# requirements",
				"## Timeline Summary\n\n",
				"```\nwhitespace prompt\n```",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := scaffoldRequirements(tt.prompt, tt.timeline)

			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("scaffoldRequirements() missing expected content %q in output: %s", want, got)
				}
			}

			// Check that it contains a timestamp
			if !strings.Contains(got, "Generated at") {
				t.Error("scaffoldRequirements() should contain timestamp")
			}
		})
	}
}

func TestWriteRequirements(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "successful write",
			content: "# Test Requirements\n\nThis is test content.",
			wantErr: false,
		},
		{
			name:    "empty content",
			content: "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create a unique temporary directory for each test
			tmpDir := t.TempDir()

			err := writeRequirements(tmpDir, tt.content)

			if (err != nil) != tt.wantErr {
				t.Errorf("writeRequirements() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Check that file was created
				filePath := filepath.Join(tmpDir, "requirements.md")
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					t.Errorf("writeRequirements() did not create file at %s", filePath)
					return
				}

				// Check file content
				got, err := os.ReadFile(filePath)
				if err != nil {
					t.Errorf("writeRequirements() created file but cannot read it: %v", err)
					return
				}

				if string(got) != tt.content {
					t.Errorf("writeRequirements() wrote %q, want %q", string(got), tt.content)
				}
			}
		})
	}
}

func TestWriteRequirements_NonExistentDirectory(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	nonExistentDir := filepath.Join(tmpDir, "nonexistent", "subdir")
	content := "# Test Requirements\n\nThis is test content."

	err := writeRequirements(nonExistentDir, content)

	if err != nil {
		t.Errorf("writeRequirements() error = %v, wantErr false", err)
		return
	}

	// Check that file was created
	filePath := filepath.Join(nonExistentDir, "requirements.md")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("writeRequirements() did not create file at %s", filePath)
		return
	}

	// Check file content
	got, err := os.ReadFile(filePath)
	if err != nil {
		t.Errorf("writeRequirements() created file but cannot read it: %v", err)
		return
	}

	if string(got) != content {
		t.Errorf("writeRequirements() wrote %q, want %q", string(got), content)
	}
}

func TestWithMetadataHeader(t *testing.T) {
	t.Parallel()

	// Create a test config
	cfg := &config.Config{
		SchemaVersion: "1.0.0",
		ProjectName:   "test-project",
		Metadata: struct {
			Owner string   `json:"owner"`
			Repo  string   `json:"repo"`
			Tags  []string `json:"tags"`
		}{
			Owner: "test-owner",
		},
		LLM: struct {
			Model       string  `json:"model"`
			Temperature float64 `json:"temperature"`
			MaxTokens   int     `json:"maxTokens"`
		}{
			Model:       "gpt-4",
			Temperature: 0.7,
			MaxTokens:   1000,
		},
	}

	tests := []struct {
		name         string
		cfg          *config.Config
		role         string
		files        []string
		runID        string
		body         string
		wantContains []string
	}{
		{
			name:  "complete metadata",
			cfg:   cfg,
			role:  "po_pm",
			files: []string{"input1.md", "input2.md"},
			runID: "run-123",
			body:  "# Test Body\n\nContent here.",
			wantContains: []string{
				"# AgentFlow — Requirements",
				"**Version:** 1.0.0",
				"**Date:** " + time.Now().Format("2006-01-02"),
				"**Owner:** test-owner",
				"Project: test-project",
				"Role: po_pm",
				"Model: gpt-4",
				"Temperature: 0.70",
				"MaxTokens: 1000",
				"RunID: run-123",
				"Inputs:",
				"- input1.md",
				"- input2.md",
				"# Test Body",
				"Content here.",
			},
		},
		{
			name: "no owner",
			cfg: &config.Config{
				SchemaVersion: "1.0.0",
				ProjectName:   "test-project",
				Metadata: struct {
					Owner string   `json:"owner"`
					Repo  string   `json:"repo"`
					Tags  []string `json:"tags"`
				}{},
				LLM: struct {
					Model       string  `json:"model"`
					Temperature float64 `json:"temperature"`
					MaxTokens   int     `json:"maxTokens"`
				}{
					Model:       "gpt-4",
					Temperature: 0.7,
					MaxTokens:   1000,
				},
			},
			role:  "po_pm",
			files: []string{},
			runID: "",
			body:  "Simple body",
			wantContains: []string{
				"# AgentFlow — Requirements",
				"**Version:** 1.0.0",
				"**Date:** " + time.Now().Format("2006-01-02"),
				"Project: test-project",
				"Role: po_pm",
				"Model: gpt-4",
				"Temperature: 0.70",
				"MaxTokens: 1000",
				"Simple body",
			},
		},
		{
			name:  "body already has title",
			cfg:   cfg,
			role:  "po_pm",
			files: []string{},
			runID: "",
			body:  "# AgentFlow — Requirements\n\nAlready has title.",
			wantContains: []string{
				"# AgentFlow — Requirements",
				"Already has title.",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := withMetadataHeader(tt.cfg, tt.role, tt.files, tt.runID, tt.body)

			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("withMetadataHeader() missing expected content %q in output", want)
				}
			}

			// Check that it doesn't duplicate the title when body already has it
			if strings.Contains(tt.body, "# AgentFlow — Requirements") {
				titleCount := strings.Count(got, "# AgentFlow — Requirements")
				if titleCount > 1 {
					t.Errorf("withMetadataHeader() duplicated title, found %d occurrences", titleCount)
				}
			}
		})
	}
}

func TestContains(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    string
		sub  string
		want bool
	}{
		{
			name: "contains substring",
			s:    "hello world",
			sub:  "world",
			want: true,
		},
		{
			name: "does not contain substring",
			s:    "hello world",
			sub:  "universe",
			want: false,
		},
		{
			name: "empty string contains empty substring",
			s:    "",
			sub:  "",
			want: true,
		},
		{
			name: "empty string does not contain non-empty substring",
			s:    "",
			sub:  "test",
			want: false,
		},
		{
			name: "exact match",
			s:    "exact",
			sub:  "exact",
			want: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := strings.Contains(tt.s, tt.sub)
			if got != tt.want {
				t.Errorf("strings.Contains(%q, %q) = %v, want %v", tt.s, tt.sub, got, tt.want)
			}
		})
	}
}
