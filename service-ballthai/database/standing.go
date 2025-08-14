
package database

import (
   "database/sql"
   "fmt"
   "go-ballthai-scraper/models"
   "log"
)

// UpdateStandingByID อัปเดตข้อมูล standings ตาม id
func UpdateStandingByID(db *sql.DB, id int, standing models.StandingDB) error {
   updateQuery := `
	   UPDATE standings SET
		   status = ?, matches_played = ?, wins = ?, draws = ?, losses = ?,
		   goals_for = ?, goals_against = ?, goal_difference = ?, points = ?, current_rank = ?
	   WHERE id = ?
   `
   _, err := db.Exec(updateQuery,
	   standing.Status, standing.MatchesPlayed, standing.Wins, standing.Draws, standing.Losses,
	   standing.GoalsFor, standing.GoalsAgainst, standing.GoalDifference, standing.Points, standing.CurrentRank,
	   id,
   )
   return err
}

// GetStandingsByLeagueID คืน standings ทั้งหมดของลีกที่ระบุ
func GetStandingsByLeagueID(db *sql.DB, leagueID int) ([]models.StandingDB, error) {
	rows, err := db.Query(`SELECT s.id, s.league_id, s.team_id, t.name_th as team_name, s.stage_id, s.status, s.matches_played, s.wins, s.draws, s.losses, s.goals_for, s.goals_against, s.goal_difference, s.points, s.current_rank FROM standings s LEFT JOIN teams t ON s.team_id = t.id WHERE s.league_id = ? ORDER BY s.points DESC, s.goal_difference DESC, s.wins DESC`, leagueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var standings []models.StandingDB
	for rows.Next() {
		var s models.StandingDB
		if err := rows.Scan(&s.ID, &s.LeagueID, &s.TeamID, &s.TeamName, &s.StageID, &s.Status, &s.MatchesPlayed, &s.Wins, &s.Draws, &s.Losses, &s.GoalsFor, &s.GoalsAgainst, &s.GoalDifference, &s.Points, &s.CurrentRank); err != nil {
			return nil, err
		}
		standings = append(standings, s)
	}
	return standings, nil
}

// InsertOrUpdateStanding inserts or updates a league standing record in the database
func InsertOrUpdateStanding(db *sql.DB, standing models.StandingDB) error {
	var existingStandingID int
	query := "SELECT id FROM standings WHERE league_id = ? AND team_id = ? AND stage_id = ?"
	err := db.QueryRow(query, standing.LeagueID, standing.TeamID, standing.StageID).Scan(&existingStandingID)

	if err == sql.ErrNoRows {
		// Insert new standing (เพิ่ม stage_id, status)
		insertQuery := `
		       INSERT INTO standings (
			       league_id, team_id, stage_id, status, matches_played, wins, draws, losses,
			       goals_for, goals_against, goal_difference, points, current_rank
		       ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	       `
		    statusVal := standing.Status
		    if !statusVal.Valid { statusVal.Int64 = 0; statusVal.Valid = true }
		    _, err := db.Exec(insertQuery,
			    standing.LeagueID, standing.TeamID, standing.StageID, statusVal, standing.MatchesPlayed,
			    standing.Wins, standing.Draws, standing.Losses, standing.GoalsFor,
			    standing.GoalsAgainst, standing.GoalDifference, standing.Points, standing.CurrentRank,
		    )
		if err != nil {
			return fmt.Errorf("failed to insert standing for team %d in league %d stage %v: %w", standing.TeamID, standing.LeagueID, standing.StageID, err)
		}
		log.Printf("Inserted new standing for team %d in league %d stage %v", standing.TeamID, standing.LeagueID, standing.StageID)
	} else if err != nil {
		return fmt.Errorf("failed to query existing standing for team %d in league %d stage %v: %w", standing.TeamID, standing.LeagueID, standing.StageID, err)
	} else {
		// Update existing standing (เพิ่ม stage_id, status)
		updateQuery := `
		       UPDATE standings SET
			       status = ?, matches_played = ?, wins = ?, draws = ?, losses = ?,
			       goals_for = ?, goals_against = ?, goal_difference = ?, points = ?, current_rank = ?
		       WHERE id = ?
	       `
		    statusVal := standing.Status
		    if !statusVal.Valid { statusVal.Int64 = 0; statusVal.Valid = true }
		    _, err := db.Exec(updateQuery,
			    statusVal, standing.MatchesPlayed, standing.Wins, standing.Draws, standing.Losses,
			    standing.GoalsFor, standing.GoalsAgainst, standing.GoalDifference, standing.Points, standing.CurrentRank,
			    existingStandingID,
		    )
		if err != nil {
			return fmt.Errorf("failed to update standing for team %d in league %d stage %v: %w", standing.TeamID, standing.LeagueID, standing.StageID, err)
		}
		log.Printf("Updated existing standing for team %d in league %d stage %v (ID: %d)", standing.TeamID, standing.LeagueID, standing.StageID, existingStandingID)
	}
	return nil
}
