package commands

import (
	"os"
	"path/filepath"
	"testing"

	"agentflow/internal/config"
)

func TestPlan_DryRun_WritesFiles(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "out")
	_ = os.MkdirAll(outDir, 0o755)

	cfg := config.DefaultConfig("Proj", "gpt-5")
	cfg.IO.OutputDir = outDir
	cfgPath := filepath.Join(dir, ".agentflow", "config.json")
	if err := config.Save(cfgPath, cfg); err != nil {
		t.Fatalf("save cfg: %v", err)
	}

	// minimal requirements
	reqPath := filepath.Join(outDir, "requirements.md")
	if err := os.WriteFile(reqPath, []byte("# req\n"), 0o644); err != nil {
		t.Fatalf("write req: %v", err)
	}

	if err := Plan(PlanOptions{
		ConfigPath:   cfgPath,
		Requirements: reqPath,
		OutputDir:    outDir,
		Role:         "sa",
		DryRun:       true,
	}); err != nil {
		t.Fatalf("plan dry-run failed: %v", err)
	}

	for _, f := range []string{"srs.md", "stories.md", "acceptance_criteria.md"} {
		if _, err := os.Stat(filepath.Join(outDir, f)); err != nil {
			t.Fatalf("missing %s: %v", f, err)
		}
	}
}
