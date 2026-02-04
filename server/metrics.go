package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Metrics collects and exposes server metrics
type Metrics struct {
	mu             sync.RWMutex
	startTime      time.Time
	requestCounts  map[string]int64
	responseTimes  map[string][]time.Duration
	statusCodes    map[int]int64
	activeRequests int64
	totalRequests  int64
	errorCount     int64
	collector      *MetricsCollector
}

// MetricsCollector periodically collects system metrics
type MetricsCollector struct {
	metrics  *Metrics
	stopChan chan bool
	interval time.Duration
}

// MetricData represents exported metric data
type MetricData struct {
	Timestamp   string                     `json:"timestamp"`
	Uptime      string                     `json:"uptime"`
	Requests    RequestMetrics             `json:"requests"`
	System      SystemMetrics              `json:"system"`
	Endpoints   map[string]EndpointMetrics `json:"endpoints"`
	StatusCodes map[string]int64           `json:"status_codes"`
}

// RequestMetrics represents request-related metrics
type RequestMetrics struct {
	Total          int64   `json:"total"`
	Active         int64   `json:"active"`
	Errors         int64   `json:"errors"`
	RatePerSecond  float64 `json:"rate_per_second"`
	AverageLatency string  `json:"average_latency"`
}

// SystemMetrics represents system metrics
type SystemMetrics struct {
	MemoryAllocated string  `json:"memory_allocated"`
	MemoryUsed      string  `json:"memory_used"`
	Goroutines      int     `json:"goroutines"`
	CPUUsage        float64 `json:"cpu_usage,omitempty"`
}

// EndpointMetrics represents metrics for a specific endpoint
type EndpointMetrics struct {
	Count      int64         `json:"count"`
	AvgTime    time.Duration `json:"avg_time"`
	ErrorCount int64         `json:"error_count"`
	LastCall   time.Time     `json:"last_call"`
}

// NewMetrics creates new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		startTime:     time.Now(),
		requestCounts: make(map[string]int64),
		responseTimes: make(map[string][]time.Duration),
		statusCodes:   make(map[int]int64),
		collector: &MetricsCollector{
			stopChan: make(chan bool),
			interval: 10 * time.Second,
		},
	}
}

// Handler handles metrics endpoint requests
func (m *Metrics) Handler(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := m.GetMetrics()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// GetMetrics returns current metrics
func (m *Metrics) GetMetrics() MetricData {
	uptime := time.Since(m.startTime)

	// Calculate request rate
	rate := float64(m.totalRequests) / uptime.Seconds()

	// Prepare endpoint metrics
	endpointMetrics := make(map[string]EndpointMetrics)
	for path, count := range m.requestCounts {
		var avgTime time.Duration
		if times, exists := m.responseTimes[path]; exists && len(times) > 0 {
			var total time.Duration
			for _, t := range times {
				total += t
			}
			avgTime = total / time.Duration(len(times))
		}

		endpointMetrics[path] = EndpointMetrics{
			Count:    count,
			AvgTime:  avgTime,
			LastCall: time.Now(), // Would track actual last call time
		}
	}

	// Format status codes
	statusCodes := make(map[string]int64)
	for code, count := range m.statusCodes {
		statusCodes[fmt.Sprintf("%d", code)] = count
	}

	return MetricData{
		Timestamp: time.Now().Format(time.RFC3339),
		Uptime:    formatDuration(uptime),
		Requests: RequestMetrics{
			Total:          m.totalRequests,
			Active:         m.activeRequests,
			Errors:         m.errorCount,
			RatePerSecond:  rate,
			AverageLatency: "0ms", // Would calculate actual average
		},
		System: SystemMetrics{
			MemoryAllocated: "0 MB", // Would get from runtime
			MemoryUsed:      "0 MB",
			Goroutines:      0, // Would get from runtime
		},
		Endpoints:   endpointMetrics,
		StatusCodes: statusCodes,
	}
}

// RecordRequest records a request
func (m *Metrics) RecordRequest(method, path string, statusCode int, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Increment counters
	key := fmt.Sprintf("%s %s", method, path)
	m.requestCounts[key]++
	m.totalRequests++
	m.statusCodes[statusCode]++

	// Track response time (keep last 100 samples)
	if _, exists := m.responseTimes[key]; !exists {
		m.responseTimes[key] = make([]time.Duration, 0, 100)
	}

	times := m.responseTimes[key]
	if len(times) >= 100 {
		times = times[1:]
	}
	m.responseTimes[key] = append(times, duration)

	// Track errors
	if statusCode >= 400 {
		m.errorCount++
	}
}

// StartCollector starts metrics collection
func (m *Metrics) StartCollector() {
	m.collector.metrics = m
	go m.collector.start()
	log.Println("Metrics collector started")
}

// StopCollector stops metrics collection
func (m *Metrics) StopCollector() {
	m.collector.stop()
	log.Println("Metrics collector stopped")
}

// IncActiveRequests increments active requests counter
func (m *Metrics) IncActiveRequests() {
	m.mu.Lock()
	m.activeRequests++
	m.mu.Unlock()
}

// DecActiveRequests decrements active requests counter
func (m *Metrics) DecActiveRequests() {
	m.mu.Lock()
	m.activeRequests--
	m.mu.Unlock()
}

// GetRequestCount returns request count for endpoint
func (m *Metrics) GetRequestCount(method, path string) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := fmt.Sprintf("%s %s", method, path)
	return m.requestCounts[key]
}

