// ...existing code...
package main

import (
	"database/sql"
	"go-ballthai-scraper/database"
	"go-ballthai-scraper/handlers"
	"go-ballthai-scraper/scraper"
	"log"
	"net/http"
	"os"
	"github.com/robfig/cron/v3"

	_ "github.com/go-sql-driver/mysql" // Driver สำหรับ MySQL
	"github.com/gorilla/mux"
	"github.com/joho/godotenv" // สำหรับโหลด .env
)

var db *sql.DB // ตัวแปร Global สำหรับเก็บ Connection ฐานข้อมูล

func main() {
   // --- Cronjob: ดึง /scraper/matches และ standings โดยใช้ BASE_URL จาก env ---
   baseURL := os.Getenv("BASE_URL")
   if baseURL == "" {
	   baseURL = "http://localhost:8080"
   }
   log.Printf("[CRON] Using BASE_URL for cronjob: %s", baseURL)

   c := cron.New()
   matchesTimes := []string{
	   "55 17 * * *", // 17:55
	   "55 18 * * *", // 18:55
	   "55 19 * * *", // 19:55
	   "55 20 * * *", // 20:55
	   "55 21 * * *", // 21:55
	   "10 8 * * *",  // 08:10
   }
   for _, expr := range matchesTimes {
	   c.AddFunc(expr, func() {
		   url := baseURL + "/scraper/matches"
		   resp, err := http.Get(url)
		   if err != nil {
			   log.Println("cron fetch error /scraper/matches:", err)
			   return
		   }
		   defer resp.Body.Close()
		   log.Println("cron fetch /scraper/matches status:", resp.Status)
	   })
   }

   standingsTimes := []string{
	   "45 17 * * *", // 17:45
	   "45 18 * * *", // 18:45
	   "45 19 * * *", // 19:45
	   "45 20 * * *", // 20:45
	   "45 21 * * *", // 21:45
	   "20 8 * * *",  // 08:20
   }
   for _, expr := range standingsTimes {
	   c.AddFunc(expr, func() {
		   url := baseURL + "/scraper/standings"
		   resp, err := http.Get(url)
		   if err != nil {
			   log.Println("cron fetch error /scraper/standings:", err)
			   return
		   }
		   defer resp.Body.Close()
		   log.Println("cron fetch /scraper/standings status:", resp.Status)
	   })
   }
   c.Start()
// ...existing code...
	// ...existing code...
// ...existing code...
       r := mux.NewRouter()

	// Debug: เพิ่ม route debug เพื่อตรวจสอบ server
	r.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("debug ok"))
	})

	// Scraper: เพิ่ม route สำหรับ trigger J-League Scraper (GET/POST)
	r.HandleFunc("/scraper/jleague", handlers.ScrapeJLeagueHandler).Methods("GET", "POST")

	// Scraper: เพิ่ม route สำหรับ trigger UpdateTeamPostBallthai (GET)
	r.HandleFunc("/scraper/team-post-ballthai", func(w http.ResponseWriter, r *http.Request) {
		err := scraper.UpdateTeamPostBallthai(db)
		if err != nil {
			log.Printf("[ERROR] UpdateTeamPostBallthai: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte("Update team_post_ballthai completed!"))
	}).Methods("GET")

	// เพิ่ม route สำหรับอัปโหลดโลโก้ช่อง (channels)
	r.HandleFunc("/api/channels/{id}/upload-logo", handlers.UploadChannelLogo).Methods("POST")
	// Debug: log routes ทั้งหมด
	log.Println("[DEBUG] Registered routes:")
	r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		tmpl, _ := route.GetPathTemplate()
		methods, _ := route.GetMethods()
		log.Printf("[DEBUG] Route: %s %v", tmpl, methods)
		return nil
	})
	log.Println("BallThai service starting...") // เพิ่ม log ตรงนี้

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

	// Set global DB for handlers
	database.DB = db

	// Test the database connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	log.Println("Database connection successful!")
	log.Println("Successfully connected to the database!")

	// Register match routes (ลำดับสำคัญ: /api/matches/{id:[0-9]+} ต้องมาก่อน /api/matches)
	r.HandleFunc("/api/matches/{id:[0-9]+}", handlers.MatchGetByIDHandler(db)).Methods("GET")
	r.HandleFunc("/api/matches/{id:[0-9]+}", handlers.MatchUpdateHandler(db)).Methods("PUT")
	r.HandleFunc("/api/matches/{id:[0-9]+}", handlers.MatchDeleteHandler(db)).Methods("DELETE")
	r.HandleFunc("/api/matches", handlers.MatchListHandler(db)).Methods("GET")
	r.HandleFunc("/api/matches", handlers.MatchCreateHandler(db)).Methods("POST")

	// Register standing update route
	r.HandleFunc("/api/standings", handlers.CreateStanding).Methods("POST")
	r.HandleFunc("/api/standings/{id:[0-9]+}", handlers.UpdateStanding).Methods("PUT")

	// เพิ่ม NotFoundHandler เพื่อ debug
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"success":false,"error":"route not found","path":"` + req.URL.Path + `"}`))
	})

	log.Println("Listening on :8080")
	http.ListenAndServe(":8080", r)
}
