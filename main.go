package main

import (
    "database/sql"
    "fmt"
    "log"

    _ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func main() {
    // ... (โค้ดเชื่อมต่อฐานข้อมูล) ...

    // กำหนดคำสั่ง SQL เป็น string
    createTableSQL := `
    CREATE TABLE IF NOT EXISTS leagues (
        id INT PRIMARY KEY AUTO_INCREMENT,
        name VARCHAR(255) NOT NULL UNIQUE
    );

    CREATE TABLE IF NOT EXISTS teams (
        id INT PRIMARY KEY AUTO_INCREMENT,
        name_th VARCHAR(255) NOT NULL UNIQUE,
        name_en VARCHAR(255),
        logo_url VARCHAR(255)
    );

    -- ... เพิ่มตารางอื่นๆ ตรงนี้ ...
    `

    // สั่งรันคำสั่ง SQL
    _, err = db.Exec(createTableSQL)
    if err != nil {
        log.Fatalf("Failed to create tables: %v", err)
    }

    fmt.Println("Database tables created successfully!")
}