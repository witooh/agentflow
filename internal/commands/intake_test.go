package commands

import (
	"testing"
)

func TestEnsureRequirementsSections_AddsMissing(t *testing.T) {
	in := "# Title\nSome content without required headings"
	timeline := "- 2025-09-13 — First idea\n- 2025-09-14 — More details"
	out := ensureRequirementsSections(in, timeline)
	checks := []string{
		"## Goals",
		"## Scope",
		"## Functional Requirements (FR)",
		"## Non-Functional Requirements (NFR)",
		"## Assumptions",
		"## Constraints",
		"## Timeline Summary",
		"## Open Questions",
		"## Questions to Human",
	}
	for _, c := range checks {
		if !contains(out, c) {
			t.Fatalf("missing section %q in output: %s", c, out)
		}
	}
	if !contains(out, "2025-09-13") || !contains(out, "2025-09-14") {
		t.Fatalf("timeline not injected: %s", out)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(s) > len(sub) && indexOf(s, sub) >= 0))
}

func indexOf(s, sub string) int {
	// simple search without importing strings to keep this test lightweight
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
