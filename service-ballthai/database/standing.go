package database

import (
	"database/sql"
	"fmt"
	"log"

	"go-ballthai-scraper/models" // ตรวจสอบให้แน่ใจว่าชื่อโมดูลตรงกับ go.mod ของคุณ
)

// InsertOrUpdateStanding inserts or updates a league standing record in the database
func InsertOrUpdateStanding(db *sql.DB, standing models.StandingDB) error {
	var existingStandingID int
	// Check for existing record based on league_id and team_id (UNIQUE constraint)
	query := "SELECT id FROM standings WHERE league_id = ? AND team_id = ?"
	err := db.QueryRow(query, standing.LeagueID, standing.TeamID).Scan(&existingStandingID)

	if err == sql.ErrNoRows {
		// Insert new standing
		insertQuery := `
			INSERT INTO standings (
				league_id, team_id, round, matches_played, wins, draws, losses,
				goals_for, goals_against, goal_difference, points, current_rank
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		_, err := db.Exec(insertQuery,
			standing.LeagueID, standing.TeamID, standing.Round, standing.MatchesPlayed,
			standing.Wins, standing.Draws, standing.Losses, standing.GoalsFor,
			standing.GoalsAgainst, standing.GoalDifference, standing.Points, standing.CurrentRank,
		)
		if err != nil {
			return fmt.Errorf("failed to insert standing for team %d in league %d: %w", standing.TeamID, standing.LeagueID, err)
		}
		log.Printf("Inserted new standing for team %d in league %d", standing.TeamID, standing.LeagueID)
	} else if err != nil {
		return fmt.Errorf("failed to query existing standing for team %d in league %d: %w", standing.TeamID, standing.LeagueID, err)
	} else {
		// Update existing standing
		updateQuery := `
			UPDATE standings SET
				round = ?, matches_played = ?, wins = ?, draws = ?, losses = ?,
				goals_for = ?, goals_against = ?, goal_difference = ?, points = ?, current_rank = ?
			WHERE id = ?
		`
		_, err := db.Exec(updateQuery,
			standing.Round, standing.MatchesPlayed, standing.Wins, standing.Draws, standing.Losses,
			standing.GoalsFor, standing.GoalsAgainst, standing.GoalDifference, standing.Points, standing.CurrentRank,
			existingStandingID,
		)
		if err != nil {
			return fmt.Errorf("failed to update standing for team %d in league %d: %w", standing.TeamID, standing.LeagueID, err)
		}
		log.Printf("Updated existing standing for team %d in league %d (ID: %d)", standing.TeamID, standing.LeagueID, existingStandingID)
	}
	return nil
}
