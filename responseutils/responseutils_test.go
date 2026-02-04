package responseutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestJSON(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		data       interface{}
		message    string
		expected   JSONResponse
	}{
		{
			name:       "success response",
			statusCode: 200,
			data:       map[string]string{"key": "value"},
			message:    "Operation successful",
			expected: JSONResponse{
				Success: true,
				Message: "Operation successful",
				Data:    map[string]string{"key": "value"},
			},
		},
		{
			name:       "error response",
			statusCode: 400,
			data:       nil,
			message:    "Bad request",
			expected: JSONResponse{
				Success: false,
				Error:   "Bad request",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			JSON(rr, tt.statusCode, tt.data, tt.message)

			if rr.Code != tt.statusCode {
				t.Errorf("expected status %d, got %d", tt.statusCode, rr.Code)
			}

			contentType := rr.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected content type application/json, got %s", contentType)
			}

			var response JSONResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if response.Success != tt.expected.Success {
				t.Errorf("expected success %v, got %v", tt.expected.Success, response.Success)
			}

			if tt.statusCode >= 200 && tt.statusCode < 300 {
				if response.Message != tt.expected.Message {
					t.Errorf("expected message %s, got %s", tt.expected.Message, response.Message)
				}
				if response.Error != "" {
					t.Errorf("expected empty error, got %s", response.Error)
				}
			} else {
				if response.Error != tt.expected.Error {
					t.Errorf("expected error %s, got %s", tt.expected.Error, response.Error)
				}
				if response.Message != "" {
					t.Errorf("expected empty message for error, got %s", response.Message)
				}
			}
		})
	}
}

func TestSuccess(t *testing.T) {
	rr := httptest.NewRecorder()
	data := map[string]string{"name": "test"}
	Success(rr, data, "Test successful")

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var response JSONResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("expected success to be true")
	}

	if response.Message != "Test successful" {
		t.Errorf("expected message 'Test successful', got %s", response.Message)
	}
}

func TestSuccessWithMeta(t *testing.T) {
	rr := httptest.NewRecorder()
	data := map[string]string{"name": "test"}
	meta := map[string]int{"count": 5}
	SuccessWithMeta(rr, data, meta, "Test successful")

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var response SuccessDetail
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("expected success to be true")
	}

	if response.Message != "Test successful" {
		t.Errorf("expected message 'Test successful', got %s", response.Message)
	}

	if response.Meta == nil {
		t.Error("expected meta data")
	}
}

func TestCreated(t *testing.T) {
	rr := httptest.NewRecorder()
	data := map[string]string{"id": "123"}
	Created(rr, data, "Created successfully")

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rr.Code)
	}

	var response JSONResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("expected success to be true")
	}

	if response.Message != "Created successfully" {
		t.Errorf("expected message 'Created successfully', got %s", response.Message)
	}
}

func TestCreatedWithLocation(t *testing.T) {
	rr := httptest.NewRecorder()
	data := map[string]string{"id": "123"}
	location := "/api/users/123"
	CreatedWithLocation(rr, data, location, "Created successfully")

	locHeader := rr.Header().Get("Location")
	if locHeader != location {
		t.Errorf("expected Location header %s, got %s", location, locHeader)
	}

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rr.Code)
	}
}

func TestNoContent(t *testing.T) {
	rr := httptest.NewRecorder()
	NoContent(rr)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", rr.Code)
	}

	if rr.Body.Len() > 0 {
		t.Error("expected empty body for NoContent")
	}
}

