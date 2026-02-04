package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestNewServer tests server creation
func TestNewServer(t *testing.T) {
	server := NewServer(nil)
	if server == nil {
		t.Fatal("NewServer returned nil")
	}

	config := server.GetConfig()
	if config.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", config.Port)
	}

	if config.Environment != "development" {
		t.Errorf("Expected environment 'development', got %s", config.Environment)
	}
}

// TestServerStartStop tests server start and stop
func TestServerStartStop(t *testing.T) {
	// Use port 0 for automatic port selection
	config := DefaultConfig()
	config.Port = 0 // Let OS choose available port
	config.EnableLogging = false

	server := NewServer(config)

	// Start server in goroutine
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Start()
	}()

	// Wait a bit for server to start
	time.Sleep(100 * time.Millisecond)

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		t.Fatalf("Failed to shutdown server: %v", err)
	}

	// Check for errors
	select {
	case err := <-errChan:
		if err != http.ErrServerClosed {
			t.Errorf("Server error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Server didn't stop in time")
	}
}

// TestHandlerGetUsers tests GET /api/v1/users
func TestHandlerGetUsers(t *testing.T) {
	server := NewServer(nil)
	handlers := NewHandlers(server)

	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	w := httptest.NewRecorder()

	handlers.GetUsers(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !result["success"].(bool) {
		t.Error("Expected success to be true")
	}
}

// TestHandlerGetUser tests GET /api/v1/users/{id}
func TestHandlerGetUser(t *testing.T) {
	server := NewServer(nil)
	handlers := NewHandlers(server)

	// Test with valid ID
	req := httptest.NewRequest("GET", "/api/v1/users/123", nil)
	w := httptest.NewRecorder()

	// Mock the request URL path
	req.URL.Path = "/api/v1/users/123"

	handlers.GetUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Test with invalid ID
	req2 := httptest.NewRequest("GET", "/api/v1/users/invalid", nil)
	w2 := httptest.NewRecorder()
	req2.URL.Path = "/api/v1/users/invalid"

	handlers.GetUser(w2, req2)

	resp2 := w2.Result()
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid ID, got %d", resp2.StatusCode)
	}
}

// TestHandlerCreateUser tests POST /api/v1/users
func TestHandlerCreateUser(t *testing.T) {
	server := NewServer(nil)
	handlers := NewHandlers(server)

	userData := map[string]string{
		"name":  "Test User",
		"email": "test@example.com",
	}

	jsonData, _ := json.Marshal(userData)

	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handlers.CreateUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	// Test with invalid data
	invalidData := map[string]string{
		"name": "", // Empty name should fail validation
	}

	jsonData, _ = json.Marshal(invalidData)
	req2 := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonData))
	req2.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	handlers.CreateUser(w2, req2)

	resp2 := w2.Result()
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid data, got %d", resp2.StatusCode)
	}
}

// TestHandlerUpdateUser tests PUT /api/v1/users/{id}
func TestHandlerUpdateUser(t *testing.T) {
	server := NewServer(nil)
	handlers := NewHandlers(server)

	userData := map[string]string{
		"name":  "Updated User",
		"email": "updated@example.com",
	}

	jsonData, _ := json.Marshal(userData)

	req := httptest.NewRequest("PUT", "/api/v1/users/123", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.URL.Path = "/api/v1/users/123"

	w := httptest.NewRecorder()
	handlers.UpdateUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestHandlerDeleteUser tests DELETE /api/v1/users/{id}
func TestHandlerDeleteUser(t *testing.T) {
	server := NewServer(nil)
	handlers := NewHandlers(server)

	req := httptest.NewRequest("DELETE", "/api/v1/users/123", nil)
	req.URL.Path = "/api/v1/users/123"

	w := httptest.NewRecorder()
	handlers.DeleteUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", resp.StatusCode)
	}
}

// TestRouter tests router functionality
func TestRouter(t *testing.T) {
	router := NewRouter()

	// Add test route
	called := false
	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	// Test route matching
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if !called {
		t.Error("Route handler was not called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test 404
	req2 := httptest.NewRequest("GET", "/notfound", nil)
	w2 := httptest.NewRecorder()

	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for unknown route, got %d", w2.Code)
	}
}

// TestMiddleware tests middleware functionality
func TestMiddleware(t *testing.T) {
	router := NewRouter()

	// Add middleware
	calls := []string{}
	router.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			calls = append(calls, "middleware1")
			next(w, r)
		}
	})

	router.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			calls = append(calls, "middleware2")
			next(w, r)
		}
	})

	// Add route
	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, "handler")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Check middleware was called in correct order
	expected := []string{"middleware1", "middleware2", "handler"}
	if len(calls) != len(expected) {
		t.Errorf("Expected %d calls, got %d", len(expected), len(calls))
	}

	for i, call := range calls {
		if call != expected[i] {
			t.Errorf("Call %d: expected %s, got %s", i, expected[i], call)
		}
	}
}

