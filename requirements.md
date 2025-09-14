
# requirements.md — Intake → SRS → Stories (Templates)

เอกสารนี้เป็น **ชุดเทมเพลตและแนวทาง** สำหรับเก็บความต้องการ, สร้าง SRS, แตกเป็น User Stories/AC, และจัดการขอบเขต/ความเสี่ยง ให้ Agents/มนุษย์ใช้ร่วมกันได้ทันที โดยโฟกัสการสร้าง “เครื่องมือ CLI ด้วยภาษา Go (Golang)” แทนเว็บแอป

---

## 1) ขั้นตอนเก็บความต้องการ (Requirement Intake)

**ถามทีละหัวข้อ (แนะนำ):**
1. เป้าหมายธุรกิจ (Business Goal/KPI)
2. กลุ่มผู้ใช้หลัก (Personas/Roles)
3. กรณีใช้งานหลัก 3 ข้อ (Top Use Cases)
4. Functional requirements หลัก
5. Non‑functional (Performance/Security/Availability/Compliance)
6. ข้อจำกัด (Time/Budget/Dependencies/Tech choices)
7. ข้อมูลอ่อนไหว/PDPA & data retention
8. ระยะเวลา/เดดไลน์โดยประมาณ

**ตัวอย่างผลลัพธ์ JSON (intake Q&A):**
```json
{
  "questions": [
    { "key": "business_goal", "ask": "ตัวชี้วัดความสำเร็จ (KPI) ของโปรเจกต์นี้คืออะไร?" },
    { "key": "primary_users", "ask": "ผู้ใช้หลักเป็นใครบ้าง และมี use case อะไร?" },
    { "key": "constraints", "ask": "มีข้อจำกัดด้านเวลา/งบ/เทคโนโลยีอะไรบ้าง?" }
      ],
      "decisions": [
        "ใช้ Go CLI (Cobra + Viper)",
        "ไม่ใช้ฐานข้อมูล — จัดเก็บข้อมูลในไฟล์ระบบภายใต้ ./.agentflow เพื่อให้ commit ด้วย Git ได้",
        "รูปแบบไฟล์มาตรฐาน: JSON/Markdown ตาม schema"
      ],
  "scope": {
    "mvp": ["คำสั่ง agentflow init/intake/srs/plan/run", "บันทึก artifacts ในโฟลเดอร์โปรเจกต์", "โหมด --json สำหรับเครื่องอ่าน"],
    "out_of_scope": ["Web UI", "ฐานข้อมูลถาวร/เซิร์ฟเวอร์ถาวร"]
  }
}
```

---

## 2) SRS Template (ย่อ, อิง IEEE‑830)

```
# Software Requirements Specification (SRS)

## 1. บทนำ
- วัตถุประสงค์, ขอบเขต, คำจำกัดความ, เอกสารอ้างอิง

## 2. ภาพรวมระบบ
- Personas, Context, สมมติฐาน/ข้อจำกัด

## 3. ความต้องการเชิงหน้าที่ (FR)
- Use Cases (UC‑01, UC‑02, ...)
- รายละเอียด Input/Output/ข้อยกเว้น

## 4. ความต้องการเชิงไม่เป็นหน้าที่ (NFR)
- Performance (เช่น P95 < 500ms), Availability (99.9%), Security (OWASP/PDPA), Usability, Accessibility

## 5. ข้อจำกัด (Constraints)
- เทคโนโลยี, งบประมาณ, เวลา, นโยบายองค์กร

## 6. ข้อมูล & ความเป็นส่วนตัว
- Data model (ย่อ), Data retention, Masking/Encryption, PDPA data flow

## 7. ความเสี่ยง & การบรรเทา
- ความเสี่ยง + แผนสำรอง

## 8. เกณฑ์ยอมรับ & การทดสอบ
- Acceptance Criteria ภาพรวม, แนวทางทดสอบ

## 9. ภาคผนวก
- คำจำกัดความ, มาตรฐาน, ลิงก์อ้างอิง
```

---

## 3) User Stories & Acceptance Criteria

**รูปแบบ Story:**
```
As a <persona>, I want <capability>, so that <benefit>.
```

**AC แบบ BDD (Given‑When‑Then):**
```
Given ฉันล็อกอินแล้ว
When ฉันกดปุ่ม Add Task พร้อมกรอกชื่อ
Then ระบบสร้าง task ใหม่ และแสดงในรายการภายใน 1 วินาที
```

**โครง JSON (stories.json):**
```json
{
  "stories": [
    {
      "id": "US-01",
      "title": "Create task",
      "ac": [
        "Given user is authenticated; When submit valid name; Then new task appears in list within 1s",
        "Given name is empty; When submit; Then show validation message"
      ]
    }
  ]
}
```

---

## 4) NFR Catalogue (ตัวอย่าง)

- **Performance**: P95 latency < 500ms (อ่าน), < 1s (เขียน); Throughput >= 100 rps
- **Availability**: 99.9% (M/M/S); RTO <= 30m, RPO <= 5m
- **Security**: OWASP Top 10, TLS 1.2+, secret rotation, least privilege, audit trail
- **Privacy/PDPA**: data minimization, consent tracking, purpose limitation, retention policy
- **Accessibility**: WCAG AA (สำคัญกับส่วน FE)
- **Observability**: Trace, metrics, structured logs (correlation id)
- **Compliance**: PDPA (TH), GDPR (ถ้ามีต่างประเทศ)

> *หมายเหตุ:* ไม่ใช่คำแนะนำทางกฎหมาย ตรวจสอบกับฝ่ายคอมพลายแยกต่างหาก

---

## 5) Decision Log (Template)

