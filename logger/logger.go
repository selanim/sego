package logger

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Level represents log level
type Level int

const (
	// DEBUG level for detailed debugging information
	DEBUG Level = iota
	// INFO level for general information
	INFO
	// WARN level for warnings
	WARN
	// ERROR level for errors
	ERROR
	// FATAL level for fatal errors (exits program)
	FATAL
	// PANIC level for panic errors (panics)
	PANIC
)

// String returns string representation of log level
func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	case PANIC:
		return "PANIC"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel parses log level from string
func ParseLevel(level string) Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	case "PANIC":
		return PANIC
	default:
		return INFO // Default to INFO
	}
}

// Config holds logger configuration
type Config struct {
	Level      Level
	Output     io.Writer
	Formatter  Formatter
	WithCaller bool
	WithTime   bool
	TimeFormat string
	Color      bool
	JSON       bool
}

// DefaultConfig returns default logger configuration
func DefaultConfig() Config {
	return Config{
		Level:      INFO,
		Output:     os.Stdout,
		Formatter:  &TextFormatter{},
		WithCaller: false,
		WithTime:   true,
		TimeFormat: time.RFC3339,
		Color:      true,
		JSON:       false,
	}
}

// Logger represents a logger instance
type Logger struct {
	config    Config
	mu        sync.Mutex
	fields    map[string]interface{}
	prefix    string
	output    io.Writer
	formatter Formatter
}

// Formatter defines interface for log formatting
type Formatter interface {
	Format(entry *Entry) ([]byte, error)
}

// Entry represents a log entry
type Entry struct {
	Time    time.Time
	Level   Level
	Message string
	Fields  map[string]interface{}
	Caller  *CallerInfo
	Prefix  string
}

// CallerInfo holds information about the caller
type CallerInfo struct {
	File     string
	Line     int
	Function string
}

// New creates a new logger with default configuration
func New() *Logger {
	config := DefaultConfig()
	return NewWithConfig(config)
}

// NewWithConfig creates a new logger with custom configuration
func NewWithConfig(config Config) *Logger {
	logger := &Logger{
		config: config,
		fields: make(map[string]interface{}),
		output: config.Output,
	}

	// Set formatter based on config
	if config.JSON {
		logger.formatter = &JSONFormatter{}
	} else {
		logger.formatter = &TextFormatter{
			Color:      config.Color,
			TimeFormat: config.TimeFormat,
		}
	}

	return logger
}

// WithField adds a field to the logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newLogger := l.clone()
	newLogger.fields[key] = value
	return newLogger
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newLogger := l.clone()
	for k, v := range fields {
		newLogger.fields[k] = v
	}
	return newLogger
}

// WithPrefix sets a prefix for the logger
func (l *Logger) WithPrefix(prefix string) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newLogger := l.clone()
	newLogger.prefix = prefix
	return newLogger
}

// clone creates a copy of the logger
func (l *Logger) clone() *Logger {
	fields := make(map[string]interface{})
	for k, v := range l.fields {
		fields[k] = v
	}

	return &Logger{
		config:    l.config,
		fields:    fields,
		prefix:    l.prefix,
		output:    l.output,
		formatter: l.formatter,
	}
}

// Log logs a message at specified level
func (l *Logger) Log(level Level, args ...interface{}) {
	if level < l.config.Level {
		return
	}

	l.log(level, fmt.Sprint(args...))
}

// Logf logs a formatted message at specified level
func (l *Logger) Logf(level Level, format string, args ...interface{}) {
	if level < l.config.Level {
		return
	}

	l.log(level, fmt.Sprintf(format, args...))
}

// Debug logs a debug message
func (l *Logger) Debug(args ...interface{}) {
	l.Log(DEBUG, args...)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Logf(DEBUG, format, args...)
}

// Info logs an info message
func (l *Logger) Info(args ...interface{}) {
	l.Log(INFO, args...)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.Logf(INFO, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(args ...interface{}) {
	l.Log(WARN, args...)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Logf(WARN, format, args...)
}

// Error logs an error message
func (l *Logger) Error(args ...interface{}) {
	l.Log(ERROR, args...)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Logf(ERROR, format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(args ...interface{}) {
	l.Log(FATAL, args...)
	os.Exit(1)
}

// Fatalf logs a formatted fatal message and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.Logf(FATAL, format, args...)
	os.Exit(1)
}

// Panic logs a panic message and panics
func (l *Logger) Panic(args ...interface{}) {
	l.Log(PANIC, args...)
	panic(fmt.Sprint(args...))
}

// Panicf logs a formatted panic message and panics
func (l *Logger) Panicf(format string, args ...interface{}) {
	l.Logf(PANIC, format, args...)
	panic(fmt.Sprintf(format, args...))
}

// log writes the log entry
func (l *Logger) log(level Level, msg string) {
	entry := &Entry{
		Time:    time.Now(),
		Level:   level,
		Message: msg,
		Fields:  l.fields,
		Prefix:  l.prefix,
	}

	// Add caller info if enabled
	if l.config.WithCaller {
		entry.Caller = getCaller()
	}

	// Format and write the entry
	formatted, err := l.formatter.Format(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to format log entry: %v\n", err)
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if _, err := l.output.Write(formatted); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write log: %v\n", err)
	}
}

// getCaller retrieves caller information
func getCaller() *CallerInfo {
	// Skip 4 callers: getCaller -> log -> Log/Logf -> actual caller
	pc, file, line, ok := runtime.Caller(4)
	if !ok {
		return nil
	}

	funcName := runtime.FuncForPC(pc).Name()
	// Simplify function name
	if idx := strings.LastIndex(funcName, "/"); idx != -1 {
		funcName = funcName[idx+1:]
	}

	return &CallerInfo{
		File:     file,
		Line:     line,
		Function: funcName,
	}
}

// SetLevel sets the log level
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.config.Level = level
}

// SetOutput sets the output writer
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = w
}

// SetFormatter sets the formatter
func (l *Logger) SetFormatter(f Formatter) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.formatter = f
}

// GetConfig returns the logger configuration
func (l *Logger) GetConfig() Config {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.config
}
