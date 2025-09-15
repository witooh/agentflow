package commands

import "testing"

func TestSplitPlanContent_AllMarkers(t *testing.T) {
	in := `--- SRS START ---
SRS body
--- STORIES START ---
Stories body
--- AC START ---
AC body`
	srs, stories, ac := splitPlanContent(in)
	if srs != "SRS body" || stories != "Stories body" || ac != "AC body" {
		t.Fatalf("unexpected split: %q | %q | %q", srs, stories, ac)
	}
}

func TestSplitPlanContent_NoMarkers_AllSRS(t *testing.T) {
	in := "Only one document"
	srs, stories, ac := splitPlanContent(in)
	if srs != in || stories != "" || ac != "" {
		t.Fatalf("expected all to be SRS; got %q | %q | %q", srs, stories, ac)
	}
}

func TestEnsureSRS_FallbackOnEmpty(t *testing.T) {
	out := ensureSRS("")
	if out == "" {
		t.Fatal("ensureSRS should provide a fallback when empty")
	}
}

func TestEnsureStories_FallbackOnEmpty(t *testing.T) {
	out := ensureStories("")
	if out == "" {
		t.Fatal("ensureStories should provide a fallback when empty")
	}
}

func TestEnsureAC_FallbackOnEmpty(t *testing.T) {
	out := ensureAC("")
	if out == "" {
		t.Fatal("ensureAC should provide a fallback when empty")
	}
}
