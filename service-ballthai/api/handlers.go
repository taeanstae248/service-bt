package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// Handler holds the database connection
type Handler struct {
	DB *sql.DB
}

// NewHandler creates a new handler with database connection
func NewHandler(db *sql.DB) *Handler {
	return &Handler{DB: db}
}

// SetupRoutes sets up all API routes
func (h *Handler) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/leagues", h.GetLeagues)
	mux.HandleFunc("/api/teams", h.GetTeams)
	mux.HandleFunc("/api/stadiums", h.GetStadiums)
	mux.HandleFunc("/api/matches", h.GetMatches)
	mux.HandleFunc("/api/teams/", h.GetTeamByID) // Handle /api/teams/{id}

	// Static file serving for images
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./img/"))))

	return mux
}

// GetLeagues returns all leagues
func (h *Handler) GetLeagues(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := `SELECT id, name FROM leagues ORDER BY name`
	rows, err := h.DB.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type League struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	var leagues []League
	for rows.Next() {
		var league League
		err := rows.Scan(&league.ID, &league.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		leagues = append(leagues, league)
	}

	response := Response{
		Success: true,
		Data:    leagues,
	}

	json.NewEncoder(w).Encode(response)
}

// GetTeams returns all teams with optional filtering
func (h *Handler) GetTeams(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := `SELECT id, name_th, name_en, logo_url, team_post_ballthai, website, shop FROM teams ORDER BY name_th`
	rows, err := h.DB.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Team struct {
		ID               int     `json:"id"`
		NameTH           string  `json:"name_th"`
		NameEN           *string `json:"name_en"`
		LogoURL          *string `json:"logo_url"`
		TeamPostBallthai *string `json:"team_post_ballthai"`
		Website          *string `json:"website"`
		Shop             *string `json:"shop"`
	}

	var teams []Team
	for rows.Next() {
		var team Team
		err := rows.Scan(
			&team.ID,
			&team.NameTH,
			&team.NameEN,
			&team.LogoURL,
			&team.TeamPostBallthai,
			&team.Website,
			&team.Shop,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		teams = append(teams, team)
	}

	response := Response{
		Success: true,
		Data:    teams,
	}

	json.NewEncoder(w).Encode(response)
}

// GetStadiums returns all stadiums
func (h *Handler) GetStadiums(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := `
		SELECT s.id, s.stadium_ref_id, s.team_id, s.name, s.short_name, s.name_en, 
		       s.short_name_en, s.year_established, s.country_name, s.country_code,
		       s.capacity, s.latitude, s.longitude, s.photo_url, t.name_th as team_name
		FROM stadiums s
		LEFT JOIN teams t ON s.team_id = t.id
		ORDER BY s.name
	`
	rows, err := h.DB.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Stadium struct {
		ID              int      `json:"id"`
		StadiumRefID    int      `json:"stadium_ref_id"`
		TeamID          *int     `json:"team_id"`
		Name            string   `json:"name"`
		ShortName       *string  `json:"short_name"`
		NameEN          *string  `json:"name_en"`
		ShortNameEN     *string  `json:"short_name_en"`
		YearEstablished *int     `json:"year_established"`
		CountryName     *string  `json:"country_name"`
		CountryCode     *string  `json:"country_code"`
		Capacity        *int     `json:"capacity"`
		Latitude        *float64 `json:"latitude"`
		Longitude       *float64 `json:"longitude"`
		PhotoURL        *string  `json:"photo_url"`
		TeamName        *string  `json:"team_name"`
	}

	var stadiums []Stadium
	for rows.Next() {
		var stadium Stadium
		err := rows.Scan(
			&stadium.ID,
			&stadium.StadiumRefID,
			&stadium.TeamID,
			&stadium.Name,
			&stadium.ShortName,
			&stadium.NameEN,
			&stadium.ShortNameEN,
			&stadium.YearEstablished,
			&stadium.CountryName,
			&stadium.CountryCode,
			&stadium.Capacity,
			&stadium.Latitude,
			&stadium.Longitude,
			&stadium.PhotoURL,
			&stadium.TeamName,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		stadiums = append(stadiums, stadium)
	}

	response := Response{
		Success: true,
		Data:    stadiums,
	}

	json.NewEncoder(w).Encode(response)
}

// GetMatches returns matches with optional filtering, separated into upcoming and past matches
func (h *Handler) GetMatches(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get query parameters for filtering
	leagueID := r.URL.Query().Get("league_id")
	league := r.URL.Query().Get("league")
	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "25" // Default limit per section
	}

	// League short name mapping
	leagueMap := map[string]string{
		"t1":    "1", // Thai League 1
		"t2":    "2", // Thai League 2
		"t3":    "3", // Thai League 3
		"fa":    "4", // FA Cup
		"lc":    "5", // League Cup
		"youth": "6", // Thai Youth League
		"cl":    "7", // Champions League
		"afc":   "8", // AFC Cup
	}

	// Check if league short name is provided
	if league != "" && leagueMap[league] != "" {
		leagueID = leagueMap[league]
	}

	type Match struct {
		ID           int     `json:"id"`
		MatchRefID   int     `json:"match_ref_id"`
		HomeTeamID   int     `json:"home_team_id"`
		AwayTeamID   int     `json:"away_team_id"`
		LeagueID     int     `json:"league_id"`
		StartDate    *string `json:"start_date"`
		StartTime    *string `json:"start_time"`
		HomeScore    *int    `json:"home_score"`
		AwayScore    *int    `json:"away_score"`
		MatchStatus  *string `json:"match_status"`
		HomeTeamName *string `json:"home_team_name"`
		AwayTeamName *string `json:"away_team_name"`
		LeagueName   *string `json:"league_name"`
	}

	// Query for upcoming matches (>= today)
	upcomingQuery := `
		SELECT m.id, m.match_ref_id, m.home_team_id, m.away_team_id, m.league_id,
		       m.start_date, m.start_time, m.home_score, m.away_score, m.match_status,
		       ht.name_th as home_team_name, at.name_th as away_team_name, l.name as league_name
		FROM matches m
		LEFT JOIN teams ht ON m.home_team_id = ht.id
		LEFT JOIN teams at ON m.away_team_id = at.id
		LEFT JOIN leagues l ON m.league_id = l.id
		WHERE m.start_date >= CURDATE()
	`

	// Query for past matches (< today)
	pastQuery := `
		SELECT m.id, m.match_ref_id, m.home_team_id, m.away_team_id, m.league_id,
		       m.start_date, m.start_time, m.home_score, m.away_score, m.match_status,
		       ht.name_th as home_team_name, at.name_th as away_team_name, l.name as league_name
		FROM matches m
		LEFT JOIN teams ht ON m.home_team_id = ht.id
		LEFT JOIN teams at ON m.away_team_id = at.id
		LEFT JOIN leagues l ON m.league_id = l.id
		WHERE m.start_date < CURDATE()
	`

	upcomingArgs := []interface{}{}
	pastArgs := []interface{}{}

	// Add league filter if specified
	if leagueID != "" {
		upcomingQuery += " AND m.league_id = ?"
		pastQuery += " AND m.league_id = ?"
		upcomingArgs = append(upcomingArgs, leagueID)
		pastArgs = append(pastArgs, leagueID)
	}

	// Add ordering and limit
	upcomingQuery += " ORDER BY m.start_date ASC, m.start_time ASC LIMIT ?"
	pastQuery += " ORDER BY m.start_date DESC, m.start_time DESC LIMIT ?"
	upcomingArgs = append(upcomingArgs, limit)
	pastArgs = append(pastArgs, limit)

	// Execute upcoming matches query
	upcomingRows, err := h.DB.Query(upcomingQuery, upcomingArgs...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer upcomingRows.Close()

	var upcomingMatches []Match
	for upcomingRows.Next() {
		var match Match
		err := upcomingRows.Scan(
			&match.ID,
			&match.MatchRefID,
			&match.HomeTeamID,
			&match.AwayTeamID,
			&match.LeagueID,
			&match.StartDate,
			&match.StartTime,
			&match.HomeScore,
			&match.AwayScore,
			&match.MatchStatus,
			&match.HomeTeamName,
			&match.AwayTeamName,
			&match.LeagueName,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		upcomingMatches = append(upcomingMatches, match)
	}

	// Execute past matches query
	pastRows, err := h.DB.Query(pastQuery, pastArgs...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer pastRows.Close()

	var pastMatches []Match
	for pastRows.Next() {
		var match Match
		err := pastRows.Scan(
			&match.ID,
			&match.MatchRefID,
			&match.HomeTeamID,
			&match.AwayTeamID,
			&match.LeagueID,
			&match.StartDate,
			&match.StartTime,
			&match.HomeScore,
			&match.AwayScore,
			&match.MatchStatus,
			&match.HomeTeamName,
			&match.AwayTeamName,
			&match.LeagueName,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		pastMatches = append(pastMatches, match)
	}

	// Create response with both upcoming and past matches
	type MatchesResponse struct {
		Upcoming []Match `json:"upcoming"`
		Past     []Match `json:"past"`
	}

	response := Response{
		Success: true,
		Data: MatchesResponse{
			Upcoming: upcomingMatches,
			Past:     pastMatches,
		},
	}

	json.NewEncoder(w).Encode(response)
} // GetTeamByID returns a specific team by ID
func (h *Handler) GetTeamByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract ID from URL path (e.g., /api/teams/123)
	path := strings.TrimPrefix(r.URL.Path, "/api/teams/")
	teamID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	type Team struct {
		ID               int     `json:"id"`
		NameTH           string  `json:"name_th"`
		NameEN           *string `json:"name_en"`
		LogoURL          *string `json:"logo_url"`
		TeamPostBallthai *string `json:"team_post_ballthai"`
		Website          *string `json:"website"`
		Shop             *string `json:"shop"`
	}

	var team Team
	query := `SELECT id, name_th, name_en, logo_url, team_post_ballthai, website, shop FROM teams WHERE id = ?`
	err = h.DB.QueryRow(query, teamID).Scan(
		&team.ID,
		&team.NameTH,
		&team.NameEN,
		&team.LogoURL,
		&team.TeamPostBallthai,
		&team.Website,
		&team.Shop,
	)

	if err == sql.ErrNoRows {
		response := Response{
			Success: false,
			Message: "Team not found",
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := Response{
		Success: true,
		Data:    team,
	}

	json.NewEncoder(w).Encode(response)
}
