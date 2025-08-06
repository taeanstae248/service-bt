package scraper

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"go-ballthai-scraper/database" // แก้ไข: ตรวจสอบให้แน่ใจว่าชื่อโมดูลตรงกับ go.mod ของคุณ
	"go-ballthai-scraper/models"   // แก้ไข: ตรวจสอบให้แน่ใจว่าชื่อโมดูลตรงกับ go.mod ของคุณ
)

// ScrapeStadiums ดึงข้อมูลสนามจาก API และบันทึกลงฐานข้อมูล
func ScrapeStadiums(db *sql.DB) error {
	baseURL := "https://competition.tl.prod.c0d1um.io/thaileague/api/stadium-public/all_stadiums_search/?page="
	maxPages := 5 // ตามที่เห็นใน PHP ต้นฉบับ

	for page := 1; page <= maxPages; page++ {
		url := fmt.Sprintf("%s%d", baseURL, page)

		var apiResponse struct {
			Results []models.StadiumAPI `json:"results"`
		}
		err := FetchAndParseAPI(url, &apiResponse)
		if err != nil {
			log.Printf("Error fetching stadiums from page %d: %v", page, err)
			continue // ดำเนินการไปยังหน้าถัดไปแม้ว่าหน้าปัจจุบันจะล้มเหลว
		}

		for _, apiStadium := range apiResponse.Results {
			// ดาวน์โหลดรูปภาพสนาม
			photoPath := ""
			if apiStadium.Photo != "" {
				downloadedPath, err := DownloadImage(apiStadium.Photo, "./img/stadiums")
				if err != nil {
					log.Printf("Warning: Failed to download stadium photo for %s: %v", apiStadium.Name, err)
				} else {
					photoPath = downloadedPath
				}
			}

			// รับ Team ID (ถ้ามี club_names และเกี่ยวข้อง)
			teamID := sql.NullInt64{Valid: false}
			if len(apiStadium.ClubNames) > 0 {
				clubNameTH := apiStadium.ClubNames[0].TH
				if clubNameTH != "" {
					tID, err := database.GetTeamIDByThaiName(db, clubNameTH, "") // ส่งโลโก้ว่างเปล่าไปก่อน
					if err != nil {
						log.Printf("Warning: Failed to get team ID for %s: %v", clubNameTH, err)
					} else {
						teamID = sql.NullInt64{Int64: int64(tID), Valid: true}
					}
				}
			}

			// Define a struct to unmarshal the country JSON
			var countryInfo struct {
				Name string `json:"name"`
				Code string `json:"code"`
			}
			// Handle country which can be an object or a string
			if len(apiStadium.Country) > 0 {
				if apiStadium.Country[0] == '{' {
					// It's an object, unmarshal into struct
					if err := json.Unmarshal(apiStadium.Country, &countryInfo); err != nil {
						log.Printf("Warning: Failed to unmarshal country object for stadium %s: %v", apiStadium.Name, err)
					}
				} else {
					// It's likely a string, unmarshal into a string variable
					var countryName string
					if err := json.Unmarshal(apiStadium.Country, &countryName); err == nil {
						countryInfo.Name = countryName
					} else {
						log.Printf("Warning: Failed to unmarshal country string for stadium %s: %v", apiStadium.Name, err)
					}
				}
			}

			// เตรียมโครงสร้าง StadiumDB
			stadiumDB := models.StadiumDB{
				StadiumRefID:    apiStadium.ID,
				Name:            apiStadium.Name,
				NameEN:          sql.NullString{String: apiStadium.NameEN, Valid: apiStadium.NameEN != ""},
				ShortName:       sql.NullString{String: apiStadium.ShortName, Valid: apiStadium.ShortName != ""},
				ShortNameEN:     sql.NullString{String: apiStadium.ShortNameEN, Valid: apiStadium.ShortNameEN != ""},
				YearEstablished: sql.NullInt64{Int64: 0, Valid: false}, // Will be populated after conversion
				Capacity:        sql.NullInt64{Int64: int64(apiStadium.Capacity), Valid: apiStadium.Capacity != 0},
				Latitude:        sql.NullFloat64{Float64: apiStadium.Latitude, Valid: apiStadium.Latitude != 0},
				Longitude:       sql.NullFloat64{Float64: apiStadium.Longitude, Valid: apiStadium.Longitude != 0},
				PhotoURL:        sql.NullString{String: photoPath, Valid: photoPath != ""},
				CountryName:     sql.NullString{String: countryInfo.Name, Valid: countryInfo.Name != ""},
				CountryCode:     sql.NullString{String: countryInfo.Code, Valid: countryInfo.Code != ""},
				TeamID:          teamID,
			}

			// Convert CreatedYear from string to int
			if year, err := strconv.Atoi(apiStadium.CreatedYear); err == nil {
				stadiumDB.YearEstablished = sql.NullInt64{Int64: int64(year), Valid: true}
			} else if apiStadium.CreatedYear != "" {
				log.Printf("Warning: Could not convert CreatedYear '%s' to int for stadium %s: %v", apiStadium.CreatedYear, apiStadium.Name, err)
			}

			// แทรกหรืออัปเดตข้อมูลสนามใน DB
			err = database.InsertOrUpdateStadium(db, stadiumDB)
			if err != nil {
				log.Printf("Error saving stadium %s to DB: %v", apiStadium.Name, err)
			}
		}
	}
	return nil
}
