package scraper

import (
	"database/sql"
	"fmt"
	"log"

	"go-ballthai-scraper/database" // แก้ไข: ตรวจสอบให้แน่ใจว่าชื่อโมดูลตรงกับ go.mod ของคุณ
	"go-ballthai-scraper/models"   // แก้ไข: ตรวจสอบให้แน่ใจว่าชื่อโมดูลตรงกับ go.mod ของคุณ
)

// scrapeMatchesByConfig เป็นฟังก์ชันทั่วไปสำหรับจัดการการกำหนดค่าการ scrape แมตช์ต่างๆ
func scrapeMatchesByConfig(db *sql.DB, baseURL string, pages []int, tournamentParam string, leagueType string, dbLeagueID int) error {
	for _, page := range pages {
		url := fmt.Sprintf("%s%d%s", baseURL, page, tournamentParam)
		log.Printf("Scraping matches for %s, page %d: %s", leagueType, page, url)

		var apiResponse struct {
			Results []models.MatchAPI `json:"results"`
		}
		err := FetchAndParseAPI(url, &apiResponse)
		if err != nil {
			log.Printf("Error fetching matches from %s: %v", url, err)
			continue // ดำเนินการไปยังหน้าถัดไปแม้ว่าหน้าปัจจุบันจะล้มเหลว
		}

		for _, apiMatch := range apiResponse.Results {
			// --- ดึง stage_id จากชื่อ stage_name แล้วนำไปเก็บใน matches (ไม่ต้องดึง/insert league แล้ว) ---
			var stageID int
			if apiMatch.StageName != "" {
				stageID, err = database.GetStageID(db, apiMatch.StageName, dbLeagueID)
				if err != nil {
					log.Printf("Warning: Failed to insert/update stage for match %d (%s): %v", apiMatch.ID, apiMatch.StageName, err)
				}
			}
			   // ไม่ต้องดาวน์โหลดโลโก้ซ้ำที่นี่ เพราะ InsertOrUpdateTeam จะจัดการให้แล้ว

			   // รับ Home Team ID (หาเฉพาะใน DB ไม่ insert/update)
			   homeTeamID := sql.NullInt64{Valid: false}
			   if apiMatch.HomeTeamName != "" {
				   tID, err := database.GetTeamIDByThaiName(db, apiMatch.HomeTeamName, "")
				   if err == nil {
					   homeTeamID = sql.NullInt64{Int64: int64(tID), Valid: true}
				   }
			   }

			   // รับ Away Team ID (หาเฉพาะใน DB ไม่ insert/update)
			   awayTeamID := sql.NullInt64{Valid: false}
			   if apiMatch.AwayTeamName != "" {
				   tID, err := database.GetTeamIDByThaiName(db, apiMatch.AwayTeamName, "")
				   if err == nil {
					   awayTeamID = sql.NullInt64{Int64: int64(tID), Valid: true}
				   }
			   }

			// รับ Channel ID (Main TV)
			channelID := sql.NullInt64{Valid: false}
			if apiMatch.ChannelInfo.Name != "" {
				cID, err := database.GetChannelID(db, apiMatch.ChannelInfo.Name, apiMatch.ChannelInfo.Logo, "TV")
				if err != nil {
					log.Printf("Warning: Failed to get channel ID for match %d (%s): %v", apiMatch.ID, apiMatch.ChannelInfo.Name, err)
				} else {
					channelID = sql.NullInt64{Int64: int64(cID), Valid: true}
				}
			}

			// รับ Live Channel ID
			liveChannelID := sql.NullInt64{Valid: false}
			if apiMatch.LiveInfo.Name != "" {
				lcID, err := database.GetChannelID(db, apiMatch.LiveInfo.Name, apiMatch.LiveInfo.Logo, "Live Stream")
				if err != nil {
					log.Printf("Warning: Failed to get live channel ID for match %d (%s): %v", apiMatch.ID, apiMatch.LiveInfo.Name, err)
				} else {
					liveChannelID = sql.NullInt64{Int64: int64(lcID), Valid: true}
				}
			}

			// กำหนดสถานะแมตช์ตาม 'match_status' ของ API
			matchStatus := sql.NullString{Valid: false}
			if statusStr, ok := apiMatch.MatchStatus.(string); ok {
				switch statusStr {
				case "2":
					matchStatus = sql.NullString{String: "FINISHED", Valid: true}
				case "1":
					matchStatus = sql.NullString{String: "FIXTURE", Valid: true}
				case "":
					matchStatus = sql.NullString{String: "ADD", Valid: true}
				default:
					matchStatus = sql.NullString{String: statusStr, Valid: true}
				}
			} else if apiMatch.MatchStatus != nil {
				// Fallback for non-string types, e.g., numbers
				statusStr := fmt.Sprintf("%v", apiMatch.MatchStatus)
				switch statusStr {
				case "2":
					matchStatus = sql.NullString{String: "FINISHED", Valid: true}
				case "1":
					matchStatus = sql.NullString{String: "FIXTURE", Valid: true}
				case "":
					matchStatus = sql.NullString{String: "ADD", Valid: true}
				default:
					matchStatus = sql.NullString{String: statusStr, Valid: true}
				}
			} else {
				// ถ้าไม่มีค่าเลย ให้เป็น "ADD"
				matchStatus = sql.NullString{String: "ADD", Valid: true}
			}

			// เตรียมโครงสร้าง MatchDB
			matchDB := models.MatchDB{
				MatchRefID:    apiMatch.ID,
				StartDate:     apiMatch.StartDate,
				StartTime:     apiMatch.StartTime,
				LeagueID:      sql.NullInt64{Int64: int64(dbLeagueID), Valid: true},      // ใช้ DB league ID จาก config
				StageID:       sql.NullInt64{Int64: int64(stageID), Valid: stageID != 0}, // ใช้ stage_id จาก DB ที่ได้จากชื่อ
				HomeTeamID:    homeTeamID,
				AwayTeamID:    awayTeamID,
				ChannelID:     channelID,
				LiveChannelID: liveChannelID,
				HomeScore:     sql.NullInt64{Int64: int64(apiMatch.HomeGoalCount), Valid: true},
				AwayScore:     sql.NullInt64{Int64: int64(apiMatch.AwayGoalCount), Valid: true},
				MatchStatus:   matchStatus,
			}

			// แทรกหรืออัปเดตแมตช์ใน DB
			err = database.InsertOrUpdateMatch(db, matchDB)
			if err != nil {
				log.Printf("Error saving match %d to DB: %v", apiMatch.ID, err)
			}
		}
	}
	return nil
}

