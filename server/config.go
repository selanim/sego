package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ConfigManager handles configuration management
type ConfigManager struct {
	config     *Config
	configFile string
	watcher    *ConfigWatcher
}

// ConfigWatcher watches for config file changes
type ConfigWatcher struct {
	filePath string
	lastMod  time.Time
	stop     chan bool
	onChange func(*Config)
}

// NewConfigManager creates a new config manager
func NewConfigManager(configFile string) (*ConfigManager, error) {
	cm := &ConfigManager{
		configFile: configFile,
	}

	// Load initial config
	if err := cm.Load(); err != nil {
		return nil, err
	}

	return cm, nil
}

// Load loads configuration from file
func (cm *ConfigManager) Load() error {
	if cm.configFile == "" {
		cm.config = DefaultConfig()
		return nil
	}

	// Check if file exists
	if _, err := os.Stat(cm.configFile); os.IsNotExist(err) {
		log.Printf("Config file %s not found, using defaults", cm.configFile)
		cm.config = DefaultConfig()
		return nil
	}

	// Read config file
	data, err := os.ReadFile(cm.configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	cm.config = &config
	log.Printf("Configuration loaded from %s", cm.configFile)
	return nil
}

// Save saves configuration to file
func (cm *ConfigManager) Save() error {
	if cm.configFile == "" {
		return fmt.Errorf("no config file specified")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(cm.configFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Convert to JSON
	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(cm.configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	log.Printf("Configuration saved to %s", cm.configFile)
	return nil
}

// GetConfig returns current configuration
func (cm *ConfigManager) GetConfig() *Config {
	return cm.config
}

// UpdateConfig updates configuration
func (cm *ConfigManager) UpdateConfig(updater func(*Config)) {
	updater(cm.config)
}

// Watch starts watching for config file changes
func (cm *ConfigManager) Watch(onChange func(*Config)) error {
	if cm.configFile == "" {
		return fmt.Errorf("no config file to watch")
	}

	cm.watcher = &ConfigWatcher{
		filePath: cm.configFile,
		stop:     make(chan bool),
		onChange: onChange,
	}

	// Get initial modification time
	info, err := os.Stat(cm.configFile)
	if err != nil {
		return err
	}
	cm.watcher.lastMod = info.ModTime()

	// Start watcher in goroutine
	go cm.watcher.start()

	log.Printf("Watching config file %s for changes", cm.configFile)
	return nil
}

// StopWatching stops the config watcher
func (cm *ConfigManager) StopWatching() {
	if cm.watcher != nil {
		cm.watcher.stop <- true
	}
}

// start starts the config watcher
func (cw *ConfigWatcher) start() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cw.checkForChanges()
		case <-cw.stop:
			return
		}
	}
}

// checkForChanges checks if config file has changed
func (cw *ConfigWatcher) checkForChanges() {
	info, err := os.Stat(cw.filePath)
	if err != nil {
		log.Printf("Error checking config file: %v", err)
		return
	}

	if info.ModTime().After(cw.lastMod) {
		log.Println("Config file changed, reloading...")
		cw.lastMod = info.ModTime()

		// Reload config
		data, err := os.ReadFile(cw.filePath)
		if err != nil {
			log.Printf("Error reading config file: %v", err)
			return
		}

		var config Config
		if err := json.Unmarshal(data, &config); err != nil {
			log.Printf("Error parsing config file: %v", err)
			return
		}

		// Call onChange callback
		if cw.onChange != nil {
			cw.onChange(&config)
		}
	}
}

// LoadFromEnv loads configuration from environment variables
func (cm *ConfigManager) LoadFromEnv() {
	if cm.config == nil {
		cm.config = DefaultConfig()
	}

	// Load from environment variables
	if host := os.Getenv("SERVER_HOST"); host != "" {
		cm.config.Host = host
	}

	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := time.ParseDuration(port); err == nil {
			cm.config.Port = int(p.Seconds())
		}
	}

	if env := os.Getenv("SERVER_ENV"); env != "" {
		cm.config.Environment = env
	}

	if timeout := os.Getenv("SERVER_READ_TIMEOUT"); timeout != "" {
		if t, err := time.ParseDuration(timeout); err == nil {
			cm.config.ReadTimeout = t
		}
	}

	if timeout := os.Getenv("SERVER_WRITE_TIMEOUT"); timeout != "" {
		if t, err := time.ParseDuration(timeout); err == nil {
			cm.config.WriteTimeout = t
		}
	}

	if enableCORS := os.Getenv("ENABLE_CORS"); enableCORS != "" {
		cm.config.EnableCORS = strings.ToLower(enableCORS) == "true"
	}

	if enableMetrics := os.Getenv("ENABLE_METRICS"); enableMetrics != "" {
		cm.config.EnableMetrics = strings.ToLower(enableMetrics) == "true"
	}

	log.Println("Configuration loaded from environment variables")
}

