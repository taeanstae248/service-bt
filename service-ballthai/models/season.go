package models

// SeasonDB represents the structure of the 'seasons' table in the database
// name = ชื่อฤดูกาล, season_start_date, season_end_date, league_id
// (id เป็น auto increment)
type SeasonDB struct {
	ID               int
	LeagueID         int
	Name             string
	SeasonStartDate  string // YYYY-MM-DD
	SeasonEndDate    string // YYYY-MM-DD
}
