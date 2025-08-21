package database

import (
	"database/sql"
)

// GetStageNameByID คืน stage_name จาก stage_id
func GetStageNameByID(db *sql.DB, stageID int64) (string, error) {
	var name string
	err := db.QueryRow("SELECT stage_name FROM stage WHERE id = ?", stageID).Scan(&name)
	if err != nil {
		return "", err
	}
	return name, nil
}
