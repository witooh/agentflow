
# requirements.md — Intake → SRS → Stories (Templates)

เอกสารนี้เป็น **ชุดเทมเพลตและแนวทาง** สำหรับเก็บความต้องการ, สร้าง SRS, แตกเป็น User Stories/AC, และจัดการขอบเขต/ความเสี่ยง ให้ Agents/มนุษย์ใช้ร่วมกันได้ทันที

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
    "ใช้ Next.js + Postgres + RabbitMQ",
    "รองรับผู้ใช้ภาษาไทยเป็นหลัก"
  ],
  "scope": {
    "mvp": ["สมัคร/ล็อกอิน", "สร้าง/แก้ไขงาน", "ค้นหา/กรอง"],
    "out_of_scope": ["การแชร์งานข้ามองค์กร", "รายงานเชิงลึก"]
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
- DL‑001: เลือก Next.js เนื่องจากทีมถนัด, deploy ง่ายบน Vercel (2025‑09‑14)
- DL‑002: ใช้ RabbitMQ สำหรับ cross‑language messaging (2025‑09‑14)
- DL‑003: เก็บ embedding ใน pgvector (2025‑09‑14)
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

## หมายเหตุสำหรับ Supabase
- โปรดออกแบบ **Scope/NFR** โดยคำนึงถึง **RLS** และ **Storage policy** ตั้งแต่แรก
- หากมีข้อมูลอ่อนไหว ให้ระบุ **data classification** และ **retention** เพื่อแปลงเป็น policy SQL ได้ทันที
