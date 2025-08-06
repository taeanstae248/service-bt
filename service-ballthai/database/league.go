package database

import (
	"database/sql"
	"fmt"
	"log"

	"go-ballthai-scraper/models" // Ensure this module name matches your go.mod
)

// GetLeagueID retrieves the league ID by its name.
// If the league does not exist, it inserts a new league record and returns its ID.
func GetLeagueID(db *sql.DB, leagueName string) (int, error) {
    var leagueID int
    query := "SELECT id FROM leagues WHERE name = ?"
    err := db.QueryRow(query, leagueName).Scan(&leagueID)

    if err == sql.ErrNoRows {
        // League does not exist, insert a new one
        insertQuery := `INSERT INTO leagues (name) VALUES (?)`
        res, err := db.Exec(insertQuery, leagueName)
        if err != nil {
            return 0, fmt.Errorf("failed to insert new league %s: %w", leagueName, err)
        }
        id, err := res.LastInsertId()
        if err != nil {
            return 0, fmt.Errorf("failed to get last insert ID for league %s: %w", leagueName, err)
        }
        log.Printf("Inserted new league: %s (ID: %d)", leagueName, id)
        return int(id), nil
    } else if err != nil {
        return 0, fmt.Errorf("error checking existing league %s: %w", leagueName, err)
    }
    // League exists, return its ID
    return leagueID, nil
}

// InsertOrUpdateLeague inserts or updates a league record in the database.
// This function can be used if you have a specific LeagueDB struct to insert/update.
func InsertOrUpdateLeague(db *sql.DB, league models.LeagueDB) error {
    var existingID int
    query := "SELECT id FROM leagues WHERE name = ?"
    err := db.QueryRow(query, league.Name).Scan(&existingID)

    if err == sql.ErrNoRows {
        // League does not exist, insert a new one
        insertQuery := `INSERT INTO leagues (name) VALUES (?)`
        _, err = db.Exec(insertQuery, league.Name)
        if err != nil {
            return fmt.Errorf("failed to insert league %s: %w", league.Name, err)
        }
        log.Printf("Inserted new league: %s", league.Name)
    } else if err != nil {
        return fmt.Errorf("error checking existing league: %w", err)
    } else {
        // League exists, update the record (only name can be updated here, if needed)
        updateQuery := `UPDATE leagues SET name = ? WHERE id = ?`
        _, err = db.Exec(updateQuery, league.Name, existingID)
        if err != nil {
            return fmt.Errorf("failed to update league %s (ID: %d): %w", league.Name, existingID, err)
        }
        log.Printf("Updated league: %s (ID: %d)", league.Name, existingID)
    }
    return nil
}
