package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/ai/claude"
	"github.com/8tcapital/ai-dep-manager/internal/ai/heuristic"
	"github.com/8tcapital/ai-dep-manager/internal/ai/ollama"
	"github.com/8tcapital/ai-dep-manager/internal/ai/openai"
	"github.com/8tcapital/ai-dep-manager/internal/ai/types"
	"github.com/8tcapital/ai-dep-manager/internal/logger"
)

var (
	// Global AI manager instance
	manager *AIManager
	
	// ErrAllProvidersFailed is returned when all providers fail
	ErrAllProvidersFailed = fmt.Errorf("all AI providers failed")
)

// Initialize initializes the AI manager with configured providers
func Initialize() error {
	return InitializeWithConfig(LoadAIConfigFromEnv())
}

// InitializeWithConfig initializes the AI manager with custom configuration
func InitializeWithConfig(config *AIConfig) error {
	if err := config.ValidateConfig(); err != nil {
		return fmt.Errorf("invalid AI configuration: %w", err)
	}
	
	manager = &AIManager{
		providers: make(map[string]types.AIProvider),
		config:    config,
	}
	
	// Initialize heuristic provider (always available)
	heuristicProvider := heuristic.NewHeuristicProvider()
	manager.providers["heuristic"] = heuristicProvider
	
	// Initialize OpenAI provider if configured
	if config.OpenAI != nil && config.OpenAI.APIKey != "" {
		openaiProvider, err := openai.NewOpenAIProvider(config.OpenAI)
		if err != nil {
			logger.Warn("Failed to initialize OpenAI provider: %v", err)
		} else {
			manager.providers["openai"] = openaiProvider
			logger.Info("OpenAI provider initialized successfully")
		}
	}
	
	// Initialize Claude provider if configured
	if config.Claude != nil && config.Claude.APIKey != "" {
		claudeProvider, err := claude.NewClaudeProvider(config.Claude)
		if err != nil {
			logger.Warn("Failed to initialize Claude provider: %v", err)
		} else {
			manager.providers["claude"] = claudeProvider
			logger.Info("Claude provider initialized successfully")
		}
	}
	
	// Initialize Ollama provider if configured
	if config.Ollama != nil && config.Ollama.BaseURL != "" {
		ollamaProvider, err := ollama.NewOllamaProvider(config.Ollama)
		if err != nil {
			logger.Warn("Failed to initialize Ollama provider: %v", err)
		} else {
			manager.providers["ollama"] = ollamaProvider
			logger.Info("Ollama provider initialized successfully with model: %s", config.Ollama.Model)
		}
	}
	
	config.LogConfiguration()
	
	return nil
}

// GetProvider returns a specific AI provider by name
func GetProvider(name string) (types.AIProvider, bool) {
	if manager == nil {
		return nil, false
	}
	
	provider, exists := manager.providers[name]
	return provider, exists
}

// GetDefaultProvider returns the default AI provider
func GetDefaultProvider() (types.AIProvider, bool) {
	if manager == nil {
		return nil, false
	}
	
	provider, exists := manager.providers[manager.config.DefaultProvider]
	return provider, exists
}

// GetAvailableProviders returns a list of available provider names
func GetAvailableProviders() []string {
	if manager == nil {
		return []string{}
	}
	
	var providers []string
	for name := range manager.providers {
		providers = append(providers, name)
	}
	
	return providers
}

// executeWithFallback executes an AI operation with fallback support
func executeWithFallback[T any](ctx context.Context, operation func(types.AIProvider) (T, error)) (T, error) {
	var zero T
	
	if manager == nil {
		return zero, ErrProviderNotAvailable
	}
	
	providerOrder := manager.config.GetProviderPriority()
	var lastErr error
	
	for _, providerName := range providerOrder {
		provider, exists := manager.providers[providerName]
		if !exists {
			logger.Debug("Provider %s not available, skipping", providerName)
			continue
		}
		
		// Check if provider is available (for AI providers)
		if providerName != "heuristic" && !provider.IsAvailable(ctx) {
			logger.Debug("Provider %s not available (API check failed), skipping", providerName)
			continue
		}
		
		logger.Debug("Attempting analysis with provider: %s", providerName)
		
		// Execute with retries
		result, err := executeWithRetry(ctx, provider, operation, manager.config.MaxRetries, manager.config.RetryDelay)
		if err == nil {
			logger.Debug("Analysis successful with provider: %s", providerName)
			return result, nil
		}
		
		logger.Warn("Provider %s failed: %v", providerName, err)
		lastErr = err
		
		// If this is a context cancellation or timeout, don't try other providers
		if ctx.Err() != nil {
			return zero, ctx.Err()
		}
	}
	
	if lastErr != nil {
		return zero, fmt.Errorf("%w: %v", ErrAllProvidersFailed, lastErr)
	}
	
	return zero, ErrProviderNotAvailable
}

// executeWithRetry executes an operation with retry logic
func executeWithRetry[T any](ctx context.Context, provider types.AIProvider, operation func(types.AIProvider) (T, error), maxRetries int, retryDelay time.Duration) (T, error) {
	var zero T
	var lastErr error
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			logger.Debug("Retrying operation with %s (attempt %d/%d)", provider.GetName(), attempt+1, maxRetries+1)
			
			// Wait before retry
			select {
			case <-ctx.Done():
				return zero, ctx.Err()
			case <-time.After(retryDelay):
			}
		}
		
		result, err := operation(provider)
		if err == nil {
			return result, nil
		}
		
		lastErr = err
		
		// Don't retry on context cancellation
		if ctx.Err() != nil {
			return zero, ctx.Err()
		}
		
		logger.Debug("Attempt %d failed with %s: %v", attempt+1, provider.GetName(), err)
	}
	
	return zero, lastErr
}

// AnalyzeChangelog delegates changelog analysis with fallback support
func AnalyzeChangelog(ctx context.Context, request *types.ChangelogAnalysisRequest) (*types.ChangelogAnalysisResponse, error) {
	return executeWithFallback(ctx, func(provider types.AIProvider) (*types.ChangelogAnalysisResponse, error) {
		return provider.AnalyzeChangelog(ctx, request)
	})
}

// AnalyzeVersionDiff delegates version diff analysis with fallback support
func AnalyzeVersionDiff(ctx context.Context, request *types.VersionDiffAnalysisRequest) (*types.VersionDiffAnalysisResponse, error) {
	return executeWithFallback(ctx, func(provider types.AIProvider) (*types.VersionDiffAnalysisResponse, error) {
		return provider.AnalyzeVersionDiff(ctx, request)
	})
}

// PredictCompatibility delegates compatibility prediction with fallback support
func PredictCompatibility(ctx context.Context, request *types.CompatibilityPredictionRequest) (*types.CompatibilityPredictionResponse, error) {
	return executeWithFallback(ctx, func(provider types.AIProvider) (*types.CompatibilityPredictionResponse, error) {
		return provider.PredictCompatibility(ctx, request)
	})
}

// ClassifyUpdate delegates update classification with fallback support
func ClassifyUpdate(ctx context.Context, request *types.UpdateClassificationRequest) (*types.UpdateClassificationResponse, error) {
	return executeWithFallback(ctx, func(provider types.AIProvider) (*types.UpdateClassificationResponse, error) {
		return provider.ClassifyUpdate(ctx, request)
	})
}
