package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	// Application settings
	LogLevel  string `mapstructure:"log_level"`
	LogFormat string `mapstructure:"log_format"`
	DataDir   string `mapstructure:"data_dir"`
	
	// Database settings
	Database DatabaseConfig `mapstructure:"database"`
	
	// Agent settings
	Agent AgentConfig `mapstructure:"agent"`
	
	// Security settings
	Security SecurityConfig `mapstructure:"security"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type     string `mapstructure:"type"`
	Path     string `mapstructure:"path"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}

// AgentConfig holds background agent configuration
type AgentConfig struct {
	Enabled          bool   `mapstructure:"enabled"`
	ScanInterval     string `mapstructure:"scan_interval"`
	MaxConcurrency   int    `mapstructure:"max_concurrency"`
	AutoUpdateLevel  string `mapstructure:"auto_update_level"`
	NotificationMode string `mapstructure:"notification_mode"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	VerifyChecksums     bool     `mapstructure:"verify_checksums"`
	TrustedSources      []string `mapstructure:"trusted_sources"`
	BlockedPackages     []string `mapstructure:"blocked_packages"`
	RequireConfirmation bool     `mapstructure:"require_confirmation"`
	WhitelistEnabled    bool     `mapstructure:"whitelist_enabled"`
	VulnerabilityScanning bool   `mapstructure:"vulnerability_scanning"`
	MasterKey           string   `mapstructure:"master_key"`
	UpdateVulnDB        bool     `mapstructure:"update_vuln_db"`
	VulnDBUpdateInterval string  `mapstructure:"vuln_db_update_interval"`
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	// Set default configuration values
	setDefaults()
	
	// Set configuration file name and paths
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	
	// Add configuration paths
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	
	viper.AddConfigPath(".")
	viper.AddConfigPath(filepath.Join(homeDir, ".ai-dep-manager"))
	viper.AddConfigPath("/etc/ai-dep-manager")
	
	// Enable environment variable support
	viper.SetEnvPrefix("AIDM")
	viper.AutomaticEnv()
	
	// Read configuration file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is acceptable, we'll use defaults
	}
	
	// Unmarshal configuration
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	// Set data directory if not specified
	if config.DataDir == "" {
		config.DataDir = filepath.Join(homeDir, ".ai-dep-manager")
	}
	
	// Ensure data directory exists
	if err := os.MkdirAll(config.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}
	
	// Set database path if using SQLite
	if config.Database.Type == "sqlite" && config.Database.Path == "" {
		config.Database.Path = filepath.Join(config.DataDir, "ai-dep-manager.db")
	}
	
	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	viper.SetDefault("log_level", "info")
	viper.SetDefault("log_format", "text")
	
	viper.SetDefault("database.type", "sqlite")
	
	viper.SetDefault("agent.enabled", true)
	viper.SetDefault("agent.scan_interval", "1h")
	viper.SetDefault("agent.max_concurrency", 5)
	viper.SetDefault("agent.auto_update_level", "none")
	viper.SetDefault("agent.notification_mode", "console")
	
	viper.SetDefault("security.verify_checksums", true)
	viper.SetDefault("security.trusted_sources", []string{})
	viper.SetDefault("security.blocked_packages", []string{})
	viper.SetDefault("security.require_confirmation", true)
}

// Global config instance
var globalConfig *Config

// GetConfig returns the global configuration instance
func GetConfig() *Config {
	if globalConfig == nil {
		// Load default config if not initialized
		config, err := Load()
		if err != nil {
			// Return a default config if loading fails
			return &Config{
				LogLevel:  "info",
				LogFormat: "text",
				DataDir:   filepath.Join(os.Getenv("HOME"), ".ai-dep-manager"),
				Database: DatabaseConfig{
					Type: "sqlite",
					Path: filepath.Join(os.Getenv("HOME"), ".ai-dep-manager", "ai-dep-manager.db"),
				},
				Agent: AgentConfig{
					Enabled:          true,
					ScanInterval:     "1h",
					MaxConcurrency:   5,
					AutoUpdateLevel:  "none",
					NotificationMode: "console",
				},
			}
		}
		globalConfig = config
	}
	return globalConfig
}

// SetConfig sets the global configuration instance
func SetConfig(config *Config) {
	globalConfig = config
}
