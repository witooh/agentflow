package commands

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agentflow/internal/config"
)

func TestPlan_NoRequirements(t *testing.T) {
	tempDir := t.TempDir()
	opts := PlanOptions{
		ConfigPath:   filepath.Join(tempDir, "config.yaml"),
		Requirements: filepath.Join(tempDir, "nonexistent.md"),
		OutputDir:    tempDir,
		DryRun:       true,
	}

	// Create minimal config file
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
	if err := os.WriteFile(opts.ConfigPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	err := Plan(opts)
	if !errors.Is(err, ErrNoRequirements) {
		t.Errorf("expected ErrNoRequirements, got %v", err)
	}
}

func TestPlan_DryRun(t *testing.T) {
	tempDir := t.TempDir()
	reqFile := filepath.Join(tempDir, "requirements.md")
	configFile := filepath.Join(tempDir, "config.yaml")

	// Create requirements file
	if err := os.WriteFile(reqFile, []byte("# Test Requirements\nSome requirements"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create config file
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
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	opts := PlanOptions{
		ConfigPath:   configFile,
		Requirements: reqFile,
		OutputDir:    tempDir,
		DryRun:       true,
	}

	err := Plan(opts)
	if err != nil {
		t.Errorf("dry run should not fail, got %v", err)
	}
}

func TestPlan_DefaultRequirementsPath(t *testing.T) {
	tempDir := t.TempDir()
	reqFile := filepath.Join(tempDir, "requirements.md")
	configFile := filepath.Join(tempDir, "config.yaml")

	// Create requirements file
	if err := os.WriteFile(reqFile, []byte("# Test Requirements\nSome requirements"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create config file
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
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	opts := PlanOptions{
		ConfigPath:   configFile,
		Requirements: "", // Empty, should default to outputDir/requirements.md
		OutputDir:    tempDir,
		DryRun:       true,
	}

	err := Plan(opts)
	if err != nil {
		t.Errorf("should use default requirements path, got %v", err)
	}
}

func TestBuildPlanSystemMessage(t *testing.T) {
	tempDir := t.TempDir()
	reqFile := filepath.Join(tempDir, "requirements.md")

	// Create requirements file
	if err := os.WriteFile(reqFile, []byte("# Test Requirements\nSome requirements"), 0644); err != nil {
		t.Fatal(err)
	}

	inputs, err := buildPlanSystemMessage(reqFile, tempDir)
	if err != nil {
		t.Fatalf("buildPlanSystemMessage failed: %v", err)
	}

	if len(inputs) == 0 {
		t.Fatal("expected at least one input")
	}

	// Check that the system message contains the template content
	systemMsg := inputs[0].OfMessage.Content.OfString.String()
	if !strings.Contains(systemMsg, reqFile) {
		t.Errorf("system message should contain requirements path %s", reqFile)
	}

	expectedPaths := []string{
		filepath.Join(tempDir, "srs.md"),
		filepath.Join(tempDir, "stories.md"),
		filepath.Join(tempDir, "acceptance_criteria.md"),
	}

	for _, path := range expectedPaths {
		if !strings.Contains(systemMsg, path) {
			t.Errorf("system message should contain path %s", path)
		}
	}
}

func TestSplitPlanContent_NoMarkers(t *testing.T) {
	content := "Just regular content without markers"
	srs, stories, ac := splitPlanContent(content)

	if srs != content {
		t.Errorf("expected srs to be full content, got %q", srs)
	}
	if stories != "" {
		t.Errorf("expected empty stories, got %q", stories)
	}
	if ac != "" {
		t.Errorf("expected empty ac, got %q", ac)
	}
}

func TestSplitPlanContent_WithMarkers(t *testing.T) {
	content := `Introduction
--- SRS START ---
SRS content here
--- STORIES START ---
Stories content here
--- AC START ---
AC content here`

	srs, stories, ac := splitPlanContent(content)

	if !strings.Contains(srs, "SRS content") {
		t.Errorf("expected SRS content, got %q", srs)
	}
	if !strings.Contains(stories, "Stories content") {
		t.Errorf("expected Stories content, got %q", stories)
	}
	if !strings.Contains(ac, "AC content") {
		t.Errorf("expected AC content, got %q", ac)
	}
}

func TestSplitPlanContent_PartialMarkers(t *testing.T) {
	content := `Introduction
--- SRS START ---
SRS content here
--- STORIES START ---
Stories content here`

	srs, stories, ac := splitPlanContent(content)

	if !strings.Contains(srs, "SRS content") {
		t.Errorf("expected SRS content, got %q", srs)
	}
	if !strings.Contains(stories, "Stories content") {
		t.Errorf("expected Stories content, got %q", stories)
	}
	if ac != "" {
		t.Errorf("expected empty AC, got %q", ac)
	}
}

func TestEnsureSRS_Empty(t *testing.T) {
	result := ensureSRS("")

	if !strings.Contains(result, "บทนำ") {
		t.Error("ensureSRS should add Thai intro section")
	}
	if !strings.Contains(result, "Use Cases") {
		t.Error("ensureSRS should add Use Cases section")
	}
	if !strings.Contains(result, "Interfaces") {
		t.Error("ensureSRS should add Interfaces section")
	}
	if !strings.Contains(result, "Constraints") {
		t.Error("ensureSRS should add Constraints section")
	}
}

func TestEnsureSRS_ValidContent(t *testing.T) {
	content := "# SRS\n## Use Cases\nUC-01\n## Interfaces\nAPI\n## Constraints\nPerformance"
	result := ensureSRS(content)

	if result != content {
		t.Errorf("ensureSRS should return original content when valid, got %q", result)
	}
}

func TestEnsureSRS_MissingSections(t *testing.T) {
	content := "Just an overview without proper sections"
	result := ensureSRS(content)

	// Should return template because it's missing most required sections
	if !strings.Contains(result, "Use Cases") {
		t.Error("ensureSRS should add template when sections are missing")
	}
}

func TestEnsureStories_Empty(t *testing.T) {
	result := ensureStories("")

	if !strings.Contains(result, "EPIC-1") {
		t.Error("ensureStories should add EPIC section")
	}
	if !strings.Contains(result, "STORY-1.1") {
		t.Error("ensureStories should add STORY section")
	}
	if !strings.Contains(result, "AC:") {
		t.Error("ensureStories should add AC reference")
	}
}

func TestEnsureStories_WithContent(t *testing.T) {
	content := "## EPIC-2\n- STORY-2.1: Custom story"
	result := ensureStories(content)

	if result != content {
		t.Errorf("ensureStories should return original content, got %q", result)
	}
}

func TestEnsureAC_Empty(t *testing.T) {
	result := ensureAC("")

	if !strings.Contains(result, "STORY-1.1") {
		t.Error("ensureAC should add STORY section")
	}
	if !strings.Contains(result, "[ ]") {
		t.Error("ensureAC should add checkbox")
	}
}

func TestEnsureAC_WithContent(t *testing.T) {
	content := "## STORY-2.1\n- [ ] Custom criteria"
	result := ensureAC(content)

	if result != content {
		t.Errorf("ensureAC should return original content, got %q", result)
	}
}

func TestWriteFileWithHeader(t *testing.T) {
	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "test.md")

	cfg := &config.Config{
		ProjectName: "TestProject",
	}
	cfg.LLM.Model = "gpt-4"
	cfg.LLM.Temperature = 0.7
	cfg.LLM.MaxTokens = 1000

	body := "# Test Content\nSome content here"
	sourcePath := "/source/requirements.md"

	err := writeFileWithHeader(cfg, sourcePath, outPath, body)
	if err != nil {
		t.Fatalf("writeFileWithHeader failed: %v", err)
	}

	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	contentStr := string(content)

	// Check that metadata is present
	if !strings.Contains(contentStr, "Project: TestProject") {
		t.Error("output should contain project name")
	}
	if !strings.Contains(contentStr, "Model: gpt-4") {
		t.Error("output should contain model name")
	}
	if !strings.Contains(contentStr, "Temperature: 0.70") {
		t.Error("output should contain temperature")
	}
	if !strings.Contains(contentStr, "SourceRequirements: /source/requirements.md") {
		t.Error("output should contain source path")
	}
	if !strings.Contains(contentStr, "Test Content") {
		t.Error("output should contain original body")
	}

	// Check timestamp format
	if !strings.Contains(contentStr, "Timestamp:") {
		t.Error("output should contain timestamp")
	}
}

func TestWriteFileWithHeader_AgentflowTitle(t *testing.T) {
	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "test.md")

	cfg := &config.Config{
		ProjectName: "TestProject",
	}
	cfg.LLM.Model = "gpt-4"
	cfg.LLM.Temperature = 0.7
	cfg.LLM.MaxTokens = 1000

	body := "# AgentFlow Project\nContent with AgentFlow title"
	sourcePath := "/source/requirements.md"

	err := writeFileWithHeader(cfg, sourcePath, outPath, body)
	if err != nil {
		t.Fatalf("writeFileWithHeader failed: %v", err)
	}

	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	contentStr := string(content)

	// Should contain the AgentFlow title
	if !strings.Contains(contentStr, "# AgentFlow Project") {
		t.Error("output should preserve AgentFlow title")
	}

	// Should still contain metadata
	if !strings.Contains(contentStr, "Project: TestProject") {
		t.Error("output should contain metadata even with AgentFlow title")
	}
}
