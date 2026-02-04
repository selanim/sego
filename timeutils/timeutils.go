package timeutils

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

// TimeFormat defines common time formats
type TimeFormat string

const (
	// RFC formats
	RFC3339Format     TimeFormat = time.RFC3339
	RFC3339NanoFormat TimeFormat = time.RFC3339Nano
	RFC1123Format     TimeFormat = time.RFC1123
	RFC1123ZFormat    TimeFormat = time.RFC1123Z
	RFC822Format      TimeFormat = time.RFC822
	RFC822ZFormat     TimeFormat = time.RFC822Z
	RFC850Format      TimeFormat = time.RFC850

	// Common date time formats
	DateTimeFormat        TimeFormat = "2006-01-02 15:04:05"
	DateFormat            TimeFormat = "2006-01-02"
	TimeFormat12Hour      TimeFormat = "03:04:05 PM"
	TimeFormat24Hour      TimeFormat = "15:04:05"
	DateTimeShortFormat   TimeFormat = "2006-01-02 15:04"
	DateTimeCompactFormat TimeFormat = "20060102150405"
	DateCompactFormat     TimeFormat = "20060102"

	// ISO formats
	ISO8601Format      TimeFormat = "2006-01-02T15:04:05Z"
	ISO8601NanoFormat  TimeFormat = "2006-01-02T15:04:05.999999999Z"
	ISO8601LocalFormat TimeFormat = "2006-01-02T15:04:05"

	// Human readable formats
	HumanDateTimeFormat TimeFormat = "Jan 02, 2006 15:04:05"
	HumanDateFormat     TimeFormat = "Jan 02, 2006"
	HumanTimeFormat     TimeFormat = "15:04:05"

	// File system safe formats
	FileSafeDateTimeFormat TimeFormat = "2006-01-02_15-04-05"
	FileSafeDateFormat     TimeFormat = "2006-01-02"

	// Database formats
	MySQLDateTimeFormat  TimeFormat = "2006-01-02 15:04:05"
	MySQLDateFormat      TimeFormat = "2006-01-02"
	PostgreSQLTimeFormat TimeFormat = "2006-01-02 15:04:05.999999-07"
)

// TimeZone defines common timezones
type TimeZone string

const (
	UTCZone   TimeZone = "UTC"
	LocalZone TimeZone = "Local"
	ESTZone   TimeZone = "America/New_York"
	PSTZone   TimeZone = "America/Los_Angeles"
	CSTZone   TimeZone = "America/Chicago"
	GMTZone   TimeZone = "Europe/London"
	CETZone   TimeZone = "Europe/Paris"
	JSTZone   TimeZone = "Asia/Tokyo"
	ISTZone   TimeZone = "Asia/Kolkata"
	AESTZone  TimeZone = "Australia/Sydney"
)

// DurationUnit defines time duration units
type DurationUnit string

const (
	NanosecondUnit  DurationUnit = "ns"
	MicrosecondUnit DurationUnit = "µs"
	MillisecondUnit DurationUnit = "ms"
	SecondUnit      DurationUnit = "s"
	MinuteUnit      DurationUnit = "m"
	HourUnit        DurationUnit = "h"
	DayUnit         DurationUnit = "d"
	WeekUnit        DurationUnit = "w"
	MonthUnit       DurationUnit = "M"
	YearUnit        DurationUnit = "y"
)

// TimeRange represents a time range with start and end
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// IsValid checks if time range is valid (start <= end)
func (tr TimeRange) IsValid() bool {
	return !tr.Start.IsZero() && !tr.End.IsZero() && !tr.Start.After(tr.End)
}

// Duration returns duration of the time range
func (tr TimeRange) Duration() time.Duration {
	if !tr.IsValid() {
		return 0
	}
	return tr.End.Sub(tr.Start)
}

// Contains checks if a time is within the range
func (tr TimeRange) Contains(t time.Time) bool {
	if !tr.IsValid() {
		return false
	}
	return !t.Before(tr.Start) && !t.After(tr.End)
}

// Overlaps checks if two time ranges overlap
func (tr TimeRange) Overlaps(other TimeRange) bool {
	if !tr.IsValid() || !other.IsValid() {
		return false
	}
	return !tr.End.Before(other.Start) && !tr.Start.After(other.End)
}

// ParseOptions provides options for time parsing
type ParseOptions struct {
	Location *time.Location
	Format   TimeFormat
	Strict   bool
}

// FormatOptions provides options for time formatting
type FormatOptions struct {
	Location *time.Location
	Format   TimeFormat
	Humanize bool // Use human readable format like "just now", "2 minutes ago"
}

