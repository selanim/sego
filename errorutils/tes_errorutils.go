package errorutils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	// Setup
	os.Setenv("TZ", "UTC")

	// Run tests
	code := m.Run()

	// Cleanup
	os.Exit(code)
}

func TestEnhancedErrorBasic(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		options     ErrorOptions
		wantMessage string
		wantLevel   ErrorLevel
		wantType    ErrorType
	}{
		{
			name:        "Simple error",
			message:     "test error",
			options:     ErrorOptions{},
			wantMessage: "test error",
			wantLevel:   Error,
			wantType:    Unknown,
		},
		{
			name:    "With custom level and type",
			message: "validation failed",
			options: ErrorOptions{
				Level: Warn,
				Type:  Validation,
			},
			wantMessage: "validation failed",
			wantLevel:   Warn,
			wantType:    Validation,
		},
		{
			name:    "With code",
			message: "not found",
			options: ErrorOptions{
				Code: "NOT_FOUND",
			},
			wantMessage: "not found",
			wantLevel:   Error,
			wantType:    Unknown,
		},
		{
			name:    "No stack",
			message: "error without stack",
			options: ErrorOptions{
				NoStack: true,
			},
			wantMessage: "error without stack",
			wantLevel:   Error,
			wantType:    Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New(tt.message, tt.options)

			if err.Error() != tt.wantMessage {
				t.Errorf("Error() = %q, want %q", err.Error(), tt.wantMessage)
			}

			info := err.Info()
			if info.Message != tt.wantMessage {
				t.Errorf("Info().Message = %q, want %q", info.Message, tt.wantMessage)
			}

			if info.Level != tt.wantLevel {
				t.Errorf("Info().Level = %q, want %q", info.Level, tt.wantLevel)
			}

			if info.Type != tt.wantType {
				t.Errorf("Info().Type = %q, want %q", info.Type, tt.wantType)
			}

			if tt.options.Code != "" && info.Code != tt.options.Code {
				t.Errorf("Info().Code = %q, want %q", info.Code, tt.options.Code)
			}

			if !tt.options.NoStack && len(info.Stack) == 0 {
				t.Error("Expected stack trace, got empty")
			}

			if tt.options.NoStack && len(info.Stack) > 0 {
				t.Error("Expected no stack trace, but got one")
			}

			if info.Timestamp.IsZero() {
				t.Error("Timestamp should not be zero")
			}
		})
	}
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")

	t.Run("Wrap simple error", func(t *testing.T) {
		wrapped := Wrap(originalErr, "wrapped message")

		if wrapped.Error() != "wrapped message" {
			t.Errorf("Error() = %q, want %q", wrapped.Error(), "wrapped message")
		}

		info := wrapped.Info()
		if info.Cause == nil {
			t.Fatal("Cause should not be nil")
		}

		if info.Cause.Message != "original error" {
			t.Errorf("Cause.Message = %q, want %q", info.Cause.Message, "original error")
		}

		// Test Unwrap
		if errors.Unwrap(wrapped) == nil {
			t.Error("Unwrap should return the cause")
		}
	})

	t.Run("Wrap with options", func(t *testing.T) {
		wrapped := Wrap(originalErr, "database error", ErrorOptions{
			Type: Database,
			Code: "DB_CONN_ERR",
		})

		info := wrapped.Info()
		if info.Type != Database {
			t.Errorf("Type = %q, want %q", info.Type, Database)
		}

		if info.Code != "DB_CONN_ERR" {
			t.Errorf("Code = %q, want %q", info.Code, "DB_CONN_ERR")
		}
	})

	t.Run("Wrap nil error", func(t *testing.T) {
		wrapped := Wrap(nil, "message")
		if wrapped != nil {
			t.Error("Wrap should return nil for nil error")
		}
	})
}

