// ...existing code...
package handlers

import (
	"database/sql"
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
		Name          string `json:"name"`
		ShirtNumber   int    `json:"shirt_number"`
		Position      string `json:"position"`
		MatchesPlayed int    `json:"matches_played"`
		Goals         int    `json:"goals"`
		YellowCards   int    `json:"yellow_cards"`
		RedCards      int    `json:"red_cards"`
		Status        int    `json:"status"`
	}
	var body reqBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	// Update player in DB
	_, err = db.Exec(`UPDATE players SET name=?, shirt_number=?, position=?, matches_played=?, goals=?, yellow_cards=?, red_cards=?, status=? WHERE id=?`,
		body.Name, body.ShirtNumber, body.Position, body.MatchesPlayed, body.Goals, body.YellowCards, body.RedCards, body.Status, id)
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
	// tournamentID ไม่ได้ใช้แล้ว เพราะ ScrapePlayers ไม่รับ argument นี้
	err := scraper.ScrapePlayers(db)
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

// GetTopScorers returns players ordered by goals (supports league_id and limit)
func GetTopScorers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	limitStr := r.URL.Query().Get("limit")
	leagueIDStr := r.URL.Query().Get("league_id")

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	query := `
		SELECT p.id, p.name, p.position, p.shirt_number, p.team_id, t.name_th as team_name,
			   t.team_post_ballthai as team_post_id, n.name as nationality,
			   p.photo_url, p.goals
		FROM players p
		LEFT JOIN teams t ON p.team_id = t.id
		LEFT JOIN nationalities n ON p.nationality_id = n.id
	`

	var args []interface{}
	if leagueIDStr != "" {
		var leagueID int
		if id, err := strconv.Atoi(leagueIDStr); err == nil {
			leagueID = id
		} else {
			// Support alias strings (e.g., "t1" -> 1). Add more mappings as needed.
			aliasMap := map[string]int{
				"t1": 1,
				"t1-jpy": 60,
			}
			if mapped, ok := aliasMap[leagueIDStr]; ok {
				leagueID = mapped
			} else {
				http.Error(w, "Invalid league_id parameter", http.StatusBadRequest)
				return
			}
		}
		query += " WHERE p.league_id = ?"
		args = append(args, leagueID)
	}

	query += " ORDER BY p.goals DESC LIMIT ?"
	args = append(args, limit)

	rows, err := DB.Query(query, args...)
	if err != nil {
		log.Printf("Database error in GetTopScorers: %v", err)
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var players []Player
	for rows.Next() {
	var player Player
	var shirt sql.NullInt64
	var pos sql.NullString
	var teamName sql.NullString
	var teamPost sql.NullString
	var nationality sql.NullString
	var photo sql.NullString
	var goals sql.NullInt64

		if err := rows.Scan(&player.ID, &player.Name, &pos, &shirt,
			&player.TeamID, &teamName, &teamPost, &nationality,
			&photo, &goals); err != nil {
			log.Printf("Scan error in GetTopScorers: %v", err)
			http.Error(w, fmt.Sprintf("Scan error: %v", err), http.StatusInternalServerError)
			return
		}
		if pos.Valid {
			s := pos.String
			player.Position = &s
		}
		if shirt.Valid {
			v := int(shirt.Int64)
			player.ShirtNumber = &v
		}
		if teamName.Valid {
			s := teamName.String
			player.TeamName = &s
		}
		if teamPost.Valid {
			if tp, err := strconv.Atoi(teamPost.String); err == nil {
				player.TeamPostID = &tp
			}
		}
		if nationality.Valid {
			s := nationality.String
			player.Nationality = &s
		}
	// player_post not present in players table in this schema
		if photo.Valid {
			s := photo.String
			player.ProfileImage = &s
		}
		if goals.Valid {
			v := int(goals.Int64)
			player.Goals = &v
		}
		// goals is used for ordering; if you want to include it in response we can extend Player struct
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
