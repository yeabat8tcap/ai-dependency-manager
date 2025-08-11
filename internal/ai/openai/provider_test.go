package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/ai/types"
)

func TestOpenAIProviderInitialization(t *testing.T) {
	tests := []struct {
		name        string
		config      *OpenAIConfig
		expectError bool
	}{
		{
			name: "Valid configuration",
			config: &OpenAIConfig{
				APIKey:      "test-key",
				Model:       "gpt-4",
				MaxTokens:   4096,
				Temperature: 0.1,
				Timeout:     30 * time.Second,
				BaseURL:     "https://api.openai.com/v1",
			},
			expectError: false,
		},
		{
			name: "Missing API key",
			config: &OpenAIConfig{
				Model:       "gpt-4",
				MaxTokens:   4096,
				Temperature: 0.1,
				Timeout:     30 * time.Second,
				BaseURL:     "https://api.openai.com/v1",
			},
			expectError: true,
		},
		{
			name: "Default values applied",
			config: &OpenAIConfig{
				APIKey: "test-key",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewOpenAIProvider(tt.config)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if provider == nil {
				t.Error("Provider should not be nil")
				return
			}
			
			if provider.GetName() != "openai" {
				t.Errorf("Expected provider name 'openai', got %s", provider.GetName())
			}
			
			if provider.GetVersion() == "" {
				t.Error("Provider version should not be empty")
			}
		})
	}
}

func TestOpenAIChangelogAnalysis(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("Expected path /chat/completions, got %s", r.URL.Path)
		}
		
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		
		// Mock successful response
		response := ChatCompletionResponse{
			ID:      "test-id",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "gpt-4",
			Choices: []Choice{
				{
					Index: 0,
					Message: Message{
						Role: "assistant",
						Content: `{
							"package_name": "test-package",
							"from_version": "1.0.0",
							"to_version": "2.0.0",
							"has_breaking_change": true,
							"breaking_changes": [
								{
									"type": "api_removal",
									"description": "Removed deprecated API",
									"impact": "Code using old API will break",
									"severity": "high",
									"confidence": 0.9,
									"mitigation": "Use new API",
									"affected_apis": ["oldFunction"]
								}
							],
							"new_features": [],
							"bug_fixes": [],
							"security_fixes": [],
							"deprecations": [],
							"risk_level": "high",
							"risk_score": 7.5,
							"confidence": 0.85,
							"summary": "Major version with breaking changes",
							"recommendations": ["Test thoroughly", "Update API usage"],
							"migration_steps": ["Replace old API calls"],
							"testing_advice": ["Run full test suite"],
							"recommended_timeline": "within_month",
							"business_impact": "High impact due to breaking changes"
						}`,
					},
					FinishReason: "stop",
				},
			},
			Usage: Usage{
				PromptTokens:     100,
				CompletionTokens: 200,
				TotalTokens:      300,
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &OpenAIConfig{
		APIKey:      "test-key",
		Model:       "gpt-4",
		MaxTokens:   4096,
		Temperature: 0.1,
		Timeout:     30 * time.Second,
		BaseURL:     server.URL,
	}

	provider, err := NewOpenAIProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	ctx := context.Background()
	request := &types.ChangelogAnalysisRequest{
		PackageName:    "test-package",
		FromVersion:    "1.0.0",
		ToVersion:      "2.0.0",
		PackageManager: "npm",
		Language:       "javascript",
		ChangelogText:  "## Breaking Changes\n- Removed deprecated API",
		ReleaseNotes:   "Major version with breaking changes",
	}

	response, err := provider.AnalyzeChangelog(ctx, request)
	if err != nil {
		t.Fatalf("Analysis failed: %v", err)
	}

	if response == nil {
		t.Fatal("Response is nil")
	}

	// Validate response fields
	if response.PackageName != request.PackageName {
		t.Errorf("Expected package name %s, got %s", request.PackageName, response.PackageName)
	}

	if response.FromVersion != request.FromVersion {
		t.Errorf("Expected from version %s, got %s", request.FromVersion, response.FromVersion)
	}

	if response.ToVersion != request.ToVersion {
		t.Errorf("Expected to version %s, got %s", request.ToVersion, response.ToVersion)
	}

	if !response.HasBreakingChange {
		t.Error("Expected breaking changes to be detected")
	}

	if len(response.BreakingChanges) == 0 {
		t.Error("Expected breaking changes list to be populated")
	}

	if response.RiskLevel != "high" {
		t.Errorf("Expected risk level 'high', got %s", response.RiskLevel)
	}

	if response.Confidence <= 0 || response.Confidence > 1 {
		t.Errorf("Confidence should be between 0 and 1, got %f", response.Confidence)
	}
}

func TestOpenAIVersionDiffAnalysis(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ChatCompletionResponse{
			ID:      "test-id",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "gpt-4",
			Choices: []Choice{
				{
					Index: 0,
					Message: Message{
						Role: "assistant",
						Content: `{
							"package_name": "test-package",
							"from_version": "1.0.0",
							"to_version": "1.1.0",
							"update_type": "minor",
							"semantic_impact": "Minor version update with new features",
							"api_changes": [
								{
									"type": "addition",
									"api": "newFunction",
									"description": "Added new utility function",
									"impact": "No breaking changes",
									"severity": "low",
									"examples": ["newFunction()"],
									"migration": "Optional to use"
								}
							],
							"behavior_changes": [],
							"risk_level": "low",
							"risk_score": 2.0,
							"confidence": 0.9,
							"summary": "Minor update with new features",
							"recommendations": ["Safe to update"],
							"migration_effort": "low",
							"backward_compatibility": true
						}`,
					},
					FinishReason: "stop",
				},
			},
			Usage: Usage{
				PromptTokens:     100,
				CompletionTokens: 200,
				TotalTokens:      300,
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &OpenAIConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	}

	provider, err := NewOpenAIProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	ctx := context.Background()
	request := &types.VersionDiffAnalysisRequest{
		PackageName:    "test-package",
		FromVersion:    "1.0.0",
		ToVersion:      "1.1.0",
		PackageManager: "npm",
		Language:       "javascript",
		DiffText:       "Added new features",
		FileChanges: []types.FileChange{
			{
				Path:         "src/utils.js",
				Type:         "modified",
				LinesAdded:   20,
				LinesRemoved: 0,
			},
		},
	}

	response, err := provider.AnalyzeVersionDiff(ctx, request)
	if err != nil {
		t.Fatalf("Analysis failed: %v", err)
	}

	if response == nil {
		t.Fatal("Response is nil")
	}

	if response.UpdateType != "minor" {
		t.Errorf("Expected update type 'minor', got %s", response.UpdateType)
	}

	if !response.BackwardCompatibility {
		t.Error("Expected backward compatibility to be true")
	}

	if response.MigrationEffort != "low" {
		t.Errorf("Expected migration effort 'low', got %s", response.MigrationEffort)
	}
}

func TestOpenAIErrorHandling(t *testing.T) {
	// Create mock server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": {"message": "Invalid API key", "type": "invalid_request_error"}}`))
	}))
	defer server.Close()

	config := &OpenAIConfig{
		APIKey:  "invalid-key",
		BaseURL: server.URL,
	}

	provider, err := NewOpenAIProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	ctx := context.Background()
	request := &types.ChangelogAnalysisRequest{
		PackageName:    "test-package",
		FromVersion:    "1.0.0",
		ToVersion:      "2.0.0",
		PackageManager: "npm",
		Language:       "javascript",
		ChangelogText:  "Test changelog",
		ReleaseNotes:   "Test release",
	}

	_, err = provider.AnalyzeChangelog(ctx, request)
	if err == nil {
		t.Error("Expected error due to invalid API key")
	}
}

