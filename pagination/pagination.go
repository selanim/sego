package pagination

import (
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
)

// Pagination represents pagination metadata
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalRows  int64 `json:"total_rows"`
	TotalPages int   `json:"total_pages"`
	Offset     int   `json:"-"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
	NextPage   int   `json:"next_page,omitempty"`
	PrevPage   int   `json:"prev_page,omitempty"`
	FirstPage  int   `json:"first_page"`
	LastPage   int   `json:"last_page"`
}

// PaginatedResult represents a paginated query result
type PaginatedResult[T any] struct {
	Data       []T         `json:"data"`
	Pagination *Pagination `json:"pagination"`
	Links      *Links      `json:"links,omitempty"`
}

// Links contains navigation links for pagination
type Links struct {
	First    string `json:"first,omitempty"`
	Last     string `json:"last,omitempty"`
	Next     string `json:"next,omitempty"`
	Previous string `json:"previous,omitempty"`
	Self     string `json:"self,omitempty"`
}

// Options represents pagination options
type Options struct {
	Page          int    `json:"page"`
	Limit         int    `json:"limit"`
	SortBy        string `json:"sort_by,omitempty"`
	SortDirection string `json:"sort_direction,omitempty"` // asc, desc
}

// DefaultOptions returns default pagination options
func DefaultOptions() *Options {
	return &Options{
		Page:          1,
		Limit:         20,
		SortBy:        "id",
		SortDirection: "desc",
	}
}

// Validate validates pagination options
func (o *Options) Validate() error {
	if o.Page < 1 {
		return fmt.Errorf("page must be greater than 0")
	}
	if o.Limit < 1 {
		return fmt.Errorf("limit must be greater than 0")
	}
	if o.Limit > 1000 {
		return fmt.Errorf("limit cannot exceed 1000")
	}
	if o.SortDirection != "" && o.SortDirection != "asc" && o.SortDirection != "desc" {
		return fmt.Errorf("sort_direction must be 'asc' or 'desc'")
	}
	return nil
}

// Calculate calculates pagination metadata
func (o *Options) Calculate(totalRows int64) *Pagination {
	totalPages := int(math.Ceil(float64(totalRows) / float64(o.Limit)))
	if totalPages < 1 {
		totalPages = 1
	}

	// Ensure page is within bounds
	page := o.Page
	if page > totalPages {
		page = totalPages
	}
	if page < 1 {
		page = 1
	}

	hasNext := page < totalPages
	hasPrev := page > 1

	pagination := &Pagination{
		Page:       page,
		Limit:      o.Limit,
		TotalRows:  totalRows,
		TotalPages: totalPages,
		Offset:     (page - 1) * o.Limit,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
		FirstPage:  1,
		LastPage:   totalPages,
	}

	if hasNext {
		pagination.NextPage = page + 1
	}
	if hasPrev {
		pagination.PrevPage = page - 1
	}

	return pagination
}

// NewPaginatedResult creates a new paginated result
func NewPaginatedResult[T any](data []T, pagination *Pagination) *PaginatedResult[T] {
	return &PaginatedResult[T]{
		Data:       data,
		Pagination: pagination,
	}
}

// NewPaginatedResultWithLinks creates a new paginated result with navigation links
func NewPaginatedResultWithLinks[T any](data []T, pagination *Pagination, baseURL string, queryParams url.Values) *PaginatedResult[T] {
	result := NewPaginatedResult(data, pagination)
	result.Links = GenerateLinks(pagination, baseURL, queryParams)
	return result
}

// GenerateLinks generates navigation links for pagination
func GenerateLinks(pagination *Pagination, baseURL string, queryParams url.Values) *Links {
	links := &Links{}

	// Helper function to build URL
	buildURL := func(page int) string {
		params := cloneQueryParams(queryParams)
		params.Set("page", strconv.Itoa(page))
		params.Set("limit", strconv.Itoa(pagination.Limit))

		return fmt.Sprintf("%s?%s", baseURL, params.Encode())
	}

	// Self link (current page)
	links.Self = buildURL(pagination.Page)

	// First page link
	links.First = buildURL(1)

	// Last page link
	links.Last = buildURL(pagination.TotalPages)

	// Next page link
	if pagination.HasNext {
		links.Next = buildURL(pagination.Page + 1)
	}

	// Previous page link
	if pagination.HasPrev {
		links.Previous = buildURL(pagination.Page - 1)
	}

	return links
}

// cloneQueryParams creates a copy of query parameters
func cloneQueryParams(params url.Values) url.Values {
	clone := url.Values{}
	for k, v := range params {
		clone[k] = v
	}
	return clone
}

// ParseFromURL parses pagination options from URL query parameters
func ParseFromURL(queryParams url.Values) *Options {
	opts := DefaultOptions()

	if pageStr := queryParams.Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			opts.Page = page
		}
	}

	if limitStr := queryParams.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			opts.Limit = limit
		}
	}

	if sortBy := queryParams.Get("sort_by"); sortBy != "" {
		opts.SortBy = sortBy
	}

	if sortDir := queryParams.Get("sort_direction"); sortDir != "" {
		opts.SortDirection = sortDir
	}

	return opts
}

// ApplyOptions applies pagination options to a query
func (o *Options) Apply() (string, []interface{}) {
	var args []interface{}
	var clauses []string

	// ORDER BY clause
	if o.SortBy != "" {
		orderDir := "ASC"
		if strings.ToLower(o.SortDirection) == "desc" {
			orderDir = "DESC"
		}
		clauses = append(clauses, fmt.Sprintf("ORDER BY %s %s", o.SortBy, orderDir))
	}

	// LIMIT and OFFSET clause
	clauses = append(clauses, fmt.Sprintf("LIMIT %d OFFSET %d", o.Limit, (o.Page-1)*o.Limit))

	return strings.Join(clauses, " "), args
}

// SQL helpers
type SQLBuilder struct {
	whereClause  string
	whereArgs    []interface{}
	orderClause  string
	limitClause  string
	offsetClause string
}

// NewSQLBuilder creates a new SQL builder
func NewSQLBuilder() *SQLBuilder {
	return &SQLBuilder{}
}

// Where adds WHERE clause
func (b *SQLBuilder) Where(condition string, args ...interface{}) *SQLBuilder {
	b.whereClause = condition
	b.whereArgs = args
	return b
}

// OrderBy adds ORDER BY clause
func (b *SQLBuilder) OrderBy(field string, direction string) *SQLBuilder {
	if direction == "" {
		direction = "ASC"
	}
	b.orderClause = fmt.Sprintf("ORDER BY %s %s", field, strings.ToUpper(direction))
	return b
}

// Paginate adds LIMIT and OFFSET clauses
func (b *SQLBuilder) Paginate(page, limit int) *SQLBuilder {
	if limit > 0 {
		b.limitClause = fmt.Sprintf("LIMIT %d", limit)
		if page > 1 {
			b.offsetClause = fmt.Sprintf("OFFSET %d", (page-1)*limit)
		}
	}
	return b
}

// Build builds the SQL query
func (b *SQLBuilder) Build() (string, []interface{}) {
	var clauses []string
	var args []interface{}

	// WHERE clause
	if b.whereClause != "" {
		clauses = append(clauses, "WHERE "+b.whereClause)
		args = append(args, b.whereArgs...)
	}

	// ORDER BY clause
	if b.orderClause != "" {
		clauses = append(clauses, b.orderClause)
	}

	// LIMIT clause
	if b.limitClause != "" {
		clauses = append(clauses, b.limitClause)
	}

	// OFFSET clause
	if b.offsetClause != "" {
		clauses = append(clauses, b.offsetClause)
	}

	return strings.Join(clauses, " "), args
}

// BuildCount builds a COUNT query
func (b *SQLBuilder) BuildCount() (string, []interface{}) {
	query := "SELECT COUNT(*) "
	if b.whereClause != "" {
		query += "WHERE " + b.whereClause
	}
	return query, b.whereArgs
}
