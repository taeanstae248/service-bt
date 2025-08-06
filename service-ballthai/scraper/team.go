package scraper

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"go-ballthai-scraper/models"
)

const idPostAPIURL = "https://serviceseoball.com/api/id_post.php"

// UpdateTeamPostBallthai fetches team post IDs from an external API
// and updates the local database.
func UpdateTeamPostBallthai(db *sql.DB) error {
	log.Println("Fetching team post IDs from API...")

	// 1. Fetch data from the API
	resp, err := http.Get(idPostAPIURL)
	if err != nil {
		return fmt.Errorf("failed to get data from API %s: %w", idPostAPIURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("api request failed with status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// 2. Unmarshal JSON response
	var apiResponse models.TeamPostAPIResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response: %w", err)
	}

	if len(apiResponse.Teams) == 0 {
		log.Println("No teams found in the API response. Nothing to update.")
		return nil
	}

	log.Printf("Found %d teams in API response. Starting database update...", len(apiResponse.Teams))

	// 3. Prepare SQL statement
	stmt, err := db.Prepare("UPDATE teams SET team_post_ballthai = ? WHERE name_th = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare update statement: %w", err)
	}
	defer stmt.Close()

	// 4. Iterate and update database
	updatedCount := 0
	for _, team := range apiResponse.Teams {
		if team.NameTh == "" || team.PostBallthaiID == "" {
			log.Printf("Skipping team with empty name or ID: %+v", team)
			continue
		}

		res, err := stmt.Exec(team.PostBallthaiID, team.NameTh)
		if err != nil {
			log.Printf("Error updating team '%s': %v", team.NameTh, err)
			continue // Continue to the next team
		}

		rowsAffected, _ := res.RowsAffected()
		if rowsAffected > 0 {
			updatedCount++
			log.Printf("Updated team_post_ballthai for team: %s", team.NameTh)
		}
	}

	log.Printf("Team post ID update process finished. Total teams updated: %d", updatedCount)
	return nil
}
