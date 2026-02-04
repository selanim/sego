package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"
)

// TestMain inafanya setup na cleanup
func TestMain(m *testing.M) {
	// Setup
	setupTestEnvironment()

	// Run tests
	code := m.Run()

	// Cleanup
	cleanupTestEnvironment()

	os.Exit(code)
}

func setupTestEnvironment() {
	// Create test directory
	os.MkdirAll("testdata", 0755)

	// Clean up old test files
	os.Remove("testdata/test.db")
	os.Remove(".env.test")

	// Create test .env file
	envContent := `# PostgreSQL Test Config
DB_TYPE=postgres
DB_USER=postgres
DB_PASSWORD=testpassword
DB_HOST=localhost
DB_PORT=5432
DB_NAME=testdb
DB_SSL_MODE=disable

# MySQL Test Config
# DB_TYPE=mysql
# DB_USER=root
# DB_PASSWORD=testpassword
# DB_HOST=localhost
# DB_PORT=3306
# DB_NAME=testdb

# SQLite Test Config
# DB_TYPE=sqlite
# DB_FILE_PATH=testdata/test.db

# MongoDB Test Config
# DB_TYPE=mongodb
# DB_HOST=localhost
# DB_PORT=27017
# DB_NAME=testdb`

	err := os.WriteFile(".env.test", []byte(envContent), 0644)
	if err != nil {
		log.Printf("Failed to create .env.test: %v", err)
	}
}

func cleanupTestEnvironment() {
	// Clean up test files
	os.Remove(".env.test")
	os.Remove("testdata/test.db")
	os.RemoveAll("testdata")

	// Close any open database connections
	if dbInstance != nil {
		dbInstance.Close()
	}
}

// TestGetEnv inatest getEnv helper function
func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		expected     string
	}{
		{"Existing env var", "PATH", "/default", os.Getenv("PATH")},
		{"Non-existing env var", "NON_EXISTENT_VAR", "default_value", "default_value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getEnv(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnv(%q, %q) = %q, want %q",
					tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

// TestConfigValidation inatest configuration validation
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		shouldError bool
	}{
		{
			name: "Valid PostgreSQL config",
			config: Config{
				Type:     PostgreSQL,
				User:     "testuser",
				Password: "testpass",
				Host:     "localhost",
				Port:     "5432",
				Name:     "testdb",
			},
			shouldError: false,
		},
		{
			name: "PostgreSQL without password",
			config: Config{
				Type: PostgreSQL,
				User: "testuser",
				// Password missing
				Host: "localhost",
				Port: "5432",
				Name: "testdb",
			},
			shouldError: true,
		},
		{
			name: "Valid SQLite config",
			config: Config{
				Type:     SQLite,
				FilePath: "testdata/test.db",
			},
			shouldError: false,
		},
		{
			name: "Valid MongoDB config",
			config: Config{
				Type: MongoDB,
				Host: "localhost",
				Port: "27017",
				Name: "testdb",
			},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Connect(&ConnectOptions{
				Config: &tt.config,
			})

			if tt.shouldError && err == nil {
				t.Errorf("Expected error for config %v, got nil", tt.config)
			}
			if !tt.shouldError && err != nil {
				// Connection might fail if database isn't running, but config should be valid
				// We accept connection errors for CI environments
				t.Logf("Connection error (expected for CI): %v", err)
			}
		})
	}
}

// TestSQLiteConnection inatest SQLite database connection
func TestSQLiteConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping SQLite test in short mode")
	}

	config := Config{
		Type:     SQLite,
		FilePath: "testdata/test.db",
	}

	db, err := Connect(&ConnectOptions{
		Config: &config,
	})

	if err != nil {
		t.Fatalf("Failed to connect to SQLite: %v", err)
	}
	defer db.Close()

	// Test HealthCheck
	if err := db.HealthCheck(); err != nil {
		t.Errorf("HealthCheck failed: %v", err)
	}

	// Test IsConnected
	if !db.IsConnected() {
		t.Error("IsConnected should return true")
	}

	// Test GetType
	if db.GetType() != SQLite {
		t.Errorf("GetType() = %v, want %v", db.GetType(), SQLite)
	}

	// Test Stats
	stats := db.Stats()
	if stats == "" {
		t.Error("Stats should return non-empty string")
	}
	t.Logf("SQLite Stats: %s", stats)
}

