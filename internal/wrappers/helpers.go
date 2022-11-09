package wrappers

import "net/http"

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
	if c.status == 0 {
		c.status = 200
	}
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
