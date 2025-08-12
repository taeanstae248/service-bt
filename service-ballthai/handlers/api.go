package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Data structures
type League struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Team struct {
	ID              int     `json:"id"`
	Name            string  `json:"name"`
	TeamPostID      *int    `json:"team_post_id,omitempty"`
	StadiumID       *int    `json:"stadium_id,omitempty"`
	StadiumName     *string `json:"stadium_name,omitempty"`
	Logo            *string `json:"logo,omitempty"`
	EstablishedYear *int    `json:"established_year,omitempty"`
}

type Stadium struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Capacity *int    `json:"capacity,omitempty"`
	Location *string `json:"location,omitempty"`
}

type Match struct {
	ID         int     `json:"id"`
	HomeTeam   string  `json:"home_team"`
	AwayTeam   string  `json:"away_team"`
	HomeScore  *int    `json:"home_score"`
	AwayScore  *int    `json:"away_score"`
	StartDate  string  `json:"start_date"`
	Stadium    *string `json:"stadium,omitempty"`
	Status     string  `json:"status"`
	LeagueID   *int    `json:"league_id,omitempty"`
	LeagueName *string `json:"league_name,omitempty"`
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

var DB *sql.DB

// SetDB sets the database connection
func SetDB(database *sql.DB) {
	DB = database
}

// GetLeagues returns all leagues
func GetLeagues(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := "SELECT id, name FROM leagues ORDER BY name"
	rows, err := DB.Query(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var leagues []League
	for rows.Next() {
		var league League
		if err := rows.Scan(&league.ID, &league.Name); err != nil {
			http.Error(w, fmt.Sprintf("Scan error: %v", err), http.StatusInternalServerError)
			return
		}
		leagues = append(leagues, league)
	}

	response := APIResponse{
		Success: true,
		Data:    leagues,
	}

	json.NewEncoder(w).Encode(response)
}

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
		if err := rows.Scan(&team.ID, &team.Name, &team.TeamPostID, &team.StadiumID,
			&team.StadiumName, &team.Logo, &team.EstablishedYear); err != nil {
			http.Error(w, fmt.Sprintf("Scan error: %v", err), http.StatusInternalServerError)
			return
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
	err = DB.QueryRow(query, teamID).Scan(&team.ID, &team.Name, &team.TeamPostID,
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

// GetMatches returns matches with optional pagination and league filter
func GetMatches(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	leagueIDStr := r.URL.Query().Get("league_id")
	leagueName := r.URL.Query().Get("league")
	scoreOnly := r.URL.Query().Get("score")

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
			 m.home_score, m.away_score, m.start_date, s.name as stadium,
			 m.match_status, m.league_id, l.name as league_name
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

	// Filter by score (matches in the past)
	if scoreOnly == "true" {
		query += " AND m.start_date <= CURDATE()"
	} else {
		query += " AND m.start_date > CURDATE()"
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
			&match.HomeScore, &match.AwayScore, &match.StartDate,
			&match.Stadium, &match.Status, &match.LeagueID, &match.LeagueName); err != nil {
			http.Error(w, fmt.Sprintf("Scan error: %v", err), http.StatusInternalServerError)
			return
		}
		matches = append(matches, match)
	}

	response := APIResponse{
		Success: true,
		Data:    matches,
	}

	json.NewEncoder(w).Encode(response)
}
