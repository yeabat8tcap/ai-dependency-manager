package main

import (
	"os"

	"github.com/8tcapital/ai-dep-manager/cmd"
	"github.com/8tcapital/ai-dep-manager/internal/config"
	"github.com/8tcapital/ai-dep-manager/internal/logger"
)

func main() {
	// Initialize configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration: %v", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.Init(cfg.LogLevel, cfg.LogFormat)

	// Execute CLI commands
	if err := cmd.Execute(); err != nil {
		logger.Fatal("Command execution failed: %v", err)
		os.Exit(1)
	}
}
