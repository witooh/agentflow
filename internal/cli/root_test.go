package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestRootHelpOutput(t *testing.T) {
	v := viper.New()
	cmd := NewRootCmd(v)

	bufOut := new(bytes.Buffer)
	bufErr := new(bytes.Buffer)
	cmd.SetOut(bufOut)
	cmd.SetErr(bufErr)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error executing help, got: %v", err)
	}

	out := bufOut.String()
	if !strings.Contains(out, "agentflow CLI") || !strings.Contains(out, "Usage:") {
		t.Fatalf("help output missing expected content, got:\n%s", out)
	}
}
