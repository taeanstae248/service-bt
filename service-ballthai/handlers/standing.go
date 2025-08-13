package handlers

import (
	"encoding/json"
	"go-ballthai-scraper/database"
	"net/http"
	"strconv"
)

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
		http.Error(w, `{"success": false, "error": "failed to fetch standings"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    standings,
	})
}
