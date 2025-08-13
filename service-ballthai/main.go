package main

import (
	"database/sql"
	"go-ballthai-scraper/handlers"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql" // Driver สำหรับ MySQL
	"github.com/gorilla/mux"
	"github.com/joho/godotenv" // สำหรับโหลด .env
)

var db *sql.DB // ตัวแปร Global สำหรับเก็บ Connection ฐานข้อมูล

func main() {
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

	// Test the database connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	log.Println("Database connection successful!")
	log.Println("Successfully connected to the database!")

	r := mux.NewRouter()

	// Register match routes (ลำดับสำคัญ: /api/matches/{id:[0-9]+} ต้องมาก่อน /api/matches)
	r.HandleFunc("/api/matches/{id:[0-9]+}", handlers.MatchGetByIDHandler(db)).Methods("GET")
	r.HandleFunc("/api/matches/{id:[0-9]+}", handlers.MatchUpdateHandler(db)).Methods("PUT")
	r.HandleFunc("/api/matches/{id:[0-9]+}", handlers.MatchDeleteHandler(db)).Methods("DELETE")
	r.HandleFunc("/api/matches", handlers.MatchListHandler(db)).Methods("GET")
	r.HandleFunc("/api/matches", handlers.MatchCreateHandler(db)).Methods("POST")

	// เพิ่ม NotFoundHandler เพื่อ debug
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"success":false,"error":"route not found","path":"` + req.URL.Path + `"}`))
	})

	log.Println("Listening on :8080")
	http.ListenAndServe(":8080", r)
}
