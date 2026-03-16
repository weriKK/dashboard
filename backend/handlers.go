package backend

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

// topRatedDashboardLimit controls how many globally ranked candidates are sent
// to the frontend for TOP badge mapping.
//
// Why this is intentionally larger than the visible TOP badge count:
//   - The frontend only highlights a small number of visible items (currently 3).
//   - Many highest-ranked items can be "hidden" from view because each feed column
//     is truncated by user-selected item count.
//   - Sending a larger candidate pool lets the frontend skip hidden links and still
//     find enough visible matches to fill the badge quota.
const topRatedDashboardLimit = 25

// CORSMiddleware adds CORS headers to all responses
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// HandleDashboard returns the complete dashboard data
func HandleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	feeds := GetAllFeeds()
	topRated := GetTopRatedItems(topRatedDashboardLimit)

	response := APIResponse{
		Feeds:    feeds,
		TopRated: topRated,
		Version:  AppVersion,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, max-age=60")
	json.NewEncoder(w).Encode(response)
}

// HandleClickFeedback records user click feedback for ML training
func HandleClickFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var feedback ClickFeedback
	if err := json.NewDecoder(r.Body).Decode(&feedback); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	feedback.Timestamp = time.Now()

	// Build a stable identity key: prefer link, fall back to title
	if feedback.ItemLink != "" {
		feedback.ItemKey = feedback.ItemLink
	} else {
		feedback.ItemKey = feedback.ItemTitle
	}

	if err := SaveClickEvent(feedback); err != nil {
		log.Printf("Failed to persist click feedback: %v", err)
	}

	log.Printf("Recorded click feedback: %s", feedback.ItemTitle)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "recorded"})
}

// HandleFrontend serves the frontend HTML
const frontendDir = "./frontend"

// HandleFrontend serves the SPA and static assets (HTML/CSS/JS)
func HandleFrontend(w http.ResponseWriter, r *http.Request) {
	// Serve the main page at root
	if r.URL.Path == "/" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, filepath.Join(frontendDir, "index.html"))
		return
	}

	if r.URL.Path == "/styles.css" {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		http.ServeFile(w, r, filepath.Join(frontendDir, "styles.css"))
		return
	}

	if r.URL.Path == "/scripts.js" {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		http.ServeFile(w, r, filepath.Join(frontendDir, "scripts.js"))
		return
	}

	http.NotFound(w, r)
}
