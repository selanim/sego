package responseutils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// JSONResponse represents a standard JSON response structure
type JSONResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	Error      string      `json:"error,omitempty"`
	StatusCode int         `json:"-"`
	Timestamp  string      `json:"timestamp,omitempty"`
	Path       string      `json:"path,omitempty"`
	RequestID  string      `json:"request_id,omitempty"`
	Version    string      `json:"version,omitempty"`
}

// PaginatedResponse represents a paginated response structure
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
	Timestamp  string      `json:"timestamp,omitempty"`
	Path       string      `json:"path,omitempty"`
}

// Pagination contains pagination metadata
type Pagination struct {
	Page       int  `json:"page"`
	PerPage    int  `json:"per_page"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
	NextPage   *int `json:"next_page,omitempty"`
	PrevPage   *int `json:"prev_page,omitempty"`
}

// ErrorDetail represents a detailed error response
type ErrorDetail struct {
	Success    bool                   `json:"success"`
	Error      string                 `json:"error"`
	Message    string                 `json:"message,omitempty"`
	StatusCode int                    `json:"status_code"`
	Timestamp  string                 `json:"timestamp"`
	Path       string                 `json:"path,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
	Details    interface{}            `json:"details,omitempty"`
	Validation []ValidationFieldError `json:"validation,omitempty"`
}

// ValidationFieldError represents a single validation error
type ValidationFieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// SuccessDetail represents a detailed success response
type SuccessDetail struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
	StatusCode int         `json:"status_code"`
	Timestamp  string      `json:"timestamp"`
	Path       string      `json:"path,omitempty"`
	RequestID  string      `json:"request_id,omitempty"`
	Meta       interface{} `json:"meta,omitempty"`
}

// FileResponse represents a file download response
type FileResponse struct {
	FileName    string
	Content     []byte
	ContentType string
	Inline      bool
}

// RedirectResponse represents a redirect response
type RedirectResponse struct {
	URL    string
	Status int
}

// HTMLResponse represents an HTML response
type HTMLResponse struct {
	HTML   string
	Status int
}

// ResponseOptions contains options for customizing responses
type ResponseOptions struct {
	StatusCode int
	Message    string
	Data       interface{}
	Error      string
	Timestamp  string
	Path       string
	RequestID  string
	Version    string
	Meta       interface{}
	Headers    map[string]string
}

// Default response configuration
var (
	DefaultVersion   = "1.0"
	TimeFormat       = time.RFC3339
	EnableTimestamps = true
	EnableRequestID  = true
	serviceStartTime = time.Now()
)

// JSON writes a JSON response with standard structure
func JSON(w http.ResponseWriter, statusCode int, data interface{}, message ...string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}

	response := JSONResponse{
		Success:    statusCode >= 200 && statusCode < 300,
		Message:    msg,
		Data:       data,
		StatusCode: statusCode,
	}

	if !response.Success && statusCode >= 400 {
		response.Error = msg
		response.Message = ""
	}

	if EnableTimestamps {
		response.Timestamp = time.Now().Format(TimeFormat)
	}

	json.NewEncoder(w).Encode(response)
}

// JSONWithOptions writes a JSON response with custom options
func JSONWithOptions(w http.ResponseWriter, options ResponseOptions) {
	w.Header().Set("Content-Type", "application/json")

	// Set custom headers
	for key, value := range options.Headers {
		w.Header().Set(key, value)
	}

	w.WriteHeader(options.StatusCode)

	response := JSONResponse{
		Success:    options.StatusCode >= 200 && options.StatusCode < 300,
		Message:    options.Message,
		Data:       options.Data,
		Error:      options.Error,
		StatusCode: options.StatusCode,
		Path:       options.Path,
		RequestID:  options.RequestID,
		Version:    options.Version,
	}

	if options.Timestamp != "" {
		response.Timestamp = options.Timestamp
	} else if EnableTimestamps {
		response.Timestamp = time.Now().Format(TimeFormat)
	}

	json.NewEncoder(w).Encode(response)
}

// Success writes a success JSON response (200 OK)
func Success(w http.ResponseWriter, data interface{}, message ...string) {
	msg := "Operation successful"
	if len(message) > 0 {
		msg = message[0]
	}
	JSON(w, http.StatusOK, data, msg)
}

