package timeutils

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestDefaultInstance(t *testing.T) {
	tu := Default()
	if tu == nil {
		t.Fatal("Default() should return non-nil instance")
	}
}

func TestNewWithLocation(t *testing.T) {
	tu, err := NewWithLocation("UTC")
	if err != nil {
		t.Fatalf("Failed to create TimeUtils with UTC: %v", err)
	}
	if tu == nil {
		t.Fatal("NewWithLocation should return non-nil instance")
	}
}

func TestParse(t *testing.T) {
	tu := Default()

	testCases := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{"RFC3339", "2023-12-25T15:04:05Z", "2023-12-25 15:04:05", false},
		{"DateTime", "2023-12-25 15:04:05", "2023-12-25 15:04:05", false},
		{"Date Only", "2023-12-25", "2023-12-25 00:00:00", false},
		{"Invalid", "invalid-date", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tu.Parse(tc.input)

			if tc.hasError {
				if err == nil {
					t.Errorf("Expected error for input: %s", tc.input)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for input %s: %v", tc.input, err)
				return
			}

			formatted := tu.Format(result, FormatOptions{Format: DateTimeFormat})
			if formatted != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, formatted)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	tu := Default()
	now := tu.Now()

	testCases := []struct {
		name     string
		format   TimeFormat
		expected string
	}{
		{"DateTime", DateTimeFormat, now.Format(string(DateTimeFormat))},
		{"Date", DateFormat, now.Format(string(DateFormat))},
		{"Time", TimeFormat24Hour, now.Format(string(TimeFormat24Hour))},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tu.Format(now, FormatOptions{Format: tc.format})
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestHumanize(t *testing.T) {
	tu := Default()
	now := tu.Now()

	testCases := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{"Just Now", now.Add(-10 * time.Second), "just now"},
		{"1 Minute", now.Add(-1 * time.Minute), "1 minute ago"},
		{"5 Minutes", now.Add(-5 * time.Minute), "5 minutes ago"},
		{"1 Hour", now.Add(-1 * time.Hour), "1 hour ago"},
		{"Yesterday", now.Add(-24 * time.Hour), "yesterday"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tu.Humanize(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestDateManipulation(t *testing.T) {
	tu := Default()
	testTime := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)

	t.Run("BeginningOfDay", func(t *testing.T) {
		result := tu.BeginningOfDay(testTime)
		expected := time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)
		if !result.Equal(expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("EndOfDay", func(t *testing.T) {
		result := tu.EndOfDay(testTime)
		expected := time.Date(2023, 12, 25, 23, 59, 59, 999999999, time.UTC)
		if !result.Equal(expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("BeginningOfWeek", func(t *testing.T) {
		result := tu.BeginningOfWeek(testTime)
		expected := time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC) // Monday
		if !result.Equal(expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("IsWeekend", func(t *testing.T) {
		weekday := time.Date(2023, 12, 23, 0, 0, 0, 0, time.UTC) // Saturday
		if !tu.IsWeekend(weekday) {
			t.Error("Saturday should be weekend")
		}

		monday := time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC) // Monday
		if tu.IsWeekend(monday) {
			t.Error("Monday should not be weekend")
		}
	})
}

func TestAddDuration(t *testing.T) {
	tu := Default()
	start := time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)

	testCases := []struct {
		name     string
		duration string
		expected time.Time
		hasError bool
	}{
		{"1 Hour", "1h", start.Add(time.Hour), false},
		{"2 Days", "2d", start.AddDate(0, 0, 2), false},
		{"1 Week", "1w", start.AddDate(0, 0, 7), false},
		{"1 Month", "1M", start.AddDate(0, 1, 0), false},
		{"Invalid", "invalid", time.Time{}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tu.AddDuration(start, tc.duration)

			if tc.hasError {
				if err == nil {
					t.Errorf("Expected error for duration: %s", tc.duration)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for duration %s: %v", tc.duration, err)
				return
			}

			if !result.Equal(tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestTimeRange(t *testing.T) {
	start := time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)
	end := time.Date(2023, 12, 26, 0, 0, 0, 0, time.UTC)

	tr := TimeRange{Start: start, End: end}

	if !tr.IsValid() {
		t.Error("TimeRange should be valid")
	}

	duration := tr.Duration()
	if duration != 24*time.Hour {
		t.Errorf("Expected 24h duration, got %v", duration)
	}

	middle := start.Add(12 * time.Hour)
	if !tr.Contains(middle) {
		t.Error("TimeRange should contain middle time")
	}

	before := start.Add(-1 * time.Hour)
	if tr.Contains(before) {
		t.Error("TimeRange should not contain time before start")
	}

	// Test overlapping ranges
	tr2 := TimeRange{
		Start: start.Add(12 * time.Hour),
		End:   end.Add(12 * time.Hour),
	}

	if !tr.Overlaps(tr2) {
		t.Error("Ranges should overlap")
	}

	nonOverlapping := TimeRange{
		Start: end.Add(time.Hour),
		End:   end.Add(2 * time.Hour),
	}

	if tr.Overlaps(nonOverlapping) {
		t.Error("Ranges should not overlap")
	}
}

func TestBusinessDaysBetween(t *testing.T) {
	tu := Default()

	// Monday to Friday (5 days)
	monday := time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC) // Monday
	friday := time.Date(2023, 12, 29, 0, 0, 0, 0, time.UTC) // Friday

	businessDays := tu.BusinessDaysBetween(monday, friday)
	if businessDays != 5 {
		t.Errorf("Expected 5 business days, got %d", businessDays)
	}

	// Friday to next Monday (1 business day)
	nextMonday := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	businessDays = tu.BusinessDaysBetween(friday, nextMonday)
	if businessDays != 1 {
		t.Errorf("Expected 1 business day, got %d", businessDays)
	}
}

func TestSleepWithContext(t *testing.T) {
	tu := Default()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := tu.SleepWithContext(ctx, 100*time.Millisecond)
	duration := time.Since(start)

	if err == nil {
		t.Error("Expected context cancellation error")
	}

	if duration > 60*time.Millisecond {
		t.Errorf("Sleep should have been cancelled, duration: %v", duration)
	}
}

func TestBatchTimes(t *testing.T) {
	tu := Default()

	start := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2023, 1, 1, 5, 0, 0, 0, time.UTC)

	batches := tu.BatchTimes(start, end, 2*time.Hour)

	if len(batches) != 3 {
		t.Errorf("Expected 3 batches, got %d", len(batches))
	}

	// Check first batch
	if !batches[0].Start.Equal(start) {
		t.Errorf("First batch should start at %v", start)
	}

	if !batches[0].End.Equal(start.Add(2 * time.Hour)) {
		t.Errorf("First batch should end at %v", start.Add(2*time.Hour))
	}

	// Check last batch
	lastBatch := batches[len(batches)-1]
	if !lastBatch.End.Equal(end) {
		t.Errorf("Last batch should end at %v", end)
	}
}

func TestIsLeapYear(t *testing.T) {
	tu := Default()

	testCases := []struct {
		year     int
		expected bool
	}{
		{2000, true},  // Divisible by 400
		{1900, false}, // Divisible by 100 but not 400
		{2020, true},  // Divisible by 4 but not 100
		{2023, false}, // Not divisible by 4
		{2024, true},  // Divisible by 4
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Year%d", tc.year), func(t *testing.T) {
			result := tu.IsLeapYear(tc.year)
			if result != tc.expected {
				t.Errorf("IsLeapYear(%d) = %v, expected %v", tc.year, result, tc.expected)
			}
		})
	}
}

func BenchmarkFormat(b *testing.B) {
	tu := Default()
	now := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tu.Format(now, FormatOptions{Format: DateTimeFormat})
	}
}

func BenchmarkParse(b *testing.B) {
	tu := Default()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tu.Parse("2023-12-25 15:04:05")
	}
}