func TestErrorResponses(t *testing.T) {
	tests := []struct {
		name       string
		fn         func(w http.ResponseWriter, message ...string)
		statusCode int
		message    string
	}{
		{
			name:       "BadRequest",
			fn:         func(w http.ResponseWriter, message ...string) { BadRequest(w, message...) },
			statusCode: http.StatusBadRequest,
			message:    "Bad request",
		},
		{
			name:       "Unauthorized",
			fn:         func(w http.ResponseWriter, message ...string) { Unauthorized(w, message...) },
			statusCode: http.StatusUnauthorized,
			message:    "Unauthorized",
		},
		{
			name:       "Forbidden",
			fn:         func(w http.ResponseWriter, message ...string) { Forbidden(w, message...) },
			statusCode: http.StatusForbidden,
			message:    "Forbidden",
		},
		{
			name:       "NotFound",
			fn:         func(w http.ResponseWriter, message ...string) { NotFound(w, message...) },
			statusCode: http.StatusNotFound,
			message:    "Resource not found",
		},
		{
			name:       "InternalServerError",
			fn:         func(w http.ResponseWriter, message ...string) { InternalServerError(w, message...) },
			statusCode: http.StatusInternalServerError,
			message:    "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			tt.fn(rr, tt.message)

			if rr.Code != tt.statusCode {
				t.Errorf("expected status %d, got %d", tt.statusCode, rr.Code)
			}

			var response ErrorDetail
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if response.Success {
				t.Error("expected success to be false for error response")
			}

			if response.Message != tt.message {
				t.Errorf("expected message %s, got %s", tt.message, response.Message)
			}

			if response.StatusCode != tt.statusCode {
				t.Errorf("expected status code %d in response, got %d", tt.statusCode, response.StatusCode)
			}

			if response.Timestamp == "" {
				t.Error("expected timestamp in response")
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	rr := httptest.NewRecorder()
	errors := map[string]string{
		"email":    "Email is required",
		"password": "Password must be at least 8 characters",
	}
	ValidationError(rr, errors)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}

	var response ErrorDetail
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Success {
		t.Error("expected success to be false for validation error")
	}

	if len(response.Validation) != len(errors) {
		t.Errorf("expected %d validation errors, got %d", len(errors), len(response.Validation))
	}

	// Check if all errors are present
	for field, expectedMsg := range errors {
		found := false
		for _, validationError := range response.Validation {
			if validationError.Field == field && validationError.Message == expectedMsg {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected validation error for field %s not found", field)
		}
	}
}

func TestValidationErrorWithCode(t *testing.T) {
	rr := httptest.NewRecorder()
	errors := []ValidationFieldError{
		{Field: "email", Message: "Invalid email format", Code: "EMAIL_FORMAT"},
		{Field: "age", Message: "Age must be positive", Code: "AGE_POSITIVE"},
	}
	ValidationErrorWithCode(rr, errors)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}

	var response ErrorDetail
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(response.Validation) != len(errors) {
		t.Errorf("expected %d validation errors, got %d", len(errors), len(response.Validation))
	}

	for i, expected := range errors {
		if response.Validation[i].Field != expected.Field {
			t.Errorf("expected field %s, got %s", expected.Field, response.Validation[i].Field)
		}
		if response.Validation[i].Code != expected.Code {
			t.Errorf("expected code %s, got %s", expected.Code, response.Validation[i].Code)
		}
	}
}

func TestPaginated(t *testing.T) {
	rr := httptest.NewRecorder()
	data := []string{"item1", "item2", "item3"}
	page, perPage, total := 2, 10, 35
	Paginated(rr, data, page, perPage, total, "Items retrieved")

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var response PaginatedResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("expected success to be true")
	}

	if response.Message != "Items retrieved" {
		t.Errorf("expected message 'Items retrieved', got %s", response.Message)
	}

	pagination := response.Pagination
	if pagination.Page != page {
		t.Errorf("expected page %d, got %d", page, pagination.Page)
	}
	if pagination.PerPage != perPage {
		t.Errorf("expected per_page %d, got %d", perPage, pagination.PerPage)
	}
	if pagination.Total != total {
		t.Errorf("expected total %d, got %d", total, pagination.Total)
	}
	if pagination.TotalPages != 4 { // ceil(35/10) = 4
		t.Errorf("expected total_pages 4, got %d", pagination.TotalPages)
	}
	if !pagination.HasPrev {
		t.Error("expected has_prev to be true")
	}
	if !pagination.HasNext {
		t.Error("expected has_next to be true")
	}
	if *pagination.NextPage != 3 {
		t.Errorf("expected next_page 3, got %d", *pagination.NextPage)
	}
	if *pagination.PrevPage != 1 {
		t.Errorf("expected prev_page 1, got %d", *pagination.PrevPage)
	}
}

func TestWritePlain(t *testing.T) {
	rr := httptest.NewRecorder()
	text := "Hello, World!"
	WritePlain(rr, http.StatusOK, text)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "text/plain; charset=utf-8" {
		t.Errorf("expected content type text/plain, got %s", contentType)
	}

	if rr.Body.String() != text {
		t.Errorf("expected body %s, got %s", text, rr.Body.String())
	}
}

func TestWriteHTML(t *testing.T) {
	rr := httptest.NewRecorder()
	html := "<html><body>Hello</body></html>"
	WriteHTML(rr, http.StatusOK, html)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("expected content type text/html, got %s", contentType)
	}

	if rr.Body.String() != html {
		t.Errorf("expected body %s, got %s", html, rr.Body.String())
	}
}