func TestOpenAIPromptGeneration(t *testing.T) {
	config := &OpenAIConfig{
		APIKey: "test-key",
	}

	provider, err := NewOpenAIProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	request := &types.ChangelogAnalysisRequest{
		PackageName:    "test-package",
		FromVersion:    "1.0.0",
		ToVersion:      "2.0.0",
		PackageManager: "npm",
		Language:       "javascript",
		ChangelogText:  "Breaking changes",
		ReleaseNotes:   "Major release",
	}

	prompt := provider.buildChangelogAnalysisPrompt(request)

	// Verify prompt contains key information
	if !contains(prompt, request.PackageName) {
		t.Error("Prompt should contain package name")
	}

	if !contains(prompt, request.FromVersion) {
		t.Error("Prompt should contain from version")
	}

	if !contains(prompt, request.ToVersion) {
		t.Error("Prompt should contain to version")
	}

	if !contains(prompt, request.ChangelogText) {
		t.Error("Prompt should contain changelog text")
	}

	if !contains(prompt, "JSON") {
		t.Error("Prompt should request JSON format")
	}
}

func TestOpenAIJSONExtraction(t *testing.T) {
	config := &OpenAIConfig{
		APIKey: "test-key",
	}

	provider, err := NewOpenAIProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "Plain JSON",
			content:  `{"key": "value"}`,
			expected: `{"key": "value"}`,
		},
		{
			name: "JSON in markdown",
			content: "Here's the analysis:\n```json\n{\"key\": \"value\"}\n```\nThat's it.",
			expected: `{"key": "value"}`,
		},
		{
			name:     "JSON with extra text",
			content:  `Some text before {"key": "value"} and after`,
			expected: `{"key": "value"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.extractJSONFromResponse(tt.content)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func BenchmarkOpenAIPromptGeneration(b *testing.B) {
	config := &OpenAIConfig{
		APIKey: "test-key",
	}

	provider, err := NewOpenAIProvider(config)
	if err != nil {
		b.Fatalf("Failed to create provider: %v", err)
	}

	request := &types.ChangelogAnalysisRequest{
		PackageName:    "benchmark-package",
		FromVersion:    "1.0.0",
		ToVersion:      "2.0.0",
		PackageManager: "npm",
		Language:       "javascript",
		ChangelogText:  "Large changelog with many changes...",
		ReleaseNotes:   "Detailed release notes...",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = provider.buildChangelogAnalysisPrompt(request)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 containsAt(s, substr, 1)))
}

func containsAt(s, substr string, start int) bool {
	if start >= len(s) {
		return false
	}
	if start+len(substr) <= len(s) && s[start:start+len(substr)] == substr {
		return true
	}
	return containsAt(s, substr, start+1)
}
