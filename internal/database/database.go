package database

import (
	"fmt"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/config"
	"github.com/8tcapital/ai-dep-manager/internal/logger"
	"github.com/8tcapital/ai-dep-manager/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

var db *gorm.DB

// SetTestDB sets the database instance for testing
func SetTestDB(testDB *gorm.DB) {
	db = testDB
}

// Init initializes the database connection and runs migrations
func Init(cfg *config.Config) error {
	var err error
	
	// Configure GORM logger
	var logLevel gormLogger.LogLevel
	switch cfg.LogLevel {
	case "debug":
		logLevel = gormLogger.Info
	case "info":
		logLevel = gormLogger.Warn
	default:
		logLevel = gormLogger.Error
	}
	
	gormConfig := &gorm.Config{
		Logger: gormLogger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}
	
	// Connect to database based on type
	switch cfg.Database.Type {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(cfg.Database.Path), gormConfig)
		if err != nil {
			return fmt.Errorf("failed to connect to SQLite database: %w", err)
		}
		logger.Info("Connected to SQLite database: %s", cfg.Database.Path)
		
	default:
		return fmt.Errorf("unsupported database type: %s", cfg.Database.Type)
	}
	
	// Configure connection pool for SQLite
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	
	// SQLite specific settings
	sqlDB.SetMaxOpenConns(1) // SQLite doesn't support concurrent writes
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(time.Hour)
	
	// Run migrations
	if err := runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	
	logger.Info("Database initialized successfully")
	return nil
}

// runMigrations runs database migrations
func runMigrations() error {
	logger.Info("Running database migrations...")
	
	// Auto-migrate all models
	err := db.AutoMigrate(
		&models.Project{},
		&models.ProjectSettings{},
		&models.Dependency{},
		&models.Update{},
		&models.UpdatePolicy{},
		&models.AIPrediction{},
		&models.ScanResult{},
		&models.AuditLog{},
		&models.RollbackPlan{},
		&models.RollbackItem{},
		&models.SecurityCheck{},
		&models.SecurityRule{},
		&models.Credential{},
		&models.VulnerabilityEntry{},
	)
	
	if err != nil {
		return fmt.Errorf("auto-migration failed: %w", err)
	}
	
	logger.Info("Database migrations completed successfully")
	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return db
}

// Close closes the database connection
func Close() error {
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// Health checks database connectivity
func Health() error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	
	return nil
}
