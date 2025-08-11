package ai

import (
	"context"
	"fmt"
	"github.com/8tcapital/ai-dep-manager/internal/ai/types"
)

// AIProvider is an alias to the types.AIProvider interface
type AIProvider = types.AIProvider

// Request/Response type aliases for convenience
type ChangelogAnalysisRequest = types.ChangelogAnalysisRequest
type ChangelogAnalysisResponse = types.ChangelogAnalysisResponse
type VersionDiffAnalysisRequest = types.VersionDiffAnalysisRequest
type VersionDiffAnalysisResponse = types.VersionDiffAnalysisResponse
type CompatibilityPredictionRequest = types.CompatibilityPredictionRequest
type CompatibilityPredictionResponse = types.CompatibilityPredictionResponse
type UpdateClassificationRequest = types.UpdateClassificationRequest
type UpdateClassificationResponse = types.UpdateClassificationResponse

// Supporting type aliases
type BreakingChange = types.BreakingChange
type Feature = types.Feature
type BugFix = types.BugFix
type SecurityFix = types.SecurityFix
type DeprecatedAPI = types.DeprecatedAPI
type UpdateType = types.UpdateType
type Priority = types.Priority
type RiskLevel = types.RiskLevel

// AIManager manages multiple AI providers
type AIManager struct {
	providers       map[string]AIProvider
	defaultProvider string
	config          *AIConfig
}

// NewAIManager creates a new AI manager
func NewAIManager() *AIManager {
	return &AIManager{
		providers: make(map[string]AIProvider),
	}
}

// RegisterProvider registers an AI provider
func (am *AIManager) RegisterProvider(provider AIProvider) {
	am.providers[provider.GetName()] = provider
	
	// Set as default if it's the first provider
	if am.defaultProvider == "" {
		am.defaultProvider = provider.GetName()
	}
}

// GetProvider returns a specific AI provider
func (am *AIManager) GetProvider(name string) (AIProvider, bool) {
	provider, exists := am.providers[name]
	return provider, exists
}

// GetDefaultProvider returns the default AI provider
func (am *AIManager) GetDefaultProvider() (AIProvider, bool) {
	if am.defaultProvider == "" {
		return nil, false
	}
	return am.GetProvider(am.defaultProvider)
}

// SetDefaultProvider sets the default AI provider
func (am *AIManager) SetDefaultProvider(name string) error {
	if _, exists := am.providers[name]; !exists {
		return ErrProviderNotFound
	}
	am.defaultProvider = name
	return nil
}

// GetAvailableProviders returns all available AI providers
func (am *AIManager) GetAvailableProviders(ctx context.Context) map[string]AIProvider {
	available := make(map[string]AIProvider)
	for name, provider := range am.providers {
		if provider.IsAvailable(ctx) {
			available[name] = provider
		}
	}
	return available
}

// ListProviders returns all registered providers
func (am *AIManager) ListProviders() []string {
	var names []string
	for name := range am.providers {
		names = append(names, name)
	}
	return names
}

// Common errors
var (
	ErrProviderNotFound     = fmt.Errorf("AI provider not found")
	ErrProviderNotAvailable = fmt.Errorf("AI provider not available")
	ErrInvalidRequest       = fmt.Errorf("invalid AI request")
	ErrAnalysisFailed       = fmt.Errorf("AI analysis failed")
)