// SuccessWithMeta writes a success response with metadata
func SuccessWithMeta(w http.ResponseWriter, data, meta interface{}, message ...string) {
	msg := "Operation successful"
	if len(message) > 0 {
		msg = message[0]
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := SuccessDetail{
		Success:    true,
		Message:    msg,
		Data:       data,
		Meta:       meta,
		StatusCode: http.StatusOK,
		Timestamp:  time.Now().Format(TimeFormat),
	}

	json.NewEncoder(w).Encode(response)
}

// Created writes a 201 Created response
func Created(w http.ResponseWriter, data interface{}, message ...string) {
	msg := "Resource created successfully"
	if len(message) > 0 {
		msg = message[0]
	}
	JSON(w, http.StatusCreated, data, msg)
}

// CreatedWithLocation writes a 201 Created response with Location header
func CreatedWithLocation(w http.ResponseWriter, data interface{}, location string, message ...string) {
	w.Header().Set("Location", location)

	msg := "Resource created successfully"
	if len(message) > 0 {
		msg = message[0]
	}

	JSON(w, http.StatusCreated, data, msg)
}

// Accepted writes a 202 Accepted response
func Accepted(w http.ResponseWriter, data interface{}, message ...string) {
	msg := "Request accepted for processing"
	if len(message) > 0 {
		msg = message[0]
	}
	JSON(w, http.StatusAccepted, data, msg)
}

// NoContent writes a 204 No Content response
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// PartialContent writes a 206 Partial Content response
func PartialContent(w http.ResponseWriter, data interface{}, message ...string) {
	msg := "Partial content"
	if len(message) > 0 {
		msg = message[0]
	}
	JSON(w, http.StatusPartialContent, data, msg)
}

// Error writes an error JSON response
func Error(w http.ResponseWriter, statusCode int, message string, details ...interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorDetail{
		Success:    false,
		Error:      http.StatusText(statusCode),
		Message:    message,
		StatusCode: statusCode,
		Timestamp:  time.Now().Format(TimeFormat),
	}

	if len(details) > 0 {
		response.Details = details[0]
	}

	json.NewEncoder(w).Encode(response)
}

// ErrorWithRequestID writes an error response with request ID
func ErrorWithRequestID(w http.ResponseWriter, statusCode int, message, requestID string, details ...interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorDetail{
		Success:    false,
		Error:      http.StatusText(statusCode),
		Message:    message,
		StatusCode: statusCode,
		Timestamp:  time.Now().Format(TimeFormat),
		RequestID:  requestID,
	}

	if len(details) > 0 {
		response.Details = details[0]
	}

	json.NewEncoder(w).Encode(response)
}

// BadRequest writes a 400 Bad Request response
func BadRequest(w http.ResponseWriter, message ...string) {
	msg := "Bad request"
	if len(message) > 0 {
		msg = message[0]
	}
	Error(w, http.StatusBadRequest, msg)
}

// Unauthorized writes a 401 Unauthorized response
func Unauthorized(w http.ResponseWriter, message ...string) {
	msg := "Unauthorized"
	if len(message) > 0 {
		msg = message[0]
	}
	Error(w, http.StatusUnauthorized, msg)
}

// Forbidden writes a 403 Forbidden response
func Forbidden(w http.ResponseWriter, message ...string) {
	msg := "Forbidden"
	if len(message) > 0 {
		msg = message[0]
	}
	Error(w, http.StatusForbidden, msg)
}

// NotFound writes a 404 Not Found response
func NotFound(w http.ResponseWriter, message ...string) {
	msg := "Resource not found"
	if len(message) > 0 {
		msg = message[0]
	}
	Error(w, http.StatusNotFound, msg)
}

// MethodNotAllowed writes a 405 Method Not Allowed response
func MethodNotAllowed(w http.ResponseWriter, message ...string) {
	msg := "Method not allowed"
	if len(message) > 0 {
		msg = message[0]
	}
	Error(w, http.StatusMethodNotAllowed, msg)
}

// Conflict writes a 409 Conflict response
func Conflict(w http.ResponseWriter, message ...string) {
	msg := "Resource conflict"
	if len(message) > 0 {
		msg = message[0]
	}
	Error(w, http.StatusConflict, msg)
}

// UnprocessableEntity writes a 422 Unprocessable Entity response
func UnprocessableEntity(w http.ResponseWriter, message ...string) {
	msg := "Unprocessable entity"
	if len(message) > 0 {
		msg = message[0]
	}
	Error(w, http.StatusUnprocessableEntity, msg)
}

// TooManyRequests writes a 429 Too Many Requests response
func TooManyRequests(w http.ResponseWriter, message ...string) {
	msg := "Too many requests"
	if len(message) > 0 {
		msg = message[0]
	}
	Error(w, http.StatusTooManyRequests, msg)
}

// InternalServerError writes a 500 Internal Server Error response
func InternalServerError(w http.ResponseWriter, message ...string) {
	msg := "Internal server error"
	if len(message) > 0 {
		msg = message[0]
	}
	Error(w, http.StatusInternalServerError, msg)
}

// NotImplemented writes a 501 Not Implemented response
func NotImplemented(w http.ResponseWriter, message ...string) {
	msg := "Not implemented"
	if len(message) > 0 {
		msg = message[0]
	}
	Error(w, http.StatusNotImplemented, msg)
}

// BadGateway writes a 502 Bad Gateway response
func BadGateway(w http.ResponseWriter, message ...string) {
	msg := "Bad gateway"
	if len(message) > 0 {
		msg = message[0]
	}
	Error(w, http.StatusBadGateway, msg)
}

// ServiceUnavailable writes a 503 Service Unavailable response
func ServiceUnavailable(w http.ResponseWriter, message ...string) {
	msg := "Service unavailable"
	if len(message) > 0 {
		msg = message[0]
	}
	Error(w, http.StatusServiceUnavailable, msg)
}

// GatewayTimeout writes a 504 Gateway Timeout response
func GatewayTimeout(w http.ResponseWriter, message ...string) {
	msg := "Gateway timeout"
	if len(message) > 0 {
		msg = message[0]
	}
	Error(w, http.StatusGatewayTimeout, msg)
}

// ValidationError writes a validation error response (400)
func ValidationError(w http.ResponseWriter, errors map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	var validationErrors []ValidationFieldError
	for field, message := range errors {
		validationErrors = append(validationErrors, ValidationFieldError{
			Field:   field,
			Message: message,
		})
	}

	response := ErrorDetail{
		Success:    false,
		Error:      "Validation failed",
		Message:    "One or more validation errors occurred",
		StatusCode: http.StatusBadRequest,
		Timestamp:  time.Now().Format(TimeFormat),
		Validation: validationErrors,
	}

	json.NewEncoder(w).Encode(response)
}

// ValidationErrorWithCode writes a validation error response with error codes
func ValidationErrorWithCode(w http.ResponseWriter, errors []ValidationFieldError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	response := ErrorDetail{
		Success:    false,
		Error:      "Validation failed",
		Message:    "One or more validation errors occurred",
		StatusCode: http.StatusBadRequest,
		Timestamp:  time.Now().Format(TimeFormat),
		Validation: errors,
	}

	json.NewEncoder(w).Encode(response)
}

// Paginated writes a paginated JSON response
func Paginated(w http.ResponseWriter, data interface{}, page, perPage, total int, message ...string) {
	w.Header().Set("Content-Type", "application/json")

	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}

	totalPages := 0
	if perPage > 0 {
		totalPages = (total + perPage - 1) / perPage
	}

	hasNext := page < totalPages
	hasPrev := page > 1

	var nextPage *int
	if hasNext {
		np := page + 1
		nextPage = &np
	}

	var prevPage *int
	if hasPrev {
		pp := page - 1
		prevPage = &pp
	}

	response := PaginatedResponse{
		Success: true,
		Message: msg,
		Data:    data,
		Pagination: Pagination{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
			HasNext:    hasNext,
			HasPrev:    hasPrev,
			NextPage:   nextPage,
			PrevPage:   prevPage,
		},
		Timestamp: time.Now().Format(TimeFormat),
	}

	json.NewEncoder(w).Encode(response)
}

