package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql" // Driver สำหรับ MySQL
	"github.com/joho/godotenv"         // สำหรับโหลด .env

	"go-ballthai-scraper/api"      // Import แพ็กเกจ api
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

	// Create connection string
	dsn := dbUser + ":" + dbPass + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName
	log.Printf("DEBUG: Connection String: %s", dsn)

	// 2. Initialize database connection
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}
	defer db.Close()

	// Test the database connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	log.Println("Database connection successful!")

	// 3. Initialize the database tables (if needed)
	database.InitializeDB(db)
	log.Println("Successfully connected to the database!")

	// Ensure image directories exist
	database.EnsureImageDirectories()
	log.Println("Image directories ensured.")

	// 4. Check for command line arguments
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "scrape":
			log.Println("Starting data scraping process...")
			scraper.StartScraping(db)
			log.Println("Data scraping completed.")
		case "api":
			log.Println("Starting API server mode...")
			database.EnsureImageDirectories()
			log.Println("Starting data scraping process...")
			go func() {
				scraper.StartScraping(db)
				log.Println("Background scraping completed.")
			}()
			startAPIServer()
		default:
			log.Printf("Unknown command: %s", os.Args[1])
			log.Println("Available commands: scrape, api")
		}
	} else {
		// Default behavior: run both scraping and API
		log.Println("No command specified. Starting data scraping process...")
		scraper.StartScraping(db)
		log.Println("Data scraping completed.")

		// Start API server
		log.Println("Starting API server...")
		startAPIServer()
	}
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
}

// startAPIServer starts the REST API server
func startAPIServer() {
	// Create handler with database connection
	handler := api.NewHandler(db)
	mux := handler.SetupRoutes()

	// Get port from environment or use default
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Starting API server on port %s...", port)
	log.Printf("📊 API Endpoints available:")
	log.Printf("   • GET /api/leagues        - All leagues")
	log.Printf("   • GET /api/teams          - All teams")
	log.Printf("   • GET /api/teams/{id}     - Specific team")
	log.Printf("   • GET /api/stadiums       - All stadiums")
	log.Printf("   • GET /api/matches        - Matches (upcoming and past)")
	log.Printf("   • GET /api/matches?league_id=1 - Matches by league ID")
	log.Printf("   • GET /api/matches?league=t1 - Matches by league name (t1, t2, t3, fa, lc, youth, cl, afc)")
	log.Printf("   • GET /images/{path}      - Static images")
	log.Printf("")
	log.Printf("🌐 API Server running at: http://localhost:%s", port)

	// Start server
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
