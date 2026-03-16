package main

import (
	"context"
	"dashboard/backend"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	backend.AppVersion = os.Getenv("DASHBOARD_VERSION")

	// Load configuration
	if err := backend.LoadConfig(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize HMAC authentication
	backend.InitHMAC("")

	// Initialize caches
	backend.InitFeedCache()

	// Initialize SQLite store for ML persistence
	dbPath := backend.Cfg.ML.DBPath
	if dbPath == "" {
		dbPath = "data/ml_preferences.db"
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}
	if err := backend.OpenStore(dbPath); err != nil {
		log.Fatalf("Failed to open store: %v", err)
	}
	defer backend.CloseStore()

	if err := backend.PruneOldEvents(backend.Cfg.ML.RetentionDays); err != nil {
		log.Printf("Warning: failed to prune old events: %v", err)
	}
	if err := backend.LoadTokenWeights(); err != nil {
		log.Fatalf("Failed to load token weights: %v", err)
	}
	if err := backend.ApplyTokenDecay(); err != nil {
		log.Printf("Warning: failed to apply token decay: %v", err)
	}

	// Start background workers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go backend.RefreshFeedsWorker(ctx)

	// Initial feed fetch
	for _, category := range backend.Cfg.Feeds {
		for _, source := range category.Sources {
			go func(cat string, src backend.FeedSource) {
				backend.FeedCacheMu.Lock()
				_, err := backend.FetchFeed(context.Background(), src, cat)
				if err != nil {
					log.Printf("Error fetching %s: %v", src.Name, err)
				}
				backend.FeedCacheMu.Unlock()
			}(category.Category, source)
		}
	}

	// Setup HTTP routes
	mux := http.NewServeMux()

	// API endpoints with rate limiting
	// Read endpoints: 300 requests/minute per IP
	// Write endpoints: 120 requests/minute per IP

	mux.Handle("/api/dashboard", backend.RateLimitMiddleware(http.HandlerFunc(backend.HandleDashboard), 300))
	// HMAC-protected write endpoint (mandatory)
	feedbackHandler := backend.RequireHMACAuth(
		backend.MaxBodySizeMiddleware(http.HandlerFunc(backend.HandleClickFeedback), 1024*10),
		true, // requireSecret=true: HMAC must be configured
	)
	feedbackHandler = backend.RateLimitMiddleware(feedbackHandler, 120)
	mux.Handle("/api/feedback", feedbackHandler)

	// Serve frontend
	mux.HandleFunc("/", backend.HandleFrontend)
	// Apply CORS middleware
	handler := backend.CORSMiddleware(mux)

	// Start server
	addr := fmt.Sprintf(":%d", backend.Cfg.Server.Port)
	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