// WritePlain writes plain text response
func WritePlain(w http.ResponseWriter, statusCode int, text string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write([]byte(text))
}

// WriteHTML writes HTML response
func WriteHTML(w http.ResponseWriter, statusCode int, html string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write([]byte(html))
}

// WriteHTMLWithStatus writes HTML response with custom status
func WriteHTMLWithStatus(w http.ResponseWriter, html string, statusCode int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write([]byte(html))
}

// WriteXML writes XML response
func WriteXML(w http.ResponseWriter, statusCode int, xml string) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write([]byte(xml))
}

// WriteJSONP writes JSONP response
func WriteJSONP(w http.ResponseWriter, statusCode int, callback string, data interface{}) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.WriteHeader(statusCode)

	jsonData, _ := json.Marshal(data)
	response := fmt.Sprintf("%s(%s);", callback, jsonData)
	w.Write([]byte(response))
}

// WriteCSV writes CSV response
func WriteCSV(w http.ResponseWriter, statusCode int, filename string, data [][]string) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.WriteHeader(statusCode)

	var csvBuilder strings.Builder
	for _, row := range data {
		for i, cell := range row {
			if i > 0 {
				csvBuilder.WriteString(",")
			}
			// Escape quotes and commas
			if strings.ContainsAny(cell, ",\"\n") {
				cell = "\"" + strings.ReplaceAll(cell, "\"", "\"\"") + "\""
			}
			csvBuilder.WriteString(cell)
		}
		csvBuilder.WriteString("\n")
	}

	w.Write([]byte(csvBuilder.String()))
}

