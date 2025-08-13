package scraper

import (
	"database/sql"
	"log"
	// "fmt" // Removed: fmt is not directly used in this file
	// The 'fmt' package was imported but not used, causing a compile error.
	// It has been removed.

	"go-ballthai-scraper/database" // ตรวจสอบให้แน่ใจว่าชื่อโมดูลตรงกับ go.mod ของคุณ
	"go-ballthai-scraper/models"  // ตรวจสอบให้แน่ใจว่าชื่อโมดูลตรงกับ go.mod ของคุณ
)

// ScrapeStandings ดึงข้อมูลตารางคะแนนลีกจาก API และบันทึกลงฐานข้อมูล
func ScrapeStandings(db *sql.DB) error {
	// แมปค่า $_GET['table'] ของ PHP กับ URL API และ DB league ID
	standingConfigs := map[string]struct {
		URL      string
		LeagueID int // สอดคล้องกับ League ID ใน DB ของคุณ
		IsT3     bool
	}{
		"t1": {URL: "https://competition.tl.prod.c0d1um.io/thaileague/api/stage-standing-public/?tournament=207", LeagueID: 1}, // ตัวอย่าง: แมปกับ DB league ID 1
		"t2": {URL: "https://competition.tl.prod.c0d1um.io/thaileague/api/stage-standing-public/?tournament=196", LeagueID: 2}, // ตัวอย่าง: แมปกับ DB league ID 2
		"t3": {URL: "https://competition.tl.prod.c0d1um.io/thaileague/api/stage-standing-public/?tournament=197", LeagueID: 3, IsT3: true}, // ตัวอย่าง: แมปกับ DB league ID 3, การจัดการพิเศษสำหรับ T3
		"samipro": {URL: "https://competition.tl.prod.c0d1um.io/thaileague/api/stage-standing-public/dashboard/?tournament=206", LeagueID: 59}, // ตัวอย่าง: แมปกับ DB league ID 59
		"revo": {URL: "https://competition.tl.prod.c0d1um.io/thaileague/api/stage-standing-public/?tournament=155", LeagueID: 4}, // ตัวอย่าง: แมปกับ DB league ID 4
		// เพิ่มการกำหนดค่าเพิ่มเติมตามความจำเป็น
	}

	for configName, config := range standingConfigs {
		log.Printf("Scraping standings for %s (%s)", configName, config.URL)
		
		var apiResponse []models.StandingAPI // API สำหรับตารางคะแนนคืนค่าเป็น array โดยตรง
		err := FetchAndParseAPI(config.URL, &apiResponse)
		if err != nil {
			log.Printf("Error fetching standings for %s: %v", configName, err)
			continue
		}

			   // เดิม: filter เฉพาะ SOUTH สำหรับ T3
			   // ใหม่: ไม่ filter ใดๆ เพื่อเก็บ standings ทุกโซนของ T3

		for _, apiStanding := range apiResponse {
			// รับ Team ID
			teamID := 0
			if apiStanding.TournamentTeamName != "" {
				tID, err := database.GetTeamIDByThaiName(db, apiStanding.TournamentTeamName, apiStanding.TournamentTeamLogo)
				if err != nil {
					log.Printf("Warning: Failed to get team ID for standing team %s: %v", apiStanding.TournamentTeamName, err)
					continue // ข้ามหากไม่สามารถแก้ไข Team ID ได้
				}
				teamID = tID
			} else {
				log.Printf("Warning: Standing entry for %s has no team name, skipping.", configName)
				continue
			}

			   // หา stage_id จาก stage_name (ถ้าไม่มีจะ insert ให้)
			   stageID := sql.NullInt64{Valid: false}
			   if apiStanding.StageName != "" {
				   id, err := database.GetStageID(db, apiStanding.StageName, config.LeagueID)
				   if err == nil {
					   stageID = sql.NullInt64{Int64: int64(id), Valid: true}
				   }
			   }
			   standingDB := models.StandingDB{
				   LeagueID:       config.LeagueID,
				   TeamID:         teamID,
				   MatchesPlayed:  apiStanding.MatchPlay,
				   Wins:           apiStanding.Win,
				   Draws:          apiStanding.Draw,
				   Losses:         apiStanding.Lose,
				   GoalsFor:       apiStanding.GoalFor,
				   GoalsAgainst:   apiStanding.GoalAgainst,
				   GoalDifference: apiStanding.GoalDifference,
				   Points:         apiStanding.Point,
				   CurrentRank:    sql.NullInt64{Int64: int64(apiStanding.CurrentRank), Valid: apiStanding.CurrentRank != 0},
				   Round:          sql.NullInt64{Valid: false},
				   StageID:        stageID,
			   }

			// แทรกหรืออัปเดตตารางคะแนนใน DB
			err = database.InsertOrUpdateStanding(db, standingDB)
			if err != nil {
				log.Printf("Error saving standing for team %s in league %s to DB: %v", apiStanding.TournamentTeamName, configName, err)
			}
		}
	}
	return nil
}