// GetErrorRate returns error rate
func (m *Metrics) GetErrorRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.totalRequests == 0 {
		return 0
	}

	return float64(m.errorCount) / float64(m.totalRequests)
}

// GetAverageResponseTime returns average response time for endpoint
func (m *Metrics) GetAverageResponseTime(method, path string) time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := fmt.Sprintf("%s %s", method, path)
	times, exists := m.responseTimes[key]
	if !exists || len(times) == 0 {
		return 0
	}

	var total time.Duration
	for _, t := range times {
		total += t
	}

	return total / time.Duration(len(times))
}

// Reset resets all metrics (for testing)
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.requestCounts = make(map[string]int64)
	m.responseTimes = make(map[string][]time.Duration)
	m.statusCodes = make(map[int]int64)
	m.totalRequests = 0
	m.errorCount = 0
	m.activeRequests = 0
	m.startTime = time.Now()
}

// MetricsCollector methods

// start starts the metrics collector
func (mc *MetricsCollector) start() {
	ticker := time.NewTicker(mc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mc.collect()
		case <-mc.stopChan:
			return
		}
	}
}

// stop stops the metrics collector
func (mc *MetricsCollector) stop() {
	select {
	case mc.stopChan <- true:
	default:
	}
}

// collect collects system metrics
func (mc *MetricsCollector) collect() {
	// In a real implementation, this would collect:
	// - Memory usage from runtime.MemStats
	// - Goroutine count from runtime.NumGoroutine
	// - CPU usage
	// - Disk usage
	// - Network stats

	log.Println("Metrics collection cycle completed")
}

// Middleware for metrics collection
func MetricsMiddleware(metrics *Metrics) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Increment active requests
			metrics.IncActiveRequests()
			defer metrics.DecActiveRequests()

			// Create response wrapper to capture status code
			rw := &metricsResponseWriter{ResponseWriter: w, statusCode: 200}

			// Process request
			next(rw, r)

			// Record metrics
			duration := time.Since(start)
			metrics.RecordRequest(r.Method, r.URL.Path, rw.statusCode, duration)

			// Add metrics headers
			w.Header().Set("X-Response-Time", duration.String())
			w.Header().Set("X-Request-Count", fmt.Sprintf("%d", metrics.GetRequestCount(r.Method, r.URL.Path)))
		}
	}
}

// metricsResponseWriter wraps ResponseWriter to capture status code
type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (mrw *metricsResponseWriter) WriteHeader(statusCode int) {
	mrw.statusCode = statusCode
	mrw.ResponseWriter.WriteHeader(statusCode)
}

// Prometheus-style metrics export
func (m *Metrics) ExportPrometheus() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var output string

	// Total requests
	output += "# HELP http_requests_total Total HTTP requests\n"
	output += "# TYPE http_requests_total counter\n"
	output += fmt.Sprintf("http_requests_total %d\n\n", m.totalRequests)

	// Active requests
	output += "# HELP http_requests_active Active HTTP requests\n"
	output += "# TYPE http_requests_active gauge\n"
	output += fmt.Sprintf("http_requests_active %d\n\n", m.activeRequests)

	// Request duration histogram (simplified)
	output += "# HELP http_request_duration_seconds HTTP request duration\n"
	output += "# TYPE http_request_duration_seconds histogram\n"

	// Endpoint-specific metrics
	for key, count := range m.requestCounts {
		output += fmt.Sprintf("http_request_count{endpoint=\"%s\"} %d\n", key, count)
	}

	return output
}

// ExportJSON exports metrics as JSON
func (m *Metrics) ExportJSON() ([]byte, error) {
	metrics := m.GetMetrics()
	return json.MarshalIndent(metrics, "", "  ")
}

// GetUptime returns server uptime
func (m *Metrics) GetUptime() string {
	return formatDuration(time.Since(m.startTime))
}
