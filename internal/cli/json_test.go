package cli

import (
	"bytes"
	"encoding/json"
	"testing"

	"agentflow/internal/buildinfo"
	"github.com/spf13/viper"
)

type rootJSON struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Capabilities []string `json:"capabilities"`
	OK           bool     `json:"ok"`
}

func TestRootJSONPlaceholder(t *testing.T) {
	// Make version deterministic
	buildinfo.Version = "test-json"

	v := viper.New()
	cmd := NewRootCmd(v)

	bufOut := new(bytes.Buffer)
	bufErr := new(bytes.Buffer)
	cmd.SetOut(bufOut)
	cmd.SetErr(bufErr)
	cmd.SetArgs([]string{"--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error executing --json, got: %v", err)
	}

	var payload rootJSON
	if err := json.Unmarshal(bufOut.Bytes(), &payload); err != nil {
		t.Fatalf("expected valid JSON output, got error: %v; out=%s", err, bufOut.String())
	}
	if payload.Name != "agentflow" {
		t.Fatalf("expected name=agentflow, got: %s", payload.Name)
	}
	if payload.Version != "test-json" {
		t.Fatalf("expected version=test-json, got: %s", payload.Version)
	}
	if !payload.OK {
		t.Fatalf("expected ok=true, got: %v", payload.OK)
	}
	if len(payload.Capabilities) == 0 {
		t.Fatalf("expected non-empty capabilities list")
	}
}
