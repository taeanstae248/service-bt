package database

import (
	"database/sql"
	"fmt"
	"log"

	"go-ballthai-scraper/models" // Ensure this module name matches your go.mod
)

// InsertOrUpdateStanding inserts or updates a standing record in the database.
func InsertOrUpdateStanding(db *sql.DB, standing models.StandingDB) error {
    var existingID int
    query := "SELECT id FROM league_standings WHERE league_id = ? AND team_id = ?"
    err := db.QueryRow(query, standing.LeagueID, standing.TeamID).Scan(&existingID)

    if err == sql.ErrNoRows {
        // Standing does not exist, insert a new one
        insertQuery := `
            INSERT INTO league_standings (
                league_id, team_id, round, matches_played, wins, draws, 
                losses, goals_for, goals_against, goal_difference, points, current_rank
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        `
        _, err = db.Exec(
            insertQuery,
            standing.LeagueID,
            standing.TeamID,
            standing.Round,
            standing.MatchesPlayed,
            standing.Wins,
            standing.Draws,
            standing.Losses,
            standing.GoalsFor,
            standing.GoalsAgainst,
            standing.GoalDifference,
            standing.Points,
            standing.CurrentRank,
        )
        if err != nil {
            return fmt.Errorf("failed to insert standing for league %d team %d: %w", standing.LeagueID, standing.TeamID, err)
        }
        log.Printf("Inserted new standing for league %d team %d", standing.LeagueID, standing.TeamID)
    } else if err != nil {
        return fmt.Errorf("error checking existing standing: %w", err)
    } else {
        // Standing exists, update the record
        updateQuery := `
            UPDATE league_standings SET
                round = ?, matches_played = ?, wins = ?, draws = ?, 
                losses = ?, goals_for = ?, goals_against = ?, 
                goal_difference = ?, points = ?, current_rank = ?
            WHERE id = ?
        `
        _, err = db.Exec(
            updateQuery,
            standing.Round,
            standing.MatchesPlayed,
            standing.Wins,
            standing.Draws,
            standing.Losses,
            standing.GoalsFor,
            standing.GoalsAgainst,
            standing.GoalDifference,
            standing.Points,
            standing.CurrentRank,
            existingID,
        )
        if err != nil {
            return fmt.Errorf("failed to update standing for league %d team %d (ID: %d): %w", standing.LeagueID, standing.TeamID, existingID, err)
        }
        log.Printf("Updated standing for league %d team %d (ID: %d)", standing.LeagueID, standing.TeamID, existingID)
    }
    return nil
}
