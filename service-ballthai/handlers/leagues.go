package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"go-ballthai-scraper/database"

	"github.com/gorilla/mux"
)

// CreateLeague สร้างลีกใหม่
func CreateLeague(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var league struct {
		Name        string `json:"name"`
		Thaileageid *int   `json:"thaileageid"`
	}

	if err := json.NewDecoder(r.Body).Decode(&league); err != nil {
		http.Error(w, `{"success": false, "error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validation
	if league.Name == "" {
		http.Error(w, `{"success": false, "error": "League name is required"}`, http.StatusBadRequest)
		return
	}

	// Create league in database
	var result sql.Result
	var err error
	if league.Thaileageid != nil {
		result, err = database.DB.Exec("INSERT INTO leagues (name, thaileageid) VALUES (?, ?)", league.Name, league.Thaileageid)
	} else {
		result, err = database.DB.Exec("INSERT INTO leagues (name) VALUES (?)", league.Name)
	}
	if err != nil {
		log.Printf("Failed to create league: %v", err)
		if isDuplicateEntry(err) {
			http.Error(w, `{"success": false, "error": "League name already exists"}`, http.StatusConflict)
		} else {
			http.Error(w, `{"success": false, "error": "Failed to create league"}`, http.StatusInternalServerError)
		}
		return
	}

	id, _ := result.LastInsertId()
	createdLeague := map[string]interface{}{
		"id":   id,
		"name": league.Name,
		"thaileageid": league.Thaileageid,
	}

	response := map[string]interface{}{
		"success": true,
		"data":    createdLeague,
	}

	json.NewEncoder(w).Encode(response)
}

// UpdateLeague อัปเดตข้อมูลลีก
func UpdateLeague(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, `{"success": false, "error": "Invalid league ID"}`, http.StatusBadRequest)
		return
	}

	var league struct {
		Name        string `json:"name"`
		Thaileageid *int   `json:"thaileageid"`
	}

	if err := json.NewDecoder(r.Body).Decode(&league); err != nil {
		http.Error(w, `{"success": false, "error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validation
	if league.Name == "" {
		http.Error(w, `{"success": false, "error": "League name is required"}`, http.StatusBadRequest)
		return
	}

	// Update league in database
	var result sql.Result
	if league.Thaileageid != nil {
		result, err = database.DB.Exec("UPDATE leagues SET name = ?, thaileageid = ? WHERE id = ?", league.Name, league.Thaileageid, id)
	} else {
		result, err = database.DB.Exec("UPDATE leagues SET name = ?, thaileageid = NULL WHERE id = ?", league.Name, id)
	}
	if err != nil {
		log.Printf("Failed to update league: %v", err)
		if isDuplicateEntry(err) {
			http.Error(w, `{"success": false, "error": "League name already exists"}`, http.StatusConflict)
		} else {
			http.Error(w, `{"success": false, "error": "Failed to update league"}`, http.StatusInternalServerError)
		}
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// อาจเป็นเพราะข้อมูลเหมือนเดิม ให้เช็คว่ามี row นี้จริงไหม
		var exists bool
		err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM leagues WHERE id = ?)", id).Scan(&exists)
		if err != nil || !exists {
			http.Error(w, `{"success": false, "error": "League not found"}`, http.StatusNotFound)
			return
		}
		// ถ้ามี row จริง ให้ถือว่า success (ข้อมูลเหมือนเดิม)
	}

	updatedLeague := map[string]interface{}{
		"id":   id,
		"name": league.Name,
		"thaileageid": league.Thaileageid,
	}

	response := map[string]interface{}{
		"success": true,
		"data":    updatedLeague,
	}

	json.NewEncoder(w).Encode(response)
}

// DeleteLeague ลบลีก
func DeleteLeague(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, `{"success": false, "error": "Invalid league ID"}`, http.StatusBadRequest)
		return
	}

	// Check if league is being used by teams, matches, etc.
	var count int
	err = database.DB.QueryRow("SELECT COUNT(*) FROM teams WHERE team_id IN (SELECT id FROM teams JOIN players ON teams.id = players.team_id WHERE players.league_id = ?)", id).Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Failed to check league usage: %v", err)
		http.Error(w, `{"success": false, "error": "Failed to check league usage"}`, http.StatusInternalServerError)
		return
	}

	if count > 0 {
		http.Error(w, `{"success": false, "error": "Cannot delete league: it is being used by teams/players"}`, http.StatusConflict)
		return
	}

	// Delete league
	result, err := database.DB.Exec("DELETE FROM leagues WHERE id = ?", id)
	if err != nil {
		log.Printf("Failed to delete league: %v", err)
		http.Error(w, `{"success": false, "error": "Failed to delete league"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, `{"success": false, "error": "League not found"}`, http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "League deleted successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// SearchLeagues ค้นหาลีก
func SearchLeagues(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	searchTerm := r.URL.Query().Get("q")

	var query string
	var args []interface{}

	if searchTerm != "" {
		query = "SELECT id, name, thaileageid FROM leagues WHERE name LIKE ? ORDER BY name"
		args = append(args, "%"+searchTerm+"%")
	} else {
		query = "SELECT id, name, thaileageid FROM leagues ORDER BY name"
	}

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		log.Printf("Failed to search leagues: %v", err)
		http.Error(w, `{"success": false, "error": "Failed to search leagues"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var leagues []map[string]interface{}
	for rows.Next() {
		var id int
		var name string
		var thaileageid sql.NullInt64

		if err := rows.Scan(&id, &name, &thaileageid); err != nil {
			log.Printf("Failed to scan league row: %v", err)
			continue
		}

		leagues = append(leagues, map[string]interface{}{
			"id":   id,
			"name": name,
			"thaileageid": func() interface{} { if thaileageid.Valid { return thaileageid.Int64 } else { return nil } }(),
		})
	}

	response := map[string]interface{}{
		"success": true,
		"data":    leagues,
	}

	json.NewEncoder(w).Encode(response)
}

// GetLeagues ดึงข้อมูลลีกทั้งหมด (พร้อม thaileageid)
func GetLeagues(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rows, err := database.DB.Query("SELECT id, name, thaileageid FROM leagues ORDER BY name")
	if err != nil {
		log.Printf("Failed to get leagues: %v", err)
		http.Error(w, `{"success": false, "error": "Failed to get leagues"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var leagues []map[string]interface{}
	for rows.Next() {
		var id int
		var name string
		var thaileageid sql.NullInt64

		if err := rows.Scan(&id, &name, &thaileageid); err != nil {
			log.Printf("Failed to scan league row: %v", err)
			continue
		}

		leagues = append(leagues, map[string]interface{}{
			"id":   id,
			"name": name,
			"thaileageid": func() interface{} { if thaileageid.Valid { return thaileageid.Int64 } else { return nil } }(),
		})
	}

	response := map[string]interface{}{
		"success": true,
		"data":    leagues,
	}

	json.NewEncoder(w).Encode(response)
}

// Helper function to check for duplicate entry errors
func isDuplicateEntry(err error) bool {
	return err != nil && (
	// MySQL duplicate entry error codes
	err.Error() == "Error 1062: Duplicate entry" ||
		// Check for "Duplicate entry" string in error message
		len(err.Error()) > 15 && err.Error()[:15] == "Error 1062: Dup")
}
