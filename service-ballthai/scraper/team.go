
package scraper

import (
   "database/sql"
   "encoding/json"
   "fmt"
   "io"
   "log"
   "net/http"
   "os"
   "path"
   "strings"

   "go-ballthai-scraper/models"
   "go-ballthai-scraper/database"
)

// SaveTeamsAndLogosByLeagueID ดึงทีมจาก API, บันทึกลง DB, ดาวน์โหลดโลโก้
func SaveTeamsAndLogosByLeagueID(db *sql.DB, leagueID string) (int, error) {
   teams, err := FetchTeamsByLeagueID(leagueID)
   if err != nil {
	   return 0, err
   }
   imported := 0
   for _, team := range teams {
	   var logoPath string
	   baseName := team.NameEN
	   if baseName == "" {
		   baseName = slugify(team.Name)
	   }
	   if team.Logo != "" {
		   ext := path.Ext(team.Logo)
		   if ext == "" {
			   ext = ".png"
		   }
		   fileName := sanitizeFileName(baseName) + ext
		   logoPath = path.Join("/img/teams", fileName)
		   if err := downloadLogoToFolder(team.Logo, baseName); err != nil {
			   log.Printf("Download logo failed for %s: %v", baseName, err)
			   logoPath = "" // ถ้าดาวน์โหลดไม่สำเร็จ
		   }
	   }
	   teamDB := models.TeamDB{
		   NameTH:   team.Name,
		   NameEN:   sql.NullString{String: team.NameEN, Valid: team.NameEN != ""},
		   LogoURL:  sql.NullString{String: logoPath, Valid: logoPath != ""},
		   Website:  sql.NullString{String: team.Website, Valid: team.Website != ""},
		   Shop:     sql.NullString{String: team.Shop, Valid: team.Shop != ""},
	   }
	   if err := database.InsertOrUpdateTeam(db, teamDB); err == nil {
		   imported++
	   }
   }
   return imported, nil
}

// slugify แปลงชื่อไทยเป็น slug ภาษาอังกฤษ (a-z, 0-9, -)

// downloadLogoToFolder ดาวน์โหลดโลโก้แล้วบันทึกลง img/teams/
func downloadLogoToFolder(logoURL, teamName string) error {
   resp, err := http.Get(logoURL)
   if err != nil {
	   return err
   }
   defer resp.Body.Close()
   if resp.StatusCode != http.StatusOK {
	   return fmt.Errorf("logo url returned status %d", resp.StatusCode)
   }
   ext := path.Ext(logoURL)
   if ext == "" {
	   ext = ".png"
   }
   fileName := sanitizeFileName(teamName) + ext
   outPath := path.Join("img/teams", fileName)
   if err := os.MkdirAll("img/teams", 0755); err != nil {
	   return err
   }
   out, err := os.Create(outPath)
   if err != nil {
	   return err
   }
   defer out.Close()
   _, err = io.Copy(out, resp.Body)
   return err
}

// slugify แปลงชื่อไทยเป็น slug ภาษาอังกฤษ (a-z, 0-9, -)
func slugify(s string) string {
   s = strings.ToLower(s)
   // แทนที่อักขระที่ไม่ใช่ a-z, 0-9 ด้วย -
   var b strings.Builder
   for _, r := range s {
	   if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
		   b.WriteRune(r)
	   } else {
		   b.WriteRune('-')
	   }
   }
   // ลบ - ซ้ำกันและขอบ
   slug := b.String()
		slug = strings.ReplaceAll(slug, "--", "-")
		slug = strings.Trim(slug, "-")
		return slug
	}

// sanitizeFileName แปลงชื่อทีมให้เป็นชื่อไฟล์ที่ปลอดภัย
func sanitizeFileName(name string) string {
   invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", " "}
   for _, c := range invalid {
	   name = strings.ReplaceAll(name, c, "_")
   }
   return name
}



const idPostAPIURL = "https://serviceseoball.com/api/id_post.php"

// UpdateTeamPostBallthai fetches team post IDs from an external API
// and updates the local database.
func UpdateTeamPostBallthai(db *sql.DB) error {
	log.Println("Fetching team post IDs from API...")

	// 1. Fetch data from the API
	resp, err := http.Get(idPostAPIURL)
	if err != nil {
		return fmt.Errorf("failed to get data from API %s: %w", idPostAPIURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("api request failed with status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// 2. Unmarshal JSON response
	var apiResponse models.TeamPostAPIResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}

	if len(apiResponse.Teams) == 0 {
		log.Println("No teams found in the API response. Nothing to update.")
		return nil
	}

	log.Printf("Found %d teams in API response. Starting database update...", len(apiResponse.Teams))

	// 3. Prepare SQL statement
	stmt, err := db.Prepare("UPDATE teams SET team_post_ballthai = ? WHERE name_th = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare update statement: %w", err)
	}
	defer stmt.Close()

	// 4. Iterate and update database
	updatedCount := 0
	for _, team := range apiResponse.Teams {
		if team.NameTh == "" || team.PostBallthaiID == "" {
			log.Printf("Skipping team with empty name or ID: %+v", team)
			continue
		}

		res, err := stmt.Exec(team.PostBallthaiID, team.NameTh)
		if err != nil {
			log.Printf("Error updating team '%s': %v", team.NameTh, err)
			continue // Continue to the next team
		}

		rowsAffected, _ := res.RowsAffected()
		if rowsAffected > 0 {
			updatedCount++
			log.Printf("Updated team_post_ballthai for team: %s", team.NameTh)
		}
	}

	log.Printf("Team post ID update process finished. Total teams updated: %d", updatedCount)
	return nil
}

// FetchTeamsByLeagueID ดึงข้อมูลทีมจาก API ตาม league id ที่ส่งเข้าไป
func FetchTeamsByLeagueID(leagueID string) ([]models.TeamAPI, error) {
	url := "https://competition.tl.prod.c0d1um.io/thaileague/api/tournament-team-dropdown-public/?tournament=" + leagueID
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch teams: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var teams []models.TeamAPI
	if err := json.Unmarshal(body, &teams); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return teams, nil
}
