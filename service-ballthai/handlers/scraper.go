package handlers

import (
	"go-ballthai-scraper/database"
	"go-ballthai-scraper/scraper"
	"log"
	"net/http"
)

func ScrapeMatchesHandler(w http.ResponseWriter, r *http.Request) {
	db := database.DB
	if db == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	err := scraper.ScrapeThaileagueMatches(db, "all")
	if err != nil {
		log.Println("Scrape error:", err)
		http.Error(w, "Scrape error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Scrape completed successfully"))
}
