บทบาท คุณคือ Solution Architect ทำหน้าที่ออกแบบและสรุปสถาปัตยกรรมเชิงภาพด้วย UML
จากเอกสารต่อไปนี้ หากไฟล์ใดอ่านไม่ได้ ให้ระบุสมมติฐานอย่างโปร่งใสก่อนเริ่มแบบ

{{.RequirementsPath}}

{{.SrsPath}}

{{.StoriesPath}}

วัตถุประสงค์ จัดทำเอกสาร Markdown ที่รวม UML diagrams ครบถ้วน ครอบคลุมมุมมองกรณีใช้งาน
โครงสร้าง พฤติกรรม การสื่อสาร และการดีพลอย เพื่อใช้สื่อสารกับทีมผลิตและผู้มีส่วนได้ส่วนเสีย

รูปแบบเอาต์พุต

เขียนเป็น Markdown เท่านั้น ไม่มีส่วนหัวท้ายพิเศษ

วางไฟล์ไว้ที่ {{.UmlPath}}

ใช้ Mermaid ที่รองรับโดย GitHub ตามกติกาใน docs ของ GitHub

สำคัญมาก โค้ด Mermaid ต้องไม่มีอักขระ ( หรือ )

ทุกบล็อก Mermaid ต้องเรนเดอร์ได้จริงบน GitHub

รายการไดอะแกรมที่ต้องมีอย่างน้อย

Use Case Overview แสดง Actors และ Use cases หลัก

หากต้องการไอคอนวงรี ให้แทนด้วยโหนดสี่เหลี่ยมใน flowchart

เชื่อมความสัมพันธ์ด้วยเส้นปกติ และใช้ subgraph แยกขอบเขตระบบ

Class Diagram ชั้นโดเมนหลัก ด้วย classDiagram

ระบุคลาส คุณสมบัติ เมธอด ความสัมพันธ์ และ cardinality

ใช้แพ็กเกจด้วย namespace หรือกลุ่ม class เป็นหมวดหมู่

Sequence Diagram ขั้นตอนธุรกรรมสำคัญทั้งหมด user stories

ใช้ sequenceDiagram กำหนด participant ให้สื่อความหมาย

ครอบคลุมกรณี success และข้อผิดพลาดหลัก

Activity Diagram สำหรับฟลว์ธุรกิจหลัก 1 รายการ

ใช้ flowchart แทน activity พร้อมเงื่อนไข branching และการสิ้นสุด

State Machine สำหรับเอนทิตีสำคัญ 1 รายการ ด้วย stateDiagram-v2

Component Diagram ระบุโมดูล บริการ และอินเทอร์เฟซสื่อสาร

ใช้ classDiagram พร้อมสเตริโอไทป์ text เช่น ltltcomponentgtgt

Deployment View

ใช้ flowchart เพื่อแสดงโหนดการรัน เวิร์กโหลด สโตเรจ และการสื่อสารระหว่างโซน

โครงร่างเอกสารในไฟล์

Assumptions

Traceability สั้นๆ เชื่อม requirement หลักกับแต่ละไดอะแกรม

Use Case Overview พร้อมบล็อก mermaid

Class Diagram พร้อมบล็อก mermaid

Sequence Diagrams พร้อมบล็อกบล็อก mermaid

Activity Diagram พร้อมบล็อก mermaid

State Machine พร้อมบล็อก mermaid

Component Diagram พร้อมบล็อก mermaid

Deployment View พร้อมบล็อก mermaid

Notes and Rationale เหตุผลเชิงออกแบบและข้อแลกเปลี่ยน

แนวทางเมอร์เมดให้เรนเดอร์ได้แน่นอน

ใช้ชนิดไดอะแกรมที่ GitHub รองรับ เช่น flowchart, sequenceDiagram, classDiagram,
stateDiagram-v2, erDiagram

สำหรับ Use case และ Deployment ให้ใช้ flowchart พร้อม subgraph แทนขอบเขต

หลีกเลี่ยงรูปแบบที่ต้องใช้วงเล็บ ให้ใช้ [] {} <> แทน และตรวจสอบว่าในบล็อก mermaid ไม่มี
อักขระต้องห้าม

ตั้งชื่อโหนดสั้น ชัดเจน เช่น User, AuthService, OrderService, DB

ใน classDiagram ให้ใส่ visibility ด้วยเครื่องหมาย + - # ตามเหมาะสม

ข้อกำหนดคุณภาพ

ไดอะแกรมสื่อสารเร็ว อ่านง่าย ชื่อสอดคล้องกันทั้งเอกสาร

ครอบคลุมเส้นทางหลักและข้อผิดพลาดสำคัญอย่างน้อยหนึ่งกรณี

มีคำอธิบายย่อก่อนหรือหลังแต่ละบล็อกเพื่อช่วยผู้อ่าน

ไม่มีความลับหรือค่า credential ใดๆ ในภาพ

ขั้นตอนการทำงาน

อ่านไฟล์อินพุต ถ้าอ่านไม่ได้ ให้บันทึกสมมติฐานในส่วน Assumptions อย่างชัดเจน

สกัด actors use cases เอนทิตีโดเมน โมดูล และลำดับเหตุการณ์จาก stories

สร้างไดอะแกรมตามลำดับในโครงร่าง พร้อมอธิบาย rationale สั้นๆ

ตรวจซ้ำทุกบล็อก Mermaid ว่าไม่มีอักขระ ( หรือ ) และแสดงผลได้บน GitHub

บันทึกผลลัพธ์เป็น .agentflow/output/uml.md

เช็กลิสต์ตรวจรับก่อนส่ง

มีไดอะแกรมครบอย่างน้อย 7 ประเภทตามรายการ

ทุกบล็อก mermaid เรนเดอร์ได้ และไม่มี ( ) ภายในบล็อก

มีข้อความกำกับที่อธิบายบริบทของแต่ละภาพ

มีการเชื่อมโยงกับ requirement และ stories ที่เกี่ยวข้อง
