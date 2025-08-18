
package database

import (
    "database/sql"
    "fmt"
    "log"
    "go-ballthai-scraper/models"
)

// GetAllLeagues returns all leagues from the database (id, name, thaileageid)
func GetAllLeagues(db *sql.DB) ([]models.LeagueDB, error) {
    rows, err := db.Query("SELECT id, name, thaileageid FROM leagues ORDER BY thaileageid ASC")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var leagues []models.LeagueDB
    for rows.Next() {
        var l models.LeagueDB
        err := rows.Scan(&l.ID, &l.Name, &l.ThaileageID)
        if err != nil {
            return nil, err
        }
        leagues = append(leagues, l)
    }
    return leagues, nil
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
