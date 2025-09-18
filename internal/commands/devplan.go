package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"agentflow/internal/agents"
	"agentflow/internal/config"
	"agentflow/internal/prompt"
)

type DevPlanOptions struct {
	ConfigPath string
	// SourceDir is where we read prior generated docs to build context (defaults to cfg.IO.OutputDir)
	SourceDir string
	// OutputDir is where we write task_list.md and tasks/*.md (defaults to cfg.IO.OutputDir)
	OutputDir string
	Role      string // usually "dev"
	DryRun    bool
}

var ErrNoContextDocs = errors.New("no context docs found for devplan")

func DevPlan(opts DevPlanOptions) error {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	cfg.ApplyEnv()
	if err := cfg.Validate(); err != nil {
		return err
	}

	// Directories
	sourceDir := opts.SourceDir
	if strings.TrimSpace(sourceDir) == "" {
		sourceDir = cfg.IO.OutputDir
	}
	outDir := opts.OutputDir
	if strings.TrimSpace(outDir) == "" {
		outDir = cfg.IO.OutputDir
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(outDir, "tasks"), 0o755); err != nil {
		return err
	}

	role := opts.Role
	if role == "" {
		role = "dev"
	}
	tpl := cfg.Roles[role]
	if strings.TrimSpace(tpl) == "" {
		tpl = "You are a Tech Lead. Produce a development plan."
	}

	// Build prompt from prior outputs
	extra := devPlanExtraSpec(cfg)
	p, files, err := prompt.BuildForRole(prompt.BuildOptions{
		RoleTemplate: tpl,
		InputsDir:    sourceDir,
		ExtraContext: extra,
	})
	if err != nil {
		return err
	}

	// Decide generation path
	var taskList string
	var runID string
	if len(files) == 0 {
		// still proceed with scaffold
		taskList = scaffoldTaskList()
	} else if opts.DryRun {
		taskList = scaffoldTaskList()
	} else {
		resp, err := agents.SA.Run(context.Background(), p)
		if err != nil {
			// fallback to scaffold
			taskList = scaffoldTaskList() + fmt.Sprintf("\n\n> Note: OpenAI call failed, wrote scaffold instead. Error: %v\n", err)
		} else {
			// runID = resp.RunID
			taskList = ensureCheckboxList(resp)
		}
	}

	// Ensure first task is scaffold project structure
	tasks := parseTasks(taskList)
	tasks = ensureScaffoldFirst(tasks)
	// Renumber and assign IDs
	assignTaskIDs(tasks)

	// Write task_list.md with metadata header
	listBody := renderTaskList(tasks)
	if err := writeFileWithHeader(cfg, role, runID, filepath.Join(sourceDir, "requirements.md"), filepath.Join(outDir, "task_list.md"), listBody); err != nil {
		return err
	}

	// For each task, write tasks/TASK-XXX.md with XML sections
	for _, t := range tasks {
		path := filepath.Join(outDir, "tasks", fmt.Sprintf("%s.md", t.ID))
		content := renderTaskFile(cfg, t, sourceDir)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return err
		}
	}

	if len(files) == 0 {
		return ErrNoContextDocs
	}
	return nil
}

type devTask struct {
	ID       string
	Title    string
	Checked  bool
	Subtasks []string
}

func devPlanExtraSpec(cfg *config.Config) string {
	return strings.TrimSpace(`ข้อกำหนดอัปเดตสำหรับคำสั่ง "agentflow devplan":
- ให้สร้าง task list เป็น Markdown ที่มี checkbox แต่ละ task เช่น "- [ ] TASK: <title>" และสามารถมี subtasks แบบ indented ได้
- จาก task list ให้สร้างไฟล์ราย task ในโฟลเดอร์ tasks/ รูปแบบไฟล์คือ TASK-XXX.md (XXX เป็นหมายเลขสามหลักเรียงลำดับ)
- เนื้อหาในไฟล์ต้องเป็น section แบบสไตล์ XML ได้แก่:
  <task>รายละเอียดของ task แบบ markdown</task>
  <context>สรุปย่อจากข้อมูลที่เกี่ยวข้องกับ task นี้</context>
  <implement>รายละเอียดในการ implement</implement>
  <subtask>รายการ subtask เป็น checkbox list</subtask>
  <dod>Definition of Done (อ้างอิง subtask ถ้ามี)</dod>
  สามารถเพิ่ม section อื่นๆ ที่จำเป็นได้ เช่น <risks>, <notes>
- ต้องเคารพ limit ของบริบทต่อ task โดยสรุป context ไม่เกิน devplan.maxContextCharsPerTask = %d อักขระ
- หากใน list ไม่มี task สำหรับการ scaffold โครงสร้างโปรเจกต์ ให้เพิ่มเป็นข้อแรกชื่อ "Project Scaffold / Bootstrap"
- หลีกเลี่ยงรายละเอียดเกินขนาดหน้าต่าง context ของ agent
`)
}

func scaffoldTaskList() string {
	return strings.Join([]string{
		"# Dev Tasks",
		"- [ ] Project Scaffold / Bootstrap",
		"- [ ] Implement devplan command",
		"- [ ] Integrate LangGraph server",
		"- [ ] Documentation update",
	}, "\n")
}

