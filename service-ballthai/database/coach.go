package database

import (
	"database/sql"
	"fmt"
	"log"

	"go-scraper/models" // ตรวจสอบให้แน่ใจว่าชื่อโมดูลตรงกับ go.mod ของคุณ
)

// InsertOrUpdateCoach inserts or updates a coach record in the database
func InsertOrUpdateCoach(db *sql.DB, coach models.CoachDB) error {
	var existingCoachID int
	query := "SELECT id FROM coaches WHERE coach_ref_id = ?"
	err := db.QueryRow(query, coach.CoachRefID).Scan(&existingCoachID)

	if err == sql.ErrNoRows {
		// Insert new coach
		insertQuery := `
			INSERT INTO coaches (
				coach_ref_id, name, birthday, team_id, nationality_id, photo_url
			) VALUES (?, ?, ?, ?, ?, ?)
		`
		_, err := db.Exec(insertQuery,
			coach.CoachRefID, coach.Name, coach.Birthday, coach.TeamID, coach.NationalityID, coach.PhotoURL,
		)
		if err != nil {
			return fmt.Errorf("failed to insert coach %s: %w", coach.Name, err)
		}
		log.Printf("Inserted new coach: %s", coach.Name)
	} else if err != nil {
		return fmt.Errorf("failed to query existing coach %d: %w", coach.CoachRefID, err)
	} else {
		// Update existing coach
		updateQuery := `
			UPDATE coaches SET
				name = ?, birthday = ?, team_id = ?, nationality_id = ?, photo_url = ?
			WHERE coach_ref_id = ?
		`
		_, err := db.Exec(updateQuery,
			coach.Name, coach.Birthday, coach.TeamID, coach.NationalityID, coach.PhotoURL,
			coach.CoachRefID,
		)
		if err != nil {
			return fmt.Errorf("failed to update coach %d: %w", coach.CoachRefID, err)
		}
		log.Printf("Updated existing coach: %s (ID: %d)", coach.Name, existingCoachID)
	}
	return nil
}
