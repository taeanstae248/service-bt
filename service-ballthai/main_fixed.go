package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql" // Driver สำหรับ MySQL
	"github.com/joho/godotenv"         // สำหรับโหลด .env

	"go-ballthai-scraper/api" // Import แพ็กเกจ api
)

var db *sql.DB // ตัวแปร Global สำหรับเก็บ Connection ฐานข้อมูล

func main() {
	// Load values from .env file
	log.Println("Attempting to load .env file...")
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Critical Error loading .env file: %v", err)
	} else {
		log.Println(".env file loaded successfully.")
	}

	// Configure database connection
	dbUser := os.Getenv("DB_USERNAME")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	log.Printf("DEBUG: DB_USERNAME: '%s'", dbUser)
	log.Printf("DEBUG: DB_PASSWORD: '%s' (length: %d)", dbPass, len(dbPass))
	log.Printf("DEBUG: DB_HOST: '%s'", dbHost)
	log.Printf("DEBUG: DB_PORT: '%s'", dbPort)
	log.Printf("DEBUG: DB_NAME: '%s'", dbName)

	// Create connection string
	dsn := dbUser + ":" + dbPass + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName
	log.Printf("DEBUG: Connection String: %s", dsn)

	// Initialize database connection
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
	log.Println("Successfully connected to the database!")

	// Check for command line arguments
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "api":
			log.Println("Starting API server mode...")
			startAPIServer()
		default:
			log.Printf("Unknown command: %s", os.Args[1])
			log.Println("Available commands: api")
		}
	} else {
		// Default behavior: start API server
		log.Println("No command specified. Starting API server...")
		startAPIServer()
	}
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
	log.Printf("   • GET /api/players        - All players (with filtering: ?team_id=X, ?league_id=X, ?position=GK)")
	log.Printf("   • GET /api/matches        - Matches (upcoming and past)")
	log.Printf("   • GET /api/matches?league_id=1 - Matches by league ID")
	log.Printf("   • GET /api/matches?league=t1 - Matches by league name (t1, t2, t3, fa, lc, youth, cl, afc)")
	log.Printf("   • GET /api/standings      - League standings")
	log.Printf("   • GET /api/standings?league=t1 - Standings by league (t1, t2, t3, fa, lc)")
	log.Printf("   • POST /api/scrape/jleague-standings - Scrape J-League standings")
	log.Printf("   • GET /images/{path}      - Static images")
	log.Printf("")
	log.Printf("🌐 API Server running at: http://localhost:%s", port)

	// Start server
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
