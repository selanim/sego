package repo

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository represents a database repository
type Repository struct {
	db        *pgxpool.Pool
	cache     *Cache
	tableName string
	modelType reflect.Type
}

// Options contains repository options
type Options struct {
	TableName  string
	CacheTTL   time.Duration
	EnableLogs bool
}

// NewRepository creates a new repository
func NewRepository(db *pgxpool.Pool, model interface{}, opts ...Options) *Repository {
	var options Options
	if len(opts) > 0 {
		options = opts[0]
	}

	if options.TableName == "" {
		// Extract table name from model type
		t := reflect.TypeOf(model)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		options.TableName = strings.ToLower(t.Name()) + "s"
	}

	repo := &Repository{
		db:        db,
		cache:     NewCache(options.CacheTTL),
		tableName: options.TableName,
		modelType: reflect.TypeOf(model),
	}

	return repo
}

// ========== CRUD OPERATIONS ==========

// Create inserts a new record
func (r *Repository) Create(ctx context.Context, data interface{}) (interface{}, error) {
	fields, values, placeholders := r.extractFieldsAndValues(data, true)

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) RETURNING *",
		r.tableName, fields, placeholders,
	)

	var result interface{}
	err := r.db.QueryRow(ctx, query, values...).Scan(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to create record: %w", err)
	}

	// Invalidate cache
	r.cache.Clear()

	return result, nil
}

// FindByID finds a record by ID
func (r *Repository) FindByID(ctx context.Context, id interface{}) (interface{}, error) {
	// Try cache first
	if cached, found := r.cache.Get(fmt.Sprintf("id:%v", id)); found {
		return cached, nil
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1", r.tableName)

	var result interface{}
	err := r.db.QueryRow(ctx, query, id).Scan(&result)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to find by ID: %w", err)
	}

	// Cache the result
	r.cache.Set(fmt.Sprintf("id:%v", id), result)

	return result, nil
}

// FindOne finds one record matching conditions
func (r *Repository) FindOne(ctx context.Context, conditions map[string]interface{}) (interface{}, error) {
	whereClause, args := r.buildWhereClause(conditions)

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s LIMIT 1", r.tableName, whereClause)

	var result interface{}
	err := r.db.QueryRow(ctx, query, args...).Scan(&result)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to find one: %w", err)
	}

	return result, nil
}

// FindAll finds all records
func (r *Repository) FindAll(ctx context.Context, opts ...QueryOptions) ([]interface{}, error) {
	options := QueryOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}

	query := fmt.Sprintf("SELECT * FROM %s", r.tableName)

	// Add WHERE clause if conditions exist
	if len(options.Conditions) > 0 {
		whereClause, args := r.buildWhereClause(options.Conditions)
		query += " WHERE " + whereClause
		options.Args = args
	}

	// Add ORDER BY
	if options.OrderBy != "" {
		query += " ORDER BY " + options.OrderBy
	}

	// Add LIMIT and OFFSET
	if options.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", options.Limit)
	}
	if options.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", options.Offset)
	}

	rows, err := r.db.Query(ctx, query, options.Args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find all: %w", err)
	}
	defer rows.Close()

	var results []interface{}
	for rows.Next() {
		var result interface{}
		if err := rows.Scan(&result); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// Update updates a record
func (r *Repository) Update(ctx context.Context, id interface{}, data interface{}) (interface{}, error) {
	fields, values, _ := r.extractFieldsAndValues(data, false)

	// Build SET clause
	var setClauses []string
	for i, field := range strings.Split(fields, ",") {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", strings.TrimSpace(field), i+1))
	}

	// Add ID as last parameter
	values = append(values, id)

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = $%d RETURNING *",
		r.tableName, strings.Join(setClauses, ", "), len(values),
	)

	var result interface{}
	err := r.db.QueryRow(ctx, query, values...).Scan(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to update record: %w", err)
	}

	// Invalidate cache
	r.cache.Delete(fmt.Sprintf("id:%v", id))
	r.cache.Clear()

	return result, nil
}

// Delete deletes a record
func (r *Repository) Delete(ctx context.Context, id interface{}) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", r.tableName)

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	// Invalidate cache
	r.cache.Delete(fmt.Sprintf("id:%v", id))
	r.cache.Clear()

	return nil
}

// ========== BATCH OPERATIONS ==========

// CreateMany inserts multiple records
func (r *Repository) CreateMany(ctx context.Context, data []interface{}) ([]interface{}, error) {
	if len(data) == 0 {
		return []interface{}{}, nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var results []interface{}
	for _, item := range data {
		fields, values, placeholders := r.extractFieldsAndValues(item, true)

		query := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s) RETURNING *",
			r.tableName, fields, placeholders,
		)

		var result interface{}
		err := tx.QueryRow(ctx, query, values...).Scan(&result)
		if err != nil {
			return nil, fmt.Errorf("failed to create record: %w", err)
		}
		results = append(results, result)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Invalidate cache
	r.cache.Clear()

	return results, nil
}