func TestWriteCSV(t *testing.T) {
	rr := httptest.NewRecorder()
	data := [][]string{
		{"Name", "Age", "Email"},
		{"John", "30", "john@example.com"},
		{"Jane", "25", "jane@example.com"},
		{"Bob", "35", "bob,smith@example.com"},
	}
	filename := "users.csv"
	WriteCSV(rr, http.StatusOK, filename, data)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "text/csv; charset=utf-8" {
		t.Errorf("expected content type text/csv, got %s", contentType)
	}

	contentDisposition := rr.Header().Get("Content-Disposition")
	expectedDisposition := fmt.Sprintf("attachment; filename=\"%s\"", filename)
	if contentDisposition != expectedDisposition {
		t.Errorf("expected Content-Disposition %s, got %s", expectedDisposition, contentDisposition)
	}

	expectedCSV := `Name,Age,Email
John,30,john@example.com
Jane,25,jane@example.com
Bob,35,"bob,smith@example.com"
`
	if rr.Body.String() != expectedCSV {
		t.Errorf("CSV output doesn't match expected\nGot: %s\nExpected: %s", rr.Body.String(), expectedCSV)
	}
}

func TestWriteFile(t *testing.T) {
	rr := httptest.NewRecorder()
	content := []byte("This is a test file content")
	fileResponse := FileResponse{
		FileName:    "test.txt",
		Content:     content,
		ContentType: "text/plain",
		Inline:      false,
	}
	WriteFile(rr, fileResponse)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "text/plain" {
		t.Errorf("expected content type text/plain, got %s", contentType)
	}

	contentDisposition := rr.Header().Get("Content-Disposition")
	expectedDisposition := "attachment; filename=\"test.txt\""
	if contentDisposition != expectedDisposition {
		t.Errorf("expected Content-Disposition %s, got %s", expectedDisposition, contentDisposition)
	}

	if rr.Body.String() != string(content) {
		t.Errorf("file content doesn't match")
	}
}

func TestRedirect(t *testing.T) {
	rr := httptest.NewRecorder()
	redirectURL := "https://example.com/new-location"
	Redirect(rr, RedirectResponse{URL: redirectURL, Status: http.StatusFound})

	if rr.Code != http.StatusFound {
		t.Errorf("expected status 302, got %d", rr.Code)
	}

	location := rr.Header().Get("Location")
	if location != redirectURL {
		t.Errorf("expected Location header %s, got %s", redirectURL, location)
	}
}

func TestPermanentRedirect(t *testing.T) {
	rr := httptest.NewRecorder()
	url := "https://example.com/permanent"
	PermanentRedirect(rr, url)

	if rr.Code != http.StatusMovedPermanently {
		t.Errorf("expected status 301, got %d", rr.Code)
	}

	location := rr.Header().Get("Location")
	if location != url {
		t.Errorf("expected Location header %s, got %s", url, location)
	}
}

func TestSetHeaders(t *testing.T) {
	rr := httptest.NewRecorder()
	headers := map[string]string{
		"X-Custom-Header": "CustomValue",
		"X-Another":       "AnotherValue",
	}
	SetHeaders(rr, headers)

	for key, value := range headers {
		if rr.Header().Get(key) != value {
			t.Errorf("expected header %s=%s, got %s", key, value, rr.Header().Get(key))
		}
	}
}

func TestSetCORS(t *testing.T) {
	rr := httptest.NewRecorder()
	origin := "https://example.com"
	methods := []string{"GET", "POST", "PUT"}
	SetCORS(rr, origin, methods...)

	if rr.Header().Get("Access-Control-Allow-Origin") != origin {
		t.Errorf("expected origin %s, got %s", origin, rr.Header().Get("Access-Control-Allow-Origin"))
	}

	expectedMethods := "GET, POST, PUT"
	if rr.Header().Get("Access-Control-Allow-Methods") != expectedMethods {
		t.Errorf("expected methods %s, got %s", expectedMethods, rr.Header().Get("Access-Control-Allow-Methods"))
	}

	if rr.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Error("expected Access-Control-Allow-Credentials to be true")
	}
}

