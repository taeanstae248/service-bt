package database

import (
	"database/sql"
	"fmt"
	"go-ballthai-scraper/models"
	"log"
)

// GetStandingsByLeagueID คืน standings ทั้งหมดของลีกที่ระบุ
func GetStandingsByLeagueID(db *sql.DB, leagueID int) ([]models.StandingDB, error) {
	rows, err := db.Query(`SELECT id, league_id, team_id, round, matches_played, wins, draws, losses, goals_for, goals_against, goal_difference, points, current_rank FROM standings WHERE league_id = ? ORDER BY points DESC, goal_difference DESC, wins DESC`, leagueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var standings []models.StandingDB
	for rows.Next() {
		var s models.StandingDB
		if err := rows.Scan(&s.ID, &s.LeagueID, &s.TeamID, &s.Round, &s.MatchesPlayed, &s.Wins, &s.Draws, &s.Losses, &s.GoalsFor, &s.GoalsAgainst, &s.GoalDifference, &s.Points, &s.CurrentRank); err != nil {
			return nil, err
		}
		standings = append(standings, s)
	}
	return standings, nil
}

// InsertOrUpdateStanding inserts or updates a league standing record in the database
func InsertOrUpdateStanding(db *sql.DB, standing models.StandingDB) error {
	var existingStandingID int
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
