package scraper

import (
	"database/sql"
	"fmt"
	"log"
	// "path/filepath" // Not directly used in this file
	// "time" // Not directly used in this file

	"go-ballthai-scraper/database" // Ensure this module name matches your go.mod
	"go-ballthai-scraper/models"  // Ensure this module name matches your go.mod
)

// scrapeMatchesByConfig is a generic function to handle various match scraping configurations
func scrapeMatchesByConfig(db *sql.DB, baseURL string, pages []int, tournamentParam string, leagueType string, dbLeagueID int) error {
	var err error // Declare err once at the function level

	for _, page := range pages {
		url := fmt.Sprintf("%s%d%s", baseURL, page, tournamentParam)
		log.Printf("Scraping matches for %s, page %d: %s", leagueType, page, url)

		var apiResponse struct {
			Results []models.MatchAPI `json:"results"`
		}
		err = FetchAndParseAPI(url, &apiResponse) // Use = for reassignment
		if err != nil {
			log.Printf("Error fetching matches from %s: %v", url, err)
			continue // Continue to next page even if one fails
		}

		for _, apiMatch := range apiResponse.Results {
			// Download home team logo and get only the filename
			homeLogoFilename := ""
			if apiMatch.HomeTeamLogo != "" {
				var downloadErr error // Declare a new error variable for DownloadImage
				downloadedFilename, downloadErr := DownloadImage(apiMatch.HomeTeamLogo, "./img/source")
				if downloadErr != nil {
					log.Printf("Warning: Failed to download home team logo for match %d: %v", apiMatch.ID, downloadErr)
				} else {
					homeLogoFilename = downloadedFilename // Store only the filename
				}
			}

			// Download away team logo and get only the filename
			awayLogoFilename := ""
			if apiMatch.AwayTeamLogo != "" {
				var downloadErr error // Declare a new error variable for DownloadImage
				downloadedFilename, downloadErr := DownloadImage(apiMatch.AwayTeamLogo, "./img/source")
				if downloadErr != nil {
					log.Printf("Warning: Failed to download away team logo for match %d: %v", apiMatch.ID, downloadErr)
				} else {
					awayLogoFilename = downloadedFilename // Store only the filename
				}
			}

			// Get Home Team ID, passing the local filename
			homeTeamID := sql.NullInt64{Valid: false}
			if apiMatch.HomeTeamName != "" {
				tID, getTeamIDErr := database.GetTeamIDByThaiName(db, apiMatch.HomeTeamName, homeLogoFilename) // Pass filename
				if getTeamIDErr != nil {
					log.Printf("Warning: Failed to get home team ID for match %d (%s): %v", apiMatch.ID, apiMatch.HomeTeamName, getTeamIDErr)
				} else {
					homeTeamID = sql.NullInt64{Int64: int64(tID), Valid: true}
				}
			}

			// Get Away Team ID, passing the local filename
			awayTeamID := sql.NullInt64{Valid: false}
			if apiMatch.AwayTeamName != "" {
				tID, getTeamIDErr := database.GetTeamIDByThaiName(db, apiMatch.AwayTeamName, awayLogoFilename) // Pass filename
				if getTeamIDErr != nil {
					log.Printf("Warning: Failed to get away team ID for match %d (%s): %v", apiMatch.ID, apiMatch.AwayTeamName, getTeamIDErr)
				} else {
					awayTeamID = sql.NullInt64{Int64: int64(tID), Valid: true}
				}
			}

			// Get Channel ID (Main TV), passing the local filename
			channelID := sql.NullInt64{Valid: false}
			if apiMatch.ChannelInfo.Name != "" {
				channelLogoFilename := ""
				if apiMatch.ChannelInfo.Logo != "" {
					var downloadErr error
					downloadedFilename, downloadErr := DownloadImage(apiMatch.ChannelInfo.Logo, "./img/source") // Assuming channel logos also go to img/source
					if downloadErr != nil {
						log.Printf("Warning: Failed to download channel logo for %s: %v", apiMatch.ChannelInfo.Name, downloadErr)
					} else {
						channelLogoFilename = downloadedFilename
					}
				}
				cID, getChannelIDErr := database.GetChannelID(db, apiMatch.ChannelInfo.Name, channelLogoFilename, "TV") // Pass filename
				if getChannelIDErr != nil {
					log.Printf("Warning: Failed to get channel ID for match %d (%s): %v", apiMatch.ID, apiMatch.ChannelInfo.Name, getChannelIDErr)
				} else {
					channelID = sql.NullInt64{Int64: int64(cID), Valid: true}
				}
			}

			// Get Live Channel ID, passing the local filename
			liveChannelID := sql.NullInt64{Valid: false}
			if apiMatch.LiveInfo.Name != "" {
				liveChannelLogoFilename := ""
				if apiMatch.LiveInfo.Logo != "" {
					var downloadErr error
					downloadedFilename, downloadErr := DownloadImage(apiMatch.LiveInfo.Logo, "./img/source") // Assuming live channel logos also go to img/source
					if downloadErr != nil {
						log.Printf("Warning: Failed to download live channel logo for %s: %v", apiMatch.LiveInfo.Name, downloadErr)
					} else {
						liveChannelLogoFilename = downloadedFilename
					}
				}
				lcID, getChannelIDErr := database.GetChannelID(db, apiMatch.LiveInfo.Name, liveChannelLogoFilename, "Live Stream") // Pass filename
				if getChannelIDErr != nil {
					log.Printf("Warning: Failed to get live channel ID for match %d (%s): %v", apiMatch.ID, apiMatch.LiveInfo.Name, getChannelIDErr)
				} else {
					liveChannelID = sql.NullInt64{Int64: int64(lcID), Valid: true}
				}
			}

			// Determine match status based on API's 'match_status'
			matchStatus := sql.NullString{Valid: false}
			if apiMatch.MatchStatus == "2" {
				matchStatus = sql.NullString{String: "FINISHED", Valid: true}
			} else if apiMatch.MatchStatus == "1" {
				matchStatus = sql.NullString{String: "FIXTURE", Valid: true}
			} else {
				matchStatus = sql.NullString{String: apiMatch.MatchStatus, Valid: true} // Use as is if not 1 or 2
			}

			// Prepare MatchDB struct
			matchDB := models.MatchDB{
				MatchRefID:    apiMatch.ID,
				StartDate:     apiMatch.StartDate,
				StartTime:     apiMatch.StartTime,
				LeagueID:      sql.NullInt64{Int64: int64(dbLeagueID), Valid: true}, // Use DB league ID from config
				HomeTeamID:    homeTeamID,
				AwayTeamID:    awayTeamID,
				ChannelID:     channelID,
				LiveChannelID: liveChannelID,
				HomeScore:     sql.NullInt64{Int64: int64(apiMatch.HomeGoalCount), Valid: true},
				AwayScore:     sql.NullInt64{Int64: int64(apiMatch.AwayGoalCount), Valid: true},
				MatchStatus:   matchStatus,
			}

			// Insert or Update match in DB
			err = database.InsertOrUpdateMatch(db, matchDB) // Use = for reassignment
			if err != nil {
				log.Printf("Error saving match %d to DB: %v", apiMatch.ID, err)
			}
		}
	}
	return nil
}

