// ...existing code...
package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go-ballthai-scraper/database"
	"go-ballthai-scraper/scraper"
)

// UpdatePlayer updates a player's info by ID
func UpdatePlayer(w http.ResponseWriter, r *http.Request) {
	db := database.DB
	if db == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid player ID", http.StatusBadRequest)
		return
	}
	type reqBody struct {
		Name        string `json:"name"`
		ShirtNumber int    `json:"shirt_number"`
		Position    string `json:"position"`
		Status      int    `json:"status"`
	}
	var body reqBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	// Update player in DB
	_, err = db.Exec(`UPDATE players SET name=?, shirt_number=?, position=?, status=? WHERE id=?`,
		body.Name, body.ShirtNumber, body.Position, body.Status, id)
	if err != nil {
		log.Println("Update player error:", err)
		http.Error(w, "Failed to update player", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// ScrapePlayersHandler handles scraping players for a given tournament
func ScrapePlayersHandler(w http.ResponseWriter, r *http.Request) {
	db := database.DB
	if db == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}
	tournamentID := 1 // default
	if tid := r.URL.Query().Get("tournament_id"); tid != "" {
		id, err := strconv.Atoi(tid)
		if err == nil {
			tournamentID = id
		}
	}
	err := scraper.ScrapePlayers(db, tournamentID)
	if err != nil {
		log.Println("Scrape players error:", err)
		http.Error(w, "Scrape players error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Scrape players completed successfully"))
}

// GetPlayers returns players with optional pagination and filtering
func GetPlayers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	teamIDStr := r.URL.Query().Get("team_id")
	positionFilter := r.URL.Query().Get("position")
	nationalityFilter := r.URL.Query().Get("nationality")

	limit := 20 // default
	offset := 0 // default

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

	// Build query with optional filters
	baseQuery := `
		SELECT p.id, p.name, p.position, p.shirt_number, p.team_id, t.name_th as team_name,
		       t.team_post_ballthai as team_post_id, p.age, p.height, p.weight,
		       n.name as nationality, p.player_post_ballthai as player_post_id,
		       p.profile_image_url, p.date_of_birth, p.place_of_birth,
		       p.career_start, p.preferred_foot
		FROM players p
		LEFT JOIN teams t ON p.team_id = t.id
		LEFT JOIN nationalities n ON p.nationality_id = n.id
	`

	var whereConditions []string
	var args []interface{}

	if teamIDStr != "" {
		if teamID, err := strconv.Atoi(teamIDStr); err == nil {
			whereConditions = append(whereConditions, "p.team_id = ?")
			args = append(args, teamID)
		}
	}

	if positionFilter != "" {
		whereConditions = append(whereConditions, "p.position LIKE ?")
		args = append(args, "%"+positionFilter+"%")
	}

	if nationalityFilter != "" {
		whereConditions = append(whereConditions, "n.name LIKE ?")
		args = append(args, "%"+nationalityFilter+"%")
	}

	query := baseQuery
	if len(whereConditions) > 0 {
		query += " WHERE "
		for i, condition := range whereConditions {
			if i > 0 {
				query += " AND "
			}
			query += condition
		}
	}

	query += " ORDER BY p.name LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := DB.Query(query, args...)
	if err != nil {
		log.Printf("Database error in GetPlayers: %v", err)
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var players []Player
	for rows.Next() {
		var player Player
		if err := rows.Scan(&player.ID, &player.Name, &player.Position, &player.ShirtNumber,
			&player.TeamID, &player.TeamName, &player.TeamPostID, &player.Age,
			&player.Height, &player.Weight, &player.Nationality, &player.PlayerPostID,
			&player.ProfileImage, &player.DateOfBirth, &player.PlaceOfBirth,
			&player.CareerStart, &player.PreferredFoot); err != nil {
			log.Printf("Scan error in GetPlayers: %v", err)
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

// GetPlayersByTeamID returns players for a specific team ID
func GetPlayersByTeamID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	teamID, err := strconv.Atoi(vars["team_id"])
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	query := `
		SELECT p.id, p.name, p.position, p.shirt_number, p.team_id, t.name_th as team_name,
		       t.team_post_ballthai as team_post_id, p.age, p.height, p.weight,
		       n.name as nationality, p.player_post_ballthai as player_post_id,
		       p.profile_image_url, p.date_of_birth, p.place_of_birth,
		       p.career_start, p.preferred_foot
		FROM players p
		LEFT JOIN teams t ON p.team_id = t.id
		LEFT JOIN nationalities n ON p.nationality_id = n.id
		WHERE p.team_id = ?
		ORDER BY p.shirt_number, p.name
	`

	rows, err := DB.Query(query, teamID)
	if err != nil {
		log.Printf("Database error in GetPlayersByTeamID: %v", err)
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var players []Player
	for rows.Next() {
		var player Player
		if err := rows.Scan(&player.ID, &player.Name, &player.Position, &player.ShirtNumber,
			&player.TeamID, &player.TeamName, &player.TeamPostID, &player.Age,
			&player.Height, &player.Weight, &player.Nationality, &player.PlayerPostID,
			&player.ProfileImage, &player.DateOfBirth, &player.PlaceOfBirth,
			&player.CareerStart, &player.PreferredFoot); err != nil {
			log.Printf("Scan error in GetPlayersByTeamID: %v", err)
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

// GetPlayersByTeamPost returns players for a specific team post ID
func GetPlayersByTeamPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	teamPostID, err := strconv.Atoi(vars["team_post_id"])
	if err != nil {
		http.Error(w, "Invalid team post ID", http.StatusBadRequest)
		return
	}

	query := `
		SELECT p.id, p.name, p.position, p.shirt_number, p.team_id, t.name_th as team_name,
		       t.team_post_ballthai as team_post_id, p.age, p.height, p.weight,
		       n.name as nationality, p.player_post_ballthai as player_post_id,
		       p.profile_image_url, p.date_of_birth, p.place_of_birth,
		       p.career_start, p.preferred_foot
		FROM players p
		LEFT JOIN teams t ON p.team_id = t.id
		LEFT JOIN nationalities n ON p.nationality_id = n.id
		WHERE t.team_post_ballthai = ?
		ORDER BY p.shirt_number, p.name
	`

	rows, err := DB.Query(query, teamPostID)
	if err != nil {
		log.Printf("Database error in GetPlayersByTeamPost: %v", err)
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var players []Player
	for rows.Next() {
		var player Player
		if err := rows.Scan(&player.ID, &player.Name, &player.Position, &player.ShirtNumber,
			&player.TeamID, &player.TeamName, &player.TeamPostID, &player.Age,
			&player.Height, &player.Weight, &player.Nationality, &player.PlayerPostID,
			&player.ProfileImage, &player.DateOfBirth, &player.PlaceOfBirth,
			&player.CareerStart, &player.PreferredFoot); err != nil {
			log.Printf("Scan error in GetPlayersByTeamPost: %v", err)
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
