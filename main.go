package main

import (
	"context"
	"dashboard/backend"
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Load configuration
	if err := backend.LoadConfig(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize HMAC authentication
	backend.InitHMAC("")

	// Initialize caches
	backend.InitFeedCache()

	// Start background workers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go backend.RefreshFeedsWorker(ctx)
	go backend.RefreshStockWorker(ctx)

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

	mux.Handle("/api/feeds", backend.RateLimitMiddleware(http.HandlerFunc(backend.HandleGetFeeds), 300))
	mux.Handle("/api/stocks", backend.RateLimitMiddleware(http.HandlerFunc(backend.HandleGetStocks), 300))
	mux.Handle("/api/timezones", backend.RateLimitMiddleware(http.HandlerFunc(backend.HandleGetTimezones), 300))
	mux.Handle("/api/dashboard", backend.RateLimitMiddleware(http.HandlerFunc(backend.HandleDashboard), 300))
	mux.Handle("/api/recommendations", backend.RateLimitMiddleware(http.HandlerFunc(backend.HandleGetRecommendations), 300))
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