// ScrapeThaileagueMatches scrapes matches for Thai League (T1, T2, T3 Regions, Samipro)
// It now accepts a targetLeague parameter to scrape specific leagues or all.
// targetLeague can be "t1", "t2", "t3_BKK", "t3_EAST", "t3_WEST", "t3_NORTH", "t3_NORTHEAST", "t3_SOUTH", or "all" (or empty for all).
func ScrapeThaileagueMatches(db *sql.DB, targetLeague string) error {
	baseURL := "https://competition.tl.prod.c0d1um.io/thaileague/api/match-day-match-public/?page="

	switch targetLeague {
	case "t1":
		log.Println("Scraping Thai League 1 (T1) Matches...")
		if err := scrapeMatchesByConfig(db, baseURL, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}, "&tournament=207", "t1", 1); err != nil {
			return fmt.Errorf("failed to scrape T1 matches: %w", err)
		}
	case "t2":
		log.Println("Scraping Thai League 2 (T2) Matches...")
		if err := scrapeMatchesByConfig(db, baseURL, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, "&tournament=196", "t2", 2); err != nil {
			return fmt.Errorf("failed to scrape T2 matches: %w", err)
		}
	case "t3_BKK":
		log.Println("Scraping Thai League 3 (T3) BKK Region Matches...")
		if err := scrapeMatchesByConfig(db, baseURL, []int{1, 2, 3, 4, 5, 6, 7, 8}, "&stage=982&tournament=197&tournament_team=", "t3_BKK", 3); err != nil {
			return fmt.Errorf("failed to scrape T3 BKK matches: %w", err)
		}
	case "t3_EAST":
		log.Println("Scraping Thai League 3 (T3) EAST Region Matches...")
		if err := scrapeMatchesByConfig(db, baseURL, []int{1, 2, 3, 4, 5, 6, 7, 8}, "&stage=999&tournament=197&tournament_team=", "t3_EAST", 3); err != nil {
			return fmt.Errorf("failed to scrape T3 EAST matches: %w", err)
		}
	case "t3_WEST":
		log.Println("Scraping Thai League 3 (T3) WEST Region Matches...")
		if err := scrapeMatchesByConfig(db, baseURL, []int{1, 2, 3, 4, 5, 6, 7, 8}, "&stage=1000&tournament=197&tournament_team=", "t3_WEST", 3); err != nil {
			return fmt.Errorf("failed to scrape T3 WEST matches: %w", err)
		}
	case "t3_NORTH":
		log.Println("Scraping Thai League 3 (T3) NORTH Region Matches...")
		if err := scrapeMatchesByConfig(db, baseURL, []int{1, 2, 3, 4, 5, 6, 7, 8}, "&stage=981&tournament=197&tournament_team=", "t3_NORTH", 3); err != nil {
			return fmt.Errorf("failed to scrape T3 NORTH matches: %w", err)
		}
	case "t3_NORTHEAST":
		log.Println("Scraping Thai League 3 (T3) NORTHEAST Region Matches...")
		if err := scrapeMatchesByConfig(db, baseURL, []int{1, 2, 3, 4, 5, 6, 7, 8}, "&stage=998&tournament=197&tournament_team=", "t3_NORTHEAST", 3); err != nil {
			return fmt.Errorf("failed to scrape T3 NORTHEAST matches: %w", err)
		}
	case "t3_SOUTH":
		log.Println("Scraping Thai League 3 (T3) SOUTH Region Matches...")
		if err := scrapeMatchesByConfig(db, baseURL, []int{1, 2, 3, 4, 5, 6, 7, 8}, "&stage=1001&tournament=197&tournament_team=", "t3_SOUTH", 3); err != nil {
			return fmt.Errorf("failed to scrape T3 SOUTH matches: %w", err)
		}
	case "samipro":
		log.Println("Scraping Samipro Matches...")
		if err := scrapeMatchesByConfig(db, baseURL, []int{1, 2, 3, 4, 5}, "&tournament=206", "samipro", 59); err != nil {
			return fmt.Errorf("failed to scrape Samipro matches: %w", err)
		}
	case "", "all": // Default to all if no specific league or "all" is provided
		log.Println("Scraping ALL Thai League Matches (T1, T2, T3 Regions, Samipro)...")
		if err := scrapeMatchesByConfig(db, baseURL, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}, "&tournament=207", "t1", 1); err != nil {
			log.Printf("Error scraping T1 matches: %v", err)
		}
		if err := scrapeMatchesByConfig(db, baseURL, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, "&tournament=196", "t2", 2); err != nil {
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
			if err := scrapeMatchesByConfig(db, baseURL, []int{1, 2, 3, 4, 5, 6, 7, 8}, param, "t3_"+region, 3); err != nil {
				log.Printf("Error scraping T3 %s matches: %v", region, err)
			}
		}
		if err := scrapeMatchesByConfig(db, baseURL, []int{1, 2, 3, 4, 5}, "&tournament=206", "samipro", 59); err != nil {
			log.Printf("Error scraping Samipro matches: %v", err)
		}
	default:
		return fmt.Errorf("invalid target league specified: %s", targetLeague)
	}

	return nil
}

