package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"notification/internal/infrastructure/models"
	"notification/pkg/config"
)

// GormDB wraps gorm.DB with additional functionality
type GormDB struct {
	*gorm.DB
	config *config.DatabaseConfig
}

// NewGormDB creates a new database connection using GORM
func NewGormDB(cfg *config.DatabaseConfig) (*GormDB, error) {
	var dialector gorm.Dialector
	var err error

	// Configure GORM logger
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// Set up dialector based on database type
	switch cfg.Type {
	case "postgres", "postgresql":
		dialector, err = createPostgresDialector(cfg)
	case "sqlite":
		dialector, err = createSQLiteDialector(cfg)
	case "sqlserver", "mssql":
		dialector, err = createSQLServerDialector(cfg)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create database dialector: %w", err)
	}

	// Open database connection
	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Minute)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &GormDB{
		DB:     db,
		config: cfg,
	}, nil
}

// createPostgresDialector creates a PostgreSQL dialector
func createPostgresDialector(cfg *config.DatabaseConfig) (gorm.Dialector, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)
	return postgres.Open(dsn), nil
}

// createSQLiteDialector creates a SQLite dialector
func createSQLiteDialector(cfg *config.DatabaseConfig) (gorm.Dialector, error) {
	// For SQLite, use DBName as the file path
	dsn := cfg.DBName
	if dsn == "" {
		dsn = "notification.db"
	}
	return sqlite.Open(dsn), nil
}

// createSQLServerDialector creates a SQL Server dialector
func createSQLServerDialector(cfg *config.DatabaseConfig) (gorm.Dialector, error) {
	dsn := fmt.Sprintf("server=%s;port=%d;user id=%s;password=%s;database=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)
	return sqlserver.Open(dsn), nil
}

// Close closes the database connection
func (db *GormDB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// RunMigrations runs GORM AutoMigrate for all models
func (db *GormDB) RunMigrations() error {
	// Run GORM AutoMigrate
	if err := models.MigrateModels(db.DB); err != nil {
		return fmt.Errorf("failed to run GORM migrations: %w", err)
	}

	// Create additional indexes
	if err := models.CreateIndexes(db.DB); err != nil {
		return fmt.Errorf("failed to create additional indexes: %w", err)
	}

	return nil
}

// GetStats returns database connection pool statistics
func (db *GormDB) GetStats() (map[string]interface{}, error) {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return nil, err
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections":   stats.MaxOpenConnections,
		"open_connections":       stats.OpenConnections,
		"in_use":                stats.InUse,
		"idle":                  stats.Idle,
		"wait_count":            stats.WaitCount,
		"wait_duration":         stats.WaitDuration.String(),
		"max_idle_closed":       stats.MaxIdleClosed,
		"max_idle_time_closed":  stats.MaxIdleTimeClosed,
		"max_lifetime_closed":   stats.MaxLifetimeClosed,
	}, nil
}

// HealthCheck performs a health check on the database
func (db *GormDB) HealthCheck() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// GetDialectorName returns the name of the current database dialector
func (db *GormDB) GetDialectorName() string {
	return db.DB.Dialector.Name()
}

// IsPostgreSQL checks if the current database is PostgreSQL
func (db *GormDB) IsPostgreSQL() bool {
	return db.GetDialectorName() == "postgres"
}

// IsSQLite checks if the current database is SQLite
func (db *GormDB) IsSQLite() bool {
	return db.GetDialectorName() == "sqlite"
}

// IsSQLServer checks if the current database is SQL Server
func (db *GormDB) IsSQLServer() bool {
	return db.GetDialectorName() == "sqlserver"
}