package database

import (
	"database/sql"
	"fmt"
	"log"
)

// GetNationalityID checks if nationality exists by code, inserts if not, and returns ID
func GetNationalityID(db *sql.DB, code, name string) (int, error) {
	var nationalityID int
	query := "SELECT id FROM nationalities WHERE code = ?"
	err := db.QueryRow(query, code).Scan(&nationalityID)

	if err == sql.ErrNoRows {
		// Insert new nationality
		insertQuery := `INSERT INTO nationalities (code, name) VALUES (?, ?)`
		result, err := db.Exec(insertQuery, code, name)
		if err != nil {
			return 0, fmt.Errorf("failed to insert new nationality %s: %w", name, err)
		}
		newID, err := result.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("failed to get last insert ID for nationality %s: %w", name, err)
		}
		log.Printf("Inserted new nationality: %s (ID: %d)", name, newID)
		return int(newID), nil
	} else if err != nil {
		return 0, fmt.Errorf("failed to query nationality by code %s: %w", code, err)
	}
	log.Printf("Found existing nationality: %s (ID: %d)", name, nationalityID)
	return nationalityID, nil
}

// GetChannelID checks if channel exists by name, inserts if not, and returns ID
func GetChannelID(db *sql.DB, name, logoURL, channelType string) (int, error) {
	var channelID int
	query := "SELECT id FROM channels WHERE REPLACE(name, ' ', '') = REPLACE(?, ' ', '')"
	err := db.QueryRow(query, name).Scan(&channelID)

	if err == sql.ErrNoRows {
		// Insert new channel
		insertQuery := `INSERT INTO channels (name, logo_url, type) VALUES (?, ?, ?)`
		result, err := db.Exec(insertQuery, name, logoURL, channelType)
		if err != nil {
			return 0, fmt.Errorf("failed to insert new channel %s: %w", name, err)
		}
		newID, err := result.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("failed to get last insert ID for channel %s: %w", name, err)
		}
		log.Printf("Inserted new channel: %s (ID: %d)", name, newID)
		return int(newID), nil
	} else if err != nil {
		return 0, fmt.Errorf("failed to query channel by name %s: %w", name, err)
	}
	log.Printf("Found existing channel: %s (ID: %d)", name, channelID)
	return channelID, nil
}

// GetLeagueID checks if league exists by name, inserts if not, and returns ID
func GetLeagueID(db *sql.DB, leagueName, leagueNameThai string) (int, error) {
	var leagueID int
	query := "SELECT id FROM leagues WHERE REPLACE(name, ' ', '') = REPLACE(?, ' ', '') OR REPLACE(name, ' ', '') = REPLACE(?, ' ', '')"
	err := db.QueryRow(query, leagueName, leagueNameThai).Scan(&leagueID)

	if err == sql.ErrNoRows {
		insertQuery := `INSERT INTO leagues (name) VALUES (?)`
		result, err := db.Exec(insertQuery, leagueNameThai) // Use Thai name as primary name
		if err != nil {
			return 0, fmt.Errorf("failed to insert new league %s: %w", leagueNameThai, err)
		}
		newID, err := result.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("failed to get last insert ID for league %s: %w", leagueNameThai, err)
		}
		log.Printf("Inserted new league: %s (ID: %d)", leagueNameThai, newID)
		return int(newID), nil
	} else if err != nil {
		return 0, fmt.Errorf("failed to query league by name %s: %w", leagueNameThai, err)
	}
	log.Printf("Found existing league: %s (ID: %d)", leagueNameThai, leagueID)
	return leagueID, nil
}
