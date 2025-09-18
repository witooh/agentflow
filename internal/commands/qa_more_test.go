package commands

import "testing"

func TestExtractTestPlan_NoMarkerReturnsTrimmed(t *testing.T) {
	in := "  body  "
	if got := extractTestPlan(in); got != "body" {
		t.Fatalf("extractTestPlan trim failed: %q", got)
	}
}
