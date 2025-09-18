package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agentflow/internal/config"
)

func TestParseTasks_AssignIDs_RenderList(t *testing.T) {
	input := strings.Join([]string{
		"- [ ] First task",
		"- [x] Second task",
	}, "\n")

	tasks := parseTasks(input)
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
	if tasks[0].Title != "First task" {
		t.Fatalf("unexpected first task parse: %#v", tasks[0])
	}
	if !strings.Contains(strings.ToLower(tasks[1].Title), "second") || !tasks[1].Checked {
		t.Fatalf("unexpected second task parse: %#v", tasks[1])
	}

	tasks = ensureScaffoldFirst(tasks)
	assignTaskIDs(tasks)
	out := renderTaskList(tasks)
	if !strings.Contains(out, "# Task List") || !strings.Contains(out, "TASK-001") || !strings.Contains(out, "TASK-002") {
		t.Fatalf("renderTaskList missing expected content: %s", out)
	}
}

func TestBuildCompactContext_TruncatesAndOrders(t *testing.T) {
	dir := t.TempDir()
	must := func(name, content string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	must("requirements.md", "R\n\nMore R")
	must("srs.md", "SRS content")
	must("stories.md", strings.Repeat("X", 50))

	got := buildCompactContext(dir, 30)
	if got == "" {
		t.Fatal("expected non-empty compact context")
	}
	// Ensure truncation to at most 30 runes
	if len([]rune(got)) != 30 {
		t.Fatalf("expected truncated to 30 runes, got %d", len([]rune(got)))
	}
}

func TestRenderTaskFile_DefaultAndCustomSubtasks(t *testing.T) {
	cfg := config.DefaultConfig("proj", "gpt-5")
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "requirements.md"), []byte("reqs"), 0o644)

	// Default subtasks
	t1 := devTask{ID: "TASK-001", Title: "Do thing"}
	out1 := renderTaskFile(cfg, t1, dir)
	if !strings.Contains(out1, "```xml") || !strings.Contains(out1, "<subtask>") || !strings.Contains(out1, "Analyze requirements") {
		t.Fatalf("renderTaskFile missing default subtasks or xml block: %s", out1)
	}

	// Custom subtasks from parse
	t2 := devTask{ID: "TASK-002", Title: "Do other", Subtasks: []string{"- [ ] A", "- [x] B"}}
	out2 := renderTaskFile(cfg, t2, dir)
	if !strings.Contains(out2, "- [ ] A") || !strings.Contains(out2, "- [x] B") {
		t.Fatalf("renderTaskFile should include provided subtasks: %s", out2)
	}
}

func TestDevPlan_DryRun_WritesFiles(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "out")
	_ = os.MkdirAll(outDir, 0o755)

	cfg := config.DefaultConfig("Proj", "gpt-5")
	cfg.IO.OutputDir = outDir
	cfgPath := filepath.Join(dir, ".agentflow", "config.json")
	if err := config.Save(cfgPath, cfg); err != nil {
		t.Fatalf("save cfg: %v", err)
	}

	// Provide minimal context so function does not return ErrNoContextDocs
	_ = os.WriteFile(filepath.Join(outDir, "requirements.md"), []byte("req"), 0o644)

	if err := DevPlan(DevPlanOptions{
		ConfigPath: cfgPath,
		SourceDir:  outDir,
		OutputDir:  outDir,
		Role:       "dev",
		DryRun:     true,
	}); err != nil {
		t.Fatalf("DevPlan dry-run failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outDir, "task_list.md")); err != nil {
		t.Fatalf("missing task_list.md: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "tasks", "TASK-001.md")); err != nil {
		t.Fatalf("missing tasks/TASK-001.md: %v", err)
	}
}

func TestDevPlan_NoContextDocs_ReturnsErr(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "out")
	_ = os.MkdirAll(outDir, 0o755)

	cfg := config.DefaultConfig("Proj", "gpt-5")
	cfg.IO.OutputDir = outDir
	cfgPath := filepath.Join(dir, ".agentflow", "config.json")
	if err := config.Save(cfgPath, cfg); err != nil {
		t.Fatalf("save cfg: %v", err)
	}

	err := DevPlan(DevPlanOptions{
		ConfigPath: cfgPath,
		SourceDir:  filepath.Join(dir, "no-such"),
		OutputDir:  outDir,
		Role:       "dev",
		DryRun:     true,
	})
	if err == nil || err != ErrNoContextDocs {
		t.Fatalf("expected ErrNoContextDocs, got %v", err)
	}
	// still writes scaffold files
	if _, err := os.Stat(filepath.Join(outDir, "task_list.md")); err != nil {
		t.Fatalf("missing task_list.md after scaffold: %v", err)
	}
}
