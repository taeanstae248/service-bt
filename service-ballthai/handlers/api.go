// UpdateMatch handles PUT /api/matches/{id}

// ...existing code...

// UpdateMatch handles PUT /api/matches/{id}

// ...existing code...

// GetMatchByID handles GET /api/matches/{id}
package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// Data structures
type League struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Team struct {
	ID              int     `json:"id"`
	Name            string  `json:"name"` // เพิ่ม field name (map จาก name_th)
	NameTh          string  `json:"name_th"`
	StadiumID       *int    `json:"stadium_id,omitempty"`
	StadiumName     *string `json:"stadium_name,omitempty"`
	Logo            *string `json:"logo"` // ส่ง logo_url เป็น logo
	EstablishedYear *int    `json:"established_year,omitempty"`
	TeamPostID      *int    `json:"team_post_id,omitempty"`
}

type Stadium struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Capacity *int    `json:"capacity,omitempty"`
	Location *string `json:"location,omitempty"`
}

type Match struct {
	ID            int     `json:"id"`
	HomeTeam      string  `json:"home_team"`
	AwayTeam      string  `json:"away_team"`
	HomeScore     *int    `json:"home_score"`
	AwayScore     *int    `json:"away_score"`
	StartDate     string  `json:"start_date"`
	StartTime     *string `json:"start_time,omitempty"`
	Stadium       *string `json:"stadium,omitempty"`
	Status        string  `json:"status"`
	MatchStatus   string  `json:"match_status"`
	LeagueID      *int    `json:"league_id,omitempty"`
	LeagueName    *string `json:"league_name,omitempty"`
	TeamPostHome  *string `json:"team_post_home,omitempty"`
	TeamPostAway  *string `json:"team_post_away,omitempty"`
	ChannelID     *int    `json:"channel_id,omitempty"`      // เพิ่ม field นี้
	LiveChannelID *int    `json:"live_channel_id,omitempty"` // เพิ่ม field นี้
}

type Player struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	Position      *string `json:"position,omitempty"`
	ShirtNumber   *int    `json:"shirt_number,omitempty"`
	TeamID        *int    `json:"team_id,omitempty"`
	TeamName      *string `json:"team_name,omitempty"`
	TeamPostID    *int    `json:"team_post_id,omitempty"`
	Age           *int    `json:"age,omitempty"`
	Height        *string `json:"height,omitempty"`
	Weight        *string `json:"weight,omitempty"`
	Nationality   *string `json:"nationality,omitempty"`
	PlayerPostID  *int    `json:"player_post_id,omitempty"`
	ProfileImage  *string `json:"profile_image,omitempty"`
	DateOfBirth   *string `json:"date_of_birth,omitempty"`
	PlaceOfBirth  *string `json:"place_of_birth,omitempty"`
	CareerStart   *int    `json:"career_start,omitempty"`
	PreferredFoot *string `json:"preferred_foot,omitempty"`
}