// TimeUtils provides time utility functions
type TimeUtils struct {
	location *time.Location
	mu       sync.RWMutex
}

var (
	globalInstance *TimeUtils
	once           sync.Once
	locationCache  = make(map[string]*time.Location)
	cacheMutex     sync.RWMutex
)

// Default returns the global TimeUtils instance
func Default() *TimeUtils {
	once.Do(func() {
		globalInstance = &TimeUtils{location: time.Local}
	})
	return globalInstance
}

// New creates a new TimeUtils instance with specified location
func New(loc *time.Location) *TimeUtils {
	if loc == nil {
		loc = time.Local
	}
	return &TimeUtils{location: loc}
}

// NewWithLocation creates TimeUtils with location name
func NewWithLocation(locationName string) (*TimeUtils, error) {
	loc, err := LoadLocation(locationName)
	if err != nil {
		return nil, err
	}
	return New(loc), nil
}

// LoadLocation loads timezone location with caching
func LoadLocation(name string) (*time.Location, error) {
	cacheMutex.RLock()
	if loc, ok := locationCache[name]; ok {
		cacheMutex.RUnlock()
		return loc, nil
	}
	cacheMutex.RUnlock()

	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil, fmt.Errorf("failed to load location %s: %w", name, err)
	}

	cacheMutex.Lock()
	locationCache[name] = loc
	cacheMutex.Unlock()

	return loc, nil
}

// SetLocation changes the timezone location
func (tu *TimeUtils) SetLocation(loc *time.Location) {
	tu.mu.Lock()
	defer tu.mu.Unlock()
	tu.location = loc
}

// Now returns current time in the configured timezone
func (tu *TimeUtils) Now() time.Time {
	tu.mu.RLock()
	defer tu.mu.RUnlock()
	return time.Now().In(tu.location)
}

// Parse parses time string with multiple format support
func (tu *TimeUtils) Parse(str string, opts ...ParseOptions) (time.Time, error) {
	var options ParseOptions
	if len(opts) > 0 {
		options = opts[0]
	}

	location := tu.location
	if options.Location != nil {
		location = options.Location
	}

	// Try specific format if provided
	if options.Format != "" {
		t, err := time.Parse(string(options.Format), str)
		if err == nil {
			return t.In(location), nil
		}
		if options.Strict {
			return time.Time{}, fmt.Errorf("failed to parse time with format %s: %w", options.Format, err)
		}
	}

	// Try common formats
	formats := []string{
		string(RFC3339Format),
		string(RFC3339NanoFormat),
		string(DateTimeFormat),
		string(DateFormat),
		string(ISO8601Format),
		string(MySQLDateTimeFormat),
		string(TimeFormat24Hour),
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05.999999",
		"2006-01-02 15:04:05.999999 -0700",
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
	}

	for _, format := range formats {
		t, err := time.Parse(format, str)
		if err == nil {
			return t.In(location), nil
		}
	}

	return time.Time{}, fmt.Errorf("failed to parse time string: %s", str)
}

// Format formats time according to specified format
func (tu *TimeUtils) Format(t time.Time, opts ...FormatOptions) string {
	var options FormatOptions
	if len(opts) > 0 {
		options = opts[0]
	}

	location := tu.location
	if options.Location != nil {
		location = options.Location
	}

	t = t.In(location)

	if options.Humanize {
		return tu.Humanize(t)
	}

	format := DateTimeFormat
	if options.Format != "" {
		format = options.Format
	}

	return t.Format(string(format))
}

// Humanize returns human readable time difference
// Humanize returns a human-readable time difference like
// "just now", "5 minutes ago", "yesterday", "2 months ago"
func (tu *TimeUtils) Humanize(t time.Time) string {
	now := tu.Now()

	// Future time
	if t.After(now) {
		return "in the future"
	}

	diff := now.Sub(t)

	seconds := int(diff.Seconds())
	minutes := int(diff.Minutes())
	hours := int(diff.Hours())
	days := hours / 24
	weeks := days / 7
	months := days / 30
	years := days / 365

	switch {
	case years > 0:
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)

	case months > 0:
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)

	case weeks > 0:
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)

	case days > 0:
		if days == 1 {
			return "yesterday"
		}
		return fmt.Sprintf("%d days ago", days)

	case hours > 0:
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)

	case minutes > 0:
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)

	case seconds >= 10:
		return fmt.Sprintf("%d seconds ago", seconds)

	default:
		return "just now"
	}
}

// Date manipulation functions

