package scraper

import (
	"database/sql"
	"fmt"
	"log"

	"go-ballthai-scraper/database" // ตรวจสอบให้แน่ใจว่าชื่อโมดูลตรงกับ go.mod ของคุณ
	"go-ballthai-scraper/models"   // ตรวจสอบให้แน่ใจว่าชื่อโมดูลตรงกับ go.mod ของคุณ
)

// ScrapePlayers ดึงข้อมูลผู้เล่นจาก API ทุกลีกใน DB และบันทึกลงฐานข้อมูล
func ScrapePlayers(db *sql.DB) error {
	leagues, err := database.GetAllLeagues(db)
	if err != nil {
		return err
	}
	for _, league := range leagues {
		if !league.ThaileageID.Valid || league.ThaileageID.Int64 == 0 {
			continue
		}
		apiURL := fmt.Sprintf("https://competition.tl.prod.c0d1um.io/thaileague/api/player-public/all_players_search/?page=1&tournament=%d", league.ThaileageID.Int64)
		log.Printf("Scraping players from: %s (leagueID=%d, thaileageid=%d)", apiURL, league.ID, league.ThaileageID.Int64)

		var apiResponse struct {
			Results []models.PlayerAPI `json:"results"`
		}
		err := FetchAndParseAPI(apiURL, &apiResponse)
		if err != nil {
			log.Printf("Error fetching players from %s: %v", apiURL, err)
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
			var shirtNumber sql.NullInt64
			if apiPlayer.ShirtNumber != nil {
				shirtNumber = sql.NullInt64{Int64: int64(*apiPlayer.ShirtNumber), Valid: true}
			} else {
				shirtNumber = sql.NullInt64{Valid: false}
			}
			playerDB := models.PlayerDB{
				PlayerRefID:   sql.NullInt64{Int64: int64(apiPlayer.ID), Valid: true},
				LeagueID:      sql.NullInt64{Int64: int64(league.ID), Valid: true},
				TeamID:        playerTeamID,
				NationalityID: nationalityID,
				Name:          apiPlayer.FullName,
				FullNameEN:    sql.NullString{String: apiPlayer.FullNameEN, Valid: apiPlayer.FullNameEN != ""},
				ShirtNumber:   shirtNumber,
				Position:      sql.NullString{String: apiPlayer.PositionShortName, Valid: apiPlayer.PositionShortName != ""},
				PhotoURL:      sql.NullString{String: photoPath, Valid: photoPath != ""},
				MatchesPlayed: apiPlayer.MatchCount,
				Goals:         apiPlayer.GoalFor,
				YellowCards:   apiPlayer.YellowCardAcc,
				RedCards:      apiPlayer.RedCardViolentConductAcc,
				Status:        0, // default เปิดข้อมูล
			}

			// แทรกหรืออัปเดตผู้เล่นใน DB
			err = database.InsertOrUpdatePlayer(db, playerDB)
			if err != nil {
				log.Printf("Error saving player %s to DB: %v", apiPlayer.FullName, err)
			}
		}
	}
	return nil
}