func TestIsEnhancedError(t *testing.T) {
	t.Run("Enhanced error", func(t *testing.T) {
		err := New("test")
		if !IsEnhancedError(err) {
			t.Error("IsEnhancedError should return true for EnhancedError")
		}
	})

	t.Run("Standard error", func(t *testing.T) {
		err := errors.New("test")
		if IsEnhancedError(err) {
			t.Error("IsEnhancedError should return false for standard error")
		}
	})

	t.Run("Wrapped standard error", func(t *testing.T) {
		stdErr := errors.New("original")
		enhancedErr := Wrap(stdErr, "wrapped")
		if !IsEnhancedError(enhancedErr) {
			t.Error("IsEnhancedError should return true for wrapped error")
		}
	})
}

func TestToEnhanced(t *testing.T) {
	t.Run("Convert standard error", func(t *testing.T) {
		stdErr := errors.New("standard error")
		enhancedErr := ToEnhanced(stdErr)

		if !IsEnhancedError(enhancedErr) {
			t.Error("ToEnhanced should return EnhancedError")
		}

		if enhancedErr.Error() != "standard error" {
			t.Errorf("Error() = %q, want %q", enhancedErr.Error(), "standard error")
		}
	})

	t.Run("Convert enhanced error", func(t *testing.T) {
		original := New("original")
		converted := ToEnhanced(original)

		if original != converted {
			t.Error("ToEnhanced should return same instance for EnhancedError")
		}
	})

	t.Run("Convert nil", func(t *testing.T) {
		converted := ToEnhanced(nil)
		if converted != nil {
			t.Error("ToEnhanced should return nil for nil input")
		}
	})
}

func TestMetadata(t *testing.T) {
	err := New("test error")

	t.Run("Add single metadata", func(t *testing.T) {
		err.WithMetadata("key1", "value1")

		if !err.HasMetadata("key1") {
			t.Error("HasMetadata should return true for added key")
		}

		value := err.GetMetadata("key1")
		if value != "value1" {
			t.Errorf("GetMetadata = %v, want %v", value, "value1")
		}
	})

	t.Run("Add multiple metadata", func(t *testing.T) {
		metadata := map[string]interface{}{
			"key2": 42,
			"key3": true,
			"key4": []string{"a", "b"},
		}

		err.WithMetadataMap(metadata)

		for key, want := range metadata {
			got := err.GetMetadata(key)
			if got != want {
				t.Errorf("GetMetadata(%q) = %v, want %v", key, got, want)
			}
		}
	})

	t.Run("Nonexistent metadata", func(t *testing.T) {
		if err.HasMetadata("nonexistent") {
			t.Error("HasMetadata should return false for nonexistent key")
		}

		value := err.GetMetadata("nonexistent")
		if value != nil {
			t.Errorf("GetMetadata should return nil for nonexistent key, got %v", value)
		}
	})
}

func TestStringAndJSON(t *testing.T) {
	err := New("test error", ErrorOptions{
		Level:    Warn,
		Type:     Validation,
		Code:     "VALID_ERR",
		Metadata: map[string]interface{}{"field": "email"},
		NoStack:  true,
	})

	t.Run("String representation", func(t *testing.T) {
		str := err.String()

		expectedParts := []string{
			"[WARN] test error",
			"(code: VALID_ERR)",
			"type: VALIDATION",
			"metadata: field=email",
		}

		for _, part := range expectedParts {
			if !strings.Contains(str, part) {
				t.Errorf("String() should contain %q, got %q", part, str)
			}
		}
	})

	t.Run("JSON representation", func(t *testing.T) {
		jsonData, err := err.JSON()
		if err != nil {
			t.Fatalf("JSON() failed: %v", err)
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("Failed to unmarshal JSON: %v", err)
		}

		if parsed["message"] != "test error" {
			t.Errorf("JSON message = %v, want %v", parsed["message"], "test error")
		}

		if parsed["level"] != "WARN" {
			t.Errorf("JSON level = %v, want %v", parsed["level"], "WARN")
		}

		if parsed["code"] != "VALID_ERR" {
			t.Errorf("JSON code = %v, want %v", parsed["code"], "VALID_ERR")
		}
	})

	t.Run("Pretty JSON", func(t *testing.T) {
		prettyJSON, err := err.PrettyJSON()
		if err != nil {
			t.Fatalf("PrettyJSON() failed: %v", err)
		}

		// Just verify it's valid JSON
		var parsed map[string]interface{}
		if err := json.Unmarshal(prettyJSON, &parsed); err != nil {
			t.Fatalf("Failed to unmarshal pretty JSON: %v", err)
		}
	})
}

