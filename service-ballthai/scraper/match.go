package scraper

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"go-ballthai-scraper/database" // Ensure this module name matches your go.mod
	"go-ballthai-scraper/models"  // Ensure this module name matches your go.mod
)

// scrapeMatchesByConfig is a generic function to handle various match scraping configurations
// It now dynamically gets the dbLeagueID based on leagueType.
func scrapeMatchesByConfig(db *sql.DB, baseURL string, pages []int, tournamentParam string, leagueType string) error {
	var err error // Declare err once at the function level

	// Get or insert league ID dynamically from the database
	dbLeagueID, err := database.GetLeagueID(db, leagueType)
	if err != nil {
		return fmt.Errorf("failed to get or insert league ID for %s: %w", leagueType, err)
	}

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
			// Declare filename variables at the beginning of the loop for wider scope
			var homeLogoFilename string
			var awayLogoFilename string
			var channelLogoFilename string
			var liveChannelLogoFilename string

			// Download home team logo and get only the filename
			if apiMatch.HomeTeamLogo != "" {
				var downloadErr error // Declare a new error variable for DownloadImage
				homeLogoFilename, downloadErr = DownloadImage(apiMatch.HomeTeamLogo, "./img/source")
				if downloadErr != nil {
					log.Printf("Warning: Failed to download home team logo for match %d: %v", apiMatch.ID, downloadErr)
				}
			}

			// Download away team logo and get only the filename
			if apiMatch.AwayTeamLogo != "" {
				var downloadErr error // Declare a new error variable for DownloadImage
				awayLogoFilename, downloadErr = DownloadImage(apiMatch.AwayTeamLogo, "./img/source")
				if downloadErr != nil {
					log.Printf("Warning: Failed to download away team logo for match %d: %v", apiMatch.ID, downloadErr)
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
				var downloadErr error
				channelLogoFilename, downloadErr = DownloadImage(apiMatch.ChannelInfo.Logo, "./img/source") // Assuming channel logos also go to img/source
				if downloadErr != nil {
					log.Printf("Warning: Failed to download channel logo for %s: %v", apiMatch.ChannelInfo.Name, downloadErr)
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
				var downloadErr error
				liveChannelLogoFilename, downloadErr = DownloadImage(apiMatch.LiveInfo.Logo, "./img/source") // Assuming live channel logos also go to img/source
				if downloadErr != nil {
					log.Printf("Warning: Failed to download live channel logo for %s: %v", apiMatch.LiveInfo.Name, downloadErr)
				}
				lcID, getChannelIDErr := database.GetChannelID(db, apiMatch.LiveInfo.Name, liveChannelLogoFilename, "Live Stream") // Pass filename
				if getChannelIDErr != nil {
					log.Printf("Warning: Failed to get live channel ID for match %d (%s): %v", apiMatch.ID, apiMatch.LiveInfo.Name, getChannelIDErr)
				} else {
					liveChannelID = sql.NullInt64{Int64: int64(lcID), Valid: true}
				}
			}

			// Determine match status based on API's 'match_status' (now interface{})
			matchStatusStr := ""
			switch v := apiMatch.MatchStatus.(type) {
			case float64: // API might send numbers as float64 when unmarshaling into interface{}
				matchStatusInt := int(v)
				if matchStatusInt == 2 {
					matchStatusStr = "FINISHED"
				} else if matchStatusInt == 1 {
					matchStatusStr = "FIXTURE"
				} else {
					matchStatusStr = strconv.Itoa(matchStatusInt) // Convert other int statuses to string
				}
			case string:
				matchStatusStr = v // Directly use the string from API
			default:
				log.Printf("Warning: Unexpected type for MatchStatus: %T, Value: %v", v, v)
				matchStatusStr = fmt.Sprintf("%v", v) // Fallback to string representation of whatever it is
			}
			
			// Prepare MatchDB struct
			matchDB := models.MatchDB{
				MatchRefID:    apiMatch.ID,
				StartDate:     apiMatch.StartDate,
				StartTime:     apiMatch.StartTime,
				LeagueID:      sql.NullInt64{Int64: int64(dbLeagueID), Valid: true}, // Use dynamically obtained DB league ID
				HomeTeamID:    homeTeamID,
				AwayTeamID:    awayTeamID,
				ChannelID:     channelID,
				LiveChannelID: liveChannelID,
				HomeScore:     sql.NullInt64{Int64: int64(apiMatch.HomeGoalCount), Valid: true},
				AwayScore:     sql.NullInt64{Int64: int64(apiMatch.AwayGoalCount), Valid: true},
				MatchStatus:   sql.NullString{String: matchStatusStr, Valid: true}, // Store as NullString
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
// targetLeague can be "t1", "t2", "t3_BKK", "t3_EAST", "t3_WEST", "t3_NORTH", "t3_NORTHEAST", "t3_SOUTH", "samipro", or "all" (or empty for all).
func ScrapeThaileagueMatches(db *sql.DB, targetLeague string) error {
	baseURL := "https://competition.tl.prod.c0d1um.io/thaileague/api/match-day-match-public/?page="

	// Use []int{1} to only fetch the first page for now to avoid 404s on subsequent pages.
	// You might need to implement more sophisticated page iteration logic later.
	singlePage := []int{1} 

	switch targetLeague {
	case "t1":
		log.Println("Scraping Thai League 1 (T1) Matches...")
		if err := scrapeMatchesByConfig(db, baseURL, singlePage, "&tournament=207", "T1"); err != nil { // Pass "T1" as league name
			return fmt.Errorf("failed to scrape T1 matches: %w", err)
		}
	case "t2":
		log.Println("Scraping Thai League 2 (T2) Matches...")
		if err := scrapeMatchesByConfig(db, baseURL, singlePage, "&tournament=196", "T2"); err != nil { // Pass "T2" as league name
			return fmt.Errorf("failed to scrape T2 matches: %w", err)
		}
	case "t3_BKK":
		log.Println("Scraping Thai League 3 (T3) BKK Region Matches...")
		if err := scrapeMatchesByConfig(db, baseURL, singlePage, "&stage=982&tournament=197&tournament_team=", "T3 Bangkok"); err != nil { // Pass specific T3 name
			return fmt.Errorf("failed to scrape T3 BKK matches: %w", err)
		}
	case "t3_EAST":
		log.Println("Scraping Thai League 3 (T3) EAST Region Matches...")
		if err := scrapeMatchesByConfig(db, baseURL, singlePage, "&stage=999&tournament=197&tournament_team=", "T3 East"); err != nil {
			return fmt.Errorf("failed to scrape T3 EAST matches: %w", err)
		}
	case "t3_WEST":
		log.Println("Scraping Thai League 3 (T3) WEST Region Matches...")
		if err := scrapeMatchesByConfig(db, baseURL, singlePage, "&stage=1000&tournament=197&tournament_team=", "T3 West"); err != nil {
			return fmt.Errorf("failed to scrape T3 WEST matches: %w", err)
		}
	case "t3_NORTH":
		log.Println("Scraping Thai League 3 (T3) NORTH Region Matches...")
		if err := scrapeMatchesByConfig(db, baseURL, singlePage, "&stage=981&tournament=197&tournament_team=", "T3 North"); err != nil {
			return fmt.Errorf("failed to scrape T3 NORTH matches: %w", err)
		}
	case "t3_NORTHEAST":
		log.Println("Scraping Thai League 3 (T3) NORTHEAST Region Matches...")
		if err := scrapeMatchesByConfig(db, baseURL, singlePage, "&stage=998&tournament=197&tournament_team=", "T3 Northeast"); err != nil {
			return fmt.Errorf("failed to scrape T3 NORTHEAST matches: %w", err)
		}
	case "t3_SOUTH":
		log.Println("Scraping Thai League 3 (T3) SOUTH Region Matches...")
		if err := scrapeMatchesByConfig(db, baseURL, singlePage, "&stage=1001&tournament=197&tournament_team=", "T3 South"); err != nil {
			return fmt.Errorf("failed to scrape T3 SOUTH matches: %w", err)
		}
	case "samipro":
		log.Println("Scraping Samipro Matches...")
		if err := scrapeMatchesByConfig(db, baseURL, singlePage, "&tournament=206", "Samipro"); err != nil { // Pass "Samipro" as league name
			return fmt.Errorf("failed to scrape Samipro matches: %w", err)
		}
	case "", "all": // Default to all if no specific league or "all" is provided
		log.Println("Scraping ALL Thai League Matches (T1, T2, T3 Regions, Samipro)...")
		// Call all of them, but still limit to single page for now
		if err := scrapeMatchesByConfig(db, baseURL, singlePage, "&tournament=207", "T1"); err != nil {
			log.Printf("Error scraping T1 matches: %v", err)
		}
		if err := scrapeMatchesByConfig(db, baseURL, singlePage, "&tournament=196", "T2"); err != nil {
			log.Printf("Error scraping T2 matches: %v", err)
		}
		t3Stages := map[string]string{
			"BKK":       "T3 Bangkok",
			"EAST":      "T3 East",
			"WEST":      "T3 West",
			"NORTH":     "T3 North",
			"NORTHEAST": "T3 Northeast",
			"SOUTH":     "T3 South",
		}
		for region, name := range t3Stages {
			param := fmt.Sprintf("&stage=%s&tournament=197&tournament_team=", getT3StageID(region)) // Helper to get stage ID
			if err := scrapeMatchesByConfig(db, baseURL, singlePage, param, name); err != nil {
				log.Printf("Error scraping T3 %s matches: %v", region, err)
			}
		}
		if err := scrapeMatchesByConfig(db, baseURL, singlePage, "&tournament=206", "Samipro"); err != nil {
			log.Printf("Error scraping Samipro matches: %v", err)
		}
	default:
		return fmt.Errorf("invalid target league specified: %s", targetLeague)
	}

	return nil
}

// Helper function to get T3 stage ID based on region name (as seen in PHP)
func getT3StageID(region string) string {
	switch region {
	case "BKK": return "982"
	case "EAST": return "999"
	case "WEST": return "1000"
	case "NORTH": return "981"
	case "NORTHEAST": return "998"
	case "SOUTH": return "1001"
	default: return ""
	}
}


// ScrapeBallthaiCupMatches scrapes matches for various cups (Revo, FA, BGC)
func ScrapeBallthaiCupMatches(db *sql.DB) error {
	baseURL := "https://competition.tl.prod.c0d1um.io/thaileague/api/match-day-match-public/?page="
	singlePage := []int{1} // Limit to single page for now

	// Revo League Cup
	if err := scrapeMatchesByConfig(db, baseURL, singlePage, "&tournament=202", "Revo League Cup"); err != nil { // Pass "Revo League Cup" as league name
		return fmt.Errorf("failed to scrape Revo Cup matches: %w", err)
	}
	// FA Cup
	if err := scrapeMatchesByConfig(db, baseURL, singlePage, "&tournament=199", "FA Cup"); err != nil { // Pass "FA Cup" as league name
		return fmt.Errorf("failed to scrape FA Cup matches: %w", err)
	}
	// BGC Cup
	if err := scrapeMatchesByConfig(db, baseURL, singlePage, "&tournament=205", "BGC Cup"); err != nil { // Pass "BGC Cup" as league name
		return fmt.Errorf("failed to scrape BGC Cup matches: %w", err)
	}

	return nil
}

// ScrapeThaileaguePlayoffMatches scrapes playoff matches for Thai League
func ScrapeThaileaguePlayoffMatches(db *sql.DB) error {
	var err error // Declare err once at the function level
	const playoffLeagueName = "Thai League Playoff" // Define a constant for playoff league name

	// Get or insert playoff league ID dynamically
	playoffLeagueID, err := database.GetLeagueID(db, playoffLeagueName)
	if err != nil {
		return fmt.Errorf("failed to get or insert league ID for %s: %w", playoffLeagueName, err)
	}

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
			// Declare filename variables at the beginning of the loop for wider scope
			var homeLogoFilename string
			var awayLogoFilename string
			var channelLogoFilename string
			var liveChannelLogoFilename string

			// Download home team logo and get only the filename
			if apiMatch.HomeTeamLogo != "" {
				var downloadErr error
				homeLogoFilename, downloadErr = DownloadImage(apiMatch.HomeTeamLogo, "./img/source")
				if downloadErr != nil {
					log.Printf("Warning: Failed to download home team logo for match %d: %v", apiMatch.ID, downloadErr)
				}
			}

			// Download away team logo and get only the filename
			if apiMatch.AwayTeamLogo != "" {
				var downloadErr error
				awayLogoFilename, downloadErr = DownloadImage(apiMatch.AwayTeamLogo, "./img/source")
				if downloadErr != nil {
					log.Printf("Warning: Failed to download away team logo for match %d: %v", apiMatch.ID, downloadErr)
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
				var downloadErr error
				channelLogoFilename, downloadErr = DownloadImage(apiMatch.ChannelInfo.Logo, "./img/source") // Assuming channel logos also go to img/source
				if downloadErr != nil {
					log.Printf("Warning: Failed to download channel logo for %s: %v", apiMatch.ChannelInfo.Name, downloadErr)
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
				var downloadErr error
				liveChannelLogoFilename, downloadErr = DownloadImage(apiMatch.LiveInfo.Logo, "./img/source") // Assuming live channel logos also go to img/source
				if downloadErr != nil {
					log.Printf("Warning: Failed to download live channel logo for %s: %v", apiMatch.LiveInfo.Name, downloadErr)
				}
				lcID, getChannelIDErr := database.GetChannelID(db, apiMatch.LiveInfo.Name, liveChannelLogoFilename, "Live Stream") // Pass filename
				if getChannelIDErr != nil {
					log.Printf("Warning: Failed to get live channel ID for match %d (%s): %v", apiMatch.ID, apiMatch.LiveInfo.Name, getChannelIDErr)
				} else {
					liveChannelID = sql.NullInt64{Int64: int64(lcID), Valid: true}
				}
			}

			// Determine match status based on API's 'match_status' (now interface{})
			matchStatusStr := ""
			switch v := apiMatch.MatchStatus.(type) {
			case float64: // API might send numbers as float64 when unmarshaling into interface{}
				matchStatusInt := int(v)
				if matchStatusInt == 2 {
					matchStatusStr = "FINISHED"
				} else if matchStatusInt == 1 {
					matchStatusStr = "FIXTURE"
				} else {
					matchStatusStr = strconv.Itoa(matchStatusInt) // Convert other int statuses to string
				}
			case string:
				matchStatusStr = v // Directly use the string from API
			default:
				log.Printf("Warning: Unexpected type for MatchStatus: %T, Value: %v", v, v)
				matchStatusStr = fmt.Sprintf("%v", v) // Fallback to string representation of whatever it is
			}
			
			// Prepare MatchDB struct
			matchDB := models.MatchDB{
				MatchRefID:    apiMatch.ID,
				StartDate:     apiMatch.StartDate,
				StartTime:     apiMatch.StartTime,
				LeagueID:      sql.NullInt64{Int64: int64(playoffLeagueID), Valid: true}, // Corrected: Use playoffLeagueID here
				HomeTeamID:    homeTeamID,
				AwayTeamID:    awayTeamID,
				ChannelID:     channelID,
				LiveChannelID: liveChannelID,
				HomeScore:     sql.NullInt64{Int64: int64(apiMatch.HomeGoalCount), Valid: true},
				AwayScore:     sql.NullInt64{Int64: int64(apiMatch.AwayGoalCount), Valid: true},
				MatchStatus:   sql.NullString{String: matchStatusStr, Valid: true}, // Store as NullString
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
