package scraper

import (
	"database/sql"
	"encoding/json" // Added for json.Unmarshal
	"fmt"
	"log"
	// "path/filepath" // Not directly used in this file
	// "time" // Not directly used in this file

	"go-ballthai-scraper/database" // Ensure this module name matches your go.mod
	"go-ballthai-scraper/models"  // Ensure this module name matches your go.mod
)

// ScrapeStadiums scrapes stadium data from the API and saves it to the database
func ScrapeStadiums(db *sql.DB) error {
	baseURL := "https://competition.tl.prod.c0d1um.io/thaileague/api/stadium-public/all_stadiums_search/?page="
	maxPages := 5 // As seen in original PHP

	for page := 1; page <= maxPages; page++ {
		url := fmt.Sprintf("%s%d", baseURL, page)
		
		var apiResponse struct {
			Results []models.StadiumAPI `json:"results"`
		}
		// เรียกใช้ FetchAndParseAPI จาก scraper/api.go
		err := FetchAndParseAPI(url, &apiResponse)
		if err != nil {
			log.Printf("Error fetching stadiums from page %d: %v", page, err)
			continue // Continue to next page even if one fails
		}

		for _, apiStadium := range apiResponse.Results {
			// Download stadium photo and get only the filename
			photoFilename := ""
			if apiStadium.Photo != "" {
				// เรียกใช้ DownloadImage จาก scraper/api.go
				downloadedFilename, err := DownloadImage(apiStadium.Photo, "./img/stadiums")
				if err != nil {
					log.Printf("Warning: Failed to download stadium photo for %s: %v", apiStadium.Name, err)
				} else {
					photoFilename = downloadedFilename // Store only the filename
				}
			}

			// Handle Country field which might be string or object
			countryName := ""
			countryCode := ""
			if apiStadium.Country != nil {
				// Try to unmarshal as CountryAPI struct
				var countryObj models.CountryAPI
				if err := json.Unmarshal(apiStadium.Country, &countryObj); err == nil {
					countryName = countryObj.Name
					countryCode = countryObj.Code
				} else {
					// If it failed, try to unmarshal as a string
					var countryStr string
					if err := json.Unmarshal(apiStadium.Country, &countryStr); err == nil {
						countryName = countryStr // Store the string directly as name
						// No code available from string, leave empty
					} else {
						log.Printf("Warning: Could not unmarshal country field for stadium %s. Raw: %s, Error: %v", apiStadium.Name, string(apiStadium.Country), err)
					}
				}
			}

			// Get Team ID (if club_names exists and is relevant)
			teamID := sql.NullInt64{Valid: false}
			if len(apiStadium.ClubNames) > 0 {
				clubNameTH := apiStadium.ClubNames[0].TH
				if clubNameTH != "" {
					// Pass empty logo for now, as stadium API might not provide team logo
					tID, err := database.GetTeamIDByThaiName(db, clubNameTH, "") 
					if err != nil {
						log.Printf("Warning: Failed to get team ID for %s: %v", clubNameTH, err)
					} else {
						teamID = sql.NullInt64{Int64: int64(tID), Valid: true}
					}
				}
			}

			// Prepare StadiumDB struct
			stadiumDB := models.StadiumDB{
				StadiumRefID:    apiStadium.ID,
				Name:            apiStadium.Name,
				NameEN:          sql.NullString{String: apiStadium.NameEN, Valid: apiStadium.NameEN != ""},
				ShortName:       sql.NullString{String: apiStadium.ShortName, Valid: apiStadium.ShortName != ""},
				ShortNameEN:     sql.NullString{String: apiStadium.ShortNameEN, Valid: apiStadium.ShortNameEN != ""},
				YearEstablished: sql.NullInt64{Int64: int64(apiStadium.CreatedYear), Valid: apiStadium.CreatedYear != 0},
				Capacity:        sql.NullInt64{Int64: int64(apiStadium.Capacity), Valid: apiStadium.Capacity != 0},
				Latitude:        sql.NullFloat64{Float64: apiStadium.Latitude, Valid: apiStadium.Latitude != 0},
				Longitude:       sql.NullFloat64{Float64: apiStadium.Longitude, Valid: apiStadium.Longitude != 0},
				PhotoURL:        sql.NullString{String: photoFilename, Valid: photoFilename != ""}, // Store filename
				CountryName:     sql.NullString{String: countryName, Valid: countryName != ""},   // Use parsed country name
				CountryCode:     sql.NullString{String: countryCode, Valid: countryCode != ""},   // Use parsed country code
				TeamID:          teamID,
			}

			// Insert or Update stadium in DB
			err = database.InsertOrUpdateStadium(db, stadiumDB)
			if err != nil {
				log.Printf("Error saving stadium %s to DB: %v", apiStadium.Name, err)
			}
		}
	}
	return nil
}
