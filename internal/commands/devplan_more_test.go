package commands

import "testing"

func TestXMLAndMDEscape(t *testing.T) {
	in := "<tag>&data>"
	got := xmlEscape(in)
	if got != "&lt;tag&gt;&amp;data&gt;" {
		t.Fatalf("xmlEscape wrong: %q", got)
	}
	if mdEscape("x") != "x" {
		t.Fatal("mdEscape should be identity")
	}
}

func TestEnsureScaffoldFirst_Reorders(t *testing.T) {
	tasks := []devTask{{Title: "Implement devplan"}, {Title: "Project Scaffold / Bootstrap"}}
	out := ensureScaffoldFirst(tasks)
	if len(out) != 2 || out[0].Title != "Project Scaffold / Bootstrap" {
		t.Fatalf("ensureScaffoldFirst should move scaffold to front: %#v", out)
	}
}
