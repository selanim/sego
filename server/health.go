package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// HealthStatus represents health status
type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Uptime    string                 `json:"uptime"`
	Checks    map[string]CheckResult `json:"checks,omitempty"`
	Version   string                 `json:"version"`
	Service   string                 `json:"service"`
}

// CheckResult represents result of a health check
type CheckResult struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Error     string    `json:"error,omitempty"`
	Latency   string    `json:"latency,omitempty"`
}

// HealthChecker manages health checks
type HealthChecker struct {
	checks    map[string]CheckFunc
	startTime time.Time
	status    HealthStatus
	mu        sync.RWMutex
	stopChan  chan bool
}

// CheckFunc is a function that performs a health check
type CheckFunc func() (bool, error, time.Duration)

// Health manages health checking
type Health struct {
	checker *HealthChecker
	server  *Server
}

// NewHealth creates new health instance
func NewHealth() *Health {
	return &Health{
		checker: NewHealthChecker(),
	}
}

// NewHealthChecker creates new health checker
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks:    make(map[string]CheckFunc),
		startTime: time.Now(),
		status: HealthStatus{
			Status:    "healthy",
			Service:   "sego-api",
			Version:   "1.0.0",
			Timestamp: time.Now().Format(time.RFC3339),
			Uptime:    "0s",
			Checks:    make(map[string]CheckResult),
		},
		stopChan: make(chan bool),
	}
}

// Handler handles health check requests
func (h *Health) Handler(w http.ResponseWriter, r *http.Request) {
	// ONDOA h.mu.RLock() - Health struct haina mu field

	// Update uptime kwenye checker's status
	h.checker.mu.Lock()
	h.checker.status.Uptime = h.GetUptime()
	h.checker.status.Timestamp = time.Now().Format(time.RFC3339)
	h.checker.mu.Unlock()

	// Run all checks
	h.checker.RunChecks()

	// Determine overall status
	h.checker.mu.RLock()
	overallStatus := "healthy"
	for _, check := range h.checker.status.Checks {
		if check.Status == "unhealthy" {
			overallStatus = "unhealthy"
			break
		}
	}
	h.checker.status.Status = overallStatus

	// Copy status for response
	responseStatus := h.checker.status
	h.checker.mu.RUnlock()

	// Set appropriate status code
	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(responseStatus)
}

// ReadyHandler handles readiness probe
func (h *Health) ReadyHandler(w http.ResponseWriter, r *http.Request) {
	// Run critical checks only
	criticalChecks := []string{"database", "cache", "storage"}
	allHealthy := true

	for _, checkName := range criticalChecks {
		if check, exists := h.checker.checks[checkName]; exists {
			healthy, err, _ := check()
			if !healthy {
				log.Printf("Readiness check failed for %s: %v", checkName, err)
				allHealthy = false
			}
		}
	}

	if allHealthy {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready"}`))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status":"not ready"}`))
	}
}

// LiveHandler handles liveness probe
func (h *Health) LiveHandler(w http.ResponseWriter, r *http.Request) {
	// Simple liveness check - just return OK if server is running
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"alive"}`))
}

// RegisterCheck registers a health check
func (h *Health) RegisterCheck(name string, check CheckFunc) {
	h.checker.RegisterCheck(name, check)
}

// AddDefaultChecks adds default health checks
func (h *Health) AddDefaultChecks() {
	// Memory usage check
	h.checker.RegisterCheck("memory", func() (bool, error, time.Duration) {
		start := time.Now()
		// Simple memory check - always return healthy
		// In production, you might check memory usage
		return true, nil, time.Since(start)
	})

	// Disk space check
	h.checker.RegisterCheck("disk", func() (bool, error, time.Duration) {
		start := time.Now()
		// Always return healthy for demo
		return true, nil, time.Since(start)
	})

	// Database check (placeholder)
	h.checker.RegisterCheck("database", func() (bool, error, time.Duration) {
		start := time.Now()
		// Simulate database check
		time.Sleep(50 * time.Millisecond)
		return true, nil, time.Since(start)
	})

	// Cache check (placeholder)
	h.checker.RegisterCheck("cache", func() (bool, error, time.Duration) {
		start := time.Now()
		// Simulate cache check
		time.Sleep(20 * time.Millisecond)
		return true, nil, time.Since(start)
	})

	// External service check (placeholder)
	h.checker.RegisterCheck("external_service", func() (bool, error, time.Duration) {
		start := time.Now()
		// Simulate external service check
		time.Sleep(100 * time.Millisecond)
		return true, nil, time.Since(start)
	})
}

// StartChecker starts periodic health checking
func (h *Health) StartChecker() {
	h.checker.Start(30 * time.Second) // Check every 30 seconds
}

