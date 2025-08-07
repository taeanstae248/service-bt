package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"

	"go-ballthai-scraper/config"
	"go-ballthai-scraper/handlers"
	"go-ballthai-scraper/middleware"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	var err error
	db, err := sql.Open("mysql", cfg.GetDSN())
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Connected to database successfully!")

	// Set database connection for handlers
	handlers.SetDB(db)

	// Create router
	router := mux.NewRouter()

	// Apply middleware
	router.Use(middleware.Logging)
	router.Use(middleware.CORS)

	// Authentication routes
	router.HandleFunc("/api/auth/login", handlers.LoginHandler).Methods("POST")
	router.HandleFunc("/api/auth/logout", handlers.LogoutHandler).Methods("POST")
	router.HandleFunc("/api/auth/verify", handlers.VerifyHandler).Methods("GET")

	// API routes
	router.HandleFunc("/api/leagues", handlers.GetLeagues).Methods("GET")
	router.HandleFunc("/api/teams", handlers.GetTeams).Methods("GET")
	router.HandleFunc("/api/teams/{id}", handlers.GetTeamByID).Methods("GET")
	router.HandleFunc("/api/stadiums", handlers.GetStadiums).Methods("GET")
	router.HandleFunc("/api/matches", handlers.GetMatches).Methods("GET")

	// Player routes
	router.HandleFunc("/api/players", handlers.GetPlayers).Methods("GET")
	router.HandleFunc("/api/players/team/{team_id}", handlers.GetPlayersByTeamID).Methods("GET")
	router.HandleFunc("/api/players/team-post/{team_post_id}", handlers.GetPlayersByTeamPost).Methods("GET")

	// Static files
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// Serve HTML templates
	router.HandleFunc("/login.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./templates/login.html")
	})

	router.HandleFunc("/dashboard.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./templates/dashboard.html")
	})

	// Redirect root to login
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login.html", http.StatusFound)
	})

	// Start server
	port := ":" + cfg.ServerPort
	log.Printf("Server starting on port %s", cfg.ServerPort)
	log.Printf("Access the application at: http://localhost%s", port)
	log.Fatal(http.ListenAndServe(port, router))
}
