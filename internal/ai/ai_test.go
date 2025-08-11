package ai

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/ai/types"
)

func TestAIManagerInitialization(t *testing.T) {
	tests := []struct {
		name           string
		config         *AIConfig
		expectError    bool
		expectedProviders []string
	}{
		{
			name:   "Default configuration",
			config: DefaultAIConfig(),
			expectError: false,
			expectedProviders: []string{"heuristic"},
		},
		{
			name: "OpenAI configuration",
			config: &AIConfig{
				DefaultProvider:         "openai",
				FallbackProviders:       []string{"openai", "heuristic"},
				EnableHeuristicFallback: true,
				OpenAI: &openai.OpenAIConfig{
					APIKey:      "test-key",
					Model:       "gpt-4",
					MaxTokens:   4096,
					Temperature: 0.1,
					Timeout:     30 * time.Second,
					BaseURL:     "https://api.openai.com/v1",
				},
				MaxRetries:     3,
				RetryDelay:     2 * time.Second,
				RequestTimeout: 30 * time.Second,
			},
			expectError: false,
			expectedProviders: []string{"heuristic", "openai"},
		},
		{
			name: "Claude configuration",
			config: &AIConfig{
				DefaultProvider:         "claude",
				FallbackProviders:       []string{"claude", "heuristic"},
				EnableHeuristicFallback: true,
				Claude: &claude.ClaudeConfig{
					APIKey:      "test-key",
					Model:       "claude-3-5-sonnet-20241022",
					MaxTokens:   4096,
					Temperature: 0.1,
					Timeout:     30 * time.Second,
					BaseURL:     "https://api.anthropic.com",
				},
				MaxRetries:     3,
				RetryDelay:     2 * time.Second,
				RequestTimeout: 30 * time.Second,
			},
			expectError: false,
			expectedProviders: []string{"heuristic", "claude"},
		},
		{
			name: "Invalid default provider",
			config: &AIConfig{
				DefaultProvider:         "invalid",
				FallbackProviders:       []string{"heuristic"},
				EnableHeuristicFallback: true,
				MaxRetries:              3,
				RetryDelay:              2 * time.Second,
				RequestTimeout:          30 * time.Second,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset manager
			manager = nil
			
			err := InitializeWithConfig(tt.config)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			// Check available providers
			availableProviders := GetAvailableProviders()
			if len(availableProviders) != len(tt.expectedProviders) {
				t.Errorf("Expected %d providers, got %d", len(tt.expectedProviders), len(availableProviders))
			}
			
			for _, expectedProvider := range tt.expectedProviders {
				found := false
				for _, provider := range availableProviders {
					if provider == expectedProvider {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected provider %s not found", expectedProvider)
				}
			}
		})
	}
}

func TestAIConfigFromEnvironment(t *testing.T) {
	// Save original environment
	originalOpenAI := os.Getenv("OPENAI_API_KEY")
	originalClaude := os.Getenv("CLAUDE_API_KEY")
	originalProvider := os.Getenv("AI_DEFAULT_PROVIDER")
	
	// Restore environment after test
	defer func() {
		os.Setenv("OPENAI_API_KEY", originalOpenAI)
		os.Setenv("CLAUDE_API_KEY", originalClaude)
		os.Setenv("AI_DEFAULT_PROVIDER", originalProvider)
	}()
	
	tests := []struct {
		name                string
		envVars             map[string]string
		expectedDefault     string
		expectedFallbacks   int
	}{
		{
			name: "No API keys",
			envVars: map[string]string{
				"OPENAI_API_KEY": "",
				"CLAUDE_API_KEY": "",
			},
			expectedDefault:   "heuristic",
			expectedFallbacks: 1,
		},
		{
			name: "OpenAI API key only",
			envVars: map[string]string{
				"OPENAI_API_KEY": "test-openai-key",
				"CLAUDE_API_KEY": "",
			},
			expectedDefault:   "openai",
			expectedFallbacks: 2,
		},
		{
			name: "Claude API key only",
			envVars: map[string]string{
				"OPENAI_API_KEY": "",
				"CLAUDE_API_KEY": "test-claude-key",
			},
			expectedDefault:   "claude",
			expectedFallbacks: 2,
		},
		{
			name: "Both API keys",
			envVars: map[string]string{
				"OPENAI_API_KEY": "test-openai-key",
				"CLAUDE_API_KEY": "test-claude-key",
			},
			expectedDefault:   "openai",
			expectedFallbacks: 3,
		},
		{
			name: "Explicit default provider",
			envVars: map[string]string{
				"OPENAI_API_KEY":     "test-openai-key",
				"CLAUDE_API_KEY":     "test-claude-key",
				"AI_DEFAULT_PROVIDER": "claude",
			},
			expectedDefault:   "claude",
			expectedFallbacks: 3,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			
			config := LoadAIConfigFromEnv()
			
			if config.DefaultProvider != tt.expectedDefault {
				t.Errorf("Expected default provider %s, got %s", tt.expectedDefault, config.DefaultProvider)
			}
			
			if len(config.FallbackProviders) != tt.expectedFallbacks {
				t.Errorf("Expected %d fallback providers, got %d", tt.expectedFallbacks, len(config.FallbackProviders))
			}
		})
	}
}

