package database

import (
	"database/sql"
	"fmt"
	"log"
)

// GetNationalityID checks if nationality exists by code, inserts if not, and returns ID
func GetNationalityID(db *sql.DB, code, name string) (int, error) {
	var nationalityID int

	// Prefer lookup by name when a name is provided. Names are generally
	// more specific than a short code returned by some APIs, and some APIs
	// unfortunately return the same code for many players.
	if name != "" {
		queryByName := "SELECT id FROM nationalities WHERE REPLACE(name, ' ', '') = REPLACE(?, ' ', '')"
		err := db.QueryRow(queryByName, name).Scan(&nationalityID)
		if err == nil {
			log.Printf("Found existing nationality by name: %s (ID: %d)", name, nationalityID)
			return nationalityID, nil
		}
		if err != sql.ErrNoRows {
			return 0, fmt.Errorf("failed to query nationality by name %s: %w", name, err)
		}
		// not found by name, fallthrough to try code (if present)
	}

	// If code is provided, try lookup by code next
	if code != "" {
		query := "SELECT id FROM nationalities WHERE code = ?"
		err := db.QueryRow(query, code).Scan(&nationalityID)
		if err == nil {
			log.Printf("Found existing nationality by code: %s (code=%s ID=%d)", name, code, nationalityID)
			return nationalityID, nil
		}
		if err != sql.ErrNoRows {
			return 0, fmt.Errorf("failed to query nationality by code %s: %w", code, err)
		}
	}

	// Insert new nationality. If code is empty, insert NULL for code.
	var codeVal interface{} = nil
	if code != "" {
		codeVal = code
	}
	insertQuery := `INSERT INTO nationalities (code, name) VALUES (?, ?)`
	result, err := db.Exec(insertQuery, codeVal, name)
	if err != nil {
		return 0, fmt.Errorf("failed to insert new nationality %s: %w", name, err)
	}
	newID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID for nationality %s: %w", name, err)
	}
	log.Printf("Inserted new nationality: %s (ID: %d) code=%v", name, newID, codeVal)
	return int(newID), nil
}

// GetChannelID checks if channel exists by name, inserts if not, and returns ID
func GetChannelID(db *sql.DB, name, logoURL, channelType string) (int, error) {
       var channelID int
       var oldLogo sql.NullString
       query := "SELECT id, logo_url FROM channels WHERE REPLACE(name, ' ', '') = REPLACE(?, ' ', '')"
       err := db.QueryRow(query, name).Scan(&channelID, &oldLogo)

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

       // Update logo_url ถ้าเปลี่ยน
       if logoURL != "" && (!oldLogo.Valid || oldLogo.String != logoURL) {
	       updateQuery := `UPDATE channels SET logo_url = ? WHERE id = ?`
	       _, err := db.Exec(updateQuery, logoURL, channelID)
	       if err != nil {
		       log.Printf("Warning: Failed to update channel logo for ID %d: %v", channelID, err)
	       } else {
		       log.Printf("Updated channel logo for ID %d: %s", channelID, logoURL)
	       }
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

// GetStageID checks if stage exists by name, inserts if not, and returns ID
func GetStageID(db *sql.DB, stageName string, leagueID int) (int, error) {
	var stageID int
	query := "SELECT id FROM stage WHERE REPLACE(stage_name, ' ', '') = REPLACE(?, ' ', '')"
	err := db.QueryRow(query, stageName).Scan(&stageID)

	if err == sql.ErrNoRows {
		insertQuery := `INSERT INTO stage (stage_name) VALUES (?)`
		result, err := db.Exec(insertQuery, stageName)
		if err != nil {
			return 0, fmt.Errorf("failed to insert new stage %s: %w", stageName, err)
		}
		newID, err := result.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("failed to get last insert ID for stage %s: %w", stageName, err)
		}
		log.Printf("Inserted new stage: %s (ID: %d)", stageName, newID)
		return int(newID), nil
	} else if err != nil {
		return 0, fmt.Errorf("failed to query stage by name %s: %w", stageName, err)
	}
	log.Printf("Found existing stage: %s (ID: %d)", stageName, stageID)
	return stageID, nil
}
