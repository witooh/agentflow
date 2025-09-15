package commands

import (
	"fmt"
	"os"

	"agentflow/internal/config"
)

// Init creates .agentflow/config.json with defaults
func Init(configPath, projectName, baseURL, model string) error {
	cfg := config.DefaultConfig(projectName, baseURL, model)
	// Ensure base directories
	if err := config.EnsureDirs(configPath, cfg); err != nil {
		return err
	}
	if err := config.Save(configPath, cfg); err != nil {
		return err
	}
	// Also generate a .gitignore suggestion for output if repo root
	writeGitignoreSuggestion(cfg.IO.OutputDir)
	return nil
}

func writeGitignoreSuggestion(outputDir string) {
	path := ".agentflow/.gitignore"
	_ = os.MkdirAll(".agentflow", 0o755)
	_ = os.WriteFile(path, []byte(fmt.Sprintf("# AgentFlow\n# Consider ignoring generated outputs if desired\n/%s/**\n", outputDir)), 0o644)
}
