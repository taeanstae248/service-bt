package scraper

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"go-ballthai-scraper/database" // ตรวจสอบให้แน่ใจว่าชื่อโมดูลตรงกับ go.mod ของคุณ
	"go-ballthai-scraper/models"  // ตรวจสอบให้แน่ใจว่าชื่อโมดูลตรงกับ go.mod ของคุณ
)

// ScrapeCoach ดึงข้อมูลโค้ชจาก API และบันทึกลงฐานข้อมูล
func ScrapeCoach(db *sql.DB) error {
	baseURL := "https://competition.tl.prod.c0d1um.io/thaileague/api/staff-public/?type=headcoach&page="
	maxPages := 10 // ตามที่เห็นใน PHP ต้นฉบับ

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
			// ดาวน์โหลดรูปภาพโค้ช
			photoPath := ""
			if apiCoach.Photo != "" {
				downloadedPath, err := DownloadImage(apiCoach.Photo, "./img/coach")
				if err != nil {
					log.Printf("Warning: Failed to download coach photo for %s: %v", apiCoach.FullName, err)
				} else {
					photoPath = downloadedPath
				}
			}

			// รับ Nationality ID
			nationalityID := sql.NullInt64{Valid: false}
			if apiCoach.Nationality.Code != "" {
				nID, err := database.GetNationalityID(db, apiCoach.Nationality.Code, apiCoach.Nationality.Name)
				if err != nil {
					log.Printf("Warning: Failed to get nationality ID for %s: %v", apiCoach.Nationality.Name, err)
				} else {
					nationalityID = sql.NullInt64{Int64: int64(nID), Valid: true}
				}
			}

			// รับ Team ID
			teamID := sql.NullInt64{Valid: false}
			if apiCoach.ClubName != "" {
				tID, err := database.GetTeamIDByThaiName(db, apiCoach.ClubName, "") // สมมติว่าโลโก้ทีมไม่พร้อมใช้งานที่นี่
				if err != nil {
					log.Printf("Warning: Failed to get team ID for coach %s's club %s: %v", apiCoach.FullName, apiCoach.ClubName, err)
				} else {
					teamID = sql.NullInt64{Int64: int64(tID), Valid: true}
				}
			}

			// แปลง BirthDate
			birthDate := sql.NullTime{Valid: false}
			if apiCoach.BirthDate != "" {
				// สมมติว่ารูปแบบวันที่ API คือ "YYYY-MM-DD"
				parsedDate, err := time.Parse("2006-01-02", apiCoach.BirthDate)
				if err != nil {
					log.Printf("Warning: Failed to parse birth date %s for coach %s: %v", apiCoach.BirthDate, apiCoach.FullName, err)
				} else {
					birthDate = sql.NullTime{Time: parsedDate, Valid: true}
				}
			}

			// เตรียมโครงสร้าง CoachDB
			coachDB := models.CoachDB{
				CoachRefID:    sql.NullInt64{Int64: int64(apiCoach.ID), Valid: true},
				Name:          apiCoach.FullName,
				Birthday:      birthDate,
				TeamID:        teamID,
				NationalityID: nationalityID,
				PhotoURL:      sql.NullString{String: photoPath, Valid: photoPath != ""},
			}

			// แทรกหรืออัปเดตโค้ชใน DB
			err = database.InsertOrUpdateCoach(db, coachDB)
			if err != nil {
				log.Printf("Error saving coach %s to DB: %v", apiCoach.FullName, err)
			}
		}
	}
	return nil
}