```
# Decision Log
- DL‑001: เลือก Go (Golang) สำหรับสร้าง CLI เนื่องจากทีมถนัดและแจกจ่ายเป็น binary ได้ง่าย (2025‑09‑14)
- DL‑002: ใช้ Cobra สำหรับโครงสร้างคำสั่ง และ Viper สำหรับ config/ENV (2025‑09‑14)
- DL‑003: (ถ้าจำเป็น) สื่อสารกับบริการภายนอกผ่าน HTTP/GRPC; หลีกเลี่ยง dependency หนักเกินไป (2025‑09‑14)
```

---

## 6) การจัดการขอบเขต (Scope)

- **MVP**: ฟีเจอร์ที่ต้องมีเพื่อใช้งานจริง
- **Nice‑to‑have**: ทำภายหลัง
- ใช้ **change request**: ทุกการเพิ่มงานใหม่ต้องผ่านการประเมินผลกระทบ (timeline/budget)

---

## 7) ตัวอย่าง (โปรเจกต์ Todo ทีม)

- MVP: สมัคร/ล็อกอิน, CRUD task, ฟิลเตอร์/ค้นหา, export CSV
- NFR: P95 < 500ms, 99.9% uptime, PDPA ready
- Risks: vendor lock‑in (ผ่อนหนักด้วย abstraction), DB scale (อ่าน replica)
- Stories: US‑01 Create, US‑02 Update, US‑03 Delete, US‑04 Filter/Search
- AC: ดูตัวอย่างด้านบน


---

## 8) นโยบายการจัดเก็บข้อมูล (Filesystem-only)
- ไม่ใช้ฐานข้อมูลถาวร ข้อมูลทั้งหมดจัดเก็บเป็นไฟล์ภายในโฟลเดอร์โปรเจกต์ เพื่อให้ตรวจสอบ/เวอร์ชันด้วย Git ได้
- รูปแบบไฟล์มาตรฐาน: JSON (.json) สำหรับข้อมูลเชิงโครงสร้าง, Markdown (.md) สำหรับเอกสารที่อ่านง่าย
- ข้อมูลแต่ละรายการควรมี metadata (createdAt, updatedAt, by, version) และ checksum หากจำเป็น
- เส้นทางหลัก: `./.agentflow/...` โฟลเดอร์ซ่อนสำหรับเก็บข้อมูลทั้งหมดของโปรเจกต์ (Git-friendly)

## 9) โครงสร้างโฟลเดอร์ .agentflow (ข้อเสนอแนะ)
```
.agentflow/
  project.json            # manifest โปรเจกต์ (name, description, owners, createdAt, version)
  README.md               # คำอธิบายสั้น + วิธีใช้ CLI สำหรับโปรเจกต์นี้
  intake/
    requirements.json     # Q&A รวม, ข้อกำหนดที่ยืนยันแล้ว
    summary.md            # สรุป Markdown สำหรับอ่านเร็ว
  analysis/
    SRS.md                # เอกสาร SRS ที่สร้างโดย SA agent
    stories.json          # User stories + AC
    risks.md              # ความเสี่ยง/การบรรเทา (ถ้ามี)
  planning/
    tasks.json            # รายการงาน (WBS/dependsOn/estimate)
    plan.md               # สรุปแผนการทำงาน
  runs/
    YYYYMMDD-HHMMSS-<id>/
      agent.json          # บันทึกการรัน agent (prompt I/O, model, params)
      logs.ndjson         # บันทึกแบบบรรทัดละ JSON
      artifacts/          # ผลลัพธ์ย่อยของการรันครั้งนี้
  artifacts/
    <artifact-id>/
      meta.json           # ชนิด, เวอร์ชัน, สคีมา, แหล่งที่มา (runId/taskId)
      data.json           # เนื้อหาหลัก (หรือ .md/.txt/.csv ตามชนิด)
  attachments/            # ไฟล์ที่ผู้ใช้แนบเข้ามา (PDF, ภาพ, ฯลฯ)
  prompts/                # แม่แบบพรอมพ์/สคีมาที่ใช้ร่วมกัน
  configs/
    cli.yaml              # ค่าคอนฟิกเฉพาะโปรเจกต์ (override ได้ด้วย ENV)
  agents/                 # ลงทะเบียน/คอนฟิกเอเจนต์ (allowlist, capabilities)
    registry.json
  schemas/                # JSON Schemas สำหรับ validate ไฟล์ต่างๆ
  indexes/                # ดัชนีค้นหา/embeddings (ถ้ามี)
  cache/                  # แคชที่ไม่ต้อง commit
  tmp/                    # ไฟล์ชั่วคราว
```

### ตัวอย่างไฟล์ manifest (project.json)
```json
{
  "name": "myproj",
  "slug": "myproj",
  "version": 1,
  "owners": ["you@company.com"],
  "createdAt": "2025-09-15T03:28:00+07:00",
  "description": "โปรเจกต์ตัวอย่างสำหรับ agentflow CLI",
  "conventionsVersion": "1.0.0"
}
```

### คำแนะนำ .gitignore ภายใน .agentflow/
```
# ไม่ต้องเวอร์ชันไฟล์ชั่วคราว/แคช
cache/
tmp/
**/*.tmp
**/*.log
runs/*/logs.ndjson
```

---

## หมายเหตุสำหรับ CLI (Go)
- โปรดออกแบบ **Scope/NFR** โดยคำนึงถึงประสบการณ์บรรทัดคำสั่ง: ชัดเจน สั้น อธิบายได้ด้วย `--help`
- รองรับไฟล์ config (`~/.config/<app>/config.yaml`) และ ENV override ด้วย Viper
- หากมีข้อมูลอ่อนไหว ให้จัดเก็บแบบเข้ารหัสใน keychain/local secure storage หรืออ่านผ่าน ENV ตามหลักความปลอดภัย