// StopChecker stops health checking
func (h *Health) StopChecker() {
	h.checker.Stop()
}

// GetUptime returns server uptime
func (h *Health) GetUptime() string {
	return formatDuration(time.Since(h.checker.startTime))
}

// GetStatus returns current health status
func (h *Health) GetStatus() HealthStatus {
	// Tumia checker's mutex badala ya h.mu
	h.checker.mu.RLock()
	defer h.checker.mu.RUnlock()

	// Update uptime
	h.checker.status.Uptime = h.GetUptime()
	h.checker.status.Timestamp = time.Now().Format(time.RFC3339)

	return h.checker.status
}

// HealthChecker methods

// RegisterCheck registers a health check
func (hc *HealthChecker) RegisterCheck(name string, check CheckFunc) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.checks[name] = check
	log.Printf("Health check registered: %s", name)
}

// RunChecks runs all registered health checks
func (hc *HealthChecker) RunChecks() {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.status.Checks = make(map[string]CheckResult)

	for name, check := range hc.checks {
		healthy, err, latency := check()

		result := CheckResult{
			Timestamp: time.Now(),
			Latency:   latency.String(),
		}

		if healthy {
			result.Status = "healthy"
		} else {
			result.Status = "unhealthy"
			if err != nil {
				result.Error = err.Error()
			}
		}

		hc.status.Checks[name] = result
	}

	// Update overall status
	hc.status.Status = "healthy"
	for _, check := range hc.status.Checks {
		if check.Status == "unhealthy" {
			hc.status.Status = "unhealthy"
			break
		}
	}

	hc.status.Timestamp = time.Now().Format(time.RFC3339)
	hc.status.Uptime = formatDuration(time.Since(hc.startTime))
}

// Start starts periodic health checking
func (hc *HealthChecker) Start(interval time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-ticker.C:
				hc.RunChecks()
				// Log if unhealthy
				if hc.status.Status == "unhealthy" {
					log.Println("Health check: UNHEALTHY")
				}
			case <-hc.stopChan:
				ticker.Stop()
				return
			}
		}
	}()

	log.Printf("Health checker started with %v interval", interval)
}

// Stop stops health checking
func (hc *HealthChecker) Stop() {
	select {
	case hc.stopChan <- true:
		log.Println("Health checker stopped")
	default:
		// Already stopped
	}
}

// GetCheckResult gets result of a specific check
func (hc *HealthChecker) GetCheckResult(name string) (CheckResult, bool) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	result, exists := hc.status.Checks[name]
	return result, exists
}

// GetAllResults returns all check results
func (hc *HealthChecker) GetAllResults() map[string]CheckResult {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	results := make(map[string]CheckResult)
	for k, v := range hc.status.Checks {
		results[k] = v
	}

	return results
}

// IsHealthy checks if all checks are healthy
func (hc *HealthChecker) IsHealthy() bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	return hc.status.Status == "healthy"
}

// Helper functions

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)

	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour

	hours := d / time.Hour
	d -= hours * time.Hour

	minutes := d / time.Minute
	d -= minutes * time.Minute

	seconds := d / time.Second

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// Built-in health checks

// HTTPCheck checks an HTTP endpoint
func HTTPCheck(name, url string, timeout time.Duration) CheckFunc {
	return func() (bool, error, time.Duration) {
		start := time.Now()

		client := &http.Client{Timeout: timeout}
		resp, err := client.Get(url)
		if err != nil {
			return false, err, time.Since(start)
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return true, nil, time.Since(start)
		}

		return false, fmt.Errorf("HTTP %d", resp.StatusCode), time.Since(start)
	}
}

// TCPCheck checks TCP connectivity
func TCPCheck(name, host string, timeout time.Duration) CheckFunc {
	return func() (bool, error, time.Duration) {
		start := time.Now()

		conn, err := net.DialTimeout("tcp", host, timeout)
		if err != nil {
			return false, err, time.Since(start)
		}
		conn.Close()

		return true, nil, time.Since(start)
	}
}

// FileExistsCheck checks if a file exists
func FileExistsCheck(name, path string) CheckFunc {
	return func() (bool, error, time.Duration) {
		start := time.Now()

		_, err := os.Stat(path)
		if err != nil {
			return false, err, time.Since(start)
		}

		return true, nil, time.Since(start)
	}
}

// DirectoryWritableCheck checks if directory is writable
func DirectoryWritableCheck(name, path string) CheckFunc {
	return func() (bool, error, time.Duration) {
		start := time.Now()

		// Try to create a test file
		testFile := filepath.Join(path, ".healthcheck_test")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			return false, err, time.Since(start)
		}

		// Clean up
		os.Remove(testFile)

		return true, nil, time.Since(start)
	}
}
