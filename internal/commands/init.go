package commands

import (
	"agentflow/internal/config"
)

// Init creates .agentflow/config.json with defaults
func Init(configPath, projectName, model string) error {
	cfg := config.DefaultConfig(projectName, model)
	// Ensure base directories
	if err := config.EnsureDirs(configPath, cfg); err != nil {
		return err
	}
	if err := config.Save(configPath, cfg); err != nil {
		return err
	}
	return nil
}
