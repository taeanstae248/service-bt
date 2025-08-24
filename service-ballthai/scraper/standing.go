
package scraper

import (
	"database/sql"
	"log"
	"fmt"
	"go-ballthai-scraper/database" // ตรวจสอบให้แน่ใจว่าชื่อโมดูลตรงกับ go.mod ของคุณ
	"go-ballthai-scraper/models"   // ตรวจสอบให้แน่ใจว่าชื่อโมดูลตรงกับ go.mod ของคุณ
)

// ScrapeStandings ดึงข้อมูลตารางคะแนนลีกจาก API และบันทึกลงฐานข้อมูล
func ScrapeStandings(db *sql.DB) error {
	   leagues, err := database.GetAllLeagues(db)
	   if err != nil {
		   return err
	   }
	   for _, league := range leagues {
		   if league.ThaileageID == 0 {
			   continue
		   }
		   url := fmt.Sprintf("https://competition.tl.prod.c0d1um.io/thaileague/api/stage-standing-public/?tournament=%d", league.ThaileageID)
		   log.Printf("Scraping standings for %s (%s)", league.Name, url)

		   var apiResponse []models.StandingAPI
		   err := FetchAndParseAPI(url, &apiResponse)
		   if err != nil {
			   log.Printf("Error fetching standings for %s: %v", league.Name, err)
			   continue
		   }

		   log.Printf("Fetched %d standing entries for league %s", len(apiResponse), league.Name)

		   for _, apiStanding := range apiResponse {
			   // รับ Team ID
			   teamID := 0
			   if apiStanding.TournamentTeamName != "" {
				   tID, err := database.GetTeamIDByThaiName(db, apiStanding.TournamentTeamName, apiStanding.TournamentTeamLogo)
				   if err != nil {
					   log.Printf("Warning: Failed to get team ID for standing team '%s': %v", apiStanding.TournamentTeamName, err)
					   // try to ensure team exists (download logo/insert team) if helper available
					   if err2 := ensureTeamAndLogo(db, apiStanding.TournamentTeamName); err2 != nil {
						   log.Printf("Also failed to ensure team '%s': %v", apiStanding.TournamentTeamName, err2)
						   continue
					   }
					   tID2, err3 := database.GetTeamIDByThaiName(db, apiStanding.TournamentTeamName, apiStanding.TournamentTeamLogo)
					   if err3 != nil {
						   log.Printf("Still failed to get team ID for '%s' after ensure: %v", apiStanding.TournamentTeamName, err3)
						   continue
					   }
					   tID = tID2
				   }
				   teamID = tID
			   } else {
				   log.Printf("Warning: Standing entry for %s has no team name, skipping.", league.Name)
				   continue
			   }

			   // หา stage_id จาก stage_name (ถ้าไม่มีจะ insert ให้)
			   stageID := sql.NullInt64{Valid: false}
			   if apiStanding.StageName != "" {
				   id, err := database.GetStageID(db, apiStanding.StageName, league.ID)
				   if err == nil {
					   stageID = sql.NullInt64{Int64: int64(id), Valid: true}
				   }
			   }
			   standingDB := models.StandingDB{
				   LeagueID:       league.ID,
				   TeamID:         teamID,
				   MatchesPlayed:  apiStanding.MatchPlay,
				   Wins:           apiStanding.Win,
				   Draws:          apiStanding.Draw,
				   Losses:         apiStanding.Lose,
				   GoalsFor:       apiStanding.GoalFor,
				   GoalsAgainst:   apiStanding.GoalAgainst,
				   GoalDifference: apiStanding.GoalDifference,
				   Points:         apiStanding.Point,
				   CurrentRank:    sql.NullInt64{Int64: int64(apiStanding.CurrentRank), Valid: apiStanding.CurrentRank != 0},
				   Status:         sql.NullInt64{Valid: false},
				   StageID:        stageID,
			   }

			   // เช็ค status ก่อนอัปเดต: ถ้า status=0 (OFF) ข้าม ไม่อัปเดต/insert
			   status, err := database.GetStandingStatus(db, league.ID, teamID, stageID)
			   if err != nil {
				   log.Printf("Error checking standing status for team %s: %v", apiStanding.TournamentTeamName, err)
				   // proceed and try to insert/update anyway
			   }
			   // New semantics: status==0 => ON / allow pull; status==1 => OFF / do not pull for this standing id
			   if status.Valid && status.Int64 == 1 {
				   log.Printf("Skipping standing update for team %s because status=1 (OFF)", apiStanding.TournamentTeamName)
				   continue
			   }
			   // proceed to save (status is either NULL or 0)
			   log.Printf("Saving standing: league=%d team=%d stage=%v points=%d rank=%d", standingDB.LeagueID, standingDB.TeamID, standingDB.StageID, standingDB.Points, apiStanding.CurrentRank)
			   err = database.InsertOrUpdateStanding(db, standingDB)
			   if err != nil {
				   log.Printf("Error saving standing for team %s in league %s to DB: %v", apiStanding.TournamentTeamName, league.Name, err)
			   } else {
				   log.Printf("Saved standing for team %s in league %s", apiStanding.TournamentTeamName, league.Name)
			   }
		   }
	   }
	   return nil
	}
