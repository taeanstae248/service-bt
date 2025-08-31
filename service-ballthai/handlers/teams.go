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

	"go-ballthai-scraper/database"

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

	// Read raw body for debugging and decoding
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, `{"success": false, "error": "Failed to read request body"}`, http.StatusBadRequest)
		return
	}

	if len(bodyBytes) == 0 {
		http.Error(w, `{"success": false, "error": "Empty request body"}`, http.StatusBadRequest)
		return
	}

	// Log the incoming body for debugging (label for CreateTeam fixed earlier)
	log.Printf("CreateTeam request body: %s", string(bodyBytes))

	if err := json.Unmarshal(bodyBytes, &team); err != nil {
		log.Printf("JSON unmarshal error in CreateTeam: %v; body: %s", err, string(bodyBytes))
		http.Error(w, `{"success": false, "error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if team.NameTh == "" {
		if team.Name != "" {
			team.NameTh = team.Name
		} else {
			http.Error(w, `{"success": false, "error": "Team name is required"}`, http.StatusBadRequest)
			return
		}
	}

	// Normalize logo URL before saving
	normalizedLogo := ""
	if team.LogoURL != "" {
		normalizedLogo = database.NormalizeLogoURL(team.LogoURL)
	}

	query := `
		INSERT INTO teams (name_th, stadium_id, team_post_ballthai, logo_url)
		VALUES (?, ?, ?, ?)
	`

	result, err := DB.Exec(query, team.NameTh, team.StadiumID, team.TeamPostID, normalizedLogo)
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

	// New implementation: accept number/string/null for numeric fields and build dynamic UPDATE
	var raw struct {
		Name       *string          `json:"name"`
		NameTh     *string          `json:"name_th"`
		StadiumRaw *json.RawMessage `json:"stadium_id"`
		TeamPostRaw *json.RawMessage `json:"team_post_id"`
		LogoURL    *string          `json:"logo_url"`
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("UpdateTeam: read body error: %v", err)
		http.Error(w, `{"success": false, "error": "Failed to read request body"}`, http.StatusBadRequest)
		return
	}
	if len(bodyBytes) == 0 {
		http.Error(w, `{"success": false, "error": "Empty request body"}`, http.StatusBadRequest)
		return
	}
	log.Printf("UpdateTeam request body: %s", string(bodyBytes))

	if err := json.Unmarshal(bodyBytes, &raw); err != nil {
		log.Printf("JSON unmarshal error in UpdateTeam: %v; body: %s", err, string(bodyBytes))
		http.Error(w, `{"success": false, "error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// helper to parse optional int from raw JSON that may be number, string, or null
	parseOptionalInt := func(raw *json.RawMessage) (*int, error) {
		if raw == nil {
			return nil, nil
		}
		if string(*raw) == "null" {
			return nil, nil
		}
		var n int
		if err := json.Unmarshal(*raw, &n); err == nil {
			return &n, nil
		}
		var s string
		if err := json.Unmarshal(*raw, &s); err == nil {
			s = strings.TrimSpace(s)
			if s == "" {
				return nil, nil
			}
			v, err := strconv.Atoi(s)
			if err != nil {
				return nil, err
			}
			return &v, nil
		}
		return nil, fmt.Errorf("unsupported value for int field: %s", string(*raw))
	}

	stadiumID, err := parseOptionalInt(raw.StadiumRaw)
	if err != nil {
		log.Printf("Invalid stadium_id value: %v", err)
		http.Error(w, `{"success": false, "error": "Invalid stadium_id"}`, http.StatusBadRequest)
		return
	}
	teamPostID, err := parseOptionalInt(raw.TeamPostRaw)
	if err != nil {
		log.Printf("Invalid team_post_id value: %v", err)
		http.Error(w, `{"success": false, "error": "Invalid team_post_id"}`, http.StatusBadRequest)
		return
	}

	var setParts []string
	var args []interface{}

	if raw.NameTh != nil {
		setParts = append(setParts, "name_th = ?")
		args = append(args, *raw.NameTh)
	} else if raw.Name != nil {
		setParts = append(setParts, "name_th = ?")
		args = append(args, *raw.Name)
	}
	if stadiumID != nil {
		setParts = append(setParts, "stadium_id = ?")
		args = append(args, stadiumID)
	}
	if teamPostID != nil {
		setParts = append(setParts, "team_post_ballthai = ?")
		args = append(args, teamPostID)
	}
	if raw.LogoURL != nil {
		if *raw.LogoURL != "" {
			normalized := database.NormalizeLogoURL(*raw.LogoURL)
			setParts = append(setParts, "logo_url = ?")
			args = append(args, normalized)
		} else {
			setParts = append(setParts, "logo_url = NULL")
		}
	}

	if len(setParts) == 0 {
		response := APIResponse{Success: true, Data: map[string]string{"message": "No changes"}}
		json.NewEncoder(w).Encode(response)
		return
	}

	query := fmt.Sprintf("UPDATE teams SET %s WHERE id = ?", strings.Join(setParts, ", "))
	args = append(args, teamID)

	_, err = DB.Exec(query, args...)
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
