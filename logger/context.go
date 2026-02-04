package logger

import (
	"context"
	"fmt"
	"sync"
)

// contextKey is a custom type for context keys
type contextKey string

const (
	// LoggerKey is the key for storing logger in context
	LoggerKey contextKey = "logger"
	// FieldsKey is the key for storing fields in context
	FieldsKey contextKey = "logger_fields"
)

// FromContext retrieves logger from context
func FromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(LoggerKey).(*Logger); ok {
		return logger
	}
	return New() // Return default logger if not found
}

// WithLogger adds logger to context
func WithLogger(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, LoggerKey, logger)
}

// WithField adds a field to context logger
func WithField(ctx context.Context, key string, value interface{}) context.Context {
	fields := getFieldsFromContext(ctx)
	fields[key] = value
	return context.WithValue(ctx, FieldsKey, fields)
}

// WithFields adds multiple fields to context logger
func WithFields(ctx context.Context, fields map[string]interface{}) context.Context {
	existingFields := getFieldsFromContext(ctx)

	// Merge fields
	for k, v := range fields {
		existingFields[k] = v
	}

	return context.WithValue(ctx, FieldsKey, existingFields)
}

// getFieldsFromContext retrieves fields from context
func getFieldsFromContext(ctx context.Context) map[string]interface{} {
	if fields, ok := ctx.Value(FieldsKey).(map[string]interface{}); ok {
		// Return a copy to avoid mutation
		copied := make(map[string]interface{})
		for k, v := range fields {
			copied[k] = v
		}
		return copied
	}
	return make(map[string]interface{})
}

// ContextLogger creates a logger with context fields
func ContextLogger(ctx context.Context) *Logger {
	logger := FromContext(ctx)
	fields := getFieldsFromContext(ctx)

	if len(fields) > 0 {
		return logger.WithFields(fields)
	}
	return logger
}

// CtxDebug logs debug message with context
func CtxDebug(ctx context.Context, args ...interface{}) {
	ContextLogger(ctx).Debug(args...)
}

// CtxDebugf logs formatted debug message with context
func CtxDebugf(ctx context.Context, format string, args ...interface{}) {
	ContextLogger(ctx).Debugf(format, args...)
}

// CtxInfo logs info message with context
func CtxInfo(ctx context.Context, args ...interface{}) {
	ContextLogger(ctx).Info(args...)
}

// CtxInfof logs formatted info message with context
func CtxInfof(ctx context.Context, format string, args ...interface{}) {
	ContextLogger(ctx).Infof(format, args...)
}

// CtxWarn logs warning message with context
func CtxWarn(ctx context.Context, args ...interface{}) {
	ContextLogger(ctx).Warn(args...)
}

// CtxWarnf logs formatted warning message with context
func CtxWarnf(ctx context.Context, format string, args ...interface{}) {
	ContextLogger(ctx).Warnf(format, args...)
}

// CtxError logs error message with context
func CtxError(ctx context.Context, args ...interface{}) {
	ContextLogger(ctx).Error(args...)
}

// CtxErrorf logs formatted error message with context
func CtxErrorf(ctx context.Context, format string, args ...interface{}) {
	ContextLogger(ctx).Errorf(format, args...)
}

// RequestLogger creates a request-scoped logger
func RequestLogger(ctx context.Context, requestID string) context.Context {
	logger := New().WithField("request_id", requestID)
	return WithLogger(ctx, logger)
}

// AsyncLogger provides asynchronous logging
type AsyncLogger struct {
	logger *Logger
	queue  chan *logEntry
	wg     *sync.WaitGroup
}

func (al *AsyncLogger) Infof(s string, i int) {
	panic("unimplemented")
}

type logEntry struct {
	level   Level
	message string
	fields  map[string]interface{}
}

// NewAsyncLogger creates a new asynchronous logger
func NewAsyncLogger(logger *Logger, bufferSize int) *AsyncLogger {
	asyncLogger := &AsyncLogger{
		logger: logger,
		queue:  make(chan *logEntry, bufferSize),
	}

	// Start worker
	asyncLogger.wg.Add(1)
	go asyncLogger.worker()

	return asyncLogger
}

// worker processes log entries asynchronously
func (al *AsyncLogger) worker() {
	defer al.wg.Done()

	for entry := range al.queue {
		logger := al.logger
		if len(entry.fields) > 0 {
			logger = logger.WithFields(entry.fields)
		}
		logger.log(entry.level, entry.message)
	}
}

// Log logs asynchronously
func (al *AsyncLogger) Log(level Level, args ...interface{}) {
	al.queue <- &logEntry{
		level:   level,
		message: fmt.Sprint(args...),
		fields:  make(map[string]interface{}),
	}
}

// Logf logs formatted message asynchronously
func (al *AsyncLogger) Logf(level Level, format string, args ...interface{}) {
	al.queue <- &logEntry{
		level:   level,
		message: fmt.Sprintf(format, args...),
		fields:  make(map[string]interface{}),
	}
}

// WithField adds a field for async logging
func (al *AsyncLogger) WithField(key string, value interface{}) *AsyncLogger {
	// This creates a wrapper that adds fields to each log entry
	return &AsyncLogger{
		logger: al.logger.WithField(key, value),
		queue:  al.queue,
		wg:     al.wg,
	}
}

// Shutdown waits for all logs to be processed
func (al *AsyncLogger) Shutdown() {
	close(al.queue)
	al.wg.Wait()
}