// DeleteMatch handles DELETE /api/matches/{id}
func DeleteMatch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"success": false, "error": "Invalid match id"}`, http.StatusBadRequest)
		return
	}
	res, err := DB.Exec("DELETE FROM matches WHERE id = ?", id)
	if err != nil {
		http.Error(w, `{"success": false, "error": "Failed to delete match"}`, http.StatusInternalServerError)
		return
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		http.Error(w, `{"success": false, "error": "Match not found"}`, http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// UpdateMatch handles PUT /api/matches/{id}
func UpdateMatch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"success": false, "error": "Invalid match id"}`, http.StatusBadRequest)
		return
	}
	var req struct {
		LeagueID      int    `json:"league_id"`
		StageID       *int   `json:"stage_id"`
		StartDate     string `json:"start_date"`
		StartTime     string `json:"start_time"`
		HomeTeamID    int    `json:"home_team_id"`
		AwayTeamID    int    `json:"away_team_id"`
		HomeScore     int    `json:"home_score"`
		AwayScore     int    `json:"away_score"`
		MatchStatus   string `json:"match_status"`
		ChannelID     *int   `json:"channel_id"`
		LiveChannelID *int   `json:"live_channel_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"success": false, "error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}
	query := `UPDATE matches SET
		league_id = ?,
		stage_id = ?,
		start_date = ?,
		start_time = ?,
		home_team_id = ?,
		away_team_id = ?,
		home_score = ?,
		away_score = ?,
		match_status = ?,
		channel_id = ?,
		live_channel_id = ?
		WHERE id = ?`
	_, err = DB.Exec(query,
		req.LeagueID, req.StageID, req.StartDate, req.StartTime,
		req.HomeTeamID, req.AwayTeamID, req.HomeScore, req.AwayScore,
		req.MatchStatus, req.ChannelID, req.LiveChannelID, id,
	)
	if err != nil {
		http.Error(w, `{"success": false, "error": "Failed to update match"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// GetMatchByID handles GET /api/matches/{id}
func GetMatchByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"success": false, "error": "Invalid match id"}`, http.StatusBadRequest)
		return
	}
	query := `
		SELECT m.id, m.league_id, m.stage_id, m.start_date, m.start_time,
			   m.home_team_id, m.away_team_id, m.home_score, m.away_score,
			   m.match_status, m.channel_id, m.live_channel_id,
			   ht.name_th as home_team, at.name_th as away_team,
			   s.name as stadium, l.name as league_name,
			   ht.team_post_ballthai as team_post_home, at.team_post_ballthai as team_post_away
		FROM matches m
		LEFT JOIN teams ht ON m.home_team_id = ht.id
		LEFT JOIN teams at ON m.away_team_id = at.id
		LEFT JOIN stadiums s ON ht.stadium_id = s.id
		LEFT JOIN leagues l ON m.league_id = l.id
		WHERE m.id = ?
		LIMIT 1
	`
	var resp struct {
		ID            int     `json:"id"`
		LeagueID      *int    `json:"league_id"`
		StageID       *int    `json:"stage_id"`
		StartDate     string  `json:"start_date"`
		StartTime     *string `json:"start_time"`
		HomeTeamID    *int    `json:"home_team_id"`
		AwayTeamID    *int    `json:"away_team_id"`
		HomeScore     *int    `json:"home_score"`
		AwayScore     *int    `json:"away_score"`
		MatchStatus   string  `json:"match_status"`
		ChannelID     *int    `json:"channel_id"`
		LiveChannelID *int    `json:"live_channel_id"`
		HomeTeam      string  `json:"home_team"`
		AwayTeam      string  `json:"away_team"`
		Stadium       *string `json:"stadium"`
		LeagueName    *string `json:"league_name"`
		TeamPostHome  *string `json:"team_post_home"`
		TeamPostAway  *string `json:"team_post_away"`
	}
	row := DB.QueryRow(query, id)
	err = row.Scan(&resp.ID, &resp.LeagueID, &resp.StageID, &resp.StartDate, &resp.StartTime,
		&resp.HomeTeamID, &resp.AwayTeamID, &resp.HomeScore, &resp.AwayScore,
		&resp.MatchStatus, &resp.ChannelID, &resp.LiveChannelID,
		&resp.HomeTeam, &resp.AwayTeam, &resp.Stadium, &resp.LeagueName,
		&resp.TeamPostHome, &resp.TeamPostAway)
	if err == sql.ErrNoRows {
		http.Error(w, `{"success": false, "error": "Match not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, `{"success": false, "error": "Database error"}`, http.StatusInternalServerError)
		return
	}
	if resp.TeamPostAway == nil || *resp.TeamPostAway == "" {
		zero := "0"
		resp.TeamPostAway = &zero
	}
	if resp.MatchStatus == "" {
		resp.MatchStatus = "ADD"
	}
	response := APIResponse{
		Success: true,
		Data:    resp,
	}
	json.NewEncoder(w).Encode(response)
}

var DB *sql.DB

// SetDB sets the database connection
func SetDB(database *sql.DB) {
	DB = database
}

// GetLeagues returns all leagues

