# Requirements: Agentflow CLI

<problem>
- ต้องการ CLI ที่แปลงไอเดียของมนุษย์ → สเปกชัดเจนเพื่อให้ dev/AI coding agents (เช่น Cursor) พัฒนาได้ตรงความต้องการ
- ลดความซับซ้อน: ไม่เก็บ run history ละเอียด เก็บเฉพาะผลลัพธ์ที่จำเป็นต่อการพัฒนา
- ให้เอเจนต์หลายบทบาททำงานบนเอกสารเดียวกันได้ โดยไม่ชนกัน
</problem>

<scope>
- In-scope:
  - โครงสร้างโปรเจกต์แบบมินิมอลภายใต้ `.agentflow/`
  - สเปกศูนย์กลาง (SPEC-driven) และไฟล์เสริมที่จำเป็นต่อการพัฒนา
  - คำสั่ง CLI เรียกเอเจนต์ตามบทบาทเพื่อเติมสเปก
  - เก็บผลลัพธ์เอเจนต์เป็น 2 ชนิดเท่านั้น: JSON และ Markdown+XML blocks
  - ตรวจความถูกต้องของแท็ก XML และ JSON schema
- Out-of-scope (เฟสแรก):
  - การติดตามรอบการรัน/เวิร์กโฟลว์เชิงเวลาแบบละเอียด
  - UI web/desktop
  - การกระจายงานข้ามหลายเครื่องแบบซับซ้อน
</scope>

<requirements>
- Functional:
  - `agentflow init` → สร้างโครง `.agentflow/` + SPEC.md (โครงแท็กพร้อมใช้งาน)
  - `agentflow intake` → เติม `<problem>`, `<scope>`, `<requirements>` ใน SPEC.md
  - `agentflow plan` → เติม `<userstories>`, `<acceptance>` และสร้าง/อัปเดต `tasks.json`
  - `agentflow design` → เติม `<architecture>`, `<decisions>`, `<tradeoffs>` และ `architecture.md`
  - `agentflow qa` → เติม `<testplan>` และขยาย `<acceptance>`
  - `agentflow devlead` → เติม `<implementation_plan>` และจัดลำดับงาน/ความเสี่ยง
  - `agentflow devops` → เติม `<delivery>` (CI/CD/env/rollout)
  - `agentflow validate` → ตรวจโครงแท็ก XML (เปิด–ปิด/allowlist) + JSON schema
  - `agentflow export` → รวมเอกสาร SPEC/ADR/Test Plan เป็นแพ็กอ่านง่าย
- Data artifacts:
  - `SPEC.md` (Markdown+XML blocks), `tasks.json`, (optional) `openapi.json`, `datamodel.json`
- Config:
  - ตั้งค่าผ่านไฟล์ใน `.agentflow/config/` และ ENV (Viper-friendly)
</requirements>

<nonfunctional>
- ใช้ได้บน macOS/Linux/Windows
- โครงไฟล์เรียบง่าย ใช้ Git ได้เต็มรูปแบบ
- จำกัดพื้นที่เขียนของเอเจนต์ไว้เฉพาะโฟลเดอร์ที่อนุญาต
- เวลาเรียกเอเจนต์หนึ่งบทบาทเสร็จในระดับวินาที–นาที (ขึ้นกับโมเดล/เครือข่าย)
</nonfunctional>

<userstories>
- PM/BA: เติมปัญหา/ขอบเขต/ความต้องการใน SPEC.md เพื่อให้ทีมเข้าใจตรงกัน
- Planner: แตกงานพร้อม DoD/AC ใน `tasks.json` เพื่อให้ dev ทำงานได้ตรง
- Architect: ระบุโครงสร้างระบบ/ข้อแลกเปลี่ยน/การตัดสินใจใน SPEC.md + `architecture.md`
- QA: เขียน test plan ผูกกับ AC เพื่อรับงานได้จริง
- Dev Lead: วางลำดับงาน/ความเสี่ยง และกำหนดเกณฑ์แล้วเสร็จ
- DevOps: กำหนดกระบวนการส่งมอบ (CI/CD, environments, rollout/rollback)
</userstories>

<acceptance>
- หลัง `agentflow init` ต้องมี `.agentflow/` และ SPEC.md ที่มีแท็กพื้นฐานพร้อมใช้งาน
- หลัง `agentflow plan` ต้องมี `tasks.json` ที่โครงสร้างครบ (id/title/DoD/AC/dependsOn)
- หลัง `agentflow design` ต้องมี `architecture.md` อธิบายส่วนประกอบ/ขอบเขต/การตัดสินใจ
- `agentflow validate` ต้องแจ้งข้อผิดพลาดเมื่อแท็ก XML ไม่ปิด/ไม่อยู่ใน allowlist หรือ JSON ไม่ผ่านสคีมา
- `agentflow export` ต้องรวมเป็นเอกสารเดียวพร้อมส่งให้ทีมพัฒนา
</acceptance>