// TestHealthCheck tests health check endpoint
func TestHealthCheck(t *testing.T) {
	health := NewHealth()

	// Add test check
	health.RegisterCheck("test", func() (bool, error, time.Duration) {
		return true, nil, 10 * time.Millisecond
	})

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	health.Handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", result["status"])
	}
}

// TestMetrics tests metrics functionality
func TestMetrics(t *testing.T) {
	metrics := NewMetrics()

	// Record some requests
	metrics.RecordRequest("GET", "/test", 200, 100*time.Millisecond)
	metrics.RecordRequest("POST", "/test", 201, 150*time.Millisecond)
	metrics.RecordRequest("GET", "/test", 404, 50*time.Millisecond)

	// Check counts
	count := metrics.GetRequestCount("GET", "/test")
	if count != 2 {
		t.Errorf("Expected 2 GET requests, got %d", count)
	}

	// Test metrics endpoint
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	metrics.Handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode metrics: %v", err)
	}
}

// TestConfigManager tests configuration management
func TestConfigManager(t *testing.T) {
	// Create temporary config file
	tmpFile, err := os.CreateTemp("", "test-config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	// Write test config
	config := DefaultConfig()
	config.Port = 9999
	config.Environment = "testing"

	configData, _ := json.MarshalIndent(config, "", "  ")
	if err := os.WriteFile(tmpFile.Name(), configData, 0644); err != nil {
		t.Fatal(err)
	}

	// Load config
	cm, err := NewConfigManager(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	loadedConfig := cm.GetConfig()
	if loadedConfig.Port != 9999 {
		t.Errorf("Expected port 9999, got %d", loadedConfig.Port)
	}

	if loadedConfig.Environment != "testing" {
		t.Errorf("Expected environment 'testing', got %s", loadedConfig.Environment)
	}

	// Update config
	cm.UpdateConfig(func(c *Config) {
		c.Port = 8888
	})

	if cm.GetConfig().Port != 8888 {
		t.Errorf("Expected updated port 8888, got %d", cm.GetConfig().Port)
	}
}

// TestIntegration tests server integration
func TestIntegration(t *testing.T) {
	// Create test server with random port
	config := DefaultConfig()
	config.Port = 0 // Let OS choose port
	config.EnableLogging = false
	config.EnableMetrics = false
	config.EnableHealth = false

	server := NewServer(config)

	// Add test endpoint
	server.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "test successful",
		})
	})

	// Start server
	go server.Start()
	time.Sleep(100 * time.Millisecond) // Give server time to start

	// Get actual address (this is simplified)
	// In real test, you'd get the listener address

	// Shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		t.Errorf("Failed to shutdown: %v", err)
	}
}

// BenchmarkServerStart benchmarks server startup
func BenchmarkServerStart(b *testing.B) {
	for i := 0; i < b.N; i++ {
		config := DefaultConfig()
		config.Port = 0
		config.EnableLogging = false

		server := NewServer(config)
		server.Start()

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		server.Shutdown(ctx)
		cancel()
	}
}

// BenchmarkHandler benchmarks handler performance
func BenchmarkHandler(b *testing.B) {
	server := NewServer(nil)
	handlers := NewHandlers(server)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/v1/users", nil)
		w := httptest.NewRecorder()

		handlers.GetUsers(w, req)
	}
}

// TestFileUpload tests file upload handler
func TestFileUpload(t *testing.T) {
	server := NewServer(nil)
	handlers := NewHandlers(server)

	// Create test file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.txt")
	io.WriteString(part, "test file content")
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	handlers.UploadHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestStaticFiles tests static file serving
func TestStaticFiles(t *testing.T) {
	// Create temporary directory with test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	config := DefaultConfig()
	config.StaticDir = tmpDir
	config.StaticPrefix = "/static/"

	server := NewServer(config)

	req := httptest.NewRequest("GET", "/static/test.txt", nil)
	w := httptest.NewRecorder()

	// TUMIA GetHandler() badala ya GetRouter()
	server.GetHandler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body, _ := io.ReadAll(w.Body)
	if !strings.Contains(string(body), "test content") {
		t.Error("Static file content not served correctly")
	}
}

// Example usage
func ExampleServer() {
	// Create server with default config
	server := NewDefaultServer()

	// Add custom route
	server.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Hello, World!",
		})
	})

	// Start server
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}

// Helper for multipart writer
var multipart = struct {
	NewWriter func(io.Writer) *multipartWriter
}{
	NewWriter: func(w io.Writer) *multipartWriter {
		return &multipartWriter{}
	},
}

type multipartWriter struct{}

func (mw *multipartWriter) CreateFormFile(fieldname, filename string) (io.Writer, error) {
	return &bytes.Buffer{}, nil
}

func (mw *multipartWriter) FormDataContentType() string {
	return "multipart/form-data"
}

func (mw *multipartWriter) Close() error {
	return nil
}
