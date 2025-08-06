package scraper

import (
	"database/sql"
	"fmt"
	"log"
	// "path/filepath" // Not directly used in this file

	"go-ballthai-scraper/database" // Ensure this module name matches your go.mod
	"go-ballthai-scraper/models"  // Ensure this module name matches your go.mod
)

// ScrapePlayers scrapes player data from the API and saves it to the database
// This function combines logic from Scrape_R*_Player and Player_Public PHP functions
func ScrapePlayers(db *sql.DB) error {
	baseAPIURL := "https://competition.tl.prod.c0d1um.io/thaileague/api/player-public/all_players_search/"

	// Define specific tournaments/team IDs as seen in PHP files
	// This is a simplified example, you might need to map tournaments to leagues
	// and handle specific team IDs as in PHP's hardcoded logic.
	playerConfigs := []struct {
		TournamentID int
		TeamIDs      []int
		LeagueID     int // Corresponds to your DB league ID
	}{
		{TournamentID: 195, TeamIDs: []int{6025, 6040, 6039, 6038, 6037, 6036, 6035, 6034, 6033, 6032, 6031, 6030, 6029, 6028, 6027, 6026}, LeagueID: 1}, // Example for R1, map to your actual league ID
		{TournamentID: 196, TeamIDs: []int{6058, 6057, 6056, 6055, 6054, 6053, 6052, 6051, 6050, 6049, 6048, 6047, 6046, 6045, 6044, 6043, 6042, 6041}, LeagueID: 2}, // Example for R2
		// Add more configurations as needed, mapping to your DB league IDs
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
				// Download player photo and get only the filename
				photoFilename := ""
				if apiPlayer.Photo != "" {
					downloadedFilename, err := DownloadImage(apiPlayer.Photo, "./img/player")
					if err != nil {
						log.Printf("Warning: Failed to download player photo for %s: %v", apiPlayer.FullName, err)
					} else {
						photoFilename = downloadedFilename // Store only the filename
					}
				}

				// Get Nationality ID
				nationalityID := sql.NullInt64{Valid: false}
				if apiPlayer.Nationality.Code != "" {
					nID, err := database.GetNationalityID(db, apiPlayer.Nationality.Code, apiPlayer.Nationality.FullName)
					if err != nil {
						log.Printf("Warning: Failed to get nationality ID for %s: %v", apiPlayer.Nationality.FullName, err)
					} else {
						nationalityID = sql.NullInt64{Int64: int64(nID), Valid: true}
					}
				}

				// Get Team ID (from club_name)
				playerTeamID := sql.NullInt64{Valid: false}
				if apiPlayer.ClubName != "" {
					// Pass empty logo for now, as player API might not provide team logo
					tID, err := database.GetTeamIDByThaiName(db, apiPlayer.ClubName, "") 
					if err != nil {
						log.Printf("Warning: Failed to get team ID for player %s's club %s: %v", apiPlayer.FullName, apiPlayer.ClubName, err)
					} else {
						playerTeamID = sql.NullInt64{Int64: int64(tID), Valid: true}
					}
				}
				// Handle specific team ID overrides as seen in PHP (if necessary)
				// e.g., if apiPlayer.ClubName == "นครศรี ยูไนเต็ด" { playerTeamID = sql.NullInt64{Int64: 923, Valid: true} }

				// Prepare PlayerDB struct
				playerDB := models.PlayerDB{
					PlayerRefID:   sql.NullInt64{Int64: int64(apiPlayer.ID), Valid: true},
					LeagueID:      sql.NullInt64{Int64: int64(config.LeagueID), Valid: true}, // Use league ID from config
					TeamID:        playerTeamID,
					NationalityID: nationalityID,
					Name:          apiPlayer.FullName,
					FullNameEN:    sql.NullString{String: apiPlayer.FullNameEN, Valid: apiPlayer.FullNameEN != ""},
					ShirtNumber:   apiPlayer.ShirtNumber,
					Position:      sql.NullString{String: apiPlayer.PositionShortName, Valid: apiPlayer.PositionShortName != ""},
					PhotoURL:      sql.NullString{String: photoFilename, Valid: photoFilename != ""}, // Store filename
					MatchesPlayed: apiPlayer.MatchCount,
					Goals:         apiPlayer.GoalFor,
					YellowCards:   apiPlayer.YellowCardAcc,
					RedCards:      apiPlayer.RedCardViolentConductAcc,
				}

				// Insert or Update player in DB
				err = database.InsertOrUpdatePlayer(db, playerDB)
				if err != nil {
					log.Printf("Error saving player %s to DB: %v", apiPlayer.FullName, err)
				}
			}
		}
	}
	return nil
}
