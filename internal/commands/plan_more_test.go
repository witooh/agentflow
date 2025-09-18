package commands

import "testing"

func TestEnsureSRS_HeuristicWhenMissing(t *testing.T) {
	s := ensureSRS("minimal text without interfaces")
	if s == "" {
		t.Fatal("ensureSRS should never return empty")
	}
}

func TestEnsureAC_DefaultWhenEmpty(t *testing.T) {
	if ensureAC("") == "" {
		t.Fatal("ensureAC should provide default")
	}
}