func TestSetSecurityHeaders(t *testing.T) {
	rr := httptest.NewRecorder()
	SetSecurityHeaders(rr)

	expectedHeaders := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
	}

	for key, expectedValue := range expectedHeaders {
		actualValue := rr.Header().Get(key)
		if actualValue != expectedValue {
			t.Errorf("expected header %s=%s, got %s", key, expectedValue, actualValue)
		}
	}
}

func TestSetPaginationHeaders(t *testing.T) {
	rr := httptest.NewRecorder()
	page, perPage, total := 2, 10, 35
	SetPaginationHeaders(rr, page, perPage, total)

	expectedHeaders := map[string]string{
		"X-Page":        "2",
		"X-Per-Page":    "10",
		"X-Total":       "35",
		"X-Total-Pages": "4",
		"X-Next-Page":   "3",
		"X-Prev-Page":   "1",
	}

	for key, expectedValue := range expectedHeaders {
		actualValue := rr.Header().Get(key)
		if actualValue != expectedValue {
			t.Errorf("expected header %s=%s, got %s", key, expectedValue, actualValue)
		}
	}
}

func TestWriteJSONP(t *testing.T) {
	rr := httptest.NewRecorder()
	callback := "callback"
	data := map[string]string{"name": "test"}
	WriteJSONP(rr, http.StatusOK, callback, data)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/javascript; charset=utf-8" {
		t.Errorf("expected content type application/javascript, got %s", contentType)
	}

	expectedPrefix := callback + "("
	if !strings.HasPrefix(rr.Body.String(), expectedPrefix) {
		t.Errorf("expected JSONP response to start with %s", expectedPrefix)
	}

	if !strings.HasSuffix(rr.Body.String(), ");") {
		t.Error("expected JSONP response to end with );")
	}
}

func TestWriteStream(t *testing.T) {
	contentType := "application/octet-stream"
	data := []byte("This is a test stream data that will be sent in chunks")
	chunkSize := 10

	// Use a custom recorder that implements http.Flusher
	recorder := &flushableRecorder{httptest.NewRecorder()}
	WriteStream(recorder, contentType, data, chunkSize)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", recorder.Code)
	}

	contentTypeHeader := recorder.Header().Get("Content-Type")
	if contentTypeHeader != contentType {
		t.Errorf("expected content type %s, got %s", contentType, contentTypeHeader)
	}

	transferEncoding := recorder.Header().Get("Transfer-Encoding")
	if transferEncoding != "chunked" {
		t.Errorf("expected Transfer-Encoding chunked, got %s", transferEncoding)
	}
}

func TestWriteHealthCheck(t *testing.T) {
	rr := httptest.NewRecorder()
	details := map[string]interface{}{
		"database": "connected",
		"redis":    "connected",
	}
	WriteHealthCheck(rr, "healthy", details)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("expected status healthy, got %s", response["status"])
	}

	if response["service"] != "API Service" {
		t.Errorf("expected service API Service, got %s", response["service"])
	}

	if response["details"] == nil {
		t.Error("expected details in response")
	}
}

func TestWriteHealthCheckWithDB(t *testing.T) {
	rr := httptest.NewRecorder()
	WriteHealthCheckWithDB(rr, "healthy", "connected")

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response["database"] != "connected" {
		t.Errorf("expected database connected, got %s", response["database"])
	}

	if response["uptime"] == "" {
		t.Error("expected uptime in response")
	}
}

func TestGetUptime(t *testing.T) {
	uptime := GetUptime()
	if uptime == "" {
		t.Error("expected non-empty uptime string")
	}

	// Should contain "seconds"
	if !strings.Contains(uptime, "seconds") {
		t.Errorf("expected uptime to contain 'seconds', got %s", uptime)
	}
}

func TestGetFormattedTimestamp(t *testing.T) {
	timestamp := GetFormattedTimestamp()
	if timestamp == "" {
		t.Error("expected non-empty timestamp")
	}

	// Try to parse it with the current time format
	_, err := time.Parse(TimeFormat, timestamp)
	if err != nil {
		t.Errorf("failed to parse timestamp with format %s: %v", TimeFormat, err)
	}
}

