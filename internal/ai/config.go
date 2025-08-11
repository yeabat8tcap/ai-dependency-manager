package ai

import (
	"fmt"
	"os"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/ai/claude"
	"github.com/8tcapital/ai-dep-manager/internal/ai/ollama"
	"github.com/8tcapital/ai-dep-manager/internal/ai/openai"
	"github.com/8tcapital/ai-dep-manager/internal/logger"
)

// AIConfig holds configuration for AI providers
type AIConfig struct {
	// Default provider to use (heuristic, openai, claude)
	DefaultProvider string `json:"default_provider" yaml:"default_provider"`
	
	// Fallback providers in order of preference
	FallbackProviders []string `json:"fallback_providers" yaml:"fallback_providers"`
	
	// Enable fallback to heuristic if all AI providers fail
	EnableHeuristicFallback bool `json:"enable_heuristic_fallback" yaml:"enable_heuristic_fallback"`
	
	// Provider-specific configurations
	OpenAI *openai.OpenAIConfig `json:"openai" yaml:"openai"`
	Claude *claude.ClaudeConfig `json:"claude" yaml:"claude"`
	Ollama *ollama.OllamaConfig `json:"ollama" yaml:"ollama"`
	
	// Global AI settings
	MaxRetries    int           `json:"max_retries" yaml:"max_retries"`
	RetryDelay    time.Duration `json:"retry_delay" yaml:"retry_delay"`
	RequestTimeout time.Duration `json:"request_timeout" yaml:"request_timeout"`
}

// DefaultAIConfig returns default AI configuration
func DefaultAIConfig() *AIConfig {
	return &AIConfig{
		DefaultProvider:         "heuristic",
		FallbackProviders:       []string{"heuristic"},
		EnableHeuristicFallback: true,
		MaxRetries:              3,
		RetryDelay:              2 * time.Second,
		RequestTimeout:          30 * time.Second,
		OpenAI: &openai.OpenAIConfig{
			Model:       "gpt-4",
			MaxTokens:   4096,
			Temperature: 0.1,
			Timeout:     30 * time.Second,
			BaseURL:     "https://api.openai.com/v1",
		},
		Claude: &claude.ClaudeConfig{
			Model:       "claude-3-5-sonnet-20241022",
			MaxTokens:   4096,
			Temperature: 0.1,
			Timeout:     30 * time.Second,
			BaseURL:     "https://api.anthropic.com",
		},
		Ollama: &ollama.OllamaConfig{
			BaseURL:     "http://localhost:11434",
			Model:       "llama2",
			Temperature: 0.7,
			TopP:        0.9,
			TopK:        40,
			NumPredict:  2048,
		},
	}
}

// LoadAIConfigFromEnv loads AI configuration from environment variables
func LoadAIConfigFromEnv() *AIConfig {
	config := DefaultAIConfig()
	
	// Load default provider
	if provider := os.Getenv("AI_DEFAULT_PROVIDER"); provider != "" {
		config.DefaultProvider = provider
	}
	
	// Load OpenAI configuration
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		config.OpenAI.APIKey = apiKey
		logger.Debug("OpenAI API key loaded from environment")
	}
	
	if model := os.Getenv("OPENAI_MODEL"); model != "" {
		config.OpenAI.Model = model
	}
	
	if baseURL := os.Getenv("OPENAI_BASE_URL"); baseURL != "" {
		config.OpenAI.BaseURL = baseURL
	}
	
	// Load Claude configuration
	if apiKey := os.Getenv("CLAUDE_API_KEY"); apiKey != "" {
		config.Claude.APIKey = apiKey
		logger.Debug("Claude API key loaded from environment")
	}
	
	if model := os.Getenv("CLAUDE_MODEL"); model != "" {
		config.Claude.Model = model
	}
	
	if baseURL := os.Getenv("CLAUDE_BASE_URL"); baseURL != "" {
		config.Claude.BaseURL = baseURL
	}
	
	// Load Ollama configuration
	if baseURL := os.Getenv("OLLAMA_BASE_URL"); baseURL != "" {
		config.Ollama.BaseURL = baseURL
	}
	
	if model := os.Getenv("OLLAMA_MODEL"); model != "" {
		config.Ollama.Model = model
		logger.Debug("Ollama model loaded from environment: %s", model)
	}
	
	// Set fallback providers based on available API keys
	config.updateFallbackProviders()
	
	return config
}

