
package handlers
import (
	"encoding/json"
	"go-ballthai-scraper/database"
	"net/http"
	"strconv"
)

// UpdateStanding อัปเดตข้อมูล standings ตาม id
func UpdateStanding(w http.ResponseWriter, r *http.Request) {
	// log เพื่อตรวจสอบว่า handler ถูกเรียก
	println("[DEBUG] UpdateStanding called for path:", r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	idStr := ""
	// รองรับทั้ง mux.Vars และ query param (กรณีใช้ mux)
	if r.URL.Path != "" {
		// /api/standings/{id}
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
	// TODO: update standing in DB (mock)
	// คุณควรเขียนฟังก์ชัน database.UpdateStandingByID(id, req) จริงในภายหลัง
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      id,
		"data":    req,
	})
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
