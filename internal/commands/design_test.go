package commands

import (
	"os"
	"path/filepath"
	"testing"

	"agentflow/internal/config"
)

func TestSplitDesignContent_AllMarkers(t *testing.T) {
	in := `--- ARCH START ---
ARCH body
--- UML START ---
UML body`
	arch, uml := splitDesignContent(in)
	if arch != "ARCH body" || uml != "UML body" {
		t.Fatalf("unexpected split: %q | %q", arch, uml)
	}
}

func TestEnsureArchitecture_FallbackOnEmpty(t *testing.T) {
	out := ensureArchitecture("")
	if out == "" || !contains(out, "Project Structure") {
		t.Fatal("ensureArchitecture should provide a fallback with Project Structure")
	}
	if !contains(out, "plantuml") {
		t.Fatal("ensureArchitecture should include a PlantUML fence when missing")
	}
}

func TestEnsureUML_FallbackOnEmpty(t *testing.T) {
	out := ensureUML("")
	if out == "" {
		t.Fatal("ensureUML should provide fallback sections")
	}
	if !contains(out, "Sequence:") || !contains(out, "Class:") || !contains(out, "Activity:") {
		t.Fatalf("ensureUML should include sequence/class/activity diagrams; got: %s", out)
	}
}

func TestDesign_DryRun_WritesFiles(t *testing.T) {
	// Prepare temp workspace
	dir := t.TempDir()
	outDir := filepath.Join(dir, "out")
	_ = os.MkdirAll(outDir, 0o755)

	cfg := config.DefaultConfig("TestProject", "gpt-4o-mini")
	cfg.IO.OutputDir = outDir
	cfgPath := filepath.Join(dir, ".agentflow", "config.json")
	if err := config.Save(cfgPath, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	// Write minimal prior docs to source
	_ = os.WriteFile(filepath.Join(outDir, "srs.md"), []byte("SRS"), 0o644)
	_ = os.WriteFile(filepath.Join(outDir, "stories.md"), []byte("Stories"), 0o644)
	_ = os.WriteFile(filepath.Join(outDir, "acceptance_criteria.md"), []byte("AC"), 0o644)

	if err := Design(DesignOptions{
		ConfigPath: cfgPath,
		SourceDir:  outDir,
		OutputDir:  outDir,
		Role:       "sa",
		DryRun:     true,
	}); err != nil {
		t.Fatalf("design dry-run failed: %v", err)
	}

	arch := filepath.Join(outDir, "architecture.md")
	uml := filepath.Join(outDir, "uml.md")
	if _, err := os.Stat(arch); err != nil {
		t.Fatalf("missing architecture.md: %v", err)
	}
	if _, err := os.Stat(uml); err != nil {
		t.Fatalf("missing uml.md: %v", err)
	}
	b, _ := os.ReadFile(arch)
	if !contains(string(b), "Project Structure") || !contains(string(b), "plantuml") {
		t.Fatalf("architecture.md missing required sections: %s", string(b))
	}
	b2, _ := os.ReadFile(uml)
	if !contains(string(b2), "Sequence:") || !contains(string(b2), "Class:") || !contains(string(b2), "Activity:") {
		t.Fatalf("uml.md missing required diagrams: %s", string(b2))
	}
}
