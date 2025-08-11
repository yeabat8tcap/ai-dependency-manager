package claude

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/ai/types"
)

func TestClaudeProviderInitialization(t *testing.T) {
	tests := []struct {
		name        string
		config      *ClaudeConfig
		expectError bool
	}{
		{
			name: "Valid configuration",
			config: &ClaudeConfig{
				APIKey:      "test-key",
				Model:       "claude-3-5-sonnet-20241022",
				MaxTokens:   4096,
				Temperature: 0.1,
				Timeout:     30 * time.Second,
				BaseURL:     "https://api.anthropic.com",
			},
			expectError: false,
		},
		{
			name: "Missing API key",
			config: &ClaudeConfig{
				Model:       "claude-3-5-sonnet-20241022",
				MaxTokens:   4096,
				Temperature: 0.1,
				Timeout:     30 * time.Second,
				BaseURL:     "https://api.anthropic.com",
			},
			expectError: true,
		},
		{
			name: "Default values applied",
			config: &ClaudeConfig{
				APIKey: "test-key",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewClaudeProvider(tt.config)
			
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
			
			if provider.GetName() != "claude" {
				t.Errorf("Expected provider name 'claude', got %s", provider.GetName())
			}
			
			if provider.GetVersion() == "" {
				t.Error("Provider version should not be empty")
			}
		})
	}
}

