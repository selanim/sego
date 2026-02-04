package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"database/sql"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL driver
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	_ "modernc.org/sqlite" // SQLite driver
)

// DBType inafafanua aina za database zinazosupportiwa
type DBType string

const (
	PostgreSQL DBType = "postgres"
	MySQL      DBType = "mysql"
	SQLite     DBType = "sqlite"
	MongoDB    DBType = "mongodb"
)

// DB ni interface ya generic database
type DB struct {
	Type         DBType
	SQLDB        *sql.DB
	PostgresPool *pgxpool.Pool
	MongoClient  *mongo.Client
	MongoDB      *mongo.Database
}

// Database globals
var (
	dbInstance *DB
	once       sync.Once
	dbType     DBType
)

// Config ina configuration za database
type Config struct {
	Type     DBType
	User     string
	Password string
	Host     string
	Port     string
	Name     string
	MaxConns int32
	MinConns int32
	SSLMode  string
	FilePath string // For SQLite
}

// ConnectOptions inaweza kubadilisha default settings
type ConnectOptions struct {
	Config  *Config
	EnvFile string
}

// Connect huanzisha database connection (singleton pattern)
func Connect(options ...*ConnectOptions) (*DB, error) {
	var err error
	once.Do(func() {
		err = initDB(options...)
	})

	if err != nil {
		return nil, err
	}

	return dbInstance, nil
}

// initDB inafanya actual initialization
func initDB(options ...*ConnectOptions) error {
	var opt *ConnectOptions
	if len(options) > 0 && options[0] != nil {
		opt = options[0]
	} else {
		opt = &ConnectOptions{}
	}

	// Load .env file
	if opt.EnvFile != "" {
		if err := godotenv.Load(opt.EnvFile); err != nil {
			log.Printf("Warning: .env file not found at %s: %v", opt.EnvFile, err)
		}
	} else {
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: .env file not found: %v", err)
		}
	}

	// Get configuration
	var config Config
	if opt.Config != nil {
		config = *opt.Config
	} else {
		// Default to PostgreSQL
		config = Config{
			Type:     DBType(getEnv("DB_TYPE", "postgres")),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			Name:     getEnv("DB_NAME", "testdb"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
			MaxConns: 10,
			MinConns: 2,
			FilePath: getEnv("DB_FILE_PATH", ""), // For SQLite
		}
	}

	dbType = config.Type
	dbInstance = &DB{Type: config.Type}

	// Connect based on database type
	switch config.Type {
	case PostgreSQL:
		return connectPostgreSQL(config)
	case MySQL:
		return connectMySQL(config)
	case SQLite:
		return connectSQLite(config)
	case MongoDB:
		return connectMongoDB(config)
	default:
		return fmt.Errorf("unsupported database type: %s", config.Type)
	}
}

// connectPostgreSQL inaunganisha na PostgreSQL
func connectPostgreSQL(config Config) error {
	// Validate required variables
	if config.Password == "" {
		return fmt.Errorf("DB_PASSWORD is required for PostgreSQL")
	}

	// Build DSN
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&pool_max_conns=%d&pool_min_conns=%d",
		config.User, config.Password, config.Host, config.Port, config.Name,
		config.SSLMode, config.MaxConns, config.MinConns)

	// Create connection pool configuration
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("failed to parse PostgreSQL config: %w", err)
	}

	// Set connection pool settings
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = time.Minute

	// Connect to database with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("unable to reach PostgreSQL database: %w", err)
	}

	dbInstance.PostgresPool = pool
	log.Printf("✅ Connected to PostgreSQL at %s:%s/%s", config.Host, config.Port, config.Name)
	log.Printf("   Connection pool: %d min, %d max connections", config.MinConns, config.MaxConns)

	return nil
}

// connectMySQL inaunganisha na MySQL
func connectMySQL(config Config) error {
	// Build DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		config.User, config.Password, config.Host, config.Port, config.Name)

	if config.SSLMode != "" && config.SSLMode != "disable" {
		dsn += "&tls=" + config.SSLMode
	}

	// Connect to MySQL
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open MySQL connection: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(int(config.MaxConns))
	db.SetMaxIdleConns(int(config.MinConns))
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(30 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("unable to reach MySQL database: %w", err)
	}

	dbInstance.SQLDB = db
	log.Printf("✅ Connected to MySQL at %s:%s/%s", config.Host, config.Port, config.Name)
	log.Printf("   Connection pool: %d min, %d max connections", config.MinConns, config.MaxConns)

	return nil
}

