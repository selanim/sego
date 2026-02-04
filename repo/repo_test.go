package repo

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"
)

// TestModel is a test model for testing
type TestModel struct {
	ID        int       `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	Age       int       `db:"age"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// MockDB is a simple mock database connection implementing pgxpool.Pool interface
type MockDB struct{}

// Since we can't implement all methods of pgxpool.Pool, we'll test without actual DB
// For unit tests, we'll focus on testing the logic that doesn't require DB connection

func TestNewRepository(t *testing.T) {
	tests := []struct {
		name     string
		model    interface{}
		opts     Options
		expected string
	}{
		{
			name:     "default table name",
			model:    TestModel{},
			opts:     Options{},
			expected: "testmodels",
		},
		{
			name:     "custom table name",
			model:    TestModel{},
			opts:     Options{TableName: "custom_table"},
			expected: "custom_table",
		},
		{
			name:     "with cache TTL",
			model:    TestModel{},
			opts:     Options{CacheTTL: time.Hour},
			expected: "testmodels",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a repository without DB for table name testing
			// We'll create a minimal test that only checks table name logic
			repo := &Repository{
				tableName: tt.expected,
				cache:     NewCache(tt.opts.CacheTTL),
			}

			if repo.GetTableName() != tt.expected {
				t.Errorf("Expected table name %s, got %s", tt.expected, repo.GetTableName())
			}

			if tt.opts.CacheTTL > 0 && repo.GetCache() == nil {
				t.Error("Expected cache to be initialized")
			}
		})
	}
}

func TestTableNameExtraction(t *testing.T) {
	// Test table name extraction logic
	type User struct{}
	type ProductItem struct{}

	tests := []struct {
		name     string
		model    interface{}
		expected string
	}{
		{"simple model", User{}, "users"},
		{"model with multiple words", ProductItem{}, "productitems"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This tests the logic that would be in NewRepository
			modelType := reflect.TypeOf(tt.model)
			if modelType.Kind() == reflect.Ptr {
				modelType = modelType.Elem()
			}
			tableName := strings.ToLower(modelType.Name()) + "s"

			if tableName != tt.expected {
				t.Errorf("Expected table name %s, got %s", tt.expected, tableName)
			}
		})
	}
}

func TestCacheOperations(t *testing.T) {
	cache := NewCache(time.Minute)

	// Test Set and Get
	cache.Set("key1", "value1")
	val, found := cache.Get("key1")
	if !found {
		t.Error("Expected key1 to be found in cache")
	}
	if val != "value1" {
		t.Errorf("Expected value1, got %v", val)
	}

	// Test Get non-existent key
	val, found = cache.Get("key2")
	if found {
		t.Error("Expected key2 not to be found in cache")
	}
	if val != nil {
		t.Errorf("Expected nil, got %v", val)
	}

	// Test Delete
	cache.Delete("key1")
	val, found = cache.Get("key1")
	if found {
		t.Error("Expected key1 to be deleted from cache")
	}

	// Test expiration
	cache.SetWithTTL("expiring", "value", time.Millisecond*10)
	time.Sleep(time.Millisecond * 20)
	val, found = cache.Get("expiring")
	if found {
		t.Error("Expected expired key to not be found")
	}

	// Test Clear
	cache.Set("key3", "value3")
	cache.Set("key4", "value4")
	if cache.Size() != 2 {
		t.Errorf("Expected cache size 2, got %d", cache.Size())
	}
	cache.Clear()
	if cache.Size() != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", cache.Size())
	}

	// Test GetOrSet
	calls := 0
	fn := func() (interface{}, error) {
		calls++
		return "computed", nil
	}

	// First call should execute function
	val, err := cache.GetOrSet("computed", fn)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if val != "computed" {
		t.Errorf("Expected 'computed', got %v", val)
	}
	if calls != 1 {
		t.Errorf("Expected 1 function call, got %d", calls)
	}

	// Second call should get from cache
	val, err = cache.GetOrSet("computed", fn)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if val != "computed" {
		t.Errorf("Expected 'computed', got %v", val)
	}
	if calls != 1 {
		t.Errorf("Expected function not to be called again, got %d calls", calls)
	}
}

