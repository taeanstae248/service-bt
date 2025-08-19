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
	var existingStatus int
	query := "SELECT id, status FROM players WHERE player_ref_id = ?"
	err := db.QueryRow(query, player.PlayerRefID).Scan(&existingPlayerID, &existingStatus)

	if err == sql.ErrNoRows {
		// Insert new player
		insertQuery := `
			INSERT INTO players (
				player_ref_id, league_id, team_id, nationality_id,
				name, full_name_en, shirt_number, position, photo_url,
				matches_played, goals, yellow_cards, red_cards, status
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		_, err := db.Exec(insertQuery,
			player.PlayerRefID, player.LeagueID, player.TeamID, player.NationalityID,
			player.Name, player.FullNameEN, player.ShirtNumber, player.Position, player.PhotoURL,
			player.MatchesPlayed, player.Goals, player.YellowCards, player.RedCards, player.Status,
		)
		if err != nil {
			return fmt.Errorf("failed to insert player %s: %w", player.Name, err)
		}
		log.Printf("Inserted new player: %s", player.Name)
	} else if err != nil {
		return fmt.Errorf("failed to query existing player %d: %w", player.PlayerRefID, err)
	} else {
		// ถ้า status = 1 ไม่ให้อัปเดต
		if existingStatus == 1 {
			log.Printf("Skip update player: %s (ID: %d) because status=1", player.Name, existingPlayerID)
			return nil
		}
		// Update existing player
		updateQuery := `
			UPDATE players SET
				league_id = ?, team_id = ?, nationality_id = ?,
				name = ?, full_name_en = ?, shirt_number = ?, position = ?, photo_url = ?,
				matches_played = ?, goals = ?, yellow_cards = ?, red_cards = ?, status = ?
			WHERE player_ref_id = ?
		`
		_, err := db.Exec(updateQuery,
			player.LeagueID, player.TeamID, player.NationalityID,
			player.Name, player.FullNameEN, player.ShirtNumber, player.Position, player.PhotoURL,
			player.MatchesPlayed, player.Goals, player.YellowCards, player.RedCards, player.Status,
			player.PlayerRefID,
		)
		if err != nil {
			return fmt.Errorf("failed to update player %d: %w", player.PlayerRefID, err)
		}
		log.Printf("Updated existing player: %s (ID: %d)", player.Name, existingPlayerID)
	}
	return nil
}
