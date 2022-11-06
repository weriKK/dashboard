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

		crw := newCustomResponseWriter(w)
		next.ServeHTTP(crw, r)

		log.Printf("INFO %s %s %s %s %d %d %s\n", r.RemoteAddr, r.Method, r.RequestURI, r.Proto, crw.status, crw.size, time.Since(start))
	})
}

type customResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

// ResponseWriter.Header is kept as it is

// Overrides ResponseWriter.WriteHeader
func (c *customResponseWriter) WriteHeader(status int) {
	c.status = status
	c.ResponseWriter.WriteHeader(status)
}

// Overrides ResponseWriter.Write
func (c *customResponseWriter) Write(b []byte) (int, error) {
	size, err := c.ResponseWriter.Write(b)
	c.size += size
	return size, err
}

// Overrides Flusher.Flush, some ResponseWriters might implement this
func (c *customResponseWriter) Flush() {
	if f, ok := c.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func newCustomResponseWriter(w http.ResponseWriter) *customResponseWriter {
	// Default status is 200 OK, it's a safe assumption when WriteHeader is
	// never called
	return &customResponseWriter{
		ResponseWriter: w,
		status:         200,
	}
}
