package integration_test

import (
	"bytes"
	"strings"
	"testing"

	"agentflow/internal/cli"

	"github.com/spf13/viper"
)

// Test that flags are bound to Viper and help still works — simulates a higher-level invocation.
func TestHelpAndFlagBinding(t *testing.T) {
	v := viper.New()
	cmd := cli.NewRootCmd(v)

	bufOut := new(bytes.Buffer)
	bufErr := new(bytes.Buffer)
	cmd.SetOut(bufOut)
	cmd.SetErr(bufErr)
	cmd.SetArgs([]string{"--project-dir", "./demo", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error executing with flags + help, got: %v", err)
	}

	if got := v.GetString("project_dir"); got != "./demo" {
		t.Fatalf("expected viper project_dir=./demo, got: %s", got)
	}

	out := bufOut.String()
	if !strings.Contains(out, "agentflow CLI") {
		t.Fatalf("expected help output to contain header; got:\n%s", out)
	}
}
