# tasks.md — งานแบบ Checklist และ User Stories (สำหรับ agentflow CLI)

เอกสารนี้สรุป “งานที่ต้องทำทีละขั้น” สำหรับสร้างและส่งมอบ agentflow แบบ CLI (Go-first) ที่จัดเก็บข้อมูลทั้งหมดในโฟลเดอร์ ./.agentflow เพื่อให้เวอร์ชันด้วย Git ได้ง่าย เอกสารแบ่งเป็น 2 มุมมอง:
- User Stories + Acceptance Criteria (เพื่อสื่อสารความต้องการ)
- Task Checklist (มีรหัสงาน, บทบาท, ประมาณเวลา, และลำดับพึ่งพา)

หมายเหตุ:
- โฟลเดอร์เก็บข้อมูล: ./.agentflow (ตามโครงที่ระบุใน requirements.md/architecture.md/blueprint.md)
- CLI Framework: Cobra, Config: Viper
- เอาต์พุต CLI ต้องรองรับ human-readable และ --json เมื่อเหมาะสม

---

## 1) User Stories (ไทย) + Acceptance Criteria

US-01 — Initialize project
- As a developer, I want to run `agentflow init` so that the project has a ready-to-use ./.agentflow structure.
- AC:
  - Given empty repo, When run `agentflow init`, Then ./.agentflow ถูกสร้างครบโครง (intake/, analysis/, planning/, runs/, artifacts/, configs/ ฯลฯ) ภายใน < 1s บนเครื่องทั่วไป
  - มีไฟล์ project.json (manifest) ที่เติมค่าเบื้องต้น (name/slug/version/createdAt)
  - ไม่เขียนทับไฟล์ที่มีอยู่แล้วโดยไม่ขออนุญาต (ใช้ --force เพื่อเขียนทับ)

US-02 — Intake requirements
- As a PM/Intake user, I want `agentflow intake --interactive` to collect Q&A and save them.
- AC:
  - When run `agentflow intake --interactive`, Then บันทึก .agentflow/intake/requirements.json และ summary.md
  - รองรับ non-interactive ผ่านไฟล์อินพุต `--in <file>` และพิมพ์ผล `--out <file>` หรือ stdout

US-03 — Generate SRS + Stories
- As a System Analyst, I want `agentflow srs` to transform intake into SRS.md and stories.json.
- AC:
  - Input: .agentflow/intake/requirements.json
  - Outputs: .agentflow/analysis/SRS.md, .agentflow/analysis/stories.json (ตามโครงในเอกสาร)
  - รองรับ `--json` เพื่อแสดงเฉพาะ stories เป็น JSON ที่ stdout

US-04 — Planning tasks
- As a Planner, I want `agentflow plan` to produce a tasks list (JSON + Markdown summary).
- AC:
  - Input: SRS.md และ/หรือ stories.json
  - Outputs: .agentflow/planning/tasks.json และ plan.md
  - ตรวจโครงตาม schema: { tasks:[{id,title,role,estimateHour,dependsOn[]}...] }

US-05 — Run agents and persist artifacts
- As a user, I want `agentflow run` to execute an agent workflow and store run logs/artifacts.
- AC:
  - Creates .agentflow/runs/<ts-id>/{agent.json,logs.ndjson,artifacts/}
  - เคารพ flags `--concurrency`, `--dry-run`, `--json`
  - มี exit codes ชัดเจน (0=สำเร็จ)

US-06 — Config & Observability
- As an operator, I want config via file/ENV and structured logs/verbosity.
- AC:
  - อ่านคอนฟิกจาก $HOME/.config/agentflow/config.yaml + ENV override (Viper)
  - รองรับ -v/--verbose และ --trace
  - ไม่พิมพ์ secrets ลง stdout/log โดยไม่จำเป็น

US-07 — Distribution
- As a releaser, I want reproducible builds and cross-platform distribution.
- AC:
  - Build binary ด้วย GoReleaser หรือ go build (มี --version แสดง commit/date)
  - README วิธีติดตั้ง/ใช้งานเบื้องต้น