// GetTeams returns all teams
func GetTeams(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := `
		SELECT t.id, t.name_th, t.team_post_ballthai, t.stadium_id, s.name as stadium_name, 
		       t.logo_url, NULL as established_year
		FROM teams t 
		LEFT JOIN stadiums s ON t.stadium_id = s.id
		ORDER BY t.name_th
	`

	rows, err := DB.Query(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	   var teams []Team
	   for rows.Next() {
		   var team Team
		   var logoUrl sql.NullString
		   if err := rows.Scan(&team.ID, &team.NameTh, &team.TeamPostID, &team.StadiumID,
			   &team.StadiumName, &logoUrl, &team.EstablishedYear); err != nil {
			   http.Error(w, fmt.Sprintf("Scan error: %v", err), http.StatusInternalServerError)
			   return
		   }
		   team.Name = team.NameTh // map name_th -> name
		   if logoUrl.Valid {
			   team.Logo = &logoUrl.String
		   } else {
			   team.Logo = nil
		   }
		   teams = append(teams, team)
	   }

	response := APIResponse{
		Success: true,
		Data:    teams,
	}

	json.NewEncoder(w).Encode(response)
}

// GetTeamByID returns a specific team
func GetTeamByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	teamID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	query := `
		SELECT t.id, t.name_th, t.team_post_ballthai, t.stadium_id, s.name as stadium_name, 
		       t.logo_url, NULL as established_year
		FROM teams t 
		LEFT JOIN stadiums s ON t.stadium_id = s.id
		WHERE t.id = ?
	`

	var team Team
	err = DB.QueryRow(query, teamID).Scan(&team.ID, &team.NameTh, &team.TeamPostID,
		&team.StadiumID, &team.StadiumName, &team.Logo, &team.EstablishedYear)

	if err == sql.ErrNoRows {
		http.Error(w, "Team not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}

	response := APIResponse{
		Success: true,
		Data:    team,
	}

	json.NewEncoder(w).Encode(response)
}

// GetStadiums returns all stadiums
func GetStadiums(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := "SELECT id, name, capacity, location FROM stadiums ORDER BY name"
	rows, err := DB.Query(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var stadiums []Stadium
	for rows.Next() {
		var stadium Stadium
		if err := rows.Scan(&stadium.ID, &stadium.Name, &stadium.Capacity, &stadium.Location); err != nil {
			http.Error(w, fmt.Sprintf("Scan error: %v", err), http.StatusInternalServerError)
			return
		}
		stadiums = append(stadiums, stadium)
	}

	response := APIResponse{
		Success: true,
		Data:    stadiums,
	}

	json.NewEncoder(w).Encode(response)
}

// GetChannels returns all channels (for TV and Live)
func GetChannels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	query := `SELECT id, name, type FROM channels ORDER BY name`
	rows, err := DB.Query(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var channels []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	}
	for rows.Next() {
		var c struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
			Type string `json:"type"`
		}
		if err := rows.Scan(&c.ID, &c.Name, &c.Type); err != nil {
			http.Error(w, fmt.Sprintf("Scan error: %v", err), http.StatusInternalServerError)
			return
		}
		channels = append(channels, c)
	}
	response := APIResponse{
		Success: true,
		Data:    channels,
	}
	json.NewEncoder(w).Encode(response)
}

// CreateMatch handles POST /api/matches
func CreateMatch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		LeagueID      int    `json:"league_id"`
		StageID       *int   `json:"stage_id"`
		StartDate     string `json:"start_date"`
		StartTime     string `json:"start_time"`
		HomeTeamID    int    `json:"home_team_id"`
		AwayTeamID    int    `json:"away_team_id"`
		HomeScore     int    `json:"home_score"`
		AwayScore     int    `json:"away_score"`
		MatchRefID    int    `json:"match_ref_id"`
		MatchStatus   string `json:"match_status"`
		ChannelID     *int   `json:"channel_id"`      // เพิ่ม field นี้
		LiveChannelID *int   `json:"live_channel_id"` // เพิ่ม field นี้
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"success": false, "error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}
	query := `INSERT INTO matches (
		match_ref_id, league_id, stage_id, start_date, start_time,
		home_team_id, away_team_id, home_score, away_score, match_status,
		channel_id, live_channel_id
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	var stageID interface{} = nil
	if req.StageID != nil && *req.StageID > 0 {
		stageID = *req.StageID
	}
	matchRefID := req.MatchRefID
	if matchRefID == 0 {
		rand.Seed(time.Now().UnixNano())
		matchRefID = 1000 + rand.Intn(9000)
	}
	_, err := DB.Exec(query,
		matchRefID, req.LeagueID, stageID, req.StartDate, req.StartTime,
		req.HomeTeamID, req.AwayTeamID, req.HomeScore, req.AwayScore, req.MatchStatus,
		req.ChannelID, req.LiveChannelID, // เพิ่มตรงนี้
	)
	if err != nil {
		fmt.Printf("CreateMatch DB error: %v\n", err)
		http.Error(w, fmt.Sprintf(`{"success": false, "error": "Failed to save match: %v"}`, err), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}
func GetMatches(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	leagueIDStr := r.URL.Query().Get("league_id")
	leagueName := r.URL.Query().Get("league")
	// scoreOnly := r.URL.Query().Get("score") // ไม่ได้ใช้งาน
	dateStr := r.URL.Query().Get("date")

	limit := 20 // default
	offset := 0 // default
	var args []interface{}

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	query := `
		 SELECT m.id, ht.name_th as home_team, at.name_th as away_team, 
			 m.home_score, m.away_score, m.start_date, m.start_time, s.name as stadium,
			 m.match_status, m.league_id, l.name as league_name,
			 ht.team_post_ballthai as team_post_home, at.team_post_ballthai as team_post_away
		 FROM matches m
		 LEFT JOIN teams ht ON m.home_team_id = ht.id
		 LEFT JOIN teams at ON m.away_team_id = at.id
		 LEFT JOIN stadiums s ON ht.stadium_id = s.id
		 LEFT JOIN leagues l ON m.league_id = l.id
		 WHERE 1=1
	`

	// Add league_id filter
	if leagueIDStr != "" {
		query += " AND m.league_id = ?"
		if leagueID, err := strconv.Atoi(leagueIDStr); err == nil {
			args = append(args, leagueID)
		} else {
			http.Error(w, "Invalid league_id parameter", http.StatusBadRequest)
			return
		}
	}

	// Add league name filter
	if leagueName != "" {
		query += " AND l.name = ?"
		args = append(args, leagueName)
	}

	// Filter by date (exact match yyyy-MM-dd)
	if dateStr != "" {
		query += " AND DATE(m.start_date) = ?"
		args = append(args, dateStr)
	} else {
		// ถ้าไม่ส่ง date ให้แสดงเฉพาะวันปัจจุบัน
		query += " AND DATE(m.start_date) = CURDATE()"
	}

	query += " ORDER BY m.start_date DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := DB.Query(query, args...)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var matches []Match
	for rows.Next() {
		var match Match
		if err := rows.Scan(&match.ID, &match.HomeTeam, &match.AwayTeam,
			&match.HomeScore, &match.AwayScore, &match.StartDate, &match.StartTime,
			&match.Stadium, &match.Status, &match.LeagueID, &match.LeagueName,
			&match.TeamPostHome, &match.TeamPostAway); err != nil {
			http.Error(w, fmt.Sprintf("Scan error: %v", err), http.StatusInternalServerError)
			return
		}
		// ถ้าไม่มีข้อมูล team_post_away ให้แสดงเป็น "0"
		if match.TeamPostAway == nil || *match.TeamPostAway == "" {
			zero := "0"
			match.TeamPostAway = &zero
		}
		// ถ้า status ว่าง ให้เติม "ADD"
		if match.Status == "" {
			match.Status = "ADD"
		}
		// เพิ่มเติม: ให้ match.MatchStatus = match.Status เพื่อให้ JS ใช้ได้
		match.MatchStatus = match.Status
		matches = append(matches, match)
	}

	response := APIResponse{
		Success: true,
		Data:    matches,
	}

	json.NewEncoder(w).Encode(response)
}

// GetStages returns unique stage_name from stage table
func GetStages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	query := `SELECT id, stage_name FROM stage WHERE stage_name IS NOT NULL AND stage_name != '' ORDER BY stage_name`
	rows, err := DB.Query(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var stages []struct {
		ID        int    `json:"id"`
		StageName string `json:"stage_name"`
	}
	for rows.Next() {
		var s struct {
			ID        int    `json:"id"`
			StageName string `json:"stage_name"`
		}
		if err := rows.Scan(&s.ID, &s.StageName); err != nil {
			http.Error(w, fmt.Sprintf("Scan error: %v", err), http.StatusInternalServerError)
			return
		}
		stages = append(stages, s)
	}
	response := APIResponse{
		Success: true,
		Data:    stages,
	}
	json.NewEncoder(w).Encode(response)
}
