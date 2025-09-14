
# ARCHITECTURE.md — TS‑first System Design

สถาปัตยกรรมเน้น **TypeScript เป็นศูนย์กลาง** (Next.js + LangGraph + OpenAI) และรองรับ **Polyglot workers** (เช่น Python) ผ่าน **RabbitMQ (AMQP)** เพื่อแยกงานที่ TS ไม่ถนัด

---

## 1) High‑level Overview

```mermaid
C4Context
title AI Software House (TS-first)

Person(User, "Human PM/Owner")
System_Boundary(WebApp, "Next.js (TS)") {
  Container(UI, "Web UI", "Next.js + shadcn/ui", "Chat/Tasks/Artifacts")
  Container(API, "API/Orchestrator", "Next.js Route Handlers + LangGraph", "Agent graph, events")
  ContainerDb(DB, "Postgres + pgvector", "RDS/Supabase", "Projects/Tasks/Artifacts/Memories")
  Container(Realtime, "WS/Realtime", "Socket.IO", "Live updates")
}
System(MQ, "RabbitMQ (AMQP)", "Messaging")
System(Workers, "Polyglot Workers", "Python/FastAPI", "ML/OCR/PDF/etc.")
User -> UI: idea & answers
UI -> API: kickoff / commands
API -> DB: persist state
API -> MQ: publish jobs (topic)
Workers -> MQ: results
API -> Realtime: push events
```

---

## 2) Components

- **UI (Next.js)**: หน้า `/new-project`, `/project/[id]` (Chat, SRS, Tasks Kanban, Agents, Artifacts, Activity)
- **API/Orchestrator (TS)**: LangGraph state machine (intake → sa → architect → planner → dev/qa → review)
- **DB (Postgres + pgvector)**: ตาราง projects, requirements, artifacts, tasks, task_dependencies, agent_runs, messages, memories
- **Messaging (RabbitMQ)**: cross‑language jobs/results
- **Workers (Python)**: งาน OCR/ML/PDF‑heavy
- **Realtime (Socket.IO/Supabase Realtime)**: สตรีมสถานะงาน/เอเจนต์

---

## 3) Data Model (สรุป)

ตารางหลัก (รายละเอียด SQL ดูไฟล์หลัก):  
- `projects`, `requirements`, `artifacts`, `tasks`, `task_dependencies`  
- `agent_runs`, `messages`, `memories (pgvector)`

---

## 4) API & Versioning

- RESTful `/api/projects`, `/api/tasks`, `/api/agents/events`  
- Versioning: `Accept: application/vnd.ai-webapp.v1+json` หรือ path `/api/v1/*`  
- Response แบบ **DTO ที่เสถียร** (ห้ามรั่ว field ภายใน)  
- Pagination & idempotency keys สำหรับ endpoints เขียน

---

## 5) Messaging & Contracts

- Exchange: `agent` (type: topic)  
- Routing keys: `ml.*`, `doc.parse.*`, `qa.run.*`, `ops.*`  
- Envelope (ดู AGENTS.md) — ต้อง **versioned** และตรวจ schema ทุกครั้ง  
- ผลลัพธ์ publish กลับด้วย `*.done` + payload อ้างอิง `taskId`

---

## 6) Security

- **Auth**: Supabase Auth/NextAuth (JWT)  
- **RLS**: บังคับผ่าน project_id ทุกตาราง (ถ้าใช้ Supabase)  
- **Secrets**: `.env` + Secret Manager สำหรับโปรดักชัน  
- **CORS/CSRF**: คุม origin และใช้ SameSite cookies  
- **Least privilege**: แยก service accounts ตามบทบาท  
- **Prompt safety**: sanitize context, block untrusted tools

---

## 7) Observability

- OpenTelemetry trace + metrics  
- Structured logs (pino) ใส่ correlation id (`projectId`, `taskId`, `runId`)  
- Error budget & SLO (เช่น 99.9% uptime) 

---

## 8) Performance & Scalability

- Horizontal scale: API stateless, Workers autoscale  
- Queue‑backpressure ผ่าน RabbitMQ; ใช้ prefetch จำกัดงานต่อ worker  
- DB: index/partition ตาม usage; อ่าน replica สำหรับ heavy read  
- Cache: edge caching (Vercel) สำหรับ public metadata (ถ้ามี)  

---

## 9) Deployment

- **FE/API**: Vercel (preview per PR)  
- **DB**: Supabase/RDS (migrations via SQL)  
- **MQ**: CloudAMQP (managed RabbitMQ)  
- **Workers**: Docker on Fargate/Cloud Run/K8s  
- **Secrets**: Manager ของ cloud ที่ใช้

**.env ตัวอย่าง**
```bash
OPENAI_API_KEY=...
OPENAI_MODEL=gpt-5.0-thinking
DATABASE_URL=postgres://...
SUPABASE_URL=...
SUPABASE_ANON_KEY=...
AMQP_URL=amqp://guest:guest@localhost:5672/
```

---

## 10) Failure Modes & Resilience

- Retry with DLQ (dead‑letter queue) ต่อ routing‑key สำคัญ  
- Idempotency key สำหรับคำสั่งซ้ำ  
- Circuit breaker สำหรับ calls ภายนอก  
- Graceful shutdown (drain messages) ใน workers

---

## 11) Mermaid Sequences

### Kickoff → Results
```mermaid
sequenceDiagram
  participant UI
  participant API as API (TS)
  participant MQ as RabbitMQ
  participant PY as Worker (Python)
  participant DB

  UI->>API: POST /api/projects
  API->>DB: create project
  API->>MQ: publish ml.extract_keywords
  PY-->>MQ: consume ml.*
  PY->>MQ: publish ml.extract_keywords.done
  API->>DB: persist outputs
  API-->>UI: ws: progress/update
```

---

## 12) Roadmap

- GitHub PR auto‑create per agent  
- Evaluation harness (self‑critique)  
- Playwright/Vitest/k6 CI  
- Knowledge base ต่อโปรเจกต์ (embeddings)  
- Monte Carlo timeline estimator
