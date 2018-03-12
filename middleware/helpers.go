package middleware

import "net/http"

// Chain applies middlewares to a http.HandlerFunc
func Chain(f http.Handler, m ...func(http.Handler) http.Handler) http.Handler {
	if len(m) == 0 {
		return f
	}

	return m[0](Chain(f, m[1:]...))
}

// UseResponseWriter wraps a handler with a middleware.ResponseWriter
func UseResponseWriter(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrw := NewResponseWriter(w)
		h.ServeHTTP(wrw, r)
	})
}

// GetStatusCode tries to get the StatusCode of a WrappedResponseWriter
func GetStatusCode(w http.ResponseWriter) int {
	if wrw, ok := w.(*ResponseWriter); ok {
		return wrw.StatusCode
	}
	return -1
}

// write to the response and with the status code
func write(w http.ResponseWriter, status int, text string) {
	writeBytes(w, status, []byte(text))
}

// writeStatus writes an empty response with an HTTP status code
func writeStatus(w http.ResponseWriter, status int) {
	writeBytes(w, status, []byte{})
}

// writeBytes to the response and with the status code
func writeBytes(w http.ResponseWriter, status int, text []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(text)
}
