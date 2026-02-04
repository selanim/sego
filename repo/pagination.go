package repo

import (
	"context"
	"fmt"
	"math"
	"reflect"
)

// Pagination represents pagination metadata
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalRows  int64 `json:"total_rows"`
	TotalPages int   `json:"total_pages"`
	Offset     int   `json:"-"`
}

// PaginatedResult represents a paginated query result
type PaginatedResult struct {
	Data       interface{} `json:"data"`
	Pagination *Pagination `json:"pagination"`
}

// PaginationOptions represents pagination input options
type PaginationOptions struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

// DefaultPagination creates default pagination options
func DefaultPagination() *PaginationOptions {
	return &PaginationOptions{
		Page:  1,
		Limit: 20,
	}
}

// Validate validates pagination options
func (p *PaginationOptions) Validate() error {
	if p.Page < 1 {
		return fmt.Errorf("page must be greater than 0")
	}
	if p.Limit < 1 {
		return fmt.Errorf("limit must be greater than 0")
	}
	if p.Limit > 1000 {
		return fmt.Errorf("limit cannot exceed 1000")
	}
	return nil
}

// ToPagination converts PaginationOptions to Pagination
func (p *PaginationOptions) ToPagination(totalRows int64) *Pagination {
	totalPages := int(math.Ceil(float64(totalRows) / float64(p.Limit)))

	return &Pagination{
		Page:       p.Page,
		Limit:      p.Limit,
		TotalRows:  totalRows,
		TotalPages: totalPages,
		Offset:     (p.Page - 1) * p.Limit,
	}
}

// NextPage gets the next page number
func (p *Pagination) NextPage() int {
	if p.Page < p.TotalPages {
		return p.Page + 1
	}
	return p.Page
}

// PrevPage gets the previous page number
func (p *Pagination) PrevPage() int {
	if p.Page > 1 {
		return p.Page - 1
	}
	return p.Page
}

// HasNext checks if there is a next page
func (p *Pagination) HasNext() bool {
	return p.Page < p.TotalPages
}

// HasPrev checks if there is a previous page
func (p *Pagination) HasPrev() bool {
	return p.Page > 1
}

// PaginatedQueryResult is a generic paginated result
type PaginatedQueryResult[T any] struct {
	Data       []T         `json:"data"`
	Pagination *Pagination `json:"pagination"`
}

// Paginate executes a paginated query
func (r *Repository) Paginate(ctx context.Context, opts *PaginationOptions, queryOptions ...QueryOptions) (*PaginatedResult, error) {
	if opts == nil {
		opts = DefaultPagination()
	}

	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid pagination options: %w", err)
	}

	// Get total count
	var conditions map[string]interface{}
	if len(queryOptions) > 0 {
		conditions = queryOptions[0].Conditions
	}

	totalRows, err := r.Count(ctx, conditions)
	if err != nil {
		return nil, fmt.Errorf("failed to count records: %w", err)
	}

	// Create pagination
	pagination := opts.ToPagination(totalRows)

	// Build query options with pagination
	queryOpts := QueryOptions{
		Conditions: conditions,
		OrderBy:    "id DESC",
		Limit:      pagination.Limit,
		Offset:     pagination.Offset,
	}

	if len(queryOptions) > 0 {
		if queryOptions[0].OrderBy != "" {
			queryOpts.OrderBy = queryOptions[0].OrderBy
		}
	}

	// Get paginated data
	data, err := r.FindAll(ctx, queryOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch paginated data: %w", err)
	}

	return &PaginatedResult{
		Data:       data,
		Pagination: pagination,
	}, nil
}

// PaginateWithQueryBuilder executes a paginated query using QueryBuilder
func (r *Repository) PaginateWithQueryBuilder(ctx context.Context, qb *QueryBuilder, opts *PaginationOptions) (*PaginatedResult, error) {
	if opts == nil {
		opts = DefaultPagination()
	}

	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid pagination options: %w", err)
	}

	// Get total count
	countQuery, countArgs := qb.BuildCount()
	var totalRows int64
	err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&totalRows)
	if err != nil {
		return nil, fmt.Errorf("failed to count records: %w", err)
	}

	// Create pagination
	pagination := opts.ToPagination(totalRows)

	// Apply pagination to query builder
	qb.Limit(pagination.Limit)
	qb.Offset(pagination.Offset)

	// Execute query
	query, args := qb.Build()
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute paginated query: %w", err)
	}
	defer rows.Close()

	// Scan results
	var results []interface{}
	for rows.Next() {
		// Create new instance of model
		model := reflect.New(r.modelType).Interface()
		err := rows.Scan(model)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, model)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return &PaginatedResult{
		Data:       results,
		Pagination: pagination,
	}, nil
}
