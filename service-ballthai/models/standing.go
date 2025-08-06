package models

import "database/sql" // Added for sql.NullInt64

// StandingAPI represents the structure of standing data from the API
type StandingAPI struct {
	ID                 int    `json:"id"` // This ID might be for the standing entry itself, not the team
	TournamentTeamName string `json:"tournament_team_name"`
	TournamentTeamNameEN string `json:"tournament_team_name_en"`
	TournamentTeamLogo string `json:"tournament_team_logo"`
	StageName          string `json:"stage_name"`
	StageNameEN        string `json:"stage_en"` // Corrected JSON tag based on typical API responses
	MatchPlay          int    `json:"match_play"`
	Win                int    `json:"win"`
	Draw               int    `json:"draw"`
	Lose               int    `json:"lose"`
	GoalFor            int    `json:"goal_for"`
	GoalAgainst        int    `json:"goal_against"`
	GoalDifference     int    `json:"goal_difference"`
	Point              int    `json:"point"`
	CurrentRank        int    `json:"current_rank"`
}

// StandingDB represents the structure of the 'league_standings' table in the database
type StandingDB struct {
	ID              int
	LeagueID        int
	TeamID          int
	Round           sql.NullInt64
	MatchesPlayed   int
	Wins            int
	Draws           int
	Losses          int
	GoalsFor        int
	GoalsAgainst    int
	GoalDifference  int
	Points          int
	CurrentRank     sql.NullInt64
}