// WriteFile sends a file as response
func WriteFile(w http.ResponseWriter, fileResponse FileResponse) {
	contentType := fileResponse.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	w.Header().Set("Content-Type", contentType)

	disposition := "attachment"
	if fileResponse.Inline {
		disposition = "inline"
	}

	if fileResponse.FileName != "" {
		w.Header().Set("Content-Disposition",
			fmt.Sprintf("%s; filename=\"%s\"", disposition, fileResponse.FileName))
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(fileResponse.Content)))
	w.Write(fileResponse.Content)
}

// Redirect sends a redirect response
func Redirect(w http.ResponseWriter, redirectResponse RedirectResponse) {
	if redirectResponse.Status == 0 {
		redirectResponse.Status = http.StatusFound // 302
	}

	w.Header().Set("Location", redirectResponse.URL)
	w.WriteHeader(redirectResponse.Status)
}

// PermanentRedirect sends a permanent redirect (301)
func PermanentRedirect(w http.ResponseWriter, url string) {
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusMovedPermanently)
}

// TemporaryRedirect sends a temporary redirect (307)
func TemporaryRedirect(w http.ResponseWriter, url string) {
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// SetHeaders sets multiple HTTP headers at once
func SetHeaders(w http.ResponseWriter, headers map[string]string) {
	for key, value := range headers {
		w.Header().Set(key, value)
	}
}

// SetHeader sets a single HTTP header
func SetHeader(w http.ResponseWriter, key, value string) {
	w.Header().Set(key, value)
}

// AddHeader adds a header (allows multiple values)
func AddHeader(w http.ResponseWriter, key, value string) {
	w.Header().Add(key, value)
}

// SetCacheControl sets Cache-Control header
func SetCacheControl(w http.ResponseWriter, value string) {
	w.Header().Set("Cache-Control", value)
}

// SetNoCache sets no-cache headers
func SetNoCache(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}

// SetCORS sets CORS headers
func SetCORS(w http.ResponseWriter, origin string, methods ...string) {
	if origin == "*" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	} else if origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}

	if len(methods) > 0 {
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ", "))
	} else {
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	}

	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Max-Age", "86400")
}

// SetContentType sets Content-Type header
func SetContentType(w http.ResponseWriter, contentType string) {
	w.Header().Set("Content-Type", contentType)
}

// SetJSONContentType sets Content-Type to application/json
func SetJSONContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

// SetHTMLContentType sets Content-Type to text/html
func SetHTMLContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
}

// SetDownloadHeader sets headers for file download
func SetDownloadHeader(w http.ResponseWriter, filename string) {
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
}

