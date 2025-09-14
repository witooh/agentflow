# Architecture: Agentflow CLI (Spec-Driven, Hybrid-ready)

<problem>
- ต้องการสถาปัตยกรรมที่เปลี่ยนไอเดีย → สเปกที่อ่านง่ายและเครื่องอ่านได้ โดยเก็บเฉพาะผลลัพธ์สำคัญ
- รองรับหลายบทบาทเอเจนต์ โดยลดโอกาสชนกันของเอกสาร
</problem>

<approach>
- **Spec-Driven Development**: ใช้ SPEC.md เป็นแหล่งความจริงเดียว
- **Minimal State**: ไม่เก็บ run history; เก็บเฉพาะ JSON และ Markdown+XML blocks
- **Go CLI** สำหรับ UX/Config/Validation + (เลือกได้) **Hybrid Runtime** (LangGraph) ผ่าน HTTP/gRPC เมื่อโปรเจกต์ใหญ่
</approach>

<architecture>
- Context:
  - ผู้ใช้สั่งงานผ่าน `agentflow` (Go CLI)
  - CLI อ่าน/เขียนไฟล์ใน `.agentflow/` และเรียกเอเจนต์เป็นรายบทบาท
  - (ทางเลือก) เรียก Runtime ภายนอกเพื่อทำ retrieval/formatting/validation ขั้นสูง
- Boundaries:
  - เอเจนต์แต่ละบทบาทแก้เฉพาะบล็อก/ไฟล์ในความรับผิดชอบ
  - จำกัดสิทธิ์การเขียนไฟล์ไว้เฉพาะ `.agentflow/` (และโฟลเดอร์อนุญาต)
</architecture>

<components>
- **Agentflow CLI (Go + Cobra/Viper)**: คำสั่ง `init/intake/plan/design/qa/devlead/devops/validate/export`
- **Agents (Roles)**: intake, planner, architect, qa, devlead, devops (เติมสเปกตามบล็อก)
- **(Optional) Agent Runtime (LangGraph)**: จัดการ retrieval-first, summarize, schema-validate เมื่อจำเป็น
</components>

<dataflow>
1) Human เริ่มไอเดีย → `agentflow init` สร้าง SPEC.md และโครง `.agentflow/`
2) เรียกเอเจนต์ตามบทบาท → เติมบล็อกใน SPEC.md และสร้างไฟล์เสริม (`tasks.json`, `openapi.json`, `datamodel.json`)
3) `agentflow validate` ตรวจรูปแบบแท็ก + JSON schema
4) `agentflow export` รวมเอกสารพร้อมส่งให้ทีมพัฒนา/AI coding agents
</dataflow>

