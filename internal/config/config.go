package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config mirrors the schema described in docs.
type Config struct {
	SchemaVersion string `json:"schemaVersion"`
	ProjectName   string `json:"projectName"`
	LangGraph     struct {
		BaseURL string `json:"baseUrl"`
	} `json:"langgraph"`
	LLM struct {
		Model       string  `json:"model"`
		Temperature float64 `json:"temperature"`
		MaxTokens   int     `json:"maxTokens"`
	} `json:"llm"`
	Roles map[string]string `json:"roles"`
	IO    struct {
		InputDir  string `json:"inputDir"`
		OutputDir string `json:"outputDir"`
	} `json:"io"`
	Security struct {
		EnvKeys []string `json:"envKeys"`
	} `json:"security"`
	Redact struct {
		Secrets bool `json:"secrets"`
	} `json:"redact"`
	DevPlan struct {
		MaxContextCharsPerTask int `json:"maxContextCharsPerTask"`
	} `json:"devplan"`
	AskHuman struct {
		Mode string `json:"mode"`
	} `json:"askHuman"`
	Metadata struct {
		Owner string   `json:"owner"`
		Repo  string   `json:"repo"`
		Tags  []string `json:"tags"`
	} `json:"metadata"`
}

func DefaultConfig(projectName, baseURL, model string) *Config {
	c := &Config{}
	c.SchemaVersion = "0.1"
	c.ProjectName = projectName
	c.LangGraph.BaseURL = baseURL
	c.LLM.Model = model
	c.LLM.Temperature = 0.2
	c.LLM.MaxTokens = 4000
	c.Roles = map[string]string{
		"po_pm": "You are a PO/PM. Convert input context into formal requirements with sections: Goals, Scope, FR, NFR, Assumptions, Open Questions.",
		"sa":    "You are a Solution Architect. Transform requirements into SRS/Stories/AC.",
		"qa":    "You are a QA Lead. Produce a concise test plan.",
		"dev":   "You are a Tech Lead. Produce dev task list and per-task context.",
	}
	c.IO.InputDir = "input"
	c.IO.OutputDir = "output"
	c.Security.EnvKeys = []string{"OPENAI_API_KEY"}
	c.Redact.Secrets = true
	c.DevPlan.MaxContextCharsPerTask = 4000
	c.AskHuman.Mode = "interactive"
	c.Metadata.Owner = ""
	c.Metadata.Repo = ""
	c.Metadata.Tags = []string{}
	return c
}

func EnsureDirs(configPath string, c *Config) error {
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return err
	}
	if c != nil {
		if c.IO.InputDir != "" {
			if err := os.MkdirAll(c.IO.InputDir, 0o755); err != nil {
				return err
			}
		}
		if c.IO.OutputDir != "" {
			if err := os.MkdirAll(c.IO.OutputDir, 0o755); err != nil {
				return err
			}
		}
	}
	return nil
}

func Save(path string, c *Config) error {
	if err := EnsureDirs(path, c); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Config) Validate() error {
	if c.SchemaVersion == "" {
		return errors.New("schemaVersion is required")
	}
	// Basic schema version check to allow forward evolution
	allowedVersions := map[string]bool{"0.1": true}
	if !allowedVersions[c.SchemaVersion] {
		return fmt.Errorf("unsupported schemaVersion: %s", c.SchemaVersion)
	}
	if strings.TrimSpace(c.ProjectName) == "" {
		return errors.New("projectName is required")
	}
	if strings.TrimSpace(c.LangGraph.BaseURL) == "" {
		return errors.New("langgraph.baseUrl is required")
	}
	if strings.TrimSpace(c.LLM.Model) == "" {
		return errors.New("llm.model is required")
	}
	if c.LLM.Temperature < 0 || c.LLM.Temperature > 2 {
		return fmt.Errorf("llm.temperature out of range [0,2]: %v", c.LLM.Temperature)
	}
	if c.LLM.MaxTokens <= 0 {
		return fmt.Errorf("llm.maxTokens must be > 0")
	}
	if strings.TrimSpace(c.IO.InputDir) == "" || strings.TrimSpace(c.IO.OutputDir) == "" {
		return fmt.Errorf("io.inputDir and io.outputDir are required")
	}
	return nil
}

// ApplyEnv overrides configuration fields from environment variables if set.
// Supported variables:
// - AGENTFLOW_BASE_URL → langgraph.baseUrl
// - AGENTFLOW_MODEL → llm.model
// - AGENTFLOW_TEMPERATURE → llm.temperature (float)
// - AGENTFLOW_MAX_TOKENS → llm.maxTokens (int)
// - AGENTFLOW_INPUT_DIR → io.inputDir
// - AGENTFLOW_OUTPUT_DIR → io.outputDir
func (c *Config) ApplyEnv() {
	if v := strings.TrimSpace(os.Getenv("AGENTFLOW_BASE_URL")); v != "" {
		c.LangGraph.BaseURL = v
	}
	if v := strings.TrimSpace(os.Getenv("AGENTFLOW_MODEL")); v != "" {
		c.LLM.Model = v
	}
	if v := strings.TrimSpace(os.Getenv("AGENTFLOW_TEMPERATURE")); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			c.LLM.Temperature = f
		}
	}
	if v := strings.TrimSpace(os.Getenv("AGENTFLOW_MAX_TOKENS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.LLM.MaxTokens = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("AGENTFLOW_INPUT_DIR")); v != "" {
		c.IO.InputDir = v
	}
	if v := strings.TrimSpace(os.Getenv("AGENTFLOW_OUTPUT_DIR")); v != "" {
		c.IO.OutputDir = v
	}
}

// RedactedEnv returns a map of configured secret env keys with redacted values
// Useful for safe logging/spec dumps without leaking secrets.
func (c *Config) RedactedEnv() map[string]string {
	m := map[string]string{}
	keys := append([]string{}, c.Security.EnvKeys...)
	// Include LANGGRAPH_API_KEY which is actually used by the client
	keys = append(keys, "LANGGRAPH_API_KEY")
	seen := map[string]bool{}
	for _, k := range keys {
		k = strings.TrimSpace(k)
		if k == "" || seen[k] {
			continue
		}
		seen[k] = true
		if _, ok := os.LookupEnv(k); ok {
			m[k] = "***"
		}
	}
	return m
}