func TestConfigure(t *testing.T) {
	// Save original values
	originalVersion := DefaultVersion
	originalTimeFormat := TimeFormat
	originalEnableTimestamps := EnableTimestamps

	defer func() {
		// Restore original values
		DefaultVersion = originalVersion
		TimeFormat = originalTimeFormat
		EnableTimestamps = originalEnableTimestamps
	}()

	config := map[string]interface{}{
		"version":           "2.0",
		"time_format":       time.RFC822,
		"enable_timestamps": false,
		"enable_request_id": true,
	}
	Configure(config)

	if DefaultVersion != "2.0" {
		t.Errorf("expected version 2.0, got %s", DefaultVersion)
	}

	if TimeFormat != time.RFC822 {
		t.Errorf("expected time format RFC822, got %s", TimeFormat)
	}

	if EnableTimestamps != false {
		t.Errorf("expected enable_timestamps false, got %v", EnableTimestamps)
	}
}

func TestResponseMiddleware(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	middleware := NewResponseMiddleware("1.0", "TestApp")
	handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Success(w, nil, "Test")
	}))

	handler.ServeHTTP(rr, req)

	expectedHeaders := map[string]string{
		"X-API-Version":   "1.0",
		"X-Service-Name":  "TestApp",
		"X-Frame-Options": "DENY",
	}

	for key, expectedValue := range expectedHeaders {
		actualValue := rr.Header().Get(key)
		if actualValue != expectedValue {
			t.Errorf("expected header %s=%s, got %s", key, expectedValue, actualValue)
		}
	}
}

func TestGetStatusCodeText(t *testing.T) {
	tests := []struct {
		code     int
		expected string
	}{
		{200, "OK"},
		{404, "Not Found"},
		{500, "Internal Server Error"},
		{999, ""}, // Unknown status code
	}

	for _, tt := range tests {
		result := GetStatusCodeText(tt.code)
		if result != tt.expected {
			t.Errorf("expected status text for %d to be %s, got %s", tt.code, tt.expected, result)
		}
	}
}

func TestIsSuccessStatusCode(t *testing.T) {
	tests := []struct {
		code     int
		expected bool
	}{
		{200, true},
		{201, true},
		{204, true},
		{299, true},
		{199, false},
		{300, false},
		{404, false},
		{500, false},
	}

	for _, tt := range tests {
		result := IsSuccessStatusCode(tt.code)
		if result != tt.expected {
			t.Errorf("IsSuccessStatusCode(%d) = %v, expected %v", tt.code, result, tt.expected)
		}
	}
}

func TestIsRedirectStatusCode(t *testing.T) {
	tests := []struct {
		code     int
		expected bool
	}{
		{301, true},
		{302, true},
		{303, true},
		{307, true},
		{308, true},
		{200, false},
		{404, false},
	}

	for _, tt := range tests {
		result := IsRedirectStatusCode(tt.code)
		if result != tt.expected {
			t.Errorf("IsRedirectStatusCode(%d) = %v, expected %v", tt.code, result, tt.expected)
		}
	}
}

func TestGetDefaultMessage(t *testing.T) {
	tests := []struct {
		code     int
		expected string
	}{
		{200, "Request successful"},
		{201, "Resource created successfully"},
		{400, "Bad request"},
		{404, "Resource not found"},
		{500, "Internal server error"},
		{999, ""}, // Unknown status code
	}

	for _, tt := range tests {
		result := GetDefaultMessage(tt.code)
		if result != tt.expected {
			t.Errorf("GetDefaultMessage(%d) = %s, expected %s", tt.code, result, tt.expected)
		}
	}
}

func TestNewValidationError(t *testing.T) {
	field := "email"
	message := "Invalid email format"
	code := "EMAIL_INVALID"

	ve := NewValidationError(field, message, code)

	if ve.Field != field {
		t.Errorf("expected field %s, got %s", field, ve.Field)
	}

	if ve.Message != message {
		t.Errorf("expected message %s, got %s", message, ve.Message)
	}

	if ve.Code != code {
		t.Errorf("expected code %s, got %s", code, ve.Code)
	}
}

func TestNewResponseOptions(t *testing.T) {
	options := NewResponseOptions()

	if options.StatusCode != http.StatusOK {
		t.Errorf("expected status code 200, got %d", options.StatusCode)
	}

	if options.Message != "" {
		t.Errorf("expected empty message, got %s", options.Message)
	}

	if options.Timestamp == "" {
		t.Error("expected timestamp")
	}

	if options.Version != DefaultVersion {
		t.Errorf("expected version %s, got %s", DefaultVersion, options.Version)
	}

	if options.Headers == nil {
		t.Error("expected headers map")
	}
}

