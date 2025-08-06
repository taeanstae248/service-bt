package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql" // Driver สำหรับ MySQL
	"github.com/joho/godotenv"         // สำหรับโหลด .env

	"go-ballthai-scraper/database" // Import แพ็กเกจ database
	"go-ballthai-scraper/scraper"  // Import แพ็กเกจ scraper
)

var db *sql.DB // ตัวแปร Global สำหรับเก็บ Connection ฐานข้อมูล

func main() {
	// 0. Check if .env file exists and is readable
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		log.Fatalf("Error: .env file not found in the current directory (%s). Please create it.", os.Getenv("PWD"))
	} else if err != nil {
		log.Fatalf("Error checking .env file: %v", err)
	}

	// Load values from .env file
	log.Println("Attempting to load .env file...")
	err := godotenv.Load()
	if err != nil {
		// If godotenv.Load() returns an Error, Fatalf immediately to see the cause
		log.Fatalf("Critical Error loading .env file: %v. Please ensure the .env file is correctly formatted and accessible.", err)
	} else {
		log.Println(".env file loaded successfully.")
	}

	// 1. Configure database connection
	// Retrieve values from Environment Variables (should now be from .env)
	dbUser := os.Getenv("DB_USERNAME")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	// Add Logs to verify retrieved values
	log.Printf("DEBUG: DB_USERNAME: '%s'", dbUser)
	log.Printf("DEBUG: DB_PASSWORD: '%s' (length: %d)", dbPass, len(dbPass))
	log.Printf("DEBUG: DB_HOST: '%s'", dbHost)
	log.Printf("DEBUG: DB_PORT: '%s'", dbPort)
	log.Printf("DEBUG: DB_NAME: '%s'", dbName)

	// Check if essential values are set
	if dbUser == "" || dbHost == "" || dbPort == "" || dbName == "" {
		log.Fatalf("Missing one or more essential database environment variables (DB_USERNAME, DB_HOST, DB_PORT, DB_NAME)")
	}

	connStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)
	log.Printf("DEBUG: Connection String: %s", connStr)

	// Call InitDB from the database package
	err = database.InitDB(connStr)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	db = database.DB
	defer db.Close()

	log.Println("Successfully connected to the database!")

	// 2. Create image folders (if they don't exist)
	imageDirs := []string{
		"./img/coach",
		"./img/player",
		"./img/source",
		"./img/stadiums",
	}
	for _, dir := range imageDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				log.Fatalf("Failed to create image directory %s: %v", dir, err)
			}
		}
	}
	log.Println("Image directories ensured.")

	// 3. Start data scraping process
	log.Println("Starting data scraping process...")

	// // Scrape Stadium data
	// log.Println("Scraping Stadiums...")
	// err = scraper.ScrapeStadiums(db)
	// if err != nil {
	// 	log.Printf("Error scraping stadiums: %v", err)
	// } else {
	// 	log.Println("Stadiums scraping completed.")
	// }

	// // Scrape Coach data
	// log.Println("Scraping Coaches...")
	// err = scraper.ScrapeCoach(db)
	// if err != nil {
	// 	log.Printf("Error scraping coaches: %v", err)
	// } else {
	// 	log.Println("Coaches scraping completed.")
	// }

	// Scrape Player data
	log.Println("Scraping Players...")
	err = scraper.ScrapePlayers(db)
	if err != nil {
		log.Printf("Error scraping players: %v", err)
	} else {
		log.Println("Players scraping completed.")
	}

	// // Scrape Standing data
	// log.Println("Scraping Standings...")
	// err = scraper.ScrapeStandings(db)
	// if err != nil {
	// 	log.Printf("Error scraping standings: %v", err)
	// } else {
	// 	log.Println("Standings scraping completed.")
	// }

	// // Scrape Match data
	// log.Println("Scraping Matches (Thaileague, Cup, Playoff)...")
	// err = scraper.ScrapeThaileagueMatches(db, "all") // Pass "all" to scrape all Thai Leagues
	// if err != nil {
	// 	log.Printf("Error scraping Thaileague matches: %v", err)
	// }
	// err = scraper.ScrapeBallthaiCupMatches(db)
	// if err != nil {
	// 	log.Printf("Error scraping Ballthai Cup matches: %v", err)
	// }
	// err = scraper.ScrapeThaileaguePlayoffMatches(db)
	// if err != nil {
	// 	log.Printf("Error scraping Thaileague Playoff matches: %v", err)
	// }
	// log.Println("Match scraping initiated. (Check logs for success/failure)")

	log.Println("Data scraping process finished.")
}
