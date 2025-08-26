package models

import "database/sql"

// LeagueDB represents the structure of the 'leagues' table in the database
type LeagueDB struct {
	ID          int
	Name        string
	ThaileageID sql.NullInt64
}
