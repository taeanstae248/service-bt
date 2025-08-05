package database

import (
	"database/sql"
	"log"
)

var DB *sql.DB // Global variable to hold the database connection

// InitDB initializes the database connection
func InitDB(connStr string) error {
	var err error
	DB, err = sql.Open("mysql", connStr)
	if err != nil {
		return err
	}

	// Test the connection
	err = DB.Ping()
	if err != nil {
		return err
	}

	log.Println("Database connection established successfully")
	return nil
}
