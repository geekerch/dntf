package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// LegacySystemConfig holds configuration for the legacy system
type LegacySystemConfig struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

// Config holds all application configuration
type Config struct {
	Server       ServerConfig
	Database     DatabaseConfig
	NATS         NATSConfig
	Logger       LoggerConfig
	LegacySystem LegacySystemConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         int    `json:"port"`
	Host         string `json:"host"`
	ReadTimeout  int    `json:"readTimeout"`
	WriteTimeout int    `json:"writeTimeout"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type           string `json:"type"` // postgres, sqlite, sqlserver
	Host           string `json:"host"`
	Port           int    `json:"port"`
	User           string `json:"user"`
	Password       string `json:"password"`
	DBName         string `json:"dbName"`
	Schema         string `json:"schema"`
	SSLMode        string `json:"sslMode"`
	MaxOpenConns   int    `json:"maxOpenConns"`
	MaxIdleConns   int    `json:"maxIdleConns"`
	MaxLifetime    int    `json:"maxLifetime"` // in minutes
	MigrationsPath string `json:"migrationsPath"`
}

// NATSConfig holds NATS configuration
type NATSConfig struct {
	URL            string `json:"url"`
	CredsPath      string `json:"credsPath"`
	MaxReconnects  int    `json:"maxReconnects"`
	ReconnectWait  int    `json:"reconnectWait"`  // in seconds
	RequestTimeout int    `json:"requestTimeout"` // in seconds
	SubjectPrefix  string `json:"subjectPrefix"`
}

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"` // json or console
	OutputPath string `json:"outputPath"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if exists
	_ = godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Port:         getEnvAsInt("SERVER_PORT", 8080),
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:  getEnvAsInt("SERVER_READ_TIMEOUT", 30),
			WriteTimeout: getEnvAsInt("SERVER_WRITE_TIMEOUT", 30),
		},
		Database: DatabaseConfig{
			Type:           getEnv("DB_TYPE", "postgres"),
			Host:           getEnv("DB_HOST", "localhost"),
			Port:           getEnvAsInt("DB_PORT", 5432),
			User:           getEnv("DB_USER", "postgres"),
			Password:       getEnv("DB_PASSWORD", ""),
			DBName:         getEnv("DB_NAME", "channel_api"),
			Schema:         getEnv("DB_SCHEMA", "public"),
			SSLMode:        getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:   getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:   getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			MaxLifetime:    getEnvAsInt("DB_MAX_LIFETIME", 5),
			MigrationsPath: getEnv("DB_MIGRATIONS_PATH", "migrations"),
		},
		NATS: NATSConfig{
			URL:            getEnv("NATS_URL", "nats://localhost:4222"),
			CredsPath:      getEnv("NATS_CREDS_PATH", ""),
			MaxReconnects:  getEnvAsInt("NATS_MAX_RECONNECTS", 10),
			ReconnectWait:  getEnvAsInt("NATS_RECONNECT_WAIT", 2),
			RequestTimeout: getEnvAsInt("NATS_REQUEST_TIMEOUT", 30),
			SubjectPrefix:  getEnv("NATS_SUBJECT_PREFIX", "eco1j.infra.eventcenter"),
		},
		Logger: LoggerConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Format:     getEnv("LOG_FORMAT", "json"),
			OutputPath: getEnv("LOG_OUTPUT_PATH", "stdout"),
		},
		LegacySystem: LegacySystemConfig{
			URL:   getEnv("LEGACY_SYSTEM_URL", ""),
			Token: getEnv("LEGACY_SYSTEM_TOKEN", ""),
		},
	}

	// Validate required fields
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate database type
	validDBTypes := map[string]bool{
		"postgres":   true,
		"postgresql": true,
		"sqlite":     true,
		"sqlserver":  true,
		"mssql":      true,
	}
	if !validDBTypes[c.Database.Type] {
		return fmt.Errorf("unsupported database type: %s", c.Database.Type)
	}

	// For non-SQLite databases, password is required
	if c.Database.Type != "sqlite" && c.Database.Password == "" {
		return fmt.Errorf("database password is required for %s", c.Database.Type)
	}

	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	// For non-SQLite databases, validate port
	if c.Database.Type != "sqlite" && (c.Database.Port <= 0 || c.Database.Port > 65535) {
		return fmt.Errorf("invalid database port: %d", c.Database.Port)
	}

	return nil
}

// GetDatabaseConnectionString returns the database connection string
func (c *Config) GetDatabaseConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

// GetServerAddress returns the server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer with a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
