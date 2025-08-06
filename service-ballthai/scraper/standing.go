package scraper

import (
	"database/sql"
	"log"
	// "fmt" // Removed: fmt is not directly used in this file
	// The 'fmt' package was imported but not used, causing a compile error.
	// It has been removed.

	"go-ballthai-scraper/database" // Ensure this module name matches your go.mod
	"go-ballthai-scraper/models"  // Ensure this module name matches your go.mod
)

// ScrapeStandings scrapes league standings data from the API and saves it to the database
func ScrapeStandings(db *sql.DB) error {
	// Map PHP's $_GET['table'] values to API URLs and DB league IDs
	standingConfigs := map[string]struct {
		URL      string
		LeagueID int // Corresponds to your DB league ID
		IsT3     bool
	}{
		"t1": {URL: "https://competition.tl.prod.c0d1um.io/thaileague/api/stage-standing-public/?tournament=207", LeagueID: 1}, // Example: map to DB league ID 1
		"t2": {URL: "https://competition.tl.prod.c0d1um.io/thaileague/api/stage-standing-public/?tournament=196", LeagueID: 2}, // Example: map to DB league ID 2
		"t3": {URL: "https://competition.tl.prod.c0d1um.io/thaileague/api/stage-standing-public/?tournament=197", LeagueID: 3, IsT3: true}, // Example: map to DB league ID 3, special handling for T3
		"samipro": {URL: "https://competition.tl.prod.c0d1um.io/thaileague/api/stage-standing-public/dashboard/?tournament=206", LeagueID: 59}, // Example: map to DB league ID 59
		"revo": {URL: "https://competition.tl.prod.c0d1um.io/thaileague/api/stage-standing-public/?tournament=155", LeagueID: 4}, // Example: map to DB league ID 4
		// Add more configurations as needed
	}

	for configName, config := range standingConfigs {
		log.Printf("Scraping standings for %s (%s)", configName, config.URL)
		
		var apiResponse []models.StandingAPI // API for standings returns an array directly
		err := FetchAndParseAPI(config.URL, &apiResponse)
		if err != nil {
			log.Printf("Error fetching standings for %s: %v", configName, err)
			continue
		}

		// Special handling for T3 (SOUTH stage) as seen in PHP
		if config.IsT3 {
			filteredResponse := []models.StandingAPI{}
			for _, s := range apiResponse {
				if s.StageName == "SOUTH" {
					filteredResponse = append(filteredResponse, s)
				}
			}
			apiResponse = filteredResponse
		}

		for _, apiStanding := range apiResponse {
			// Get Team ID
			teamID := 0
			tID, err := database.GetTeamIDByThaiName(db, apiStanding.TournamentTeamName, apiStanding.TournamentTeamLogo)
			if err != nil {
				log.Printf("Warning: Failed to get team ID for standing team %s: %v", apiStanding.TournamentTeamName, err)
				continue // Skip if team ID cannot be resolved
			}
			teamID = tID

			// Prepare StandingDB struct
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
				// Round is often not directly in standing API, might need to be derived or set to NULL
				Round:          sql.NullInt64{Valid: false}, // Default to null if not available
			}

			// Insert or Update standing in DB
			err = database.InsertOrUpdateStanding(db, standingDB)
			if err != nil {
				log.Printf("Error saving standing for team %s in league %s to DB: %v", apiStanding.TournamentTeamName, configName, err)
			}
		}
	}
	return nil
}
