package main

import (
	"os"
	"path/filepath"
	"testing"
)

// We test only argument parsing branches that don't exit the process.

func TestUsageDoesNotPanic(t *testing.T) {
	// Just ensure the function runs; it writes to stdout
	usage()
}

func TestInitCmd_CreatesConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".agentflow", "config.json")
	initCmd([]string{"-project-name", "X", "-model", "gpt-5", "-config", cfgPath})
	if _, err := os.Stat(cfgPath); err != nil {
		t.Fatalf("config not created: %v", err)
	}
}

func TestPlanCmd_AndOthers_DryRun(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".agentflow", "config.json")
	// init
	initCmd([]string{"-project-name", "X", "-model", "gpt-5", "-config", cfgPath})

	outDir := filepath.Join(dir, "out")
	_ = os.MkdirAll(outDir, 0o755)
	// intake dry-run
	intakeCmd([]string{"-config", cfgPath, "-input", filepath.Join(dir, "input"), "-output", outDir, "-dry-run"})
	// plan dry-run; ensure requirements exists
	_ = os.WriteFile(filepath.Join(outDir, "requirements.md"), []byte("# req"), 0o644)
	planCmd([]string{"-config", cfgPath, "-requirements", filepath.Join(outDir, "requirements.md"), "-output", outDir, "-dry-run"})
	// design dry-run
	designCmd([]string{"-config", cfgPath, "-source", outDir, "-output", outDir, "-dry-run"})
	// qa dry-run
	qaCmd([]string{"-config", cfgPath, "-source", outDir, "-output", outDir, "-dry-run"})
	// devplan dry-run
	devplanCmd([]string{"-config", cfgPath, "-source", outDir, "-output", outDir, "-dry-run"})
}
