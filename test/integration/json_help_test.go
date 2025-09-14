package integration_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"agentflow/internal/cli"
	"github.com/spf13/viper"
)

type rootPayload struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	OK      bool   `json:"ok"`
}

func TestRootJSONPlaceholderIntegration(t *testing.T) {
	v := viper.New()
	cmd := cli.NewRootCmd(v)

	bufOut := new(bytes.Buffer)
	bufErr := new(bytes.Buffer)
	cmd.SetOut(bufOut)
	cmd.SetErr(bufErr)
	cmd.SetArgs([]string{"--project-dir", ".", "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	var payload rootPayload
	if err := json.Unmarshal(bufOut.Bytes(), &payload); err != nil {
		t.Fatalf("expected valid JSON, got error: %v; out=%s", err, bufOut.String())
	}
	if payload.Name == "" || payload.Version == "" || !payload.OK {
		t.Fatalf("unexpected payload values: %+v", payload)
	}

	if got := v.GetString("project_dir"); got != "." {
		t.Fatalf("expected viper project_dir='.', got: %s", got)
	}
}
