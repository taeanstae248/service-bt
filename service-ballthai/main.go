package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	// "time" // ลบออกหรือ comment เพราะไม่ได้ใช้ใน main.go โดยตรง

	_ "github.com/go-sql-driver/mysql" // Driver สำหรับ MySQL
	"github.com/joho/godotenv"         // สำหรับโหลด .env

	"go-ballthai-scraper/database" // แก้ไข: เปลี่ยนเป็นชื่อโมดูลที่ถูกต้องตาม go.mod
	"go-ballthai-scraper/scraper"  // แก้ไข: เปลี่ยนเป็นชื่อโมดูลที่ถูกต้องตาม go.mod
)

var db *sql.DB // ตัวแปร Global สำหรับเก็บ Connection ฐานข้อมูล

func main() {
	// โหลดค่าจาก .env ไฟล์ก่อน
	log.Println("Attempting to load .env file...")
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v. Assuming environment variables are set directly.", err)
		// ไม่ต้อง Fatalf ที่นี่ เพราะอาจจะรันบน Server ที่ตั้งค่า env vars โดยตรง
	} else {
		log.Println(".env file loaded successfully.")
	}

	// 1. ตั้งค่าการเชื่อมต่อฐานข้อมูล
	// ดึงค่าจาก Environment Variables
	dbUser := os.Getenv("DB_USERNAME")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	// *** เพิ่ม Log เพื่อตรวจสอบค่าที่ได้มา ***
	log.Printf("DEBUG: DB_USERNAME: '%s'", dbUser)
	log.Printf("DEBUG: DB_PASSWORD: '%s' (length: %d)", dbPass, len(dbPass)) // ตรวจสอบความยาวของรหัสผ่าน
	log.Printf("DEBUG: DB_HOST: '%s'", dbHost)
	log.Printf("DEBUG: DB_PORT: '%s'", dbPort)
	log.Printf("DEBUG: DB_NAME: '%s'", dbName)

	// ตรวจสอบว่าค่าที่จำเป็นถูกตั้งค่าหรือไม่
	// ในกรณีของ XAMPP, root มักไม่มีรหัสผ่าน
	if dbUser == "" || dbHost == "" || dbPort == "" || dbName == "" {
		log.Fatalf("Missing one or more essential database environment variables (DB_USERNAME, DB_HOST, DB_PORT, DB_NAME)")
	}

	connStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)
	log.Printf("DEBUG: Connection String: %s", connStr) // แสดง Connection String (ระวังข้อมูล Sensitive ใน Log จริง)

	// เรียกใช้ InitDB จากแพ็กเกจ database
	err = database.InitDB(connStr)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	db = database.DB // กำหนด global db variable ให้ชี้ไปยัง Connection ที่ InitDB สร้างขึ้น
	defer db.Close()  // ปิด Connection เมื่อโปรแกรมทำงานเสร็จ

	log.Println("Successfully connected to the database!")

	// 2. สร้างโฟลเดอร์สำหรับเก็บรูปภาพ (ถ้ายังไม่มี)
	imageDirs := []string{
		"./img/coach",
		"./img/player",
		"./img/source", // สำหรับโลโก้ทีม (จาก API หลายตัว)
		"./img/stadiums",
	}
	for _, dir := range imageDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755) // สร้างโฟลเดอร์พร้อมสิทธิ์
			if err != nil {
				log.Fatalf("Failed to create image directory %s: %v", dir, err)
			}
		}
	}
	log.Println("Image directories ensured.")

	// 3. เริ่มต้นการ Scrape ข้อมูล
	log.Println("Starting data scraping process...")

	// *** ขั้นตอนสำคัญ: ตรวจสอบและเพิ่มลีกในฐานข้อมูลก่อนเสมอ ***
	log.Println("Ensuring Leagues exist in DB...")
	leaguesToEnsure := []string{
		"T1", "T2", "T3 Bangkok", "T3 East", "T3 West", "T3 North", "T3 Northeast", "T3 South",
		"Samipro", "Revo League Cup", "FA Cup", "BGC Cup", "Thai League Playoff",
	}
	for _, leagueName := range leaguesToEnsure {
		_, err := database.GetLeagueID(db, leagueName) // GetLeagueID จะเพิ่มลีกถ้ายังไม่มี
		if err != nil {
			log.Printf("Error ensuring league %s exists: %v", leagueName, err)
		}
	}
	log.Println("Leagues ensured in DB.")


	// Scrape ข้อมูลสนาม (ควรทำก่อนทีม เพราะทีมอาจอ้างอิง stadium_id)
	// log.Println("Scraping Stadiums...")
	// err = scraper.ScrapeStadiums(db)
	// if err != nil {
	// 	log.Printf("Error scraping stadiums: %v", err)
	// } else {
	// 	log.Println("Stadiums scraping completed.")
	// }

	// Scrape ข้อมูลโค้ช (อ้างอิง nationality และ team)
	// log.Println("Scraping Coaches...")
	// err = scraper.ScrapeCoach(db)
	// if err != nil {
	// 	log.Printf("Error scraping coaches: %v", err)
	// } else {
	// 	log.Println("Coaches scraping completed.")
	// }

	// Scrape ข้อมูลผู้เล่น (อ้างอิง nationality และ team)
	// log.Println("Scraping Players...")
	// err = scraper.ScrapePlayers(db)
	// if err != nil {
	// 	log.Printf("Error scraping players: %v", err)
	// } else {
	// 	log.Println("Players scraping completed.")
	// }

	// Scrape ข้อมูลตารางคะแนน (อ้างอิง league และ team)
	// log.Println("Scraping Standings...")
	// err = scraper.ScrapeStandings(db)
	// if err != nil {
	// 	log.Printf("Error scraping standings: %v", err)
	// } else {
	// 	log.Println("Standings scraping completed.")
	// }

	// Scrape ข้อมูลแมตช์ (อ้างอิง league, home_team, away_team, channel)
	log.Println("Scraping Matches (Thaileague, Cup, Playoff)...")
	// err = scraper.ScrapeThaileagueMatches(db, "all") // ถ้าต้องการเรียกทั้งหมด
	err = scraper.ScrapeThaileagueMatches(db, "t1") // เรียกเฉพาะ T1
	if err != nil {
		log.Printf("Error scraping Thaileague matches: %v", err)
	}
	err = scraper.ScrapeBallthaiCupMatches(db)
	// if err != nil {
	// 	log.Printf("Error scraping Ballthai Cup matches: %v", err)
	// }
	// err = scraper.ScrapeThaileaguePlayoffMatches(db)
	// if err != nil {
	// 	log.Printf("Error scraping Thaileague Playoff matches: %v", err)
	// }
	log.Println("Match scraping initiated. (Check logs for success/failure)")

	log.Println("Data scraping process finished.")
}
