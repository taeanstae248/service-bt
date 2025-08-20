package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"go-ballthai-scraper/models" // ตรวจสอบให้แน่ใจว่าชื่อโมดูลตรงกับ go.mod ของคุณ
)

func init() {
	// สร้างโฟลเดอร์ img/teams ถ้ายังไม่มี
	os.MkdirAll("img/teams", os.ModePerm)
}

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
       // ถ้าพบทีมแล้ว ให้อัปเดตโลโก้ เฉพาะเมื่อ teamLogoURL ไม่ว่าง
       if teamLogoURL != "" {
	       updateLogoQuery := "UPDATE teams SET logo_url = ? WHERE id = ?"
	       _, err = db.Exec(updateLogoQuery, sql.NullString{String: teamLogoURL, Valid: true}, teamID)
	       if err != nil {
		       log.Printf("Warning: Failed to update team logo for ID %d: %v", teamID, err)
	       }
       }

	log.Printf("Found existing team: %s (ID: %d)", teamNameThai, teamID)
	return teamID, nil
}

// InsertOrUpdateTeam inserts or updates a team record in the database (ใช้ name_th แทน team_ref_id)
func InsertOrUpdateTeam(db *sql.DB, team models.TeamDB) error {
	var existingTeamID int
	query := "SELECT id FROM teams WHERE name_th = ?"
	err := db.QueryRow(query, team.NameTH).Scan(&existingTeamID)

		// รับ path local จาก team.LogoURL โดยตรง ไม่ดาวน์โหลดซ้ำ
		logoDBPath := team.LogoURL.String

       if err == sql.ErrNoRows {
	       // Insert new team
	       insertQuery := `
		       INSERT INTO teams (
			       name_th, name_en, logo_url,
			       team_post_ballthai, website, shop, stadium_id
		       ) VALUES (?, ?, ?, ?, ?, ?, ?)
	       `
	       _, err := db.Exec(insertQuery,
		       team.NameTH, team.NameEN, sql.NullString{String: logoDBPath, Valid: logoDBPath != ""},
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
	       // ถ้า logoDBPath ไม่ว่าง ให้ update logo_url ด้วย, ถ้าว่างให้ข้าม logo_url
	       if logoDBPath != "" {
		       updateQuery := `
			       UPDATE teams SET
				       name_en = ?, logo_url = ?,
				       team_post_ballthai = ?, website = ?, shop = ?, stadium_id = ?
			       WHERE id = ?
		       `
		       _, err := db.Exec(updateQuery,
			       team.NameEN, sql.NullString{String: logoDBPath, Valid: true},
			       team.TeamPostBallthai, team.Website, team.Shop, team.StadiumID,
			       existingTeamID,
		       )
		       if err != nil {
			       return fmt.Errorf("failed to update team %s: %w", team.NameTH, err)
		       }
		       log.Printf("Updated existing team: %s (ID: %d)", team.NameTH, existingTeamID)
	       } else {
		       updateQuery := `
			       UPDATE teams SET
				       name_en = ?,
				       team_post_ballthai = ?, website = ?, shop = ?, stadium_id = ?
			       WHERE id = ?
		       `
		       _, err := db.Exec(updateQuery,
			       team.NameEN,
			       team.TeamPostBallthai, team.Website, team.Shop, team.StadiumID,
			       existingTeamID,
		       )
		       if err != nil {
			       return fmt.Errorf("failed to update team %s: %w", team.NameTH, err)
		       }
		       log.Printf("Updated existing team (no logo change): %s (ID: %d)", team.NameTH, existingTeamID)
	       }
       }
	return nil
}

// ฟังก์ชันช่วย sanitize ชื่อไฟล์
func sanitizeFileName(name string) string {
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	return name
}