func TestMultiError(t *testing.T) {
	t.Run("Empty multi error", func(t *testing.T) {
		me := NewMultiError()

		if me.HasErrors() {
			t.Error("HasErrors should return false for empty MultiError")
		}

		if me.Count() != 0 {
			t.Errorf("Count() = %d, want 0", me.Count())
		}

		if me.Error() == "" {
			t.Error("Error() should return a message even for empty MultiError")
		}
	})

	t.Run("With errors", func(t *testing.T) {
		me := NewMultiError(
			errors.New("error 1"),
			errors.New("error 2"),
			errors.New("error 3"),
		)

		if !me.HasErrors() {
			t.Error("HasErrors should return true")
		}

		if me.Count() != 3 {
			t.Errorf("Count() = %d, want 3", me.Count())
		}

		errs := me.Errors()
		if len(errs) != 3 {
			t.Errorf("Errors() returned %d errors, want 3", len(errs))
		}

		errorStr := me.Error()
		if !strings.Contains(errorStr, "3 error(s) occurred:") {
			t.Errorf("Error() should contain count, got: %s", errorStr)
		}

		for i, err := range errs {
			if !strings.Contains(errorStr, fmt.Sprintf("[%d]", i+1)) {
				t.Errorf("Error() should contain error number %d", i+1)
			}
			if !strings.Contains(errorStr, err.Error()) {
				t.Errorf("Error() should contain error message %q", err.Error())
			}
		}
	})

	t.Run("Add errors", func(t *testing.T) {
		me := NewMultiError()

		me.Add(errors.New("first"))
		if me.Count() != 1 {
			t.Errorf("Count() after first add = %d, want 1", me.Count())
		}

		me.Add(nil) // Should ignore nil
		if me.Count() != 1 {
			t.Errorf("Count() after nil add = %d, want 1", me.Count())
		}

		me.Add(errors.New("second"))
		if me.Count() != 2 {
			t.Errorf("Count() after second add = %d, want 2", me.Count())
		}
	})

	t.Run("Clear errors", func(t *testing.T) {
		me := NewMultiError(
			errors.New("error 1"),
			errors.New("error 2"),
		)

		me.Clear()

		if me.HasErrors() {
			t.Error("HasErrors should return false after Clear()")
		}

		if me.Count() != 0 {
			t.Errorf("Count() after Clear() = %d, want 0", me.Count())
		}
	})

	t.Run("Filter errors", func(t *testing.T) {
		err1 := New("error 1", ErrorOptions{Type: Database})
		err2 := New("error 2", ErrorOptions{Type: Validation})
		err3 := New("error 3", ErrorOptions{Type: Database})

		me := NewMultiError(err1, err2, err3)

		filtered := me.Filter(func(err error) bool {
			info := ErrorInfoFromError(err)
			return info.Type == Database
		})

		if filtered.Count() != 2 {
			t.Errorf("Filtered count = %d, want 2", filtered.Count())
		}

		errs := filtered.Errors()
		for _, err := range errs {
			info := ErrorInfoFromError(err)
			if info.Type != Database {
				t.Errorf("Filtered error has type %v, want %v", info.Type, Database)
			}
		}
	})
}

