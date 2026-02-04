package server

import (
	"compress/gzip"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

// Middleware struct
type Middleware struct {
	server *Server
	chain  []func(http.HandlerFunc) http.HandlerFunc
}

// NewMiddleware creates new middleware
func NewMiddleware(server *Server) *Middleware {
	return &Middleware{
		server: server,
		chain:  []func(http.HandlerFunc) http.HandlerFunc{},
	}
}

// Use adds middleware
func (m *Middleware) Use(middleware func(http.HandlerFunc) http.HandlerFunc) {
	m.chain = append(m.chain, middleware)
}

// apply applies middleware to handler
func (m *Middleware) apply(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.ServeHTTP(w, r)
		})

		// Apply middleware in reverse order
		for i := len(m.chain) - 1; i >= 0; i-- {
			h = m.chain[i](h)
		}

		h(w, r)
	})
}

// Common middleware functions

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create response wrapper
		rw := &responseWriter{w, http.StatusOK}

		// Process request
		next(rw, r)

		// Log
		duration := time.Since(start)
		log.Printf("%s %s %d %s %s",
			r.Method,
			r.URL.Path,
			rw.status,
			duration,
			r.UserAgent(),
		)
	}
}

// CORSMiddleware handles CORS
func CORSMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// SecurityMiddleware adds security headers
func SecurityMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// CSP
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")

		if r.TLS != nil {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		next(w, r)
	}
}

// RequestIDMiddleware adds request ID
func RequestIDMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := generateRequestID()
		w.Header().Set("X-Request-ID", requestID)

		ctx := context.WithValue(r.Context(), "request_id", requestID)
		next(w, r.WithContext(ctx))
	}
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC: %v\n%s", err, debug.Stack())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next(w, r)
	}
}

// TimeoutMiddleware adds timeout
func TimeoutMiddleware(timeout time.Duration) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			next(w, r.WithContext(ctx))
		}
	}
}

// CompressionMiddleware compresses responses
func CompressionMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")

		gz := gzip.NewWriter(w)
		defer gz.Close()

		gzw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		next(gzw, r)
	}
}

// AuthMiddleware for authentication
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		// Validate token
		token := parts[1]
		if !isValidToken(token) {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add user to context
		user, err := getUserFromToken(token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user", user)
		next(w, r.WithContext(ctx))
	}
}

// RateLimitMiddleware limits requests
func RateLimitMiddleware(limit int, window time.Duration) func(http.HandlerFunc) http.HandlerFunc {
	type clientInfo struct {
		count    int
		lastSeen time.Time
	}

	clients := make(map[string]*clientInfo)
	var mu sync.RWMutex

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)

			mu.Lock()

			// Clean old entries
			for ip, info := range clients {
				if time.Since(info.lastSeen) > window {
					delete(clients, ip)
				}
			}

			// Get client info
			info, exists := clients[clientIP]
			if !exists {
				info = &clientInfo{}
				clients[clientIP] = info
			}

			// Reset counter if window passed
			if time.Since(info.lastSeen) > window {
				info.count = 0
			}

			// Check limit
			if info.count >= limit {
				mu.Unlock()
				w.Header().Set("Retry-After", fmt.Sprintf("%.0f", window.Seconds()))
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}

			// Update
			info.count++
			info.lastSeen = time.Now()

			mu.Unlock()

			// Set headers
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", limit-info.count))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(window).Unix()))

			next(w, r)
		}
	}
}

// JSONMiddleware sets JSON content type
func JSONMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next(w, r)
	}
}

// CacheControlMiddleware adds cache headers
func CacheControlMiddleware(maxAge time.Duration) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%.0f", maxAge.Seconds()))
			next(w, r)
		}
	}
}

// NoCacheMiddleware adds no-cache headers
func NoCacheMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		next(w, r)
	}
}

// SizeLimitMiddleware limits request size
func SizeLimitMiddleware(maxBytes int64) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next(w, r)
		}
	}
}

// Helper functions

func generateRequestID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func getClientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}

	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	return strings.Split(r.RemoteAddr, ":")[0]
}

func isValidToken(token string) bool {
	// Implement your token validation
	return len(token) > 0
}

func getUserFromToken(token string) (map[string]interface{}, error) {
	// Implement your user extraction
	return map[string]interface{}{
		"id":    "user123",
		"email": "user@example.com",
		"role":  "user",
	}, nil
}

// Response writer wrappers
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
