package database

import (
	"database/sql"
	"fmt"
	"log"
)

var DB *sql.DB // Global variable for the database connection

// InitDB initializes the database connection.
func InitDB(connStr string) error {
	var err error
	DB, err = sql.Open("mysql", connStr)
	if err != nil {
		return fmt.Errorf("error opening database connection: %w", err)
	}

	// Ping the database to verify the connection
	if err = DB.Ping(); err != nil {
		DB.Close() // Close the connection if ping fails
		return fmt.Errorf("error connecting to the database: %w", err)
	}
	log.Println("Database connection successful!")
	return nil
}
