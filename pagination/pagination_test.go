package pagination

import (
	"net/url"
	"testing"
	"time"
)

func TestOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		options Options
		wantErr bool
	}{
		{
			name: "valid options",
			options: Options{
				Page:  1,
				Limit: 10,
			},
			wantErr: false,
		},
		{
			name: "zero page",
			options: Options{
				Page:  0,
				Limit: 10,
			},
			wantErr: true,
		},
		{
			name: "negative page",
			options: Options{
				Page:  -1,
				Limit: 10,
			},
			wantErr: true,
		},
		{
			name: "zero limit",
			options: Options{
				Page:  1,
				Limit: 0,
			},
			wantErr: true,
		},
		{
			name: "negative limit",
			options: Options{
				Page:  1,
				Limit: -5,
			},
			wantErr: true,
		},
		{
			name: "excessive limit",
			options: Options{
				Page:  1,
				Limit: 2000,
			},
			wantErr: true,
		},
		{
			name: "invalid sort direction",
			options: Options{
				Page:          1,
				Limit:         10,
				SortDirection: "invalid",
			},
			wantErr: true,
		},
		{
			name: "valid sort direction asc",
			options: Options{
				Page:          1,
				Limit:         10,
				SortDirection: "asc",
			},
			wantErr: false,
		},
		{
			name: "valid sort direction desc",
			options: Options{
				Page:          1,
				Limit:         10,
				SortDirection: "desc",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.options.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Options.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err == nil {
				t.Error("Expected error but got none")
			}
		})
	}
}

func TestOptions_Calculate(t *testing.T) {
	tests := []struct {
		name        string
		options     Options
		totalRows   int64
		wantPage    int
		wantTotal   int
		wantOffset  int
		wantHasNext bool
		wantHasPrev bool
	}{
		{
			name: "first page of many",
			options: Options{
				Page:  1,
				Limit: 10,
			},
			totalRows:   100,
			wantPage:    1,
			wantTotal:   10,
			wantOffset:  0,
			wantHasNext: true,
			wantHasPrev: false,
		},
		{
			name: "middle page",
			options: Options{
				Page:  3,
				Limit: 10,
			},
			totalRows:   100,
			wantPage:    3,
			wantTotal:   10,
			wantOffset:  20,
			wantHasNext: true,
			wantHasPrev: true,
		},
		{
			name: "last page",
			options: Options{
				Page:  10,
				Limit: 10,
			},
			totalRows:   100,
			wantPage:    10,
			wantTotal:   10,
			wantOffset:  90,
			wantHasNext: false,
			wantHasPrev: true,
		},
		{
			name: "page out of bounds high",
			options: Options{
				Page:  15,
				Limit: 10,
			},
			totalRows:   100,
			wantPage:    10, // Corrected to last page
			wantTotal:   10,
			wantOffset:  90,
			wantHasNext: false,
			wantHasPrev: true,
		},
		{
			name: "page out of bounds low",
			options: Options{
				Page:  0,
				Limit: 10,
			},
			totalRows:   100,
			wantPage:    1, // Corrected to first page
			wantTotal:   10,
			wantOffset:  0,
			wantHasNext: true,
			wantHasPrev: false,
		},
		{
			name: "empty result set",
			options: Options{
				Page:  1,
				Limit: 10,
			},
			totalRows:   0,
			wantPage:    1,
			wantTotal:   1, // At least 1 page
			wantOffset:  0,
			wantHasNext: false,
			wantHasPrev: false,
		},
		{
			name: "exact page fit",
			options: Options{
				Page:  2,
				Limit: 25,
			},
			totalRows:   50,
			wantPage:    2,
			wantTotal:   2,
			wantOffset:  25,
			wantHasNext: false,
			wantHasPrev: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pagination := tt.options.Calculate(tt.totalRows)

			if pagination.Page != tt.wantPage {
				t.Errorf("Page = %d, want %d", pagination.Page, tt.wantPage)
			}

			if pagination.TotalPages != tt.wantTotal {
				t.Errorf("TotalPages = %d, want %d", pagination.TotalPages, tt.wantTotal)
			}

			if pagination.Offset != tt.wantOffset {
				t.Errorf("Offset = %d, want %d", pagination.Offset, tt.wantOffset)
			}

			if pagination.HasNext != tt.wantHasNext {
				t.Errorf("HasNext = %v, want %v", pagination.HasNext, tt.wantHasNext)
			}

			if pagination.HasPrev != tt.wantHasPrev {
				t.Errorf("HasPrev = %v, want %v", pagination.HasPrev, tt.wantHasPrev)
			}

			if pagination.TotalRows != tt.totalRows {
				t.Errorf("TotalRows = %d, want %d", pagination.TotalRows, tt.totalRows)
			}

			if pagination.Limit != tt.options.Limit {
				t.Errorf("Limit = %d, want %d", pagination.Limit, tt.options.Limit)
			}
		})
	}
}

