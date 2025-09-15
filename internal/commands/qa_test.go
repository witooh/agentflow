package commands

import "testing"

func TestExtractTestPlan(t *testing.T) {
	in := "preamble\n--- TESTPLAN START ---\n## Test Strategy\nX"
	got := extractTestPlan(in)
	if got == "" || got[0] != '#' {
		t.Fatalf("extractTestPlan did not return content after marker: %q", got)
	}
}

func TestEnsureTestPlanHasSections(t *testing.T) {
	plan := "## Test Strategy\nA"
	got := ensureTestPlan(plan)
	// A few key sections must exist
	keys := []string{
		"## Scope",
		"## Test Types",
		"## Mapping to Acceptance Criteria",
		"## Entry/Exit Criteria",
		"## Risks & Mitigations",
	}
	for _, k := range keys {
		if !containsQA(got, k) {
			t.Fatalf("ensureTestPlan missing section %q in:\n%s", k, got)
		}
	}
}

func containsQA(s, sub string) bool {
	return len(s) >= len(sub) && (indexQA(s, sub) >= 0)
}

// small helper to avoid importing strings; deterministic for tests
func indexQA(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		match := true
		for j := 0; j < len(needle); j++ {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}
