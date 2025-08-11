package github

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/logger"
)

// Config represents GitHub integration configuration
type Config struct {
	// Authentication
	AuthType           string `json:"auth_type" yaml:"auth_type"`                       // "pat" or "app"
	PersonalAccessToken string `json:"personal_access_token,omitempty" yaml:"personal_access_token,omitempty"`
	
	// GitHub App configuration
	AppID              int64  `json:"app_id,omitempty" yaml:"app_id,omitempty"`
	InstallationID     int64  `json:"installation_id,omitempty" yaml:"installation_id,omitempty"`
	PrivateKeyPath     string `json:"private_key_path,omitempty" yaml:"private_key_path,omitempty"`
	PrivateKeyPEM      string `json:"private_key_pem,omitempty" yaml:"private_key_pem,omitempty"`
	
	// API Configuration
	BaseURL            string        `json:"base_url" yaml:"base_url"`
	Timeout            time.Duration `json:"timeout" yaml:"timeout"`
	RateLimitRetries   int           `json:"rate_limit_retries" yaml:"rate_limit_retries"`
	
	// Webhook Configuration
	WebhookURL         string `json:"webhook_url" yaml:"webhook_url"`
	WebhookSecret      string `json:"webhook_secret" yaml:"webhook_secret"`
	WebhookPort        int    `json:"webhook_port" yaml:"webhook_port"`
	
	// Patch Configuration
	PatchBranchPrefix  string   `json:"patch_branch_prefix" yaml:"patch_branch_prefix"`
	DefaultReviewers   []string `json:"default_reviewers" yaml:"default_reviewers"`
	DefaultLabels      []string `json:"default_labels" yaml:"default_labels"`
	AutoMergePatch     bool     `json:"auto_merge_patch" yaml:"auto_merge_patch"`
	AutoMergeSecurity  bool     `json:"auto_merge_security" yaml:"auto_merge_security"`
	
	// Repository Configuration
	Repositories       []RepositoryConfig `json:"repositories" yaml:"repositories"`
	
	// Cleanup Configuration
	CleanupOldBranches bool          `json:"cleanup_old_branches" yaml:"cleanup_old_branches"`
	BranchMaxAge       time.Duration `json:"branch_max_age" yaml:"branch_max_age"`
}

// RepositoryConfig represents configuration for a specific repository
type RepositoryConfig struct {
	Owner             string   `json:"owner" yaml:"owner"`
	Name              string   `json:"name" yaml:"name"`
	Enabled           bool     `json:"enabled" yaml:"enabled"`
	BaseBranch        string   `json:"base_branch,omitempty" yaml:"base_branch,omitempty"`
	Reviewers         []string `json:"reviewers,omitempty" yaml:"reviewers,omitempty"`
	Labels            []string `json:"labels,omitempty" yaml:"labels,omitempty"`
	AutoMergePatch    *bool    `json:"auto_merge_patch,omitempty" yaml:"auto_merge_patch,omitempty"`
	AutoMergeSecurity *bool    `json:"auto_merge_security,omitempty" yaml:"auto_merge_security,omitempty"`
	WebhookEnabled    bool     `json:"webhook_enabled" yaml:"webhook_enabled"`
}

// DefaultConfig returns a default GitHub configuration
func DefaultConfig() *Config {
	return &Config{
		AuthType:           "pat",
		BaseURL:            DefaultBaseURL,
		Timeout:            30 * time.Second,
		RateLimitRetries:   3,
		WebhookPort:        8080,
		PatchBranchPrefix:  "ai-dep-manager/patch",
		DefaultLabels:      []string{"dependencies", "automated"},
		AutoMergePatch:     false,
		AutoMergeSecurity:  false,
		CleanupOldBranches: true,
		BranchMaxAge:       7 * 24 * time.Hour, // 7 days
		Repositories:       []RepositoryConfig{},
	}
}

