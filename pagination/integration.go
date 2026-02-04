package pagination

import (
	"context"
	"fmt"
	"net/url"
)

// Repository is a simplified interface for repositories
type Repository interface {
	Count(ctx context.Context, conditions ...map[string]interface{}) (int64, error)
	FindAll(ctx context.Context, query string, args ...interface{}) ([]interface{}, error)
}

// Paginate performs pagination on a repository
func Paginate[T any](ctx context.Context, repo Repository, opts *Options,
	baseQuery string, queryArgs []interface{}, mapper func(interface{}) T) (*PaginatedResult[T], error) {

	if opts == nil {
		opts = DefaultOptions()
	}

	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid pagination options: %w", err)
	}

	// Count total rows
	totalRows, err := repo.Count(ctx, map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("failed to count rows: %w", err)
	}

	// Calculate pagination
	pagination := opts.Calculate(totalRows)

	// Apply pagination to query
	paginationClause, paginationArgs := opts.Apply()
	fullQuery := fmt.Sprintf("%s %s", baseQuery, paginationClause)

	// Combine query arguments
	allArgs := append(queryArgs, paginationArgs...)

	// Execute query
	rows, err := repo.FindAll(ctx, fullQuery, allArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute paginated query: %w", err)
	}

	// Map results to type T
	var data []T
	for _, row := range rows {
		data = append(data, mapper(row))
	}

	return NewPaginatedResult(data, pagination), nil
}

// PaginateWithURL performs pagination with URL generation
func PaginateWithURL[T any](ctx context.Context, repo Repository, opts *Options,
	baseQuery string, queryArgs []interface{}, mapper func(interface{}) T,
	baseURL string, queryParams url.Values) (*PaginatedResult[T], error) {

	result, err := Paginate(ctx, repo, opts, baseQuery, queryArgs, mapper)
	if err != nil {
		return nil, err
	}

	// Add navigation links
	if result.Pagination != nil {
		result.Links = GenerateLinks(result.Pagination, baseURL, queryParams)
	}

	return result, nil
}

// PaginateCursor performs cursor-based pagination
func PaginateCursor[T any](ctx context.Context, repo Repository, opts *CursorOptions,
	field string, mapper func(interface{}) T, getCursorValue func(T) interface{}) (*CursorResult[T], error) {

	if opts == nil {
		opts = DefaultCursorOptions()
	}

	// Decode cursor
	var cursorValue interface{}
	if opts.Cursor != "" {
		cursor, err := DecodeCursor(opts.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}
		if cursor != nil {
			cursorValue = cursor.Value
			if opts.Direction == "" {
				opts.Direction = cursor.Direction
			}
		}
	}

	// Build query
	query, args := ApplyCursorQuery(field, cursorValue, opts.Direction, opts.Limit)

	// Execute query
	rows, err := repo.FindAll(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute cursor query: %w", err)
	}

	// Map results
	var data []T
	for _, row := range rows {
		data = append(data, mapper(row))
	}

	// Build cursor result
	return BuildCursorResult(data, opts.Limit, getCursorValue)
}

// PaginationMiddleware provides middleware for pagination
type PaginationMiddleware struct {
	DefaultLimit int
	MaxLimit     int
	AllowedSorts map[string]bool
	DefaultSort  string
	DefaultOrder string
}

// NewPaginationMiddleware creates a new pagination middleware
func NewPaginationMiddleware() *PaginationMiddleware {
	return &PaginationMiddleware{
		DefaultLimit: 20,
		MaxLimit:     100,
		AllowedSorts: map[string]bool{
			"id":         true,
			"created_at": true,
			"updated_at": true,
		},
		DefaultSort:  "created_at",
		DefaultOrder: "desc",
	}
}

// ParseOptions parses and validates pagination options from query parameters
func (m *PaginationMiddleware) ParseOptions(queryParams url.Values) (*Options, error) {
	opts := ParseFromURL(queryParams)

	// Apply defaults and limits
	if opts.Limit <= 0 {
		opts.Limit = m.DefaultLimit
	}
	if opts.Limit > m.MaxLimit {
		opts.Limit = m.MaxLimit
	}

	// Validate sort field
	if opts.SortBy != "" {
		if !m.AllowedSorts[opts.SortBy] {
			opts.SortBy = m.DefaultSort
		}
	} else {
		opts.SortBy = m.DefaultSort
	}

	// Validate sort direction
	if opts.SortDirection == "" {
		opts.SortDirection = m.DefaultOrder
	}

	if err := opts.Validate(); err != nil {
		return nil, err
	}

	return opts, nil
}

// PaginateQuery builds paginated SQL query
func (m *PaginationMiddleware) PaginateQuery(baseQuery string, opts *Options) (string, []interface{}) {
	if opts == nil {
		opts = DefaultOptions()
	}

	clause, args := opts.Apply()
	return fmt.Sprintf("%s %s", baseQuery, clause), args
}