// updateFallbackProviders updates fallback providers based on available API keys
func (c *AIConfig) updateFallbackProviders() {
	var providers []string
	
	// Add providers based on available API keys and availability
	if c.OpenAI != nil && c.OpenAI.APIKey != "" {
		providers = append(providers, "openai")
	}
	
	if c.Claude != nil && c.Claude.APIKey != "" {
		providers = append(providers, "claude")
	}
	
	// Ollama doesn't require API key, check if configured
	if c.Ollama != nil && c.Ollama.BaseURL != "" {
		providers = append(providers, "ollama")
	}
	
	// Always include heuristic as final fallback
	if c.EnableHeuristicFallback {
		providers = append(providers, "heuristic")
	}
	
	c.FallbackProviders = providers
	
	// Set default provider to first available AI provider if not set to AI provider
	if c.DefaultProvider == "heuristic" && len(providers) > 1 {
		c.DefaultProvider = providers[0]
		logger.Info("Default AI provider set to %s (API key available)", c.DefaultProvider)
	}
}

// ValidateConfig validates the AI configuration
func (c *AIConfig) ValidateConfig() error {
	// Validate default provider
	validProviders := map[string]bool{
		"heuristic": true,
		"openai":    true,
		"claude":    true,
	}
	
	if !validProviders[c.DefaultProvider] {
		return fmt.Errorf("invalid default provider: %s", c.DefaultProvider)
	}
	
	// Validate fallback providers
	for _, provider := range c.FallbackProviders {
		if !validProviders[provider] {
			return fmt.Errorf("invalid fallback provider: %s", provider)
		}
	}
	
	// Validate OpenAI config if provider is enabled
	if c.hasProvider("openai") {
		if c.OpenAI == nil {
			return fmt.Errorf("OpenAI provider enabled but configuration is missing")
		}
		if c.OpenAI.APIKey == "" {
			return fmt.Errorf("OpenAI API key is required when OpenAI provider is enabled")
		}
	}
	
	// Validate Claude config if provider is enabled
	if c.hasProvider("claude") {
		if c.Claude == nil {
			return fmt.Errorf("Claude provider enabled but configuration is missing")
		}
		if c.Claude.APIKey == "" {
			return fmt.Errorf("Claude API key is required when Claude provider is enabled")
		}
	}
	
	return nil
}

// hasProvider checks if a provider is configured
func (c *AIConfig) hasProvider(provider string) bool {
	if c.DefaultProvider == provider {
		return true
	}
	
	for _, p := range c.FallbackProviders {
		if p == provider {
			return true
		}
	}
	
	return false
}

// GetProviderPriority returns the priority order of providers
func (c *AIConfig) GetProviderPriority() []string {
	providers := []string{c.DefaultProvider}
	
	for _, provider := range c.FallbackProviders {
		if provider != c.DefaultProvider {
			providers = append(providers, provider)
		}
	}
	
	return providers
}

// IsAIProviderAvailable checks if any AI provider (non-heuristic) is available
func (c *AIConfig) IsAIProviderAvailable() bool {
	for _, provider := range c.GetProviderPriority() {
		if provider != "heuristic" {
			return true
		}
	}
	return false
}

// GetProviderConfig returns configuration for a specific provider
func (c *AIConfig) GetProviderConfig(provider string) interface{} {
	switch provider {
	case "openai":
		return c.OpenAI
	case "claude":
		return c.Claude
	default:
		return nil
	}
}

// LogConfiguration logs the current AI configuration (without sensitive data)
func (c *AIConfig) LogConfiguration() {
	logger.Info("AI Configuration:")
	logger.Info("  Default Provider: %s", c.DefaultProvider)
	logger.Info("  Fallback Providers: %v", c.FallbackProviders)
	logger.Info("  Heuristic Fallback: %t", c.EnableHeuristicFallback)
	logger.Info("  Max Retries: %d", c.MaxRetries)
	logger.Info("  Request Timeout: %v", c.RequestTimeout)
	
	if c.OpenAI != nil {
		hasKey := c.OpenAI.APIKey != ""
		logger.Info("  OpenAI: model=%s, api_key=%t", c.OpenAI.Model, hasKey)
	}
	
	if c.Claude != nil {
		hasKey := c.Claude.APIKey != ""
		logger.Info("  Claude: model=%s, api_key=%t", c.Claude.Model, hasKey)
	}
}