// SetRateLimitHeaders sets rate limiting headers
func SetRateLimitHeaders(w http.ResponseWriter, limit, remaining, reset string) {
	w.Header().Set("X-RateLimit-Limit", limit)
	w.Header().Set("X-RateLimit-Remaining", remaining)
	w.Header().Set("X-RateLimit-Reset", reset)
}

// SetPaginationHeaders sets pagination headers (for API clients that read headers)
func SetPaginationHeaders(w http.ResponseWriter, page, perPage, total int) {
	w.Header().Set("X-Page", fmt.Sprintf("%d", page))
	w.Header().Set("X-Per-Page", fmt.Sprintf("%d", perPage))
	w.Header().Set("X-Total", fmt.Sprintf("%d", total))

	totalPages := 0
	if perPage > 0 {
		totalPages = (total + perPage - 1) / perPage
	}
	w.Header().Set("X-Total-Pages", fmt.Sprintf("%d", totalPages))

	if page < totalPages {
		w.Header().Set("X-Next-Page", fmt.Sprintf("%d", page+1))
	}
	if page > 1 {
		w.Header().Set("X-Prev-Page", fmt.Sprintf("%d", page-1))
	}
}

// SetRequestIDHeader sets request ID header
func SetRequestIDHeader(w http.ResponseWriter, requestID string) {
	w.Header().Set("X-Request-ID", requestID)
}

// SetVersionHeader sets API version header
func SetVersionHeader(w http.ResponseWriter, version string) {
	w.Header().Set("X-API-Version", version)
}

// SetSecurityHeaders sets security-related headers
func SetSecurityHeaders(w http.ResponseWriter) {
	// Security headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")

	// CSP header (Content Security Policy)
	w.Header().Set("Content-Security-Policy",
		"default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline';")

	// HSTS header (only for HTTPS)
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
}

// WriteError writes an error with stack trace in development
func WriteError(w http.ResponseWriter, statusCode int, message string, stackTrace ...string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"success":    false,
		"error":      http.StatusText(statusCode),
		"message":    message,
		"statusCode": statusCode,
		"timestamp":  time.Now().Format(TimeFormat),
	}

	// Include stack trace in development
	if len(stackTrace) > 0 {
		response["stack_trace"] = stackTrace[0]
	}

	json.NewEncoder(w).Encode(response)
}

// WriteSuccessMap writes a success response from a map
func WriteSuccessMap(w http.ResponseWriter, data map[string]interface{}, message ...string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	msg := "Operation successful"
	if len(message) > 0 {
		msg = message[0]
	}

	response := map[string]interface{}{
		"success":   true,
		"message":   msg,
		"data":      data,
		"timestamp": time.Now().Format(TimeFormat),
	}

	json.NewEncoder(w).Encode(response)
}

// WriteSuccessSlice writes a success response from a slice
func WriteSuccessSlice(w http.ResponseWriter, data []interface{}, message ...string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	msg := "Operation successful"
	if len(message) > 0 {
		msg = message[0]
	}

	response := map[string]interface{}{
		"success":   true,
		"message":   msg,
		"data":      data,
		"timestamp": time.Now().Format(TimeFormat),
	}

	json.NewEncoder(w).Encode(response)
}

// WriteCustom writes a completely custom response
func WriteCustom(w http.ResponseWriter, statusCode int, contentType string, body []byte) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(statusCode)
	w.Write(body)
}

// WriteStream writes a streaming response
func WriteStream(w http.ResponseWriter, contentType string, data []byte, chunkSize int) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)

	// Send data in chunks
	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		w.Write(data[i:end])
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
}

// Status writes just the status code without body
func Status(w http.ResponseWriter, statusCode int) {
	w.WriteHeader(statusCode)
}

// OK writes 200 OK
func OK(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}

// WriteHealthCheck writes a health check response
func WriteHealthCheck(w http.ResponseWriter, status string, details ...map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"status":    status,
		"timestamp": time.Now().Format(TimeFormat),
		"service":   "API Service",
	}

	if len(details) > 0 {
		response["details"] = details[0]
	}

	json.NewEncoder(w).Encode(response)
}

