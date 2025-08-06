package database

import (
	"database/sql"
	"fmt"
	"log"
	"time" // สำหรับ time.Time ใน sql.NullTime

	_ "github.com/go-sql-driver/mysql" // Driver สำหรับ MySQL
)

var DB *sql.DB // ตัวแปร Global สำหรับเก็บ Connection ฐานข้อมูล

// InitDB initializes the database connection
func InitDB(connStr string) error {
	var err error
	DB, err = sql.Open("mysql", connStr)
	if err != nil {
		return fmt.Errorf("error opening database connection: %w", err)
	}

	// ตั้งค่า Pool ของ Connection
	DB.SetMaxOpenConns(25)                 // จำนวน Connection สูงสุดที่เปิดพร้อมกัน
	DB.SetMaxIdleConns(25)                 // จำนวน Connection ที่จะเก็บไว้ใน Pool
	DB.SetConnMaxLifetime(5 * time.Minute) // ระยะเวลาที่ Connection จะอยู่ใน Pool ก่อนถูกปิด

	// ตรวจสอบว่าเชื่อมต่อฐานข้อมูลได้จริงหรือไม่
	if err = DB.Ping(); err != nil {
		DB.Close() // ปิด Connection ถ้า Ping ไม่ผ่าน
		return fmt.Errorf("error connecting to the database: %w", err)
	}
	log.Println("Database connection successful!")
	return nil
}

// GetLastInsertID is a helper function to get the last inserted ID
// (Note: In Go's database/sql, LastInsertId() is part of sql.Result from db.Exec)
// This function is generally not needed as a standalone, but for PHP's nextNumber equivalent
func GetLastInsertID(result sql.Result) (int64, error) {
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}
	return id, nil
}
