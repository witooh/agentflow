package commands

import (
	"strings"
	"testing"
)

func TestSplitDesignContent_Variants(t *testing.T) {
	s := "x\n--- UML START ---\nU\n--- ARCH START ---\nA"
	arch, uml := splitDesignContent(s)
	if arch == "" || uml == "" {
		t.Fatalf("expected both parts, got arch=%q uml=%q", arch, uml)
	}
}

func TestEnsureArchitecture_AddsMissingSections(t *testing.T) {
	out := ensureArchitecture("Overview only")
	if out == "" || !containsLower(out, "project structure") || !containsLower(out, "plantuml") {
		t.Fatalf("ensureArchitecture did not add required sections: %s", out)
	}
}

func TestEnsureUML_AddsMissingDiagrams(t *testing.T) {
	out := ensureUML("just text")
	if !containsLower(out, "sequence") || !containsLower(out, "class") || !containsLower(out, "activity") {
		t.Fatalf("ensureUML missing diagrams: %s", out)
	}
}

func containsLower(s, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}
