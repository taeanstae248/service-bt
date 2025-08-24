package scraper

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"time"
	"strings"
	"errors"
	"go-ballthai-scraper/database"
	"go-ballthai-scraper/models"
)

var (
	ErrInvalidPage = errors.New("invalid page")
	ErrNoResults   = errors.New("no results")
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
	// If pages provided explicitly, use them
	if len(pages) > 0 {
		for _, p := range pages {
			if err := scrapeMatchesPage(db, baseURL, p, tournamentParam, leagueType, dbLeagueID); err != nil {
				log.Printf("Error scraping page %d for %s: %v", p, leagueType, err)
			}
		}
		return nil
	}

	// Auto-pagination: iterate pages until an empty result set or a safety maxPages
	maxPages := 200
	for page := 1; page <= maxPages; page++ {
		err := scrapeMatchesPage(db, baseURL, page, tournamentParam, leagueType, dbLeagueID)
		if err == ErrInvalidPage {
			log.Printf("API reports invalid page %d for %s, stopping pagination", page, leagueType)
			break
		}
		if err == ErrNoResults {
			log.Printf("No results on page %d for %s, stopping pagination", page, leagueType)
			break
		}
		if err != nil {
			log.Printf("Error scraping page %d for %s: %v", page, leagueType, err)
			time.Sleep(200 * time.Millisecond)
			continue
		}
		// be polite
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}


// scrapeMatchesPage processes a single page of matches
func scrapeMatchesPage(db *sql.DB, baseURL string, page int, tournamentParam string, leagueType string, dbLeagueID int) error {
	url := fmt.Sprintf("%s%d%s", baseURL, page, tournamentParam)
	log.Printf("Scraping matches for %s, page %d: %s", leagueType, page, url)

	var apiResponse struct {
		Results []models.MatchAPI `json:"results"`
	}
	if err := FetchAndParseAPI(url, &apiResponse); err != nil {
		// check for invalid page message from API
		if strings.Contains(err.Error(), "Invalid page") || strings.Contains(err.Error(), "404") {
			return ErrInvalidPage
		}
		return fmt.Errorf("Error fetching matches from %s: %w", url, err)
	}

	if len(apiResponse.Results) == 0 {
		log.Printf("No results on page %d for %s", page, leagueType)
		return ErrNoResults
	}

	for _, apiMatch := range apiResponse.Results {
		var stageID int
		if apiMatch.StageName != "" {
			sid, errStage := database.GetStageID(db, apiMatch.StageName, dbLeagueID)
			if errStage != nil {
				log.Printf("Warning: Failed to insert/update stage for match %d (%s): %v", apiMatch.ID, apiMatch.StageName, errStage)
			} else {
				stageID = sid
			}
		}

		var currentStatus string
		err := db.QueryRow("SELECT match_status FROM matches WHERE match_ref_id = ?", apiMatch.ID).Scan(&currentStatus)
		if err == nil && currentStatus == "OFF" {
			log.Printf("Skip update match %d (status=OFF)", apiMatch.ID)
			continue
		}

		var homeTeamID, awayTeamID int
		if apiMatch.HomeTeamName != "" {
			id, err := database.GetTeamIDByThaiName(db, apiMatch.HomeTeamName, "")
			if err != nil {
				log.Printf("Warning: GetTeamIDByThaiName home team '%s' failed: %v", apiMatch.HomeTeamName, err)
			} else {
				homeTeamID = id
			}
		}
		if apiMatch.AwayTeamName != "" {
			id, err := database.GetTeamIDByThaiName(db, apiMatch.AwayTeamName, "")
			if err != nil {
				log.Printf("Warning: GetTeamIDByThaiName away team '%s' failed: %v", apiMatch.AwayTeamName, err)
			} else {
				awayTeamID = id
			}
		}

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
			if _, err := os.Stat(localPath); os.IsNotExist(err) {
				err := downloadChannelLogoToFolder(apiMatch.ChannelInfo.Logo, safeName)
				if err != nil {
					log.Printf("Warning: Failed to download channel logo for %s: %v", apiMatch.ChannelInfo.Name, err)
					webPath = apiMatch.ChannelInfo.Logo
				}
			}
			channelLogoPath = webPath
		} else if apiMatch.ChannelInfo.Name != "" {
			channelLogoPath = ""
		}
		if apiMatch.ChannelInfo.Name != "" {
			_, err := database.GetChannelID(db, apiMatch.ChannelInfo.Name, channelLogoPath, "TV")
			if err != nil {
				log.Printf("Warning: Failed to get channel ID for match %d (%s): %v", apiMatch.ID, apiMatch.ChannelInfo.Name, err)
			}
		}

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
		if apiMatch.LiveInfo.Name != "" {
			_, err := database.GetChannelID(db, apiMatch.LiveInfo.Name, liveChannelLogoPath, "Live Stream")
			if err != nil {
				log.Printf("Warning: Failed to get live channel ID for match %d (%s): %v", apiMatch.ID, apiMatch.LiveInfo.Name, err)
			}
		}

		log.Printf("Processing match %d: leagueID=%d stageID=%d home='%s' away='%s'", apiMatch.ID, dbLeagueID, stageID, apiMatch.HomeTeamName, apiMatch.AwayTeamName)
		matchDB := models.MatchDB{
			MatchRefID: apiMatch.ID,
			StartDate:  apiMatch.StartDate,
			StartTime:  apiMatch.StartTime,
			LeagueID:   sql.NullInt64{Valid: dbLeagueID != 0, Int64: int64(dbLeagueID)},
			StageID:    sql.NullInt64{Valid: stageID != 0, Int64: int64(stageID)},
			HomeTeamID: sql.NullInt64{Valid: homeTeamID != 0, Int64: int64(homeTeamID)},
			AwayTeamID: sql.NullInt64{Valid: awayTeamID != 0, Int64: int64(awayTeamID)},
			ChannelID:  sql.NullInt64{Valid: false},
			LiveChannelID: sql.NullInt64{Valid: false},
			HomeScore:   sql.NullInt64{Valid: true, Int64: int64(apiMatch.HomeGoalCount)},
			AwayScore:   sql.NullInt64{Valid: true, Int64: int64(apiMatch.AwayGoalCount)},
			MatchStatus: sql.NullString{String: fmt.Sprint(apiMatch.MatchStatus), Valid: apiMatch.MatchStatus != nil},
		}

		if apiMatch.ChannelInfo.Name != "" {
			if chID, err := database.GetChannelID(db, apiMatch.ChannelInfo.Name, channelLogoPath, "TV"); err == nil {
				matchDB.ChannelID = sql.NullInt64{Valid: true, Int64: int64(chID)}
			}
		}
		if apiMatch.LiveInfo.Name != "" {
			if lchID, err := database.GetChannelID(db, apiMatch.LiveInfo.Name, liveChannelLogoPath, "Live Stream"); err == nil {
				matchDB.LiveChannelID = sql.NullInt64{Valid: true, Int64: int64(lchID)}
			}
		}

		if err := database.InsertOrUpdateMatch(db, matchDB); err != nil {
			log.Printf("Error saving match %d: %v", apiMatch.ID, err)
		} else {
			log.Printf("Saved match %d to DB", apiMatch.ID)
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
		// pass nil pages to enable auto-pagination (iterate until no more results)
		var singlePage []int = nil

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


