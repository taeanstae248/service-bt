package database

import (
	"database/sql"
	"fmt"
	"log"

	"go-ballthai-scraper/models" // ตรวจสอบให้แน่ใจว่าชื่อโมดูลตรงกับ go.mod ของคุณ
)

// InsertOrUpdatePlayer inserts or updates a player record in the database
func InsertOrUpdatePlayer(db *sql.DB, player models.PlayerDB) error {
	var existingPlayerID int
	query := "SELECT id FROM players WHERE player_ref_id = ?"
	err := db.QueryRow(query, player.PlayerRefID).Scan(&existingPlayerID)

	if err == sql.ErrNoRows {
		// Insert new player
		insertQuery := `
			INSERT INTO players (
				player_ref_id, league_id, team_id, nationality_id,
				name, full_name_en, shirt_number, position, photo_url,
				matches_played, goals, yellow_cards, red_cards
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		_, err := db.Exec(insertQuery,
			player.PlayerRefID, player.LeagueID, player.TeamID, player.NationalityID,
			player.Name, player.FullNameEN, player.ShirtNumber, player.Position, player.PhotoURL,
			player.MatchesPlayed, player.Goals, player.YellowCards, player.RedCards,
		)
		if err != nil {
			return fmt.Errorf("failed to insert player %s: %w", player.Name, err)
		}
		log.Printf("Inserted new player: %s", player.Name)
	} else if err != nil {
		return fmt.Errorf("failed to query existing player %d: %w", player.PlayerRefID, err)
	} else {
		// Update existing player
		updateQuery := `
			UPDATE players SET
				league_id = ?, team_id = ?, nationality_id = ?,
				name = ?, full_name_en = ?, shirt_number = ?, position = ?, photo_url = ?,
				matches_played = ?, goals = ?, yellow_cards = ?, red_cards = ?
			WHERE player_ref_id = ?
		`
		_, err := db.Exec(updateQuery,
			player.LeagueID, player.TeamID, player.NationalityID,
			player.Name, player.FullNameEN, player.ShirtNumber, player.Position, player.PhotoURL,
			player.MatchesPlayed, player.Goals, player.YellowCards, player.RedCards,
			player.PlayerRefID,
		)
		if err != nil {
			return fmt.Errorf("failed to update player %d: %w", player.PlayerRefID, err)
		}
		log.Printf("Updated existing player: %s (ID: %d)", player.Name, existingPlayerID)
	}
	return nil
}
