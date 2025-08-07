package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
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

	connStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)

	// Connect to database
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test query
	query := `
		SELECT id, username, email, full_name, role, is_active, created_at, updated_at, last_login
		FROM users 
		WHERE username = ? AND is_active = TRUE
	`

	rows, err := db.Query(query, "admin")
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var username, email, role string
		var fullName *string
		var isActive bool
		var createdAt, updatedAt string
		var lastLogin *string

		err := rows.Scan(&id, &username, &email, &fullName, &role, &isActive, &createdAt, &updatedAt, &lastLogin)
		if err != nil {
			log.Fatalf("Scan failed: %v", err)
		}

		fmt.Printf("Found user: ID=%d, Username=%s, Email=%s, FullName=%v, Role=%s, Active=%t\n",
			id, username, email, fullName, role, isActive)
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Rows error: %v", err)
	}
}
