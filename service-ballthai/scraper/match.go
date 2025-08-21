package scraper

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"go-ballthai-scraper/database"
	"go-ballthai-scraper/models"
)

// ensureTeamAndLogo ตรวจสอบและอัปเดตข้อมูลทีมและโลโก้ในตาราง teams
func ensureTeamAndLogo(db *sql.DB, teamName string) error {
	tID, err := database.GetTeamIDByThaiName(db, teamName, "")
	needUpdate := false
	if err == nil {
		// ตรวจสอบโลโก้ ถ้าไม่มีโลโก้ให้ดึงใหม่
		var logo sql.NullString
		err2 := db.QueryRow("SELECT logo_url FROM teams WHERE id = ?", tID).Scan(&logo)
		if err2 != nil || !logo.Valid || logo.String == "" {
			needUpdate = true
		}
	} else {
		needUpdate = true
	}
	if needUpdate {
		teams, _ := FetchTeamsByLeagueID("")
		for _, team := range teams {
			if team.Name == teamName {
				baseName := team.NameEN
				if baseName == "" {
					baseName = slugify(team.Name)
				}
				var logoPath string
				if team.Logo != "" {
					ext := path.Ext(team.Logo)
					if ext == "" {
						ext = ".png"
					}
					fileName := sanitizeFileName(baseName) + ext
					logoPath = "/img/teams/" + fileName
					_ = downloadLogoToFolder(team.Logo, baseName)
				}
				teamDB := models.TeamDB{
					NameTH:   team.Name,
					NameEN:   sql.NullString{String: team.NameEN, Valid: team.NameEN != ""},
					LogoURL:  sql.NullString{String: logoPath, Valid: logoPath != ""},
					Website:  sql.NullString{String: team.Website, Valid: team.Website != ""},
					Shop:     sql.NullString{String: team.Shop, Valid: team.Shop != ""},
				}
				_ = database.InsertOrUpdateTeam(db, teamDB)
				break
			}
		}
	}
	return nil
}


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
			if apiMatch.StageName != "" {
				_, err = database.GetStageID(db, apiMatch.StageName, dbLeagueID)
				if err != nil {
					log.Printf("Warning: Failed to insert/update stage for match %d (%s): %v", apiMatch.ID, apiMatch.StageName, err)
				}
			}

			// เช็ค match_status ใน DB ถ้าเป็น OFF ให้ข้ามการอัปเดต
			var currentStatus string
			err := db.QueryRow("SELECT match_status FROM matches WHERE match_ref_id = ?", apiMatch.ID).Scan(&currentStatus)
			if err == nil && currentStatus == "OFF" {
				log.Printf("Skip update match %d (status=OFF)", apiMatch.ID)
				continue
			}

			// ไม่ต้องดาวน์โหลดโลโก้ซ้ำที่นี่ เพราะ InsertOrUpdateTeam จะจัดการให้แล้ว

			   // รับ Home Team ID
			if apiMatch.HomeTeamName != "" {
				_, _ = database.GetTeamIDByThaiName(db, apiMatch.HomeTeamName, "")
			}
			if apiMatch.AwayTeamName != "" {
				_, _ = database.GetTeamIDByThaiName(db, apiMatch.AwayTeamName, "")
			}

			// รับ Channel ID (Main TV) - ดาวน์โหลดโลโก้ลง server ถ้ายังไม่มี
			// channelID := sql.NullInt64{Valid: false}
			channelLogoPath := ""
			if apiMatch.ChannelInfo.Name != "" && apiMatch.ChannelInfo.Logo != "" {
				ext := path.Ext(apiMatch.ChannelInfo.Logo)
				if ext == "" {
					ext = ".png"
				}
				safeName := sanitizeFileName(apiMatch.ChannelInfo.Name)
				fileName := safeName + ext
				localPath := path.Join("img/channels", fileName)
				webPath := "/img/channels/" + fileName
				// ถ้ายังไม่มีไฟล์ ให้ดาวน์โหลด
				if _, err := os.Stat(localPath); os.IsNotExist(err) {
					err := downloadChannelLogoToFolder(apiMatch.ChannelInfo.Logo, safeName)
					if err != nil {
						log.Printf("Warning: Failed to download channel logo for %s: %v", apiMatch.ChannelInfo.Name, err)
						webPath = apiMatch.ChannelInfo.Logo // fallback เป็น url เดิม
					}
				}
				channelLogoPath = webPath
			} else if apiMatch.ChannelInfo.Name != "" {
				channelLogoPath = ""
			}
			// var channelID sql.NullInt64 = sql.NullInt64{Valid: false}
			if apiMatch.ChannelInfo.Name != "" {
				_, err := database.GetChannelID(db, apiMatch.ChannelInfo.Name, channelLogoPath, "TV")
				if err != nil {
					log.Printf("Warning: Failed to get channel ID for match %d (%s): %v", apiMatch.ID, apiMatch.ChannelInfo.Name, err)
				}
			}

			// รับ Live Channel ID - ดาวน์โหลดโลโก้ลง server ถ้ายังไม่มี
			// liveChannelID := sql.NullInt64{Valid: false}
			liveChannelLogoPath := ""
			if apiMatch.LiveInfo.Name != "" && apiMatch.LiveInfo.Logo != "" {
				ext := path.Ext(apiMatch.LiveInfo.Logo)
				if ext == "" {
					ext = ".png"
				}
				safeName := sanitizeFileName(apiMatch.LiveInfo.Name)
				fileName := safeName + ext
				localPath := path.Join("img/channels", fileName)
				webPath := "/img/channels/" + fileName
				if _, err := os.Stat(localPath); os.IsNotExist(err) {
					err := downloadChannelLogoToFolder(apiMatch.LiveInfo.Logo, safeName)
					if err != nil {
						log.Printf("Warning: Failed to download live channel logo for %s: %v", apiMatch.LiveInfo.Name, err)
						webPath = apiMatch.LiveInfo.Logo
					}
				}
				liveChannelLogoPath = webPath
			} else if apiMatch.LiveInfo.Name != "" {
				liveChannelLogoPath = ""
			}
			// var liveChannelID sql.NullInt64 = sql.NullInt64{Valid: false}
			if apiMatch.LiveInfo.Name != "" {
				_, err := database.GetChannelID(db, apiMatch.LiveInfo.Name, liveChannelLogoPath, "Live Stream")
				if err != nil {
					log.Printf("Warning: Failed to get live channel ID for match %d (%s): %v", apiMatch.ID, apiMatch.LiveInfo.Name, err)
				}
			}
		}
	}
	return nil
}

func downloadChannelLogoToFolder(logoURL, channelName string) error {
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
	fileName := sanitizeFileName(channelName) + ext
	outPath := path.Join("img/channels", fileName)
	if err := os.MkdirAll("img/channels", 0755); err != nil {
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

func sanitizeFileName(name string) string {
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", " "}
	for _, c := range invalid {
		name = strings.ReplaceAll(name, c, "_")
	}
	return name
}

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


