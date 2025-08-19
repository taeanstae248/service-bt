package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"go-ballthai-scraper/database"

	"golang.org/x/crypto/bcrypt"
)

// Request/Response structures
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	User      database.User `json:"user"`
	SessionID string        `json:"session_id"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Generate a secure session ID
func generateSessionID() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// LoginHandler handles user login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var loginReq LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, `{"success": false, "error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate input
	if loginReq.Username == "" || loginReq.Password == "" {
		http.Error(w, `{"success": false, "error": "Username and password are required"}`, http.StatusBadRequest)
		return
	}

	// Get user password hash
	passwordHash, err := database.GetUserPasswordHash(loginReq.Username)
	if err == sql.ErrNoRows {
		log.Printf("User not found: %s", loginReq.Username)
		http.Error(w, `{"success": false, "error": "Invalid username or password"}`, http.StatusUnauthorized)
		return
	} else if err != nil {
		log.Printf("Database error getting password hash: %v", err)
		http.Error(w, `{"success": false, "error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(loginReq.Password)); err != nil {
		log.Printf("Password verification failed for user: %s", loginReq.Username)
		http.Error(w, `{"success": false, "error": "Invalid username or password"}`, http.StatusUnauthorized)
		return
	}

	// Get user details
	user, err := database.GetUserByUsername(loginReq.Username)
	if err != nil {
		log.Printf("Failed to get user details: %v", err)
		http.Error(w, `{"success": false, "error": "Failed to get user details"}`, http.StatusInternalServerError)
		return
	}

	// Generate session ID
	sessionID, err := generateSessionID()
	if err != nil {
		http.Error(w, `{"success": false, "error": "Failed to generate session"}`, http.StatusInternalServerError)
		return
	}

	// Create session (expires in 24 hours)
	expiresAt := time.Now().Add(24 * time.Hour)
	if err := database.CreateSession(sessionID, user.ID, expiresAt); err != nil {
		http.Error(w, `{"success": false, "error": "Failed to create session"}`, http.StatusInternalServerError)
		return
	}

	// Update last login
	database.UpdateLastLogin(user.ID)

	// Set session_id cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		// Secure: true, // เปิดถ้าใช้ HTTPS
		Expires:  expiresAt,
	})

	// Return response
	response := APIResponse{
		Success: true,
		Data: LoginResponse{
			User:      *user,
			SessionID: sessionID,
		},
	}

	json.NewEncoder(w).Encode(response)
}

// LogoutHandler handles user logout
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get session ID from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, `{"success": false, "error": "Authorization header required"}`, http.StatusUnauthorized)
		return
	}

	// Extract session ID (Bearer <session_id>)
	sessionID := strings.TrimPrefix(authHeader, "Bearer ")
	if sessionID == authHeader {
		http.Error(w, `{"success": false, "error": "Invalid authorization format"}`, http.StatusUnauthorized)
		return
	}

	// Delete session
	if err := database.DeleteSession(sessionID); err != nil {
		log.Printf("Failed to delete session: %v", err)
		http.Error(w, `{"success": false, "error": "Failed to logout"}`, http.StatusInternalServerError)
		return
	}

	response := APIResponse{
		Success: true,
		Data:    "Logged out successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// VerifyHandler verifies user session
func VerifyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get session ID from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, `{"success": false, "error": "Authorization header required"}`, http.StatusUnauthorized)
		return
	}

	// Extract session ID (Bearer <session_id>)
	sessionID := strings.TrimPrefix(authHeader, "Bearer ")
	if sessionID == authHeader {
		http.Error(w, `{"success": false, "error": "Invalid authorization format"}`, http.StatusUnauthorized)
		return
	}

	// Verify session
	session, err := database.GetSession(sessionID)
	if err == sql.ErrNoRows {
		http.Error(w, `{"success": false, "error": "Invalid session"}`, http.StatusUnauthorized)
		return
	} else if err != nil {
		log.Printf("Database error verifying session: %v", err)
		http.Error(w, `{"success": false, "error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	// Check if session is expired
	now := time.Now().Local()
	if now.After(session.ExpiresAt) {
		database.DeleteSession(sessionID) // Clean up expired session
		http.Error(w, `{"success": false, "error": "Session expired"}`, http.StatusUnauthorized)
		return
	}

	// Get user details
	user, err := database.GetUserByID(session.UserID)
	if err != nil {
		log.Printf("Failed to get user details: %v", err)
		http.Error(w, `{"success": false, "error": "Failed to get user details"}`, http.StatusInternalServerError)
		return
	}

	response := APIResponse{
		Success: true,
		Data:    user,
	}

	json.NewEncoder(w).Encode(response)
}
