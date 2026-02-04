package pagination

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// Cursor represents cursor-based pagination
type Cursor struct {
	Value     interface{} `json:"value"`
	Direction string      `json:"direction"` // "next" or "prev"
	Timestamp time.Time   `json:"timestamp,omitempty"`
}

// CursorOptions represents cursor pagination options
type CursorOptions struct {
	Cursor    string `json:"cursor,omitempty"`
	Limit     int    `json:"limit"`
	Direction string `json:"direction,omitempty"` // "forward" or "backward"
}

// CursorResult represents cursor-based paginated result
type CursorResult[T any] struct {
	Data     []T    `json:"data"`
	Next     string `json:"next_cursor,omitempty"`
	Previous string `json:"previous_cursor,omitempty"`
	HasMore  bool   `json:"has_more"`
	Limit    int    `json:"limit"`
}

// DefaultCursorOptions returns default cursor options
func DefaultCursorOptions() *CursorOptions {
	return &CursorOptions{
		Limit:     20,
		Direction: "forward",
	}
}

// EncodeCursor encodes a cursor to base64 string
func EncodeCursor(cursor *Cursor) (string, error) {
	data, err := json.Marshal(cursor)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cursor: %w", err)
	}
	return base64.URLEncoding.EncodeToString(data), nil
}

// DecodeCursor decodes a base64 string to cursor
func DecodeCursor(cursorStr string) (*Cursor, error) {
	if cursorStr == "" {
		return nil, nil
	}

	data, err := base64.URLEncoding.DecodeString(cursorStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cursor: %w", err)
	}

	var cursor Cursor
	if err := json.Unmarshal(data, &cursor); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cursor: %w", err)
	}

	return &cursor, nil
}

// CreateNextCursor creates a next cursor from the last item
func CreateNextCursor[T any](items []T, getCursorValue func(T) interface{}) (*Cursor, error) {
	if len(items) == 0 {
		return nil, nil
	}

	lastItem := items[len(items)-1]
	cursor := &Cursor{
		Value:     getCursorValue(lastItem),
		Direction: "next",
		Timestamp: time.Now(),
	}

	return cursor, nil
}

// CreatePrevCursor creates a previous cursor from the first item
func CreatePrevCursor[T any](items []T, getCursorValue func(T) interface{}) (*Cursor, error) {
	if len(items) == 0 {
		return nil, nil
	}

	firstItem := items[0]
	cursor := &Cursor{
		Value:     getCursorValue(firstItem),
		Direction: "prev",
		Timestamp: time.Now(),
	}

	return cursor, nil
}

// ApplyCursorQuery applies cursor-based pagination to SQL query
func ApplyCursorQuery(field string, cursorValue interface{}, direction string, limit int) (string, []interface{}) {
	var query string
	var args []interface{}

	if cursorValue != nil {
		operator := ">"
		orderDirection := "ASC"

		if direction == "prev" {
			operator = "<"
			orderDirection = "DESC"
		}

		query = fmt.Sprintf("WHERE %s %s ? ORDER BY %s %s LIMIT ?",
			field, operator, field, orderDirection)
		args = []interface{}{cursorValue, limit}
	} else {
		query = fmt.Sprintf("ORDER BY %s ASC LIMIT ?", field)
		args = []interface{}{limit}
	}

	return query, args
}

// BuildCursorResult builds a cursor pagination result
func BuildCursorResult[T any](items []T, limit int, getCursorValue func(T) interface{}) (*CursorResult[T], error) {
	result := &CursorResult[T]{
		Data:  items,
		Limit: limit,
	}

	// Check if there are more items
	result.HasMore = len(items) == limit

	// Create cursors if there are items
	if len(items) > 0 {
		// Next cursor
		if result.HasMore {
			nextCursor, err := CreateNextCursor(items, getCursorValue)
			if err != nil {
				return nil, err
			}
			if nextCursor != nil {
				result.Next, err = EncodeCursor(nextCursor)
				if err != nil {
					return nil, err
				}
			}
		}

		// Previous cursor (for backward pagination)
		prevCursor, err := CreatePrevCursor(items, getCursorValue)
		if err != nil {
			return nil, err
		}
		if prevCursor != nil {
			result.Previous, err = EncodeCursor(prevCursor)
			if err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}
