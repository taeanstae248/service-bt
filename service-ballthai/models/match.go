package models

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

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

// MatchInsertRequest represents the structure for inserting a new match
type MatchInsertRequest struct {
	LeagueID      int    `json:"league_id"`
	StageID       int    `json:"stage_id"`
	StartDate     string `json:"start_date"`
	StartTime     string `json:"start_time"`
	HomeTeamID    int    `json:"home_team_id"`
	AwayTeamID    int    `json:"away_team_id"`
	HomeScore     *int   `json:"home_score"`
	AwayScore     *int   `json:"away_score"`
	MatchStatus   string `json:"match_status"`
	ChannelID     *int   `json:"channel_id"`
	LiveChannelID *int   `json:"live_channel_id"`
}

// InsertMatch inserts a new match into the database
func InsertMatch(db *sql.DB, req MatchInsertRequest) error {
	log.Printf("InsertMatch payload: %+v\n", req)
	sqlStr := `
		INSERT INTO matches (
			league_id, stage_id, start_date, start_time,
			home_team_id, away_team_id, home_score, away_score,
			match_status, channel_id, live_channel_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	args := []interface{}{
		req.LeagueID, req.StageID, req.StartDate, req.StartTime,
		req.HomeTeamID, req.AwayTeamID, req.HomeScore, req.AwayScore,
		req.MatchStatus, req.ChannelID, req.LiveChannelID,
	}
	log.Printf("InsertMatch args: %+v\n", args)
	res, err := db.Exec(sqlStr, args...)
	if err != nil {
		log.Printf("InsertMatch error: %v\n", err)
	} else {
		id, _ := res.LastInsertId()
		log.Printf("InsertMatch success, inserted id: %d\n", id)
	}
	return err
}

// Handler ควรรับ db *sql.DB เป็น argument
func MatchCreateHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req MatchInsertRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		log.Printf("Handler decoded req: %+v\n", req)
		if err := InsertMatch(db, req); err != nil {
			log.Printf("Handler InsertMatch error: %v\n", err)
			http.Error(w, "Insert failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(`{"success":true}`))
	}
}
