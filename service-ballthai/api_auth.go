package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"golang.org/x/crypto/bcrypt"

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

type Player struct {
	ID            int     `json:"id"`
	PlayerRefID   *int    `json:"player_ref_id,omitempty"`
	LeagueID      *int    `json:"league_id,omitempty"`
	TeamID        *int    `json:"team_id,omitempty"`
	NationalityID *int    `json:"nationality_id,omitempty"`
	Name          string  `json:"name"`
	FullNameEN    *string `json:"full_name_en,omitempty"`
	ShirtNumber   *int    `json:"shirt_number,omitempty"`
	Position      *string `json:"position,omitempty"`
	PhotoURL      *string `json:"photo_url,omitempty"`
	MatchesPlayed int     `json:"matches_played"`
	Goals         int     `json:"goals"`
	YellowCards   int     `json:"yellow_cards"`
	RedCards      int     `json:"red_cards"`
	TeamName      *string `json:"team_name,omitempty"`
	LeagueName    *string `json:"league_name,omitempty"`
	Nationality   *string `json:"nationality,omitempty"`
}

// Authentication structures
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	User      *database.User `json:"user"`
	SessionID string         `json:"session_id"`
}

type AuthResponse struct {
	Authenticated bool           `json:"authenticated"`
	User          *database.User `json:"user,omitempty"`
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

	// Authentication routes
	router.HandleFunc("/api/auth/login", loginHandler).Methods("POST")
	router.HandleFunc("/api/auth/logout", logoutHandler).Methods("POST")
	router.HandleFunc("/api/auth/verify", verifyHandler).Methods("GET")

	// API routes
	router.HandleFunc("/api/leagues", getLeagues).Methods("GET")
	router.HandleFunc("/api/teams", getTeams).Methods("GET")
	router.HandleFunc("/api/teams/{id}", getTeamByID).Methods("GET")
	router.HandleFunc("/api/stadiums", getStadiums).Methods("GET")
	router.HandleFunc("/api/matches", getMatches).Methods("GET")
	router.HandleFunc("/api/players", getPlayers).Methods("GET")
	router.HandleFunc("/api/players/team/{id}", getPlayersByTeamID).Methods("GET")
	router.HandleFunc("/api/players/team-post/{post_id}", getPlayersByTeamPost).Methods("GET")

	// Static file serving
	router.PathPrefix("/images/").Handler(http.StripPrefix("/images/", http.FileServer(http.Dir("./img/"))))
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./")))

	// Enable CORS
	c := cors.AllowAll()
	handler := c.Handler(router)

	log.Println("Starting API server on :8080")
	log.Println("API endpoints available:")
	log.Println("  POST /api/auth/login - User login")
	log.Println("  POST /api/auth/logout - User logout")
	log.Println("  GET  /api/auth/verify - Verify session")
	log.Println("  GET  /api/leagues - Get all leagues")
	log.Println("  GET  /api/teams - Get all teams")
	log.Println("  GET  /api/teams/{id} - Get team by ID")
	log.Println("  GET  /api/stadiums - Get all stadiums")
	log.Println("  GET  /api/matches?league_id={id}&limit={limit} - Get matches")
	log.Println("  GET  /api/players?team_id={id}&league_id={id}&position={pos}&limit={limit} - Get players")
	log.Println("  GET  /api/players/team/{id} - Get players by team ID")
	log.Println("  GET  /api/players/team-post/{post_id} - Get players by team_post_ballthai")
	log.Println("  GET  /images/{path} - Serve static images")
	log.Println("  GET  / - Serve static files (login.html, dashboard.html)")

	log.Fatal(http.ListenAndServe(":8080", handler))
}

// Authentication handlers

func loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var loginReq LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, `{"success": false, "error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate input
	if loginReq.Username == "" || loginReq.Password == "" {
		http.Error(w, `{"success": false, "error": "Username and password are required"}`, http.StatusBadRequest)
		return
	}

	// Get user password hash
	passwordHash, err := database.GetUserPasswordHash(loginReq.Username)
	if err == sql.ErrNoRows {
		log.Printf("User not found: %s", loginReq.Username)
		http.Error(w, `{"success": false, "error": "Invalid username or password"}`, http.StatusUnauthorized)
		return
	} else if err != nil {
		log.Printf("Database error getting password hash: %v", err)
		http.Error(w, `{"success": false, "error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(loginReq.Password)); err != nil {
		log.Printf("Password verification failed for user: %s", loginReq.Username)
		http.Error(w, `{"success": false, "error": "Invalid username or password"}`, http.StatusUnauthorized)
		return
	}

	// Get user details
	user, err := database.GetUserByUsername(loginReq.Username)
	if err != nil {
		log.Printf("Failed to get user details: %v", err)
		http.Error(w, `{"success": false, "error": "Failed to get user details"}`, http.StatusInternalServerError)
		return
	}

	// Generate session ID
	sessionID, err := generateSessionID()
	if err != nil {
		http.Error(w, `{"success": false, "error": "Failed to generate session"}`, http.StatusInternalServerError)
		return
	}

	// Create session (expires in 24 hours)
	expiresAt := time.Now().Add(24 * time.Hour)
	if err := database.CreateSession(sessionID, user.ID, expiresAt); err != nil {
		http.Error(w, `{"success": false, "error": "Failed to create session"}`, http.StatusInternalServerError)
		return
	}

	// Update last login
	database.UpdateLastLogin(user.ID)

	// Return response
	response := APIResponse{
		Success: true,
		Data: LoginResponse{
			User:      user,
			SessionID: sessionID,
		},
	}

	json.NewEncoder(w).Encode(response)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get session ID from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, `{"success": false, "error": "No authorization header"}`, http.StatusUnauthorized)
		return
	}

	// Extract session ID (format: "Bearer <sessionID>")
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, `{"success": false, "error": "Invalid authorization header format"}`, http.StatusUnauthorized)
		return
	}

	sessionID := parts[1]

	// Delete session
	database.DeleteSession(sessionID)

	response := APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Logged out successfully"},
	}

	json.NewEncoder(w).Encode(response)
}

func verifyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get session ID from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Data:    AuthResponse{Authenticated: false},
		})
		return
	}

	// Extract session ID
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Data:    AuthResponse{Authenticated: false},
		})
		return
	}

	sessionID := parts[1]

	// Verify session
	session, err := database.GetSession(sessionID)
	if err != nil {
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Data:    AuthResponse{Authenticated: false},
		})
		return
	}

	// Get user details
	user, err := database.GetUserByID(session.UserID)
	if err != nil {
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Data:    AuthResponse{Authenticated: false},
		})
		return
	}

	response := APIResponse{
		Success: true,
		Data: AuthResponse{
			Authenticated: true,
			User:          user,
		},
	}

	json.NewEncoder(w).Encode(response)
}

func generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
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
		SELECT t.id, t.name_th, t.team_post_ballthai, t.stadium_id, s.name as stadium_name, 
		       t.logo_url, NULL as established_year
		FROM teams t 
		LEFT JOIN stadiums s ON t.stadium_id = s.id
		ORDER BY t.name_th
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
		SELECT t.id, t.name_th, t.team_post_ballthai, t.stadium_id, s.name as stadium_name, 
		       t.logo_url, NULL as established_year
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

	query := "SELECT id, name, country_name, capacity FROM stadiums ORDER BY name"
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
		SELECT m.id, ht.name_th as home_team, at.name_th as away_team,
		       m.start_date, m.start_time, m.home_score, m.away_score, m.match_status,
		       m.league_id, l.name as league_name, s.name as stadium_name
		FROM matches m
		JOIN teams ht ON m.home_team_id = ht.id
		JOIN teams at ON m.away_team_id = at.id
		JOIN leagues l ON m.league_id = l.id
		LEFT JOIN stadiums s ON ht.stadium_id = s.id
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

func getPlayers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse query parameters
	teamIDStr := r.URL.Query().Get("team_id")
	leagueIDStr := r.URL.Query().Get("league_id")
	position := r.URL.Query().Get("position")
	limitStr := r.URL.Query().Get("limit")

	var args []interface{}
	query := `
		SELECT p.id, p.player_ref_id, p.league_id, p.team_id, p.nationality_id,
		       p.name, p.full_name_en, p.shirt_number, p.position, p.photo_url,
		       p.matches_played, p.goals, p.yellow_cards, p.red_cards,
		       t.name_th as team_name, l.name as league_name, n.name as nationality
		FROM players p
		LEFT JOIN teams t ON p.team_id = t.id
		LEFT JOIN leagues l ON p.league_id = l.id
		LEFT JOIN nationalities n ON p.nationality_id = n.id
		WHERE 1=1
	`

	// Add filters
	if teamIDStr != "" {
		teamID, err := strconv.Atoi(teamIDStr)
		if err != nil {
			http.Error(w, "Invalid team_id parameter", http.StatusBadRequest)
			return
		}
		query += " AND p.team_id = ?"
		args = append(args, teamID)
	}

	if leagueIDStr != "" {
		leagueID, err := strconv.Atoi(leagueIDStr)
		if err != nil {
			http.Error(w, "Invalid league_id parameter", http.StatusBadRequest)
			return
		}
		query += " AND p.league_id = ?"
		args = append(args, leagueID)
	}

	if position != "" {
		query += " AND p.position = ?"
		args = append(args, position)
	}

	query += " ORDER BY p.name"

	// Add limit
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

	var players []Player
	for rows.Next() {
		var player Player
		if err := rows.Scan(
			&player.ID, &player.PlayerRefID, &player.LeagueID, &player.TeamID, &player.NationalityID,
			&player.Name, &player.FullNameEN, &player.ShirtNumber, &player.Position, &player.PhotoURL,
			&player.MatchesPlayed, &player.Goals, &player.YellowCards, &player.RedCards,
			&player.TeamName, &player.LeagueName, &player.Nationality,
		); err != nil {
			http.Error(w, fmt.Sprintf("Scan error: %v", err), http.StatusInternalServerError)
			return
		}
		players = append(players, player)
	}

	response := APIResponse{
		Success: true,
		Data:    players,
	}

	json.NewEncoder(w).Encode(response)
}

