package handlers

import (
   "encoding/json"
   "go-ballthai-scraper/database"
   "go-ballthai-scraper/models"
   "net/http"
   "strconv"
   "database/sql"
)

// UpdateStandingsOrder อัปเดต current_rank ของ standings หลายรายการ
func UpdateStandingsOrder(w http.ResponseWriter, r *http.Request) {
   w.Header().Set("Content-Type", "application/json")
   var req struct {
	   LeagueID int `json:"league_id"`
	   Order []struct {
		   ID int `json:"id"`
		   CurrentRank int `json:"current_rank"`
	   } `json:"order"`
   }
   if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	   http.Error(w, `{"success": false, "error": "invalid json"}`, http.StatusBadRequest)
	   return
   }
   // Validate
   if len(req.Order) == 0 {
	   http.Error(w, `{"success": false, "error": "order is required"}`, http.StatusBadRequest)
	   return
   }
   // อัปเดตทีละรายการ
   var updated, failed int
   for _, o := range req.Order {
	   err := database.UpdateStandingRankByID(database.DB, o.ID, o.CurrentRank)
	   if err != nil {
		   failed++
	   } else {
		   updated++
	   }
   }
   json.NewEncoder(w).Encode(map[string]interface{}{
	   "success": failed == 0,
	   "updated": updated,
	   "failed": failed,
   })
}

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

	// Build a map of fields provided by caller to avoid overwriting unspecified fields with zeros
	fields := make(map[string]interface{})
	if v, ok := req["matches_played"]; ok { if i, ok2 := toInt(v); ok2 { fields["matches_played"] = i } }
	if v, ok := req["wins"]; ok { if i, ok2 := toInt(v); ok2 { fields["wins"] = i } }
	if v, ok := req["draws"]; ok { if i, ok2 := toInt(v); ok2 { fields["draws"] = i } }
	if v, ok := req["losses"]; ok { if i, ok2 := toInt(v); ok2 { fields["losses"] = i } }
	if v, ok := req["goals_for"]; ok { if i, ok2 := toInt(v); ok2 { fields["goals_for"] = i } }
	if v, ok := req["goals_against"]; ok { if i, ok2 := toInt(v); ok2 { fields["goals_against"] = i } }
	if v, ok := req["goal_difference"]; ok { if i, ok2 := toInt(v); ok2 { fields["goal_difference"] = i } }
	if v, ok := req["points"]; ok { if i, ok2 := toInt(v); ok2 { fields["points"] = i } }
	if v, ok := req["current_rank"]; ok { if i, ok2 := toInt64(v); ok2 { fields["current_rank"] = i } }
	if v, ok := req["status"]; ok { if i, ok2 := toInt64(v); ok2 { fields["status"] = i } }

	err = database.UpdateStandingFieldsByID(database.DB, id, fields)
	if err != nil {
		println("[ERROR] UpdateStandingByID:", err.Error())
		http.Error(w, `{"success": false, "error": "failed to update standing"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      id,
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
       // ใช้ leagueNameMap สำหรับชื่อเต็มลีก
       leagueNameMap := map[string]string{
	       "t1": "ไทยลีก 1",
	       "t2": "ไทยลีก 2",
	       "t3": "ไทยลีก 3",
	       "fa": "FA Cup",
	       "league_cup": "League Cup",
	       "bgc": "BGC Cup",
	       "samipro": "Samipro",
       }
	w.Header().Set("Content-Type", "application/json")
       leagueIDStr := r.URL.Query().Get("league_id")
       if leagueIDStr == "" {
	       http.Error(w, `{"success": false, "error": "league_id is required"}`, http.StatusBadRequest)
	       return
       }
       // รองรับ league_id เป็นตัวเลข หรือ t1/t2/t3 ที่มีใน leagueNameMap เท่านั้น
       var leagueID int
       var err error
       if _, ok := leagueNameMap[leagueIDStr]; ok {
	       // ถ้าเป็น t1/t2/t3 ให้ map เป็นเลข
	       switch leagueIDStr {
	       case "t1": leagueID = 1
	       case "t2": leagueID = 2
	       case "t3": leagueID = 3
	       default:
		       http.Error(w, `{"success": false, "error": "invalid league_id"}`, http.StatusBadRequest)
		       return
	       }
       } else {
	       leagueID, err = strconv.Atoi(leagueIDStr)
	       if err != nil {
		       http.Error(w, `{"success": false, "error": "invalid league_id"}`, http.StatusBadRequest)
		       return
	       }
       }
	// ...existing code...
       leagueName := ""
       if name, ok := leagueNameMap[leagueIDStr]; ok {
	       leagueName = name
       }
       // รองรับ stage (stage_id) จาก query string
	stageStr := r.URL.Query().Get("stage")
	var standings []models.StandingDB
	if stageStr != "" {
	       // ถ้า stage เป็นตัวเลข ให้ filter ด้วย stage_id
	       var stageID sql.NullInt64
	       if sID, err := strconv.ParseInt(stageStr, 10, 64); err == nil {
		       stageID.Int64 = sID
		       stageID.Valid = true
	       } else {
		       // ถ้าไม่ใช่ตัวเลข ให้ถือว่าไม่ filter stage_id (หรือจะ map ชื่อเป็น id เพิ่มเติมได้)
		       stageID.Valid = false
	       }
	       standings, err = database.GetStandingsByLeagueIDAndStageID(database.DB, leagueID, stageID)
	       if err != nil {
		       println("[ERROR] GetStandingsByLeagueIDAndStageID:", err.Error())
		       http.Error(w, `{"success": false, "error": "failed to fetch standings by stage"}`, http.StatusInternalServerError)
		       return
	       }
       } else {
	       standings, err = database.GetStandingsByLeagueID(database.DB, leagueID)
	       if err != nil {
		       // log error detail for debugging
		       println("[ERROR] GetStandingsByLeagueID:", err.Error())
		       http.Error(w, `{"success": false, "error": "failed to fetch standings"}`, http.StatusInternalServerError)
		       return
	       }
       }
       // เพิ่ม stage_name ให้แต่ละ standing ถ้ามี stage_id
       type standingAPI struct {
	       ID             int             `json:"id"`
	       LeagueID       int             `json:"league_id"`
	       TeamID         int             `json:"team_id"`
	       TeamName       *string         `json:"team_name"`
	       MatchesPlayed  int             `json:"matches_played"`
	       Wins           int             `json:"wins"`
	       Draws          int             `json:"draws"`
	       Losses         int             `json:"losses"`
	       GoalsFor       int             `json:"goals_for"`
	       GoalsAgainst   int             `json:"goals_against"`
	       GoalDifference int             `json:"goal_difference"`
	       Points         int             `json:"points"`
	       CurrentRank    int             `json:"current_rank"`
			   StageName      string          `json:"stage_name"`
			   Status         sql.NullInt64   `json:"status"`
			   LogoHome       *string         `json:"logo_home"`
			   LogoAway       *string         `json:"logo_away"`
       }
       var result []standingAPI
       for _, s := range standings {
	       stageName := ""
	       if s.StageID.Valid {
		       name, err := database.GetStageNameByID(database.DB, s.StageID.Int64)
		       if err == nil {
			       stageName = name
		       }
	       }
	       // fallback: ถ้า stageName ยังว่าง ให้ลองใช้ s.TeamName (กรณี scraper เคยบันทึก stage_name ลง DB โดยตรง)
	       if stageName == "" && s.TeamName != nil {
		       // ลอง parse จาก team_name ถ้ามีรูปแบบ "... (stage)" เช่น "ทีม A (โซนเหนือ)"
		       tn := *s.TeamName
		       if idx := len(tn) - 1; idx > 0 && tn[idx] == ')' {
			       if open := idx - 1; open > 0 {
				       for ; open >= 0 && tn[open] != '('; open-- {}
				       if open >= 0 && open < idx-1 {
					       stageName = tn[open+1 : idx]
				       }
			       }
		       }
	       }
	       currentRank := 0
	       if s.CurrentRank.Valid {
		       currentRank = int(s.CurrentRank.Int64)
	       }
			   result = append(result, standingAPI{
		       ID:             s.ID,
		       LeagueID:       s.LeagueID,
		       TeamID:         s.TeamID,
		       TeamName:       s.TeamName,
		       MatchesPlayed:  s.MatchesPlayed,
		       Wins:           s.Wins,
		       Draws:          s.Draws,
		       Losses:         s.Losses,
		       GoalsFor:       s.GoalsFor,
		       GoalsAgainst:   s.GoalsAgainst,
		       GoalDifference: s.GoalDifference,
		       Points:         s.Points,
		       CurrentRank:    currentRank,
		       StageName:      stageName,
				   Status:         s.Status,
				   LogoHome:       s.TeamLogo,
				   LogoAway:       s.TeamLogo,
	       })
       }
       json.NewEncoder(w).Encode(map[string]interface{}{
	       "success": true,
	       "data":    result,
	       "league_name": leagueName,
       })
}
