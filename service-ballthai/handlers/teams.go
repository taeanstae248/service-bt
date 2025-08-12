package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// CreateTeam creates a new team
func CreateTeam(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var team struct {
		Name       string `json:"name"`
		NameTh     string `json:"name_th"`
		StadiumID  *int   `json:"stadium_id"`
		TeamPostID *int   `json:"team_post_id"`
		LogoURL    string `json:"logo_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		http.Error(w, `{"success": false, "error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if team.NameTh == "" {
		http.Error(w, `{"success": false, "error": "Team name is required"}`, http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO teams (name_th, stadium_id, team_post_ballthai, logo_url)
		VALUES (?, ?, ?, ?)
	`

	result, err := DB.Exec(query, team.NameTh, team.StadiumID, team.TeamPostID, team.LogoURL)
	if err != nil {
		log.Printf("Error creating team: %v", err)
		http.Error(w, `{"success": false, "error": "Failed to create team"}`, http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()

	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"id":      id,
			"message": "Team created successfully",
		},
	}

	json.NewEncoder(w).Encode(response)
}

// UpdateTeam updates an existing team
func UpdateTeam(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	teamID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, `{"success": false, "error": "Invalid team ID"}`, http.StatusBadRequest)
		return
	}

	var team struct {
		Name       string `json:"name"`
		NameTh     string `json:"name_th"`
		StadiumID  *int   `json:"stadium_id"`
		TeamPostID *int   `json:"team_post_id"`
		LogoURL    string `json:"logo_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		http.Error(w, `{"success": false, "error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	query := `
		UPDATE teams 
		SET name_th = ?, stadium_id = ?, team_post_ballthai = ?, logo_url = ?
		WHERE id = ?
	`

	_, err = DB.Exec(query, team.NameTh, team.StadiumID, team.TeamPostID, team.LogoURL, teamID)
	if err != nil {
		log.Printf("Error updating team: %v", err)
		http.Error(w, `{"success": false, "error": "Failed to update team"}`, http.StatusInternalServerError)
		return
	}

	response := APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Team updated successfully"},
	}

	json.NewEncoder(w).Encode(response)
}

// DeleteTeam deletes a team
func DeleteTeam(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	teamID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, `{"success": false, "error": "Invalid team ID"}`, http.StatusBadRequest)
		return
	}

	// Check if team has players
	var playerCount int
	err = DB.QueryRow("SELECT COUNT(*) FROM players WHERE team_id = ?", teamID).Scan(&playerCount)
	if err != nil {
		log.Printf("Error checking team players: %v", err)
		http.Error(w, `{"success": false, "error": "Failed to check team players"}`, http.StatusInternalServerError)
		return
	}

	if playerCount > 0 {
		http.Error(w, `{"success": false, "error": "Cannot delete team with existing players"}`, http.StatusConflict)
		return
	}

	query := `DELETE FROM teams WHERE id = ?`
	_, err = DB.Exec(query, teamID)
	if err != nil {
		log.Printf("Error deleting team: %v", err)
		http.Error(w, `{"success": false, "error": "Failed to delete team"}`, http.StatusInternalServerError)
		return
	}

	response := APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Team deleted successfully"},
	}

	json.NewEncoder(w).Encode(response)
}

// SearchTeams searches teams by name
func SearchTeams(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	searchQuery := r.URL.Query().Get("q")
	if searchQuery == "" {
		GetTeams(w, r) // Return all teams if no search query
		return
	}

	query := `
		SELECT t.id, t.name_th, t.team_post_ballthai, t.stadium_id, s.name as stadium_name, 
		       t.logo_url, NULL as established_year
		FROM teams t 
		LEFT JOIN stadiums s ON t.stadium_id = s.id
		WHERE t.name_th LIKE ?
		ORDER BY t.name_th
	`

	rows, err := DB.Query(query, "%"+searchQuery+"%")
	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var teams []Team
	for rows.Next() {
		var team Team
		if err := rows.Scan(&team.ID, &team.NameTh, &team.TeamPostID, &team.StadiumID,
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

// UploadTeamLogo handles team logo upload
func UploadTeamLogo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	teamID := vars["id"]

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, `{"success": false, "error": "Failed to parse multipart form"}`, http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("logo")
	if err != nil {
		http.Error(w, `{"success": false, "error": "Failed to get file from form"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file type
	if !isValidImageType(handler.Filename) {
		http.Error(w, `{"success": false, "error": "Invalid file type. Only JPG, PNG, and GIF are allowed"}`, http.StatusBadRequest)
		return
	}

	// Create upload directory if it doesn't exist
	uploadDir := "./img/teams"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("Error creating upload directory: %v", err)
		http.Error(w, `{"success": false, "error": "Failed to create upload directory"}`, http.StatusInternalServerError)
		return
	}

	// Generate unique filename
	timestamp := time.Now().Format("20060102-150405")
	ext := filepath.Ext(handler.Filename)
	filename := fmt.Sprintf("%s-team-%s%s", timestamp, teamID, ext)
	filepath := filepath.Join(uploadDir, filename)

	// Create the file
	dst, err := os.Create(filepath)
	if err != nil {
		log.Printf("Error creating file: %v", err)
		http.Error(w, `{"success": false, "error": "Failed to create file"}`, http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, file); err != nil {
		log.Printf("Error copying file: %v", err)
		http.Error(w, `{"success": false, "error": "Failed to save file"}`, http.StatusInternalServerError)
		return
	}

	// Update team logo URL in database
	logoURL := fmt.Sprintf("/img/teams/%s", filename)
	_, err = DB.Exec("UPDATE teams SET logo_url = ? WHERE id = ?", logoURL, teamID)
	if err != nil {
		log.Printf("Error updating team logo: %v", err)
		http.Error(w, `{"success": false, "error": "Failed to update team logo"}`, http.StatusInternalServerError)
		return
	}

	response := APIResponse{
		Success: true,
		Data: map[string]string{
			"message":  "Logo uploaded successfully",
			"logo_url": logoURL,
		},
	}

	json.NewEncoder(w).Encode(response)
}

// Helper function to validate image file types
func isValidImageType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExtensions := []string{".jpg", ".jpeg", ".png", ".gif"}

	for _, validExt := range validExtensions {
		if ext == validExt {
			return true
		}
	}
	return false
}
