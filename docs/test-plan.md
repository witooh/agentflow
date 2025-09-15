# AgentFlow — Test Plan (v2)
**Date:** 2025-09-15

## Strategy
- Unit + Integration + E2E (snapshot docs)
- Mock LangGraph สำหรับ deterministic และ 1 ชุด E2E จริง
- Contract tests ฝั่ง REST

## Entry/Exit
- Entry: binary + docker langgraph + config ok
- Exit: docs ครบ, tasks ครบ, coverage ≥ 70%, ไม่มี secret ใน logs

## Risks
- Prompt drift, Context windows, API changes