// TestSQLiteCRUD inatest CRUD operations kwa SQLite
func TestSQLiteCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping SQLite CRUD test in short mode")
	}

	config := Config{
		Type:     SQLite,
		FilePath: "testdata/test_crud.db",
	}

	db, err := Connect(&ConnectOptions{
		Config: &config,
	})

	if err != nil {
		t.Fatalf("Failed to connect to SQLite: %v", err)
	}
	defer os.Remove("testdata/test_crud.db")
	defer db.Close()

	ctx := context.Background()

	// Create table
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS test_users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

	_, err = db.ExecuteExec(ctx, createTableSQL)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert data
	insertSQL := "INSERT INTO test_users (name, email) VALUES (?, ?)"
	rowsAffected, err := db.ExecuteExec(ctx, insertSQL, "John Doe", "john@example.com")
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	if rowsAffected != 1 {
		t.Errorf("RowsAffected = %d, want 1", rowsAffected)
	}

	// Query data
	querySQL := "SELECT id, name, email FROM test_users WHERE email = ?"
	rows, err := db.ExecuteQuery(ctx, querySQL, "john@example.com")
	if err != nil {
		t.Fatalf("Failed to query data: %v", err)
	}
	defer rows.Close()

	var id int64
	var name, email string

	if rows.Next() {
		err = rows.Scan(&id, &name, &email)
		if err != nil {
			t.Fatalf("Failed to scan row: %v", err)
		}

		if name != "John Doe" || email != "john@example.com" {
			t.Errorf("Got name=%s, email=%s, want name=John Doe, email=john@example.com",
				name, email)
		}
	} else {
		t.Error("No rows returned from query")
	}

	// Update data
	updateSQL := "UPDATE test_users SET name = ? WHERE email = ?"
	rowsAffected, err = db.ExecuteExec(ctx, updateSQL, "John Updated", "john@example.com")
	if err != nil {
		t.Fatalf("Failed to update data: %v", err)
	}

	if rowsAffected != 1 {
		t.Errorf("RowsAffected after update = %d, want 1", rowsAffected)
	}

	// Delete data
	deleteSQL := "DELETE FROM test_users WHERE email = ?"
	rowsAffected, err = db.ExecuteExec(ctx, deleteSQL, "john@example.com")
	if err != nil {
		t.Fatalf("Failed to delete data: %v", err)
	}

	if rowsAffected != 1 {
		t.Errorf("RowsAffected after delete = %d, want 1", rowsAffected)
	}
}

