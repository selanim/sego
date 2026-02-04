package utils

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestPointer(t *testing.T) {
	str := "hello"
	ptr := Pointer(str)
	if *ptr != str {
		t.Errorf("Pointer() = %v, want %v", *ptr, str)
	}

	num := 42
	numPtr := Pointer(num)
	if *numPtr != num {
		t.Errorf("Pointer() = %v, want %v", *numPtr, num)
	}
}

func TestDeref(t *testing.T) {
	var ptr *string
	result := Deref(ptr, "default")
	if result != "default" {
		t.Errorf("Deref() = %v, want %v", result, "default")
	}

	value := "hello"
	ptr = &value
	result = Deref(ptr, "default")
	if result != "hello" {
		t.Errorf("Deref() = %v, want %v", result, "hello")
	}
}

func TestMap(t *testing.T) {
	numbers := []int{1, 2, 3, 4}
	doubled := Map(numbers, func(n int) int {
		return n * 2
	})

	expected := []int{2, 4, 6, 8}
	for i, v := range doubled {
		if v != expected[i] {
			t.Errorf("Map()[%d] = %v, want %v", i, v, expected[i])
		}
	}
}

func TestFilter(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	even := Filter(numbers, func(n int) bool {
		return n%2 == 0
	})

	expected := []int{2, 4}
	if len(even) != len(expected) {
		t.Errorf("Filter() length = %v, want %v", len(even), len(expected))
	}

	for i, v := range even {
		if v != expected[i] {
			t.Errorf("Filter()[%d] = %v, want %v", i, v, expected[i])
		}
	}
}

func TestContains(t *testing.T) {
	slice := []string{"apple", "banana", "cherry"}

	if !Contains(slice, "banana") {
		t.Error("Contains() should find 'banana'")
	}

	if Contains(slice, "orange") {
		t.Error("Contains() should not find 'orange'")
	}
}

func TestUnique(t *testing.T) {
	slice := []int{1, 2, 2, 3, 4, 4, 4, 5}
	unique := Unique(slice)

	expected := []int{1, 2, 3, 4, 5}
	if len(unique) != len(expected) {
		t.Errorf("Unique() length = %v, want %v", len(unique), len(expected))
	}

	for i, v := range unique {
		if v != expected[i] {
			t.Errorf("Unique()[%d] = %v, want %v", i, v, expected[i])
		}
	}
}

func TestChunk(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5, 6, 7}
	chunks := Chunk(slice, 3)

	if len(chunks) != 3 {
		t.Errorf("Chunk() created %v chunks, want 3", len(chunks))
	}

	expected := [][]int{{1, 2, 3}, {4, 5, 6}, {7}}
	for i, chunk := range chunks {
		for j, v := range chunk {
			if v != expected[i][j] {
				t.Errorf("Chunk()[%d][%d] = %v, want %v", i, j, v, expected[i][j])
			}
		}
	}
}

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_KEY", "test_value")
	defer os.Unsetenv("TEST_KEY")

	value := GetEnv("TEST_KEY", "default")
	if value != "test_value" {
		t.Errorf("GetEnv() = %v, want %v", value, "test_value")
	}

	value = GetEnv("NON_EXISTENT_KEY", "default")
	if value != "default" {
		t.Errorf("GetEnv() = %v, want %v", value, "default")
	}
}

func TestFileExists(t *testing.T) {
	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if !FileExists(tmpfile.Name()) {
		t.Error("FileExists() should return true for existing file")
	}

	if FileExists("/non/existent/file") {
		t.Error("FileExists() should return false for non-existent file")
	}
}

func TestHashMD5(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "5d41402abc4b2a76b9719d911017c592"},
		{"", "d41d8cd98f00b204e9800998ecf8427e"},
		{"test", "098f6bcd4621d373cade4e832627b4f6"},
	}

	for _, test := range tests {
		result := HashMD5(test.input)
		if result != test.expected {
			t.Errorf("HashMD5(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestGenerateRandomString(t *testing.T) {
	length := 10
	result := GenerateRandomString(length)

	if len(result) != length {
		t.Errorf("GenerateRandomString() length = %v, want %v", len(result), length)
	}

	// Generate another to ensure randomness
	another := GenerateRandomString(length)
	if result == another {
		t.Error("GenerateRandomString() should generate different strings")
	}
}

func TestReadWriteJSON(t *testing.T) {
	type TestData struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	data := TestData{
		Name:  "Test",
		Value: 42,
	}

	tmpfile, err := os.CreateTemp("", "test.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Write JSON
	err = WriteJSON(tmpfile.Name(), data, true)
	if err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}

	// Read JSON
	var readData TestData
	err = ReadJSON(tmpfile.Name(), &readData)
	if err != nil {
		t.Fatalf("ReadJSON failed: %v", err)
	}

	if readData.Name != data.Name || readData.Value != data.Value {
		t.Errorf("Read data = %+v, want %+v", readData, data)
	}
}

func TestRetry(t *testing.T) {
	attempts := 0
	err := Retry(3, time.Millisecond, func() error {
		attempts++
		if attempts < 3 {
			return fmt.Errorf("not yet")
		}
		return nil
	})

	if err != nil {
		t.Errorf("Retry() failed: %v", err)
	}

	if attempts != 3 {
		t.Errorf("Retry() made %v attempts, want 3", attempts)
	}
}

func TestIsZeroValue(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected bool
	}{
		{0, true},
		{42, false},
		{"", true},
		{"hello", false},
		{[]int{}, true},
		{[]int{1, 2}, false},
		{nil, true},
	}

	for _, test := range tests {
		result := IsZeroValue(test.value)
		if result != test.expected {
			t.Errorf("IsZeroValue(%v) = %v, want %v", test.value, result, test.expected)
		}
	}
}

func TestGetFileExtension(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"test.txt", "txt"},
		{"archive.tar.gz", "gz"},
		{"noextension", ""},
		{".hidden", ""},
	}

	for _, test := range tests {
		result := GetFileExtension(test.filename)
		if result != test.expected {
			t.Errorf("GetFileExtension(%q) = %q, want %q", test.filename, result, test.expected)
		}
	}
}

func TestConcurrentMap(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	results := ConcurrentMap(items, func(n int) int {
		time.Sleep(time.Millisecond)
		return n * 2
	}, 2)

	expected := []int{2, 4, 6, 8, 10}
	for i, v := range results {
		if v != expected[i] {
			t.Errorf("ConcurrentMap()[%d] = %v, want %v", i, v, expected[i])
		}
	}
}

func TestSafeClose(t *testing.T) {
	// Test with nil closer
	SafeClose(nil) // Should not panic

	// Test with actual closer
	tmpfile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// This should not panic even if we close multiple times
	SafeClose(tmpfile)
	SafeClose(tmpfile)
}
