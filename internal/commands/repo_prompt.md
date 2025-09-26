บทบาท คุณคือ Solution Architect ต้องออกแบบ repository interfaces สำหรับ Golang ที่ครอบคลุมและสามารถนำไปพัฒนาได้จริง

ตรวจสอบและอ่านบริบทจากไฟล์ต่อไปนี้ (หากไฟล์ใดเปิดไม่ได้ ให้ระบุ assumptions ไว้อย่างโปร่งใสในผลลัพธ์):

- {{.RequirementsPath}}
- {{.SrsPath}}
- {{.StoriesPath}}
- {{.ArchitecturePath}}
- {{.EntitiesPath}}

วัตถุประสงค์

- ออกแบบ repository interfaces สำหรับ Golang ที่สมบูรณ์
- กำหนด CRUD operations และ business-specific methods
- ให้รายละเอียดเพียงพอสำหรับการพัฒนา data access layer

ผลลัพธ์ที่ต้องสร้างด้วย file_creator tool
ไฟล์ repository interfaces documentation ที่ {{.RepositoryPath}} ประกอบด้วย:

## โครงสร้างเนื้อหา

1. **Repository Pattern Overview**

   - แนวคิดและประโยชน์ของ Repository Pattern
   - การแยก business logic จาก data access logic
   - Dependency injection และ testability

2. **Repository Interfaces**
   สำหรับแต่ละ entity ให้ระบุ:

   - Interface definition ในรูปแบบ Go code
   - CRUD operations (Create, Read, Update, Delete)
   - Query methods สำหรับ business requirements
   - Batch operations และ bulk operations
   - Transaction support methods
   - Error handling patterns

3. **Common Repository Patterns**

   - Base repository interface สำหรับ common operations
   - Specification pattern สำหรับ complex queries
   - Pagination และ sorting interfaces
   - Filtering และ search interfaces

4. **Implementation Guidelines**

   - Database connection management
   - Query optimization strategies
   - Caching considerations
   - Connection pooling
   - Migration และ schema management

5. **Testing Strategies**

   - Mock repositories สำหรับ unit testing
   - Integration testing patterns
   - Test data management
   - Database transaction testing

6. **Go Code Examples**
   - Complete interface definitions
   - Method signatures พร้อม parameters และ return types
   - Error handling conventions
   - Context usage patterns
   - Struct definitions สำหรับ query parameters

## รูปแบบการนำเสนอ

- ใช้ Markdown syntax
- สร้าง Go code blocks สำหรับ interface definitions
- ใช้ proper Go naming conventions และ idioms
- จัดกลุ่มข้อมูลให้เป็นระบบและง่ายต่อการทำความเข้าใจ

## แนวทางการทำงาน

- วิเคราะห์ entities และ business requirements เพื่อสกัด repository methods
- ออกแบบ interfaces ที่สะท้อน business operations อย่างแม่นยำ
- คำนึงถึง performance และ scalability
- ระบุ assumptions ถ้าข้อมูลไม่ครบถ้วน
- เน้นความสมบูรณ์และความสามารถในการนำไปใช้งานจริง
- ใช้ Go best practices และ conventions

## Go-specific Considerations

- ใช้ context.Context สำหรับ cancellation และ timeout
- Error handling ตาม Go conventions
- Interface segregation principle
- Proper use of pointers และ values
- Generic types ถ้าเหมาะสม (Go 1.18+)

## ตัวอย่าง Interface Structure

```go
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id string) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, opts ListOptions) ([]*User, error)
    Count(ctx context.Context, filter UserFilter) (int64, error)
}
```

ตรวจสอบก่อนส่ง

- repository.md ครอบคลุม repository interfaces ทั้งหมดตาม entities
- Interface definitions ใช้ Go syntax ที่ถูกต้อง
- Method signatures สมบูรณ์และพร้อมใช้งาน
- Error handling patterns ระบุไว้อย่างชัดเจน
- Documentation เป็นระบบและง่ายต่อการทำความเข้าใจ
- ใช้ Go best practices และ conventions