var checkboxRe = regexp.MustCompile(`(?i)^(\s*)-\s*\[( |x)\]\s*(.+?)\s*$`)

func ensureCheckboxList(s string) string {
	return strings.TrimSpace(s)
}

func parseTasks(s string) []devTask {
	var out []devTask
	lines := strings.Split(s, "\n")
	lastTop := -1
	for _, ln := range lines {
		m := checkboxRe.FindStringSubmatch(ln)
		if m == nil {
			continue
		}
		indent := m[1]
		checked := strings.ToLower(m[2]) == "x"
		text := strings.TrimSpace(m[3])
		if strings.TrimSpace(indent) == "" { // top-level task
			out = append(out, devTask{Title: text, Checked: checked})
			lastTop = len(out) - 1
		} else if lastTop >= 0 { // subtask of the previous top-level task
			prefix := "- [ ] "
			if checked {
				prefix = "- [x] "
			}
			out[lastTop].Subtasks = append(out[lastTop].Subtasks, prefix+text)
		}
	}
	return out
}

func ensureScaffoldFirst(tasks []devTask) []devTask {
	if len(tasks) == 0 {
		return []devTask{{Title: "Project Scaffold / Bootstrap"}}
	}
	firstIsScaffold := strings.Contains(strings.ToLower(tasks[0].Title), "scaffold") || strings.Contains(strings.ToLower(tasks[0].Title), "bootstrap")
	if firstIsScaffold {
		return tasks
	}
	// Check if any existing task is scaffold-like
	idx := -1
	for i, t := range tasks {
		low := strings.ToLower(t.Title)
		if strings.Contains(low, "scaffold") || strings.Contains(low, "bootstrap") {
			idx = i
			break
		}
	}
	if idx >= 0 {
		// move to front
		sc := tasks[idx]
		rest := append([]devTask{}, tasks[:idx]...)
		rest = append(rest, tasks[idx+1:]...)
		return append([]devTask{sc}, rest...)
	}
	// insert new scaffold
	return append([]devTask{{Title: "Project Scaffold / Bootstrap"}}, tasks...)
}

func assignTaskIDs(tasks []devTask) {
	for i := range tasks {
		id := fmt.Sprintf("TASK-%03d", i+1)
		tasks[i].ID = id
	}
}

func renderTaskList(tasks []devTask) string {
	b := &strings.Builder{}
	b.WriteString("# Task List\n")
	b.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format(time.RFC3339)))
	for _, t := range tasks {
		chk := " "
		if t.Checked {
			chk = "x"
		}
		b.WriteString(fmt.Sprintf("- [%s] %s — %s\n", chk, t.ID, t.Title))
	}
	return b.String()
}

func renderTaskFile(cfg *config.Config, t devTask, sourceDir string) string {
	// Build compact context from prior docs
	ctx := buildCompactContext(sourceDir, cfg.DevPlan.MaxContextCharsPerTask)
	impl := fmt.Sprintf("Implement task '%s' in the codebase.\n", t.Title)
	dod := "- [ ] Code implemented\n- [ ] Tests updated (if applicable)\n- [ ] Docs updated\n"
	// Subtasks: use parsed ones if present; otherwise, provide a reasonable default scaffold
	subtaskBlock := "- [ ] Analyze requirements\n- [ ] Design changes\n- [ ] Implement\n- [ ] Review & Test"
	if len(t.Subtasks) > 0 {
		subtaskBlock = strings.Join(t.Subtasks, "\n")
	}
	return strings.Join([]string{
		fmt.Sprintf("# %s — %s\n**Date:** %s\n", t.ID, t.Title, time.Now().Format("2006-01-02")),
		"```xml",
		fmt.Sprintf("<task>%s</task>", mdEscape(t.Title)),
		fmt.Sprintf("<context>%s</context>", xmlEscape(ctx)),
		fmt.Sprintf("<implement>%s</implement>", xmlEscape(impl)),
		fmt.Sprintf("<subtask>%s</subtask>", xmlEscape(subtaskBlock)),
		fmt.Sprintf("<dod>%s</dod>", xmlEscape(dod)),
		"```\n",
	}, "\n")
}

func buildCompactContext(dir string, max int) string {
	// Prefer these files if present
	candidates := []string{"requirements.md", "srs.md", "stories.md", "acceptance_criteria.md", "architecture.md", "uml.md"}
	var parts []string
	for _, name := range candidates {
		path := filepath.Join(dir, name)
		b, err := os.ReadFile(path)
		if err == nil {
			parts = append(parts, fmt.Sprintf("# %s\n%s", name, string(b)))
		}
	}
	if len(parts) == 0 {
		return ""
	}
	joined := strings.Join(parts, "\n\n---\n\n")
	if max > 0 && len([]rune(joined)) > max {
		// truncate by rune count to avoid cutting multibyte
		r := []rune(joined)
		joined = string(r[:max])
	}
	return joined
}

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func mdEscape(s string) string { return s }