func TestErrorHandler(t *testing.T) {
	t.Run("Create handler", func(t *testing.T) {
		handler := NewErrorHandler(
			WithLogLevel(Warn),
			WithLogFormat("json"),
			WithIncludeStack(false),
			WithStackDepth(20),
		)

		if handler.LogLevel != Warn {
			t.Errorf("LogLevel = %v, want %v", handler.LogLevel, Warn)
		}

		if handler.LogFormat != "json" {
			t.Errorf("LogFormat = %v, want %v", handler.LogFormat, "json")
		}

		if handler.IncludeStack {
			t.Error("IncludeStack should be false")
		}

		if handler.StackDepth != 20 {
			t.Errorf("StackDepth = %d, want %d", handler.StackDepth, 20)
		}
	})

	t.Run("Handle error with context", func(t *testing.T) {
		var logOutput bytes.Buffer
		logger := log.New(&logOutput, "", 0)

		handler := NewErrorHandler(
			WithLogger(logger),
			WithLogLevel(Error),
			WithCaptureContext(true),
		)

		ctx := context.WithValue(context.Background(), "request_id", "req-123")
		err := errors.New("test error")

		handledErr := handler.Handle(ctx, err)

		if handledErr == nil {
			t.Fatal("Handle should return non-nil error")
		}

		// Check if request_id metadata was added
		if enhancedErr, ok := AsEnhancedError(handledErr); ok {
			if !enhancedErr.HasMetadata("request_id") {
				t.Error("Error should have request_id metadata")
			}

			if enhancedErr.GetMetadata("request_id") != "req-123" {
				t.Errorf("request_id metadata = %v, want %v",
					enhancedErr.GetMetadata("request_id"), "req-123")
			}
		}

		// Check if error was logged
		logStr := logOutput.String()
		if !strings.Contains(logStr, "test error") {
			t.Errorf("Log output should contain error message, got: %s", logStr)
		}
	})

	t.Run("Filter by log level", func(t *testing.T) {
		var logOutput bytes.Buffer
		logger := log.New(&logOutput, "", 0)

		handler := NewErrorHandler(
			WithLogger(logger),
			WithLogLevel(Error),
		)

		// Info level should not be logged
		infoErr := New("info error", ErrorOptions{Level: Info})
		handler.Handle(context.Background(), infoErr)

		if logOutput.Len() > 0 {
			t.Error("Info error should not be logged when log level is Error")
		}

		// Error level should be logged
		logOutput.Reset()
		errorErr := New("error error", ErrorOptions{Level: Error})
		handler.Handle(context.Background(), errorErr)

		if logOutput.Len() == 0 {
			t.Error("Error error should be logged when log level is Error")
		}
	})

	t.Run("Recover from panic", func(t *testing.T) {
		var panicCallbackCalled bool

		handler := NewErrorHandler(
			WithOnPanic(func(err error) {
				panicCallbackCalled = true
			}),
		)

		func() {
			defer handler.Recover(context.Background())

			// This will be caught by Recover
			panic("test panic")
		}()

		if !panicCallbackCalled {
			t.Error("OnPanic callback should have been called")
		}
	})

	t.Run("Recover with callback", func(t *testing.T) {
		var callbackCalled bool

		handler := NewErrorHandler()

		func() {
			defer handler.RecoverWithCallback(context.Background(), func(err error) {
				callbackCalled = true
				if err == nil {
					t.Error("Callback should receive error")
				}
			})

			panic("test panic")
		}()

		if !callbackCalled {
			t.Error("Recovery callback should have been called")
		}
	})
}

func TestIsFunctions(t *testing.T) {
	err := New("test", ErrorOptions{
		Type:  Database,
		Level: Error,
		Code:  "DB_001",
	})

	t.Run("IsType", func(t *testing.T) {
		if !IsType(err, Database) {
			t.Error("IsType should return true for Database type")
		}

		if IsType(err, Validation) {
			t.Error("IsType should return false for wrong type")
		}

		// Test with standard error
		stdErr := errors.New("standard")
		if IsType(stdErr, Database) {
			t.Error("IsType should return false for standard error without type")
		}
	})

	t.Run("IsLevel", func(t *testing.T) {
		if !IsLevel(err, Error) {
			t.Error("IsLevel should return true for Error level")
		}

		if IsLevel(err, Warn) {
			t.Error("IsLevel should return false for wrong level")
		}
	})

	t.Run("IsCode", func(t *testing.T) {
		if !IsCode(err, "DB_001") {
			t.Error("IsCode should return true for DB_001 code")
		}

		if IsCode(err, "OTHER") {
			t.Error("IsCode should return false for wrong code")
		}
	})
}