---

## 2) Task Checklist (มีรหัสงาน/บทบาท/เวลา/พึ่งพา)

Legend:
- role: pm | sa | architect | techlead | fe | be | qa | devops | writer | cli-dev (Go)
- estimateHour: ชม. โดยประมาณ (1–8 ต่อชิ้นงาน หากใหญ่กว่านั้นควรแตกย่อย)
- dependsOn: รหัสงานที่ต้องเสร็จก่อน

Phase A — Project Skeleton & Conventions
- [x] T-01 Setup Go module และ dependencies (cobra, viper) — role: cli-dev — 2h
- [x] T-02 กำหนดเวอร์ชัน/บิลด์ข้อมูล (--version; ldflags) — role: cli-dev — 1h — dependsOn: T-01
- [x] T-03 จัดทำ root command + --help + --json placeholder — role: cli-dev — 2h — dependsOn: T-01

Phase B — Filesystem IO & .agentflow
- [ ] T-04 ฟังก์ชัน helper จัดการพาธ ./.agentflow และการสร้างโครงโฟลเดอร์ — role: cli-dev — 3h — dependsOn: T-01
- [ ] T-05 สคีมา/โครงสร้างไฟล์พื้นฐาน (project.json, gitignore ข้อยกเว้น) — role: sa | cli-dev — 2h — dependsOn: T-04

Phase C — Commands
- [ ] T-06 คำสั่ง init สร้าง ./.agentflow + project.json — role: cli-dev — 3h — dependsOn: T-04, T-05
- [ ] T-07 คำสั่ง intake (interactive + non-interactive) เขียน requirements.json + summary.md — role: cli-dev — 4h — dependsOn: T-06
- [ ] T-08 คำสั่ง srs แปลงเป็น SRS.md + stories.json — role: cli-dev — 4h — dependsOn: T-07
- [ ] T-09 คำสั่ง plan ผลิต planning/tasks.json + plan.md — role: cli-dev — 4h — dependsOn: T-08
- [ ] T-10 คำสั่ง run จัดการ runs/<ts-id>/ และ artifacts/ — role: cli-dev — 6h — dependsOn: T-09

Phase D — Config, Logging, Security
- [ ] T-11 Viper config ($HOME/.config/agentflow/config.yaml + ENV) — role: cli-dev — 2h — dependsOn: T-03
- [ ] T-12 Structured logging + -v/--verbose/--trace — role: cli-dev — 2h — dependsOn: T-03
- [ ] T-13 Guardrails: ไม่ log secrets, กำหนดสิทธิ์ไฟล์ (0600) — role: cli-dev — 2h — dependsOn: T-11, T-12

Phase E — Docs & Examples
- [ ] T-14 อัปเดต README/USAGE สำหรับคำสั่งทั้งหมด + ตัวอย่าง — role: writer — 3h — dependsOn: T-06..T-10
- [ ] T-15 เพิ่มตัวอย่างไฟล์ภายใน .agentflow (dummy) — role: writer | cli-dev — 2h — dependsOn: T-06..T-09

Phase F — QA & Release
- [ ] T-16 Test Plan (unit/smoke) + เคสสำคัญต่อคำสั่ง — role: qa — 3h — dependsOn: T-06..T-10
- [ ] T-17 เขียน unit tests สำหรับ helpers/commands หลัก — role: cli-dev — 6h — dependsOn: T-16
- [ ] T-18 GoReleaser config (multi-OS/arch) + verify binary flags — role: devops — 4h — dependsOn: T-02, T-10
- [ ] T-19 Release notes v0.1 — role: techlead | writer — 2h — dependsOn: T-18

Notes:
- ทุกคำสั่งควรรองรับ --project-dir เพื่อระบุรากโปรเจกต์ (ค่าเริ่มต้น = cwd)
- tasks.json/plan.md ที่สร้างจากคำสั่ง plan สามารถนำรายการงานนี้ไปใช้เป็น baseline ได้

---

