package database

import (
	"database/sql"
	"fmt"
	"log"

	"go-ballthai-scraper/models" // Ensure this module name matches your go.mod
)

// InsertOrUpdateStadium inserts or updates a stadium record in the database.
func InsertOrUpdateStadium(db *sql.DB, stadium models.StadiumDB) error {
	var existingID int
	query := "SELECT id FROM stadiums WHERE name = ?"
	err := db.QueryRow(query, stadium.Name).Scan(&existingID)

	if err == sql.ErrNoRows {
		// Stadium does not exist, insert a new one
		insertQuery := `
            INSERT INTO stadiums (
                stadium_ref_id, name, short_name, name_en, short_name_en, 
                year_established, capacity, latitude, longitude, photo_url, 
                country_name, country_code, team_id
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        `
		_, err = db.Exec(
			insertQuery,
			stadium.StadiumRefID,
			stadium.Name,
			stadium.ShortName,
			stadium.NameEN,
			stadium.ShortNameEN,
			stadium.YearEstablished,
			stadium.Capacity,
			stadium.Latitude,
			stadium.Longitude,
			stadium.PhotoURL,
			stadium.CountryName,
			stadium.CountryCode,
			stadium.TeamID,
		)
		if err != nil {
			return fmt.Errorf("failed to insert stadium %s: %w", stadium.Name, err)
		}
		log.Printf("Inserted new stadium: %s", stadium.Name)
	} else if err != nil {
		return fmt.Errorf("error checking existing stadium: %w", err)
	} else {
		// Stadium exists, update the record
		updateQuery := `
            UPDATE stadiums SET
                stadium_ref_id = ?, short_name = ?, name_en = ?, short_name_en = ?, 
                year_established = ?, capacity = ?, latitude = ?, longitude = ?, 
                photo_url = ?, country_name = ?, country_code = ?, team_id = ?
            WHERE id = ?
        `
		_, err = db.Exec(
			updateQuery,
			stadium.StadiumRefID,
			stadium.ShortName,
			stadium.NameEN,
			stadium.ShortNameEN,
			stadium.YearEstablished,
			stadium.Capacity,
			stadium.Latitude,
			stadium.Longitude,
			stadium.PhotoURL,
			stadium.CountryName,
			stadium.CountryCode,
			stadium.TeamID,
			existingID,
		)
		if err != nil {
			return fmt.Errorf("failed to update stadium %s (ID: %d): %w", stadium.Name, existingID, err)
		}
		log.Printf("Updated stadium: %s (ID: %d)", stadium.Name, existingID)
	}
	return nil
}
