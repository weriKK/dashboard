package backend

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

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

// HandleGetFeeds returns all feeds
func HandleGetFeeds(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	feeds := GetAllFeeds()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, max-age=60")
	json.NewEncoder(w).Encode(feeds)
}

// HandleGetStocks returns cached stock data
func HandleGetStocks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	StockCacheMu.RLock()
	defer StockCacheMu.RUnlock()

	var stocks []StockData
	for _, stockConfig := range Cfg.Stocks.Items {
		if data, ok := StockCache[stockConfig.Symbol]; ok {
			stocks = append(stocks, *data)
		} else {
			stocks = append(stocks, StockData{
				Symbol:    stockConfig.Symbol,
				Label:     stockConfig.Label,
				UpdatedAt: time.Now(),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stocks)
}

// HandleGetRecommendations returns ML-ranked recommendations
func HandleGetRecommendations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	recommendations := GetRecommendations()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recommendations)
}

// HandleGetTimezones returns configured timezones
func HandleGetTimezones(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Cfg.Timezones)
}

// HandleDashboard returns the complete dashboard data
func HandleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	feeds := GetAllFeeds()

	StockCacheMu.RLock()
	var stocks []StockData
	for _, stockConfig := range Cfg.Stocks.Items {
		if data, ok := StockCache[stockConfig.Symbol]; ok {
			stocks = append(stocks, *data)
		} else {
			// Return empty stock data if not yet fetched
			stocks = append(stocks, StockData{
				Symbol:    stockConfig.Symbol,
				Label:     stockConfig.Label,
				Price:     0,
				UpdatedAt: time.Now(),
			})
		}
	}
	StockCacheMu.RUnlock()

	recommendations := GetRecommendations()

	response := APIResponse{
		Feeds:           feeds,
		Stocks:          stocks,
		Recommendations: recommendations,
		Timezones:       Cfg.Timezones,
		CurrentTime:     time.Now(),
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

	ClickHistoryMu.Lock()
	ClickHistory = append(ClickHistory, feedback)
	ClickHistoryMu.Unlock()

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
