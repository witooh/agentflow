package commands

import (
	"os"
	"path/filepath"
	"testing"

	"agentflow/internal/config"
)

func TestQA_DryRun_WritesFile(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "out")
	_ = os.MkdirAll(outDir, 0o755)

	// prior docs (any may be missing)
	_ = os.WriteFile(filepath.Join(outDir, "srs.md"), []byte("SRS"), 0o644)
	_ = os.WriteFile(filepath.Join(outDir, "stories.md"), []byte("Stories"), 0o644)

	cfg := config.DefaultConfig("Proj", "gpt-5")
	cfg.IO.OutputDir = outDir
	cfgPath := filepath.Join(dir, ".agentflow", "config.json")
	if err := config.Save(cfgPath, cfg); err != nil {
		t.Fatalf("save cfg: %v", err)
	}

	if err := QA(QAOptions{
		ConfigPath: cfgPath,
		SourceDir:  outDir,
		OutputDir:  outDir,
		Role:       "qa",
		DryRun:     true,
	}); err != nil {
		t.Fatalf("qa dry-run failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "test-plan.md")); err != nil {
		t.Fatalf("missing test-plan.md: %v", err)
	}
}