// BeginningOfDay returns start of the day (00:00:00)
func (tu *TimeUtils) BeginningOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay returns end of the day (23:59:59.999999999)
func (tu *TimeUtils) EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// BeginningOfWeek returns start of week (Monday)
func (tu *TimeUtils) BeginningOfWeek(t time.Time) time.Time {
	weekday := t.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	daysToSubtract := int(weekday - time.Monday)
	return tu.BeginningOfDay(t.AddDate(0, 0, -daysToSubtract))
}

// EndOfWeek returns end of week (Sunday)
func (tu *TimeUtils) EndOfWeek(t time.Time) time.Time {
	beginning := tu.BeginningOfWeek(t)
	return tu.EndOfDay(beginning.AddDate(0, 0, 6))
}

// BeginningOfMonth returns start of month
func (tu *TimeUtils) BeginningOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth returns end of month
func (tu *TimeUtils) EndOfMonth(t time.Time) time.Time {
	firstDay := tu.BeginningOfMonth(t)
	lastDay := firstDay.AddDate(0, 1, -1)
	return tu.EndOfDay(lastDay)
}

// BeginningOfYear returns start of year
func (tu *TimeUtils) BeginningOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

// EndOfYear returns end of year
func (tu *TimeUtils) EndOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 12, 31, 23, 59, 59, 999999999, t.Location())
}

// IsToday checks if date is today
func (tu *TimeUtils) IsToday(t time.Time) bool {
	now := tu.Now()
	return t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day()
}

// IsWeekend checks if date is weekend
func (tu *TimeUtils) IsWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// AddDuration adds duration with string support (e.g., "1h30m", "2d", "1w")
func (tu *TimeUtils) AddDuration(t time.Time, duration string) (time.Time, error) {
	// Try standard duration parsing first
	d, err := time.ParseDuration(duration)
	if err == nil {
		return t.Add(d), nil
	}

	// Try custom units (weeks, months, years)
	value, unit, err := parseDurationString(duration)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid duration format: %s", duration)
	}

	switch unit {
	case "w", "week", "weeks":
		return t.AddDate(0, 0, value*7), nil
	case "M", "month", "months":
		return t.AddDate(0, value, 0), nil
	case "y", "year", "years":
		return t.AddDate(value, 0, 0), nil
	default:
		return time.Time{}, fmt.Errorf("unsupported duration unit: %s", unit)
	}
}

// parseDurationString parses duration strings with custom units
func parseDurationString(s string) (int, string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, "", errors.New("empty duration string")
	}

	// Find first non-digit character
	i := 0
	for i < len(s) && (s[i] >= '0' && s[i] <= '9') {
		i++
	}

	if i == 0 {
		return 0, "", errors.New("no numeric value found")
	}

	value, err := strconv.Atoi(s[:i])
	if err != nil {
		return 0, "", fmt.Errorf("invalid numeric value: %w", err)
	}

	unit := strings.ToLower(strings.TrimSpace(s[i:]))
	if unit == "" {
		return 0, "", errors.New("no unit specified")
	}

	return value, unit, nil
}

// Difference calculates difference between two times in specified unit
func (tu *TimeUtils) Difference(t1, t2 time.Time, unit DurationUnit) (float64, error) {
	duration := t2.Sub(t1)

	switch unit {
	case NanosecondUnit:
		return float64(duration.Nanoseconds()), nil
	case MicrosecondUnit:
		return duration.Seconds() * 1e6, nil
	case MillisecondUnit:
		return duration.Seconds() * 1e3, nil
	case SecondUnit:
		return duration.Seconds(), nil
	case MinuteUnit:
		return duration.Minutes(), nil
	case HourUnit:
		return duration.Hours(), nil
	case DayUnit:
		return duration.Hours() / 24, nil
	case WeekUnit:
		return duration.Hours() / (24 * 7), nil
	default:
		return 0, fmt.Errorf("unsupported duration unit: %s", unit)
	}
}

// BusinessDaysBetween calculates number of business days between two dates
func (tu *TimeUtils) BusinessDaysBetween(start, end time.Time) int {
	if start.After(end) {
		start, end = end, start
	}

	days := 0
	current := tu.BeginningOfDay(start)
	endDay := tu.BeginningOfDay(end)

	for !current.After(endDay) {
		if !tu.IsWeekend(current) {
			days++
		}
		current = current.AddDate(0, 0, 1)
	}

	return days
}

// IsLeapYear checks if year is leap year
func (tu *TimeUtils) IsLeapYear(year int) bool {
	return (year%4 == 0 && year%100 != 0) || (year%400 == 0)
}