<folder_structure>
กติกาไฟล์ (สำคัญ)
รูปแบบที่อนุญาตเท่านั้น:
JSON — ใช้กับสิ่งที่ต้องแม่นตามสคีมา (เช่น tasks.json, openapi.json, datamodel.json)
Markdown + XML blocks — เอกสารบรรยายที่มีแท็ก <problem>, <approach>, <architecture>, <acceptance>, …
ห้ามมี runs/ หรือประวัติการรันย่อย — เก็บเฉพาะ “สิ่งที่พร้อมให้ dev/AI coding agents ใช้พัฒนา”
ขอบเขตการเขียนของเอเจนต์: เฉพาะใน .agentflow/ (และโฟลเดอร์อนุญาตอื่น ๆ ตามนโยบายโปรเจกต์)
โค้ดจริง อยู่ นอก .agentflow/ (เช่นใน src/)
ใช้งานจริง: ให้ทุกบทบาทเติมเฉพาะบล็อก/ไฟล์ของตัวเองในชุดไฟล์ด้านบน แล้วให้ dev/AI tools (เช่น Cursor) อ่าน SPEC.md + JSON เสริม เพื่อ implement ได้ตรงตาม requirement.
.agentflow/
  SPEC.md                # สเปกศูนย์กลาง (Markdown + XML blocks)
  requirements.md        # สรุปปัญหา/ขอบเขต/ความต้องการ (MD+XML)
  architecture.md        # ภาพรวมสถาปัตย์/การตัดสินใจ (MD+XML)
    
  tasks.json             # รายการงาน + DoD/AC (JSON)        [optional]
  openapi.json           # สคีมา API (OpenAPI) (JSON)        [optional]
  datamodel.json         # ดาต้าโมเดล/สคีมา (JSON)          [optional]
    
  config/
    agents.yaml          # แมปบทบาทเอเจนต์ ↔ บล็อกใน SPEC.md/ไฟล์ที่ดูแล
    # (ไฟล์ตั้งค่าอื่น ๆ ที่จำเป็นเท่านั้น เช่น โมเดล/พารามิเตอร์)
    
  agents/                # (แยกเอาต์พุตต่อบทบาท—ถ้าต้องการ)
    intake/
      outputs/
        requirements.md  # ผลลัพธ์ของ Intake (MD+XML)
      state.json         # สถานะย่อยล่าสุดของบทบาท (JSON)  [optional]
    planner/
      outputs/
        plan.md          # สรุปแผน (MD+XML)                  [optional]
        tasks.json       # รายการงาน (JSON)
    architect/
      outputs/
        adr-001.md       # ADR/Design Notes (MD+XML)
        design-overview.md
    qa/
      outputs/
        test-plan.md     # แผนทดสอบ + AC (MD+XML)
    devlead/
      outputs/
        dev-breakdown.md # แผนลงมือ/ลำดับงาน (MD+XML)      [optional]
        dev-breakdown.json                                   [optional]
    devops/
      outputs/
        pipeline-plan.json                                   [optional]
        release-strategy.md                                  [optional]
</folder_structure>

<apis>
- ภายใน CLI:
  - `agentflow init` สร้างโครง
  - `agentflow plan` อัปเดต `<userstories>/<acceptance>` + `tasks.json`
  - `agentflow design` เติม `<architecture>/<decisions>` + `architecture.md`
  - `agentflow validate` ตรวจ XML blocks + JSON schema
- ภายนอก (ถ้าใช้ Runtime):
  - `POST /run/role { role, inputs:[fileRefs] }` → คืน `{ artifacts:[{path,kind}] }`
</apis>

<constraints>
- ไม่มีระบบติดตามรอบการรันละเอียด
- ต้องรักษาความสอดคล้องของ SPEC.md กับไฟล์เสริมทุกครั้งก่อน export
- ควรใช้ retrieval-first เพื่อลดคอนเท็กซ์ของโมเดล
</constraints>

<decisions>
- เลือก Markdown+XML blocks เป็น “สัญญาเอกสาร” กลาง
- ใช้ JSON สำหรับส่วนที่ต้องแม่นยำ (tasks/API/datamodel)
- เตรียมรองรับ Hybrid Runtime โดยไม่ต้องเปลี่ยนรูปแบบไฟล์
</decisions>

<risks>
- สเปกไม่ครบถ้วน → โค้ดอาจเบี่ยง
- โปรเจกต์ใหญ่ → การค้นคืนอาจไม่แม่น
</risks>

<mitigations>
- ใช้เทมเพลต SPEC ที่เข้ม: บังคับมี `<acceptance>`/DoD ต่อ task
- ทำ chunking/retrieval-first และ (ถ้าจำเป็น) เพิ่ม index เบาบางต่อไฟล์ Markdown/JSON
</mitigations>

<nonfunctional>
- DX ดี: ใช้ Git ธรรมดา, โครงไฟล์เรียบง่าย
- Portable: ข้ามแพลตฟอร์มได้
- Scalable: เพิ่ม Runtime ภายนอกเมื่อโปรเจกต์โต โดยไม่รื้อสัญญาไฟล์
</nonfunctional>