func TestMustFunctions(t *testing.T) {
	t.Run("Must with nil error", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Error("Must should not panic for nil error")
			}
		}()

		Must(nil)
	})

	t.Run("Must with error", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Must should panic for non-nil error")
			}
		}()

		Must(errors.New("error"))
	})

	t.Run("Must2", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			value := Must2("test", nil)
			if value != "test" {
				t.Errorf("Must2 = %v, want %v", value, "test")
			}
		})

		t.Run("Error", func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Error("Must2 should panic when error is not nil")
				}
			}()

			_ = Must2("", errors.New("error"))
		})
	})

	t.Run("Must3", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			a, b := Must3("a", "b", nil)
			if a != "a" || b != "b" {
				t.Errorf("Must3 = %v, %v, want %v, %v", a, b, "a", "b")
			}
		})

		t.Run("Error", func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Error("Must3 should panic when error is not nil")
				}
			}()

			_, _ = Must3("", "", errors.New("error"))
		})
	})
}

func TestSafeExecute(t *testing.T) {
	t.Run("No panic", func(t *testing.T) {
		err := SafeExecute(func() error {
			return nil
		})

		if err != nil {
			t.Errorf("SafeExecute returned error: %v", err)
		}
	})

	t.Run("With panic", func(t *testing.T) {
		panicErr := errors.New("panic error")

		err := SafeExecute(func() error {
			panic(panicErr)
		})

		if err == nil {
			t.Fatal("SafeExecute should return error after panic")
		}

		if !strings.Contains(err.Error(), "panic in SafeExecute") {
			t.Errorf("Error should contain 'panic in SafeExecute', got: %v", err)
		}

		if !strings.Contains(err.Error(), "panic error") {
			t.Errorf("Error should contain panic error, got: %v", err)
		}
	})

	t.Run("With string panic", func(t *testing.T) {
		err := SafeExecute(func() error {
			panic("string panic")
		})

		if err == nil {
			t.Fatal("SafeExecute should return error after string panic")
		}

		if !strings.Contains(err.Error(), "string panic") {
			t.Errorf("Error should contain panic string, got: %v", err)
		}
	})
}

func TestRetry(t *testing.T) {
	t.Run("Success on first try", func(t *testing.T) {
		attempts := 0
		err := Retry(3, time.Millisecond, func() error {
			attempts++
			return nil
		})

		if err != nil {
			t.Errorf("Retry returned error: %v", err)
		}

		if attempts != 1 {
			t.Errorf("Function called %d times, want 1", attempts)
		}
	})

	t.Run("Success after retries", func(t *testing.T) {
		attempts := 0
		err := Retry(3, time.Millisecond, func() error {
			attempts++
			if attempts < 2 {
				return errors.New("temporary error")
			}
			return nil
		})

		if err != nil {
			t.Errorf("Retry returned error: %v", err)
		}

		if attempts != 2 {
			t.Errorf("Function called %d times, want 2", attempts)
		}
	})

	t.Run("Failure after all retries", func(t *testing.T) {
		attempts := 0
		err := Retry(2, time.Millisecond, func() error {
			attempts++
			return errors.New("persistent error")
		})

		if err == nil {
			t.Fatal("Retry should return error after all retries fail")
		}

		if attempts != 2 {
			t.Errorf("Function called %d times, want 2", attempts)
		}

		if !strings.Contains(err.Error(), "failed after 2 attempts") {
			t.Errorf("Error should contain retry count, got: %v", err)
		}

		if !strings.Contains(err.Error(), "persistent error") {
			t.Errorf("Error should contain original error, got: %v", err)
		}
	})
}