// DaysInMonth returns number of days in month
func (tu *TimeUtils) DaysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// Time conversion functions

// ToUnix converts time to Unix timestamp
func (tu *TimeUtils) ToUnix(t time.Time) int64 {
	return t.Unix()
}

// FromUnix converts Unix timestamp to time
func (tu *TimeUtils) FromUnix(timestamp int64) time.Time {
	return time.Unix(timestamp, 0).In(tu.location)
}

// ToUnixMilli converts time to Unix timestamp in milliseconds
func (tu *TimeUtils) ToUnixMilli(t time.Time) int64 {
	return t.UnixMilli()
}

// FromUnixMilli converts Unix timestamp in milliseconds to time
func (tu *TimeUtils) FromUnixMilli(timestamp int64) time.Time {
	return time.UnixMilli(timestamp).In(tu.location)
}

// ToUnixMicro converts time to Unix timestamp in microseconds
func (tu *TimeUtils) ToUnixMicro(t time.Time) int64 {
	return t.UnixMicro()
}

// FromUnixMicro converts Unix timestamp in microseconds to time
func (tu *TimeUtils) FromUnixMicro(timestamp int64) time.Time {
	return time.UnixMicro(timestamp).In(tu.location)
}

// Timezone conversion

// ConvertTimezone converts time from one timezone to another
func (tu *TimeUtils) ConvertTimezone(t time.Time, from, to string) (time.Time, error) {
	fromLoc, err := LoadLocation(from)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to load source timezone: %w", err)
	}

	toLoc, err := LoadLocation(to)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to load target timezone: %w", err)
	}

	return t.In(fromLoc).In(toLoc), nil
}

// Schedule utilities

// NextOccurrence calculates next occurrence of a scheduled time
func (tu *TimeUtils) NextOccurrence(scheduleTime time.Time, interval time.Duration) time.Time {
	now := tu.Now()

	if scheduleTime.After(now) {
		return scheduleTime
	}

	// Calculate number of intervals passed
	diff := now.Sub(scheduleTime)
	intervals := int(math.Ceil(diff.Seconds() / interval.Seconds()))

	return scheduleTime.Add(time.Duration(intervals) * interval)
}

// Cron utilities

// ParseCron parses cron expression (simplified version)
func (tu *TimeUtils) ParseCron(expr string) (minute, hour, day, month, weekday string, err error) {
	parts := strings.Fields(expr)
	if len(parts) != 5 {
		return "", "", "", "", "", errors.New("cron expression must have 5 fields")
	}
	return parts[0], parts[1], parts[2], parts[3], parts[4], nil
}

// NextCronTime calculates next time based on cron expression
func (tu *TimeUtils) NextCronTime(expr string, fromTime ...time.Time) (time.Time, error) {
	from := tu.Now()
	if len(fromTime) > 0 {
		from = fromTime[0]
	}

	// Simple implementation - in production, use a proper cron parser library
	// This is a placeholder implementation
	minute, hour, day, month, weekday, err := tu.ParseCron(expr)
	if err != nil {
		return time.Time{}, err
	}

	// Basic validation
	if minute == "*" && hour == "*" && day == "*" && month == "*" && weekday == "*" {
		return from.Add(time.Minute), nil
	}

	// Simplified next minute calculation
	return from.Add(time.Minute), nil
}

// Utility functions

// SleepWithContext sleeps for duration but can be cancelled by context
func (tu *TimeUtils) SleepWithContext(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// RetryWithBackoff retries a function with exponential backoff
func (tu *TimeUtils) RetryWithBackoff(ctx context.Context, maxRetries int, initialDelay time.Duration, fn func() error) error {
	var err error
	delay := initialDelay

	for i := 0; i < maxRetries; i++ {
		err = fn()
		if err == nil {
			return nil
		}

		if i == maxRetries-1 {
			break
		}

		select {
		case <-time.After(delay):
			delay *= 2 // Exponential backoff
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		}
	}

	return fmt.Errorf("failed after %d retries: %w", maxRetries, err)
}

// BatchTimes splits time range into batches
func (tu *TimeUtils) BatchTimes(start, end time.Time, batchSize time.Duration) []TimeRange {
	if start.After(end) || batchSize <= 0 {
		return nil
	}

	var batches []TimeRange
	current := start

	for current.Before(end) {
		batchEnd := current.Add(batchSize)
		if batchEnd.After(end) {
			batchEnd = end
		}

		batches = append(batches, TimeRange{
			Start: current,
			End:   batchEnd,
		})

		current = batchEnd
		if !current.Before(end) {
			break
		}
	}

	return batches
}
