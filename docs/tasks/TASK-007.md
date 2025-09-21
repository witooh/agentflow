# TASK-007 — LangGraph: integrate real backend (auth, endpoints, health, retries)
**Date:** 2025-09-15

## ContextEngineering
```xml
<task>LangGraph: integrate real backend (auth, endpoints, health, retries)</task>
<context>
  <requirement># AgentFlow — Requirements (v2)

**Date:** 2025-09-15

## เป้าหมายเฉพาะของงานนี้
- เชื่อมต่อ LangGraph ของจริงผ่าน REST แทน mock ให้ครบวงจร (health, /agents/run, /agents/questions)
- รองรับ Authorization: Bearer จาก env LANGGRAPH_API_KEY
- ปรับ config ให้เลือก baseUrl ได้จาก .agentflow/config.json และ trim trailing slash
- รีไทร่ด้วย backoff/jitter สำหรับ error แบบ transient
- ครอบทับ timeout ค่าเริ่มต้น 30s และปรับได้

## ขอบเขต
- ไม่เปลี่ยน CLI flags ใหญ่ ๆ แต่ทำให้การตั้งค่า baseUrl และ API key ใช้งานได้จริง
- เพิ่ม/ปรับ unit tests โดยไม่พึ่ง network จริง (ใช้ httptest)
</requirement>
  <srs># ประเด็นสถาปัตยกรรม/บูรณาการ
- internal/langgraph/client.go: ตรวจสอบ base URL, header Authorization, timeout, retry
- internal/config: field security.envKeys และ langgraph.baseUrl ใช้งานจริง
- docker-compose ใช้ทดสอบกับ mock ได้ แต่ unit test ต้องไม่แตะ network
</srs>
  <stories># Stories ที่เกี่ยวข้อง
- STORY-LG-1: As a developer, I can configure a real LangGraph endpoint and API key.
- STORY-LG-2: As a developer, I get reliable retries on transient failures.
- STORY-LG-3: As a developer, CLI respects baseUrl and timeout from config.
</stories>
</context>
```

## Implement
- ตรวจทาน internal/langgraph/client.go ว่ามี:
  - trimTrailingSlash(baseUrl)
  - header Authorization: Bearer ${LANGGRAPH_API_KEY} เมื่อมีค่าใน env
  - Retry เฉพาะ transient (timeout, connection reset, DNS) สูงสุด 3 ครั้ง พร้อม backoff + jitter
  - Health check GET /healthz → 200 ok
  - POST /agents/run และ /agents/questions ส่ง/รับ JSON ตามเอกสาร
- เพิ่ม unit tests ใน internal/langgraph/client_test.go ครอบคลุม:
  - trimTrailingSlash
  - transient(...) ตรวจจับ net.Error ที่ Timeout()=true
  - ใส่ Authorization header เมื่อมี LANGGRAPH_API_KEY
  - เคส retry โดย mock server ล้มชั่วคราวแล้วสำเร็จรอบถัดไป

## Subtasks
1. อ่านและยืนยันโค้ด client ปัจจุบันว่าครอบคลุมข้อกำหนด ถ้ายัง ให้ปรับปรุงโค้ดให้น้อยที่สุด
2. เพิ่ม/ปรับ unit tests ให้สะท้อนพฤติกรรมข้างต้น โดยใช้ httptest.Server
3. ปรับเอกสาร GUIDELINES/README สั้น ๆ ถ้าจำเป็น (optional ในรอบนี้)
4. ทดสอบแบบแห้งด้วย docker-compose (mock) เพื่อยืนยัน flow CLI ยังปกติ

## Definition of Done (DoD)
- [ ] เรียกใช้งานจริงกับ baseUrl และ API key ได้ (ทดสอบด้วย mock + review โค้ดสำหรับ production)
- [ ] Unit tests langgraph ผ่านทั้งหมด และไม่มี network calls จริง
- [ ] ไม่พัง CLI คำสั่ง: intake, plan, devplan เมื่อใช้ --dry-run
- [ ] เอกสาร config/ENV ระบุ LANGGRAPH_API_KEY ชัดเจน

## Risk
- การ retry ที่ aggressive อาจทำให้ latency สูง → จำกัดรอบ/เวลา
- สัญญา API ฝั่ง LangGraph เปลี่ยน → ทำให้ struct JSON ต้องอัปเดต

## Notes
- คงค่าเริ่มต้นจาก config.DefaultConfig; baseUrl: http://localhost:8123
- ใช้ CGO_ENABLED=0 ตามสคริปต์ build, คง timeout 30s เว้นผู้ใช้ override

## Traceability
- Linked: TASK-002 (REST client), Stories LG-1..LG-3
