package database

import (
	"database/sql"
)

// GetTeamLogoByName returns the logo URL for a team by Thai name
func GetTeamLogoByName(db *sql.DB, nameTh string) (string, error) {
	var logo sql.NullString
	err := db.QueryRow("SELECT logo_url FROM teams WHERE name_th = ?", nameTh).Scan(&logo)
	if err != nil {
		return "", err
	}
	if logo.Valid {
		return logo.String, nil
	}
	return "", nil
}
