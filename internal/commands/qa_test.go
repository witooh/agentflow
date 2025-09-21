package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test helper to create a minimal valid config file
func createTestConfig(t *testing.T, tempDir string) string {
	configContent := `{
  "schemaVersion": "0.1",
  "projectName": "test",
  "llm": {
    "model": "gpt-4",
    "temperature": 0.7,
    "maxTokens": 1000
  },
  "io": {
    "inputDir": "` + tempDir + `",
    "outputDir": "` + tempDir + `"
  }
}`
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}
	return configPath
}

func TestQA_ConfigLoadError(t *testing.T) {
	opts := QAOptions{
		ConfigPath: "/nonexistent/config.yaml",
		DryRun:     true,
	}

	err := QA(opts)
	if err == nil {
		t.Error("expected error when config file doesn't exist")
	}
	if !strings.Contains(err.Error(), "load config") {
		t.Errorf("expected 'load config' error, got %v", err)
	}
}

func TestQA_DryRun(t *testing.T) {
	tempDir := t.TempDir()
	configPath := createTestConfig(t, tempDir)

	opts := QAOptions{
		ConfigPath: configPath,
		SourceDir:  tempDir,
		OutputDir:  tempDir,
		DryRun:     true,
	}

	err := QA(opts)
	if err != nil {
		t.Errorf("dry run should not fail, got %v", err)
	}
}

func TestQA_DefaultSourceDir_DryRun(t *testing.T) {
	tempDir := t.TempDir()
	configPath := createTestConfig(t, tempDir)

	opts := QAOptions{
		ConfigPath: configPath,
		SourceDir:  "", // Empty, should default to cfg.IO.OutputDir
		OutputDir:  tempDir,
		DryRun:     true,
	}

	err := QA(opts)
	if err != nil {
		t.Errorf("should use default source dir, got %v", err)
	}
}

func TestQA_OutputDirOverride_DryRun(t *testing.T) {
	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "custom_output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatal(err)
	}

	configPath := createTestConfig(t, tempDir)

	opts := QAOptions{
		ConfigPath: configPath,
		SourceDir:  tempDir,
		OutputDir:  outputDir, // Override the output directory
		DryRun:     true,
	}

	err := QA(opts)
	if err != nil {
		t.Errorf("should override output dir, got %v", err)
	}
}

func TestQA_SourceDirTrimming_DryRun(t *testing.T) {
	tempDir := t.TempDir()
	configPath := createTestConfig(t, tempDir)

	opts := QAOptions{
		ConfigPath: configPath,
		SourceDir:  "  " + tempDir + "  ", // With whitespace
		OutputDir:  tempDir,
		DryRun:     true,
	}

	err := QA(opts)
	if err != nil {
		t.Errorf("should trim source dir whitespace, got %v", err)
	}
}

func TestBuildQASystemMessage_Success(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := tempDir
	outputDir := tempDir

	inputs, err := buildQASystemMessage(sourceDir, outputDir)
	if err != nil {
		t.Fatalf("buildQASystemMessage failed: %v", err)
	}

	if len(inputs) == 0 {
		t.Fatal("expected at least one input")
	}

	// Check that the system message contains the template content
	systemMsg := inputs[0].OfMessage.Content.OfString.String()

	// Check for expected file paths
	expectedPaths := []string{
		filepath.Join(sourceDir, "requirements.md"),
		filepath.Join(sourceDir, "srs.md"),
		filepath.Join(sourceDir, "stories.md"),
		filepath.Join(sourceDir, "acceptance_criteria.md"),
		filepath.Join(outputDir, "test-plan.md"),
	}

	for _, path := range expectedPaths {
		if !strings.Contains(systemMsg, path) {
			t.Errorf("system message should contain path %s", path)
		}
	}

	// Check for Thai content from the template
	if !strings.Contains(systemMsg, "QA Lead") {
		t.Error("system message should contain QA Lead role description")
	}
	if !strings.Contains(systemMsg, "Test Strategy") {
		t.Error("system message should contain Test Strategy section")
	}
	if !strings.Contains(systemMsg, "AgentFlow — Test Plan") {
		t.Error("system message should contain AgentFlow title format")
	}
}

