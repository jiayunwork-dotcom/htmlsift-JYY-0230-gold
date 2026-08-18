// middleware.go provides HTTP middleware utilities for the API server.
package api

import (
	"log"
	"net/http"
	"strings"
	"time"
)

// CORSMiddleware adds permissive CORS headers for development.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware logs incoming requests with method, path, status and duration.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(sw, r)
		duration := time.Since(start)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, sw.status, duration)
	})
}

// statusWriter captures the response status code.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// ContentTypeJSON ensures responses have JSON content type for API endpoints.
func ContentTypeJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
		}
		next.ServeHTTP(w, r)
	})
}

// MaxBodySize limits request body size to prevent abuse.
func MaxBodySize(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// RecoverMiddleware catches panics in handlers and returns a 500 error.
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v", err)
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// RateLimitInfo holds simple rate limiting state (informational only in this
// implementation).
type RateLimitInfo struct {
	RequestCount int
	WindowStart  time.Time
	Limit        int
	Window       time.Duration
}

// NewRateLimitInfo creates a rate limiter info tracker.
func NewRateLimitInfo(limit int, window time.Duration) *RateLimitInfo {
	return &RateLimitInfo{
		WindowStart: time.Now(),
		Limit:       limit,
		Window:      window,
	}
}

// Allow checks if a request is within the rate limit.
func (rl *RateLimitInfo) Allow() bool {
	now := time.Now()
	if now.Sub(rl.WindowStart) > rl.Window {
		rl.WindowStart = now
		rl.RequestCount = 0
	}
	rl.RequestCount++
	return rl.RequestCount <= rl.Limit
}
