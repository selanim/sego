package examples

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/selanim/sego/pagination"
)

// User represents a user model
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// MockRepository is a mock repository for examples
type MockRepository struct {
	users []User
}

func NewMockRepository() *MockRepository {
	users := make([]User, 100)
	for i := 0; i < 100; i++ {
		users[i] = User{
			ID:        i + 1,
			Name:      fmt.Sprintf("User %d", i+1),
			Email:     fmt.Sprintf("user%d@example.com", i+1),
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Hour),
		}
	}
	return &MockRepository{users: users}
}

func (r *MockRepository) Count(ctx context.Context, conditions ...map[string]interface{}) (int64, error) {
	return int64(len(r.users)), nil
}

func (r *MockRepository) FindAll(ctx context.Context, query string, args ...interface{}) ([]interface{}, error) {
	// Simplified mock - in real implementation, you'd parse the query
	limit := 20
	offset := 0

	// Parse limit and offset from query (simplified)
	if len(args) >= 1 {
		if l, ok := args[0].(int); ok {
			limit = l
		}
	}
	if len(args) >= 2 {
		if o, ok := args[1].(int); ok {
			offset = o
		}
	}

	// Apply pagination
	end := offset + limit
	if end > len(r.users) {
		end = len(r.users)
	}

	var result []interface{}
	for i := offset; i < end; i++ {
		result = append(result, r.users[i])
	}

	return result, nil
}

func ExampleOffsetPagination() {
	ctx := context.Background()
	repo := NewMockRepository()

	// Parse options from URL query parameters
	queryParams := url.Values{}
	queryParams.Set("page", "2")
	queryParams.Set("limit", "10")
	queryParams.Set("sort_by", "created_at")
	queryParams.Set("sort_direction", "desc")

	opts := pagination.ParseFromURL(queryParams)
	fmt.Printf("Page: %d, Limit: %d\n", opts.Page, opts.Limit)

	// Create base URL for links
	baseURL := "https://api.example.com/users"

	// Paginate users
	result, err := pagination.PaginateWithURL(ctx, repo, opts,
		"SELECT * FROM users ORDER BY created_at DESC",
		[]interface{}{},
		func(row interface{}) User {
			return row.(User)
		},
		baseURL, queryParams)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Total users: %d\n", result.Pagination.TotalRows)
	fmt.Printf("Page %d of %d\n", result.Pagination.Page, result.Pagination.TotalPages)
	fmt.Printf("Showing %d users\n", len(result.Data))

	// Navigation links
	if result.Links != nil {
		fmt.Printf("First page: %s\n", result.Links.First)
		fmt.Printf("Last page: %s\n", result.Links.Last)
		if result.Links.Next != "" {
			fmt.Printf("Next page: %s\n", result.Links.Next)
		}
		if result.Links.Previous != "" {
			fmt.Printf("Previous page: %s\n", result.Links.Previous)
		}
	}

	// Display first few users
	for i, user := range result.Data {
		if i < 3 { // Show only first 3
			fmt.Printf("User %d: %s (%s)\n", user.ID, user.Name, user.Email)
		}
	}
}

func ExampleCursorPagination() {
	ctx := context.Background()
	repo := NewMockRepository()

	// Cursor pagination options
	opts := &pagination.CursorOptions{
		Limit:     15,
		Direction: "forward",
	}

	// Get cursor value function
	getCursorValue := func(user User) interface{} {
		return user.ID
	}

	// Perform cursor-based pagination
	result, err := pagination.PaginateCursor(ctx, repo, opts,
		"id",
		func(row interface{}) User {
			return row.(User)
		},
		getCursorValue)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Retrieved %d users\n", len(result.Data))
	fmt.Printf("Has more: %v\n", result.HasMore)

	if result.Next != "" {
		fmt.Printf("Next cursor: %s\n", result.Next)
	}
	if result.Previous != "" {
		fmt.Printf("Previous cursor: %s\n", result.Previous)
	}

	// Decode cursor to see its contents
	if result.Next != "" {
		cursor, err := pagination.DecodeCursor(result.Next)
		if err == nil && cursor != nil {
			fmt.Printf("Next cursor value: %v, direction: %s\n", cursor.Value, cursor.Direction)
		}
	}
}

func ExampleSQLBuilder() {
	// Create SQL builder
	builder := pagination.NewSQLBuilder().
		Where("active = ? AND age >= ?", true, 18).
		OrderBy("created_at", "desc").
		Paginate(2, 20)

	// Build query
	query, args := builder.Build()
	fmt.Printf("Query: SELECT * FROM users %s\n", query)
	fmt.Printf("Args: %v\n", args)

	// Build count query
	countQuery, countArgs := builder.BuildCount()
	fmt.Printf("Count Query: %s\n", countQuery)
	fmt.Printf("Count Args: %v\n", countArgs)
}