// WriteHealthCheckWithDB writes a health check response with database status
func WriteHealthCheckWithDB(w http.ResponseWriter, status string, dbStatus string) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"status":    status,
		"timestamp": time.Now().Format(TimeFormat),
		"service":   "API Service",
		"database":  dbStatus,
		"uptime":    fmt.Sprintf("%.0f seconds", time.Since(serviceStartTime).Seconds()),
	}

	json.NewEncoder(w).Encode(response)
}

// WriteMetrics writes metrics data
func WriteMetrics(w http.ResponseWriter, metrics map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")

	metrics["timestamp"] = time.Now().Format(TimeFormat)
	json.NewEncoder(w).Encode(metrics)
}

// GetUptime returns the service uptime
func GetUptime() string {
	return fmt.Sprintf("%.0f seconds", time.Since(serviceStartTime).Seconds())
}

// GetFormattedTimestamp returns current timestamp in configured format
func GetFormattedTimestamp() string {
	return time.Now().Format(TimeFormat)
}

// Configure sets global configuration
func Configure(config map[string]interface{}) {
	if version, ok := config["version"].(string); ok {
		DefaultVersion = version
	}

	if timeFormat, ok := config["time_format"].(string); ok {
		TimeFormat = timeFormat
	}

	if enableTimestamps, ok := config["enable_timestamps"].(bool); ok {
		EnableTimestamps = enableTimestamps
	}

	if enableRequestID, ok := config["enable_request_id"].(bool); ok {
		EnableRequestID = enableRequestID
	}
}

// Middleware for adding request context
type ResponseMiddleware struct {
	Version string
	AppName string
}

// NewResponseMiddleware creates a new response middleware
func NewResponseMiddleware(version, appName string) *ResponseMiddleware {
	return &ResponseMiddleware{
		Version: version,
		AppName: appName,
	}
}

// Handler wraps an HTTP handler with response enhancements
func (m *ResponseMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add version header
		w.Header().Set("X-API-Version", m.Version)
		w.Header().Set("X-Service-Name", m.AppName)

		// Add security headers
		SetSecurityHeaders(w)

		// Add timestamp
		if EnableTimestamps {
			w.Header().Set("X-Response-Time", GetFormattedTimestamp())
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// GetStatusCodeText returns the HTTP status text for a status code
func GetStatusCodeText(statusCode int) string {
	return http.StatusText(statusCode)
}

// IsSuccessStatusCode checks if status code is in 2xx range
func IsSuccessStatusCode(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}

// IsClientErrorStatusCode checks if status code is in 4xx range
func IsClientErrorStatusCode(statusCode int) bool {
	return statusCode >= 400 && statusCode < 500
}

// IsServerErrorStatusCode checks if status code is in 5xx range
func IsServerErrorStatusCode(statusCode int) bool {
	return statusCode >= 500 && statusCode < 600
}

// IsRedirectStatusCode checks if status code is a redirect
func IsRedirectStatusCode(statusCode int) bool {
	return statusCode == 301 || statusCode == 302 || statusCode == 303 ||
		statusCode == 307 || statusCode == 308
}

// GetDefaultMessage returns default message for a status code
func GetDefaultMessage(statusCode int) string {
	switch statusCode {
	case 200:
		return "Request successful"
	case 201:
		return "Resource created successfully"
	case 204:
		return "No content"
	case 400:
		return "Bad request"
	case 401:
		return "Unauthorized"
	case 403:
		return "Forbidden"
	case 404:
		return "Resource not found"
	case 409:
		return "Conflict"
	case 422:
		return "Unprocessable entity"
	case 429:
		return "Too many requests"
	case 500:
		return "Internal server error"
	case 502:
		return "Bad gateway"
	case 503:
		return "Service unavailable"
	case 504:
		return "Gateway timeout"
	default:
		return http.StatusText(statusCode)
	}
}

// Helper function to create a validation error
func NewValidationError(field, message, code string) ValidationFieldError {
	return ValidationFieldError{
		Field:   field,
		Message: message,
		Code:    code,
	}
}

// Helper function to create response options
func NewResponseOptions() ResponseOptions {
	return ResponseOptions{
		StatusCode: http.StatusOK,
		Message:    "",
		Timestamp:  GetFormattedTimestamp(),
		Version:    DefaultVersion,
		Headers:    make(map[string]string),
	}
}
