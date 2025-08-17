package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"

	"go-ballthai-scraper/config"
	"go-ballthai-scraper/database"
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
	database.SetDB(db)

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
	router.HandleFunc("/api/leagues/search", handlers.SearchLeagues).Methods("GET")
	router.HandleFunc("/api/leagues", handlers.CreateLeague).Methods("POST")
	router.HandleFunc("/api/leagues/{id}", handlers.UpdateLeague).Methods("PUT")
	router.HandleFunc("/api/leagues/{id}", handlers.DeleteLeague).Methods("DELETE")
	router.HandleFunc("/api/teams", handlers.GetTeams).Methods("GET")
	router.HandleFunc("/api/stages", handlers.GetStages).Methods("GET")
	router.HandleFunc("/api/teams/search", handlers.SearchTeams).Methods("GET")
	router.HandleFunc("/api/teams", handlers.CreateTeam).Methods("POST")
	// Standings API
	router.HandleFunc("/api/standings", handlers.GetStandings).Methods("GET")
	router.HandleFunc("/api/standings/{id:[0-9]+}", handlers.UpdateStanding).Methods("PUT")
	router.HandleFunc("/api/teams/{id}", handlers.GetTeamByID).Methods("GET")
	router.HandleFunc("/api/teams/{id}", handlers.UpdateTeam).Methods("PUT")
	router.HandleFunc("/api/teams/{id}", handlers.DeleteTeam).Methods("DELETE")
	router.HandleFunc("/api/teams/{id}/logo", handlers.UploadTeamLogo).Methods("POST")
	router.HandleFunc("/api/stadiums", handlers.GetStadiums).Methods("GET")
	router.HandleFunc("/api/matches", handlers.GetMatches).Methods("GET")
	router.HandleFunc("/api/matches", handlers.CreateMatch).Methods("POST")
	router.HandleFunc("/api/matches/{id}", handlers.GetMatchByID).Methods("GET")
	router.HandleFunc("/api/matches/{id}", handlers.DeleteMatch).Methods("DELETE")
	router.HandleFunc("/api/matches/{id}", handlers.UpdateMatch).Methods("PUT")
	router.HandleFunc("/api/channels", handlers.GetChannels).Methods("GET")
	// เพิ่ม route สำหรับ scraper
	router.HandleFunc("/scraper/matches", handlers.ScrapeMatchesHandler).Methods("GET")
	router.HandleFunc("/scraper/standing", handlers.ScrapeStandingsHandler).Methods("GET")
	router.HandleFunc("/scraper/jleague", handlers.ScrapeJLeagueHandler).Methods("GET")
	router.HandleFunc("/scraper/player", handlers.ScrapePlayersHandler).Methods("GET")

	// Player routes
	router.HandleFunc("/api/players", handlers.GetPlayers).Methods("GET")
	router.HandleFunc("/api/players/team/{team_id}", handlers.GetPlayersByTeamID).Methods("GET")
	router.HandleFunc("/api/players/team-post/{team_post_id}", handlers.GetPlayersByTeamPost).Methods("GET")

	// Static files
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	router.PathPrefix("/img/").Handler(http.StripPrefix("/img/", http.FileServer(http.Dir("./img/"))))

	// Serve HTML templates
	router.HandleFunc("/login.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./templates/login.html")
	})

	router.HandleFunc("/dashboard.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./templates/dashboard.html")
	})

	router.HandleFunc("/teams.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./templates/teams.html")
	})

	router.HandleFunc("/leagues.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./templates/leagues.html")
	})

	router.HandleFunc("/matches.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./templates/matches.html")
	})

	router.HandleFunc("/standings.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./templates/standings.html")
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
