package main

import (
	"database/sql"
	"log"
	"net/http"
	"html/template"
	"strings"

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
router.HandleFunc("/api/players/{id:[0-9]+}", handlers.UpdatePlayer).Methods("PUT")
	router.HandleFunc("/players.html", func(w http.ResponseWriter, r *http.Request) {
		db := database.DB
		if db == nil {
			http.Error(w, "Database not initialized", http.StatusInternalServerError)
			return
		}
		// Filter by team_id and name if present
		teamID := r.URL.Query().Get("team_id")
		name := r.URL.Query().Get("name")
		baseQuery := `SELECT p.id, p.name, p.full_name_en, p.shirt_number, p.position, p.photo_url, p.matches_played, p.goals, p.yellow_cards, p.red_cards, p.status, t.name_th as team_name FROM players p LEFT JOIN teams t ON p.team_id = t.id`
		var where []string
		var args []interface{}
		if teamID != "" {
			where = append(where, "p.team_id = ?")
			args = append(args, teamID)
		}
		if name != "" {
			where = append(where, "p.name LIKE ?")
			args = append(args, "%"+name+"%")
		}
		query := baseQuery
		if len(where) > 0 {
			query += " WHERE " + strings.Join(where, " AND ")
		}
		rows, err := db.Query(query, args...)
		if err != nil {
			log.Printf("Query error in /players.html: %v", err)
			http.Error(w, "Query error: "+err.Error(), 500)
			return
		}
		defer rows.Close()
		type Player struct {
			ID            int
			Name          string
			FullNameEN    string
			ShirtNumber   int
			Position      string
			PhotoURL      string
			MatchesPlayed int
			Goals         int
			YellowCards   int
			RedCards      int
			Status        int
			TeamName      string
		}
		var players []Player
		for rows.Next() {
			var p Player
			var shirtNumber sql.NullInt64
			var position sql.NullString
			var photoURL sql.NullString
			var teamName sql.NullString
			if err := rows.Scan(&p.ID, &p.Name, &p.FullNameEN, &shirtNumber, &position, &photoURL, &p.MatchesPlayed, &p.Goals, &p.YellowCards, &p.RedCards, &p.Status, &teamName); err != nil {
				continue
			}
			p.ShirtNumber = int(shirtNumber.Int64)
			p.Position = position.String
			p.PhotoURL = photoURL.String
			p.TeamName = teamName.String
			players = append(players, p)
		}
		tmpl, err := template.ParseFiles("templates/players.html", "templates/_nav.html")
		if err != nil {
			http.Error(w, "Template error", 500)
			return
		}
		data := struct{ Players []Player }{Players: players}
		tmpl.Execute(w, data)
	})

	router.HandleFunc("/dashboard.html", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/dashboard.html", "templates/_nav.html")
		if err != nil {
			http.Error(w, "Template error", 500)
			return
		}
		tmpl.Execute(w, nil)
	})

	router.HandleFunc("/teams.html", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/teams.html", "templates/_nav.html")
		if err != nil {
			http.Error(w, "Template error", 500)
			return
		}
		tmpl.Execute(w, nil)
	})

	router.HandleFunc("/leagues.html", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/leagues.html", "templates/_nav.html")
		if err != nil {
			http.Error(w, "Template error", 500)
			return
		}
		tmpl.Execute(w, nil)
	})

	router.HandleFunc("/matches.html", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/matches.html", "templates/_nav.html")
		if err != nil {
			http.Error(w, "Template error", 500)
			return
		}
		tmpl.Execute(w, nil)
	})

	router.HandleFunc("/standings.html", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/standings.html", "templates/_nav.html")
		if err != nil {
			http.Error(w, "Template error", 500)
			return
		}
		tmpl.Execute(w, nil)
	})


	// Serve static files (css, js, images)
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

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
