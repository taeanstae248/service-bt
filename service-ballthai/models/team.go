package models

import "database/sql" // Added for sql.NullString, sql.NullInt64

// TeamAPI represents the structure of team data from the API (e.g., from tournament-team-public)
type TeamAPI struct {
	ID         int    `json:"id"`
	Website    string `json:"website"`
	Shop       string `json:"shop"`
	Stadium    string `json:"stadium"`     // Name of stadium
	StadiumEN  string `json:"stadium_en"`  // English name of stadium
	StadiumPhoto string `json:"stadium_photo"` // Photo URL of stadium (might be in team API)
	Name       string `json:"name"`        // Team name Thai
	NameEN     string `json:"name_en"`     // Team name English
	Logo       string `json:"logo"`        // Team logo URL
	TournamentTeamID int `json:"tournament_team_id"` // Specific ID for team in a tournament
}

// TeamDB represents the structure of the 'teams' table in the database
type TeamDB struct {
	ID               int
	TeamRefID        sql.NullInt64 // ID from API if exists
	NameTH           string
	NameEN           sql.NullString
	LogoURL          sql.NullString
	TeamPostBallthai sql.NullString
	Website          sql.NullString
	Shop             sql.NullString
	StadiumID        sql.NullInt64 // Foreign Key to stadiums table
	LeagueID         sql.NullInt64 // Foreign Key to leagues table (if team is tied to a specific league)
}