func getPlayersByTeamID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	teamID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	query := `
		SELECT p.id, p.player_ref_id, p.league_id, p.team_id, p.nationality_id,
		       p.name, p.full_name_en, p.shirt_number, p.position, p.photo_url,
		       p.matches_played, p.goals, p.yellow_cards, p.red_cards,
		       t.name_th as team_name, l.name as league_name, n.name as nationality
		FROM players p
		LEFT JOIN teams t ON p.team_id = t.id
		LEFT JOIN leagues l ON p.league_id = l.id
		LEFT JOIN nationalities n ON p.nationality_id = n.id
		WHERE p.team_id = ?
		ORDER BY p.shirt_number, p.name
	`

	rows, err := db.Query(query, teamID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var players []Player
	for rows.Next() {
		var player Player
		if err := rows.Scan(
			&player.ID, &player.PlayerRefID, &player.LeagueID, &player.TeamID, &player.NationalityID,
			&player.Name, &player.FullNameEN, &player.ShirtNumber, &player.Position, &player.PhotoURL,
			&player.MatchesPlayed, &player.Goals, &player.YellowCards, &player.RedCards,
			&player.TeamName, &player.LeagueName, &player.Nationality,
		); err != nil {
			http.Error(w, fmt.Sprintf("Scan error: %v", err), http.StatusInternalServerError)
			return
		}
		players = append(players, player)
	}

	response := APIResponse{
		Success: true,
		Data:    players,
	}

	json.NewEncoder(w).Encode(response)
}

func getPlayersByTeamPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	teamPost := vars["post_id"]
	if teamPost == "" {
		http.Error(w, "Invalid team post ID", http.StatusBadRequest)
		return
	}

	// First, get team ID from team_post_ballthai
	var teamID int
	teamQuery := `SELECT id FROM teams WHERE team_post_ballthai = ?`
	err := db.QueryRow(teamQuery, teamPost).Scan(&teamID)
	if err == sql.ErrNoRows {
		http.Error(w, "Team not found with the provided team_post_ballthai", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}

	query := `
		SELECT p.id, p.player_ref_id, p.league_id, p.team_id, p.nationality_id,
		       p.name, p.full_name_en, p.shirt_number, p.position, p.photo_url,
		       p.matches_played, p.goals, p.yellow_cards, p.red_cards,
		       t.name_th as team_name, l.name as league_name, n.name as nationality
		FROM players p
		LEFT JOIN teams t ON p.team_id = t.id
		LEFT JOIN leagues l ON p.league_id = l.id
		LEFT JOIN nationalities n ON p.nationality_id = n.id
		WHERE p.team_id = ?
		ORDER BY p.shirt_number, p.name
	`

	rows, err := db.Query(query, teamID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var players []Player
	for rows.Next() {
		var player Player
		if err := rows.Scan(
			&player.ID, &player.PlayerRefID, &player.LeagueID, &player.TeamID, &player.NationalityID,
			&player.Name, &player.FullNameEN, &player.ShirtNumber, &player.Position, &player.PhotoURL,
			&player.MatchesPlayed, &player.Goals, &player.YellowCards, &player.RedCards,
			&player.TeamName, &player.LeagueName, &player.Nationality,
		); err != nil {
			http.Error(w, fmt.Sprintf("Scan error: %v", err), http.StatusInternalServerError)
			return
		}
		players = append(players, player)
	}

	response := APIResponse{
		Success: true,
		Data:    players,
	}

	json.NewEncoder(w).Encode(response)
}
