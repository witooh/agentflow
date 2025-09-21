# AgentFlow — Requirements (v2)

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
