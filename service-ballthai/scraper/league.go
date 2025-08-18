package scraper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"database/sql"
	"log"
)
// ScrapeAndSyncSeasonsFromAPI ดึงข้อมูลฤดูกาลจาก API แล้ว sync กับ DB
func ScrapeAndSyncSeasonsFromAPI(db *sql.DB) error {
	apiURL := "https://competition.tl.prod.c0d1um.io/thaileague/api/tournament-public/?latest_activated_season=True&show_in_public_website=True"
	resp, err := http.Get(apiURL)
	if err != nil {
		return fmt.Errorf("failed to fetch API: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read API response: %w", err)
	}

	var apiResp struct {
		Results []struct {
			ID               int    `json:"id"`
			Name             string `json:"name"`
			SeasonStartDate  string `json:"season_start_date"`
			SeasonEndDate    string `json:"season_end_date"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("failed to parse API JSON: %w", err)
	}

	for _, apiLeague := range apiResp.Results {
		// หา league_id ใน DB ที่ thaileageid ตรงกับ apiLeague.ID
		var leagueID int
		err := db.QueryRow("SELECT id FROM leagues WHERE thaileageid = ?", apiLeague.ID).Scan(&leagueID)
		if err == sql.ErrNoRows {
			log.Printf("[season-sync] No league found for thaileageid=%d", apiLeague.ID)
			continue
		} else if err != nil {
			log.Printf("[season-sync] DB error for thaileageid=%d: %v", apiLeague.ID, err)
			continue
		}

		// check ว่ามี season นี้อยู่แล้วหรือยัง (เช็คซ้ำด้วย name + league_id)
		var seasonID int
		err = db.QueryRow("SELECT id FROM seasons WHERE league_id = ? AND name = ?", leagueID, apiLeague.Name).Scan(&seasonID)
		if err == sql.ErrNoRows {
			// insert ใหม่
			_, err = db.Exec(`INSERT INTO seasons (league_id, name, season_start_date, season_end_date) VALUES (?, ?, ?, ?)`,
				leagueID, apiLeague.Name, apiLeague.SeasonStartDate, apiLeague.SeasonEndDate)
			if err != nil {
				log.Printf("[season-sync] Insert error for league_id=%d name=%s: %v", leagueID, apiLeague.Name, err)
				continue
			}
			log.Printf("[season-sync] Inserted season: league_id=%d name=%s", leagueID, apiLeague.Name)
		} else if err == nil {
			// update
			_, err = db.Exec(`UPDATE seasons SET season_start_date=?, season_end_date=? WHERE id=?`,
				apiLeague.SeasonStartDate, apiLeague.SeasonEndDate, seasonID)
			if err != nil {
				log.Printf("[season-sync] Update error for season_id=%d: %v", seasonID, err)
				continue
			}
			log.Printf("[season-sync] Updated season: id=%d name=%s", seasonID, apiLeague.Name)
		} else {
			log.Printf("[season-sync] Query error: %v", err)
		}
	}
	return nil
}

// ScrapeLeagues scrapes league data from BallThai API
func ScrapeLeagues(db *sql.DB) error {
	log.Println("Starting to scrape leagues...")

	// Common leagues in Thai football - insert basic data first
	leagues := []string{
		"Thai League 1",
		"Thai League 2",
		"Thai League 3",
		"FA Cup",
		"League Cup",
		"Thai Youth League",
		"Champions League",
		"AFC Cup",
	}

	// Insert basic leagues
	for _, leagueName := range leagues {
		query := `INSERT IGNORE INTO leagues (name) VALUES (?)`
		_, err := db.Exec(query, leagueName)
		if err != nil {
			log.Printf("Error inserting league %s: %v", leagueName, err)
			continue
		}
		log.Printf("League '%s' inserted successfully", leagueName)
	}

	log.Println("League scraping completed.")
	return nil
}

// ScrapeTeams scrapes team data from existing database and updates missing team info
func ScrapeTeams(db *sql.DB) error {
	log.Println("Starting to scrape teams...")

	// This function will be used to update teams with additional information
	// For now, we'll just ensure we have basic teams from matches

	// Update team post ballthai data
	err := UpdateTeamPostBallthai(db)
	if err != nil {
		log.Printf("Error updating team post ballthai: %v", err)
		return err
	}

	log.Println("Team scraping completed.")
	return nil
}
