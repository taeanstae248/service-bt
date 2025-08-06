package database

import (
	"database/sql"
	"fmt"
	"log"
	// เพิ่ม import ที่จำเป็นอื่นๆ ที่นี่ เช่น "time" หากมีการใช้ Date/Time
	
	"go-ballthai-scraper/models" // แก้ไข: ตรวจสอบให้แน่ใจว่า import เฉพาะแพ็กเกจ models
)

// GetTeamIDByThaiName retrieves the team ID by its Thai name.
// If the team does not exist, it inserts a new team record and returns its ID.
func GetTeamIDByThaiName(db *sql.DB, nameTH string, logoURL string) (int, error) {
    var teamID int
    query := "SELECT id FROM teams WHERE name_th = ?"
    err := db.QueryRow(query, nameTH).Scan(&teamID)

    if err == sql.ErrNoRows {
        // Team does not exist, insert a new one
        insertQuery := `
            INSERT INTO teams (name_th, logo_url) VALUES (?, ?)
        `
        res, err := db.Exec(insertQuery, nameTH, logoURL)
        if err != nil {
            return 0, fmt.Errorf("failed to insert new team %s: %w", nameTH, err)
        }
        id, err := res.LastInsertId()
        if err != nil {
            return 0, fmt.Errorf("failed to get last insert ID for team %s: %w", nameTH, err)
        }
        log.Printf("Inserted new team: %s (ID: %d)", nameTH, id)
        return int(id), nil
    } else if err != nil {
        return 0, fmt.Errorf("error checking existing team %s: %w", nameTH, err)
    }
    // Team exists, return its ID
    return teamID, nil
}

// InsertOrUpdateTeam inserts or updates a team record in the database.
// This function assumes TeamDB struct is defined in models package.
func InsertOrUpdateTeam(db *sql.DB, team models.TeamDB) error {
    var existingID int
    query := "SELECT id FROM teams WHERE name_th = ?"
    err := db.QueryRow(query, team.NameTH).Scan(&existingID)

    if err == sql.ErrNoRows {
        // Team does not exist, insert a new one
        insertQuery := `
            INSERT INTO teams (
                team_ref_id, name_th, name_en, logo_url, team_post_ballthai, 
                website, shop, stadium_id, league_id
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
        `
        _, err = db.Exec(
            insertQuery,
            team.TeamRefID,
            team.NameTH,
            team.NameEN,
            team.LogoURL,
            team.TeamPostBallthai,
            team.Website,
            team.Shop,
            team.StadiumID,
            team.LeagueID,
        )
        if err != nil {
            return fmt.Errorf("failed to insert team %s: %w", team.NameTH, err)
        }
        log.Printf("Inserted new team: %s", team.NameTH)
    } else if err != nil {
        return fmt.Errorf("error checking existing team: %w", err)
    } else {
        // Team exists, update the record
        updateQuery := `
            UPDATE teams SET
                team_ref_id = ?, name_en = ?, logo_url = ?, team_post_ballthai = ?,
                website = ?, shop = ?, stadium_id = ?, league_id = ?
            WHERE id = ?
        `
        _, err = db.Exec(
            updateQuery,
            team.TeamRefID,
            team.NameEN,
            team.LogoURL,
            team.TeamPostBallthai,
            team.Website,
            team.Shop,
            team.StadiumID,
            team.LeagueID,
            existingID,
        )
        if err != nil {
            return fmt.Errorf("failed to update team %s (ID: %d): %w", team.NameTH, existingID, err)
        }
        log.Printf("Updated team: %s (ID: %d)", team.NameTH, existingID)
    }
    return nil
}