// LoadConfigFromEnv loads GitHub configuration from environment variables
func LoadConfigFromEnv() *Config {
	config := DefaultConfig()
	
	// Authentication configuration
	if authType := os.Getenv("GITHUB_AUTH_TYPE"); authType != "" {
		config.AuthType = authType
	}
	
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		config.PersonalAccessToken = token
		config.AuthType = "pat"
	}
	
	if appIDStr := os.Getenv("GITHUB_APP_ID"); appIDStr != "" {
		if appID, err := strconv.ParseInt(appIDStr, 10, 64); err == nil {
			config.AppID = appID
			config.AuthType = "app"
		}
	}
	
	if installationIDStr := os.Getenv("GITHUB_INSTALLATION_ID"); installationIDStr != "" {
		if installationID, err := strconv.ParseInt(installationIDStr, 10, 64); err == nil {
			config.InstallationID = installationID
		}
	}
	
	if privateKeyPath := os.Getenv("GITHUB_PRIVATE_KEY_PATH"); privateKeyPath != "" {
		config.PrivateKeyPath = privateKeyPath
	}
	
	if privateKeyPEM := os.Getenv("GITHUB_PRIVATE_KEY_PEM"); privateKeyPEM != "" {
		config.PrivateKeyPEM = privateKeyPEM
	}
	
	// API configuration
	if baseURL := os.Getenv("GITHUB_BASE_URL"); baseURL != "" {
		config.BaseURL = baseURL
	}
	
	if timeoutStr := os.Getenv("GITHUB_TIMEOUT"); timeoutStr != "" {
		if timeout, err := time.ParseDuration(timeoutStr); err == nil {
			config.Timeout = timeout
		}
	}
	
	if retriesStr := os.Getenv("GITHUB_RATE_LIMIT_RETRIES"); retriesStr != "" {
		if retries, err := strconv.Atoi(retriesStr); err == nil {
			config.RateLimitRetries = retries
		}
	}
	
	// Webhook configuration
	if webhookURL := os.Getenv("GITHUB_WEBHOOK_URL"); webhookURL != "" {
		config.WebhookURL = webhookURL
	}
	
	if webhookSecret := os.Getenv("GITHUB_WEBHOOK_SECRET"); webhookSecret != "" {
		config.WebhookSecret = webhookSecret
	}
	
	if webhookPortStr := os.Getenv("GITHUB_WEBHOOK_PORT"); webhookPortStr != "" {
		if webhookPort, err := strconv.Atoi(webhookPortStr); err == nil {
			config.WebhookPort = webhookPort
		}
	}
	
	// Patch configuration
	if branchPrefix := os.Getenv("GITHUB_PATCH_BRANCH_PREFIX"); branchPrefix != "" {
		config.PatchBranchPrefix = branchPrefix
	}
	
	if reviewersStr := os.Getenv("GITHUB_DEFAULT_REVIEWERS"); reviewersStr != "" {
		config.DefaultReviewers = strings.Split(reviewersStr, ",")
		// Trim whitespace
		for i, reviewer := range config.DefaultReviewers {
			config.DefaultReviewers[i] = strings.TrimSpace(reviewer)
		}
	}
	
	if labelsStr := os.Getenv("GITHUB_DEFAULT_LABELS"); labelsStr != "" {
		config.DefaultLabels = strings.Split(labelsStr, ",")
		// Trim whitespace
		for i, label := range config.DefaultLabels {
			config.DefaultLabels[i] = strings.TrimSpace(label)
		}
	}
	
	if autoMergePatchStr := os.Getenv("GITHUB_AUTO_MERGE_PATCH"); autoMergePatchStr != "" {
		config.AutoMergePatch = strings.ToLower(autoMergePatchStr) == "true"
	}
	
	if autoMergeSecurityStr := os.Getenv("GITHUB_AUTO_MERGE_SECURITY"); autoMergeSecurityStr != "" {
		config.AutoMergeSecurity = strings.ToLower(autoMergeSecurityStr) == "true"
	}
	
	// Cleanup configuration
	if cleanupStr := os.Getenv("GITHUB_CLEANUP_OLD_BRANCHES"); cleanupStr != "" {
		config.CleanupOldBranches = strings.ToLower(cleanupStr) == "true"
	}
	
	if maxAgeStr := os.Getenv("GITHUB_BRANCH_MAX_AGE"); maxAgeStr != "" {
		if maxAge, err := time.ParseDuration(maxAgeStr); err == nil {
			config.BranchMaxAge = maxAge
		}
	}
	
	// Repository configuration from environment
	if reposStr := os.Getenv("GITHUB_REPOSITORIES"); reposStr != "" {
		repos := strings.Split(reposStr, ",")
		for _, repo := range repos {
			repo = strings.TrimSpace(repo)
			if parts := strings.Split(repo, "/"); len(parts) == 2 {
				config.Repositories = append(config.Repositories, RepositoryConfig{
					Owner:          parts[0],
					Name:           parts[1],
					Enabled:        true,
					WebhookEnabled: true,
				})
			}
		}
	}
	
	return config
}

