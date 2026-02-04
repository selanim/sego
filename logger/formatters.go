package logger

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// TextFormatter formats logs as plain text
type TextFormatter struct {
	Color      bool
	TimeFormat string
}

// Format formats a log entry as text
func (f *TextFormatter) Format(entry *Entry) ([]byte, error) {
	var buf strings.Builder

	// Add time if enabled
	if f.TimeFormat != "" {
		buf.WriteString(entry.Time.Format(f.TimeFormat))
		buf.WriteString(" ")
	}

	// Add log level with color
	levelStr := entry.Level.String()
	if f.Color {
		levelStr = f.colorizeLevel(levelStr, entry.Level)
	}
	buf.WriteString(fmt.Sprintf("[%s]", levelStr))
	buf.WriteString(" ")

	// Add prefix if exists
	if entry.Prefix != "" {
		buf.WriteString(fmt.Sprintf("[%s] ", entry.Prefix))
	}

	// Add message
	buf.WriteString(entry.Message)

	// Add fields if any
	if len(entry.Fields) > 0 {
		buf.WriteString(" | ")
		first := true
		for k, v := range entry.Fields {
			if !first {
				buf.WriteString(", ")
			}
			buf.WriteString(fmt.Sprintf("%s=%v", k, v))
			first = false
		}
	}

	// Add caller info if available
	if entry.Caller != nil {
		buf.WriteString(fmt.Sprintf(" | %s:%d", entry.Caller.File, entry.Caller.Line))
	}

	buf.WriteString("\n")
	return []byte(buf.String()), nil
}

// colorizeLevel adds ANSI color codes to log level
func (f *TextFormatter) colorizeLevel(level string, lvl Level) string {
	switch lvl {
	case DEBUG:
		return fmt.Sprintf("\033[36m%s\033[0m", level) // Cyan
	case INFO:
		return fmt.Sprintf("\033[32m%s\033[0m", level) // Green
	case WARN:
		return fmt.Sprintf("\033[33m%s\033[0m", level) // Yellow
	case ERROR:
		return fmt.Sprintf("\033[31m%s\033[0m", level) // Red
	case FATAL, PANIC:
		return fmt.Sprintf("\033[35m%s\033[0m", level) // Magenta
	default:
		return level
	}
}

// JSONFormatter formats logs as JSON
type JSONFormatter struct{}

// Format formats a log entry as JSON
func (f *JSONFormatter) Format(entry *Entry) ([]byte, error) {
	logData := map[string]interface{}{
		"timestamp": entry.Time.Format(time.RFC3339),
		"level":     entry.Level.String(),
		"message":   entry.Message,
	}

	// Add prefix if exists
	if entry.Prefix != "" {
		logData["prefix"] = entry.Prefix
	}

	// Add fields
	for k, v := range entry.Fields {
		logData[k] = v
	}

	// Add caller info if available
	if entry.Caller != nil {
		logData["caller"] = map[string]interface{}{
			"file":     entry.Caller.File,
			"line":     entry.Caller.Line,
			"function": entry.Caller.Function,
		}
	}

	return json.Marshal(logData)
}
