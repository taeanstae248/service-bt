package database

import (
	"database/sql"
	"fmt"
	"log"

	"go-scraper/models" // ตรวจสอบให้แน่ใจว่าชื่อโมดูลตรงกับ go.mod ของคุณ
)

// GetStadiumID checks if stadium exists by ref_id, inserts if not, and returns ID
func GetStadiumID(db *sql.DB, stadiumRefID int, stadiumName, stadiumNameEN, photoURL string) (int, error) {
	var stadiumID int
	query := "SELECT id FROM stadiums WHERE stadium_ref_id = ?"
	err := db.QueryRow(query, stadiumRefID).Scan(&stadiumID)

	if err == sql.ErrNoRows {
		// Insert new stadium
		insertQuery := `
			INSERT INTO stadiums (stadium_ref_id, name, name_en, photo_url)
			VALUES (?, ?, ?, ?)
		`
		result, err := db.Exec(insertQuery, stadiumRefID, stadiumName, sql.NullString{String: stadiumNameEN, Valid: stadiumNameEN != ""}, sql.NullString{String: photoURL, Valid: photoURL != ""})
		if err != nil {
			return 0, fmt.Errorf("failed to insert new stadium %s: %w", stadiumName, err)
		}
		newID, err := result.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("failed to get last insert ID for stadium %s: %w", stadiumName, err)
		}
		log.Printf("Inserted new stadium: %s (ID: %d)", stadiumName, newID)
		return int(newID), nil
	} else if err != nil {
		return 0, fmt.Errorf("failed to query stadium by ref ID %d: %w", stadiumRefID, err)
	}
	log.Printf("Found existing stadium: %s (ID: %d)", stadiumName, stadiumID)
	return stadiumID, nil
}

// InsertOrUpdateStadium inserts or updates a stadium record in the database
func InsertOrUpdateStadium(db *sql.DB, stadium models.StadiumDB) error {
	var existingStadiumID int
	query := "SELECT id FROM stadiums WHERE stadium_ref_id = ?"
	err := db.QueryRow(query, stadium.StadiumRefID).Scan(&existingStadiumID)

	if err == sql.ErrNoRows {
		// Insert new stadium
		insertQuery := `
			INSERT INTO stadiums (
				stadium_ref_id, name, name_en, photo_url,
				year_established, country_name, country_code, capacity, latitude, longitude, team_id
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		_, err := db.Exec(insertQuery,
			stadium.StadiumRefID, stadium.Name, stadium.NameEN, stadium.PhotoURL,
			stadium.YearEstablished, stadium.CountryName, stadium.CountryCode, stadium.Capacity, stadium.Latitude, stadium.Longitude, stadium.TeamID,
		)
		if err != nil {
			return fmt.Errorf("failed to insert stadium %s: %w", stadium.Name, err)
		}
		log.Printf("Inserted new stadium: %s", stadium.Name)
	} else if err != nil {
		return fmt.Errorf("failed to query existing stadium %d: %w", stadium.StadiumRefID, err)
	} else {
		// Update existing stadium
		updateQuery := `
			UPDATE stadiums SET
				name = ?, name_en = ?, photo_url = ?,
				year_established = ?, country_name = ?, country_code = ?, capacity = ?, latitude = ?, longitude = ?, team_id = ?
			WHERE stadium_ref_id = ?
		`
		_, err := db.Exec(updateQuery,
			stadium.Name, stadium.NameEN, stadium.PhotoURL,
			stadium.YearEstablished, stadium.CountryName, stadium.CountryCode, stadium.Capacity, stadium.Latitude, stadium.Longitude, stadium.TeamID,
			stadium.StadiumRefID,
		)
		if err != nil {
			return fmt.Errorf("failed to update stadium %d: %w", stadium.StadiumRefID, err)
		}
		log.Printf("Updated existing stadium: %s (ID: %d)", stadium.Name, existingStadiumID)
	}
	return nil
}
