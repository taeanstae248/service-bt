package database

import (
	"database/sql"
	"fmt"
	"log"

	// "go-ballthai-scraper/models" // ลบออก: ไม่ได้ใช้ models ในไฟล์นี้โดยตรง
)

// GetNationalityID retrieves the nationality ID by its code.
// If the nationality does not exist, it inserts a new nationality record and returns its ID.
func GetNationalityID(db *sql.DB, code string, name string) (int, error) {
	var nationalityID int
	query := "SELECT id FROM nationalities WHERE code = ?"
	err := db.QueryRow(query, code).Scan(&nationalityID)

	if err == sql.ErrNoRows {
		// Nationality does not exist, insert a new one
		insertQuery := `INSERT INTO nationalities (code, name) VALUES (?, ?)`
		res, err := db.Exec(insertQuery, code, name)
		if err != nil {
			return 0, fmt.Errorf("failed to insert new nationality %s: %w", name, err)
		}
		id, err := res.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("failed to get last insert ID for nationality %s: %w", name, err)
		}
		log.Printf("Inserted new nationality: %s (ID: %d)", name, id)
		return int(id), nil
	} else if err != nil {
		return 0, fmt.Errorf("error checking existing nationality %s: %w", name, err)
	}
	// Nationality exists, return its ID
	return nationalityID, nil
}

// GetChannelID retrieves the channel ID by its name and type.
// If the channel does not exist, it inserts a new channel record and returns its ID.
func GetChannelID(db *sql.DB, channelName string, logoURL string, channelType string) (int, error) {
	var channelID int
	query := "SELECT id FROM channels WHERE name = ? AND type = ?"
	err := db.QueryRow(query, channelName, channelType).Scan(&channelID)

	if err == sql.ErrNoRows {
		// Channel does not exist, insert a new one
		insertQuery := `INSERT INTO channels (name, logo_url, type) VALUES (?, ?, ?)`
		res, err := db.Exec(insertQuery, channelName, logoURL, channelType)
		if err != nil {
			return 0, fmt.Errorf("failed to insert new channel %s (%s): %w", channelName, channelType, err)
		}
		id, err := res.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("failed to get last insert ID for channel %s (%s): %w", channelName, channelType, err)
		}
		log.Printf("Inserted new channel: %s (%s) (ID: %d)", channelName, channelType, id)
		return int(id), nil
	} else if err != nil {
		return 0, fmt.Errorf("error checking existing channel %s (%s): %w", channelName, channelType, err)
	}
	// Channel exists, return its ID
	return channelID, nil
}