// UpdateMany updates multiple records
func (r *Repository) UpdateMany(ctx context.Context, updates map[interface{}]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for id, data := range updates {
		fields, values, _ := r.extractFieldsAndValues(data, false)

		// Build SET clause
		var setClauses []string
		for i, field := range strings.Split(fields, ",") {
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d", strings.TrimSpace(field), i+1))
		}

		// Add ID as last parameter
		values = append(values, id)

		query := fmt.Sprintf(
			"UPDATE %s SET %s WHERE id = $%d",
			r.tableName, strings.Join(setClauses, ", "), len(values),
		)

		_, err := tx.Exec(ctx, query, values...)
		if err != nil {
			return fmt.Errorf("failed to update record: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Invalidate cache
	r.cache.Clear()

	return nil
}

// ========== QUERY OPERATIONS ==========

// Count counts records
func (r *Repository) Count(ctx context.Context, conditions ...map[string]interface{}) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", r.tableName)

	var args []interface{}
	if len(conditions) > 0 && len(conditions[0]) > 0 {
		whereClause, whereArgs := r.buildWhereClause(conditions[0])
		query += " WHERE " + whereClause
		args = whereArgs
	}

	var count int64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count records: %w", err)
	}

	return count, nil
}

// Exists checks if a record exists
func (r *Repository) Exists(ctx context.Context, conditions map[string]interface{}) (bool, error) {
	count, err := r.Count(ctx, conditions)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// FindByField finds records by field value
func (r *Repository) FindByField(ctx context.Context, field string, value interface{}) ([]interface{}, error) {
	return r.FindAll(ctx, QueryOptions{
		Conditions: map[string]interface{}{field: value},
	})
}

// FindByFields finds records by multiple field values
func (r *Repository) FindByFields(ctx context.Context, fields map[string]interface{}) ([]interface{}, error) {
	return r.FindAll(ctx, QueryOptions{
		Conditions: fields,
	})
}

// ========== TRANSACTION SUPPORT ==========

// Transaction executes a function within a transaction
func (r *Repository) Transaction(ctx context.Context, fn func(*Repository) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Create a new repository with transaction
	txRepo := &Repository{
		db:        &pgxpool.Pool{}, // This would need proper tx wrapper
		cache:     r.cache,
		tableName: r.tableName,
		modelType: r.modelType,
	}

	if err := fn(txRepo); err != nil {
		tx.Rollback(ctx)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ========== UTILITY METHODS ==========

// GetTableName returns the table name
func (r *Repository) GetTableName() string {
	return r.tableName
}

// GetDB returns the database connection
func (r *Repository) GetDB() *pgxpool.Pool {
	return r.db
}

// GetCache returns the cache instance
func (r *Repository) GetCache() *Cache {
	return r.cache
}

// ClearCache clears the cache
func (r *Repository) ClearCache() {
	r.cache.Clear()
}

// Ping checks database connection
func (r *Repository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

// ========== PRIVATE HELPER METHODS ==========

// extractFieldsAndValues extracts fields and values from a struct
func (r *Repository) extractFieldsAndValues(data interface{}, includeID bool) (string, []interface{}, string) {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	var fields []string
	var values []interface{}
	var placeholders []string

	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		value := v.Field(i).Interface()
		dbTag := field.Tag.Get("db")

		if dbTag == "" {
			dbTag = strings.ToLower(field.Name)
		}

		// Skip ID field if not included
		if dbTag == "id" && !includeID {
			continue
		}

		fields = append(fields, dbTag)
		values = append(values, value)
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(values)))
	}

	return strings.Join(fields, ", "), values, strings.Join(placeholders, ", ")
}

// buildWhereClause builds WHERE clause from conditions
func (r *Repository) buildWhereClause(conditions map[string]interface{}) (string, []interface{}) {
	var clauses []string
	var args []interface{}

	i := 1
	for field, value := range conditions {
		clauses = append(clauses, fmt.Sprintf("%s = $%d", field, i))
		args = append(args, value)
		i++
	}

	return strings.Join(clauses, " AND "), args
}

// ========== ERRORS ==========

var (
	ErrNotFound    = errors.New("record not found")
	ErrInvalidData = errors.New("invalid data")
	ErrDuplicate   = errors.New("duplicate record")
)

// ========== QUERY OPTIONS ==========

// QueryOptions contains query options
type QueryOptions struct {
	Conditions map[string]interface{}
	OrderBy    string
	Limit      int
	Offset     int
	Args       []interface{}
}
