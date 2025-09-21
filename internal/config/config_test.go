package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfigAndSaveLoad(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".agentflow", "config.json")

	c := DefaultConfig("Demo", "gpt-4o-mini")
	// Ensure IO dirs are scoped under temp dir to avoid polluting repo
	c.IO.InputDir = filepath.Join(dir, "input")
	c.IO.OutputDir = filepath.Join(dir, "output")

	if err := Save(cfgPath, c); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	// Ensure directories created
	if _, err := os.Stat(c.IO.InputDir); err != nil {
		t.Fatalf("expected input dir created: %v", err)
	}
	if _, err := os.Stat(c.IO.OutputDir); err != nil {
		t.Fatalf("expected output dir created: %v", err)
	}

	loaded, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if err := loaded.Validate(); err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if got, want := loaded.LLM.Model, "gpt-4o-mini"; got != want {
		t.Fatalf("Model mismatch: got %q want %q", got, want)
	}
	if got := loaded.IO.InputDir; got != c.IO.InputDir {
		t.Fatalf("InputDir mismatch: got %q want %q", got, c.IO.InputDir)
	}
	if got := loaded.IO.OutputDir; got != c.IO.OutputDir {
		t.Fatalf("OutputDir mismatch: got %q want %q", got, c.IO.OutputDir)
	}
}

func TestEnvOverridesAndValidation(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".agentflow", "config.json")
	c := DefaultConfig("Demo", "gpt-4o-mini")
	c.IO.InputDir = filepath.Join(dir, "input")
	c.IO.OutputDir = filepath.Join(dir, "output")
	if err := Save(cfgPath, c); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Set env overrides
	t.Setenv("AGENTFLOW_MODEL", "gpt-4o-mini-2025")
	t.Setenv("AGENTFLOW_TEMPERATURE", "0.7")
	t.Setenv("AGENTFLOW_MAX_TOKENS", "12345")
	t.Setenv("AGENTFLOW_INPUT_DIR", filepath.Join(dir, "custom_in"))
	t.Setenv("AGENTFLOW_OUTPUT_DIR", filepath.Join(dir, "custom_out"))

	loaded, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	loaded.ApplyEnv()
	if err := loaded.Validate(); err != nil {
		t.Fatalf("Validate after env: %v", err)
	}
	if got, want := loaded.LLM.Model, "gpt-4o-mini-2025"; got != want {
		t.Fatalf("Model env override got %q want %q", got, want)
	}
	if got := loaded.LLM.Temperature; got < 0.69 || got > 0.71 {
		t.Fatalf("Temperature env override got %v want ~0.7", got)
	}
	if got, want := loaded.LLM.MaxTokens, 12345; got != want {
		t.Fatalf("MaxTokens env override got %d want %d", got, want)
	}
	if got, want := loaded.IO.InputDir, filepath.Join(dir, "custom_in"); got != want {
		t.Fatalf("InputDir env override got %q want %q", got, want)
	}
	if got, want := loaded.IO.OutputDir, filepath.Join(dir, "custom_out"); got != want {
		t.Fatalf("OutputDir env override got %q want %q", got, want)
	}
}

func TestValidateConstraints(t *testing.T) {
	c := DefaultConfig("Demo", "gpt-4o-mini")
	c.IO.InputDir = "in"
	c.IO.OutputDir = "out"
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected validate error: %v", err)
	}
	c.LLM.Temperature = 3.5
	if err := c.Validate(); err == nil {
		t.Fatalf("expected temperature out-of-range error")
	}
	c.LLM.Temperature = 0.2
	c.LLM.MaxTokens = 0
	if err := c.Validate(); err == nil {
		t.Fatalf("expected maxTokens > 0 error")
	}
}

func TestRedactedEnv(t *testing.T) {
	c := DefaultConfig("Demo", "gpt-4o-mini")
	t.Setenv("OPENAI_API_KEY", "secret1")
	t.Setenv("LANGGRAPH_API_KEY", "secret2")
	m := c.RedactedEnv()
	if m["OPENAI_API_KEY"] != "***" {
		t.Fatalf("expected OPENAI_API_KEY to be redacted")
	}
	if m["LANGGRAPH_API_KEY"] != "***" {
		t.Fatalf("expected LANGGRAPH_API_KEY to be redacted")
	}
}
