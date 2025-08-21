package database

import (
	"database/sql"
)

type ChannelInfo struct {
	ID      int
	Name    string
	LogoURL string
}

// GetChannelInfoByID returns channel name and logo_url by id
func GetChannelInfoByID(db *sql.DB, id int) (*ChannelInfo, error) {
	var c ChannelInfo
	query := "SELECT id, name, logo_url FROM channels WHERE id = ?"
	err := db.QueryRow(query, id).Scan(&c.ID, &c.Name, &c.LogoURL)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
