package utils

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"
)

// Pointer returns a pointer to the given value
func Pointer[T any](value T) *T {
	return &value
}

// Deref safely dereferences a pointer, returning default value if nil
func Deref[T any](ptr *T, defaultValue T) T {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

// Map applies a function to each element of a slice
func Map[T any, U any](slice []T, fn func(T) U) []U {
	result := make([]U, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}

// Filter filters a slice based on a predicate function
func Filter[T any](slice []T, fn func(T) bool) []T {
	var result []T
	for _, v := range slice {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

// Reduce reduces a slice to a single value
func Reduce[T any, U any](slice []T, initial U, fn func(U, T) U) U {
	result := initial
	for _, v := range slice {
		result = fn(result, v)
	}
	return result
}

// Contains checks if a slice contains a value
func Contains[T comparable](slice []T, value T) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// Unique returns a slice with only unique values
func Unique[T comparable](slice []T) []T {
	seen := make(map[T]bool)
	var result []T
	for _, v := range slice {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

// Chunk splits a slice into chunks of specified size
func Chunk[T any](slice []T, size int) [][]T {
	var chunks [][]T
	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

// MergeMaps merges multiple maps into one
func MergeMaps[K comparable, V any](maps ...map[K]V) map[K]V {
	result := make(map[K]V)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// GetEnv gets an environment variable or returns a default value
func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// GetEnvRequired gets an environment variable or panics if not found
func GetEnvRequired(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	panic(fmt.Sprintf("Environment variable %s is required", key))
}

// FileExists checks if a file exists
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// DirExists checks if a directory exists
func DirExists(dirname string) bool {
	info, err := os.Stat(dirname)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// CreateDir creates a directory if it doesn't exist
func CreateDir(dirname string) error {
	if !DirExists(dirname) {
		return os.MkdirAll(dirname, 0755)
	}
	return nil
}

// ReadJSON reads and unmarshals a JSON file
func ReadJSON(filename string, v interface{}) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// WriteJSON marshals and writes data to a JSON file
func WriteJSON(filename string, v interface{}, indent bool) error {
	var data []byte
	var err error

	if indent {
		data, err = json.MarshalIndent(v, "", "  ")
	} else {
		data, err = json.Marshal(v)
	}

	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// HashMD5 calculates MD5 hash of a string
func HashMD5(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// HashSHA256 calculates SHA256 hash of a string
func HashSHA256(text string) string {
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:])
}

// GenerateRandomString generates a random string of specified length
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	rand.Read(b)
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b)
}

// Retry executes a function with retry logic
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
	return fmt.Errorf("after %d attempts, last error: %v", attempts, err)
}

// DownloadFile downloads a file from a URL
func DownloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// GetFunctionName returns the name of the calling function
func GetFunctionName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}

// MeasureExecutionTime measures the execution time of a function
func MeasureExecutionTime(fn func()) time.Duration {
	start := time.Now()
	fn()
	return time.Since(start)
}

// IsZeroValue checks if a value is the zero value for its type
func IsZeroValue(v interface{}) bool {
	return reflect.DeepEqual(v, reflect.Zero(reflect.TypeOf(v)).Interface())
}

// SafeClose safely closes a Closer and logs any error
func SafeClose(closer io.Closer) {
	if closer != nil {
		if err := closer.Close(); err != nil {
			fmt.Printf("Error closing: %v\n", err)
		}
	}
}

// ParseDurationSafe safely parses a duration string
func ParseDurationSafe(durationStr string, defaultDuration time.Duration) time.Duration {
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return defaultDuration
	}
	return duration
}

// GetFileExtension returns the file extension (without dot)
func GetFileExtension(filename string) string {
	return strings.TrimPrefix(filepath.Ext(filename), ".")
}

// GetFileSize returns the size of a file in bytes
func GetFileSize(filename string) (int64, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// ConcurrentMap applies a function to each element concurrently
func ConcurrentMap[T any, U any](items []T, fn func(T) U, maxWorkers int) []U {
	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU()
	}

	type result struct {
		index int
		value U
	}

	results := make([]U, len(items))
	ch := make(chan result, len(items))
	sem := make(chan struct{}, maxWorkers)

	for i, item := range items {
		go func(idx int, item T) {
			sem <- struct{}{}
			defer func() { <-sem }()
			ch <- result{index: idx, value: fn(item)}
		}(i, item)
	}

	for i := 0; i < len(items); i++ {
		res := <-ch
		results[res.index] = res.value
	}

	return results
}