func ExamplePaginationMiddleware() {
	// Create middleware
	middleware := pagination.NewPaginationMiddleware()

	// Simulate HTTP request query parameters
	queryParams := url.Values{
		"page":           {"3"},
		"limit":          {"25"},
		"sort_by":        {"name"},
		"sort_direction": {"asc"},
	}

	// Parse and validate options
	opts, err := middleware.ParseOptions(queryParams)
	if err != nil {
		fmt.Printf("Error parsing options: %v\n", err)
		return
	}

	fmt.Printf("Parsed options: Page=%d, Limit=%d, Sort=%s, Direction=%s\n",
		opts.Page, opts.Limit, opts.SortBy, opts.SortDirection)

	// Build paginated query
	baseQuery := "SELECT * FROM users WHERE status = 'active'"
	finalQuery, queryArgs := middleware.PaginateQuery(baseQuery, opts)

	fmt.Printf("\nFinal Query:\n%s\n", finalQuery)
	fmt.Printf("Query Args: %v\n", queryArgs)

	// Calculate pagination for response
	totalRows := int64(150)
	pagination := opts.Calculate(totalRows)

	fmt.Printf("\nPagination Info:\n")
	fmt.Printf("  Page: %d/%d\n", pagination.Page, pagination.TotalPages)
	fmt.Printf("  Total Rows: %d\n", pagination.TotalRows)
	fmt.Printf("  Has Next: %v\n", pagination.HasNext)
	fmt.Printf("  Has Prev: %v\n", pagination.HasPrev)
	fmt.Printf("  Offset: %d\n", pagination.Offset)
}

func ExampleIntegrationWithHTTP() {
	// This example shows how to use pagination in an HTTP handler

	/*
		// In your HTTP handler:
		func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
			// Parse query parameters
			queryParams := r.URL.Query()

			// Parse pagination options
			opts := pagination.ParseFromURL(queryParams)

			// Validate options
			if err := opts.Validate(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// Get total count
			totalRows, err := h.repo.Count(r.Context())
			if err != nil {
				http.Error(w, "Failed to count users", http.StatusInternalServerError)
				return
			}

			// Calculate pagination
			pagination := opts.Calculate(totalRows)

			// Build query with pagination
			query := fmt.Sprintf(
				"SELECT * FROM users ORDER BY %s %s LIMIT %d OFFSET %d",
				opts.SortBy,
				opts.SortDirection,
				opts.Limit,
				pagination.Offset,
			)

			// Execute query
			users, err := h.repo.FindAll(r.Context(), query)
			if err != nil {
				http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
				return
			}

			// Generate navigation links
			baseURL := fmt.Sprintf("%s://%s%s", r.URL.Scheme, r.URL.Host, r.URL.Path)
			links := pagination.GenerateLinks(pagination, baseURL, queryParams)

			// Create response
			response := pagination.NewPaginatedResultWithLinks(users, pagination, baseURL, queryParams)

			// Send JSON response
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	*/

	fmt.Println("See code comments for HTTP handler integration example")
}

func ExampleAdvancedPagination() {
	// Advanced pagination with custom SQL
	ctx := context.Background()
	repo := NewMockRepository()

	// Custom pagination options
	opts := &pagination.Options{
		Page:          2,
		Limit:         15,
		SortBy:        "created_at",
		SortDirection: "desc",
	}

	// Complex query with joins
	baseQuery := `
		SELECT u.*, COUNT(o.id) as order_count 
		FROM users u
		LEFT JOIN orders o ON u.id = o.user_id
		WHERE u.active = true
		GROUP BY u.id
	`
	queryArgs := []interface{}{}

	// Paginate
	result, err := pagination.Paginate(ctx, repo, opts, baseQuery, queryArgs,
		func(row interface{}) User {
			// In real implementation, you'd map the joined result
			return row.(User)
		})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Advanced Pagination Result:\n")
	fmt.Printf("  Page: %d/%d\n", result.Pagination.Page, result.Pagination.TotalPages)
	fmt.Printf("  Items per page: %d\n", result.Pagination.Limit)
	fmt.Printf("  Total items: %d\n", result.Pagination.TotalRows)
	fmt.Printf("  Retrieved items: %d\n", len(result.Data))
}

func ExampleEdgeCases() {
	fmt.Println("Testing pagination edge cases:")

	// Test 1: Empty result set
	opts1 := &pagination.Options{Page: 1, Limit: 10}
	pagination1 := opts1.Calculate(0)
	fmt.Printf("1. Empty set: TotalPages=%d, HasNext=%v, HasPrev=%v\n",
		pagination1.TotalPages, pagination1.HasNext, pagination1.HasPrev)

	// Test 2: Single page
	opts2 := &pagination.Options{Page: 1, Limit: 10}
	pagination2 := opts2.Calculate(5)
	fmt.Printf("2. Single page: TotalPages=%d, HasNext=%v, HasPrev=%v\n",
		pagination2.TotalPages, pagination2.HasNext, pagination2.HasPrev)

	// Test 3: Page out of bounds (too high)
	opts3 := &pagination.Options{Page: 10, Limit: 10}
	pagination3 := opts3.Calculate(25)
	fmt.Printf("3. Page too high: Corrected Page=%d/%d\n",
		pagination3.Page, pagination3.TotalPages)

	// Test 4: Page out of bounds (too low)
	opts4 := &pagination.Options{Page: 0, Limit: 10}
	pagination4 := opts4.Calculate(100)
	fmt.Printf("4. Page too low: Corrected Page=%d\n", pagination4.Page)

	// Test 5: Large limit
	opts5 := &pagination.Options{Page: 1, Limit: 10000}
	if err := opts5.Validate(); err != nil {
		fmt.Printf("5. Large limit rejected: %v\n", err)
	}
}
