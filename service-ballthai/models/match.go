package models

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
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

// MatchUpdateRequest represents the structure for updating a match
type MatchUpdateRequest struct {
	ID            int    `json:"id"`
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

// UpdateMatch updates a match in the database
func UpdateMatch(db *sql.DB, req MatchUpdateRequest) error {
	sqlStr := `
		UPDATE matches SET
			league_id = ?, stage_id = ?, start_date = ?, start_time = ?,
			home_team_id = ?, away_team_id = ?, home_score = ?, away_score = ?,
			match_status = ?, channel_id = ?, live_channel_id = ?
		WHERE id = ?
	`
	args := []interface{}{
		req.LeagueID, req.StageID, req.StartDate, req.StartTime,
		req.HomeTeamID, req.AwayTeamID, req.HomeScore, req.AwayScore,
		req.MatchStatus, req.ChannelID, req.LiveChannelID, req.ID,
	}
	_, err := db.Exec(sqlStr, args...)
	return err
}

// GetMatchByID fetches a match by id
func GetMatchByID(db *sql.DB, id int) (*MatchDB, error) {
	sqlStr := `
		SELECT id, match_ref_id, start_date, start_time, league_id, stage_id,
			home_team_id, away_team_id, channel_id, live_channel_id,
			home_score, away_score, match_status
		FROM matches WHERE id = ?
	`
	var match MatchDB
	err := db.QueryRow(sqlStr, id).Scan(
		&match.ID, &match.MatchRefID, &match.StartDate, &match.StartTime,
		&match.LeagueID, &match.StageID, &match.HomeTeamID, &match.AwayTeamID,
		&match.ChannelID, &match.LiveChannelID, &match.HomeScore, &match.AwayScore, &match.MatchStatus,
	)
	if err != nil {
		return nil, err
	}
	return &match, nil
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

// Handler สำหรับ GET /api/matches/{id}
func MatchGetByIDHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr := vars["id"]
		if idStr == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Missing match id",
			})
			return
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Invalid match id",
			})
			return
		}
		match, err := GetMatchByID(db, id)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Match not found",
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    match,
		})
	}
}

// Handler สำหรับ PUT /api/matches/{id}
func MatchUpdateHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req MatchUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		// ดึง id จาก path parameter ถ้าไม่ได้ส่งมาใน body
		if req.ID == 0 {
			vars := mux.Vars(r)
			idStr := vars["id"]
			if idStr != "" {
				id, err := strconv.Atoi(idStr)
				if err == nil {
					req.ID = id
				}
			}
		}
		if req.ID == 0 {
			http.Error(w, "Missing match id", http.StatusBadRequest)
			return
		}
		if err := UpdateMatch(db, req); err != nil {
			http.Error(w, "Update failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"success":true}`))
	}
}

// Handler สำหรับ GET /api/matches (list all matches)
func MatchListHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sqlStr := `
			SELECT id, match_ref_id, start_date, start_time, league_id, stage_id,
				home_team_id, away_team_id, channel_id, live_channel_id,
				home_score, away_score, match_status
			FROM matches
			ORDER BY start_date DESC, start_time DESC
		`
		rows, err := db.Query(sqlStr)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Database error",
			})
			return
		}
		defer rows.Close()
		var matches []MatchDB
		for rows.Next() {
			var match MatchDB
			err := rows.Scan(
				&match.ID, &match.MatchRefID, &match.StartDate, &match.StartTime,
				&match.LeagueID, &match.StageID, &match.HomeTeamID, &match.AwayTeamID,
				&match.ChannelID, &match.LiveChannelID, &match.HomeScore, &match.AwayScore, &match.MatchStatus,
			)
			if err != nil {
				continue
			}
			matches = append(matches, match)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    matches,
		})
	}
}

// Handler สำหรับ DELETE /api/matches/{id}
func MatchDeleteHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr := vars["id"]
		if idStr == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Missing match id",
			})
			return
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Invalid match id",
			})
			return
		}
		sqlStr := `DELETE FROM matches WHERE id = ?`
		_, err = db.Exec(sqlStr, id)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Delete failed",
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	}
}
