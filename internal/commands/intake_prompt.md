### 🎯 Output format (Markdown)

- **Business Goals & Success KPIs**  
  - Describe business drivers (compliance, UX, marketing agility, cost savings).  
  - Define measurable KPIs (e.g., opt-in rate target, consent sync SLA, regulator reporting turnaround).  

- **User Personas & Journeys**  
  - Customer (mobile/web) → manage consent, banner UX.  
  - Marketing/Analytics team → use dashboard, reporting.  
  - Regulator/Audit → compliance log, proof of consent.  
  - Backend/System Integrator → consume consent via API/SDK.  

- **Scope (MVP vs Future Phases)**  
  - Clearly separate **MVP features** vs **future expansion**.  
  - Use MoSCoW or phased roadmap (MVP → Phase 2 → Mature state).  

- **Functional Requirements (FR)**  
  - Detail user-facing and system-facing capabilities.  
  - Link each FR back to persona & business goal.  

- **Non-Functional Requirements (NFR)**  
  - Scale, latency, retention, compliance, UX accessibility, availability.  
  - Prioritize what is critical at MVP vs later.  

- **Dependencies & Risks**  
  - Dependencies on other teams (e.g., Data Lake, Security, Compliance).  
  - Risks (regulatory, adoption, tech feasibility).  

- **Constraints**  
  - Jurisdiction: Thailand only (PDPA).  
  - Data residency: PDPA compliant.  
  - Migration: cutover from OneTrust (big bang).  
  - Certifications: not required at MVP.  

- **Deliverables to Solution Architect (SA)**  
  - Consent use cases & flows (opt-in, revoke, merge, reporting).  
  - High-level data model (consent record, audit log, mapping to customer/device).  
  - Integration points (mobile, web, backend, data lake).  
  - Prioritized features (MVP vs future).  
  - Reporting requirements (dimensions, regulator templates).  

- **Timeline Summary (Product Roadmap)**  
  - Narrate evolution chronologically:  
    - MVP (core consent, banner, reporting baseline).  
    - Phase 2 (advanced analytics, audience targeting, cookie discovery).  
    - Future (scalability, certifications, multi-region compliance).  

- **Questions to Human (Stakeholders)**  
  - Business-side clarifications (regulator reporting expectation, marketing KPIs, branding rules).  
  - Technical-side clarifications (DB choice, API standards, realtime infra).  

เมื่อสรุปข้อมูลทั้งหมดแล้ว ให้สร้าง Markdown เดียวบันทึกที่ {{.RequirementsPath}}

เขียนเนื้อหาให้อ่านง่าย กระชับ มีหัวข้อย่อยครบถ้วน และระบุ assumptions หรือข้อมูลที่ต้องถามต่ออย่างโปร่งใส
