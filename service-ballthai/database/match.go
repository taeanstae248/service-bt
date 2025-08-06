package database

import (
	"database/sql"
	"fmt"
	"log"

	"go-ballthai-scraper/models" // Ensure this module name matches your go.mod
)

// InsertOrUpdateMatch inserts or updates a match record in the database.
func InsertOrUpdateMatch(db *sql.DB, match models.MatchDB) error {
	var existingID int
	query := "SELECT id FROM matches WHERE match_ref_id = ?"
	err := db.QueryRow(query, match.MatchRefID).Scan(&existingID)

	if err == sql.ErrNoRows {
		// Match does not exist, insert a new one
		insertQuery := `
            INSERT INTO matches (
                match_ref_id, start_date, start_time, league_id, home_team_id, 
                away_team_id, channel_id, live_channel_id, home_score, away_score, match_status
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        `
		_, err = db.Exec(
			insertQuery,
			match.MatchRefID,
			match.StartDate,
			match.StartTime,
			match.LeagueID,
			match.HomeTeamID,
			match.AwayTeamID,
			match.ChannelID,
			match.LiveChannelID,
			match.HomeScore,
			match.AwayScore,
			match.MatchStatus,
		)
		if err != nil {
			return fmt.Errorf("failed to insert match %d: %w", match.MatchRefID, err)
		}
		log.Printf("Inserted new match: %d", match.MatchRefID)
	} else if err != nil {
		return fmt.Errorf("error checking existing match: %w", err)
	} else {
		// Match exists, update the record
		updateQuery := `
            UPDATE matches SET
                start_date = ?, start_time = ?, league_id = ?, home_team_id = ?, 
                away_team_id = ?, channel_id = ?, live_channel_id = ?, 
                home_score = ?, away_score = ?, match_status = ?
            WHERE id = ?
        `
		_, err = db.Exec(
			updateQuery,
			match.StartDate,
			match.StartTime,
			match.LeagueID,
			match.HomeTeamID,
			match.AwayTeamID,
			match.ChannelID,
			match.LiveChannelID,
			match.HomeScore,
			match.AwayScore,
			match.MatchStatus,
			existingID,
		)
		if err != nil {
			return fmt.Errorf("failed to update match %d (ID: %d): %w", match.MatchRefID, existingID, err)
		}
		log.Printf("Updated match: %d (ID: %d)", match.MatchRefID, existingID)
	}
	return nil
}
