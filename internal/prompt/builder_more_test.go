package prompt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildForRole_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	p, files, err := BuildForRole(BuildOptions{RoleTemplate: "ROLE", InputsDir: dir, ExtraContext: "EXTRA"})
	if err != nil {
		t.Fatalf("BuildForRole error: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected 0 files, got %d", len(files))
	}
	if !strings.Contains(p, "SYSTEM:\nROLE") || !strings.Contains(p, "EXTRA:\nEXTRA") {
		t.Fatalf("prompt missing SYSTEM/EXTRA sections: %s", p)
	}
}

func TestBuildForRole_WithFiles_OrdersAndEmbeds(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	write("b.txt", "B content")
	write("a.md", "A content")

	p, files, err := BuildForRole(BuildOptions{RoleTemplate: "ROLE", InputsDir: dir})
	if err != nil {
		t.Fatalf("BuildForRole error: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
	// files are full paths; ensure order corresponds to lexicographic path order
	if filepath.Base(files[0]) != "a.md" || filepath.Base(files[1]) != "b.txt" {
		t.Fatalf("unexpected order: %#v", files)
	}
	if !strings.Contains(p, "# File: a.md") || !strings.Contains(p, "A content") {
		t.Fatalf("prompt missing content for a.md: %s", p)
	}
	if !strings.Contains(p, "# File: b.txt") || !strings.Contains(p, "B content") {
		t.Fatalf("prompt missing content for b.txt: %s", p)
	}
}