// TestTransactions inatest transaction support
func TestTransactions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping transaction test in short mode")
	}

	config := Config{
		Type:     SQLite,
		FilePath: "testdata/test_transaction.db",
	}

	db, err := Connect(&ConnectOptions{
		Config: &config,
	})

	if err != nil {
		t.Fatalf("Failed to connect to SQLite: %v", err)
	}
	defer os.Remove("testdata/test_transaction.db")
	defer db.Close()

	ctx := context.Background()

	// Create table
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS test_accounts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		balance DECIMAL(10, 2) NOT NULL DEFAULT 0
	)`

	_, err = db.ExecuteExec(ctx, createTableSQL)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert initial data
	_, err = db.ExecuteExec(ctx, "INSERT INTO test_accounts (balance) VALUES (1000.00)")
	if err != nil {
		t.Fatalf("Failed to insert initial data: %v", err)
	}

	// Test successful transaction
	err = db.WithTransaction(ctx, func(tx interface{}) error {
		var sqlTx *sql.Tx
		var ok bool

		if sqlTx, ok = tx.(*sql.Tx); !ok {
			return fmt.Errorf("invalid transaction type")
		}

		// Update balance
		_, err := sqlTx.ExecContext(ctx, "UPDATE test_accounts SET balance = balance - 100 WHERE id = 1")
		if err != nil {
			return err
		}

		// Another update
		_, err = sqlTx.ExecContext(ctx, "UPDATE test_accounts SET balance = balance + 100 WHERE id = 1")
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Transaction failed: %v", err)
	}

	// Test failed transaction (rollback)
	err = db.WithTransaction(ctx, func(tx interface{}) error {
		var sqlTx *sql.Tx
		var ok bool

		if sqlTx, ok = tx.(*sql.Tx); !ok {
			return fmt.Errorf("invalid transaction type")
		}

		// First update should succeed
		_, err := sqlTx.ExecContext(ctx, "UPDATE test_accounts SET balance = balance - 100 WHERE id = 1")
		if err != nil {
			return err
		}

		// Second update should fail
		_, err = sqlTx.ExecContext(ctx, "UPDATE test_accounts SET balance = balance + 'invalid' WHERE id = 1")
		if err != nil {
			return err
		}

		return nil
	})

	if err == nil {
		t.Error("Expected transaction to fail, but it succeeded")
	}
}

// TestSingletonPattern inatest singleton pattern ya Connect function
func TestSingletonPattern(t *testing.T) {
	// Reset global instance for this test
	dbInstance = nil
	once = sync.Once{}

	config := Config{
		Type:     SQLite,
		FilePath: "testdata/test_singleton.db",
	}

	// First connection
	db1, err := Connect(&ConnectOptions{
		Config: &config,
	})
	if err != nil {
		t.Fatalf("First connection failed: %v", err)
	}
	defer os.Remove("testdata/test_singleton.db")

	// Second connection should return same instance
	db2, err := Connect(&ConnectOptions{
		Config: &config,
	})
	if err != nil {
		t.Fatalf("Second connection failed: %v", err)
	}

	if db1 != db2 {
		t.Error("Connect should return the same instance (singleton pattern)")
	}

	db1.Close()
}

// TestMustConnect inatest MustConnect function
func TestMustConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping MustConnect test in short mode")
	}

	// Test with invalid config (should panic)
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustConnect should panic with invalid config")
		}
	}()

	invalidConfig := Config{
		Type:     PostgreSQL,
		User:     "invalid",
		Password: "",
		Host:     "localhost",
		Port:     "5432",
		Name:     "nonexistent",
	}

	_ = MustConnect(&ConnectOptions{
		Config: &invalidConfig,
	})
}

// TestConnectionPoolSettings inatest connection pool settings
func TestConnectionPoolSettings(t *testing.T) {
	config := Config{
		Type:     SQLite,
		FilePath: "testdata/test_pool.db",
		MaxConns: 5,
		MinConns: 2,
	}

	db, err := Connect(&ConnectOptions{
		Config: &config,
	})

	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer os.Remove("testdata/test_pool.db")
	defer db.Close()

	stats := db.Stats()
	if stats == "" {
		t.Error("Stats should return non-empty string")
	}

	t.Logf("Connection pool stats: %s", stats)
}

// TestContextCancellation inatest context cancellation
func TestContextCancellation(t *testing.T) {
	config := Config{
		Type:     SQLite,
		FilePath: "testdata/test_context.db",
	}

	db, err := Connect(&ConnectOptions{
		Config: &config,
	})

	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer os.Remove("testdata/test_context.db")
	defer db.Close()

	// Create table
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS test_context (
		id INTEGER PRIMARY KEY,
		data TEXT
	)`

	ctx := context.Background()
	_, err = db.ExecuteExec(ctx, createTableSQL)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Test with cancelled context
	cancelledCtx, cancel := context.WithCancel(ctx)
	cancel() // Cancel immediately

	_, err = db.ExecuteExec(cancelledCtx, "INSERT INTO test_context (data) VALUES ('test')")
	if err == nil {
		t.Error("Expected error with cancelled context")
	}

	// Test with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
	defer cancel()
	time.Sleep(2 * time.Nanosecond) // Ensure timeout expires

	_, err = db.ExecuteExec(timeoutCtx, "INSERT INTO test_context (data) VALUES ('test')")
	if err == nil {
		t.Error("Expected error with timeout context")
	}
}

