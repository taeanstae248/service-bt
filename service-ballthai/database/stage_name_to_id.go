package database

import (
	"database/sql"
)

// GetStageIDByName returns stage id from stage_name (case-insensitive)
func GetStageIDByName(db *sql.DB, stageName string) (int, error) {
	var id int
	err := db.QueryRow("SELECT id FROM stage WHERE LOWER(stage_name) = LOWER(?)", stageName).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
