package examples

import (
	"context"
	"fmt"
	"github.com/selanim/sego/logger"
	"os"
	"time"
)

func ExampleBasicLogging() {
	// Create a new logger with default configuration
	log := logger.New()

	// Set log level
	log.SetLevel(logger.DEBUG)

	// Basic logging
	log.Debug("Debug message")
	log.Info("Info message")
	log.Warn("Warning message")
	log.Error("Error message")

	// Formatted logging
	log.Infof("User %s logged in at %v", "john", time.Now())

	// With fields (structured logging)
	log.WithField("user_id", 123).
		WithField("action", "login").
		Info("User action performed")

	// With multiple fields
	fields := map[string]interface{}{
		"user_id":    456,
		"ip_address": "192.168.1.1",
		"user_agent": "Mozilla/5.0",
	}
	log.WithFields(fields).Info("Request received")
}

func ExampleFileLogging() {
	// Create file handler with rotation (10MB max, keep 5 files)
	fileHandler, err := logger.NewFileHandler("logs/app.log", 10*1024*1024, 5)
	if err != nil {
		panic(err)
	}
	defer fileHandler.Close()

	// Create multi-writer to log to both console and file
	multiWriter := logger.NewMultiWriter(os.Stdout, fileHandler)

	// Configure logger
	config := logger.Config{
		Level:      logger.INFO,
		Output:     multiWriter,
		WithTime:   true,
		TimeFormat: "2006-01-02 15:04:05",
		Color:      true,
	}

	log := logger.NewWithConfig(config)

	// Log to both console and file
	log.Info("This will be logged to both console and file")

	// Add more writers
	errorFile, _ := os.OpenFile("logs/error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	multiWriter.AddWriter(errorFile)

	log.Error("This will also go to error.log")
}

func ExampleJSONLogging() {
	// Configure logger for JSON output
	config := logger.Config{
		Level:    logger.INFO,
		Output:   os.Stdout,
		JSON:     true,
		WithTime: true,
	}

	log := logger.NewWithConfig(config)

	// JSON formatted logs
	log.WithField("user", "john").
		WithField("action", "purchase").
		WithField("amount", 99.99).
		Info("Purchase completed")

	// Output will be:
	// {"timestamp":"2024-01-15T10:30:00Z","level":"INFO","message":"Purchase completed","user":"john","action":"purchase","amount":99.99}
}

func ExampleContextLogging() {
	ctx := context.Background()

	// Create request-scoped logger
	ctx = logger.RequestLogger(ctx, "req-12345")
	ctx = logger.WithField(ctx, "user_id", 789)
	ctx = logger.WithField(ctx, "endpoint", "/api/users")

	// Use context logger
	logger.CtxInfo(ctx, "Request started")
	logger.CtxDebugf(ctx, "Processing user %d", 789)

	// Or get logger from context
	log := logger.FromContext(ctx)
	log.WithField("duration_ms", 150).Info("Request completed")
}

func ExampleAsyncLogging() {
	log := logger.New()
	asyncLog := logger.NewAsyncLogger(log, 1000) // Buffer size 1000

	// These won't block
	for i := 0; i < 100; i++ {
		asyncLog.WithField("iteration", i).Infof("Processing item %d", i)
	}

	// Shutdown gracefully
	asyncLog.Shutdown()
}

func ExampleDailyLogging() {
	// Create daily file handler
	dailyHandler, err := logger.NewDailyFileHandler("logs/app")
	if err != nil {
		panic(err)
	}
	defer dailyHandler.Close()

	config := logger.Config{
		Level:  logger.INFO,
		Output: dailyHandler,
	}

	log := logger.NewWithConfig(config)

	// This will create logs like app-2024-01-15.log, app-2024-01-16.log, etc.
	log.Info("Daily log entry")
}

func ExampleCustomFormatter() {
	// Create custom formatter
	customFormatter := &logger.TextFormatter{
		Color:      true,
		TimeFormat: "15:04:05",
	}

	config := logger.Config{
		Level:      logger.DEBUG,
		Output:     os.Stdout,
		Formatter:  customFormatter,
		WithCaller: true,
	}

	log := logger.NewWithConfig(config)

	log.Debug("Debug with caller info")
	log.Error("Error with caller info")
}

func ExampleLogLevels() {
	log := logger.New()

	// Set different log levels
	log.SetLevel(logger.DEBUG)
	log.Debug("This will show") // Shows
	log.Info("This will show")  // Shows
	log.Warn("This will show")  // Shows

	log.SetLevel(logger.WARN)
	log.Debug("This won't show") // Hidden
	log.Info("This won't show")  // Hidden
	log.Warn("This will show")   // Shows
	log.Error("This will show")  // Shows

	// Parse level from string
	level := logger.ParseLevel("ERROR")
	log.SetLevel(level)
}

func ExampleIntegrationWithRepo() {
	// Create a logger
	log := logger.New()
	log.SetLevel(logger.DEBUG)

	// Create a repository with logger prefix
	repoLogger := log.WithPrefix("UserRepository")

	// Simulate repository methods
	findUserByID := func(ctx context.Context, id int) (interface{}, error) {
		repoLogger.WithField("user_id", id).Debug("Finding user by ID")

		// Add request ID from context if available
		if reqID, ok := ctx.Value("request_id").(string); ok {
			repoLogger = repoLogger.WithField("request_id", reqID)
		}

		// Simulate database operation
		time.Sleep(10 * time.Millisecond)

		repoLogger.WithField("duration_ms", 45).Info("User found")
		return map[string]interface{}{"id": id, "name": "John Doe"}, nil
	}

	createUser := func(ctx context.Context, user interface{}) error {
		repoLogger.WithFields(map[string]interface{}{
			"operation": "create",
			"user":      user,
		}).Info("Creating user")

		// Simulate database operation
		time.Sleep(15 * time.Millisecond)

		repoLogger.Info("User created successfully")
		return nil
	}

	// Create context with request ID
	ctx := context.WithValue(context.Background(), "request_id", "req-12345")

	// Use repository methods
	user, err := findUserByID(ctx, 123)
	if err != nil {
		log.WithField("error", err).Error("Failed to find user")
	} else {
		log.WithField("user", user).Debug("Retrieved user")
	}

	// Create a new user
	newUser := map[string]interface{}{
		"name":  "Jane Doe",
		"email": "jane@example.com",
	}

	if err := createUser(ctx, newUser); err != nil {
		log.WithField("error", err).Error("Failed to create user")
	}
}

func ExampleErrorHandling() {
	log := logger.New()

	// Simulate an error
	err := fmt.Errorf("database connection failed")

	// Log error with stack trace or additional context
	log.WithField("error", err).
		WithField("retry_count", 3).
		Error("Failed to connect to database")

	// Note: Uncomment these for demonstration only
	// Fatal logs and exits
	// log.Fatal("Critical error, exiting...")

	// Panic logs and panics
	// log.Panic("Something went terribly wrong")
}

func ExamplePerformanceLogging() {
	log := logger.New().WithPrefix("Performance")

	start := time.Now()

	// Simulate some work
	time.Sleep(100 * time.Millisecond)

	duration := time.Since(start)

	log.WithField("duration_ms", duration.Milliseconds()).
		WithField("operation", "complex_calculation").
		Info("Operation completed")
}

// UserRepositoryExample demonstrates logger integration with repository
type UserRepositoryExample struct {
	logger *logger.Logger
}

// NewUserRepositoryExample creates a new user repository example
func NewUserRepositoryExample(log *logger.Logger) *UserRepositoryExample {
	return &UserRepositoryExample{
		logger: log.WithPrefix("UserRepo"),
	}
}

// FindByID finds user by ID (example method)
func (r *UserRepositoryExample) FindByID(ctx context.Context, id int) (interface{}, error) {
	r.logger.WithField("user_id", id).Debug("Finding user by ID")

	// Add request ID from context if available
	if reqID, ok := ctx.Value("request_id").(string); ok {
		r.logger = r.logger.WithField("request_id", reqID)
	}

	// Simulate database operation
	time.Sleep(10 * time.Millisecond)

	r.logger.WithField("duration_ms", 45).Info("User found")
	return map[string]interface{}{"id": id, "name": "John Doe"}, nil
}

// Create creates a new user (example method)
func (r *UserRepositoryExample) Create(ctx context.Context, user interface{}) error {
	r.logger.WithFields(map[string]interface{}{
		"operation": "create",
		"user":      user,
	}).Info("Creating user")

	// Simulate database operation
	time.Sleep(15 * time.Millisecond)

	r.logger.Info("User created successfully")
	return nil
}

func ExampleRepositoryIntegration() {
	// Create a logger
	log := logger.New()
	log.SetLevel(logger.DEBUG)

	// Create repository with logger
	repo := NewUserRepositoryExample(log)

	// Create context with request ID
	ctx := context.WithValue(context.Background(), "request_id", "req-12345")

	// Use repository methods
	user, err := repo.FindByID(ctx, 123)
	if err != nil {
		log.WithField("error", err).Error("Failed to find user")
	} else {
		log.WithField("user", user).Debug("Retrieved user")
	}

	// Create a new user
	newUser := map[string]interface{}{
		"name":  "Jane Doe",
		"email": "jane@example.com",
	}

	if err := repo.Create(ctx, newUser); err != nil {
		log.WithField("error", err).Error("Failed to create user")
	}
}