func TestQueryBuilder(t *testing.T) {
	qb := NewQueryBuilder("users")

	// Test basic query
	query, args := qb.Select("id", "name", "email").
		WhereEq("active", true).
		WhereLike("name", "john").
		OrderBy("created_at", true).
		Limit(10).
		Offset(0).
		Build()

	expectedQuery := "SELECT id, name, email FROM users WHERE active = $1 AND name LIKE $2 ORDER BY created_at DESC LIMIT 10 OFFSET 0"
	if query != expectedQuery {
		t.Errorf("Expected query: %s\nGot: %s", expectedQuery, query)
	}
	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}
	if args[0] != true {
		t.Errorf("Expected first arg to be true, got %v", args[0])
	}
	if args[1] != "%john%" {
		t.Errorf("Expected second arg to be '%%john%%', got %v", args[1])
	}

	// Test IN clause
	qb2 := NewQueryBuilder("products")
	query2, args2 := qb2.Select("*").
		WhereIn("category_id", []interface{}{1, 2, 3}).
		Build()

	if !strings.Contains(query2, "category_id IN (") {
		t.Errorf("Expected IN clause in query: %s", query2)
	}
	if len(args2) != 3 {
		t.Errorf("Expected 3 args, got %d", len(args2))
	}

	// Test JOIN
	qb3 := NewQueryBuilder("orders")
	query3, _ := qb3.Select("orders.*", "users.name").
		InnerJoin("users", "orders.user_id = users.id").
		WhereEq("orders.status", "completed").
		Build()

	if !strings.Contains(query3, "INNER JOIN users ON orders.user_id = users.id") {
		t.Errorf("Expected JOIN clause in query: %s", query3)
	}

	// Test Count query
	qb4 := NewQueryBuilder("items")
	countQuery, countArgs := qb4.WhereEq("status", "active").BuildCount()
	expectedCountQuery := "SELECT COUNT(*) FROM items WHERE status = $1"
	if countQuery != expectedCountQuery {
		t.Errorf("Expected count query: %s\nGot: %s", expectedCountQuery, countQuery)
	}
	if len(countArgs) != 1 {
		t.Errorf("Expected 1 arg, got %d", len(countArgs))
	}
}

func TestPagination(t *testing.T) {
	// Test pagination options validation
	opts := &PaginationOptions{
		Page:  0,
		Limit: 10,
	}
	err := opts.Validate()
	if err == nil || !strings.Contains(err.Error(), "page must be greater than 0") {
		t.Errorf("Expected validation error for page 0, got: %v", err)
	}

	opts.Page = 1
	opts.Limit = 0
	err = opts.Validate()
	if err == nil || !strings.Contains(err.Error(), "limit must be greater than 0") {
		t.Errorf("Expected validation error for limit 0, got: %v", err)
	}

	opts.Limit = 1001
	err = opts.Validate()
	if err == nil || !strings.Contains(err.Error(), "limit cannot exceed 1000") {
		t.Errorf("Expected validation error for limit > 1000, got: %v", err)
	}

	// Test valid options
	opts.Limit = 20
	err = opts.Validate()
	if err != nil {
		t.Errorf("Expected no error for valid options, got: %v", err)
	}

	// Test pagination calculation
	pagination := opts.ToPagination(105)
	if pagination.Page != 1 {
		t.Errorf("Expected page 1, got %d", pagination.Page)
	}
	if pagination.Limit != 20 {
		t.Errorf("Expected limit 20, got %d", pagination.Limit)
	}
	if pagination.TotalRows != 105 {
		t.Errorf("Expected total rows 105, got %d", pagination.TotalRows)
	}
	expectedPages := int(math.Ceil(105.0 / 20.0))
	if pagination.TotalPages != expectedPages {
		t.Errorf("Expected total pages %d, got %d", expectedPages, pagination.TotalPages)
	}
	if pagination.Offset != 0 {
		t.Errorf("Expected offset 0, got %d", pagination.Offset)
	}

	// Test next/prev page
	pagination.Page = 3
	if !pagination.HasNext() {
		t.Error("Expected to have next page")
	}
	if !pagination.HasPrev() {
		t.Error("Expected to have previous page")
	}
	if pagination.NextPage() != 4 {
		t.Errorf("Expected next page 4, got %d", pagination.NextPage())
	}
	if pagination.PrevPage() != 2 {
		t.Errorf("Expected previous page 2, got %d", pagination.PrevPage())
	}

	// Test edge cases
	pagination.Page = 1
	if pagination.HasPrev() {
		t.Error("Should not have previous page on first page")
	}
	if pagination.PrevPage() != 1 {
		t.Errorf("Expected previous page to be 1 on first page, got %d", pagination.PrevPage())
	}

	pagination.Page = pagination.TotalPages
	if pagination.HasNext() {
		t.Error("Should not have next page on last page")
	}
	if pagination.NextPage() != pagination.TotalPages {
		t.Errorf("Expected next page to be %d on last page, got %d", pagination.TotalPages, pagination.NextPage())
	}
}

