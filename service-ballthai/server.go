
package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"html/template"
	"strings"
	"os"
	"strconv"
	"github.com/robfig/cron/v3"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"

	"go-ballthai-scraper/config"
	"go-ballthai-scraper/database"
	"go-ballthai-scraper/handlers"
	"go-ballthai-scraper/middleware"
	"go-ballthai-scraper/scraper"
)


	// scrapePostHandler ดึงข้อมูลจาก serviceseoball.com แล้วส่งต่อให้ client
	func scrapePostHandler(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("This is /scrape/post endpoint (no external fetch)."))
	}

func main() {
   // --- Cronjob: ดึง /scraper/matches ทุก 30 นาที ---
   c := cron.New()
   // ดึงทุกชั่วโมง เฉพาะ 7, 15, 16, 17, 18, 19, 20, 21 น.
   c.AddFunc("0 7,15-21 * * *", func() {
	   resp, err := http.Get("https://svc.ballthai.com/scraper/matches")
	   if err != nil {
		   log.Println("cron fetch error /scraper/matches:", err)
		   return
	   }
	   defer resp.Body.Close()
	   log.Println("cron fetch /scraper/matches status:", resp.Status)
   })
	// Also fetch standings on the same schedule but every 10 minutes from the base hour
	c.AddFunc("10 7,15-21 * * *", func() {
	   resp, err := http.Get("https://svc.ballthai.com/scraper/standing")
	   if err != nil {
		   log.Println("cron fetch error /scraper/standing:", err)
		   return
	   }
	   defer resp.Body.Close()
	   log.Println("cron fetch /scraper/standing status:", resp.Status)
   })
   c.Start()
	// ประกาศตัวแปร db และ err
	var db *sql.DB
	var err error

	// Create router
	router := mux.NewRouter()

	// เพิ่ม route สำหรับอัปเดต team_post_ballthai
	router.HandleFunc("/scraper/team-post-ballthai", func(w http.ResponseWriter, r *http.Request) {
		db := database.DB
		if db == nil {
			http.Error(w, "Database not initialized", http.StatusInternalServerError)
			return
		}
		err := scraper.UpdateTeamPostBallthai(db)
		if err != nil {
			log.Printf("[ERROR] UpdateTeamPostBallthai: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte("Update team_post_ballthai completed!"))
	}).Methods("GET")

	// Scrape post proxy
	router.HandleFunc("/scrape/post", scrapePostHandler).Methods("GET")
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	db, err = sql.Open("mysql", cfg.GetDSN())
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
	router.HandleFunc("/api/standings/order", handlers.UpdateStandingsOrder).Methods("POST")
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
	router.HandleFunc("/scraper/seasons", handlers.ScrapeSeasonsHandler).Methods("GET")

	// Player routes
	router.HandleFunc("/api/players", handlers.GetPlayers).Methods("GET")
	router.HandleFunc("/api/players/top-scorers", handlers.GetTopScorers).Methods("GET")
	router.HandleFunc("/api/players/team/{team_id}", handlers.GetPlayersByTeamID).Methods("GET")
	router.HandleFunc("/api/players/team-post/{team_post_id}", handlers.GetPlayersByTeamPost).Methods("GET")
	router.HandleFunc("/api/players/{id:[0-9]+}", handlers.UpdatePlayer).Methods("PUT")
	router.Handle("/players.html", middleware.CheckAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db := database.DB
		if db == nil {
			http.Error(w, "Database not initialized", http.StatusInternalServerError)
			return
		}
		// Filter by team_id and name if present
		teamID := r.URL.Query().Get("team_id")
		name := r.URL.Query().Get("name")
		pageStr := r.URL.Query().Get("page")
		pageSizeStr := r.URL.Query().Get("page_size")
		page := 1
		pageSize := 50
		if pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}
		if pageSizeStr != "" {
			if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 500 {
				pageSize = ps
			}
		}

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

		// Count total
		countQuery := "SELECT COUNT(*) FROM players p LEFT JOIN teams t ON p.team_id = t.id"
		if len(where) > 0 {
			countQuery += " WHERE " + strings.Join(where, " AND ")
		}
		var totalCount int
		if err := db.QueryRow(countQuery, args...).Scan(&totalCount); err != nil {
			log.Printf("Count query error in /players.html: %v", err)
			http.Error(w, "Query error: "+err.Error(), 500)
			return
		}

		query := baseQuery
		if len(where) > 0 {
			query += " WHERE " + strings.Join(where, " AND ")
		}

		// pagination
		offset := (page - 1) * pageSize
		query += " ORDER BY p.shirt_number, p.name LIMIT ? OFFSET ?"
		args = append(args, pageSize, offset)

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

		rows, err := db.Query(query, args...)
		if err != nil {
			log.Printf("Query error in /players.html: %v", err)
			http.Error(w, "Query error: "+err.Error(), 500)
			return
		}
		defer rows.Close()

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

		// prepare paging metadata
		totalPages := 1
		if totalCount > 0 {
			totalPages = (totalCount + pageSize - 1) / pageSize
		}
		hasPrev := page > 1
		hasNext := page < totalPages
		prevPage := page - 1
		nextPage := page + 1

		tmpl, err := template.ParseFiles("templates/players.html", "templates/_nav.html")
		if err != nil {
			http.Error(w, "Template error", 500)
			return
		}
		data := struct{
			Players []Player
			CurrentPage int
			PageSize int
			TotalPages int
			TotalCount int
			HasPrev bool
			HasNext bool
			PrevPage int
			NextPage int
			TeamID string
			Name string
		}{
			Players: players,
			CurrentPage: page,
			PageSize: pageSize,
			TotalPages: totalPages,
			TotalCount: totalCount,
			HasPrev: hasPrev,
			HasNext: hasNext,
			PrevPage: prevPage,
			NextPage: nextPage,
			TeamID: teamID,
			Name: name,
		}
		tmpl.Execute(w, data)
	})))


	// กลุ่ม route ที่ต้อง login
	router.Handle("/dashboard.html", middleware.CheckAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/dashboard.html", "templates/_nav.html")
		if err != nil {
			http.Error(w, "Template error", 500)
			return
		}
		tmpl.Execute(w, nil)
	})))

	router.Handle("/teams.html", middleware.CheckAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/teams.html", "templates/_nav.html")
		if err != nil {
			http.Error(w, "Template error", 500)
			return
		}
		tmpl.Execute(w, nil)
	})))

	router.Handle("/leagues.html", middleware.CheckAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/leagues.html", "templates/_nav.html")
		if err != nil {
			http.Error(w, "Template error", 500)
			return
		}
		tmpl.Execute(w, nil)
	})))

	router.Handle("/matches.html", middleware.CheckAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/matches.html", "templates/_nav.html")
		if err != nil {
			http.Error(w, "Template error", 500)
			return
		}
		tmpl.Execute(w, nil)
	})))

	router.Handle("/standings.html", middleware.CheckAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/standings.html", "templates/_nav.html")
		if err != nil {
			http.Error(w, "Template error", 500)
			return
		}
		tmpl.Execute(w, nil)
	})))



	// เพิ่ม route สำหรับหน้า login.html
	router.HandleFunc("/login.html", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/login.html", "templates/_nav.html")
		if err != nil {
			http.Error(w, "Template error", 500)
			return
		}
		tmpl.Execute(w, nil)
	})

	// Serve static files (css, js, images)
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	// Serve team logos from /img/teams/ -> ./img/teams/
	router.PathPrefix("/img/teams/").Handler(http.StripPrefix("/img/teams/", http.FileServer(http.Dir("img/teams/"))))
	// Serve channel images from /img/channels/ -> ./img/channels/
	router.PathPrefix("/img/channels/").Handler(http.StripPrefix("/img/channels/", http.FileServer(http.Dir("img/channels/"))))
	// Serve player images from /img/player/ -> ./img/player/
	router.PathPrefix("/img/player/").Handler(http.StripPrefix("/img/player/", http.FileServer(http.Dir("img/player/"))))

	// Ensure image directories exist to avoid 404s when files are created at runtime
	os.MkdirAll("img/teams", 0755)
	os.MkdirAll("img/channels", 0755)
	os.MkdirAll("img/player", 0755)

	// Redirect root to login
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login.html", http.StatusFound)
	})

	// เพิ่ม route สำหรับ scrape teams by leagueid (บันทึกลง database ด้วย)
	router.HandleFunc("/scrape/teams/{leagueid}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		leagueID := vars["leagueid"]
		db := database.DB
		if db == nil {
			http.Error(w, "Database not initialized", 500)
			return
		}
		   imported, err := scraper.SaveTeamsAndLogosByLeagueID(db, leagueID)
		   if err != nil {
			   http.Error(w, err.Error(), 500)
			   return
		   }
		   w.Header().Set("Content-Type", "application/json")
		   json.NewEncoder(w).Encode(map[string]interface{}{
			   "success": true,
			   "imported": imported,
			   "message": "นำเข้าทีมและโลโก้สำเร็จ (logo_url เป็น path local เท่านั้น)",
		   })
	}).Methods("GET")

	// อ่าน host/port จาก environment
	host := os.Getenv("API_HOST")
	if host == "" {
		host = "0.0.0.0"
	}
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}
	addr := host + ":" + port

	if port == "443" {
		log.Printf("Starting HTTPS server on %s", addr)
		err := http.ListenAndServeTLS(addr, "/etc/letsencrypt/live/svc.ballthai.com/fullchain.pem", "/etc/letsencrypt/live/svc.ballthai.com/privkey.pem", router)
		if err != nil {
			log.Fatal("ListenAndServeTLS error:", err)
		}
	} else {
		log.Printf("Starting HTTP server on %s", addr)
		log.Fatal(http.ListenAndServe(addr, router))
	}
}
