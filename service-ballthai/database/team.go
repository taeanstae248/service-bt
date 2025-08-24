package database

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

	// Normalize / download logo URL before inserting
	normalizedLogo := NormalizeLogoURL(teamLogoURL)

	if err == sql.ErrNoRows {
		// ถ้าไม่พบทีม ให้เพิ่มทีมใหม่
		insertQuery := `
			INSERT INTO teams (name_th, logo_url)
			VALUES (?, ?)
		`
		result, err := db.Exec(insertQuery, teamNameThai, sql.NullString{String: normalizedLogo, Valid: normalizedLogo != ""})
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
	   if normalizedLogo != "" {
		   updateLogoQuery := "UPDATE teams SET logo_url = ? WHERE id = ?"
		   _, err = db.Exec(updateLogoQuery, sql.NullString{String: normalizedLogo, Valid: true}, teamID)
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

	// Normalize / download logo path from team.LogoURL
	logoDBPath := NormalizeLogoURL(team.LogoURL.String)

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
		// Build SET clause dynamically so we don't overwrite fields that caller didn't provide.
		// Always update name_en. Only update logo_url when logoDBPath != "".
		// Only update team_post_ballthai when caller provided a valid value (Valid == true).
		setClauses := []string{"name_en = ?"}
		args := []interface{}{team.NameEN}

		if logoDBPath != "" {
			setClauses = append(setClauses, "logo_url = ?")
			args = append(args, sql.NullString{String: logoDBPath, Valid: true})
		}

		if team.TeamPostBallthai.Valid {
			setClauses = append(setClauses, "team_post_ballthai = ?")
			args = append(args, team.TeamPostBallthai)
		}

		// Common optional fields
		setClauses = append(setClauses, "website = ?", "shop = ?", "stadium_id = ?")
		args = append(args, team.Website, team.Shop, team.StadiumID)

		// Finalize query
		updateQuery := "UPDATE teams SET " + strings.Join(setClauses, ", ") + " WHERE id = ?"
		args = append(args, existingTeamID)

		_, err := db.Exec(updateQuery, args...)
		if err != nil {
			return fmt.Errorf("failed to update team %s: %w", team.NameTH, err)
		}

		if logoDBPath != "" {
			log.Printf("Updated existing team (logo updated): %s (ID: %d)", team.NameTH, existingTeamID)
		} else {
			log.Printf("Updated existing team: %s (ID: %d)", team.NameTH, existingTeamID)
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

// NormalizeLogoURL will convert external logo URLs into local server paths when possible.
// - If logo is empty, returns empty string.
// - If logo already points to a local path (starts with /img/ or img\\), return normalized local path.
// - If logo is an absolute external URL (http/https), attempt to download into ./img/teams and
//   return the server-local path (/img/teams/<filename>). If download fails, return original URL.
func NormalizeLogoURL(logo string) string {
	logo = strings.TrimSpace(logo)
	if logo == "" {
		return ""
	}

	// Already local path
	if strings.HasPrefix(logo, "/img/") || strings.HasPrefix(logo, "img\\") || strings.HasPrefix(logo, "img/") {
		// Normalize backslashes
		normalized := strings.ReplaceAll(logo, "\\", "/")
		return normalized
	}

	// If protocol-relative URL
	if strings.HasPrefix(logo, "//") {
		logo = "https:" + logo
	}

	// If it's an absolute URL, try download
	if strings.HasPrefix(logo, "http://") || strings.HasPrefix(logo, "https://") {
		// download to ./img/teams/
			filename := filepath.Base(logo)
		filename = sanitizeFileName(filename)
			destDir := filepath.Join(".", "img", "teams")
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return logo
		}
		destPath := filepath.Join(destDir, filename)
		// if file already exists, return server path
		if _, err := os.Stat(destPath); err == nil {
				return "/img/teams/" + filename
		}

		// Download
		resp, err := http.Get(logo)
		if err != nil {
			log.Printf("NormalizeLogoURL: download error for %s: %v", logo, err)
			return logo
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			log.Printf("NormalizeLogoURL: bad status %d for %s", resp.StatusCode, logo)
			return logo
		}

		out, err := os.Create(destPath)
		if err != nil {
			log.Printf("NormalizeLogoURL: create file error %v", err)
			return logo
		}
		defer out.Close()

		if _, err := io.Copy(out, resp.Body); err != nil {
			log.Printf("NormalizeLogoURL: save file error %v", err)
			return logo
		}
	log.Printf("NormalizeLogoURL: downloaded logo to %s", destPath)
	return "/img/teams/" + filename
	}

	// otherwise leave as-is
	return logo
}
