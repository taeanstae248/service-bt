
package handlers
import (
	"encoding/json"
	"go-ballthai-scraper/database"
	"go-ballthai-scraper/models"
	"net/http"
	"strconv"
	"database/sql"
)

// UpdateStanding อัปเดตข้อมูล standings ตาม id
func UpdateStanding(w http.ResponseWriter, r *http.Request) {
	println("[DEBUG] UpdateStanding called for path:", r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	idStr := ""
	if r.URL.Path != "" {
		parts := splitPath(r.URL.Path)
		if len(parts) > 0 {
			idStr = parts[len(parts)-1]
		}
	}
	if idStr == "" {
		http.Error(w, `{"success": false, "error": "id is required"}`, http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"success": false, "error": "invalid id"}`, http.StatusBadRequest)
		return
	}
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"success": false, "error": "invalid json"}`, http.StatusBadRequest)
		return
	}

	// Map json -> StandingDB (handle sql.NullInt64)
	standing := models.StandingDB{ID: id}
	if v, ok := req["matches_played"]; ok { standing.MatchesPlayed, _ = toInt(v) }
	if v, ok := req["wins"]; ok { standing.Wins, _ = toInt(v) }
	if v, ok := req["draws"]; ok { standing.Draws, _ = toInt(v) }
	if v, ok := req["losses"]; ok { standing.Losses, _ = toInt(v) }
	if v, ok := req["goals_for"]; ok { standing.GoalsFor, _ = toInt(v) }
	if v, ok := req["goals_against"]; ok { standing.GoalsAgainst, _ = toInt(v) }
	if v, ok := req["goal_difference"]; ok { standing.GoalDifference, _ = toInt(v) }
	if v, ok := req["points"]; ok { standing.Points, _ = toInt(v) }
	if v, ok := req["current_rank"]; ok {
		i, _ := toInt64(v)
		standing.CurrentRank = sqlNullInt64(i)
	}
		if v, ok := req["status"]; ok { 
			i, _ := toInt64(v) 
			standing.Status = sqlNullInt64(i) 
		} else { 
			standing.Status = sqlNullInt64(0) // default OFF 
		}

	err = database.UpdateStandingByID(database.DB, id, standing)
	if err != nil {
		println("[ERROR] UpdateStandingByID:", err.Error())
		http.Error(w, `{"success": false, "error": "failed to update standing"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      id,
		"data":    standing,
	})
}

// toInt helper
func toInt(v interface{}) (int, bool) {
	switch val := v.(type) {
	case float64:
		return int(val), true
	case int:
		return val, true
	case int64:
		return int(val), true
	case string:
		i, err := strconv.Atoi(val)
		if err == nil { return i, true }
	}
	return 0, false
}
func toInt64(v interface{}) (int64, bool) {
	switch val := v.(type) {
	case float64:
		return int64(val), true
	case int:
		return int64(val), true
	case int64:
		return val, true
	case string:
		i, err := strconv.ParseInt(val, 10, 64)
		if err == nil { return i, true }
	}
	return 0, false
}
func sqlNullInt64(i int64) (n sql.NullInt64) {
	n.Int64 = i
	n.Valid = true
	return
}

// splitPath แยก path เป็น slice (เช่น /api/standings/16 -> ["api","standings","16"])
func splitPath(path string) []string {
	var out []string
	seg := ""
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if seg != "" {
				out = append(out, seg)
				seg = ""
			}
		} else {
			seg += string(path[i])
		}
	}
	if seg != "" {
		out = append(out, seg)
	}
	return out
}

// GetStandings คืนข้อมูล standings ตาม league_id
func GetStandings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	leagueIDStr := r.URL.Query().Get("league_id")
	if leagueIDStr == "" {
		http.Error(w, `{"success": false, "error": "league_id is required"}`, http.StatusBadRequest)
		return
	}
	leagueID, err := strconv.Atoi(leagueIDStr)
	if err != nil {
		http.Error(w, `{"success": false, "error": "invalid league_id"}`, http.StatusBadRequest)
		return
	}
	standings, err := database.GetStandingsByLeagueID(database.DB, leagueID)
	if err != nil {
		// log error detail for debugging
		println("[ERROR] GetStandingsByLeagueID:", err.Error())
		http.Error(w, `{"success": false, "error": "failed to fetch standings"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    standings,
	})
}