func TestClaudeChangelogAnalysis(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" {
			t.Errorf("Expected path /v1/messages, got %s", r.URL.Path)
		}
		
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		
		// Check headers
		if r.Header.Get("x-api-key") == "" {
			t.Error("Expected x-api-key header")
		}
		
		if r.Header.Get("anthropic-version") == "" {
			t.Error("Expected anthropic-version header")
		}
		
		// Mock successful response
		response := MessageResponse{
			ID:   "test-id",
			Type: "message",
			Role: "assistant",
			Content: []ContentBlock{
				{
					Type: "text",
					Text: `{
						"package_name": "test-package",
						"from_version": "1.0.0",
						"to_version": "2.0.0",
						"has_breaking_change": true,
						"breaking_changes": [
							{
								"type": "api_removal",
								"description": "Removed deprecated API endpoints",
								"impact": "Applications using old endpoints will fail",
								"severity": "high",
								"confidence": 0.95,
								"mitigation": "Update to use new API endpoints",
								"affected_apis": ["oldEndpoint", "deprecatedFunction"]
							}
						],
						"new_features": [
							{
								"name": "Enhanced Security",
								"description": "Added OAuth2 authentication",
								"type": "security",
								"impact": "Improved security posture",
								"confidence": 0.9,
								"benefits": ["Better security", "OAuth2 support"],
								"usage_example": "Use new auth methods"
							}
						],
						"bug_fixes": [],
						"security_fixes": [
							{
								"description": "Fixed SQL injection vulnerability",
								"severity": "critical",
								"cve": "CVE-2024-1234",
								"cvss": 9.1,
								"impact": "Prevents SQL injection attacks",
								"confidence": 1.0,
								"references": ["https://security.example.com/advisory"],
								"urgency": "critical"
							}
						],
						"deprecations": [],
						"risk_level": "high",
						"risk_score": 8.5,
						"confidence": 0.9,
						"summary": "Major version with breaking changes and critical security fixes",
						"recommendations": ["Update immediately", "Test thoroughly", "Review API usage"],
						"migration_steps": ["Replace deprecated APIs", "Update authentication"],
						"testing_advice": ["Run security tests", "Validate API changes"],
						"recommended_timeline": "immediate",
						"business_impact": "High impact due to security fixes and breaking changes"
					}`,
				},
			},
			Model:      "claude-3-5-sonnet-20241022",
			StopReason: "end_turn",
			Usage: Usage{
				InputTokens:  150,
				OutputTokens: 300,
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &ClaudeConfig{
		APIKey:      "test-key",
		Model:       "claude-3-5-sonnet-20241022",
		MaxTokens:   4096,
		Temperature: 0.1,
		Timeout:     30 * time.Second,
		BaseURL:     server.URL,
	}

	provider, err := NewClaudeProvider(config)
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
		ChangelogText:  "## Breaking Changes\n- Removed deprecated API\n## Security\n- Fixed SQL injection",
		ReleaseNotes:   "Major version with security fixes",
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

	if !response.HasBreakingChange {
		t.Error("Expected breaking changes to be detected")
	}

	if len(response.BreakingChanges) == 0 {
		t.Error("Expected breaking changes list to be populated")
	}

	if len(response.SecurityFixes) == 0 {
		t.Error("Expected security fixes to be detected")
	}

	if response.RiskLevel != "high" {
		t.Errorf("Expected risk level 'high', got %s", response.RiskLevel)
	}

	// Validate security fix details
	securityFix := response.SecurityFixes[0]
	if securityFix.CVE != "CVE-2024-1234" {
		t.Errorf("Expected CVE 'CVE-2024-1234', got %s", securityFix.CVE)
	}

	if securityFix.CVSS != 9.1 {
		t.Errorf("Expected CVSS 9.1, got %f", securityFix.CVSS)
	}
}

func TestClaudeCompatibilityPrediction(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := MessageResponse{
			ID:   "test-id",
			Type: "message",
			Role: "assistant",
			Content: []ContentBlock{
				{
					Type: "text",
					Text: `{
						"package_name": "test-package",
						"from_version": "1.0.0",
						"to_version": "2.0.0",
						"compatibility_score": 0.7,
						"risk_level": "medium",
						"risk_score": 5.5,
						"confidence": 0.85,
						"potential_issues": [
							{
								"type": "breaking_change",
								"description": "API signature changes may cause compilation errors",
								"severity": "medium",
								"likelihood": 0.8,
								"impact": "Code using changed APIs will need updates",
								"mitigation": "Update API calls to match new signatures",
								"detection": "Compile-time errors will indicate affected code"
							}
						],
						"migration_steps": [
							{
								"step": "Update API calls",
								"description": "Replace old API calls with new signatures",
								"priority": "high",
								"effort": "medium",
								"risk": "low",
								"validation": "Run unit tests to verify changes"
							}
						],
						"testing_recommendations": [
							{
								"type": "integration",
								"description": "Test all API integrations thoroughly",
								"priority": "high",
								"test_cases": ["API call validation", "Error handling"],
								"tools": ["Jest", "Mocha"]
							}
						],
						"summary": "Moderate compatibility risk due to API changes",
						"recommendations": ["Plan migration carefully", "Test thoroughly"],
						"estimated_effort": "medium",
						"rollback_complexity": "low"
					}`,
				},
			},
			Model:      "claude-3-5-sonnet-20241022",
			StopReason: "end_turn",
			Usage: Usage{
				InputTokens:  200,
				OutputTokens: 250,
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &ClaudeConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	}

	provider, err := NewClaudeProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
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

	response, err := provider.PredictCompatibility(ctx, request)
	if err != nil {
		t.Fatalf("Analysis failed: %v", err)
	}

	if response == nil {
		t.Fatal("Response is nil")
	}

	if response.CompatibilityScore != 0.7 {
		t.Errorf("Expected compatibility score 0.7, got %f", response.CompatibilityScore)
	}

	if response.RiskLevel != "medium" {
		t.Errorf("Expected risk level 'medium', got %s", response.RiskLevel)
	}

	if len(response.PotentialIssues) == 0 {
		t.Error("Expected potential issues to be identified")
	}

	if len(response.MigrationSteps) == 0 {
		t.Error("Expected migration steps to be provided")
	}
}

func TestClaudeErrorHandling(t *testing.T) {
	// Create mock server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{
			"type": "error",
			"error": {
				"type": "authentication_error",
				"message": "Invalid API key"
			}
		}`))
	}))
	defer server.Close()

	config := &ClaudeConfig{
		APIKey:  "invalid-key",
		BaseURL: server.URL,
	}

	provider, err := NewClaudeProvider(config)
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

	if !contains(err.Error(), "authentication_error") {
		t.Errorf("Expected authentication error, got: %v", err)
	}
}

func TestClaudeJSONExtraction(t *testing.T) {
	config := &ClaudeConfig{
		APIKey: "test-key",
	}

	provider, err := NewClaudeProvider(config)
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
		{
			name: "Complex JSON with nested objects",
			content: `Analysis: {"outer": {"inner": "value"}, "array": [1, 2, 3]} Done.`,
			expected: `{"outer": {"inner": "value"}, "array": [1, 2, 3]}`,
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

func TestClaudeSystemPrompts(t *testing.T) {
	config := &ClaudeConfig{
		APIKey: "test-key",
	}

	provider, err := NewClaudeProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	tests := []struct {
		name   string
		prompt string
	}{
		{
			name:   "Changelog system prompt",
			prompt: provider.getChangelogSystemPrompt(),
		},
		{
			name:   "Version diff system prompt",
			prompt: provider.getVersionDiffSystemPrompt(),
		},
		{
			name:   "Compatibility system prompt",
			prompt: provider.getCompatibilitySystemPrompt(),
		},
		{
			name:   "Update classification system prompt",
			prompt: provider.getUpdateClassificationSystemPrompt(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.prompt == "" {
				t.Error("System prompt should not be empty")
			}

			if !contains(tt.prompt, "JSON") {
				t.Error("System prompt should mention JSON format")
			}

			if len(tt.prompt) < 100 {
				t.Error("System prompt seems too short")
			}
		})
	}
}

func TestClaudePromptGeneration(t *testing.T) {
	config := &ClaudeConfig{
		APIKey: "test-key",
	}

	provider, err := NewClaudeProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	request := &types.UpdateClassificationRequest{
		PackageName:    "test-package",
		FromVersion:    "1.0.0",
		ToVersion:      "1.0.1",
		PackageManager: "npm",
		Language:       "javascript",
		ChangelogText:  "Security fixes",
		ReleaseNotes:   "Patch release",
	}

	prompt := provider.buildUpdateClassificationPrompt(request)

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

	if !contains(prompt, "security") {
		t.Error("Prompt should contain classification categories")
	}
}

func BenchmarkClaudePromptGeneration(b *testing.B) {
	config := &ClaudeConfig{
		APIKey: "test-key",
	}

	provider, err := NewClaudeProvider(config)
	if err != nil {
		b.Fatalf("Failed to create provider: %v", err)
	}

	request := &types.ChangelogAnalysisRequest{
		PackageName:    "benchmark-package",
		FromVersion:    "1.0.0",
		ToVersion:      "2.0.0",
		PackageManager: "npm",
		Language:       "javascript",
		ChangelogText:  "Large changelog with many changes and detailed descriptions...",
		ReleaseNotes:   "Comprehensive release notes with extensive details...",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = provider.buildChangelogAnalysisPrompt(request)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
