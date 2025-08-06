package database

import (
	"database/sql"
	"fmt"
	"log"

	"go-scraper/models" // ตรวจสอบให้แน่ใจว่าชื่อโมดูลตรงกับ go.mod ของคุณ
)

// InsertOrUpdateMatch inserts or updates a match record in the database
func InsertOrUpdateMatch(db *sql.DB, match models.MatchDB) error {
	var existingMatchID int
	query := "SELECT id FROM matches WHERE match_ref_id = ?"
	err := db.QueryRow(query, match.MatchRefID).Scan(&existingMatchID)

	if err == sql.ErrNoRows {
		// Insert new match
		insertQuery := `
			INSERT INTO matches (
				match_ref_id, start_date, start_time, league_id,
				home_team_id, away_team_id, channel_id, live_channel_id,
				home_score, away_score, match_status
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		_, err := db.Exec(insertQuery,
			match.MatchRefID, match.StartDate, match.StartTime, match.LeagueID,
			match.HomeTeamID, match.AwayTeamID, match.ChannelID, match.LiveChannelID,
			match.HomeScore, match.AwayScore, match.MatchStatus,
		)
		if err != nil {
			return fmt.Errorf("failed to insert match %d: %w", match.MatchRefID, err)
		}
		log.Printf("Inserted new match: %d", match.MatchRefID)
	} else if err != nil {
		return fmt.Errorf("failed to query existing match %d: %w", match.MatchRefID, err)
	} else {
		// Update existing match
		updateQuery := `
			UPDATE matches SET
				start_date = ?, start_time = ?, league_id = ?,
				home_team_id = ?, away_team_id = ?, channel_id = ?, live_channel_id = ?,
				home_score = ?, away_score = ?, match_status = ?
			WHERE match_ref_id = ?
		`
		_, err := db.Exec(updateQuery,
			match.StartDate, match.StartTime, match.LeagueID,
			match.HomeTeamID, match.AwayTeamID, match.ChannelID, match.LiveChannelID,
			match.HomeScore, match.AwayScore, match.MatchStatus,
			match.MatchRefID,
		)
		if err != nil {
			return fmt.Errorf("failed to update match %d: %w", match.MatchRefID, err)
		}
		log.Printf("Updated existing match: %d (ID: %d)", match.MatchRefID, existingMatchID)
	}
	return nil
}
