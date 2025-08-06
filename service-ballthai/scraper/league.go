package scraper

import (
	"database/sql"
	"log"
)

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
