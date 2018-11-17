package logger

import (
	"net/http"
	"time"
)

// TODO: Log before and after ServeHTTP, that way errors that happen during
// 		 processing the request are easier to identify. (maybe also use a
// 		 correlation id for all logs that belong to a request?)
func Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		Infof("(%s)   > %s %s %s", r.RemoteAddr, r.Method, r.RequestURI, r.Proto)

		crw := newCustomResponseWriter(w)
		next.ServeHTTP(crw, r)

		Infof("(%s) <   %d %d %s", r.RemoteAddr, crw.status, crw.size, time.Since(start))
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
