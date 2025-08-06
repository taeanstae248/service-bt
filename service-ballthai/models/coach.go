package models

import (
	"database/sql"
	"time" // For time.Time
)

// CoachAPI represents the structure of coach data from the API
type CoachAPI struct {
	ID          int            `json:"id"`
	FullName    string         `json:"full_name"`
	ClubName    string         `json:"club_name"`
	BirthDate   string         `json:"birth_date"` // API provides as string
	Photo       string         `json:"photo"`
	Nationality NationalityAPI `json:"nationality"`
}

// CoachDB represents the structure of the 'coaches' table in the database
type CoachDB struct {
	ID            int
	CoachRefID    sql.NullInt64
	Name          string
	Birthday      sql.NullTime // Use sql.NullTime for nullable DATE
	TeamID        sql.NullInt64
	NationalityID sql.NullInt64
	PhotoURL      sql.NullString
}
