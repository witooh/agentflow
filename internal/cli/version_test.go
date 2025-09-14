package cli

import (
	"bytes"
	"strings"
	"testing"

	"agentflow/internal/buildinfo"
	"github.com/spf13/viper"
)

func TestVersionFlagOutputsBuildInfo(t *testing.T) {
	// Set predictable buildinfo for test
	buildinfo.Version = "test-version"
	buildinfo.Commit = "abc1234"
	buildinfo.Date = "2025-09-15"

	v := viper.New()
	cmd := NewRootCmd(v)

	bufOut := new(bytes.Buffer)
	bufErr := new(bytes.Buffer)
	cmd.SetOut(bufOut)
	cmd.SetErr(bufErr)
	cmd.SetArgs([]string{"--version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error executing --version, got: %v", err)
	}

	out := bufOut.String()
	if !strings.Contains(out, "agentflow test-version") {
		t.Fatalf("expected version output to contain agentflow test-version, got: %s", out)
	}
	if !strings.Contains(out, "abc1234") || !strings.Contains(out, "2025-09-15") {
		t.Fatalf("expected output to contain commit and date, got: %s", out)
	}
}