func TestWriteCustom(t *testing.T) {
	rr := httptest.NewRecorder()
	contentType := "application/xml"
	body := []byte("<response><status>success</status></response>")
	WriteCustom(rr, http.StatusOK, contentType, body)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	actualContentType := rr.Header().Get("Content-Type")
	if actualContentType != contentType {
		t.Errorf("expected content type %s, got %s", contentType, actualContentType)
	}

	if !bytes.Equal(rr.Body.Bytes(), body) {
		t.Errorf("body doesn't match expected")
	}
}

func TestSetNoCache(t *testing.T) {
	rr := httptest.NewRecorder()
	SetNoCache(rr)

	expectedHeaders := map[string]string{
		"Cache-Control": "no-cache, no-store, must-revalidate",
		"Pragma":        "no-cache",
		"Expires":       "0",
	}

	for key, expectedValue := range expectedHeaders {
		actualValue := rr.Header().Get(key)
		if actualValue != expectedValue {
			t.Errorf("expected header %s=%s, got %s", key, expectedValue, actualValue)
		}
	}
}

// flushableRecorder implements http.Flusher for testing
type flushableRecorder struct {
	*httptest.ResponseRecorder
}

func (f *flushableRecorder) Flush() {
	// For testing purposes, we don't need to actually flush
}

func TestWriteSuccessMap(t *testing.T) {
	rr := httptest.NewRecorder()
	data := map[string]interface{}{
		"id":   1,
		"name": "Test",
	}
	WriteSuccessMap(rr, data, "Test successful")

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !response["success"].(bool) {
		t.Error("expected success to be true")
	}

	if response["message"] != "Test successful" {
		t.Errorf("expected message 'Test successful', got %s", response["message"])
	}

	responseData := response["data"].(map[string]interface{})
	if responseData["name"] != "Test" {
		t.Errorf("expected data.name 'Test', got %s", responseData["name"])
	}
}

func TestWriteSuccessSlice(t *testing.T) {
	rr := httptest.NewRecorder()
	data := []interface{}{"item1", "item2", "item3"}
	WriteSuccessSlice(rr, data, "Items retrieved")

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	responseData := response["data"].([]interface{})
	if len(responseData) != len(data) {
		t.Errorf("expected %d items in data, got %d", len(data), len(responseData))
	}
}

func TestJSONWithOptions(t *testing.T) {
	rr := httptest.NewRecorder()
	options := ResponseOptions{
		StatusCode: http.StatusCreated,
		Message:    "Resource created",
		Data:       map[string]string{"id": "123"},
		Timestamp:  "2024-01-01T00:00:00Z",
		Path:       "/api/users",
		RequestID:  "req-123",
		Version:    "2.0",
		Headers: map[string]string{
			"X-Custom-Header": "CustomValue",
		},
	}
	JSONWithOptions(rr, options)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rr.Code)
	}

	// Check custom headers
	if rr.Header().Get("X-Custom-Header") != "CustomValue" {
		t.Error("expected custom header to be set")
	}

	var response JSONResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("expected success to be true")
	}

	if response.Message != options.Message {
		t.Errorf("expected message %s, got %s", options.Message, response.Message)
	}

	if response.Path != options.Path {
		t.Errorf("expected path %s, got %s", options.Path, response.Path)
	}

	if response.RequestID != options.RequestID {
		t.Errorf("expected request_id %s, got %s", options.RequestID, response.RequestID)
	}

	if response.Version != options.Version {
		t.Errorf("expected version %s, got %s", options.Version, response.Version)
	}
}

func TestErrorWithRequestID(t *testing.T) {
	rr := httptest.NewRecorder()
	statusCode := http.StatusBadRequest
	message := "Validation failed"
	requestID := "req-123"
	details := map[string]string{"field": "email"}

	ErrorWithRequestID(rr, statusCode, message, requestID, details)

	if rr.Code != statusCode {
		t.Errorf("expected status %d, got %d", statusCode, rr.Code)
	}

	var response ErrorDetail
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.RequestID != requestID {
		t.Errorf("expected request_id %s, got %s", requestID, response.RequestID)
	}

	if response.Details == nil {
		t.Error("expected details in response")
	}
}
