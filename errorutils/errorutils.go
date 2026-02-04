// Package errorutils provides enhanced error handling utilities
// including error wrapping, stack traces, multi-error handling,
// error categorization, and structured error reporting.
package errorutils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// ErrorLevel defines the severity level of an error
type ErrorLevel string

const (
	// Debug - debug-level errors for development
	Debug ErrorLevel = "DEBUG"
	// Info - informational errors
	Info ErrorLevel = "INFO"
	// Warn - warning-level errors that don't break functionality
	Warn ErrorLevel = "WARN"
	// Error - standard error level for operational issues
	Error ErrorLevel = "ERROR"
	// Fatal - critical errors that require immediate attention
	Fatal ErrorLevel = "FATAL"
	// Panic - errors that cause panic
	Panic ErrorLevel = "PANIC"
)

// ErrorType categorizes the type of error
type ErrorType string

const (
	// System - system-level errors (file system, network, etc.)
	System ErrorType = "SYSTEM"
	// Database - database-related errors
	Database ErrorType = "DATABASE"
	// Validation - validation errors
	Validation ErrorType = "VALIDATION"
	// BusinessLogic - business logic errors
	BusinessLogic ErrorType = "BUSINESS_LOGIC"
	// ExternalService - third-party service errors
	ExternalService ErrorType = "EXTERNAL_SERVICE"
	// Authentication - authentication/authorization errors
	Authentication ErrorType = "AUTHENTICATION"
	// Configuration - configuration errors
	Configuration ErrorType = "CONFIGURATION"
	// Timeout - timeout errors
	Timeout ErrorType = "TIMEOUT"
	// Unknown - unknown/unclassified errors
	Unknown ErrorType = "UNKNOWN"
)

// StackFrame represents a single frame in the call stack
type StackFrame struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Function string `json:"function"`
}

// ErrorInfo contains structured error information
type ErrorInfo struct {
	Message     string                 `json:"message"`
	Level       ErrorLevel             `json:"level"`
	Type        ErrorType              `json:"type"`
	Code        string                 `json:"code,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Stack       []StackFrame           `json:"stack,omitempty"`
	Cause       *ErrorInfo             `json:"cause,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	UserMessage string                 `json:"user_message,omitempty"`
}

// EnhancedError is a rich error type with structured information
type EnhancedError struct {
	info *ErrorInfo
	mu   sync.RWMutex
}

// ErrorOptions defines options for creating enhanced errors
type ErrorOptions struct {
	Level       ErrorLevel
	Type        ErrorType
	Code        string
	Metadata    map[string]interface{}
	UserMessage string
	SkipFrames  int  // Number of stack frames to skip
	NoStack     bool // If true, don't capture stack trace
	Cause       error
}

// DefaultOptions provides default error options
var DefaultOptions = ErrorOptions{
	Level:      Error,
	Type:       Unknown,
	SkipFrames: 2, // Skip error creation functions
}

// Error returns the error message
func (e *EnhancedError) Error() string {
	if e == nil || e.info == nil {
		return "<nil>"
	}
	return e.info.Message
}

// Unwrap returns the underlying cause error
func (e *EnhancedError) Unwrap() error {
	if e == nil || e.info == nil || e.info.Cause == nil {
		return nil
	}
	// Convert cause ErrorInfo to EnhancedError
	return &EnhancedError{info: e.info.Cause}
}

