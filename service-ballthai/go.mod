module go-ballthai-scraper

go 1.22

require (
	github.com/go-sql-driver/mysql v1.8.1 // ตรวจสอบเวอร์ชันล่าสุดที่ https://pkg.go.dev/github.com/go-sql-driver/mysql
	github.com/joho/godotenv v1.5.1 // เพิ่มไลบรารีสำหรับโหลด .env
)

require filippo.io/edwards25519 v1.1.0 // indirect

// สามารถเพิ่ม require อื่นๆ ได้ในอนาคต หากใช้ไลบรารีเพิ่มเติม
