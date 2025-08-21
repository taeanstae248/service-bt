package handlers


import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
)

// UploadChannelLogo handles channel logo upload
func UploadChannelLogo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	channelID := vars["id"]

	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, `{"success": false, "error": "Failed to parse multipart form"}`, http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("logo")
	if err != nil {
		http.Error(w, `{"success": false, "error": "Failed to get file from form"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file type (optional)
	ext := filepath.Ext(handler.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
		http.Error(w, `{"success": false, "error": "Invalid file type. Only JPG, PNG, and GIF are allowed"}`, http.StatusBadRequest)
		return
	}

	uploadDir := "./img/channels"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("Error creating upload directory: %v", err)
		http.Error(w, `{"success": false, "error": "Failed to create upload directory"}`, http.StatusInternalServerError)
		return
	}

	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s-channel-%s%s", timestamp, channelID, ext)
	filepath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(filepath)
	if err != nil {
		log.Printf("Error creating file: %v", err)
		http.Error(w, `{"success": false, "error": "Failed to create file"}`, http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		log.Printf("Error copying file: %v", err)
		http.Error(w, `{"success": false, "error": "Failed to save file"}`, http.StatusInternalServerError)
		return
	}

	logoURL := fmt.Sprintf("/img/channels/%s", filename)
	_, err = DB.Exec("UPDATE channels SET logo_url = ? WHERE id = ?", logoURL, channelID)
	if err != nil {
		log.Printf("Error updating channel logo: %v", err)
		http.Error(w, `{"success": false, "error": "Failed to update channel logo"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"success":  "true",
		"logo_url": logoURL,
	}
	json.NewEncoder(w).Encode(response)
}
