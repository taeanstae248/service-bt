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
			// --- เพิ่ม: จัดเก็บข้อมูลลีก ---
			var leagueID int
			if apiMatch.TournamentName != "" {
				leagueID, err = database.GetLeagueID(db, apiMatch.TournamentNameEN, apiMatch.TournamentName)
				if err != nil {
					log.Printf("Warning: Failed to insert/update league for match %d (%s): %v", apiMatch.ID, apiMatch.TournamentName, err)
				}
			}

			// --- เพิ่ม: ดึง stage_id จากชื่อ stage_name แล้วนำไปเก็บใน matches ---
			var stageID int
			if apiMatch.StageName != "" {
				stageID, err = database.GetStageID(db, apiMatch.StageName, leagueID)
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

	singlePage := []int{1} // ใช้แค่หน้าแรกเพื่อหลีกเลี่ยง 404s ในหน้าถัดไป

	switch targetLeague {
	case "t1":
		log.Println("Scraping Thai League 1 (T1) Matches...")
		if err := scrapeMatchesByConfig(db, baseURL, singlePage, "&tournament=207", "T1", 1); err != nil { // แมปกับ DB league ID สำหรับ T1
			return fmt.Errorf("failed to scrape T1 matches: %w", err)
		}
	case "t2":
		log.Println("Scraping Thai League 2 (T2) Matches...")
		if err := scrapeMatchesByConfig(db, baseURL, singlePage, "&tournament=208", "T2", 2); err != nil { // แมปกับ DB league ID สำหรับ T2
			return fmt.Errorf("failed to scrape T2 matches: %w", err)
		}
	case "t3_BKK":
		log.Println("Scraping Thai League 3 (T3) BKK Region Matches...")
		if err := scrapeMatchesByConfig(db, baseURL, singlePage, "&stage=982&tournament=197&tournament_team=", "T3 Bangkok", 3); err != nil { // แมปกับ DB league ID สำหรับ T3
			return fmt.Errorf("failed to scrape T3 BKK matches: %w", err)
		}
	case "t3_EAST":
		log.Println("Scraping Thai League 3 (T3) EAST Region Matches...")
		if err := scrapeMatchesByConfig(db, baseURL, singlePage, "&stage=999&tournament=197&tournament_team=", "T3 East", 3); err != nil {
			return fmt.Errorf("failed to scrape T3 EAST matches: %w", err)
		}
	case "t3_WEST":
		log.Println("Scraping Thai League 3 (T3) WEST Region Matches...")
		if err := scrapeMatchesByConfig(db, baseURL, singlePage, "&stage=1000&tournament=197&tournament_team=", "T3 West", 3); err != nil {
			return fmt.Errorf("failed to scrape T3 WEST matches: %w", err)
		}
	case "t3_NORTH":
		log.Println("Scraping Thai League 3 (T3) NORTH Region Matches...")
		if err := scrapeMatchesByConfig(db, baseURL, singlePage, "&stage=981&tournament=197&tournament_team=", "T3 North", 3); err != nil {
			return fmt.Errorf("failed to scrape T3 NORTH matches: %w", err)
		}
	case "t3_NORTHEAST":
		log.Println("Scraping Thai League 3 (T3) NORTHEAST Region Matches...")
		if err := scrapeMatchesByConfig(db, baseURL, singlePage, "&stage=998&tournament=197&tournament_team=", "T3 Northeast", 3); err != nil {
			return fmt.Errorf("failed to scrape T3 NORTHEAST matches: %w", err)
		}
	case "t3_SOUTH":
		log.Println("Scraping Thai League 3 (T3) SOUTH Region Matches...")
		if err := scrapeMatchesByConfig(db, baseURL, singlePage, "&stage=1001&tournament=197&tournament_team=", "T3 South", 3); err != nil {
			return fmt.Errorf("failed to scrape T3 SOUTH matches: %w", err)
		}
	case "samipro":
		log.Println("Scraping Samipro Matches...")
		if err := scrapeMatchesByConfig(db, baseURL, singlePage, "&tournament=206", "Samipro", 59); err != nil { // แมปกับ DB league ID สำหรับ Samipro
			return fmt.Errorf("failed to scrape Samipro matches: %w", err)
		}
	case "", "all": // Default to all if no specific league or "all" is provided
		log.Println("Scraping ALL Thai League Matches (T1, T2, T3 Regions, Samipro)...")
		// Call all of them, but still limit to single page for now
		if err := scrapeMatchesByConfig(db, baseURL, singlePage, "&tournament=207", "T1", 1); err != nil {
			log.Printf("Error scraping T1 matches: %v", err)
		}
		if err := scrapeMatchesByConfig(db, baseURL, singlePage, "&tournament=196", "T2", 2); err != nil {
			log.Printf("Error scraping T2 matches: %v", err)
		}
		t3Stages := map[string]string{
			"BKK":       "&stage=982&tournament=197&tournament_team=",
			"EAST":      "&stage=999&tournament=197&tournament_team=",
			"WEST":      "&stage=1000&tournament=197&tournament_team=",
			"NORTH":     "&stage=981&tournament=197&tournament_team=",
			"NORTHEAST": "&stage=998&tournament=197&tournament_team=",
			"SOUTH":     "&stage=1001&tournament=197&tournament_team=",
		}
		for region, param := range t3Stages {
			if err := scrapeMatchesByConfig(db, baseURL, singlePage, param, "T3 "+region, 3); err != nil { // ใช้ league ID สำหรับ T3
				log.Printf("Error scraping T3 %s matches: %v", region, err)
			}
		}
		if err := scrapeMatchesByConfig(db, baseURL, singlePage, "&tournament=206", "Samipro", 59); err != nil {
			log.Printf("Error scraping Samipro matches: %v", err)
		}
	default:
		return fmt.Errorf("invalid target league specified: %s", targetLeague)
	}

	return nil
}

