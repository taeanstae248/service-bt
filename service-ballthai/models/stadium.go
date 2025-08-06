package models

import (
	"database/sql" // Added for sql.NullString, sql.NullInt64, sql.NullFloat64
	"encoding/json" // Added for json.RawMessage
)

// StadiumAPI represents the structure of stadium data from the API
type StadiumAPI struct {
	ID             int           `json:"id"`
	Photo          string        `json:"photo"`
	Country        json.RawMessage `json:"country"` // Changed to json.RawMessage to handle inconsistent types (string or object)
	ClubNames      []ClubNameAPI `json:"club_names"` // Array of club names, usually only one relevant
	Name           string        `json:"name"`
	ShortName      string        `json:"short_name"`
	NameEN         string        `json:"name_en"`
	ShortNameEN    string        `json:"short_en"` // Corrected JSON tag based on typical API responses
	CreatedYear    int           `json:"created_year"`
	Capacity       int           `json:"capacity"`
	Latitude       float64       `json:"latitude"`
	Longitude      float64       `json:"longitude"`
}

// StadiumDB represents the structure of the 'stadiums' table in the database
type StadiumDB struct {
	ID                 int
	StadiumRefID       int
	TeamID             sql.NullInt64 // Use sql.NullInt64 for nullable FK
	Name               string
	ShortName          sql.NullString
	NameEN             sql.NullString
	ShortNameEN        sql.NullString
	YearEstablished    sql.NullInt64
	CountryName        sql.NullString
	CountryCode        sql.NullString
	Capacity           sql.NullInt64
	Latitude           sql.NullFloat64
	Longitude          sql.NullFloat64
	PhotoURL           sql.NullString
}
