package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Route represents a single route
type Route struct {
	Method      string
	Path        string
	Handler     http.HandlerFunc
	Middlewares []func(http.HandlerFunc) http.HandlerFunc
}

// Router handles routing
type Router struct {
	routes     []Route
	middleware []func(http.HandlerFunc) http.HandlerFunc
	mu         sync.RWMutex
}

// Server represents the HTTP server
type Server struct {
	httpServer *http.Server
	router     *Router
	config     *Config
	middleware *Middleware
	handlers   *Handlers
	metrics    *Metrics
	health     *Health
	listener   net.Listener
}

// Config holds server configuration
type Config struct {
	Host              string
	Port              int
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	MaxHeaderBytes    int
	ShutdownTimeout   time.Duration
	EnableCORS        bool
	EnableCompression bool
	EnableLogging     bool
	EnableMetrics     bool
	EnableHealth      bool
	EnableHTTP2       bool
	TLSCertFile       string
	TLSKeyFile        string
	TrustedProxies    []string
	Environment       string
	StaticDir         string
	StaticPrefix      string
}

// DefaultConfig returns default server configuration
func DefaultConfig() *Config {
	return &Config{
		Host:              "0.0.0.0",
		Port:              8080,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MB
		ShutdownTimeout:   10 * time.Second,
		EnableCORS:        true,
		EnableCompression: true,
		EnableLogging:     true,
		EnableMetrics:     true,
		EnableHealth:      true,
		EnableHTTP2:       true,
		TrustedProxies:    []string{"127.0.0.1", "localhost"},
		Environment:       "development",
		StaticDir:         "./static",
		StaticPrefix:      "/static/",
	}
}

// NewRouter creates a new router
func NewRouter() *Router {
	return &Router{
		routes:     []Route{},
		middleware: []func(http.HandlerFunc) http.HandlerFunc{},
	}
}

// NewServer creates a new HTTP server
func NewServer(config *Config) *Server {
	if config == nil {
		config = DefaultConfig()
	}

	// Create router
	router := NewRouter()

	// Create server instance
	server := &Server{
		router: router,
		config: config,
	}

	// Initialize components
	server.middleware = NewMiddleware(server)
	server.handlers = NewHandlers(server)
	server.metrics = NewMetrics()
	server.health = NewHealth()

	// Setup routes and middleware
	server.setupMiddleware()
	server.setupRoutes()

	// Create HTTP server
	server.httpServer = &http.Server{
		Addr:           fmt.Sprintf("%s:%d", config.Host, config.Port),
		Handler:        server.middleware.apply(router),
		ReadTimeout:    config.ReadTimeout,
		WriteTimeout:   config.WriteTimeout,
		IdleTimeout:    config.IdleTimeout,
		MaxHeaderBytes: config.MaxHeaderBytes,
	}

	// Configure TLS if enabled
	if config.TLSCertFile != "" && config.TLSKeyFile != "" {
		server.httpServer.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			CurvePreferences: []tls.CurveID{
				tls.X25519,
				tls.CurveP256,
			},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		}
	}

	return server
}

// setupMiddleware configures server middleware
func (s *Server) setupMiddleware() {
	if s.config.EnableLogging {
		s.middleware.Use(LoggingMiddleware)
	}

	if s.config.EnableCORS {
		s.middleware.Use(CORSMiddleware)
	}

	if s.config.EnableCompression {
		s.middleware.Use(CompressionMiddleware)
	}

	// Always add security middleware
	s.middleware.Use(SecurityMiddleware)

	// Add request ID middleware
	s.middleware.Use(RequestIDMiddleware)

	// Add recovery middleware
	s.middleware.Use(RecoveryMiddleware)

	// Add timeout middleware
	s.middleware.Use(TimeoutMiddleware(30 * time.Second))
}