// connectSQLite inaunganisha na SQLite
func connectSQLite(config Config) error {
	filePath := config.FilePath
	if filePath == "" {
		filePath = config.Name + ".db"
	}

	// Build DSN
	dsn := fmt.Sprintf("file:%s?_journal=WAL&_timeout=5000", filePath)

	// Connect to SQLite
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("failed to open SQLite connection: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(1) // SQLite ina limitations kwa concurrency
	db.SetMaxIdleConns(1)

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("unable to reach SQLite database: %w", err)
	}

	dbInstance.SQLDB = db
	log.Printf("✅ Connected to SQLite at %s", filePath)

	return nil
}

// connectMongoDB inaunganisha na MongoDB
func connectMongoDB(config Config) error {
	// Build connection URI
	var uri string
	if config.User != "" && config.Password != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%s/%s",
			config.User, config.Password, config.Host, config.Port, config.Name)
	} else {
		uri = fmt.Sprintf("mongodb://%s:%s/%s",
			config.Host, config.Port, config.Name)
	}

	// Set client options
	clientOptions := options.Client().ApplyURI(uri)
	clientOptions.SetMaxPoolSize(uint64(config.MaxConns))
	clientOptions.SetMinPoolSize(uint64(config.MinConns))
	clientOptions.SetConnectTimeout(15 * time.Second)

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test connection
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("unable to reach MongoDB: %w", err)
	}

	dbInstance.MongoClient = client
	dbInstance.MongoDB = client.Database(config.Name)
	log.Printf("✅ Connected to MongoDB at %s:%s/%s", config.Host, config.Port, config.Name)
	log.Printf("   Connection pool: %d min, %d max connections", config.MinConns, config.MaxConns)

	return nil
}

// MustConnect inafanya connection na kufail kama kuna error
func MustConnect(options ...*ConnectOptions) *DB {
	db, err := Connect(options...)
	if err != nil {
		log.Fatalf("FATAL: Database connection failed: %v", err)
	}
	return db
}

// Close inafunga database connection
func (db *DB) Close() {
	if db == nil {
		return
	}

	switch db.Type {
	case PostgreSQL:
		if db.PostgresPool != nil {
			db.PostgresPool.Close()
		}
	case MySQL, SQLite:
		if db.SQLDB != nil {
			db.SQLDB.Close()
		}
	case MongoDB:
		if db.MongoClient != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			db.MongoClient.Disconnect(ctx)
		}
	}
	log.Println("Database connection closed")
}

// GetDB inarudisha database instance
func GetDB() *DB {
	if dbInstance == nil {
		log.Fatal("Database not connected. Call Connect() first")
	}
	return dbInstance
}

