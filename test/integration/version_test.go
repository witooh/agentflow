package integration_test

import (
	"bytes"
	"strings"
	"testing"

	"agentflow/internal/buildinfo"
	"agentflow/internal/cli"
	"github.com/spf13/viper"
)

func TestVersionFlagIntegration(t *testing.T) {
	// Override build info for predictable output during test
	buildinfo.Version = "0.0.0-test"
	buildinfo.Commit = "deadbeef"
	buildinfo.Date = "2025-09-15"

	v := viper.New()
	cmd := cli.NewRootCmd(v)

	bufOut := new(bytes.Buffer)
	bufErr := new(bytes.Buffer)
	cmd.SetOut(bufOut)
	cmd.SetErr(bufErr)
	cmd.SetArgs([]string{"--version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error executing --version, got: %v", err)
	}

	out := bufOut.String()
	if !strings.Contains(out, "agentflow 0.0.0-test") {
		t.Fatalf("expected version output to contain agentflow 0.0.0-test, got: %s", out)
	}
	if !strings.Contains(out, "deadbeef") || !strings.Contains(out, "2025-09-15") {
		t.Fatalf("expected output to contain commit and date, got: %s", out)
	}
}
