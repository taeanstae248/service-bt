package database

import (
	"database/sql"
	"fmt"
	"log"
	// "strings" // ลบออก: ไม่ได้ใช้ strings ในไฟล์นี้แล้ว

	"go-ballthai-scraper/models" // ตรวจสอบให้แน่ใจว่าชื่อโมดูลตรงกับ go.mod ของคุณ
)

// GetTeamIDByThaiName checks if team exists by Thai name, inserts if not, and returns ID
func GetTeamIDByThaiName(db *sql.DB, teamNameThai, teamLogoURL string) (int, error) {
	var teamID int
	// ใช้ REPLACE (name_th, ' ', '') = REPLACE (?, ' ', '') เพื่อเทียบชื่อโดยไม่สนใจช่องว่าง
	query := "SELECT id FROM teams WHERE REPLACE(name_th, ' ', '') = REPLACE(?, ' ', '')"
	err := db.QueryRow(query, teamNameThai).Scan(&teamID)

	if err == sql.ErrNoRows {
		// ถ้าไม่พบทีม ให้เพิ่มทีมใหม่
		insertQuery := `
			INSERT INTO teams (name_th, logo_url)
			VALUES (?, ?)
		`
		result, err := db.Exec(insertQuery, teamNameThai, sql.NullString{String: teamLogoURL, Valid: teamLogoURL != ""})
		if err != nil {
			return 0, fmt.Errorf("failed to insert new team: %w", err)
		}
		newID, err := result.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("failed to get last insert ID for team: %w", err)
		}
		log.Printf("Inserted new team: %s (ID: %d)", teamNameThai, newID)
		return int(newID), nil
	} else if err != nil {
		return 0, fmt.Errorf("failed to query team by name: %w", err)
	}
	// ถ้าพบทีมแล้ว ให้อัปเดตโลโก้ (ถ้าจำเป็น)
	updateLogoQuery := "UPDATE teams SET logo_url = ? WHERE id = ?"
	_, err = db.Exec(updateLogoQuery, sql.NullString{String: teamLogoURL, Valid: teamLogoURL != ""}, teamID)
	if err != nil {
		log.Printf("Warning: Failed to update team logo for ID %d: %v", teamID, err)
	}

	log.Printf("Found existing team: %s (ID: %d)", teamNameThai, teamID)
	return teamID, nil
}

// InsertOrUpdateTeam inserts or updates a team record in the database (ใช้ name_th แทน team_ref_id)
func InsertOrUpdateTeam(db *sql.DB, team models.TeamDB) error {
	var existingTeamID int
	query := "SELECT id FROM teams WHERE name_th = ?"
	err := db.QueryRow(query, team.NameTH).Scan(&existingTeamID)

	if err == sql.ErrNoRows {
		// Insert new team
		insertQuery := `
			INSERT INTO teams (
				name_th, name_en, logo_url,
				team_post_ballthai, website, shop, stadium_id
			) VALUES (?, ?, ?, ?, ?, ?, ?)
		`
		_, err := db.Exec(insertQuery,
			team.NameTH, team.NameEN, team.LogoURL,
			team.TeamPostBallthai, team.Website, team.Shop, team.StadiumID,
		)
		if err != nil {
			return fmt.Errorf("failed to insert team %s: %w", team.NameTH, err)
		}
		log.Printf("Inserted new team: %s", team.NameTH)
	} else if err != nil {
		return fmt.Errorf("failed to query existing team %s: %w", team.NameTH, err)
	} else {
		// Update existing team
		updateQuery := `
			UPDATE teams SET
				name_en = ?, logo_url = ?,
				team_post_ballthai = ?, website = ?, shop = ?, stadium_id = ?
			WHERE id = ?
		`
		_, err := db.Exec(updateQuery,
			team.NameEN, team.LogoURL,
			team.TeamPostBallthai, team.Website, team.Shop, team.StadiumID,
			existingTeamID,
		)
		if err != nil {
			return fmt.Errorf("failed to update team %s: %w", team.NameTH, err)
		}
		log.Printf("Updated existing team: %s (ID: %d)", team.NameTH, existingTeamID)
	}
	return nil
}
