package backend

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter tracks requests per IP
type RateLimiter struct {
	mu      sync.RWMutex
	clients map[string]*ClientRateLimit
}

// ClientRateLimit tracks rate limit info for a single IP
type ClientRateLimit struct {
	requests  []time.Time
	lastCheck time.Time
}

var rateLimiter = &RateLimiter{
	clients: make(map[string]*ClientRateLimit),
}

// GetClientIP extracts client IP from request
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For first (proxy/reverse proxy)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	// Check X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// IsAllowed checks if IP is within rate limit
// Returns true if request is allowed, false if rate limited
func (rl *RateLimiter) IsAllowed(ip string, requestsPerMinute int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	client, exists := rl.clients[ip]

	// Clean up old entries (older than 2 minutes)
	if exists {
		cutoff := now.Add(-2 * time.Minute)
		validRequests := []time.Time{}
		for _, req := range client.requests {
			if req.After(cutoff) {
				validRequests = append(validRequests, req)
			}
		}
		client.requests = validRequests
	}

	// Create new entry if doesn't exist
	if !exists {
		rl.clients[ip] = &ClientRateLimit{
			requests:  []time.Time{now},
			lastCheck: now,
		}
		return true
	}

	// Check if within limit
	if len(client.requests) < requestsPerMinute {
		client.requests = append(client.requests, now)
		return true
	}

	return false
}

// RateLimitMiddleware enforces per-IP rate limiting
// Use with: rateLimitedHandler := RateLimitMiddleware(myHandler, 60) // 60 req/min
func RateLimitMiddleware(handler http.Handler, requestsPerMinute int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := GetClientIP(r)

		if !rateLimiter.IsAllowed(clientIP, requestsPerMinute) {
			w.Header().Set("Retry-After", "60")
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

// MaxBodySizeMiddleware enforces maximum request body size
// Use with: sizedHandler := MaxBodySizeMiddleware(myHandler, 1024*100) // 100 KB
func MaxBodySizeMiddleware(handler http.Handler, maxBytes int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
			return
		}

		handler.ServeHTTP(w, r)
	})
}