// Info returns a copy of the error information
func (e *EnhancedError) Info() ErrorInfo {
	if e == nil || e.info == nil {
		return ErrorInfo{}
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	// Deep copy to prevent modification
	info := *e.info

	// Deep copy metadata
	if info.Metadata != nil {
		metadata := make(map[string]interface{})
		for k, v := range info.Metadata {
			metadata[k] = v
		}
		info.Metadata = metadata
	}

	// Deep copy stack
	if info.Stack != nil {
		stack := make([]StackFrame, len(info.Stack))
		copy(stack, info.Stack)
		info.Stack = stack
	}

	return info
}

// String returns a formatted string representation
func (e *EnhancedError) String() string {
	if e == nil || e.info == nil {
		return "<nil>"
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("[%s] %s", e.info.Level, e.info.Message))

	if e.info.Code != "" {
		buf.WriteString(fmt.Sprintf(" (code: %s)", e.info.Code))
	}

	if e.info.Type != Unknown {
		buf.WriteString(fmt.Sprintf(" type: %s", e.info.Type))
	}

	if len(e.info.Metadata) > 0 {
		buf.WriteString(" metadata:")
		for k, v := range e.info.Metadata {
			buf.WriteString(fmt.Sprintf(" %s=%v", k, v))
		}
	}

	if e.info.Cause != nil {
		buf.WriteString(fmt.Sprintf("\nCaused by: %v", e.info.Cause))
	}

	return buf.String()
}

// JSON returns the error as JSON
func (e *EnhancedError) JSON() ([]byte, error) {
	if e == nil || e.info == nil {
		return []byte("null"), nil
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	return json.Marshal(e.info)
}

// PrettyJSON returns indented JSON representation
func (e *EnhancedError) PrettyJSON() ([]byte, error) {
	if e == nil || e.info == nil {
		return []byte("null"), nil
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	return json.MarshalIndent(e.info, "", "  ")
}

// HasMetadata checks if the error has specific metadata
func (e *EnhancedError) HasMetadata(key string) bool {
	if e == nil || e.info == nil {
		return false
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	_, exists := e.info.Metadata[key]
	return exists
}

// GetMetadata returns metadata value
func (e *EnhancedError) GetMetadata(key string) interface{} {
	if e == nil || e.info == nil {
		return nil
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.info.Metadata[key]
}

// WithMetadata adds or updates metadata
func (e *EnhancedError) WithMetadata(key string, value interface{}) *EnhancedError {
	if e == nil || e.info == nil {
		return e
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if e.info.Metadata == nil {
		e.info.Metadata = make(map[string]interface{})
	}
	e.info.Metadata[key] = value

	return e
}

// WithMetadataMap adds multiple metadata entries
func (e *EnhancedError) WithMetadataMap(metadata map[string]interface{}) *EnhancedError {
	if e == nil || e.info == nil || metadata == nil {
		return e
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if e.info.Metadata == nil {
		e.info.Metadata = make(map[string]interface{})
	}

	for k, v := range metadata {
		e.info.Metadata[k] = v
	}

	return e
}

// WithUserMessage sets a user-friendly message
func (e *EnhancedError) WithUserMessage(message string) *EnhancedError {
	if e == nil || e.info == nil {
		return e
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.info.UserMessage = message
	return e
}

// New creates a new enhanced error
func New(message string, opts ...ErrorOptions) *EnhancedError {
	var options ErrorOptions
	if len(opts) > 0 {
		options = opts[0]
	} else {
		options = DefaultOptions
	}

	if options.Level == "" {
		options.Level = DefaultOptions.Level
	}
	if options.Type == "" {
		options.Type = DefaultOptions.Type
	}
	if options.SkipFrames == 0 {
		options.SkipFrames = DefaultOptions.SkipFrames
	}

	info := &ErrorInfo{
		Message:     message,
		Level:       options.Level,
		Type:        options.Type,
		Code:        options.Code,
		Timestamp:   time.Now(),
		Metadata:    options.Metadata,
		UserMessage: options.UserMessage,
	}

	// Capture stack trace if not disabled
	if !options.NoStack {
		info.Stack = captureStack(options.SkipFrames)
	}

	// Handle cause
	if options.Cause != nil {
		info.Cause = errorInfoFromError(options.Cause)
	}

	return &EnhancedError{info: info}
}

// Wrap wraps an existing error with enhanced information
func Wrap(err error, message string, opts ...ErrorOptions) *EnhancedError {
	if err == nil {
		return nil
	}

	var options ErrorOptions
	if len(opts) > 0 {
		options = opts[0]
	} else {
		options = DefaultOptions
	}

	options.Cause = err
	return New(message, options)
}

// Wrapf wraps an error with formatted message
func Wrapf(err error, format string, args ...interface{}) *EnhancedError {
	if err == nil {
		return nil
	}

	return Wrap(err, fmt.Sprintf(format, args...))
}

// Newf creates a new error with formatted message
func Newf(format string, args ...interface{}) *EnhancedError {
	return New(fmt.Sprintf(format, args...))
}

// NewWithCode creates an error with a specific code
func NewWithCode(message, code string, opts ...ErrorOptions) *EnhancedError {
	options := ErrorOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	options.Code = code
	return New(message, options)
}

// NewValidation creates a validation error
func NewValidation(message string, opts ...ErrorOptions) *EnhancedError {
	options := ErrorOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	options.Type = Validation
	if options.Level == "" {
		options.Level = Warn
	}
	return New(message, options)
}

// NewDatabase creates a database error
func NewDatabase(message string, opts ...ErrorOptions) *EnhancedError {
	options := ErrorOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	options.Type = Database
	return New(message, options)
}

// NewSystem creates a system error
func NewSystem(message string, opts ...ErrorOptions) *EnhancedError {
	options := ErrorOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	options.Type = System
	return New(message, options)
}

// NewTimeout creates a timeout error
func NewTimeout(message string, opts ...ErrorOptions) *EnhancedError {
	options := ErrorOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	options.Type = Timeout
	return New(message, options)
}

// IsEnhancedError checks if an error is an EnhancedError
func IsEnhancedError(err error) bool {
	if err == nil {
		return false
	}

	_, ok := err.(*EnhancedError)
	return ok
}

// AsEnhancedError converts an error to EnhancedError if possible
func AsEnhancedError(err error) (*EnhancedError, bool) {
	if err == nil {
		return nil, false
	}

	var enhancedErr *EnhancedError
	if errors.As(err, &enhancedErr) {
		return enhancedErr, true
	}
	return nil, false
}

// ToEnhanced converts any error to EnhancedError
func ToEnhanced(err error, opts ...ErrorOptions) *EnhancedError {
	if err == nil {
		return nil
	}

	// If already enhanced, return as is
	if enhancedErr, ok := AsEnhancedError(err); ok {
		return enhancedErr
	}

	// Wrap the error
	return Wrap(err, err.Error(), opts...)
}

// ErrorInfoFromError extracts ErrorInfo from any error
func ErrorInfoFromError(err error) ErrorInfo {
	if err == nil {
		return ErrorInfo{}
	}

	if enhancedErr, ok := AsEnhancedError(err); ok {
		return enhancedErr.Info()
	}

	// Convert standard error
	return ErrorInfo{
		Message:   err.Error(),
		Level:     Error,
		Type:      Unknown,
		Timestamp: time.Now(),
		Metadata:  map[string]interface{}{"original_error_type": fmt.Sprintf("%T", err)},
	}
}

// Is checks if error matches specific criteria
func Is(err error, criteria func(ErrorInfo) bool) bool {
	if err == nil {
		return false
	}

	info := ErrorInfoFromError(err)
	return criteria(info)
}

// IsType checks if error is of specific type
func IsType(err error, errorType ErrorType) bool {
	return Is(err, func(info ErrorInfo) bool {
		return info.Type == errorType
	})
}

// IsLevel checks if error is of specific level
func IsLevel(err error, level ErrorLevel) bool {
	return Is(err, func(info ErrorInfo) bool {
		return info.Level == level
	})
}

// IsCode checks if error has specific code
func IsCode(err error, code string) bool {
	return Is(err, func(info ErrorInfo) bool {
		return info.Code == code
	})
}

// MultiError represents multiple errors
type MultiError struct {
	errors []error
	mu     sync.RWMutex
}

// NewMultiError creates a new MultiError
func NewMultiError(errs ...error) *MultiError {
	me := &MultiError{}
	for _, err := range errs {
		if err != nil {
			me.Add(err)
		}
	}
	return me
}

// Error returns concatenated error messages
func (m *MultiError) Error() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.errors) == 0 {
		return "<no errors>"
	}

	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("%d error(s) occurred:", len(m.errors)))

	for i, err := range m.errors {
		buf.WriteString(fmt.Sprintf("\n  [%d] %v", i+1, err))
	}

	return buf.String()
}

// Add adds an error to the collection
func (m *MultiError) Add(err error) {
	if err == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.errors = append(m.errors, err)
}

// Errors returns all errors
func (m *MultiError) Errors() []error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	errs := make([]error, len(m.errors))
	copy(errs, m.errors)
	return errs
}

// Count returns the number of errors
func (m *MultiError) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.errors)
}

// HasErrors checks if there are any errors
func (m *MultiError) HasErrors() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.errors) > 0
}

