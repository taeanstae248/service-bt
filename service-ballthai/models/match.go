package models

import "database/sql"

// MatchAPI represents the structure of match data from the API
type MatchAPI struct {
	ID               int         `json:"id"`
	StartDate        string      `json:"start_date"` // API provides as string
	StartTime        string      `json:"start_time"` // API provides as string
	HomeTeamName     string      `json:"home_team_name"`
	HomeTeamNameEN   string      `json:"home_team_name_en"`
	HomeTeamLogo     string      `json:"home_team_logo"`
	AwayTeamName     string      `json:"away_team_name"`
	AwayTeamNameEN   string      `json:"away_team_name_en"`
	AwayTeamLogo     string      `json:"away_team_logo"`
	StadiumName      string      `json:"stadium_name"` // Name of stadium
	TournamentName   string      `json:"tournament_name"`
	TournamentNameEN string      `json:"tournament_name_en"`
	ChannelInfo      ChannelAPI  `json:"channel_info"` // Main TV channel
	LiveInfo         ChannelAPI  `json:"live_info"`    // Live streaming channel
	HomeGoalCount    int         `json:"home_goal_count"`
	AwayGoalCount    int         `json:"away_goal_count"`
	MatchStatus      interface{} `json:"match_status"` // Changed to interface{} to handle inconsistent types (number or string)
	StageName        string      `json:"stage_name"`
	StageNameEN      string      `json:"stage_en"` // Corrected JSON tag based on typical API responses
	StageID          int         `json:"stage_id"` // เพิ่มฟิลด์สำหรับ stage_id จาก JSON
}

// MatchDB represents the structure of the 'matches' table in the database
type MatchDB struct {
	ID            int
	MatchRefID    int
	StartDate     string // Use string for DATE/TIME if not parsing to time.Time
	StartTime     string
	LeagueID      sql.NullInt64
	StageID       sql.NullInt64 // เพิ่มฟิลด์สำหรับ stage_id
	HomeTeamID    sql.NullInt64
	AwayTeamID    sql.NullInt64
	ChannelID     sql.NullInt64
	LiveChannelID sql.NullInt64
	HomeScore     sql.NullInt64
	AwayScore     sql.NullInt64
	MatchStatus   sql.NullString
}
