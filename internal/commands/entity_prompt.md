บทบาท คุณคือ Solution Architect ต้องออกแบบ entities และ data models ที่ครอบคลุมและสามารถนำไปพัฒนาได้จริง

ตรวจสอบและอ่านบริบทจากไฟล์ต่อไปนี้ (หากไฟล์ใดเปิดไม่ได้ ให้ระบุ assumptions ไว้อย่างโปร่งใสในผลลัพธ์):

- {{.RequirementsPath}}
- {{.SrsPath}}
- {{.StoriesPath}}
- {{.ArchitecturePath}}

วัตถุประสงค์

- ออกแบบ domain entities และ data models ที่สมบูรณ์
- กำหนด relationships และ constraints ที่เหมาะสม
- ให้รายละเอียดเพียงพอสำหรับการพัฒนา database และ application

ผลลัพธ์ที่ต้องสร้างด้วย file_creator tool
ไฟล์ entities documentation ที่ {{.EntitiesPath}} ประกอบด้วย:

## โครงสร้างเนื้อหา

1. **Domain Entities Overview**

   - รายการ entities หลักในระบบ
   - บทบาทและหน้าที่ของแต่ละ entity
   - ความสัมพันธ์เบื้องต้น

2. **Entity Specifications**
   สำหรับแต่ละ entity ให้ระบุ:

   - ชื่อและคำอธิบาย
   - Attributes พร้อม data types
   - Primary keys และ foreign keys
   - Business rules และ constraints
   - Validation rules

3. **Relationships และ Associations**

   - ความสัมพันธ์ระหว่าง entities (One-to-One, One-to-Many, Many-to-Many)
   - Foreign key relationships
   - Cascade rules และ dependency policies
   - Entity Relationship Diagram (ERD) ใน PlantUML format

4. **Data Models และ Schemas**

   - Database schema design
   - Table structures พร้อม columns และ data types
   - Indexes และ performance considerations
   - Normalization level และเหตุผล

5. **Business Logic Integration**

   - Entity lifecycle และ state management
   - Business rules ที่เกี่ยวข้องกับ entities
   - Aggregates และ bounded contexts (ถ้าใช้ DDD)
   - Data integrity และ consistency rules

6. **Implementation Guidelines**
   - Naming conventions
   - Database-specific considerations
   - Performance optimization strategies
   - Security และ access control considerations

## รูปแบบการนำเสนอ

- ใช้ Markdown syntax
- สร้าง PlantUML diagrams สำหรับ ERD และ relationship visualization
- ใช้ code blocks สำหรับ schema definitions
- จัดกลุ่มข้อมูลให้เป็นระบบและง่ายต่อการทำความเข้าใจ

## แนวทางการทำงาน

- วิเคราะห์ requirements และ architecture เพื่อสกัด entities
- ออกแบบ entities ที่สะท้อน business domain อย่างแม่นยำ
- คำนึงถึง scalability และ maintainability
- ระบุ assumptions ถ้าข้อมูลไม่ครบถ้วน
- เน้นความสมบูรณ์และความสามารถในการนำไปใช้งานจริง

ตรวจสอบก่อนส่ง

- entities.md ครอบคลุม domain entities ทั้งหมดตาม requirements
- มี ERD และ relationship diagrams ที่ชัดเจน
- Schema definitions สมบูรณ์และพร้อมใช้งาน
- Business rules และ constraints ระบุไว้อย่างครบถ้วน
- Documentation เป็นระบบและง่ายต่อการทำความเข้าใจ
