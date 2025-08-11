package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/8tcapital/ai-dep-manager/cmd"
	"github.com/8tcapital/ai-dep-manager/internal/config"
	"github.com/8tcapital/ai-dep-manager/internal/database"
	"github.com/8tcapital/ai-dep-manager/internal/logger"
)

var (
	Version   = "production-v1.0.0"
	BuildTime = "2025-01-01T00:00:00Z"
	GitCommit = "production"
)

func main() {
	// Initialize configuration
	cfg := config.Load()
	
	// Initialize logger
	logger.Initialize(cfg.Logging)
	
	// Initialize database
	db, err := database.Initialize(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	
	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	
	// Set version info
	cmd.SetVersionInfo(Version, BuildTime, GitCommit)
	
	// Execute CLI
	ctx := context.Background()
	if err := cmd.Execute(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
