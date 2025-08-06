package database

import (
	"database/sql"
	"fmt"
	"log"

	"go-ballthai-scraper/models" // Ensure this module name matches your go.mod
)

// InsertOrUpdatePlayer inserts or updates a player record in the database.
func InsertOrUpdatePlayer(db *sql.DB, player models.PlayerDB) error {
    var existingID int
    query := "SELECT id FROM players WHERE player_ref_id = ?"
    err := db.QueryRow(query, player.PlayerRefID).Scan(&existingID)

    if err == sql.ErrNoRows {
        // Player does not exist, insert a new one
        insertQuery := `
            INSERT INTO players (
                player_ref_id, league_id, team_id, nationality_id, name, 
                full_name_en, shirt_number, position, photo_url, matches_played, 
                goals, yellow_cards, red_cards
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        `
        _, err = db.Exec(
            insertQuery,
            player.PlayerRefID,
            player.LeagueID,
            player.TeamID,
            player.NationalityID,
            player.Name,
            player.FullNameEN,
            player.ShirtNumber,
            player.Position,
            player.PhotoURL,
            player.MatchesPlayed,
            player.Goals,
            player.YellowCards,
            player.RedCards,
        )
        if err != nil {
            return fmt.Errorf("failed to insert player %s: %w", player.Name, err)
        }
        log.Printf("Inserted new player: %s", player.Name)
    } else if err != nil {
        return fmt.Errorf("error checking existing player: %w", err)
    } else {
        // Player exists, update the record
        updateQuery := `
            UPDATE players SET
                league_id = ?, team_id = ?, nationality_id = ?, name = ?, 
                full_name_en = ?, shirt_number = ?, position = ?, photo_url = ?, 
                matches_played = ?, goals = ?, yellow_cards = ?, red_cards = ?
            WHERE id = ?
        `
        _, err = db.Exec(
            updateQuery,
            player.LeagueID,
            player.TeamID,
            player.NationalityID,
            player.Name,
            player.FullNameEN,
            player.ShirtNumber,
            player.Position,
            player.PhotoURL,
            player.MatchesPlayed,
            player.Goals,
            player.YellowCards,
            player.RedCards,
            existingID,
        )
        if err != nil {
            return fmt.Errorf("failed to update player %s (ID: %d): %w", player.Name, existingID, err)
        }
        log.Printf("Updated player: %s (ID: %d)", player.Name, existingID)
    }
    return nil
}
