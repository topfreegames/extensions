package middleware

import "net/http"

// ResponseWriter is a wrapper around http.ResponseWriter
type ResponseWriter struct {
	ResponseWriter http.ResponseWriter
	StatusCode     int
}

// NewResponseWriter ctor
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
	}
}

// WriteHeader wraps http.ResponseWriter.WriteHeader
func (w *ResponseWriter) WriteHeader(status int) {
	w.StatusCode = status
	w.ResponseWriter.WriteHeader(status)
}

// Header wraps http.ResponseWriter.Header
func (w *ResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

// Write wraps http.ResponseWriter.Write
func (w *ResponseWriter) Write(b []byte) (int, error) {
	return w.ResponseWriter.Write(b)
}