func TestParseFromURL(t *testing.T) {
	tests := []struct {
		name      string
		query     url.Values
		wantPage  int
		wantLimit int
		wantSort  string
	}{
		{
			name: "basic query",
			query: url.Values{
				"page":  {"2"},
				"limit": {"15"},
			},
			wantPage:  2,
			wantLimit: 15,
			wantSort:  "id", // default
		},
		{
			name: "with sort parameters",
			query: url.Values{
				"page":           {"3"},
				"limit":          {"20"},
				"sort_by":        {"name"},
				"sort_direction": {"asc"},
			},
			wantPage:  3,
			wantLimit: 20,
			wantSort:  "name",
		},
		{
			name: "invalid page number",
			query: url.Values{
				"page":  {"invalid"},
				"limit": {"10"},
			},
			wantPage:  1, // default
			wantLimit: 10,
		},
		{
			name:      "empty query",
			query:     url.Values{},
			wantPage:  1,  // default
			wantLimit: 20, // default
		},
		{
			name: "zero limit",
			query: url.Values{
				"limit": {"0"},
			},
			wantPage:  1,
			wantLimit: 20, // default because 0 is invalid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := ParseFromURL(tt.query)

			if opts.Page != tt.wantPage {
				t.Errorf("Page = %d, want %d", opts.Page, tt.wantPage)
			}

			if opts.Limit != tt.wantLimit {
				t.Errorf("Limit = %d, want %d", opts.Limit, tt.wantLimit)
			}

			if opts.SortBy != tt.wantSort {
				t.Errorf("SortBy = %s, want %s", opts.SortBy, tt.wantSort)
			}
		})
	}
}

func TestGenerateLinks(t *testing.T) {
	baseURL := "https://api.example.com/users"
	queryParams := url.Values{
		"filter": {"active"},
		"search": {"john"},
	}

	pagination := &Pagination{
		Page:       3,
		Limit:      10,
		TotalRows:  100,
		TotalPages: 10,
		HasNext:    true,
		HasPrev:    true,
		NextPage:   4,
		PrevPage:   2,
	}

	links := GenerateLinks(pagination, baseURL, queryParams)

	if links == nil {
		t.Fatal("Expected links, got nil")
	}

	// Check self link contains current page
	if links.Self == "" {
		t.Error("Self link should not be empty")
	}
	if !containsParam(links.Self, "page=3") {
		t.Error("Self link should contain page=3")
	}

	// Check first link
	if links.First == "" {
		t.Error("First link should not be empty")
	}
	if !containsParam(links.First, "page=1") {
		t.Error("First link should contain page=1")
	}

	// Check last link
	if links.Last == "" {
		t.Error("Last link should not be empty")
	}
	if !containsParam(links.Last, "page=10") {
		t.Error("Last link should contain page=10")
	}

	// Check next link
	if links.Next == "" {
		t.Error("Next link should not be empty")
	}
	if !containsParam(links.Next, "page=4") {
		t.Error("Next link should contain page=4")
	}

	// Check previous link
	if links.Previous == "" {
		t.Error("Previous link should not be empty")
	}
	if !containsParam(links.Previous, "page=2") {
		t.Error("Previous link should contain page=2")
	}

	// Check that original query params are preserved
	if !containsParam(links.Self, "filter=active") {
		t.Error("Links should preserve original query parameters")
	}
	if !containsParam(links.Self, "search=john") {
		t.Error("Links should preserve original query parameters")
	}
}

func containsParam(urlStr, param string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	query := u.Query().Encode()
	return containsSubstring(query, param)
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsSubstring(s[1:], substr))
}

func TestCursorEncoding(t *testing.T) {
	originalCursor := &Cursor{
		Value:     12345,
		Direction: "next",
		Timestamp: time.Now().Truncate(time.Millisecond), // Truncate for comparison
	}

	// Encode
	encoded, err := EncodeCursor(originalCursor)
	if err != nil {
		t.Fatalf("Failed to encode cursor: %v", err)
	}

	if encoded == "" {
		t.Error("Encoded cursor should not be empty")
	}

	// Decode
	decoded, err := DecodeCursor(encoded)
	if err != nil {
		t.Fatalf("Failed to decode cursor: %v", err)
	}

	if decoded == nil {
		t.Fatal("Decoded cursor should not be nil")
	}

	// Compare
	if decoded.Value != originalCursor.Value {
		t.Errorf("Value mismatch: got %v, want %v", decoded.Value, originalCursor.Value)
	}

	if decoded.Direction != originalCursor.Direction {
		t.Errorf("Direction mismatch: got %s, want %s", decoded.Direction, originalCursor.Direction)
	}

	// Compare timestamps (within a second)
	timeDiff := decoded.Timestamp.Sub(originalCursor.Timestamp)
	if timeDiff > time.Second || timeDiff < -time.Second {
		t.Errorf("Timestamp mismatch: got %v, want %v", decoded.Timestamp, originalCursor.Timestamp)
	}

	// Test empty cursor
	decodedEmpty, err := DecodeCursor("")
	if err != nil {
		t.Fatalf("Failed to decode empty cursor: %v", err)
	}
	if decodedEmpty != nil {
		t.Error("Empty cursor string should decode to nil")
	}
}