// setupRoutes configures server routes
func (s *Server) setupRoutes() {
	// Health check endpoints
	if s.config.EnableHealth {
		s.router.Get("/api/v1/health", s.health.Handler)
		s.router.Get("/api/v1/health/ready", s.health.ReadyHandler)
		s.router.Get("/api/v1/health/live", s.health.LiveHandler)
	}

	// Metrics endpoint
	if s.config.EnableMetrics {
		s.router.Get("/api/v1/metrics", s.metrics.Handler)
	}

	// Example API routes
	s.router.Get("/api/v1/users", s.handlers.GetUsers)
	s.router.Get("/api/v1/users/{id}", s.handlers.GetUser)
	s.router.Post("/api/v1/users", s.handlers.CreateUser)
	s.router.Put("/api/v1/users/{id}", s.handlers.UpdateUser)
	s.router.Delete("/api/v1/users/{id}", s.handlers.DeleteUser)

	// Static file serving
	if s.config.StaticDir != "" {
		s.router.Get(s.config.StaticPrefix+"*", s.serveStaticFiles())
	}

	// Frontend routes
	s.router.Get("/", s.handlers.ServeFrontend)
	s.router.Get("/about", s.handlers.ServeFrontend)
	s.router.Get("/contact", s.handlers.ServeFrontend)

	// 404 handler
	s.router.NotFound(s.handlers.NotFound)
}

// serveStaticFiles serves static files
func (s *Server) serveStaticFiles() http.HandlerFunc {
	fs := http.FileServer(http.Dir(s.config.StaticDir))
	return func(w http.ResponseWriter, r *http.Request) {
		// Strip prefix
		path := strings.TrimPrefix(r.URL.Path, s.config.StaticPrefix)
		r.URL.Path = path
		fs.ServeHTTP(w, r)
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := s.httpServer.Addr
	log.Printf("🚀 Server starting on http://%s", addr)
	log.Printf("📁 Environment: %s", s.config.Environment)
	log.Printf("⏱️  Timeouts - Read: %v, Write: %v, Idle: %v",
		s.config.ReadTimeout, s.config.WriteTimeout, s.config.IdleTimeout)

	// Start metrics collector if enabled
	if s.config.EnableMetrics {
		s.metrics.StartCollector()
	}

	// Start health checker if enabled
	if s.config.EnableHealth {
		s.health.StartChecker()
	}

	// Create listener
	var err error
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	// Start server
	if s.config.TLSCertFile != "" && s.config.TLSKeyFile != "" {
		log.Printf("🔒 HTTPS enabled")
		return s.httpServer.ServeTLS(s.listener, s.config.TLSCertFile, s.config.TLSKeyFile)
	}

	return s.httpServer.Serve(s.listener)
}

// StartTLS starts the server with TLS
func (s *Server) StartTLS(certFile, keyFile string) error {
	s.config.TLSCertFile = certFile
	s.config.TLSKeyFile = keyFile
	return s.Start()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("🛑 Server shutting down gracefully...")

	// Stop metrics collector
	if s.config.EnableMetrics {
		s.metrics.StopCollector()
	}

	// Stop health checker
	if s.config.EnableHealth {
		s.health.StopChecker()
	}

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, s.config.ShutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("⚠️  Graceful shutdown failed: %v", err)
		return s.httpServer.Close()
	}

	log.Println("✅ Server shutdown complete")
	return nil
}

// Run starts the server with graceful shutdown
func (s *Server) Run() error {
	// Create channel for OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		if err := s.Start(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Wait for signal or error
	select {
	case sig := <-sigChan:
		log.Printf("📶 Received signal: %v", sig)
		ctx := context.Background()
		return s.Shutdown(ctx)
	case err := <-serverErr:
		log.Printf("❌ Server error: %v", err)
		return err
	}
}

// RunTLS starts the server with TLS and graceful shutdown
func (s *Server) RunTLS(certFile, keyFile string) error {
	s.config.TLSCertFile = certFile
	s.config.TLSKeyFile = keyFile
	return s.Run()
}

// GetHandler returns the HTTP handler for integration
func (s *Server) GetHandler() http.Handler {
	return s.middleware.apply(s.router)
}

// GetHTTPHandler returns the HTTP handler (alias for GetHandler)
func (s *Server) GetHTTPHandler() http.Handler {
	return s.GetHandler()
}

