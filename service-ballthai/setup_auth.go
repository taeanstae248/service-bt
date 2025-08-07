package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get database configuration
	dbUser := os.Getenv("DB_USERNAME")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	if dbUser == "" || dbHost == "" || dbPort == "" || dbName == "" {
		log.Fatalf("Missing database environment variables")
	}

	connStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)

	// Connect to database
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Connected to database successfully!")

	// Create users table
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INT PRIMARY KEY AUTO_INCREMENT,
		username VARCHAR(50) NOT NULL UNIQUE,
		email VARCHAR(100) NOT NULL UNIQUE,
		password_hash VARCHAR(255) NOT NULL,
		full_name VARCHAR(100),
		role ENUM('admin', 'editor', 'viewer') DEFAULT 'viewer',
		is_active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		last_login TIMESTAMP NULL
	)`

	_, err = db.Exec(createUsersTable)
	if err != nil {
		log.Fatalf("Failed to create users table: %v", err)
	}
	log.Println("Users table created successfully!")

	// Create sessions table
	createSessionsTable := `
	CREATE TABLE IF NOT EXISTS sessions (
		id VARCHAR(128) PRIMARY KEY,
		user_id INT NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	)`

	_, err = db.Exec(createSessionsTable)
	if err != nil {
		log.Fatalf("Failed to create sessions table: %v", err)
	}
	log.Println("Sessions table created successfully!")

	// Generate password hash for admin123
	password := "admin123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to generate password hash: %v", err)
	}

	// Insert admin user
	insertAdmin := `
	INSERT INTO users (username, email, password_hash, full_name, role)
	VALUES (?, ?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE username = username
	`

	_, err = db.Exec(insertAdmin, "admin", "admin@ballthai.com", string(hash), "ผู้ดูแลระบบ", "admin")
	if err != nil {
		log.Fatalf("Failed to insert admin user: %v", err)
	}
	log.Println("Admin user created successfully!")
	log.Printf("Admin credentials - Username: admin, Password: %s", password)
}