func TestRetryWithContext(t *testing.T) {
	t.Run("Context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		attempts := 0
		err := RetryWithContext(ctx, 3, time.Second, func(ctx context.Context) error {
			attempts++
			return errors.New("error")
		})

		if err == nil {
			t.Fatal("RetryWithContext should return error when context is cancelled")
		}

		if attempts != 0 {
			t.Errorf("Function should not be called when context is cancelled, got %d calls", attempts)
		}

		if !strings.Contains(err.Error(), "context cancelled") {
			t.Errorf("Error should mention context cancellation, got: %v", err)
		}
	})

	t.Run("Success with context", func(t *testing.T) {
		ctx := context.Background()
		attempts := 0

		err := RetryWithContext(ctx, 3, time.Millisecond, func(ctx context.Context) error {
			attempts++
			if attempts < 2 {
				return errors.New("temporary error")
			}
			return nil
		})

		if err != nil {
			t.Errorf("RetryWithContext returned error: %v", err)
		}

		if attempts != 2 {
			t.Errorf("Function called %d times, want 2", attempts)
		}
	})
}

func TestPrettyPrint(t *testing.T) {
	err := New("test error", ErrorOptions{
		Type:        Validation,
		Code:        "ERR_001",
		Level:       Warn,
		UserMessage: "Please check your input",
		Metadata: map[string]interface{}{
			"field": "email",
			"value": "invalid@",
		},
		NoStack: true,
	})

	var buf bytes.Buffer
	PrettyPrint(err, &buf)

	output := buf.String()

	expectedParts := []string{
		"Error: test error",
		"Code: ERR_001",
		"Type: VALIDATION",
		"Level: WARN",
		"User Message: Please check your input",
		"Metadata:",
		"  field: email",
		"  value: invalid@",
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("PrettyPrint output should contain %q, got:\n%s", part, output)
		}
	}
}

func TestErrorChaining(t *testing.T) {
	// Create a chain of errors
	root := errors.New("root cause")

	middle := Wrap(root, "middle error", ErrorOptions{
		Type:  Database,
		Code:  "DB_ERR",
		Level: Error,
	})

	top := Wrap(middle, "top error", ErrorOptions{
		Type:  System,
		Code:  "SYS_ERR",
		Level: Fatal,
	})

	// Test error chain
	if top.Error() != "top error" {
		t.Errorf("Top error message = %q, want %q", top.Error(), "top error")
	}

	// Test unwrapping
	unwrapped := errors.Unwrap(top)
	if unwrapped != middle {
		t.Error("Unwrap should return middle error")
	}

	// Test chain length
	info := top.Info()
	if info.Cause == nil {
		t.Fatal("Top error should have cause")
	}

	if info.Cause.Message != "middle error" {
		t.Errorf("Cause message = %q, want %q", info.Cause.Message, "middle error")
	}

	if info.Cause.Cause == nil {
		t.Fatal("Middle error should have cause")
	}

	if info.Cause.Cause.Message != "root cause" {
		t.Errorf("Root cause message = %q, want %q", info.Cause.Cause.Message, "root cause")
	}
}

func BenchmarkEnhancedErrorCreation(b *testing.B) {
	b.Run("Simple error", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = New("benchmark error", ErrorOptions{NoStack: true})
		}
	})

	b.Run("Error with stack", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = New("benchmark error with stack")
		}
	})

	b.Run("Error with metadata", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = New("benchmark error", ErrorOptions{
				NoStack: true,
				Metadata: map[string]interface{}{
					"key1": "value1",
					"key2": 123,
					"key3": true,
				},
			})
		}
	})
}

func BenchmarkErrorWrapping(b *testing.B) {
	original := errors.New("original error")

	b.Run("Wrap standard error", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Wrap(original, "wrapped")
		}
	})

	b.Run("ToEnhanced", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = ToEnhanced(original)
		}
	})
}