func TestSQLBuilder(t *testing.T) {
	builder := NewSQLBuilder().
		Where("status = ? AND age > ?", "active", 18).
		OrderBy("created_at", "desc").
		Paginate(2, 20)

	query, args := builder.Build()

	expectedQueryParts := []string{
		"WHERE status = ? AND age > ?",
		"ORDER BY created_at DESC",
		"LIMIT 20",
		"OFFSET 20",
	}

	for _, part := range expectedQueryParts {
		if !containsSubstring(query, part) {
			t.Errorf("Query should contain: %s", part)
		}
	}

	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}

	if args[0] != "active" || args[1] != 18 {
		t.Errorf("Args mismatch: got %v", args)
	}

	// Test count query
	countQuery, countArgs := builder.BuildCount()
	if !containsSubstring(countQuery, "SELECT COUNT(*)") {
		t.Error("Count query should start with SELECT COUNT(*)")
	}
	if !containsSubstring(countQuery, "WHERE status = ? AND age > ?") {
		t.Error("Count query should contain WHERE clause")
	}
	if len(countArgs) != 2 {
		t.Errorf("Expected 2 count args, got %d", len(countArgs))
	}
}

func TestPaginationMiddleware(t *testing.T) {
	middleware := NewPaginationMiddleware()

	queryParams := url.Values{
		"page":           {"2"},
		"limit":          {"30"},
		"sort_by":        {"name"},
		"sort_direction": {"asc"},
	}

	opts, err := middleware.ParseOptions(queryParams)
	if err != nil {
		t.Fatalf("Failed to parse options: %v", err)
	}

	if opts.Page != 2 {
		t.Errorf("Page = %d, want 2", opts.Page)
	}

	if opts.Limit != 30 {
		t.Errorf("Limit = %d, want 30", opts.Limit)
	}

	if opts.SortBy != "name" {
		t.Errorf("SortBy = %s, want name", opts.SortBy)
	}

	if opts.SortDirection != "asc" {
		t.Errorf("SortDirection = %s, want asc", opts.SortDirection)
	}

	// Test with invalid sort field
	queryParams.Set("sort_by", "invalid_field")
	opts, err = middleware.ParseOptions(queryParams)
	if err != nil {
		t.Fatalf("Failed to parse options: %v", err)
	}

	if opts.SortBy != middleware.DefaultSort {
		t.Errorf("Invalid sort should default to %s, got %s", middleware.DefaultSort, opts.SortBy)
	}

	// Test with excessive limit
	queryParams.Set("limit", "200")
	opts, err = middleware.ParseOptions(queryParams)
	if err != nil {
		t.Fatalf("Failed to parse options: %v", err)
	}

	if opts.Limit != middleware.MaxLimit {
		t.Errorf("Excessive limit should be capped at %d, got %d", middleware.MaxLimit, opts.Limit)
	}
}

func TestApplyCursorQuery(t *testing.T) {
	tests := []struct {
		name        string
		field       string
		cursorValue interface{}
		direction   string
		limit       int
		wantQuery   string
		wantArgsLen int
	}{
		{
			name:        "forward pagination with cursor",
			field:       "id",
			cursorValue: 100,
			direction:   "next",
			limit:       20,
			wantQuery:   "WHERE id > ? ORDER BY id ASC LIMIT ?",
			wantArgsLen: 2,
		},
		{
			name:        "backward pagination with cursor",
			field:       "created_at",
			cursorValue: "2024-01-15T10:30:00Z",
			direction:   "prev",
			limit:       15,
			wantQuery:   "WHERE created_at < ? ORDER BY created_at DESC LIMIT ?",
			wantArgsLen: 2,
		},
		{
			name:        "first page no cursor",
			field:       "id",
			cursorValue: nil,
			direction:   "next",
			limit:       10,
			wantQuery:   "ORDER BY id ASC LIMIT ?",
			wantArgsLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, args := ApplyCursorQuery(tt.field, tt.cursorValue, tt.direction, tt.limit)

			if query != tt.wantQuery {
				t.Errorf("Query = %s, want %s", query, tt.wantQuery)
			}

			if len(args) != tt.wantArgsLen {
				t.Errorf("Args length = %d, want %d", len(args), tt.wantArgsLen)
			}

			if tt.cursorValue != nil && args[0] != tt.cursorValue {
				t.Errorf("First arg = %v, want %v", args[0], tt.cursorValue)
			}

			if args[len(args)-1] != tt.limit {
				t.Errorf("Last arg (limit) = %v, want %v", args[len(args)-1], tt.limit)
			}
		})
	}
}

func BenchmarkOptionsCalculate(b *testing.B) {
	opts := Options{Page: 5, Limit: 20}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		opts.Calculate(1000)
	}
}

func BenchmarkParseFromURL(b *testing.B) {
	query := url.Values{
		"page":           {"3"},
		"limit":          {"25"},
		"sort_by":        {"name"},
		"sort_direction": {"asc"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseFromURL(query)
	}
}