// Clear removes all errors
func (m *MultiError) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.errors = nil
}

// Filter filters errors by predicate
func (m *MultiError) Filter(predicate func(error) bool) *MultiError {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := NewMultiError()
	for _, err := range m.errors {
		if predicate(err) {
			result.Add(err)
		}
	}
	return result
}

// WrapAll wraps all errors with a message
func (m *MultiError) WrapAll(message string, opts ...ErrorOptions) *MultiError {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := NewMultiError()
	for _, err := range m.errors {
		result.Add(Wrap(err, message, opts...))
	}
	return result
}

// ErrorHandler provides configurable error handling
type ErrorHandler struct {
	Logger         *log.Logger
	LogLevel       ErrorLevel
	LogFormat      string // "text", "json"
	Metadata       map[string]interface{}
	OnError        func(error)
	OnPanic        func(error)
	StackDepth     int
	IncludeStack   bool
	CaptureContext bool
	mu             sync.RWMutex
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(opts ...ErrorHandlerOption) *ErrorHandler {
	handler := &ErrorHandler{
		Logger:       log.New(os.Stderr, "[ERROR] ", log.LstdFlags),
		LogLevel:     Error,
		LogFormat:    "text",
		Metadata:     make(map[string]interface{}),
		StackDepth:   10,
		IncludeStack: true,
	}

	for _, opt := range opts {
		opt(handler)
	}

	return handler
}

// ErrorHandlerOption configures ErrorHandler
type ErrorHandlerOption func(*ErrorHandler)

// WithLogger sets the logger
func WithLogger(logger *log.Logger) ErrorHandlerOption {
	return func(h *ErrorHandler) {
		h.Logger = logger
	}
}

// WithLogLevel sets the minimum log level
func WithLogLevel(level ErrorLevel) ErrorHandlerOption {
	return func(h *ErrorHandler) {
		h.LogLevel = level
	}
}

// WithLogFormat sets the log format
func WithLogFormat(format string) ErrorHandlerOption {
	return func(h *ErrorHandler) {
		h.LogFormat = format
	}
}

// WithMetadata adds global metadata
func WithMetadata(metadata map[string]interface{}) ErrorHandlerOption {
	return func(h *ErrorHandler) {
		h.mu.Lock()
		defer h.mu.Unlock()

		if h.Metadata == nil {
			h.Metadata = make(map[string]interface{})
		}
		for k, v := range metadata {
			h.Metadata[k] = v
		}
	}
}

// WithStackDepth sets stack trace depth
func WithStackDepth(depth int) ErrorHandlerOption {
	return func(h *ErrorHandler) {
		h.StackDepth = depth
	}
}

// WithIncludeStack enables/disables stack traces
func WithIncludeStack(include bool) ErrorHandlerOption {
	return func(h *ErrorHandler) {
		h.IncludeStack = include
	}
}

// WithOnError sets error callback
func WithOnError(callback func(error)) ErrorHandlerOption {
	return func(h *ErrorHandler) {
		h.OnError = callback
	}
}

// WithOnPanic sets panic callback
func WithOnPanic(callback func(error)) ErrorHandlerOption {
	return func(h *ErrorHandler) {
		h.OnPanic = callback
	}
}

// WithCaptureContext enables context capture
func WithCaptureContext(enable bool) ErrorHandlerOption {
	return func(h *ErrorHandler) {
		h.CaptureContext = enable
	}
}

// Handle processes an error
func (h *ErrorHandler) Handle(ctx context.Context, err error, opts ...ErrorOptions) error {
	if err == nil {
		return nil
	}

	// Convert to enhanced error if needed
	enhancedErr := ToEnhanced(err, opts...)

	// Add global metadata
	h.mu.RLock()
	if len(h.Metadata) > 0 {
		enhancedErr.WithMetadataMap(h.Metadata)
	}
	h.mu.RUnlock()

	// Add context metadata if enabled
	if h.CaptureContext && ctx != nil {
		if requestID := ctx.Value("request_id"); requestID != nil {
			enhancedErr.WithMetadata("request_id", requestID)
		}
		if userID := ctx.Value("user_id"); userID != nil {
			enhancedErr.WithMetadata("user_id", userID)
		}
	}

	// Check if error should be logged based on level
	info := enhancedErr.Info()
	if h.shouldLog(info.Level) {
		h.logError(enhancedErr)
	}

	// Call callback if set
	if h.OnError != nil {
		h.OnError(enhancedErr)
	}

	return enhancedErr
}

// HandleAndLog logs and returns the error
func (h *ErrorHandler) HandleAndLog(ctx context.Context, err error, opts ...ErrorOptions) error {
	return h.Handle(ctx, err, opts...)
}

// Recover handles panics
func (h *ErrorHandler) Recover(ctx context.Context) {
	if r := recover(); r != nil {
		var err error
		switch v := r.(type) {
		case error:
			err = v
		case string:
			err = errors.New(v)
		default:
			err = fmt.Errorf("panic: %v", v)
		}

		enhancedErr := h.Handle(ctx, err, ErrorOptions{
			Level: Panic,
			Type:  System,
		})

		// Call panic callback
		if h.OnPanic != nil {
			h.OnPanic(enhancedErr)
		}

		// Re-throw if panic level
		panic(enhancedErr)
	}
}

// RecoverWithCallback recovers with a custom function
func (h *ErrorHandler) RecoverWithCallback(ctx context.Context, callback func(error)) {
	if r := recover(); r != nil {
		var err error
		switch v := r.(type) {
		case error:
			err = v
		case string:
			err = errors.New(v)
		default:
			err = fmt.Errorf("panic: %v", v)
		}

		enhancedErr := h.Handle(ctx, err, ErrorOptions{
			Level: Panic,
			Type:  System,
		})

		if callback != nil {
			callback(enhancedErr)
		}

		if h.OnPanic != nil {
			h.OnPanic(enhancedErr)
		}
	}
}

func (h *ErrorHandler) shouldLog(level ErrorLevel) bool {
	levels := map[ErrorLevel]int{
		Debug: 1,
		Info:  2,
		Warn:  3,
		Error: 4,
		Fatal: 5,
		Panic: 6,
	}

	return levels[level] >= levels[h.LogLevel]
}

func (h *ErrorHandler) logError(err *EnhancedError) {
	if h.Logger == nil {
		return
	}

	switch h.LogFormat {
	case "json":
		if jsonData, jsonErr := err.JSON(); jsonErr == nil {
			h.Logger.Println(string(jsonData))
		} else {
			h.Logger.Printf("Failed to marshal error to JSON: %v", jsonErr)
		}
	default:
		h.Logger.Println(err.String())
	}
}

// Helper functions
func captureStack(skip int) []StackFrame {
	var stack []StackFrame

	// Get callers
	pc := make([]uintptr, 50)
	n := runtime.Callers(skip, pc)
	if n == 0 {
		return stack
	}

	pc = pc[:n]
	frames := runtime.CallersFrames(pc)

	for {
		frame, more := frames.Next()
		stack = append(stack, StackFrame{
			File:     frame.File,
			Line:     frame.Line,
			Function: frame.Function,
		})

		if !more {
			break
		}
	}

	return stack
}

func errorInfoFromError(err error) *ErrorInfo {
	if err == nil {
		return nil
	}

	if enhancedErr, ok := AsEnhancedError(err); ok {
		info := enhancedErr.Info()
		return &info
	}

	return &ErrorInfo{
		Message:   err.Error(),
		Level:     Error,
		Type:      Unknown,
		Timestamp: time.Now(),
	}
}

// FormatStack formats stack trace for display
func FormatStack(stack []StackFrame) string {
	var buf strings.Builder
	for i, frame := range stack {
		buf.WriteString(fmt.Sprintf("%d. %s\n   %s:%d\n", i+1, frame.Function, frame.File, frame.Line))
	}
	return buf.String()
}

// PrettyPrint prints error in a readable format
func PrettyPrint(err error, w io.Writer) {
	if err == nil {
		return
	}

	info := ErrorInfoFromError(err)
	fmt.Fprintf(w, "Error: %s\n", info.Message)

	if info.Code != "" {
		fmt.Fprintf(w, "Code: %s\n", info.Code)
	}

	if info.Type != Unknown {
		fmt.Fprintf(w, "Type: %s\n", info.Type)
	}

	if info.Level != "" {
		fmt.Fprintf(w, "Level: %s\n", info.Level)
	}

	if info.Timestamp != (time.Time{}) {
		fmt.Fprintf(w, "Time: %s\n", info.Timestamp.Format(time.RFC3339))
	}

	if len(info.Metadata) > 0 {
		fmt.Fprintf(w, "Metadata:\n")
		for k, v := range info.Metadata {
			fmt.Fprintf(w, "  %s: %v\n", k, v)
		}
	}

	if info.UserMessage != "" {
		fmt.Fprintf(w, "User Message: %s\n", info.UserMessage)
	}

	if len(info.Stack) > 0 {
		fmt.Fprintf(w, "Stack Trace:\n%s", FormatStack(info.Stack))
	}

	if info.Cause != nil {
		fmt.Fprintf(w, "\nCaused by:\n")
		PrettyPrint(&EnhancedError{info: info.Cause}, w)
	}
}

// Global error handler instance
var (
	defaultHandler *ErrorHandler
	handlerOnce    sync.Once
)

// GetDefaultHandler returns the global error handler
func GetDefaultHandler() *ErrorHandler {
	handlerOnce.Do(func() {
		defaultHandler = NewErrorHandler(
			WithLogLevel(Error),
			WithLogFormat("text"),
			WithIncludeStack(true),
			WithStackDepth(5),
		)
	})
	return defaultHandler
}

// SetDefaultHandler sets the global error handler
func SetDefaultHandler(handler *ErrorHandler) {
	defaultHandler = handler
}

// HandleError uses the global handler
func HandleError(ctx context.Context, err error, opts ...ErrorOptions) error {
	return GetDefaultHandler().Handle(ctx, err, opts...)
}

// Must panics if error is not nil
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// MustWithMessage panics with custom message if error is not nil
func MustWithMessage(err error, message string) {
	if err != nil {
		panic(fmt.Sprintf("%s: %v", message, err))
	}
}

// Must2 is like Must but for 2 return values
func Must2[T any](value T, err error) T {
	Must(err)
	return value
}

// Must3 is like Must but for 3 return values
func Must3[T1, T2 any](value1 T1, value2 T2, err error) (T1, T2) {
	Must(err)
	return value1, value2
}

// SafeExecute executes a function and recovers any panic
func SafeExecute(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				err = v
			case string:
				err = errors.New(v)
			default:
				err = fmt.Errorf("panic: %v", v)
			}
			err = Wrap(err, "panic in SafeExecute")
		}
	}()

	return fn()
}

// Retry retries a function on error
func Retry(attempts int, delay time.Duration, fn func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		err = fn()
		if err == nil {
			return nil
		}

		if i < attempts-1 {
			time.Sleep(delay)
			delay *= 2 // Exponential backoff
		}
	}

	return Wrapf(err, "failed after %d attempts", attempts)
}

// RetryWithContext retries with context support
func RetryWithContext(ctx context.Context, attempts int, delay time.Duration, fn func(context.Context) error) error {
	var err error
	for i := 0; i < attempts; i++ {
		select {
		case <-ctx.Done():
			return Wrap(ctx.Err(), "context cancelled during retry")
		default:
			err = fn(ctx)
			if err == nil {
				return nil
			}

			if i < attempts-1 {
				select {
				case <-time.After(delay):
					delay *= 2
				case <-ctx.Done():
					return Wrap(ctx.Err(), "context cancelled during retry delay")
				}
			}
		}
	}

	return Wrapf(err, "failed after %d attempts", attempts)
}