// Export exports configuration to writer
func (cm *ConfigManager) Export(w io.Writer) error {
	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

// Validate validates configuration
func (cm *ConfigManager) Validate() error {
	if cm.config == nil {
		return fmt.Errorf("configuration is nil")
	}

	// Validate port
	if cm.config.Port < 1 || cm.config.Port > 65535 {
		return fmt.Errorf("invalid port: %d", cm.config.Port)
	}

	// Validate timeouts
	if cm.config.ReadTimeout < 0 {
		return fmt.Errorf("read timeout cannot be negative")
	}

	if cm.config.WriteTimeout < 0 {
		return fmt.Errorf("write timeout cannot be negative")
	}

	if cm.config.IdleTimeout < 0 {
		return fmt.Errorf("idle timeout cannot be negative")
	}

	// Validate static directory exists if specified
	if cm.config.StaticDir != "" {
		if _, err := os.Stat(cm.config.StaticDir); os.IsNotExist(err) {
			log.Printf("Warning: Static directory %s does not exist", cm.config.StaticDir)
		}
	}

	return nil
}

// Merge merges another config into current config
func (cm *ConfigManager) Merge(other *Config) {
	if other.Host != "" {
		cm.config.Host = other.Host
	}

	if other.Port != 0 {
		cm.config.Port = other.Port
	}

	if other.Environment != "" {
		cm.config.Environment = other.Environment
	}

	if other.ReadTimeout != 0 {
		cm.config.ReadTimeout = other.ReadTimeout
	}

	if other.WriteTimeout != 0 {
		cm.config.WriteTimeout = other.WriteTimeout
	}

	if other.IdleTimeout != 0 {
		cm.config.IdleTimeout = other.IdleTimeout
	}

	// Merge boolean flags (only if explicitly set in other config)
	// For booleans, we can't distinguish between false and unset
	// So we rely on the user to set all values
}

// CreateDefaultConfigFile creates a default config file
func CreateDefaultConfigFile(path string) error {
	config := DefaultConfig()

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// ExampleConfig returns example configuration as JSON
func ExampleConfig() string {
	config := DefaultConfig()
	data, _ := json.MarshalIndent(config, "", "  ")
	return string(data)
}

// Config templates for different environments
func DevelopmentConfig() *Config {
	config := DefaultConfig()
	config.Environment = "development"
	config.EnableLogging = true
	config.EnableCORS = true
	config.ReadTimeout = 30 * time.Second
	config.WriteTimeout = 30 * time.Second
	return config
}

func ProductionConfig() *Config {
	config := DefaultConfig()
	config.Environment = "production"
	config.EnableLogging = true
	config.EnableCORS = false // Usually handled by reverse proxy
	config.ReadTimeout = 30 * time.Second
	config.WriteTimeout = 30 * time.Second
	config.IdleTimeout = 120 * time.Second
	config.ShutdownTimeout = 30 * time.Second
	return config
}

func TestingConfig() *Config {
	config := DefaultConfig()
	config.Environment = "testing"
	config.Port = 0 // Let OS choose port
	config.EnableLogging = false
	config.EnableMetrics = false
	config.EnableHealth = false
	return config
}

// Helper function to get config by environment
func GetConfigByEnvironment(env string) *Config {
	switch strings.ToLower(env) {
	case "development", "dev":
		return DevelopmentConfig()
	case "production", "prod":
		return ProductionConfig()
	case "testing", "test":
		return TestingConfig()
	default:
		return DefaultConfig()
	}
}