func TestChangelogAnalysisWithFallback(t *testing.T) {
	// Initialize with heuristic provider only
	config := DefaultAIConfig()
	err := InitializeWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to initialize AI manager: %v", err)
	}
	
	ctx := context.Background()
	request := &types.ChangelogAnalysisRequest{
		PackageName:    "test-package",
		FromVersion:    "1.0.0",
		ToVersion:      "2.0.0",
		PackageManager: "npm",
		Language:       "javascript",
		ChangelogText:  "## Breaking Changes\n- Removed deprecated API\n- Changed function signatures",
		ReleaseNotes:   "Major version with breaking changes",
	}
	
	response, err := AnalyzeChangelog(ctx, request)
	if err != nil {
		t.Fatalf("Changelog analysis failed: %v", err)
	}
	
	if response == nil {
		t.Fatal("Response is nil")
	}
	
	if response.PackageName != request.PackageName {
		t.Errorf("Expected package name %s, got %s", request.PackageName, response.PackageName)
	}
	
	if response.FromVersion != request.FromVersion {
		t.Errorf("Expected from version %s, got %s", request.FromVersion, response.FromVersion)
	}
	
	if response.ToVersion != request.ToVersion {
		t.Errorf("Expected to version %s, got %s", request.ToVersion, response.ToVersion)
	}
	
	// Should detect breaking changes
	if !response.HasBreakingChange {
		t.Error("Expected breaking changes to be detected")
	}
	
	if len(response.BreakingChanges) == 0 {
		t.Error("Expected breaking changes list to be populated")
	}
}

func TestVersionDiffAnalysisWithFallback(t *testing.T) {
	// Initialize with heuristic provider only
	config := DefaultAIConfig()
	err := InitializeWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to initialize AI manager: %v", err)
	}
	
	ctx := context.Background()
	request := &types.VersionDiffAnalysisRequest{
		PackageName:    "test-package",
		FromVersion:    "1.0.0",
		ToVersion:      "1.1.0",
		PackageManager: "npm",
		Language:       "javascript",
		DiffText:       "Added new features and bug fixes",
		FileChanges: []types.FileChange{
			{
				Path:         "src/main.js",
				Type:         "modified",
				LinesAdded:   50,
				LinesRemoved: 10,
			},
		},
	}
	
	response, err := AnalyzeVersionDiff(ctx, request)
	if err != nil {
		t.Fatalf("Version diff analysis failed: %v", err)
	}
	
	if response == nil {
		t.Fatal("Response is nil")
	}
	
	if response.PackageName != request.PackageName {
		t.Errorf("Expected package name %s, got %s", request.PackageName, response.PackageName)
	}
	
	if response.UpdateType == "" {
		t.Error("Expected update type to be set")
	}
}

func TestCompatibilityPredictionWithFallback(t *testing.T) {
	// Initialize with heuristic provider only
	config := DefaultAIConfig()
	err := InitializeWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to initialize AI manager: %v", err)
	}
	
	ctx := context.Background()
	request := &types.CompatibilityPredictionRequest{
		PackageName:    "test-package",
		FromVersion:    "1.0.0",
		ToVersion:      "2.0.0",
		PackageManager: "npm",
		Language:       "javascript",
		ProjectContext: types.ProjectContext{
			Framework:       "react",
			LanguageVersion: "18.0.0",
			BuildSystem:     "webpack",
		},
		DependencyGraph: []types.DependencyInfo{
			{
				Name:    "react",
				Version: "18.0.0",
				Type:    "direct",
			},
		},
	}
	
	response, err := PredictCompatibility(ctx, request)
	if err != nil {
		t.Fatalf("Compatibility prediction failed: %v", err)
	}
	
	if response == nil {
		t.Fatal("Response is nil")
	}
	
	if response.PackageName != request.PackageName {
		t.Errorf("Expected package name %s, got %s", request.PackageName, response.PackageName)
	}
	
	if response.CompatibilityScore < 0 || response.CompatibilityScore > 1 {
		t.Errorf("Compatibility score should be between 0 and 1, got %f", response.CompatibilityScore)
	}
}