// ScrapeBallthaiCupMatches ดึงข้อมูลแมตช์สำหรับถ้วยต่างๆ (Revo, FA, BGC)
func ScrapeBallthaiCupMatches(db *sql.DB) error {
	baseURL := "https://competition.tl.prod.c0d1um.io/thaileague/api/match-day-match-public/?page="

	// Revo League Cup
	if err := scrapeMatchesByConfig(db, baseURL, []int{1}, "&tournament=202", "revo", 4); err != nil { // แมปกับ DB league ID สำหรับ Revo
		return fmt.Errorf("failed to scrape Revo Cup matches: %w", err)
	}
	// FA Cup
	if err := scrapeMatchesByConfig(db, baseURL, []int{1}, "&tournament=199", "fa", 5); err != nil { // แมปกับ DB league ID สำหรับ FA
		return fmt.Errorf("failed to scrape FA Cup matches: %w", err)
	}
	// BGC Cup
	if err := scrapeMatchesByConfig(db, baseURL, []int{1}, "&tournament=205", "bgc", 6); err != nil { // แมปกับ DB league ID สำหรับ BGC
		return fmt.Errorf("failed to scrape BGC Cup matches: %w", err)
	}

	return nil
}

// ScrapeThaileaguePlayoffMatches ดึงข้อมูลแมตช์เพลย์ออฟสำหรับ Thai League
func ScrapeThaileaguePlayoffMatches(db *sql.DB) error {
	// Playoff T3 stages
	playoffURLs := []string{
		"https://competition.tl.prod.c0d1um.io/thaileague/api/match-day-match-public/?tournament=197&tournament_team=&stage=1032",
		"https://competition.tl.prod.c0d1um.io/thaileague/api/match-day-match-public/?tournament=197&tournament_team=&stage=1033",
		"https://competition.tl.prod.c0d1um.io/thaileague/api/match-day-match-public/?tournament=197&tournament_team=&stage=1034",
		"https://competition.tl.prod.c0d1um.io/thaileague/api/match-day-match-public/?tournament=197&tournament_team=&stage=&match_day=&match_status=results&only_valid_match=true",
	}

	for _, url := range playoffURLs {
		log.Printf("Scraping playoff matches from: %s", url)
		var apiResponse struct {
			Results []models.MatchAPI `json:"results"`
		}
		// หมายเหตุ: URL เพลย์ออฟไม่มีพารามิเตอร์ page= เหมือนอื่นๆ ดังนั้นจึงส่ง slice ของ pages ว่างเปล่า
		if err := FetchAndParseAPI(url, &apiResponse); err != nil {
			log.Printf("Error fetching playoff matches from %s: %v", url, err)
			continue
		}

		for _, apiMatch := range apiResponse.Results {
			// ดาวน์โหลดโลโก้ทีมเหย้า
			homeLogoPath := ""
			if apiMatch.HomeTeamLogo != "" {
				downloadedPath, err := DownloadImage(apiMatch.HomeTeamLogo, "./img/source")
				if err != nil {
					log.Printf("Warning: Failed to download home team logo for match %d: %v", apiMatch.ID, err)
				} else {
					homeLogoPath = downloadedPath
				}
			}

			// ดาวน์โหลดโลโก้ทีมเยือน
			awayLogoPath := ""
			if apiMatch.AwayTeamLogo != "" {
				downloadedPath, err := DownloadImage(apiMatch.AwayTeamLogo, "./img/source")
				if err != nil {
					log.Printf("Warning: Failed to download away team logo for match %d: %v", apiMatch.ID, err)
				} else {
					awayLogoPath = downloadedPath
				}
			}

			// รับ Home Team ID
			homeTeamID := sql.NullInt64{Valid: false}
			if apiMatch.HomeTeamName != "" {
				tID, err := database.GetTeamIDByThaiName(db, apiMatch.HomeTeamName, homeLogoPath)
				if err != nil {
					log.Printf("Warning: Failed to get home team ID for match %d (%s): %v", apiMatch.ID, apiMatch.HomeTeamName, err)
				} else {
					homeTeamID = sql.NullInt64{Int64: int64(tID), Valid: true}
				}
			}

			// รับ Away Team ID
			awayTeamID := sql.NullInt64{Valid: false}
			if apiMatch.AwayTeamName != "" {
				tID, err := database.GetTeamIDByThaiName(db, apiMatch.AwayTeamName, awayLogoPath)
				if err != nil {
					log.Printf("Warning: Failed to get away team ID for match %d (%s): %v", apiMatch.ID, apiMatch.AwayTeamName, err)
				} else {
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
				default:
					matchStatus = sql.NullString{String: statusStr, Valid: true}
				}
			}

			// เตรียมโครงสร้าง MatchDB
			matchDB := models.MatchDB{
				MatchRefID:    apiMatch.ID,
				StartDate:     apiMatch.StartDate,
				StartTime:     apiMatch.StartTime,
				LeagueID:      sql.NullInt64{Int64: 7, Valid: true}, // กำหนด DB league ID เฉพาะสำหรับเพลย์ออฟหากจำเป็น เช่น 7
				HomeTeamID:    homeTeamID,
				AwayTeamID:    awayTeamID,
				ChannelID:     channelID,
				LiveChannelID: liveChannelID,
				HomeScore:     sql.NullInt64{Int64: int64(apiMatch.HomeGoalCount), Valid: true},
				AwayScore:     sql.NullInt64{Int64: int64(apiMatch.AwayGoalCount), Valid: true},
				MatchStatus:   matchStatus,
			}

			// แทรกหรืออัปเดตแมตช์ใน DB
			if err := database.InsertOrUpdateMatch(db, matchDB); err != nil {
				log.Printf("Error saving match %d to DB: %v", apiMatch.ID, err)
			}
		}
	}
	return nil
}
