package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/8tcapital/ai-dep-manager/internal/config"
	"github.com/8tcapital/ai-dep-manager/internal/database"
	"github.com/8tcapital/ai-dep-manager/internal/logger"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	cfg     *config.Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ai-dep-manager",
	Short: "AI-powered dependency management tool",
	Long: `AI Dependency Manager (AutoUpdateAgent) is an intelligent CLI tool that helps you
manage software dependencies across multiple package managers using AI-powered analysis.

It can:
- Scan your projects for outdated dependencies
- Analyze changelogs to detect breaking changes
- Suggest safe update strategies
- Run as a background agent for continuous monitoring
- Integrate with npm, pip, Maven, Gradle and more`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		var err error
		cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Initialize logger
		logger.Init(cfg.LogLevel, cfg.LogFormat)

		// Initialize database
		if err := database.Init(cfg); err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		// Clean up database connection
		if err := database.Close(); err != nil {
			logger.Error("Failed to close database connection: %v", err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ai-dep-manager/config.yaml)")
	rootCmd.PersistentFlags().String("log-level", "info", "log level (debug, info, warn, error, fatal)")
	rootCmd.PersistentFlags().String("log-format", "text", "log format (text, json)")
	rootCmd.PersistentFlags().Bool("verbose", false, "enable verbose output")

	// Bind flags to viper
	viper.BindPFlag("log_level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("log_format", rootCmd.PersistentFlags().Lookup("log-format"))
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	}
}
