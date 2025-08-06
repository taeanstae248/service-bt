package database

import (
	"database/sql"
	"fmt"
	"log"
	
	"go-ballthai-scraper/models" // Ensure this module name matches your go.mod
)

// InsertOrUpdateCoach inserts or updates a coach record in the database.
func InsertOrUpdateCoach(db *sql.DB, coach models.CoachDB) error {
    // Check if coach_ref_id already exists
    var existingID int
    query := "SELECT id FROM coaches WHERE coach_ref_id = ?"
    err := db.QueryRow(query, coach.CoachRefID).Scan(&existingID)

    if err == sql.ErrNoRows {
        // Coach does not exist, insert a new one
        insertQuery := `
            INSERT INTO coaches (
                coach_ref_id, name, birthday, team_id, nationality_id, photo_url
            ) VALUES (?, ?, ?, ?, ?, ?)
        `
        _, err = db.Exec(
            insertQuery,
            coach.CoachRefID,
            coach.Name,
            coach.Birthday,
            coach.TeamID,
            coach.NationalityID,
            coach.PhotoURL,
        )
        if err != nil {
            return fmt.Errorf("failed to insert coach %s: %w", coach.Name, err)
        }
        log.Printf("Inserted new coach: %s", coach.Name)
    } else if err != nil {
        return fmt.Errorf("error checking existing coach: %w", err)
    } else {
        // Coach exists, update the record
        updateQuery := `
            UPDATE coaches SET
                name = ?, birthday = ?, team_id = ?, nationality_id = ?, photo_url = ?
            WHERE id = ?
        `
        _, err = db.Exec(
            updateQuery,
            coach.Name,
            coach.Birthday,
            coach.TeamID,
            coach.NationalityID,
            coach.PhotoURL,
            existingID,
        )
        if err != nil {
            return fmt.Errorf("failed to update coach %s (ID: %d): %w", coach.Name, existingID, err)
        }
        log.Printf("Updated coach: %s (ID: %d)", coach.Name, existingID)
    }
    return nil
}
