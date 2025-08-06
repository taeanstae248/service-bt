package scraper

import (
	"database/sql"
	"fmt"
	"log"
	// "path/filepath" // Not directly used in this file
	"time"

	"go-ballthai-scraper/database" // Ensure this module name matches your go.mod
	"go-ballthai-scraper/models"  // Ensure this module name matches your go.mod
)

// ScrapeCoach scrapes coach data from the API and saves it to the database
func ScrapeCoach(db *sql.DB) error {
	baseURL := "https://competition.tl.prod.c0d1um.io/thaileague/api/staff-public/?type=headcoach&page="
	maxPages := 10 // As seen in original PHP

	for page := 1; page <= maxPages; page++ {
		url := fmt.Sprintf("%s%d", baseURL, page)
		
		var apiResponse struct {
			Results []models.CoachAPI `json:"results"`
		}
		err := FetchAndParseAPI(url, &apiResponse)
		if err != nil {
			log.Printf("Error fetching coaches from page %d: %v", page, err)
			continue
		}

		for _, apiCoach := range apiResponse.Results {
			// Download coach photo and get only the filename
			photoFilename := ""
			if apiCoach.Photo != "" {
				downloadedFilename, err := DownloadImage(apiCoach.Photo, "./img/coach")
				if err != nil {
					log.Printf("Warning: Failed to download coach photo for %s: %v", apiCoach.FullName, err)
				} else {
					photoFilename = downloadedFilename // Store only the filename
				}
			}

			// Get Nationality ID
			nationalityID := sql.NullInt64{Valid: false}
			if apiCoach.Nationality.Code != "" {
				nID, err := database.GetNationalityID(db, apiCoach.Nationality.Code, apiCoach.Nationality.Name)
				if err != nil {
					log.Printf("Warning: Failed to get nationality ID for %s: %v", apiCoach.Nationality.Name, err)
				} else {
					nationalityID = sql.NullInt64{Int64: int64(nID), Valid: true}
				}
			}

			// Get Team ID
			teamID := sql.NullInt64{Valid: false}
			if apiCoach.ClubName != "" {
				// Pass empty logo for now, as coach API might not provide team logo
				tID, err := database.GetTeamIDByThaiName(db, apiCoach.ClubName, "") 
				if err != nil {
					log.Printf("Warning: Failed to get team ID for coach %s's club %s: %v", apiCoach.FullName, apiCoach.ClubName, err)
				} else {
					teamID = sql.NullInt64{Int64: int64(tID), Valid: true}
				}
			}

			// Parse BirthDate
			birthDate := sql.NullTime{Valid: false}
			if apiCoach.BirthDate != "" {
				// Assuming API date format is "YYYY-MM-DD"
				parsedDate, err := time.Parse("2006-01-02", apiCoach.BirthDate)
				if err != nil {
					log.Printf("Warning: Failed to parse birth date %s for coach %s: %v", apiCoach.BirthDate, apiCoach.FullName, err)
				} else {
					birthDate = sql.NullTime{Time: parsedDate, Valid: true}
				}
			}

			// Prepare CoachDB struct
			coachDB := models.CoachDB{
				CoachRefID:    sql.NullInt64{Int64: int64(apiCoach.ID), Valid: true},
				Name:          apiCoach.FullName,
				Birthday:      birthDate,
				TeamID:        teamID,
				NationalityID: nationalityID,
				PhotoURL:      sql.NullString{String: photoFilename, Valid: photoFilename != ""}, // Store filename
			}

			// Insert or Update coach in DB
			err = database.InsertOrUpdateCoach(db, coachDB)
			if err != nil {
				log.Printf("Error saving coach %s to DB: %v", apiCoach.FullName, err)
			}
		}
	}
	return nil
}
