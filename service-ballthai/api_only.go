package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"

	"go-ballthai-scraper/database"
)

var db *sql.DB

// API Response structures
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

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
	Location *string `json:"location,omitempty"`
	Capacity *int    `json:"capacity,omitempty"`
}

type Match struct {
	ID          int     `json:"id"`
	HomeTeam    string  `json:"home_team"`
	AwayTeam    string  `json:"away_team"`
	StartDate   *string `json:"start_date"`
	StartTime   *string `json:"start_time"`
	HomeScore   *int    `json:"home_score"`
	AwayScore   *int    `json:"away_score"`
	MatchStatus *string `json:"match_status"`
	LeagueID    int     `json:"league_id"`
	LeagueName  string  `json:"league_name"`
	StadiumName *string `json:"stadium_name"`
}

func main() {
	// Load environment variables
	log.Printf("Attempting to load .env file...")
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	} else {
		log.Printf(".env file loaded successfully.")
	}

	// Get database configuration from environment variables
	dbUser := os.Getenv("DB_USERNAME")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	log.Printf("DEBUG: DB_USERNAME: '%s'", dbUser)
	log.Printf("DEBUG: DB_PASSWORD: '%s' (length: %d)", dbPass, len(dbPass))
	log.Printf("DEBUG: DB_HOST: '%s'", dbHost)
	log.Printf("DEBUG: DB_PORT: '%s'", dbPort)
	log.Printf("DEBUG: DB_NAME: '%s'", dbName)

	if dbUser == "" || dbHost == "" || dbPort == "" || dbName == "" {
		log.Fatalf("Missing one or more essential database environment variables")
	}

	connStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)
	log.Printf("DEBUG: Connection String: %s", connStr)

	// Initialize database connection
	err = database.InitDB(connStr)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	db = database.DB
	defer db.Close()

	log.Println("Database connection successful!")
	startAPIServer()
}

func startAPIServer() {
	router := mux.NewRouter()

	// API routes
	router.HandleFunc("/api/leagues", getLeagues).Methods("GET")
	router.HandleFunc("/api/teams", getTeams).Methods("GET")
	router.HandleFunc("/api/teams/{id}", getTeamByID).Methods("GET")
	router.HandleFunc("/api/stadiums", getStadiums).Methods("GET")
	router.HandleFunc("/api/matches", getMatches).Methods("GET")

	// Static file serving for images
	router.PathPrefix("/images/").Handler(http.StripPrefix("/images/", http.FileServer(http.Dir("./img/"))))

	// Enable CORS
	c := cors.AllowAll()
	handler := c.Handler(router)

	log.Println("Starting API server on :8080")
	log.Println("API endpoints available:")
	log.Println("  GET /api/leagues - Get all leagues")
	log.Println("  GET /api/teams - Get all teams")
	log.Println("  GET /api/teams/{id} - Get team by ID")
	log.Println("  GET /api/stadiums - Get all stadiums")
	log.Println("  GET /api/matches?league_id={id}&limit={limit} - Get matches")
	log.Println("  GET /images/{path} - Serve static images")

	log.Fatal(http.ListenAndServe(":8080", handler))
}

func getLeagues(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := "SELECT id, name FROM leagues ORDER BY name"
	rows, err := db.Query(query)
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

func getTeams(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := `
		SELECT t.id, t.name, t.team_post_ballthai, t.stadium_id, s.name as stadium_name, 
		       t.logo, t.established_year
		FROM teams t 
		LEFT JOIN stadiums s ON t.stadium_id = s.id
		ORDER BY t.name
	`

	rows, err := db.Query(query)
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

func getTeamByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	teamID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	query := `
		SELECT t.id, t.name, t.team_post_ballthai, t.stadium_id, s.name as stadium_name, 
		       t.logo, t.established_year
		FROM teams t 
		LEFT JOIN stadiums s ON t.stadium_id = s.id
		WHERE t.id = ?
	`

	var team Team
	err = db.QueryRow(query, teamID).Scan(&team.ID, &team.Name, &team.TeamPostID,
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

func getStadiums(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := "SELECT id, name, location, capacity FROM stadiums ORDER BY name"
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var stadiums []Stadium
	for rows.Next() {
		var stadium Stadium
		if err := rows.Scan(&stadium.ID, &stadium.Name, &stadium.Location, &stadium.Capacity); err != nil {
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

func getMatches(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse query parameters
	leagueIDStr := r.URL.Query().Get("league_id")
	limitStr := r.URL.Query().Get("limit")

	var args []interface{}
	query := `
		SELECT m.id, ht.name as home_team, at.name as away_team,
		       m.start_date, m.start_time, m.home_score, m.away_score, m.match_status,
		       m.league_id, l.name as league_name, s.name as stadium_name
		FROM matches m
		JOIN teams ht ON m.home_team_id = ht.id
		JOIN teams at ON m.away_team_id = at.id
		JOIN leagues l ON m.league_id = l.id
		LEFT JOIN stadiums s ON m.stadium_id = s.id
		WHERE 1=1
	`

	if leagueIDStr != "" {
		leagueID, err := strconv.Atoi(leagueIDStr)
		if err != nil {
			http.Error(w, "Invalid league_id parameter", http.StatusBadRequest)
			return
		}
		query += " AND m.league_id = ?"
		args = append(args, leagueID)
	}

	query += " ORDER BY m.start_date DESC, m.start_time DESC"

	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}
		query += " LIMIT ?"
		args = append(args, limit)
	} else {
		query += " LIMIT 50" // default limit
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var matches []Match
	for rows.Next() {
		var match Match
		if err := rows.Scan(&match.ID, &match.HomeTeam, &match.AwayTeam,
			&match.StartDate, &match.StartTime, &match.HomeScore, &match.AwayScore, &match.MatchStatus,
			&match.LeagueID, &match.LeagueName, &match.StadiumName); err != nil {
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
