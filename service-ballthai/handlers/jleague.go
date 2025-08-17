package handlers

import (
	"go-ballthai-scraper/database"
	"go-ballthai-scraper/scraper"
	"log"
	"net/http"
)

// ScrapeJLeagueHandler handles scraping J-League standings
func ScrapeJLeagueHandler(w http.ResponseWriter, r *http.Request) {
	db := database.DB
	if db == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}
	err := scraper.ScrapeJLeagueStandings(db)
	if err != nil {
		log.Println("Scrape J-League error:", err)
		http.Error(w, "Scrape J-League error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Scrape J-League completed successfully"))
}
