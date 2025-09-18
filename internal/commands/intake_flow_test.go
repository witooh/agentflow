package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agentflow/internal/config"
)

func TestIntake_DryRun_WithInputs_WritesRequirements(t *testing.T) {
	dir := t.TempDir()
	inDir := filepath.Join(dir, "in")
	outDir := filepath.Join(dir, "out")
	_ = os.MkdirAll(inDir, 0o755)
	_ = os.MkdirAll(outDir, 0o755)

	// Create a couple of input files
	if err := os.WriteFile(filepath.Join(inDir, "2025-01-01.md"), []byte("First idea"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := os.WriteFile(filepath.Join(inDir, "notes.md"), []byte("some notes"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	cfg := config.DefaultConfig("Proj", "gpt-5")
	cfg.IO.InputDir = inDir
	cfg.IO.OutputDir = outDir
	cfgPath := filepath.Join(dir, ".agentflow", "config.json")
	if err := config.Save(cfgPath, cfg); err != nil {
		t.Fatalf("save cfg: %v", err)
	}

	if err := Intake(IntakeOptions{
		ConfigPath: cfgPath,
		InputsDir:  inDir,
		OutputDir:  outDir,
		Role:       "po_pm",
		DryRun:     true,
	}); err != nil {
		t.Fatalf("intake dry-run failed: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(outDir, "requirements.md"))
	if err != nil {
		t.Fatalf("read requirements: %v", err)
	}
	s := string(b)
	if !strings.Contains(s, "# requirements") || !strings.Contains(s, "## Goals") {
		t.Fatalf("requirements missing expected sections: %s", s)
	}
}

func TestIntake_DryRun_NoInputs_ReturnsErrNoInputs_WritesFile(t *testing.T) {
	dir := t.TempDir()
	inDir := filepath.Join(dir, "in")
	outDir := filepath.Join(dir, "out")
	_ = os.MkdirAll(inDir, 0o755)
	_ = os.MkdirAll(outDir, 0o755)

	cfg := config.DefaultConfig("Proj", "gpt-5")
	cfg.IO.InputDir = inDir
	cfg.IO.OutputDir = outDir
	cfgPath := filepath.Join(dir, ".agentflow", "config.json")
	if err := config.Save(cfgPath, cfg); err != nil {
		t.Fatalf("save cfg: %v", err)
	}

	err := Intake(IntakeOptions{
		ConfigPath: cfgPath,
		InputsDir:  inDir,
		OutputDir:  outDir,
		Role:       "po_pm",
		DryRun:     true,
	})
	if err == nil || err != ErrNoInputs {
		t.Fatalf("expected ErrNoInputs, got %v", err)
	}
	b, rerr := os.ReadFile(filepath.Join(outDir, "requirements.md"))
	if rerr != nil {
		t.Fatalf("requirements not written on no-inputs: %v", rerr)
	}
	if !strings.Contains(string(b), "No inputs found") {
		t.Fatalf("expected scaffold note in requirements: %s", string(b))
	}
}