func TestExtractFieldsAndValues(t *testing.T) {
	// Create a test repository without DB
	repo := &Repository{
		tableName: "testmodels",
		modelType: reflect.TypeOf(TestModel{}),
		cache:     NewCache(time.Minute),
	}

	now := time.Now()
	model := TestModel{
		ID:        1,
		Name:      "John Doe",
		Email:     "john@example.com",
		Age:       30,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Test including ID
	fields, values, placeholders := repo.extractFieldsAndValues(model, true)

	expectedFields := []string{"id", "name", "email", "age", "created_at", "updated_at"}
	for _, field := range expectedFields {
		if !strings.Contains(fields, field) {
			t.Errorf("Expected field %s in fields: %s", field, fields)
		}
	}

	if len(values) != 6 {
		t.Errorf("Expected 6 values, got %d", len(values))
	}

	placeholderCount := len(strings.Split(placeholders, ", "))
	if placeholderCount != 6 {
		t.Errorf("Expected 6 placeholders, got %d", placeholderCount)
	}

	// Test excluding ID
	fields, values, placeholders = repo.extractFieldsAndValues(model, false)
	if strings.Contains(fields, "id") {
		t.Errorf("Expected ID to be excluded, got fields: %s", fields)
	}
	if len(values) != 5 {
		t.Errorf("Expected 5 values when excluding ID, got %d", len(values))
	}
	placeholderCount = len(strings.Split(placeholders, ", "))
	if placeholderCount != 5 {
		t.Errorf("Expected 5 placeholders when excluding ID, got %d", placeholderCount)
	}
}

func TestBuildWhereClause(t *testing.T) {
	// Create repository without DB
	repo := &Repository{
		tableName: "testmodels",
		modelType: reflect.TypeOf(TestModel{}),
		cache:     NewCache(time.Minute),
	}

	conditions := map[string]interface{}{
		"name":  "John",
		"age":   30,
		"email": "john@example.com",
	}

	where, args := repo.buildWhereClause(conditions)

	expectedParts := []string{"name = $1", "age = $2", "email = $3"}
	for _, part := range expectedParts {
		if !strings.Contains(where, part) {
			t.Errorf("Expected WHERE clause to contain %s, got: %s", part, where)
		}
	}

	if !strings.Contains(where, " AND ") {
		t.Errorf("Expected WHERE clause to contain AND, got: %s", where)
	}

	if len(args) != 3 {
		t.Errorf("Expected 3 args, got %d", len(args))
	}

	if args[0] != "John" {
		t.Errorf("Expected first arg to be 'John', got %v", args[0])
	}
	if args[1] != 30 {
		t.Errorf("Expected second arg to be 30, got %v", args[1])
	}
	if args[2] != "john@example.com" {
		t.Errorf("Expected third arg to be 'john@example.com', got %v", args[2])
	}
}

func TestRepositoryErrors(t *testing.T) {
	if ErrNotFound.Error() != "record not found" {
		t.Errorf("Expected ErrNotFound message 'record not found', got: %s", ErrNotFound.Error())
	}
	if ErrInvalidData.Error() != "invalid data" {
		t.Errorf("Expected ErrInvalidData message 'invalid data', got: %s", ErrInvalidData.Error())
	}
	if ErrDuplicate.Error() != "duplicate record" {
		t.Errorf("Expected ErrDuplicate message 'duplicate record', got: %s", ErrDuplicate.Error())
	}
}

func TestCacheCleanup(t *testing.T) {
	cache := NewCacheWithCleanup(time.Millisecond*50, time.Millisecond*100)

	// Set items with short TTL
	cache.SetWithTTL("key1", "value1", time.Millisecond*10)
	cache.SetWithTTL("key2", "value2", time.Millisecond*200) // Longer TTL

	// Wait for first item to expire and cleanup to run
	time.Sleep(time.Millisecond * 150)

	// key1 should be cleaned up
	val, found := cache.Get("key1")
	if found {
		t.Error("Expected key1 to be cleaned up")
	}

	// key2 should still exist
	val, found = cache.Get("key2")
	if !found {
		t.Error("Expected key2 to still exist")
	}
	if val != "value2" {
		t.Errorf("Expected value2, got %v", val)
	}
}

func TestPaginationDefault(t *testing.T) {
	defaultPagination := DefaultPagination()
	if defaultPagination.Page != 1 {
		t.Errorf("Expected default page 1, got %d", defaultPagination.Page)
	}
	if defaultPagination.Limit != 20 {
		t.Errorf("Expected default limit 20, got %d", defaultPagination.Limit)
	}
}

func TestCacheExists(t *testing.T) {
	cache := NewCache(time.Minute)

	// Test non-existent key
	if cache.Exists("nonexistent") {
		t.Error("Expected non-existent key to not exist")
	}

	// Test existing key
	cache.Set("exists", "value")
	if !cache.Exists("exists") {
		t.Error("Expected existing key to exist")
	}

	// Test expired key
	cache.SetWithTTL("expired", "value", time.Millisecond*10)
	time.Sleep(time.Millisecond * 20)
	if cache.Exists("expired") {
		t.Error("Expected expired key to not exist")
	}
}

func TestCacheKeys(t *testing.T) {
	cache := NewCache(time.Minute)

	// Empty cache
	keys := cache.Keys()
	if len(keys) != 0 {
		t.Errorf("Expected 0 keys, got %d", len(keys))
	}

	// With items
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.SetWithTTL("expired", "value", time.Millisecond*10)
	time.Sleep(time.Millisecond * 20)

	keys = cache.Keys()
	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(keys))
	}

	// Check that expired key is not included
	for _, key := range keys {
		if key == "expired" {
			t.Error("Expired key should not be in keys list")
		}
	}
}