func TestUpdateClassificationWithFallback(t *testing.T) {
	// Initialize with heuristic provider only
	config := DefaultAIConfig()
	err := InitializeWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to initialize AI manager: %v", err)
	}
	
	ctx := context.Background()
	request := &types.UpdateClassificationRequest{
		PackageName:    "test-package",
		FromVersion:    "1.0.0",
		ToVersion:      "1.0.1",
		PackageManager: "npm",
		Language:       "javascript",
		ChangelogText:  "## Bug Fixes\n- Fixed critical security vulnerability\n- Resolved memory leak",
		ReleaseNotes:   "Security patch release",
	}
	
	response, err := ClassifyUpdate(ctx, request)
	if err != nil {
		t.Fatalf("Update classification failed: %v", err)
	}
	
	if response == nil {
		t.Fatal("Response is nil")
	}
	
	if response.PackageName != request.PackageName {
		t.Errorf("Expected package name %s, got %s", request.PackageName, response.PackageName)
	}
	
	if response.UpdateType == "" {
		t.Error("Expected update type to be set")
	}
	
	if response.Priority == "" {
		t.Error("Expected priority to be set")
	}
}

func TestProviderFallbackMechanism(t *testing.T) {
	// Create a config with multiple providers but invalid API keys
	config := &AIConfig{
		DefaultProvider:         "openai",
		FallbackProviders:       []string{"openai", "claude", "heuristic"},
		EnableHeuristicFallback: true,
		OpenAI: &openai.OpenAIConfig{
			APIKey:      "invalid-key",
			Model:       "gpt-4",
			MaxTokens:   4096,
			Temperature: 0.1,
			Timeout:     5 * time.Second,
			BaseURL:     "https://api.openai.com/v1",
		},
		Claude: &claude.ClaudeConfig{
			APIKey:      "invalid-key",
			Model:       "claude-3-5-sonnet-20241022",
			MaxTokens:   4096,
			Temperature: 0.1,
			Timeout:     5 * time.Second,
			BaseURL:     "https://api.anthropic.com",
		},
		MaxRetries:     1,
		RetryDelay:     1 * time.Second,
		RequestTimeout: 10 * time.Second,
	}
	
	err := InitializeWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to initialize AI manager: %v", err)
	}
	
	ctx := context.Background()
	request := &types.ChangelogAnalysisRequest{
		PackageName:    "test-package",
		FromVersion:    "1.0.0",
		ToVersion:      "2.0.0",
		PackageManager: "npm",
		Language:       "javascript",
		ChangelogText:  "Breaking changes in this release",
		ReleaseNotes:   "Major version update",
	}
	
	// This should fallback to heuristic provider since AI providers will fail with invalid keys
	response, err := AnalyzeChangelog(ctx, request)
	if err != nil {
		t.Fatalf("Analysis should succeed with fallback: %v", err)
	}
	
	if response == nil {
		t.Fatal("Response should not be nil")
	}
	
	// Verify the response is from heuristic analysis
	if response.PackageName != request.PackageName {
		t.Errorf("Expected package name %s, got %s", request.PackageName, response.PackageName)
	}
}

func BenchmarkChangelogAnalysis(b *testing.B) {
	config := DefaultAIConfig()
	err := InitializeWithConfig(config)
	if err != nil {
		b.Fatalf("Failed to initialize AI manager: %v", err)
	}
	
	ctx := context.Background()
	request := &types.ChangelogAnalysisRequest{
		PackageName:    "benchmark-package",
		FromVersion:    "1.0.0",
		ToVersion:      "2.0.0",
		PackageManager: "npm",
		Language:       "javascript",
		ChangelogText:  "## Breaking Changes\n- API changes\n- Removed features",
		ReleaseNotes:   "Major release with breaking changes",
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := AnalyzeChangelog(ctx, request)
		if err != nil {
			b.Fatalf("Analysis failed: %v", err)
		}
	}
}
