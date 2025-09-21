# TASK-002 — LangGraph REST client (run/questions) + retry/backoff
**Date:** 2025-09-15

## ContextEngineering
```xml
<task>LangGraph REST client (run/questions) + retry/backoff</task>
<context>
  <requirement># AgentFlow — Requirements (v2)

**Date:** 2025-09-15

## 1. บทนำ
AgentFlow คือ CLI ที่ช่วยแปลงไอเดียจากมนุษย์ให้เป็นเอกสารซอฟต์แวร์ครบวงจรและงานสำหรับนักพัฒนา โดย orchestrate AI Agents ผ่าน LangGraph Server (REST) และเก็บหลักฐานผลลัพธ์อย่างเป็นระบบ

## 2. เป้าหมาย
- จัดโฟลว์จาก idea → requirements → srs → stories → acceptance criteria → architecture/UML → test plan → dev tasks
- รองรับ role-based prompting (PO/PM, SA, QA, Dev)
- เก็บ idea แบบมี **timeline** (ไฟล์รายวันใน `input/`), และให้ prompt เล่าการเปลี่ยนแปลงตามลำดับเวลา

## 3. ขอบเขต
- CLI: `agentflow init | intake | plan | design | qa | devplan`
- แหล่งข้อมูล: `.agentflow/input/YYYY-MM-DD.md`
- เอาต์พุต: `.agentflow/output/*.md` และ `tasks/*.md`

## 4. Functional Requirements
- FR-1: `init` สร้าง `.agentflow/config.json` ด้วยสคีมาเวอร์ชัน, model, roles, io, security, devplan ฯลฯ
- FR-2: `intake` รวม idea ตาม **timeline** → สร้าง prompt สำหรับ PO/PM → gen `requirements.md` (มี Goals/Scope/FR/NFR/Assumptions/Open Questions & Timeline Summary & Questions to Human)
- FR-3: `plan` วิเคราะห์ `requirements.md` → gen `srs.md`, `stories.md`, `acceptance_criteria.md` โดย SA (และสร้างคำถามถ้าข้อมูลไม่พอ)
- FR-4: `design` วิเคราะห์ทั้งหมด → gen `architecture.md` (มี **Project Structure**), `uml.md` (PlantUML)
- FR-5: `qa` สร้าง `test-plan.md`
- FR-6: `devplan` gen `task_list.md` และ `tasks/<task_id>.md` โดยแต่ละไฟล์มี XML sections:  
  `<task/> <context/> <implement/> <subtask/> <dod/>` (และ extras เช่น `<risk/> <notes/> <trace/>`)
- FR-7: รองรับ human-in-the-loop ผ่านคำถามย้อนกลับของ agent

## 5. Non-Functional Requirements
- Go 1.22+, single static binary, cross-platform
- Config JSON + schema versioning
- Security: อ่าน API keys จาก env, บันทึก/เรดแด็คต์ใน log
- Test coverage ≥ 70%, Snapshot tests สำหรับเอกสาร

## 6. ความเสี่ยง
- คุณภาพผลลัพธ์ขึ้นกับโมเดลและ prompt ⇒ ใช้ templates + snapshots
- ข้อจำกัด context ⇒ สรุปย่อ per-task
</requirement>
  <srs># AgentFlow — SRS (v2)

**Date:** 2025-09-15

## 1. คำนิยาม/Actors
- **Developer**: ผู้สั่ง CLI
- **PO/PM**: แนวคิด/ลำดับเวลา/ความต้องการระดับธุรกิจ
- **SA**: วิเคราะห์ระบบ, SRS, Stories/AC, Design/UML
- **QA**: แผนทดสอบ
- **LangGraph Server**: orchestration ของ agents, REST
- **LLM Provider**: โมเดลภายนอก

## 2. Use Cases
- UC-01 Init, UC-02 Intake (Timeline-aware), UC-03 Plan, UC-04 Design, UC-05 QA, UC-06 DevPlan

## 3. System Features
### 3.1 Config (Init)
ตัวอย่างสคีมา (ย่อ):


### 3.2 Intake (Timeline Prompting)
- รวมไฟล์ใน `input/*.md` ตามวันที่ (ชื่อไฟล์ YYYY-MM-DD.md) → สร้างส่วน **Timeline Narrative** แสดงการเปลี่ยนแปลงไอเดีย
- โครงสร้าง Requirements Output ต้องมีหัวข้อ: Goals, Scope, FR/NFR, Assumptions, Constraints, Timeline Summary, Questions to Human

### 3.3 Plan
- จาก `requirements.md` → สร้าง `srs.md` (Use Cases, Interfaces, API, Data, Constraints)
- `stories.md` (INVEST) + `acceptance_criteria.md` (Gherkin)

### 3.4 Design
- `architecture.md` (ภาพรวมสถาปัตยกรรม + **Project Structure**) + `uml.md` (PlantUML)

### 3.5 QA
- `test-plan.md`

### 3.6 DevPlan
- `task_list.md` + `tasks/<id>.md` (XML sections: `<task/> <context/> <implement/> <subtask/> <dod/>` + extras)

## 4. External Interfaces
- CLI flags: `--config`, `--noninteractive`, `--overwrite`, `--append`
- REST: `POST /agents/run`, `POST /agents/questions`

## 5. Constraints
- PlantUML syntax ใน codefence ```plantuml
- Docker สำหรับ LangGraph

## 6. Quality Requirements
- Testability, Traceability (story ↔ AC ↔ task), Observability (logs/metrics)
</srs>
  <stories># AgentFlow — User Stories (v2)

**Date:** 2025-09-15

## EPIC-INTAKE
- STORY-INT-1: As a PO/PM, I can generate requirements with a **timeline summary** of ideas.

## EPIC-PLAN
- STORY-PLAN-1: As a SA, I can generate SRS from requirements.
- STORY-PLAN-2: As a SA, I can produce INVEST stories and mapped AC in Gherkin.

## EPIC-DESIGN
- STORY-DES-1: As a SA, I can generate architecture.md containing **Project Structure** and PlantUML diagrams.
- STORY-DES-2: As a SA, I can generate uml.md (Sequence/Class/Activity/Component).

## EPIC-QA
- STORY-QA-1: As a QA, I can craft a detailed test-plan.md.

## EPIC-DEVPLAN
- STORY-DEV-1: As a Dev, I can get a task_list.md with checkbox items and per-task context-engineered files.
- STORY-DEV-2: As a Dev, the **first task is Project Scaffold** if not present.
</stories>
  <acceptanceCriteria># AgentFlow — Acceptance Criteria (v2)

**Date:** 2025-09-15

## Intake
- [ ] `requirements.md` มี Timeline Summary + Questions to Human

## Plan
- [ ] `srs.md` ครบ Use Cases/Interfaces/Constraints/Data
- [ ] `stories.md` เป็น INVEST และจับคู่กับ AC
- [ ] `acceptance_criteria.md` เป็น Gherkin ต่อ story

## Design
- [ ] `architecture.md` มี **Project Structure** และ PlantUML Component/Deployment
- [ ] `uml.md` มี Sequence/Class/Activity อย่างน้อย

## QA
- [ ] `test-plan.md` ครบ Strategy/Types/Entry-Exit/Risks

## DevPlan
- [ ] `task_list.md` สอดคล้องกับ stories/AC
- [ ] ไฟล์ `tasks/<id>.md` มี XML sections: `<task/> <context/> <implement/> <subtask/> <dod/>`
- [ ] งานแรกคือ **Project Scaffold**
</acceptanceCriteria>
  <architecture># AgentFlow — Architecture (v2)

**Date:** 2025-09-15

## 1. Overview
CLI (Go) ↔ LangGraph (Docker, REST) ↔ LLM Provider. เอกสารถูกเก็บใน `.agentflow/output`

## 2. Component Diagram


## 3. Deployment Diagram


## 4. Project Structure (Go + LangGraph)


## 5. REST Contracts
- `POST /agents/run` → `{role, prompt, params}` → `{runId, content}`
- `POST /agents/questions` → human-in-the-loop

## 6. Data/Metadata
- เขียน YAML frontmatter (runId, stage, model, time) บนเอกสารทุกไฟล์เพื่อ traceability
</architecture>
  <uml># AgentFlow — UML (v2)

**Date:** 2025-09-15

## Sequence: Timeline Intake


## Class: Core


## Activity: DevPlan

</uml>
</context>
<implement>
HTTP client with context/timeout, retry (exponential backoff), healthcheck function, structured logging.
</implement>
<subtask>
- [x] LangGraph REST client (run/questions) + retry/backoff: design
- [x] LangGraph REST client (run/questions) + retry/backoff: implement
- [x] LangGraph REST client (run/questions) + retry/backoff: unit/integration tests
- [x] LangGraph REST client (run/questions) + retry/backoff: docs & snapshots
</subtask>
<dod>
- Build runs on macOS/Linux/Windows
- Tests pass with coverage ≥ 70% (where applicable)
- Artifacts written to .agentflow/output (where applicable)
- No secrets in logs, configs validated
</dod>
<risk>
Model drift; network failures; context size — mitigated by templates, retries, summarization.
</risk>
<trace>
Story: See stories.md (v2) for relevant EPIC/STORY as listed above.
AC: See acceptance_criteria.md (v2) related section.
</trace>
<notes>Keep commits small; add golden snapshots for deterministic assertions.</notes>
```