// ScrapeBallthaiCupMatches scrapes matches for various cups (Revo, FA, BGC)
func ScrapeBallthaiCupMatches(db *sql.DB) error {
	baseURL := "https://competition.tl.prod.c0d1um.io/thaileague/api/match-day-match-public/?page="

	// Revo League Cup
	if err := scrapeMatchesByConfig(db, baseURL, []int{1, 2, 3, 4, 5, 6, 7, 8}, "&tournament=202", "revo", 4); err != nil { // Map to your DB league ID for Revo
		return fmt.Errorf("failed to scrape Revo Cup matches: %w", err)
	}
	// FA Cup
	if err := scrapeMatchesByConfig(db, baseURL, []int{1, 2, 3, 4, 5, 6, 7, 8}, "&tournament=199", "fa", 5); err != nil { // Map to your DB league ID for FA
		return fmt.Errorf("failed to scrape FA Cup matches: %w", err)
	}
	// BGC Cup
	if err := scrapeMatchesByConfig(db, baseURL, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, "&tournament=205", "bgc", 6); err != nil { // Map to your DB league ID for BGC
		return fmt.Errorf("failed to scrape BGC Cup matches: %w", err)
	}

	return nil
}

// ScrapeThaileaguePlayoffMatches scrapes playoff matches for Thai League
func ScrapeThaileaguePlayoffMatches(db *sql.DB) error {
	var err error // Declare err once at the function level
	const playoffLeagueID = 7 // Define a constant for playoff league ID

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
		// Note: Playoff URLs do not have a page= parameter like others, so send an empty slice of pages
		err = FetchAndParseAPI(url, &apiResponse) // Use = for reassignment
		if err != nil {
			log.Printf("Error fetching playoff matches from %s: %v", url, err)
			continue
		}

		for _, apiMatch := range apiResponse.Results {
			// Download home team logo and get only the filename
			homeLogoFilename := ""
			if apiMatch.HomeTeamLogo != "" {
				var downloadErr error // Declare a new error variable for DownloadImage
				downloadedFilename, downloadErr := DownloadImage(apiMatch.HomeTeamLogo, "./img/source")
				if downloadErr != nil {
					log.Printf("Warning: Failed to download home team logo for match %d: %v", apiMatch.ID, downloadErr)
				} else {
					homeLogoFilename = downloadedFilename // Store only the filename
				}
			}

			// Download away team logo and get only the filename
			awayLogoFilename := ""
			if apiMatch.AwayTeamLogo != "" {
				var downloadErr error // Declare a new error variable for DownloadImage
				downloadedFilename, downloadErr := DownloadImage(apiMatch.AwayTeamLogo, "./img/source")
				if downloadErr != nil {
					log.Printf("Warning: Failed to download away team logo for match %d: %v", apiMatch.ID, downloadErr)
				} else {
					awayLogoFilename = downloadedFilename // Store only the filename
				}
			}

			// Get Home Team ID, passing the local filename
			homeTeamID := sql.NullInt64{Valid: false}
			if apiMatch.HomeTeamName != "" {
				tID, getTeamIDErr := database.GetTeamIDByThaiName(db, apiMatch.HomeTeamName, homeLogoFilename) // Pass filename
				if getTeamIDErr != nil {
					log.Printf("Warning: Failed to get home team ID for match %d (%s): %v", apiMatch.ID, apiMatch.HomeTeamName, getTeamIDErr)
				} else {
					homeTeamID = sql.NullInt64{Int64: int64(tID), Valid: true}
				}
			}

			// Get Away Team ID, passing the local filename
			awayTeamID := sql.NullInt64{Valid: false}
			if apiMatch.AwayTeamName != "" {
				tID, getTeamIDErr := database.GetTeamIDByThaiName(db, apiMatch.AwayTeamName, awayLogoFilename) // Pass filename
				if getTeamIDErr != nil {
					log.Printf("Warning: Failed to get away team ID for match %d (%s): %v", apiMatch.ID, apiMatch.AwayTeamName, getTeamIDErr)
				} else {
					awayTeamID = sql.NullInt64{Int64: int64(tID), Valid: true}
				}
			}

			// Get Channel ID (Main TV), passing the local filename
			channelID := sql.NullInt64{Valid: false}
			if apiMatch.ChannelInfo.Name != "" {
				channelLogoFilename := ""
				if apiMatch.ChannelInfo.Logo != "" {
					var downloadErr error
					downloadedFilename, downloadErr := DownloadImage(apiMatch.ChannelInfo.Logo, "./img/source") // Assuming channel logos also go to img/source
					if downloadErr != nil {
						log.Printf("Warning: Failed to download channel logo for %s: %v", apiMatch.ChannelInfo.Name, downloadErr)
					} else {
						channelLogoFilename = downloadedFilename
					}
				}
				cID, getChannelIDErr := database.GetChannelID(db, apiMatch.ChannelInfo.Name, channelLogoFilename, "TV") // Pass filename
				if getChannelIDErr != nil {
					log.Printf("Warning: Failed to get channel ID for match %d (%s): %v", apiMatch.ID, apiMatch.ChannelInfo.Name, getChannelIDErr)
				} else {
					channelID = sql.NullInt64{Int64: int64(cID), Valid: true}
				}
			}

			// Get Live Channel ID, passing the local filename
			liveChannelID := sql.NullInt64{Valid: false}
			if apiMatch.LiveInfo.Name != "" {
				liveChannelLogoFilename := ""
				if apiMatch.LiveInfo.Logo != "" {
					var downloadErr error
					downloadedFilename, downloadErr := DownloadImage(apiMatch.LiveInfo.Logo, "./img/source") // Assuming live channel logos also go to img/source
					if downloadErr != nil {
						log.Printf("Warning: Failed to download live channel logo for %s: %v", apiMatch.LiveInfo.Name, downloadErr)
					} else {
						liveChannelLogoFilename = downloadedFilename
					}
				}
				lcID, getChannelIDErr := database.GetChannelID(db, apiMatch.LiveInfo.Name, liveChannelLogoFilename, "Live Stream") // Pass filename
				if getChannelIDErr != nil {
					log.Printf("Warning: Failed to get live channel ID for match %d (%s): %v", apiMatch.ID, apiMatch.LiveInfo.Name, getChannelIDErr)
				} else {
					liveChannelID = sql.NullInt64{Int64: int64(lcID), Valid: true}
				}
			}

			// Determine match status based on API's 'match_status'
			matchStatus := sql.NullString{Valid: false}
			if apiMatch.MatchStatus == "2" {
				matchStatus = sql.NullString{String: "FINISHED", Valid: true}
			} else if apiMatch.MatchStatus == "1" {
				matchStatus = sql.NullString{String: "FIXTURE", Valid: true}
			} else {
				matchStatus = sql.NullString{String: apiMatch.MatchStatus, Valid: true} // Use as is if not 1 or 2
			}

			// Prepare MatchDB struct
			matchDB := models.MatchDB{
				MatchRefID:    apiMatch.ID,
				StartDate:     apiMatch.StartDate,
				StartTime:     apiMatch.StartTime,
				LeagueID:      sql.NullInt64{Int64: playoffLeagueID, Valid: true}, // Use the constant for playoff league ID
				HomeTeamID:    homeTeamID,
				AwayTeamID:    awayTeamID,
				ChannelID:     channelID,
				LiveChannelID: liveChannelID,
				HomeScore:     sql.NullInt64{Int64: int64(apiMatch.HomeGoalCount), Valid: true},
				AwayScore:     sql.NullInt64{Int64: int64(apiMatch.AwayGoalCount), Valid: true},
				MatchStatus:   matchStatus,
			}

			// Insert or Update match in DB
			err = database.InsertOrUpdateMatch(db, matchDB) // Use = for reassignment
			if err != nil {
				log.Printf("Error saving playoff match %d to DB: %v", apiMatch.ID, err)
			}
		}
	}
	return nil
}