## 3) ตัวอย่าง JSON (สอดคล้องกับ .agentflow/planning/tasks.json)

คัดลอกไปวางที่ .agentflow/planning/tasks.json ได้โดยตรง และปรับแก้ตามโปรเจกต์จริง

```json
{
  "tasks": [
    { "id": "T-01", "title": "Setup Go module and deps (cobra,viper)", "role": "cli-dev", "estimateHour": 2, "dependsOn": [] },
    { "id": "T-02", "title": "Build version/ldflags --version", "role": "cli-dev", "estimateHour": 1, "dependsOn": ["T-01"] },
    { "id": "T-03", "title": "Root command + --help + --json placeholder", "role": "cli-dev", "estimateHour": 2, "dependsOn": ["T-01"] },
    { "id": "T-04", "title": ".agentflow path helpers + folder scaffold", "role": "cli-dev", "estimateHour": 3, "dependsOn": ["T-01"] },
    { "id": "T-05", "title": "Base schemas/files (project.json, gitignore)", "role": "cli-dev", "estimateHour": 2, "dependsOn": ["T-04"] },
    { "id": "T-06", "title": "init command creates .agentflow + project.json", "role": "cli-dev", "estimateHour": 3, "dependsOn": ["T-04","T-05"] },
    { "id": "T-07", "title": "intake command (interactive/non-interactive)", "role": "cli-dev", "estimateHour": 4, "dependsOn": ["T-06"] },
    { "id": "T-08", "title": "srs command (SRS.md + stories.json)", "role": "cli-dev", "estimateHour": 4, "dependsOn": ["T-07"] },
    { "id": "T-09", "title": "plan command (tasks.json + plan.md)", "role": "cli-dev", "estimateHour": 4, "dependsOn": ["T-08"] },
    { "id": "T-10", "title": "run command (runs/<ts-id> + artifacts)", "role": "cli-dev", "estimateHour": 6, "dependsOn": ["T-09"] },
    { "id": "T-11", "title": "Viper config ($HOME/.config/agentflow)", "role": "cli-dev", "estimateHour": 2, "dependsOn": ["T-03"] },
    { "id": "T-12", "title": "Structured logging + verbosity flags", "role": "cli-dev", "estimateHour": 2, "dependsOn": ["T-03"] },
    { "id": "T-13", "title": "Guardrails: avoid logging secrets, file perms", "role": "cli-dev", "estimateHour": 2, "dependsOn": ["T-11","T-12"] },
    { "id": "T-14", "title": "Docs/README usage for all commands", "role": "writer", "estimateHour": 3, "dependsOn": ["T-06","T-07","T-08","T-09","T-10"] },
    { "id": "T-15", "title": "Example files in .agentflow (dummy)", "role": "cli-dev", "estimateHour": 2, "dependsOn": ["T-06","T-07","T-08","T-09"] },
    { "id": "T-16", "title": "QA Test Plan + key cases per command", "role": "qa", "estimateHour": 3, "dependsOn": ["T-06","T-07","T-08","T-09","T-10"] },
    { "id": "T-17", "title": "Unit tests for helpers/commands", "role": "cli-dev", "estimateHour": 6, "dependsOn": ["T-16"] },
    { "id": "T-18", "title": "GoReleaser config + verify binaries", "role": "devops", "estimateHour": 4, "dependsOn": ["T-02","T-10"] },
    { "id": "T-19", "title": "Release notes v0.1", "role": "techlead", "estimateHour": 2, "dependsOn": ["T-18"] }
  ]
}
```

---

## 4) วิธีติดตามความคืบหน้า

- ใช้กล่องเช็คในเอกสารนี้ระหว่างทำงานจริง
- นำรายการ JSON ไปใส่ .agentflow/planning/tasks.json แล้วให้เอเจนต์/สคริปต์อ่านเพื่ออัปเดตรายงานความคืบหน้า
- ทุกผลลัพธ์/หลักฐาน (artifacts/logs) ให้เก็บลง .agentflow ตามมาตรฐาน

อัปเดตล่าสุด: 2025-09-15 03:55
