package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestLoggerLevels(t *testing.T) {
	var buf bytes.Buffer

	config := Config{
		Level:  INFO,
		Output: &buf,
		Color:  false,
	}

	logger := NewWithConfig(config)

	// Test that DEBUG is filtered out
	logger.Debug("debug message")
	if buf.Len() > 0 {
		t.Error("DEBUG message should not be logged when level is INFO")
	}

	// Test that INFO is logged
	logger.Info("info message")
	if !strings.Contains(buf.String(), "info message") {
		t.Error("INFO message should be logged")
	}

	// Change level to DEBUG
	logger.SetLevel(DEBUG)
	buf.Reset()
	logger.Debug("debug message 2")
	if !strings.Contains(buf.String(), "debug message 2") {
		t.Error("DEBUG message should be logged when level is DEBUG")
	}
}

func TestLoggerWithFields(t *testing.T) {
	var buf bytes.Buffer

	config := Config{
		Level:  INFO,
		Output: &buf,
		Color:  false,
	}

	logger := NewWithConfig(config)

	// Test WithField
	logger.WithField("user_id", 123).Info("user action")
	output := buf.String()
	if !strings.Contains(output, "user_id=123") {
		t.Error("Field should be included in log output")
	}

	// Test WithFields
	buf.Reset()
	fields := map[string]interface{}{
		"ip":   "192.168.1.1",
		"port": 8080,
	}
	logger.WithFields(fields).Info("connection")
	output = buf.String()
	if !strings.Contains(output, "ip=192.168.1.1") || !strings.Contains(output, "port=8080") {
		t.Error("Multiple fields should be included in log output")
	}
}

func TestJSONFormatter(t *testing.T) {
	var buf bytes.Buffer

	config := Config{
		Level:  INFO,
		Output: &buf,
		JSON:   true,
	}

	logger := NewWithConfig(config)
	logger.WithField("action", "login").Info("user logged in")

	// Parse JSON output
	var data map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &data); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if data["level"] != "INFO" {
		t.Errorf("Expected level INFO, got %v", data["level"])
	}

	if data["action"] != "login" {
		t.Errorf("Expected action login, got %v", data["action"])
	}

	if !strings.Contains(data["message"].(string), "user logged in") {
		t.Error("Message not found in JSON output")
	}
}

func TestContextLogger(t *testing.T) {
	var buf bytes.Buffer

	config := Config{
		Level:  INFO,
		Output: &buf,
		Color:  false,
	}

	logger := NewWithConfig(config)
	ctx := context.Background()

	// Add logger to context
	ctx = WithLogger(ctx, logger)

	// Add fields to context
	ctx = WithField(ctx, "request_id", "abc123")
	ctx = WithField(ctx, "user_id", 456)

	// Get logger from context
	ctxLogger := FromContext(ctx)
	ctxLogger.Info("request processed")

	output := buf.String()
	if !strings.Contains(output, "request_id=abc123") || !strings.Contains(output, "user_id=456") {
		t.Error("Context fields should be included in log output")
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
	}{
		{"debug", DEBUG},
		{"DEBUG", DEBUG},
		{"info", INFO},
		{"INFO", INFO},
		{"warn", WARN},
		{"WARN", WARN},
		{"warning", WARN},
		{"error", ERROR},
		{"ERROR", ERROR},
		{"fatal", FATAL},
		{"FATAL", FATAL},
		{"panic", PANIC},
		{"PANIC", PANIC},
		{"unknown", INFO}, // Default to INFO
	}

	for _, test := range tests {
		result := ParseLevel(test.input)
		if result != test.expected {
			t.Errorf("ParseLevel(%q) = %v, want %v", test.input, result, test.expected)
		}
	}
}

func TestTextFormatterColors(t *testing.T) {

	formatter := &TextFormatter{
		Color:      true,
		TimeFormat: "",
	}

	entry := &Entry{
		Time:    time.Now(),
		Level:   ERROR,
		Message: "test error",
	}

	output, err := formatter.Format(entry)
	if err != nil {
		t.Fatalf("Failed to format entry: %v", err)
	}

	// Check for ANSI color codes
	if !strings.Contains(string(output), "\033[") {
		t.Error("Color codes should be present in output")
	}
}

func TestMultiWriter(t *testing.T) {
	var buf1, buf2 bytes.Buffer

	multiWriter := NewMultiWriter(&buf1, &buf2)
	message := "test message"

	n, err := multiWriter.Write([]byte(message))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if n != len(message) {
		t.Errorf("Write returned %d, want %d", n, len(message))
	}

	if buf1.String() != message {
		t.Errorf("Buffer1 contains %q, want %q", buf1.String(), message)
	}

	if buf2.String() != message {
		t.Errorf("Buffer2 contains %q, want %q", buf2.String(), message)
	}
}

func TestFileHandler(t *testing.T) {
	// Create temporary file
	tmpFile := "/tmp/test.log"

	handler, err := NewFileHandler(tmpFile, 1024, 3) // 1KB max, keep 3 files
	if err != nil {
		t.Fatalf("Failed to create file handler: %v", err)
	}
	defer handler.Close()

	// Write some data
	message := "test log entry\n"
	n, err := handler.Write([]byte(message))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if n != len(message) {
		t.Errorf("Write returned %d, want %d", n, len(message))
	}
}

func TestAsyncLogger(t *testing.T) {
	var buf bytes.Buffer

	config := Config{
		Level:  INFO,
		Output: &buf,
		Color:  false,
	}

	baseLogger := NewWithConfig(config)
	asyncLogger := NewAsyncLogger(baseLogger, 10)

	// Log asynchronously
	for i := 0; i < 5; i++ {
		asyncLogger.Infof("Message %d", i)
	}

	// Give time for processing
	asyncLogger.Shutdown()

	// Check that all messages were logged
	output := buf.String()
	for i := 0; i < 5; i++ {
		if !strings.Contains(output, fmt.Sprintf("Message %d", i)) {
			t.Errorf("Message %d not found in output", i)
		}
	}
}

func TestPrefix(t *testing.T) {
	var buf bytes.Buffer

	config := Config{
		Level:  INFO,
		Output: &buf,
		Color:  false,
	}

	logger := NewWithConfig(config)

	// Test WithPrefix
	prefixedLogger := logger.WithPrefix("Auth")
	prefixedLogger.Info("user authenticated")

	output := buf.String()
	if !strings.Contains(output, "[Auth]") {
		t.Error("Prefix should be included in log output")
	}
}

func BenchmarkLogger(b *testing.B) {
	var buf bytes.Buffer

	config := Config{
		Level:  INFO,
		Output: &buf,
		Color:  false,
		JSON:   false,
	}

	logger := NewWithConfig(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message")
		buf.Reset()
	}
}

func BenchmarkLoggerWithFields(b *testing.B) {
	var buf bytes.Buffer

	config := Config{
		Level:  INFO,
		Output: &buf,
		Color:  false,
		JSON:   false,
	}

	logger := NewWithConfig(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.WithField("iteration", i).Info("benchmark message")
		buf.Reset()
	}
}

func BenchmarkJSONLogger(b *testing.B) {
	var buf bytes.Buffer

	config := Config{
		Level:  INFO,
		Output: &buf,
		JSON:   true,
	}

	logger := NewWithConfig(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.WithField("iteration", i).Info("benchmark message")
		buf.Reset()
	}
}