// AddRoute adds a route to the server
func (s *Server) AddRoute(method, path string, handler http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) {
	s.router.AddRoute(method, path, handler, middleware...)
}

// Get adds GET route
func (s *Server) Get(path string, handler http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) {
	s.router.Get(path, handler, middleware...)
}

// Post adds POST route
func (s *Server) Post(path string, handler http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) {
	s.router.Post(path, handler, middleware...)
}

// Put adds PUT route
func (s *Server) Put(path string, handler http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) {
	s.router.Put(path, handler, middleware...)
}

// Delete adds DELETE route
func (s *Server) Delete(path string, handler http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) {
	s.router.Delete(path, handler, middleware...)
}

// Patch adds PATCH route
func (s *Server) Patch(path string, handler http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) {
	s.router.Patch(path, handler, middleware...)
}

// Options adds OPTIONS route
func (s *Server) Options(path string, handler http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) {
	s.router.Options(path, handler, middleware...)
}

// Head adds HEAD route
func (s *Server) Head(path string, handler http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) {
	s.router.Head(path, handler, middleware...)
}

// AddMiddleware adds global middleware
func (s *Server) AddMiddleware(middleware func(http.HandlerFunc) http.HandlerFunc) {
	s.middleware.Use(middleware)
}

// AddRouteMiddleware adds middleware to specific route
func (s *Server) AddRouteMiddleware(method, path string, middleware ...func(http.HandlerFunc) http.HandlerFunc) {
	s.router.AddMiddlewareToRoute(method, path, middleware...)
}

// SetNotFoundHandler sets custom 404 handler
func (s *Server) SetNotFoundHandler(handler http.HandlerFunc) {
	s.router.NotFound(handler)
}

// GetConfig returns server configuration
func (s *Server) GetConfig() *Config {
	return s.config
}

// GetMetrics returns metrics instance
func (s *Server) GetMetrics() *Metrics {
	return s.metrics
}

// GetHealth returns health instance
func (s *Server) GetHealth() *Health {
	return s.health
}

// UpdateConfig updates server configuration
func (s *Server) UpdateConfig(updater func(*Config)) {
	updater(s.config)

	// Recreate HTTP server with new config
	s.httpServer = &http.Server{
		Addr:           fmt.Sprintf("%s:%d", s.config.Host, s.config.Port),
		Handler:        s.middleware.apply(s.router),
		ReadTimeout:    s.config.ReadTimeout,
		WriteTimeout:   s.config.WriteTimeout,
		IdleTimeout:    s.config.IdleTimeout,
		MaxHeaderBytes: s.config.MaxHeaderBytes,
	}
}

// PrintRoutes prints all registered routes
func (s *Server) PrintRoutes() {
	log.Println("🛣️  Registered Routes:")
	s.router.PrintRoutes()
}

// WaitForShutdown waits for shutdown signal
func (s *Server) WaitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	s.Shutdown(context.Background())
}

// ServeStatic serves static files
func (s *Server) ServeStatic(prefix, directory string) {
	s.config.StaticPrefix = prefix
	s.config.StaticDir = directory

	handler := func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, prefix)
		r.URL.Path = path
		http.FileServer(http.Dir(directory)).ServeHTTP(w, r)
	}

	s.router.Get(prefix+"*", handler)
}

// Quick helper functions

// NewDefaultServer creates a server with default configuration
func NewDefaultServer() *Server {
	return NewServer(DefaultConfig())
}

// NewDevelopmentServer creates a server for development
func NewDevelopmentServer() *Server {
	config := DefaultConfig()
	config.Environment = "development"
	config.EnableLogging = true
	return NewServer(config)
}

// NewProductionServer creates a server for production
func NewProductionServer() *Server {
	config := DefaultConfig()
	config.Environment = "production"
	config.ReadTimeout = 30 * time.Second
	config.WriteTimeout = 30 * time.Second
	config.IdleTimeout = 120 * time.Second
	config.ShutdownTimeout = 30 * time.Second
	return NewServer(config)
}

// SimpleStart starts a simple server
func SimpleStart(port int) error {
	config := DefaultConfig()
	config.Port = port
	server := NewServer(config)
	return server.Run()
}