// ScrapeThaileagueMatches ดึงข้อมูลแมตช์สำหรับ Thai League (T1, T2, T3 regions, Samipro)
func ScrapeThaileagueMatches(db *sql.DB, targetLeague string) error { // เพิ่ม targetLeague parameter
	   baseURL := "https://competition.tl.prod.c0d1um.io/thaileague/api/match-day-match-public/?page="
	   singlePage := []int{1}

	   // ดึงลีกทั้งหมดจาก DB
	   leagues, err := database.GetAllLeagues(db)
	   if err != nil {
		   return fmt.Errorf("failed to get leagues from DB: %w", err)
	   }

	   for _, league := range leagues {
		   // ถ้า targetLeague ไม่ว่าง ให้ filter เฉพาะลีกที่ตรง
		   if targetLeague != "" && targetLeague != "all" && league.Name != targetLeague {
			   continue
		   }
		   // ข้ามลีกที่ thaileageid เป็น 0 หรือว่าง
		   if league.ThaileageID == 0 {
			   continue
		   }
		   tournamentParam := fmt.Sprintf("&tournament=%d", league.ThaileageID)
		   log.Printf("Scraping league: %s (thaileageid=%d)", league.Name, league.ThaileageID)
		   if err := scrapeMatchesByConfig(db, baseURL, singlePage, tournamentParam, league.Name, league.ID); err != nil {
			   log.Printf("Error scraping %s: %v", league.Name, err)
		   }
	   }
	   return nil
}


