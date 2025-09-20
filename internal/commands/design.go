package commands

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"agentflow/internal/agents"
	"agentflow/internal/config"
)

type DesignOptions struct {
	ConfigPath string
	SourceDir  string // where to read prior docs (srs/stories/acceptance_criteria). If empty, use cfg.IO.OutputDir
	OutputDir  string // where to write architecture.md and uml.md
	Role       string
	DryRun     bool
}

func Design(opts DesignOptions) error {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	cfg.ApplyEnv()
	if opts.OutputDir != "" {
		cfg.IO.OutputDir = opts.OutputDir
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	if err := config.EnsureDirs(opts.ConfigPath, cfg); err != nil {
		return err
	}

	systemMessages := createDesignSystemMessage(opts.SourceDir)

	if opts.DryRun {
		return nil
	} else {
		_, err := agents.SA.RunInputs(context.Background(), systemMessages)
		if err != nil {
			fmt.Printf("\n\n> Note: OpenAI call failed, wrote scaffold instead. Error: %v\n", err)
		}
	}

	return err
}

func createDesignSystemMessage(outputDir string) []agents.TResponseInputItem {
	return agents.InputList(
		agents.SystemMessage(fmt.Sprintf(`คุณคือ Solution Architect ที่ออกแบบสถาปัตยกรรมระบบและโครงสร้างโปรเจกต์โดยอ้างอิงจากไฟล์ต่อไปนี้ หากเข้าถึงไม่ได้ให้ทำงานแบบสมมติฐานที่ระบุชัดเจน
%s
%s
%s

วัตถุประสงค์
1 ออกแบบ Infrastructure Architecture และ Project File Structure
2 สร้างเอกสาร Markdown ที่มี Mermaid diagrams ตามกติกา GitHub Mermaid
3 บันทึกไฟล์ผลลัพธ์ไว้ที่ .agentflow/output/architecture.md

ข้อกำหนดเอาต์พุต
เอกสารเป็น Markdown เท่านั้น ห้ามมี header หรือ footer ที่ไม่จำเป็น
โค้ด Mermaid ต้อง ไม่ใช้อักขระ ( หรือ ) และต้องเป็นไวยากรณ์ที่ GitHub Mermaid รองรับ เช่น flowchart TD หรือ LR, subgraph, [] สำหรับ node
แยก diagram อย่างน้อย 2 ส่วน
Infrastructure Diagram
Project File Structure Diagram

เพิ่มส่วน Assumptions, Rationale, Security and Observability, CI CD overview แบบย่อใน Markdown ธรรมดา

ขั้นตอนการทำงาน
1 พยายามอ่านไฟล์อินพุตที่กำหนด หากอ่านไม่ได้ ให้ระบุในเอกสารส่วน “Assumptions” ว่าใช้สมมติฐานมาตรฐาน production แทน
2 สกัด requirement หลัก NFRs ถ้ามี เช่น throughput latency RTO RPO data residency budget หากไม่พบให้กำหนดค่าเริ่มต้นที่เหมาะสมและระบุเหตุผล
3 ออกแบบ Infrastructure ด้วย Mermaid ให้สอดคล้องกับข้อกำหนด เช่น cloud ที่ระบุ ถ้าไม่ระบุให้ใช้ AWS เป็นดีฟอลต์ พร้อมส่วนประกอบมาตรฐาน VPC subnets ALB compute service datastore cache object storage queue CDN WAF IAM secrets monitoring logging tracing และ CI CD
4 ออกแบบ Project structure ด้วย markdown ให้รองรับ backend และ frontend แยกกัน พร้อม infra tests automation และ configs ชัดเจน

เขียนผลลัพธ์ไปที่ %s เท่านั้น

เกณฑ์คุณภาพ

สั้น ชัด มีเหตุผลกำกับแต่ละกลุ่มบริการ
ปลอดภัยตามหลัก least privilege และมีการสังเกตการณ์ครบถ้วน
โค้ด Mermaid แสดงผลได้จริงใน GitHub
หากมีข้อจำกัดหรือสมมติฐาน ให้บอกอย่างโปร่งใสในส่วน Assumptions
รูปแบบไฟล์ผลลัพธ์
Markdown ล้วน มีส่วนย่อยต่อไปนี้เรียงลำดับ
Assumptions
Infrastructure overview and rationale
mermaid flowchart for infrastructure
CI CD overview เลือกใช้ flowchart อีกบล็อกได้
Project file structure for tree
Security and observability checklist
ห้ามทำสิ่งต่อไปนี้
ห้ามเขียนเนื้อหาใดๆ นอกเหนือไฟล์เป้าหมาย
ห้ามใส่วงเล็บในโค้ด Mermaid
ห้ามเปิดเผยคีย์หรือความลับใดๆ ในตัวอย่าง`,
			filepath.Join(outputDir, "requirements.md."),
			filepath.Join(outputDir, "srs.md"),
			filepath.Join(outputDir, "stories.md"),
			filepath.Join(outputDir, "architecture.md"),
		)),
	)
}

func splitDesignContent(s string) (string, string) {
	// Similar to plan split: find markers
	const archMark = "--- ARCH START ---"
	const umlMark = "--- UML START ---"
	idxA := strings.Index(s, archMark)
	idxU := strings.Index(s, umlMark)
	if idxA == -1 && idxU == -1 {
		return s, ""
	}
	var arch, uml string
	if idxA != -1 && idxU != -1 {
		if idxA < idxU {
			arch = strings.TrimSpace(s[idxA+len(archMark) : idxU])
			uml = strings.TrimSpace(s[idxU+len(umlMark):])
			return strings.TrimSpace(arch), strings.TrimSpace(uml)
		}
		// UML first (unexpected) but handle
		uml = strings.TrimSpace(s[idxU+len(umlMark) : idxA])
		arch = strings.TrimSpace(s[idxA+len(archMark):])
		return strings.TrimSpace(arch), strings.TrimSpace(uml)
	}
	if idxA != -1 {
		arch = strings.TrimSpace(s[idxA+len(archMark):])
		return strings.TrimSpace(arch), ""
	}
	uml = strings.TrimSpace(s[idxU+len(umlMark):])
	return "", strings.TrimSpace(uml)
}

func ensureArchitecture(s string) string {
	s = strings.TrimSpace(s)
	// Ensure Project Structure section exists
	return s
}

func ensureUML(s string) string {
	s = strings.TrimSpace(s)
	lower := strings.ToLower(s)
	needSeq := !strings.Contains(lower, "sequence")
	needClass := !strings.Contains(lower, "class")
	needAct := !strings.Contains(lower, "activity")
	if needSeq || needClass || needAct {
		var b strings.Builder
		b.WriteString(s)
		b.WriteString("\n\n")
		s = b.String()
	}
	return s
}
