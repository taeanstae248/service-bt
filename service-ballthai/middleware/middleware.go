package middleware

import (
	"log"
	"net/http"
	"time"
	"go-ballthai-scraper/database"
)

// Auth middleware: ตรวจสอบ session จาก Authorization header (Bearer <session_id>)
func CheckAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// อย่า redirect ซ้ำถ้า path คือ /login.html หรือ /api/auth/login หรือ /api/auth/verify
		if r.URL.Path == "/login.html" || r.URL.Path == "/api/auth/login" || r.URL.Path == "/api/auth/verify" {
			next.ServeHTTP(w, r)
			return
		}
		// 1. ลองอ่านจาก Authorization header
		sessionID := ""
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			sessionID = authHeader
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				sessionID = authHeader[7:]
			}
		} else {
			// 2. ถ้าไม่มี header ให้ลองอ่านจาก cookie ชื่อ session_id
			cookie, err := r.Cookie("session_id")
			if err == nil {
				sessionID = cookie.Value
			}
		}
		if sessionID == "" {
			http.Redirect(w, r, "/login.html", http.StatusFound)
			return
		}
		// ตรวจสอบ session ใน database
		session, err := database.GetSession(sessionID)
		if err != nil || session == nil || session.ExpiresAt.Before(time.Now()) {
			http.Redirect(w, r, "/login.html", http.StatusFound)
			return
		}
		// ผ่าน auth
		next.ServeHTTP(w, r)
	})
}

// CORS middleware
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Logging middleware
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}
