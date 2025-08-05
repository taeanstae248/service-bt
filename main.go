package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    "time"

    _ "github.com/go-sql-driver/mysql"
    "github.com/joho/godotenv" // เพิ่ม import นี้

    "go-ballthai-scraper/database"
    "go-ballthai-scraper/scraper"
)

var db *sql.DB

func main() {
    // โหลดค่าจาก .env ไฟล์ก่อน
    err := godotenv.Load()
    if err != nil {
        log.Println("Error loading .env file, assuming environment variables are set directly.")
        // ไม่ต้อง Fatalf ที่นี่ เพราะอาจจะรันบน Server ที่ตั้งค่า env vars โดยตรง
    }

    // 1. ตั้งค่าการเชื่อมต่อฐานข้อมูล
    // ดึงค่าจาก Environment Variables
    dbUser := os.Getenv("DB_USERNAME")
    dbPass := os.Getenv("DB_PASSWORD")
    dbHost := os.Getenv("DB_HOST")
    dbPort := os.Getenv("DB_PORT")
    dbName := os.Getenv("DB_NAME")

    // ตรวจสอบว่าค่าที่จำเป็นถูกตั้งค่าหรือไม่
    if dbUser == "" || dbPass == "" || dbHost == "" || dbPort == "" || dbName == "" {
        log.Fatalf("Missing one or more database environment variables (DB_USERNAME, DB_PASSWORD, DB_HOST, DB_PORT, DB_NAME)")
    }

    connStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)

    var dbErr error // ใช้ชื่อตัวแปร err อื่นเพื่อไม่ให้ชนกับ err ของ godotenv.Load()
    db, dbErr = sql.Open("mysql", connStr)
    if dbErr != nil {
        log.Fatalf("Error opening database connection: %v", dbErr)
    }
    defer db.Close()

    dbErr = db.Ping()
    if dbErr != nil {
        log.Fatalf("Error connecting to the database: %v", dbErr)
    }
    log.Println("Successfully connected to the database!")

    // ... โค้ดส่วนที่เหลือเหมือนเดิม ...
}

// ... ฟังก์ชัน initDB (ถ้ายังไม่ได้ลบออก) ...