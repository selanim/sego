package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// MultiWriter writes to multiple writers
type MultiWriter struct {
	writers []io.Writer
	mu      sync.Mutex
}

// NewMultiWriter creates a new MultiWriter
func NewMultiWriter(writers ...io.Writer) *MultiWriter {
	return &MultiWriter{writers: writers}
}

// Write writes to all writers
func (mw *MultiWriter) Write(p []byte) (n int, err error) {
	mw.mu.Lock()
	defer mw.mu.Unlock()

	for _, w := range mw.writers {
		n, err = w.Write(p)
		if err != nil {
			return n, err
		}
	}
	return len(p), nil
}

// AddWriter adds a writer to MultiWriter
func (mw *MultiWriter) AddWriter(w io.Writer) {
	mw.mu.Lock()
	defer mw.mu.Unlock()
	mw.writers = append(mw.writers, w)
}

// FileHandler handles file-based logging
type FileHandler struct {
	file        *os.File
	filename    string
	maxSize     int64 // in bytes
	maxFiles    int
	currentSize int64
	mu          sync.Mutex
}

// NewFileHandler creates a new file handler
func NewFileHandler(filename string, maxSize int64, maxFiles int) (*FileHandler, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &FileHandler{
		file:        file,
		filename:    filename,
		maxSize:     maxSize,
		maxFiles:    maxFiles,
		currentSize: info.Size(),
	}, nil
}

// Write writes data to file with rotation
func (h *FileHandler) Write(p []byte) (n int, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if we need to rotate
	if h.maxSize > 0 && h.currentSize+int64(len(p)) > h.maxSize {
		if err := h.rotate(); err != nil {
			return 0, err
		}
	}

	n, err = h.file.Write(p)
	if err == nil {
		h.currentSize += int64(n)
	}
	return n, err
}

// rotate rotates the log file
func (h *FileHandler) rotate() error {
	// Close current file
	if err := h.file.Close(); err != nil {
		return err
	}

	// Rotate old files
	for i := h.maxFiles - 1; i > 0; i-- {
		oldName := fmt.Sprintf("%s.%d", h.filename, i)
		newName := fmt.Sprintf("%s.%d", h.filename, i+1)

		if _, err := os.Stat(oldName); err == nil {
			os.Rename(oldName, newName)
		}
	}

	// Rename current file
	backupName := fmt.Sprintf("%s.1", h.filename)
	if err := os.Rename(h.filename, backupName); err != nil {
		return err
	}

	// Create new file
	file, err := os.OpenFile(h.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	h.file = file
	h.currentSize = 0
	return nil
}

// Close closes the file handler
func (h *FileHandler) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.file.Close()
}

// DailyFileHandler rotates logs daily
type DailyFileHandler struct {
	file       *os.File
	baseName   string
	currentDay int
	mu         sync.Mutex
}

// NewDailyFileHandler creates a new daily file handler
func NewDailyFileHandler(baseName string) (*DailyFileHandler, error) {
	today := time.Now().Day()
	filename := fmt.Sprintf("%s-%s.log", baseName, time.Now().Format("2006-01-02"))

	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &DailyFileHandler{
		file:       file,
		baseName:   baseName,
		currentDay: today,
	}, nil
}

// Write writes data to file with daily rotation
func (h *DailyFileHandler) Write(p []byte) (n int, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if day has changed
	today := time.Now().Day()
	if today != h.currentDay {
		// Close current file
		h.file.Close()

		// Open new file for today
		filename := fmt.Sprintf("%s-%s.log", h.baseName, time.Now().Format("2006-01-02"))
		file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return 0, err
		}

		h.file = file
		h.currentDay = today
	}

	return h.file.Write(p)
}

// Close closes the daily file handler
func (h *DailyFileHandler) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.file.Close()
}