// HealthCheck inaangalia kama database iko live
func (db *DB) HealthCheck() error {
	if db == nil {
		return fmt.Errorf("database not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	switch db.Type {
	case PostgreSQL:
		if db.PostgresPool != nil {
			return db.PostgresPool.Ping(ctx)
		}
	case MySQL, SQLite:
		if db.SQLDB != nil {
			return db.SQLDB.PingContext(ctx)
		}
	case MongoDB:
		if db.MongoClient != nil {
			return db.MongoClient.Ping(ctx, nil)
		}
	}
	return fmt.Errorf("database not properly initialized")
}

// Stats inarudisha connection pool statistics
func (db *DB) Stats() string {
	if db == nil {
		return "Database not connected"
	}

	switch db.Type {
	case PostgreSQL:
		if db.PostgresPool != nil {
			stats := db.PostgresPool.Stat()
			return fmt.Sprintf(
				"PostgreSQL Connections: %d total, %d idle, %d max, Acquires: %d",
				stats.TotalConns(), stats.IdleConns(), stats.MaxConns(), stats.AcquireCount(),
			)
		}
	case MySQL, SQLite:
		if db.SQLDB != nil {
			stats := db.SQLDB.Stats()
			return fmt.Sprintf(
				"%s Connections: %d open, %d in use, %d idle, WaitCount: %d",
				db.Type, stats.OpenConnections, stats.InUse, stats.Idle, stats.WaitCount,
			)
		}
	case MongoDB:
		return "MongoDB connection active"
	}
	return "Database stats not available"
}

// ExecuteQuery inafanya query rahisi kwa SQL databases
func (db *DB) ExecuteQuery(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	switch db.Type {
	case PostgreSQL:
		if db.PostgresPool != nil {
			// Kwa PostgreSQL, tunaweza kutumia pgx.Rows moja kwa moja
			// Au kureturn error tukikosa converter
			return nil, fmt.Errorf("PostgreSQL ExecuteQuery: use QueryRow or ScanRows instead for pgx")
		}
	case MySQL, SQLite:
		if db.SQLDB != nil {
			return db.SQLDB.QueryContext(ctx, query, args...)
		}
	}
	return nil, fmt.Errorf("ExecuteQuery not supported for database type: %s", db.Type)
}

// ExecuteQueryRows inafanya query na kureturn pgx.Rows kwa PostgreSQL
func (db *DB) ExecuteQueryRows(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	if db.Type == PostgreSQL && db.PostgresPool != nil {
		return db.PostgresPool.Query(ctx, query, args...)
	}
	return nil, fmt.Errorf("ExecuteQueryRows only supported for PostgreSQL")
}

// QueryRow inafanya query na kureturn single row
func (db *DB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if db.Type == MySQL || db.Type == SQLite {
		if db.SQLDB != nil {
			return db.SQLDB.QueryRowContext(ctx, query, args...)
		}
	}
	return nil
}

// QueryRowPgx inafanya query na kureturn single row kwa PostgreSQL
func (db *DB) QueryRowPgx(ctx context.Context, query string, args ...interface{}) pgx.Row {
	if db.Type == PostgreSQL && db.PostgresPool != nil {
		return db.PostgresPool.QueryRow(ctx, query, args...)
	}
	return nil
}

// ScanRows inafanya scan za results kutoka pgx.Rows
func (db *DB) ScanRows(rows pgx.Rows, dest ...interface{}) error {
	if db.Type != PostgreSQL {
		return fmt.Errorf("ScanRows only supported for PostgreSQL")
	}

	// Simple scanning - in production, unaweza kuwa na complex scanning logic
	return rows.Scan(dest...)
}

// ScanSQLRows inafanya scan za results kutoka sql.Rows
func (db *DB) ScanSQLRows(rows *sql.Rows, dest ...interface{}) error {
	return rows.Scan(dest...)
}

// ExecuteExec inafanya exec command kwa SQL databases (INSERT, UPDATE, DELETE)
func (db *DB) ExecuteExec(ctx context.Context, query string, args ...interface{}) (int64, error) {
	switch db.Type {
	case PostgreSQL:
		if db.PostgresPool != nil {
			result, err := db.PostgresPool.Exec(ctx, query, args...)
			if err != nil {
				return 0, err
			}
			// pgx inarudisha int64 tu
			return result.RowsAffected(), nil
		}
	case MySQL, SQLite:
		if db.SQLDB != nil {
			result, err := db.SQLDB.ExecContext(ctx, query, args...)
			if err != nil {
				return 0, err
			}
			// database/sql inarudisha (int64, error)
			rowsAffected, err := result.RowsAffected()
			if err != nil {
				return 0, fmt.Errorf("failed to get rows affected: %w", err)
			}
			return rowsAffected, nil
		}
	}
	return 0, fmt.Errorf("ExecuteExec not supported for database type: %s", db.Type)
}

// WithTransaction inafanya transaction safely kwa SQL databases
func (db *DB) WithTransaction(ctx context.Context, fn func(tx interface{}) error) error {
	switch db.Type {
	case PostgreSQL:
		if db.PostgresPool == nil {
			return fmt.Errorf("PostgreSQL not connected")
		}
		conn, err := db.PostgresPool.Acquire(ctx)
		if err != nil {
			return err
		}
		defer conn.Release()

		tx, err := conn.Begin(ctx)
		if err != nil {
			return err
		}

		defer func() {
			if p := recover(); p != nil {
				tx.Rollback(ctx)
				panic(p)
			}
		}()

		if err := fn(tx); err != nil {
			tx.Rollback(ctx)
			return err
		}

		return tx.Commit(ctx)

	case MySQL, SQLite:
		if db.SQLDB == nil {
			return fmt.Errorf("SQL database not connected")
		}
		tx, err := db.SQLDB.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		defer func() {
			if p := recover(); p != nil {
				tx.Rollback()
				panic(p)
			}
		}()

		if err := fn(tx); err != nil {
			tx.Rollback()
			return err
		}

		return tx.Commit()
	}
	return fmt.Errorf("transactions not supported for database type: %s", db.Type)
}

// MongoDB helpers
func (db *DB) MongoCollection(collectionName string) *mongo.Collection {
	if db.Type != MongoDB || db.MongoDB == nil {
		return nil
	}
	return db.MongoDB.Collection(collectionName)
}

// Helper function to get environment variable
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// IsConnected inaangalia kama database imeconnect
func (db *DB) IsConnected() bool {
	if db == nil {
		return false
	}
	return db.HealthCheck() == nil
}

// GetType inarudisha aina ya database
func (db *DB) GetType() DBType {
	if db == nil {
		return ""
	}
	return db.Type
}
