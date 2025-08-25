package database

import (
	"database/sql"
	"go-ballthai-scraper/models"
)

// GetStandingsByLeagueIDAndStageID คืน standings ของลีกและ stage ที่ระบุ
func GetStandingsByLeagueIDAndStageID(db *sql.DB, leagueID int, stageID sql.NullInt64) ([]models.StandingDB, error) {
	var rows *sql.Rows
	var err error
	if stageID.Valid {
		rows, err = db.Query(`SELECT s.id, s.league_id, s.team_id, t.name_th as team_name, t.team_post_ballthai as team_post, s.stage_id, s.status, s.matches_played, s.wins, s.draws, s.losses, s.goals_for, s.goals_against, s.goal_difference, s.points, s.current_rank FROM standings s LEFT JOIN teams t ON s.team_id = t.id WHERE s.league_id = ? AND s.stage_id = ? ORDER BY s.current_rank ASC, s.id ASC`, leagueID, stageID.Int64)
	} else {
		rows, err = db.Query(`SELECT s.id, s.league_id, s.team_id, t.name_th as team_name, t.team_post_ballthai as team_post, s.stage_id, s.status, s.matches_played, s.wins, s.draws, s.losses, s.goals_for, s.goals_against, s.goal_difference, s.points, s.current_rank FROM standings s LEFT JOIN teams t ON s.team_id = t.id WHERE s.league_id = ? AND s.stage_id IS NULL ORDER BY s.current_rank ASC, s.id ASC`, leagueID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var standings []models.StandingDB
	for rows.Next() {
		var s models.StandingDB
	if err := rows.Scan(&s.ID, &s.LeagueID, &s.TeamID, &s.TeamName, &s.TeamPost, &s.StageID, &s.Status, &s.MatchesPlayed, &s.Wins, &s.Draws, &s.Losses, &s.GoalsFor, &s.GoalsAgainst, &s.GoalDifference, &s.Points, &s.CurrentRank); err != nil {
			return nil, err
		}
		standings = append(standings, s)
	}
	return standings, nil
}