// TestMultipleConnections inatest multiple database connections
func TestMultipleConnections(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping multiple connections test in short mode")
	}

	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "SQLite connection 1",
			config: Config{
				Type:     SQLite,
				FilePath: "testdata/multi1.db",
			},
		},
		{
			name: "SQLite connection 2",
			config: Config{
				Type:     SQLite,
				FilePath: "testdata/multi2.db",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset for each test
			dbInstance = nil
			once = sync.Once{}

			db, err := Connect(&ConnectOptions{
				Config: &tt.config,
			})

			if err != nil {
				t.Errorf("Failed to connect %s: %v", tt.name, err)
				return
			}
			defer os.Remove(tt.config.FilePath)
			defer db.Close()

			if !db.IsConnected() {
				t.Errorf("%s should be connected", tt.name)
			}
		})
	}
}

// TestEnvFileLoading inatest .env file loading
func TestEnvFileLoading(t *testing.T) {
	// Create a test .env file
	testEnvContent := `
DB_TYPE=sqlite
DB_FILE_PATH=testdata/env_test.db
TEST_VAR=test_value
`

	envFilePath := "testdata/.env.test"
	err := os.WriteFile(envFilePath, []byte(testEnvContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}
	defer os.Remove(envFilePath)
	defer os.Remove("testdata/env_test.db")

	db, err := Connect(&ConnectOptions{
		EnvFile: envFilePath,
	})

	if err != nil {
		// Connection might fail for other reasons, but env file should load
		t.Logf("Connection error (env file loaded): %v", err)
	}

	// Check if test variable was loaded
	if os.Getenv("TEST_VAR") != "test_value" {
		t.Error("Env variable not loaded from file")
	}

	if db != nil {
		db.Close()
	}
}

// TestErrorHandling inatest error handling
func TestErrorHandling(t *testing.T) {
	// Test with nil database instance
	if dbInstance != nil {
		dbInstance.Close()
		dbInstance = nil
	}

	// GetDB should panic when no connection
	defer func() {
		if r := recover(); r == nil {
			t.Error("GetDB should panic when no connection")
		}
	}()

	_ = GetDB()
}

// BenchmarkSQLiteQuery benchmarks SQLite query performance
func BenchmarkSQLiteQuery(b *testing.B) {
	config := Config{
		Type:     SQLite,
		FilePath: "testdata/benchmark.db",
	}

	db, err := Connect(&ConnectOptions{
		Config: &config,
	})
	if err != nil {
		b.Fatalf("Failed to connect: %v", err)
	}
	defer os.Remove("testdata/benchmark.db")
	defer db.Close()

	ctx := context.Background()

	// Create table
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS benchmark_data (
		id INTEGER PRIMARY KEY,
		value TEXT
	)`

	_, err = db.ExecuteExec(ctx, createTableSQL)
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// Insert test data
	for i := 0; i < 1000; i++ {
		_, err = db.ExecuteExec(ctx, "INSERT INTO benchmark_data (value) VALUES (?)",
			fmt.Sprintf("value_%d", i))
		if err != nil {
			b.Fatalf("Failed to insert data: %v", err)
		}
	}

	b.ResetTimer()

	// Benchmark queries
	for i := 0; i < b.N; i++ {
		rows, err := db.ExecuteQuery(ctx, "SELECT id, value FROM benchmark_data WHERE id % 10 = ?", i%10)
		if err != nil {
			b.Fatalf("Query failed: %v", err)
		}
		rows.Close()
	}
}
