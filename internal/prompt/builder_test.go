package prompt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildTimelineSummary_SortsByDateAndIgnoresNonDated(t *testing.T) {
	dir := t.TempDir()
	mustWrite := func(name, content string) {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	mustWrite("2025-09-14.md", "Second day\nAdded more details")
	mustWrite("2025-09-13.md", "First day\nInitial idea")
	mustWrite("notes.md", "random note")

	timeline, names, err := BuildTimelineSummary(dir)
	if err != nil {
		t.Fatalf("BuildTimelineSummary error: %v", err)
	}
	if len(names) != 2 {
		t.Fatalf("expected 2 dated files, got %d", len(names))
	}
	if names[0] != "2025-09-13.md" || names[1] != "2025-09-14.md" {
		t.Fatalf("unexpected order: %#v", names)
	}
	if !(strings.Contains(timeline, "2025-09-13") && strings.Contains(timeline, "2025-09-14")) {
		t.Fatalf("timeline missing dates: %q", timeline)
	}
}