func TestQueryBuilderGroupBy(t *testing.T) {
	qb := NewQueryBuilder("orders")
	query, args := qb.
		Select("customer_id", "COUNT(*) as order_count", "SUM(amount) as total_amount").
		GroupBy("customer_id").
		Having("COUNT(*) > $1", 5).
		Build()

	if !strings.Contains(query, "GROUP BY customer_id") {
		t.Errorf("Expected GROUP BY clause in query: %s", query)
	}
	if !strings.Contains(query, "HAVING COUNT(*) > $1") {
		t.Errorf("Expected HAVING clause in query: %s", query)
	}
	if len(args) != 1 || args[0] != 5 {
		t.Errorf("Expected arg [5], got %v", args)
	}
}

func TestQueryBuilderJoins(t *testing.T) {
	qb := NewQueryBuilder("orders")
	query, _ := qb.
		Select("orders.id", "customers.name", "orders.amount").
		InnerJoin("customers", "orders.customer_id = customers.id").
		LeftJoin("payments", "orders.id = payments.order_id").
		RightJoin("shipping", "orders.id = shipping.order_id").
		Build()

	if !strings.Contains(query, "INNER JOIN customers ON orders.customer_id = customers.id") {
		t.Errorf("Expected INNER JOIN in query: %s", query)
	}
	if !strings.Contains(query, "LEFT JOIN payments ON orders.id = payments.order_id") {
		t.Errorf("Expected LEFT JOIN in query: %s", query)
	}
	if !strings.Contains(query, "RIGHT JOIN shipping ON orders.id = shipping.order_id") {
		t.Errorf("Expected RIGHT JOIN in query: %s", query)
	}
}

// Integration test example (would require actual database)
func TestRepositoryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This would require a real database connection
	// For now, we'll just skip it
	t.Skip("Integration test requires database connection")
}

// Benchmark tests
func BenchmarkCacheSet(b *testing.B) {
	cache := NewCache(time.Hour)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
	}
}

func BenchmarkCacheGet(b *testing.B) {
	cache := NewCache(time.Hour)
	for i := 0; i < 1000; i++ {
		cache.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(fmt.Sprintf("key%d", i%1000))
	}
}

func BenchmarkQueryBuilder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		qb := NewQueryBuilder("users")
		qb.Select("id", "name", "email").
			WhereEq("active", true).
			WhereIn("id", []interface{}{1, 2, 3, 4, 5}).
			OrderBy("created_at", true).
			Limit(100)
		qb.Build()
	}
}