// Validate validates the GitHub configuration
func (c *Config) Validate() error {
	// Validate authentication
	switch strings.ToLower(c.AuthType) {
	case "pat", "token", "personal_access_token":
		if c.PersonalAccessToken == "" {
			return fmt.Errorf("personal access token is required for PAT authentication")
		}
	case "app", "github_app":
		if c.AppID == 0 {
			return fmt.Errorf("app ID is required for GitHub App authentication")
		}
		if c.InstallationID == 0 {
			return fmt.Errorf("installation ID is required for GitHub App authentication")
		}
		if c.PrivateKeyPath == "" && c.PrivateKeyPEM == "" {
			return fmt.Errorf("private key (path or PEM) is required for GitHub App authentication")
		}
	default:
		return fmt.Errorf("unsupported authentication type: %s", c.AuthType)
	}
	
	// Validate API configuration
	if c.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}
	
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	
	if c.RateLimitRetries < 0 {
		return fmt.Errorf("rate limit retries cannot be negative")
	}
	
	// Validate webhook configuration
	if c.WebhookPort <= 0 || c.WebhookPort > 65535 {
		return fmt.Errorf("webhook port must be between 1 and 65535")
	}
	
	// Validate patch configuration
	if c.PatchBranchPrefix == "" {
		return fmt.Errorf("patch branch prefix is required")
	}
	
	if c.BranchMaxAge <= 0 {
		return fmt.Errorf("branch max age must be positive")
	}
	
	// Validate repositories
	for i, repo := range c.Repositories {
		if repo.Owner == "" {
			return fmt.Errorf("repository[%d]: owner is required", i)
		}
		if repo.Name == "" {
			return fmt.Errorf("repository[%d]: name is required", i)
		}
	}
	
	return nil
}

// GetRepositoryConfig returns configuration for a specific repository
func (c *Config) GetRepositoryConfig(owner, name string) *RepositoryConfig {
	for _, repo := range c.Repositories {
		if repo.Owner == owner && repo.Name == name {
			return &repo
		}
	}
	return nil
}

// IsRepositoryEnabled checks if a repository is enabled for GitHub integration
func (c *Config) IsRepositoryEnabled(owner, name string) bool {
	repo := c.GetRepositoryConfig(owner, name)
	return repo != nil && repo.Enabled
}

// GetReviewers returns reviewers for a repository (repo-specific or default)
func (c *Config) GetReviewers(owner, name string) []string {
	if repo := c.GetRepositoryConfig(owner, name); repo != nil && len(repo.Reviewers) > 0 {
		return repo.Reviewers
	}
	return c.DefaultReviewers
}

// GetLabels returns labels for a repository (repo-specific or default)
func (c *Config) GetLabels(owner, name string) []string {
	if repo := c.GetRepositoryConfig(owner, name); repo != nil && len(repo.Labels) > 0 {
		return repo.Labels
	}
	return c.DefaultLabels
}

// ShouldAutoMergePatch checks if patch PRs should be auto-merged for a repository
func (c *Config) ShouldAutoMergePatch(owner, name string) bool {
	if repo := c.GetRepositoryConfig(owner, name); repo != nil && repo.AutoMergePatch != nil {
		return *repo.AutoMergePatch
	}
	return c.AutoMergePatch
}

// ShouldAutoMergeSecurity checks if security PRs should be auto-merged for a repository
func (c *Config) ShouldAutoMergeSecurity(owner, name string) bool {
	if repo := c.GetRepositoryConfig(owner, name); repo != nil && repo.AutoMergeSecurity != nil {
		return *repo.AutoMergeSecurity
	}
	return c.AutoMergeSecurity
}

