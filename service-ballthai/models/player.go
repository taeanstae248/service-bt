package models

import "database/sql" // Added for sql.NullString, sql.NullInt64

// PlayerAPI represents the structure of player data from the API
type PlayerAPI struct {
	ID                     int            `json:"id"`
	FullName               string         `json:"full_name"`
	MatchCount             int            `json:"match_count"`
	GoalFor                int            `json:"goal_for"`
	YellowCardAcc          int            `json:"yellow_card_acc"`
	RedCardViolentConductAcc int          `json:"red_card_violent_conduct_acc"`
	ShirtNumber            sql.NullInt64  `json:"tnm_shirt_number"` // Can be null in API
	Photo                  string         `json:"photo"`
	ClubName               string         `json:"club_name"`
	Nationality            NationalityAPI `json:"nationality"`
	PositionShortName      string         `json:"position_short_name"`
	FullNameEN             string         `json:"full_name_en"`
}

// PlayerDB represents the structure of the 'players' table in the database
type PlayerDB struct {
	ID            int
	PlayerRefID   sql.NullInt64
	LeagueID      sql.NullInt64
	TeamID        sql.NullInt64
	NationalityID sql.NullInt64
	Name          string
	FullNameEN    sql.NullString
	ShirtNumber   sql.NullInt64
	Position      sql.NullString
	PhotoURL      sql.NullString
	MatchesPlayed int
	Goals         int
	YellowCards   int
	RedCards      int
	Status        int
}
