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

	// Build prompt from prior outputs
	extra := devPlanExtraSpec(cfg)
	files, err := prompt.GetInputFiles(sourceDir, []string{"requirements.md", "srs.md", "stories.md", "acceptance_criteria.md", "architecture.md", "uml.md"})
	if err != nil {
		return err
	}
	systemMessages, err := prompt.GetPromptFromFiles(files)
	if err != nil {
		return err
	}
	userMessage := agents.UserMessage(strings.Join([]string{
		"You are a Tech Lead. Produce a development plan.",
		extra,
	}, "\n\n"))
	prompts := agents.InputList(systemMessages, userMessage)

	var taskList string
	if len(files) == 0 {
		// still proceed with scaffold
		taskList = ""
	} else if opts.DryRun {
		taskList = drayRunTaskList
	} else {
		resp, err := agents.SA.RunInputs(context.Background(), prompts)
		if err != nil {
			// fallback to scaffold
			taskList = fmt.Sprintf("\n\n> Note: OpenAI call failed, wrote scaffold instead. Error: %v\n", err)
		} else {
			//delete tasks/ directory
			os.RemoveAll(filepath.Join(outDir, "tasks"))
			if err := os.MkdirAll(filepath.Join(outDir, "tasks"), 0o755); err != nil {
				return err
			}
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
	if err := writeFileWithHeader(cfg, filepath.Join(sourceDir, "requirements.md"), filepath.Join(outDir, "task_list.md"), listBody); err != nil {
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
	Context  string
}

func devPlanExtraSpec(cfg *config.Config) string {
	return strings.TrimSpace(`ข้อกำหนดอัปเดตสำหรับคำสั่ง "agentflow devplan":
- ให้สร้าง task list เป็น Markdown ที่มี checkbox แต่ละ task เช่น "- [ ] TASK: <title>" และสามารถมี subtasks แบบ indented ได้
- จาก task list ให้สร้างไฟล์ราย task ในโฟลเดอร์ tasks/ รูปแบบไฟล์คือ TASK-XXX.md (XXX เป็นหมายเลขสามหลักเรียงลำดับ)
- เนื้อหาในไฟล์ต้องเป็น section แบบสไตล์ XML ได้แก่:
  <task>รายละเอียดของ task แบบ markdown</task>
  <context>ข้อมูลที่จำเป็นจำหรับ implement task นี้แบบละเอียด</context>
  <implement>รายละเอียดในการ implement</implement>
  <subtask>รายการ subtask เป็น checkbox list</subtask>
  <dod>Definition of Done (อ้างอิง subtask ถ้ามี)</dod>
  สามารถเพิ่ม section อื่นๆ ที่จำเป็นได้ เช่น <risks>, <notes>
- ต้องเคารพ limit ของบริบทต่อ task โดยสรุป context ไม่เกิน devplan.maxContextCharsPerTask = %d อักขระ
- หากใน list ไม่มี task สำหรับการ scaffold โครงสร้างโปรเจกต์ ให้เพิ่มเป็นข้อแรกชื่อ "Project Scaffold / Bootstrap"
- หลีกเลี่ยงรายละเอียดเกินขนาดหน้าต่าง context ของ agent
`)
}

var checkboxRe = regexp.MustCompile(`(?i)^(\s*)-\s*\[( |x)\]\s*(.+?)\s*$`)

func ensureCheckboxList(s string) string {
	return strings.TrimSpace(s)
}

func parseTasks(s string) []devTask {
	var out []devTask
	lines := strings.Split(s, "\n")
	lastTop := -1

	// For context extraction: look for <context>...</context> blocks per task
	// We'll build a map: task title (normalized) -> context string
	contextMap := make(map[string]string)
	var currentTaskTitle, currentContext string
	// Track discovery order of <task> tags and collect subtasks per task title
	titlesOrder := []string{}
	subtaskMap := make(map[string][]string)

	// First pass: extract all <task> and <context> blocks
	for i := 0; i < len(lines); i++ {
		ln := strings.TrimSpace(lines[i])
		if strings.HasPrefix(ln, "<task>") && strings.HasSuffix(ln, "</task>") {
			// Single-line <task>...</task>
			currentTaskTitle = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(ln, "<task>"), "</task>"))
			if currentTaskTitle != "" {
				titlesOrder = append(titlesOrder, currentTaskTitle)
			}
		} else if strings.HasPrefix(ln, "<task>") {
			// Multi-line <task>
			currentTaskTitle = ""
			for j := i; j < len(lines); j++ {
				line := lines[j]
				if strings.Contains(line, "</task>") {
					currentTaskTitle += strings.TrimSpace(strings.ReplaceAll(line, "</task>", ""))
					i = j
					break
				}
				if j != i {
					currentTaskTitle += "\n"
				}
				currentTaskTitle += strings.TrimSpace(line)
			}
			if currentTaskTitle != "" {
				titlesOrder = append(titlesOrder, currentTaskTitle)
			}
		}
		if strings.HasPrefix(ln, "<context>") {
			currentContext = strings.TrimPrefix(ln, "<context>")
			if strings.Contains(currentContext, "</context>") {
				// Single-line context
				currentContext = strings.TrimSuffix(currentContext, "</context>")
			} else {
				currentContext = strings.TrimSpace(currentContext)
				// Multi-line context
				for j := i + 1; j < len(lines); j++ {
					line := lines[j]
					if strings.Contains(line, "</context>") {
						currentContext += "\n" + strings.TrimSuffix(line, "</context>")
						i = j
						break
					}
					currentContext += "\n" + line
				}
			}
			// Map context to the most recent <task>
			if currentTaskTitle != "" {
				contextMap[strings.ToLower(strings.TrimSpace(currentTaskTitle))] = strings.TrimSpace(currentContext)
			}
		}
		if strings.HasPrefix(ln, "<subtask>") {
			// Collect subtask block content
			sub := strings.TrimPrefix(ln, "<subtask>")
			if strings.Contains(sub, "</subtask>") {
				// Single-line subtask block
				sub = strings.TrimSuffix(sub, "</subtask>")
			} else {
				sub = strings.TrimSpace(sub)
				for j := i + 1; j < len(lines); j++ {
					line := lines[j]
					if strings.Contains(line, "</subtask>") {
						sub += "\n" + strings.TrimSuffix(line, "</subtask>")
						i = j
						break
					}
					sub += "\n" + line
				}
			}
			if currentTaskTitle != "" {
				key := strings.ToLower(strings.TrimSpace(currentTaskTitle))
				var subs []string
				for _, l := range strings.Split(sub, "\n") {
					tline := strings.TrimSpace(l)
					if tline == "" {
						continue
					}
					// Normalize to checkbox lines
					if m2 := checkboxRe.FindStringSubmatch(tline); m2 != nil {
						checked := strings.ToLower(m2[2]) == "x"
						text := strings.TrimSpace(m2[3])
						prefix := "- [ ] "
						if checked {
							prefix = "- [x] "
						}
						subs = append(subs, prefix+text)
					} else {
						// Convert dash list or plain line to unchecked checkbox
						if strings.HasPrefix(tline, "- ") {
							tline = strings.TrimSpace(strings.TrimPrefix(tline, "- "))
						}
						subs = append(subs, "- [ ] "+tline)
					}
				}
				subtaskMap[key] = subs
			}
		}
	}

	// Second pass: parse checkbox list and attach context if available
	for _, ln := range lines {
		m := checkboxRe.FindStringSubmatch(ln)
		if m == nil {
			continue
		}
		indent := m[1]
		checked := strings.ToLower(m[2]) == "x"
		text := strings.TrimSpace(m[3])
		if strings.TrimSpace(indent) == "" { // top-level task
			// Try to find context for this task
			ctx := ""
			// Try exact match, then prefix match (for "TASK: ..." etc)
			lowTitle := strings.ToLower(text)
			if v, ok := contextMap[lowTitle]; ok {
				ctx = v
			} else {
				for k, v := range contextMap {
					if strings.HasPrefix(lowTitle, k) || strings.HasPrefix(k, lowTitle) {
						ctx = v
						break
					}
				}
			}
			out = append(out, devTask{Title: text, Checked: checked, Context: ctx})
			lastTop = len(out) - 1
		} else if lastTop >= 0 { // subtask of the previous top-level task
			prefix := "- [ ] "
			if checked {
				prefix = "- [x] "
			}
			out[lastTop].Subtasks = append(out[lastTop].Subtasks, prefix+text)
		}
	}
	// Fallback: If no top-level checkbox tasks were found, but XML blocks were parsed, build tasks from them
	if len(out) == 0 && len(titlesOrder) > 0 {
		for _, title := range titlesOrder {
			key := strings.ToLower(strings.TrimSpace(title))
			ctx := ""
			if v, ok := contextMap[key]; ok {
				ctx = v
			}
			subs := subtaskMap[key]
			out = append(out, devTask{Title: strings.TrimSpace(title), Checked: false, Subtasks: subs, Context: strings.TrimSpace(ctx)})
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
	ctx := strings.TrimSpace(t.Context)
	if ctx == "" {
		// Fallback to compact context from prior docs
		ctx = buildCompactContext(sourceDir, cfg.DevPlan.MaxContextCharsPerTask)
	}
	impl := fmt.Sprintf("Implement task '%s' in the codebase.\n", t.Title)
	dod := "- [ ] Code implemented\n- [ ] Tests updated (if applicable)\n- [ ] Docs updated\n"
	// Subtasks: use parsed ones if present; otherwise, provide a reasonable default scaffold
	subtaskBlock := "- [ ] Analyze requirements\n- [ ] Design changes\n- [ ] Implement\n- [ ] Review & Test"
	if len(t.Subtasks) > 0 {
		subtaskBlock = strings.Join(t.Subtasks, "\n")
	}
	return strings.Join([]string{
		fmt.Sprintf("# %s — %s\n**Date:** %s\n", t.ID, t.Title, time.Now().Format("2006-01-02")), "",
		fmt.Sprintf("<task>%s</task>", mdEscape(t.Title)),
		fmt.Sprintf("<context>%s</context>", xmlEscape(ctx)),
		fmt.Sprintf("<implement>%s</implement>", xmlEscape(impl)),
		fmt.Sprintf("<subtask>%s</subtask>", xmlEscape(subtaskBlock)),
		fmt.Sprintf("<dod>%s</dod>", xmlEscape(dod)),
		"\n",
	}, "\n")
}

// buildCompactContext composes a concise context string from prior docs in dir,
// limited to maxRunes characters (runes, not bytes).
func buildCompactContext(dir string, maxRunes int) string {
	if maxRunes <= 0 {
		maxRunes = 4000
	}
	// Order matters: prioritize core derivations
	candidates := []string{
		filepath.Join(dir, "requirements.md"),
		filepath.Join(dir, "srs.md"),
		filepath.Join(dir, "stories.md"),
		filepath.Join(dir, "acceptance_criteria.md"),
		filepath.Join(dir, "architecture.md"),
		filepath.Join(dir, "uml.md"),
	}
	var b strings.Builder
	for _, p := range candidates {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		s := strings.TrimSpace(string(data))
		if s == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString(s)
		// Early stop if we've exceeded the limit by a margin
		if len([]rune(b.String())) >= maxRunes {
			break
		}
	}
	runes := []rune(b.String())
	if len(runes) > maxRunes {
		runes = runes[:maxRunes]
	}
	return string(runes)
}

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func mdEscape(s string) string { return s }

const drayRunTaskList = `
Task List
    - [ ] TASK: Project Scaffold / Bootstrap
      - [ ] Initialize monorepo structure (services, sdk, wrapper, infra)
      - [ ] Setup CI/CD, code quality, branch protection
      - [ ] IaC baseline (TH region), environments (dev/stage/prod)
    - [ ] TASK: Domain Model and Admin CRUD (Purposes/Categories/Vendors/Policies)
      - [ ] Define schemas and migrations
      - [ ] Admin CRUD with versioning and audit
      - [ ] Locale (TH/EN) fields and validations
    - [ ] TASK: Consent Service APIs (Write/Read, Idempotency)
      - [ ] POST /consents, GET /consents
      - [ ] Idempotency keys, receipts
      - [ ] Evidence capture fields
    - [ ] TASK: Consent Lifecycle (Expiry/Refresh, Re-Consent Triggers)
      - [ ] Policy change triggers not-set where required
      - [ ] Expiry policies per purpose/region
      - [ ] Banner prompt logic surfacing via SDK
    - [ ] TASK: Identity Merge & Unlink
      - [ ] Merge by most-recent per purpose/vendor
      - [ ] Lineage persistence and audit entries
      - [ ] Unlink handling (logout)
    - [ ] TASK: Rules/Targeting Engine (Region/Channel/Property/Audience)
      - [ ] Default rules (TH) and overrides
      - [ ] Audience targeting hooks
      - [ ] SDK evaluation contract
    - [ ] TASK: Web SDK (JS)
      - [ ] Initialization, cache, get/set consent
      - [ ] Offline queue, request signing
      - [ ] Preference center UI hooks
    - [ ] TASK: Web Wrapper Governance
      - [ ] Block-by-default script/cookie gating
      - [ ] data-cms-purpose/vendor attributes
      - [ ] Cookie shim and post-consent injection
    - [ ] TASK: Discovery Logging (Scripts/Cookies)
      - [ ] Discovery events endpoint
      - [ ] Admin review/mapping UI
      - [ ] PII sanitization
    - [ ] TASK: Mobile SDKs (iOS/Android)
      - [ ] Init, cache, offline queue, retry
      - [ ] Signing, local storage governance
      - [ ] Preference UI hooks
    - [ ] TASK: Evaluate Consent API + Caching (p95 ≤ 120 ms)
      - [ ] Read path optimization and cache policy
      - [ ] Scope-based evaluation (purpose/vendor)
      - [ ] SLA monitoring
    - [ ] TASK: Propagation (Webhooks/Event Bus)
      - [ ] Signed webhooks, retries, idempotency
      - [ ] Event schemas v1, bus topics
      - [ ] Subscriber guide
    - [ ] TASK: Data Lake Export (Parquet + Partitioning)
      - [ ] ConsentChange and BannerInteraction schemas
      - [ ] Partitions dt/region/channel/property
      - [ ] Schema evolution policy
    - [ ] TASK: Dashboards & Analytics
      - [ ] Funnel, opt-in/out, time-to-consent
      - [ ] Breakdowns (region/category/vendor/channel/property/campaign/variant)
      - [ ] Filters and refresh
    - [ ] TASK: Compliance Reports & Exports
      - [ ] Per-user/purpose audit trails
      - [ ] PDF/CSV exports with branding
      - [ ] Legal hash/policy version visibility
    - [ ] TASK: Admin Console SSO/RBAC/MFA
      - [ ] SAML/OIDC integration
      - [ ] Role mapping and guards
      - [ ] Audit for admin actions
    - [ ] TASK: Observability (Metrics/Tracing/Alerts)
      - [ ] Service metrics, wrapper KPIs
      - [ ] OpenTelemetry tracing
      - [ ] SLO alerts
    - [ ] TASK: Security Hardening (TLS, Signing, KMS, Secrets)
      - [ ] TLS 1.2+, mTLS internal
      - [ ] HMAC signing, key rotation
      - [ ] KMS at rest, least-privilege IAM
    - [ ] TASK: Resilience & DR (Health, Rate Limits, CB, Backups)
      - [ ] Health checks, rate limiting, circuit breakers
      - [ ] Backups, RPO/RTO drills
      - [ ] Restore playbooks
    - [ ] TASK: Cutover & QA (Feature Flags, WCAG, Localization)
      - [ ] Feature flags per property, schedule
      - [ ] WCAG AA audits; TH/EN QA
      - [ ] Runbooks and rollback
    
    tasks/TASK-001.md
    <task>
    Project Scaffold / Bootstrap for the bank-owned CMS. Establish monorepo, services, SDKs, wrapper, infra scaffolding, CI/CD, environments (dev/stage/prod) in Thailand region, and baseline security/quality gates.
    </task>
    <context>
    Target: TH region deployment; services: consent, policy, identity-merge, admin, webhook, etl; sdk: web/ios/android; wrapper web. Use bank-approved CDN and IdP. Git with branch protection. CI/CD to k8s. Secrets via KMS. Default language TH/EN, bank design system. No OneTrust import. WORM storage needed.
    </context>
    <implement>
    - Create monorepo (e.g., Nx/Turbo) with workspaces for services, sdk, wrapper.
    - Initialize API gateway config, k8s manifests, Terraform for TH region.
    - Setup CI: lint/test/build, container scan, IaC scan, SAST; CD: blue/green.
    - Configure environments, KMS keys, object lock bucket for audit.
    - Add CODEOWNERS, PR templates, conventional commits, versioning.
    </implement>
    <subtask>
    - [ ] Monorepo created with packages and service templates
    - [ ] Terraform baseline for TH region and buckets (audit/data)
    - [ ] K8s cluster and namespaces per env
    - [ ] CI pipelines (lint/test/build) and CD (deploy)
    - [ ] Security scanners and branch protections
    - [ ] Secrets/KMS setup and docs
    </subtask>
    <dod>
    - All pipelines green; sample deploy to dev succeeds.
    - Infra applied in TH; audit bucket with object lock verified.
    - README with repo structure and contribution guide.
    </dod>
    <risks>
    Tooling sprawl; IaC drift; permission constraints in bank cloud.
    </risks>
    
    tasks/TASK-002.md
    <task>
    Domain Model and Admin CRUD for Purposes, Categories, Vendors, and Policies with versioning, localization, and audit logging.
    </task>
    <context>
    Must support categories (SN, Functional, Analytics, Advertising), many-to-many vendor↔purpose, TH/EN fields, policy versioning with legal text hashes. Admin actions audited to WORM. No IAB TCF. Data residency TH.
    </context>
    <implement>
    - Define DB schemas and migrations (purpose, vendor, category, policy, mappings).
    - Implement Admin API + UI for CRUD with TH/EN fields and validation.
    - Add versioning for policies; compute and store legal text checksum.
    - Integrate AdminAudit log writes to WORM after changes.
    </implement>
    <subtask>
    - [ ] DB schemas/migrations ready
    - [ ] Admin API endpoints secured
    - [ ] Admin UI for CRUD with TH/EN
    - [ ] Policy versioning + checksum
    - [ ] Admin audit WORM writes
    </subtask>
    <dod>
    - Create/edit/delete visible in UI; changes are versioned and audit-logged.
    - Locale names/descriptions render correctly.
    </dod>
    
    tasks/TASK-003.md
    <task>
    Consent Service APIs for write/read with idempotency and evidence capture.
    </task>
    <context>
    Consent states: granted/denied/not-set. Evidence: ts, channel, property, locale, region, hashed IP, UA, action, deviceId/customerId, campaign, abVariant, policyVersion/legalHash. Idempotency keys required on writes.
    </context>
    <implement>
    - Define POST /v1/consents and GET /v1/consents (query by identifiers/time).
    - Implement idempotency store (key→receipt) with TTL.
    - Compute legalHash from active policy; include in stored record.
    - Hash IP at edge; normalize UA.
    - Return ConsentReceipt with record id and hash.
    </implement>
    <subtask>
    - [ ] API contracts (OpenAPI)
    - [ ] Write path with idempotency
    - [ ] Read path with filters/pagination
    - [ ] Evidence capture implemented
    - [ ] Unit and contract tests
    </subtask>
    <dod>
    - Duplicate idempotencyKey returns same receipt.
    - Evidence fields present; tests pass.
    </dod>
    
    tasks/TASK-004.md
    <task>
    Consent Lifecycle: expiry/refresh policies and re-consent triggers on policy changes.
    </task>
    <context>
    Per-purpose/region expiry (e.g., 180d TH Analytics). Policy activation should mark affected scopes as not-set until user acts. SDK must surface prompts; evaluation respects not-set.
    </context>
    <implement>
    - Model expiry policies; jobs to evaluate and transition to not-set.
    - On policy version change, compute affected scopes and flag re-consent.
    - SDK contract: expose needReconsent() and prompt flows.
    </implement>
    <subtask>
    - [ ] Expiry policy schema + scheduler
    - [ ] Re-consent flagging on policy change
    - [ ] SDK prompt integration
    - [ ] Tests for expiry and policy change
    </subtask>
    <dod>
    - Expired consents become not-set; re-consent is prompted; evaluation reflects states.
    </dod>
    
    tasks/TASK-005.md
    <task>
    Identity Merge & Unlink: merge deviceId and customerId on login; maintain lineage; handle logout.
    </task>
    <context>
    On authentication, unify device/customer consents; conflict policy: most recent per purpose/vendor wins. Keep lineage of source records; support queries by either id. Logout stops using customerId without losing history.
    </context>
    <implement>
    - Implement merge service to fetch both records, resolve winners, persist merged state.
    - Write lineage entries and audit to WORM.
    - Expose unlink endpoint/event handling to revert to device-only scope.
    </implement>
    <subtask>
    - [ ] Merge algorithm + API
    - [ ] Lineage persistence
    - [ ] Unlink/logout handling
    - [ ] Tests for merge/unlink paths
    </subtask>
    <dod>
    - Post-login read by either id returns merged state; lineage visible in audit.
    </dod>
    
    tasks/TASK-006.md
    <task>
    Rules/Targeting Engine for region/channel/property/audience defaults and banner variants.
    </task>
    <context>
    Thailand baseline; property/channel overrides; audience targeting (geo/device/customer segments). No pre-ticked consent. Offline/mobile graceful degradation.
    </context>
    <implement>
    - Define rule DSL/config; evaluation library shared across services/SDKs.
    - Store banner variants and default choices mapping.
    - Provide evaluation API and SDK helper to fetch effective rules.
    </implement>
    <subtask>
    - [ ] Rule schema + evaluator
    - [ ] Variant/targeting storage
    - [ ] SDK helper integration
    - [ ] Tests for rule precedence
    </subtask>
    <dod>
    - Correct defaults/variants applied per region/property/channel/audience in tests.
    </dod>
    
    tasks/TASK-007.md
    <task>
    Web SDK (JS): init, getConsent, setConsent, cache, offline queue, preference UI hooks, signed requests.
    </task>
    <context>
    Must be lightweight, cache-aware, offline tolerant; capture campaign/variant; expose openPreferences(); integrate with wrapper; support request signing and idempotency.
    </context>
    <implement>
    - Build ESM/UMD bundle with init({property,channel,region?,locale?}).
    - Implement getConsent(scope) with cache TTL; setConsent(choices,opts) with queue.
    - Capture UTM/campaign metadata; emit events.
    - Add HMAC signing and idempotency key generation.
    </implement>
    <subtask>
    - [ ] SDK core init/cache
    - [ ] get/set consent APIs
    - [ ] Offline queue + retry
    - [ ] Signing + idempotency
    - [ ] Preference hooks
    </subtask>
    <dod>
    - Works offline with queued writes; cache returns quickly; signed requests verified.
    </dod>
    
    tasks/TASK-008.md
    <task>
    Web Wrapper Governance: block-by-default scripts/cookies; allow post-consent; support data-cms-purpose/vendor attributes.
    </task>
    <context>
    First-party wrapper must gate execution and cookie access per consent; prevent document.cookie set/read for non-allowed purposes; inject allowed scripts after consent.
    </context>
    <implement>
    - Implement script loader parsing DOM for tagged scripts; delay execution until allowed.
    - Provide cookie shim intercepting document.cookie.
    - Listen to SDK consent changes to re-evaluate and inject.
    </implement>
    <subtask>
    - [ ] Script gating engine
    - [ ] Cookie shim
    - [ ] Consent change listener
    - [ ] E2E governance tests
    </subtask>
    <dod>
    - Non-allowed scripts/cookies blocked pre-consent; allowed execute after consent.
    </dod>
    
    tasks/TASK-009.md
    <task>
    Discovery Logging for unknown scripts/cookies with admin mapping workflow.
    </task>
    <context>
    Discovery mode logs unregistered scripts/cookies (URL/name, context), redacts PII, sends to /v1/discovery; admin UI proposes inventory entries for mapping to vendor/purpose.
    </context>
    <implement>
    - Wrapper hooks to emit discovery events with sanitized context.
    - Backend endpoint to collect/store findings.
    - Admin list/review/map and create vendor/purpose entries.
    </implement>
    <subtask>
    - [ ] Wrapper discovery hooks
    - [ ] API + storage
    - [ ] Admin review UI
    - [ ] Redaction tests
    </subtask>
    <dod>
    - Discoveries appear in Admin; mapping persists; PII not leaked.
    </dod>
    
    tasks/TASK-010.md
    <task>
    Mobile SDKs (iOS/Android): init, cache, offline queue, retry, signing, preference UI hooks.
    </task>
    <context>
    Apps must function offline; queue writes; signed requests; local cache TTL; integrate with native preferences screens; minimal footprint; TH/EN localization.
    </context>
    <implement>
    - Build shared spec; implement native SDKs with identical APIs.
    - Add persistent queue with backoff/jitter; reachability listeners.
    - Provide sample UI modules for preferences.
    </implement>
    <subtask>
    - [ ] iOS SDK (Swift Package)
    - [ ] Android SDK (Maven/AAR)
    - [ ] Offline queue + retry
    - [ ] Signing + cache
    - [ ] Samples + docs
    </subtask>
    <dod>
    - Offline operations queue reliably; parity across iOS/Android; sample apps demo flows.
    </dod>
    
    tasks/TASK-011.md
    <task>
    Evaluate Consent API and Caching to meet p95 ≤ 120 ms.
    </task>
    <context>
    Read path must be low latency; cache hot paths; evaluate per purpose/vendor; include policyVersion; observability for SLA.
    </context>
    <implement>
    - Add in-memory/redis cache keyed by identifiers + scope.
    - Optimize DB indexes; avoid N+1; batch lookups.
    - Expose POST /consents/evaluate; implement fast-path return with cache.
    </implement>
    <subtask>
    - [ ] Cache strategy + invalidation
    - [ ] Optimized queries/indexes
    - [ ] Evaluate endpoint
    - [ ] Latency dashboards + SLO
    </subtask>
    <dod>
    - Load tests show p95 ≤ 120 ms; cache hit ratio targets met.
    </dod>
    
    tasks/TASK-012.md
    <task>
    Propagation via Webhooks and Event Bus with signing and idempotency.
    </task>
    <context>
    Near real-time updates: median ≤ 5s, p95 ≤ 30s. Signed webhook delivery with retries; bus topics cms.consent.events; idempotency keys prevent duplicates.
    </context>
    <implement>
    - Build webhook dispatcher with retry/backoff and HMAC signatures.
    - Publish to event bus with schema v1; include idempotencyKey and timestamps.
    - Subscriber guide and sample consumers.
    </implement>
    <subtask>
    - [ ] Webhook delivery engine
    - [ ] Event bus producer + schemas
    - [ ] Subscriber docs/samples
    - [ ] Lag monitoring
    </subtask>
    <dod>
    - Delivery SLOs met; duplicates deduped by subscribers using key.
    </dod>
    
    tasks/TASK-013.md
    <task>
    Data Lake Export (Parquet) for consent changes and banner interactions with partitioning.
    </task>
    <context>
    Export to TH-region data lake in Parquet; partitions dt=YYYY-MM-DD/region/channel/property; pseudonymize identifiers; schema evolution documented.
    </context>
    <implement>
    - Define Parquet schemas; implement ETL sink with rolling files.
    - Partition writer; handle late events and compaction.
    - Document schema versions and changes.
    </implement>
    <subtask>
    - [ ] Schemas defined
    - [ ] ETL writer + partitions
    - [ ] Pseudonymization
    - [ ] Docs and sample queries
    </subtask>
    <dod>
    - Files land in correct partitions; sample queries succeed; schemas published.
    </dod>
    
    tasks/TASK-014.md
    <task>
    Dashboards & Analytics: funnel, opt-in/out, time-to-consent, breakdowns with filters.
    </task>
    <context>
    Use data lake events to power dashboards; dimensions: region/category/vendor/channel/property/campaign/variant; update within 5s on filter changes.
    </context>
    <implement>
    - Build semantic model; create dashboards (e.g., Superset/Looker).
    - Implement funnel stages and breakdown slices.
    - Add date filters and caching for responsiveness.
    </implement>
    <subtask>
    - [ ] Semantic model
    - [ ] Core dashboards
    - [ ] Filters + caching
    - [ ] KPI definitions documented
    </subtask>
    <dod>
    - Dashboards render metrics correctly and update within 5s when filtering.
    </dod>
    
    tasks/TASK-015.md
    <task>
    Compliance Reports & Exports (PDF/CSV) regulator-ready with audit trails.
    </task>
    <context>
    Per-user/purpose audit with evidence, policy versions, legal hashes. Branded exports; PDPA-compliant formatting; secure access.
    </context>
    <implement>
    - Build reporting queries and API.
    - Generate PDF/CSV with bank branding.
    - Access control checks; watermarking and audit of downloads.
    </implement>
    <subtask>
    - [ ] Queries/APIs
    - [ ] PDF/CSV generators
    - [ ] Access control + audit
    - [ ] Templates and tests
    </subtask>
    <dod>
    - Reports contain required fields; exports download successfully; access enforced.
    </dod>
    
    tasks/TASK-016.md
    <task>
    Admin Console SSO/RBAC/MFA integration and permission guards.
    </task>
    <context>
    SAML/OIDC with Bank IdP. Map groups/claims to roles: Admin, Compliance, Analyst, Developer. Enforce UI/API guards; MFA per policy.
    </context>
    <implement>
    - Implement SSO login flow; configure IdP.
    - Role mapping middleware; protect routes and APIs.
    - Enforce MFA and session policies; log admin actions to WORM.
    </implement>
    <subtask>
    - [ ] SSO integration
    - [ ] Role mapping and guards
    - [ ] MFA enforcement
    - [ ] Admin action audit
    </subtask>
    <dod>
    - Role-based access verified; MFA required; admin actions audited.
    </dod>
    
    tasks/TASK-017.md
    <task>
    Observability: metrics, tracing, logs, and SLO alerts across services and wrapper.
    </task>
    <context>
    Track latency, error rates, propagation lag, wrapper block/allow counts, funnel KPIs. Use OpenTelemetry and Prometheus/OpenMetrics; correlate with IDs.
    </context>
    <implement>
    - Instrument services and SDK/wrapper events.
    - Centralize logs with correlation IDs.
    - Define SLOs and configure alert routes/on-call.
    </implement>
    <subtask>
    - [ ] Metrics + dashboards
    - [ ] Tracing spans
    - [ ] Structured logging
    - [ ] Alerts configured
    </subtask>
    <dod>
    - Dashboards and alerts operational; traces link requests across services.
    </dod>
    
    tasks/TASK-018.md
    <task>
    Security Hardening: TLS, HMAC signing, KMS encryption, secrets rotation, IAM least privilege.
    </task>
    <context>
    Data residency in TH; TLS 1.2+; mTLS internal; HMAC request/webhook signatures; KMS at rest; secret rotation with no downtime; RBAC for services.
    </context>
    <implement>
    - Enforce TLS and mTLS; configure cert management.
    - Implement request/webhook signing and verification.
    - Encrypt at rest; define IAM policies; setup rotation schedules.
    </implement>
    <subtask>
    - [ ] TLS/mTLS configured
    - [ ] HMAC signing verified
    - [ ] KMS policies applied
    - [ ] Secrets rotation runbook
    </subtask>
    <dod>
    - Pen test/security checks pass; rotations tested without outage.
    </dod>
    
    tasks/TASK-019.md
    <task>
    Resilience & DR: health checks, rate limiting, circuit breakers, backups, RPO/RTO drills and restore playbooks.
    </task>
    <context>
    Availability 99.9%; RPO ≤ 15 min; RTO ≤ 2 h. Idempotent writes; graceful degradation with 503 on breaker; backups scheduled and tested.
    </context>
    <implement>
    - Add readiness/liveness and dependency health endpoints.
    - Implement rate limiting and circuit breakers.
    - Configure backups and restore runbooks; perform DR exercise.
    </implement>
    <subtask>
    - [ ] Health endpoints
    - [ ] Rate limiters + CBs
    - [ ] Backup schedules
    - [ ] DR drill + restore test
    </subtask>
    <dod>
    - DR simulation meets RPO/RTO; graceful failure observed under fault injection.
    </dod>
    
    tasks/TASK-020.md
    <task>
    Cutover & QA: feature flags per property, scheduled big-bang cutover, WCAG AA and localization QA, runbooks and rollback.
    </task>
    <context>
    No data backfill; feature-flag control; TH/EN localization; WCAG AA for banner and admin; cutover date with monitoring and rollback path.
    </context>
    <implement>
    - Implement per-property flags in wrapper/CDN and APIs.
    - Prepare runbooks; dry-run shadow deploy; finalize rollback plan.
    - Conduct WCAG audits and localization QA; fix defects.
    </implement>
    <subtask>
    - [ ] Flags implemented and tested
    - [ ] Shadow run + monitoring
    - [ ] WCAG AA audit fixes
    - [ ] TH/EN localization QA
    - [ ] Cutover + rollback runbooks
    </subtask>
    <dod>
    - Flags enable/disable properties atomically; WCAG checks pass; cutover executed with rollback readiness.
    </dod>
    <notes>
    Coordinate with marketing/analytics for discovery inventory iteration post-cutover.
    </notes>
`
