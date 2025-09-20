package database

import (
	"fmt"
	"time"

	"voicegenie/internal/config"
	"voicegenie/pkg/logger"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// DB wraps the GORM database connection
type DB struct {
	*gorm.DB
}

// New creates a new database connection
func New(cfg config.DatabaseConfig) (*DB, error) {
	var dialector gorm.Dialector

	switch cfg.Type {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
		dialector = mysql.Open(dsn)
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
			cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port, cfg.SSLMode, cfg.Timezone)
		dialector = postgres.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(cfg.Name)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	// Configure GORM logger
	var dbLogger gormLogger.Interface
	if cfg.LogLevel == "silent" {
		dbLogger = gormLogger.Default.LogMode(gormLogger.Silent)
	} else if cfg.LogLevel == "error" {
		dbLogger = gormLogger.Default.LogMode(gormLogger.Error)
	} else if cfg.LogLevel == "warn" {
		dbLogger = gormLogger.Default.LogMode(gormLogger.Warn)
	} else {
		dbLogger = gormLogger.Default.LogMode(gormLogger.Info)
	}

	// Open database connection
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: dbLogger,
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get generic database interface
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connected successfully")

	dbWrapper := &DB{DB: db}

	// Auto-migrate models
	if cfg.AutoMigrate {
		if err := dbWrapper.AutoMigrate(); err != nil {
			return nil, fmt.Errorf("failed to auto-migrate: %w", err)
		}
	}

	return dbWrapper, nil
}

// AutoMigrate runs auto-migration for all models
func (db *DB) AutoMigrate() error {
	logger.Info("Starting database migration...")

	// Import models
	var models = []interface{}{
		&User{},
		&Conversation{},
		&Message{},
		&Setting{},
		&AudioFile{},
		&APIKey{},
		&Usage{},
		&ErrorLog{},
	}

	// Migrate all models
	for _, model := range models {
		if err := db.DB.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate model %T: %w", model, err)
		}
	}

	logger.Info("Database migration completed successfully")
	return nil
}

// Health checks database health
func (db *DB) Health() bool {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return false
	}
	return sqlDB.Ping() == nil
}

// Close closes the database connection
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	logger.Info("Closing database connection")
	return sqlDB.Close()
}

// Transaction executes a function within a database transaction
func (db *DB) Transaction(fn func(*gorm.DB) error) error {
	return db.DB.Transaction(fn)
}

// GetStats returns database statistics
func (db *DB) GetStats() map[string]interface{} {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections":     stats.MaxOpenConnections,
		"open_connections":         stats.OpenConnections,
		"in_use":                   stats.InUse,
		"idle":                     stats.Idle,
		"wait_count":               stats.WaitCount,
		"wait_duration":            stats.WaitDuration,
		"max_idle_closed":          stats.MaxIdleClosed,
		"max_idle_time_closed":     stats.MaxIdleTimeClosed,
		"max_lifetime_closed":      stats.MaxLifetimeClosed,
	}
}