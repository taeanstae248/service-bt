package handlers

import (
	"go-ballthai-scraper/database"
	"go-ballthai-scraper/scraper"
	"log"
	"net/http"
	"fmt"
)

func ScrapeStandingsHandler(w http.ResponseWriter, r *http.Request) {
	db := database.DB
	if db == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}
	err := scraper.ScrapeStandings(db)
	if err != nil {
		log.Println("Scrape standings error:", err)
		http.Error(w, "Scrape standings error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Scrape standings completed successfully"))
}

func ScrapeMatchesHandler(w http.ResponseWriter, r *http.Request) {
	db := database.DB
	if db == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	// ปรับให้แสดงผลลัพธ์ลีกและลิงก์ที่ดึง
	var resultMsg string
	leagues, err := database.GetAllLeagues(db)
	if err != nil {
		log.Println("Scrape error:", err)
		http.Error(w, "Scrape error", http.StatusInternalServerError)
		return
	}

	baseURL := "https://competition.tl.prod.c0d1um.io/thaileague/api/match-day-match-public/?page=1&tournament="

	for _, league := range leagues {
		if league.ThaileageID != 0 {
			url := baseURL + fmt.Sprintf("%d", league.ThaileageID)
			resultMsg += fmt.Sprintf("%s\n%s\n\n", league.Name, url)
		}
	}

	err = scraper.ScrapeThaileagueMatches(db, "all")
	if err != nil {
		log.Println("Scrape error:", err)
		http.Error(w, "Scrape error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if resultMsg == "" {
		w.Write([]byte("Scrape completed successfully (no leagues found)"))
	} else {
		w.Write([]byte("Scrape completed successfully.\n\nLeagues scraped:\n" + resultMsg))
	}
}