// GetBaseBranch returns the base branch for a repository
func (c *Config) GetBaseBranch(owner, name string) string {
	if repo := c.GetRepositoryConfig(owner, name); repo != nil && repo.BaseBranch != "" {
		return repo.BaseBranch
	}
	return "main" // Default base branch
}

// IsWebhookEnabled checks if webhooks are enabled for a repository
func (c *Config) IsWebhookEnabled(owner, name string) bool {
	if repo := c.GetRepositoryConfig(owner, name); repo != nil {
		return repo.WebhookEnabled
	}
	return false
}

// AddRepository adds a repository to the configuration
func (c *Config) AddRepository(owner, name string, enabled bool) {
	// Check if repository already exists
	for i, repo := range c.Repositories {
		if repo.Owner == owner && repo.Name == name {
			c.Repositories[i].Enabled = enabled
			return
		}
	}
	
	// Add new repository
	c.Repositories = append(c.Repositories, RepositoryConfig{
		Owner:          owner,
		Name:           name,
		Enabled:        enabled,
		WebhookEnabled: enabled,
	})
}

// RemoveRepository removes a repository from the configuration
func (c *Config) RemoveRepository(owner, name string) {
	for i, repo := range c.Repositories {
		if repo.Owner == owner && repo.Name == name {
			c.Repositories = append(c.Repositories[:i], c.Repositories[i+1:]...)
			return
		}
	}
}

// GetEnabledRepositories returns all enabled repositories
func (c *Config) GetEnabledRepositories() []RepositoryConfig {
	var enabled []RepositoryConfig
	for _, repo := range c.Repositories {
		if repo.Enabled {
			enabled = append(enabled, repo)
		}
	}
	return enabled
}

// GetWebhookEnabledRepositories returns all repositories with webhooks enabled
func (c *Config) GetWebhookEnabledRepositories() []RepositoryConfig {
	var enabled []RepositoryConfig
	for _, repo := range c.Repositories {
		if repo.Enabled && repo.WebhookEnabled {
			enabled = append(enabled, repo)
		}
	}
	return enabled
}

// CreateAuthConfig creates an AuthConfig from the GitHub configuration
func (c *Config) CreateAuthConfig() *AuthConfig {
	authConfig := &AuthConfig{
		Type: c.AuthType,
	}
	
	switch strings.ToLower(c.AuthType) {
	case "pat", "token", "personal_access_token":
		authConfig.Token = c.PersonalAccessToken
	case "app", "github_app":
		authConfig.AppID = c.AppID
		authConfig.InstallationID = c.InstallationID
		authConfig.PrivateKeyPath = c.PrivateKeyPath
		authConfig.PrivateKeyPEM = c.PrivateKeyPEM
	}
	
	return authConfig
}

// LogConfiguration logs the current configuration (without sensitive data)
func (c *Config) LogConfiguration() {
	logger.Info("GitHub Integration Configuration:")
	logger.Info("  Auth Type: %s", c.AuthType)
	logger.Info("  Base URL: %s", c.BaseURL)
	logger.Info("  Timeout: %v", c.Timeout)
	logger.Info("  Rate Limit Retries: %d", c.RateLimitRetries)
	logger.Info("  Webhook Port: %d", c.WebhookPort)
	logger.Info("  Patch Branch Prefix: %s", c.PatchBranchPrefix)
	logger.Info("  Default Labels: %v", c.DefaultLabels)
	logger.Info("  Auto Merge Patch: %t", c.AutoMergePatch)
	logger.Info("  Auto Merge Security: %t", c.AutoMergeSecurity)
	logger.Info("  Cleanup Old Branches: %t", c.CleanupOldBranches)
	logger.Info("  Branch Max Age: %v", c.BranchMaxAge)
	logger.Info("  Repositories: %d configured", len(c.Repositories))
	
	for i, repo := range c.Repositories {
		logger.Info("    [%d] %s/%s (enabled: %t, webhook: %t)", 
			i+1, repo.Owner, repo.Name, repo.Enabled, repo.WebhookEnabled)
	}
}