func TestBuildQASystemMessage_FilePathConstruction(t *testing.T) {
	sourceDir := "/custom/source"
	outputDir := "/custom/output"

	inputs, err := buildQASystemMessage(sourceDir, outputDir)
	if err != nil {
		t.Fatalf("buildQASystemMessage failed: %v", err)
	}

	systemMsg := inputs[0].OfMessage.Content.OfString.String()

	// Test specific path construction
	tests := []struct {
		expected string
		desc     string
	}{
		{"/custom/source/requirements.md", "requirements path"},
		{"/custom/source/srs.md", "SRS path"},
		{"/custom/source/stories.md", "stories path"},
		{"/custom/source/acceptance_criteria.md", "acceptance criteria path"},
		{"/custom/output/test-plan.md", "test plan output path"},
	}

	for _, test := range tests {
		if !strings.Contains(systemMsg, test.expected) {
			t.Errorf("system message should contain %s: %s", test.desc, test.expected)
		}
	}
}

func TestBuildQASystemMessage_ReturnStructure(t *testing.T) {
	tempDir := t.TempDir()

	inputs, err := buildQASystemMessage(tempDir, tempDir)
	if err != nil {
		t.Fatalf("buildQASystemMessage failed: %v", err)
	}

	// Should return exactly one input item
	if len(inputs) != 1 {
		t.Errorf("expected 1 input item, got %d", len(inputs))
	}

	// Should be a system message
	input := inputs[0]
	if input.OfMessage == nil {
		t.Error("input should be a message")
	}

	content := input.OfMessage.Content.OfString.String()
	if content == "" {
		t.Error("system message content should not be empty")
	}
}

// Test table-driven scenarios for dry run mode
func TestQA_TableDriven_DryRun(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) QAOptions
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_config_dry_run",
			setupFunc: func(t *testing.T) QAOptions {
				tempDir := t.TempDir()
				configPath := createTestConfig(t, tempDir)
				return QAOptions{
					ConfigPath: configPath,
					SourceDir:  tempDir,
					OutputDir:  tempDir,
					DryRun:     true,
				}
			},
			expectError: false,
		},
		{
			name: "missing_config",
			setupFunc: func(t *testing.T) QAOptions {
				return QAOptions{
					ConfigPath: "/nonexistent/config.yaml",
					DryRun:     true,
				}
			},
			expectError: true,
			errorMsg:    "load config",
		},
		{
			name: "empty_source_dir_defaults",
			setupFunc: func(t *testing.T) QAOptions {
				tempDir := t.TempDir()
				configPath := createTestConfig(t, tempDir)
				return QAOptions{
					ConfigPath: configPath,
					SourceDir:  "", // Should default to config output dir
					OutputDir:  tempDir,
					DryRun:     true,
				}
			},
			expectError: false,
		},
		{
			name: "whitespace_trimming",
			setupFunc: func(t *testing.T) QAOptions {
				tempDir := t.TempDir()
				configPath := createTestConfig(t, tempDir)
				return QAOptions{
					ConfigPath: configPath,
					SourceDir:  "  " + tempDir + "  ", // With whitespace
					OutputDir:  tempDir,
					DryRun:     true,
				}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := tt.setupFunc(t)

			err := QA(opts)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectError && err != nil && tt.errorMsg != "" {
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got %v", tt.errorMsg, err)
				}
			}
		})
	}
}

func TestBuildQASystemMessage_EmptyDirs(t *testing.T) {
	inputs, err := buildQASystemMessage("", "")
	if err != nil {
		t.Fatalf("buildQASystemMessage with empty dirs should not fail: %v", err)
	}

	if len(inputs) != 1 {
		t.Errorf("expected 1 input item, got %d", len(inputs))
	}

	systemMsg := inputs[0].OfMessage.Content.OfString.String()

	// Should still contain template structure even with empty paths
	if !strings.Contains(systemMsg, "QA Lead") {
		t.Error("system message should contain QA Lead role even with empty paths")
	}
}

func TestBuildQASystemMessage_TemplateContent(t *testing.T) {
	tempDir := t.TempDir()

	inputs, err := buildQASystemMessage(tempDir, tempDir)
	if err != nil {
		t.Fatalf("buildQASystemMessage failed: %v", err)
	}

	systemMsg := inputs[0].OfMessage.Content.OfString.String()

	// Test for key template sections
	expectedSections := []string{
		"Test Strategy",
		"Scope",
		"Test Types",
		"Mapping to Acceptance Criteria",
		"Test Environments & Data",
		"Entry/Exit Criteria",
		"Risks & Mitigations",
		"Execution Plan & Responsibilities",
		"--- TESTPLAN START ---",
	}

	for _, section := range expectedSections {
		if !strings.Contains(systemMsg, section) {
			t.Errorf("system message should contain section: %s", section)
		}
	}
}
