package wrappers

import (
	"log"
	"net/http"
	"time"
)

// WithRequestLogger wraps the http handler with request logging
func WithRequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		crw := customResponseWriter{ResponseWriter: w}
		next.ServeHTTP(&crw, r)

		log.Printf("INFO %s %s %s %s %d %d %s\n", r.RemoteAddr, r.Method, r.RequestURI, r.Proto, crw.status, crw.size, time.Since(start))
	})
}
