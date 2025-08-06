package scraper

import (
	"database/sql"
	"fmt"
	"log"

	"go-ballthai-scraper/database" // ตรวจสอบให้แน่ใจว่าชื่อโมดูลตรงกับ go.mod ของคุณ
	"go-ballthai-scraper/models"  // ตรวจสอบให้แน่ใจว่าชื่อโมดูลตรงกับ go.mod ของคุณ
)

// ScrapePlayers ดึงข้อมูลผู้เล่นจาก API และบันทึกลงฐานข้อมูล
// ฟังก์ชันนี้รวมตรรกะจากฟังก์ชัน PHP Scrape_R*_Player และ Player_Public
func ScrapePlayers(db *sql.DB) error {
	baseAPIURL := "https://competition.tl.prod.c0d1um.io/thaileague/api/player-public/all_players_search/"

	// กำหนดการกำหนดค่าทัวร์นาเมนต์/Team ID เฉพาะตามที่เห็นในไฟล์ PHP
	// นี่เป็นตัวอย่างที่ง่ายขึ้น คุณอาจต้องแมปทัวร์นาเมนต์กับลีก
	// และจัดการ Team ID เฉพาะตามตรรกะที่ฮาร์ดโค้ดใน PHP
	playerConfigs := []struct {
		TournamentID int
		TeamIDs      []int
		LeagueID     int // สอดคล้องกับ League ID ใน DB ของคุณ
	}{
		{TournamentID: 195, TeamIDs: []int{6025, 6040, 6039, 6038, 6037, 6036, 6035, 6034, 6033, 6032, 6031, 6030, 6029, 6028, 6027, 6026}, LeagueID: 1}, // ตัวอย่างสำหรับ R1, แมปกับ League ID จริงของคุณ
		{TournamentID: 196, TeamIDs: []int{6058, 6057, 6056, 6055, 6054, 6053, 6052, 6051, 6050, 6049, 6048, 6047, 6046, 6045, 6044, 6043, 6042, 6041}, LeagueID: 2}, // ตัวอย่างสำหรับ R2
		// เพิ่มการกำหนดค่าเพิ่มเติมตามความจำเป็น, แมปกับ League ID ใน DB ของคุณ
	}

	for _, config := range playerConfigs {
		for _, teamID := range config.TeamIDs {
			url := fmt.Sprintf("%s?page=1&tournament=%d&tournament_team=%d", baseAPIURL, config.TournamentID, teamID)
			log.Printf("Scraping players for Tournament %d, Team %d: %s", config.TournamentID, teamID, url)

			var apiResponse struct {
				Results []models.PlayerAPI `json:"results"`
			}
			err := FetchAndParseAPI(url, &apiResponse)
			if err != nil {
				log.Printf("Error fetching players from %s: %v", url, err)
				continue
			}

			for _, apiPlayer := range apiResponse.Results {
				// ดาวน์โหลดรูปภาพผู้เล่น
				photoPath := ""
				if apiPlayer.Photo != "" {
					downloadedPath, err := DownloadImage(apiPlayer.Photo, "./img/player")
					if err != nil {
						log.Printf("Warning: Failed to download player photo for %s: %v", apiPlayer.FullName, err)
					} else {
						photoPath = downloadedPath
					}
				}

				// รับ Nationality ID
				nationalityID := sql.NullInt64{Valid: false}
				if apiPlayer.Nationality.Code != "" {
					nID, err := database.GetNationalityID(db, apiPlayer.Nationality.Code, apiPlayer.Nationality.FullName)
					if err != nil {
						log.Printf("Warning: Failed to get nationality ID for %s: %v", apiPlayer.Nationality.FullName, err)
					} else {
						nationalityID = sql.NullInt64{Int64: int64(nID), Valid: true}
					}
				}

				// รับ Team ID (จาก club_name)
				playerTeamID := sql.NullInt64{Valid: false}
				if apiPlayer.ClubName != "" {
					tID, err := database.GetTeamIDByThaiName(db, apiPlayer.ClubName, "") // สมมติว่าโลโก้ไม่พร้อมใช้งานที่นี่
					if err != nil {
						log.Printf("Warning: Failed to get team ID for player %s's club %s: %v", apiPlayer.FullName, apiPlayer.ClubName, err)
					} else {
						playerTeamID = sql.NullInt64{Int64: int64(tID), Valid: true}
					}
				}
				// จัดการการแทนที่ Team ID เฉพาะตามที่เห็นใน PHP (ถ้าจำเป็น)
				// เช่น if apiPlayer.ClubName == "นครศรี ยูไนเต็ด" { playerTeamID = sql.NullInt64{Int64: 923, Valid: true} }

				// เตรียมโครงสร้าง PlayerDB
				playerDB := models.PlayerDB{
					PlayerRefID:   sql.NullInt64{Int64: int64(apiPlayer.ID), Valid: true},
					LeagueID:      sql.NullInt64{Int64: int64(config.LeagueID), Valid: true}, // ใช้ League ID จาก config
					TeamID:        playerTeamID,
					NationalityID: nationalityID,
					Name:          apiPlayer.FullName,
					FullNameEN:    sql.NullString{String: apiPlayer.FullNameEN, Valid: apiPlayer.FullNameEN != ""},
					ShirtNumber:   apiPlayer.ShirtNumber,
					Position:      sql.NullString{String: apiPlayer.PositionShortName, Valid: apiPlayer.PositionShortName != ""},
					PhotoURL:      sql.NullString{String: photoPath, Valid: photoPath != ""},
					MatchesPlayed: apiPlayer.MatchCount,
					Goals:         apiPlayer.GoalFor,
					YellowCards:   apiPlayer.YellowCardAcc,
					RedCards:      apiPlayer.RedCardViolentConductAcc,
				}

				// แทรกหรืออัปเดตผู้เล่นใน DB
				err = database.InsertOrUpdatePlayer(db, playerDB)
				if err != nil {
					log.Printf("Error saving player %s to DB: %v", apiPlayer.FullName, err)
				}
			}
		}
	}
	return nil
}
