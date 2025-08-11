package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/ai/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOllamaProvider(t *testing.T) {
	tests := []struct {
		name        string
		config      *OllamaConfig
		expectError bool
	}{
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
		},
		{
			name: "valid config with defaults",
			config: &OllamaConfig{
				BaseURL: "http://localhost:11434",
				Model:   "llama2",
			},
			expectError: false,
		},
		{
			name: "empty config with defaults applied",
			config: &OllamaConfig{},
			expectError: false,
		},
		{
			name: "custom config",
			config: &OllamaConfig{
				BaseURL:     "http://custom:8080",
				Model:       "codellama",
				Temperature: 0.5,
				TopP:        0.8,
				TopK:        30,
				NumPredict:  1024,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewOllamaProvider(tt.config)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, provider)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
				assert.Equal(t, "ollama", provider.GetName())
				assert.Equal(t, "1.0.0", provider.GetVersion())
				
				// Check defaults were applied
				if tt.config != nil && tt.config.Model == "" {
					assert.Equal(t, "llama2", provider.config.Model)
				}
				if tt.config != nil && tt.config.BaseURL == "" {
					assert.Equal(t, "http://localhost:11434", provider.config.BaseURL)
				}
			}
		})
	}
}

func TestOllamaProvider_AnalyzeChangelog(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			response := ChatResponse{
				Message: Message{
					Role: "assistant",
					Content: `{
						"risk_level": "medium",
						"confidence": 0.85,
						"summary": "This update includes breaking changes to the API",
						"breaking_changes": [
							{
								"type": "api_change",
								"description": "Removed deprecated method",
								"impact": "high",
								"migration_required": true
							}
						],
						"recommendations": [
							"Review API usage before updating",
							"Test thoroughly in staging environment"
						],
						"security_implications": ["No security concerns identified"]
					}`,
				},
				Done: true,
			}
			json.NewEncoder(w).Encode(response)
		} else if r.URL.Path == "/api/tags" {
			response := ModelsResponse{
				Models: []ModelInfo{
					{Name: "llama2", Size: 3825819519},
				},
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	config := &OllamaConfig{
		BaseURL:     server.URL,
		Model:       "llama2",
		Temperature: 0.7,
	}

	provider, err := NewOllamaProvider(config)
	require.NoError(t, err)

	request := &types.ChangelogAnalysisRequest{
		PackageName:    "test-package",
		FromVersion:    "1.0.0",
		ToVersion:      "2.0.0",
		ChangelogText:  "Breaking: Removed deprecated API methods",
		ReleaseNotes:   "Major version update with API changes",
		PackageManager: "npm",
		Language:       "javascript",
	}

	ctx := context.Background()
	response, err := provider.AnalyzeChangelog(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "test-package", response.PackageName)
	assert.Equal(t, "1.0.0", response.FromVersion)
	assert.Equal(t, "2.0.0", response.ToVersion)
	assert.Equal(t, types.RiskLevelMedium, response.RiskLevel)
	assert.Equal(t, 0.85, response.Confidence)
	assert.Contains(t, response.Summary, "breaking changes")
	assert.Len(t, response.BreakingChanges, 1)
	assert.Len(t, response.Recommendations, 2)
}

func TestOllamaProvider_AnalyzeVersionDiff(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			response := ChatResponse{
				Message: Message{
					Role: "assistant",
					Content: `{
						"risk_level": "low",
						"confidence": 0.92,
						"summary": "Minor bug fixes and performance improvements",
						"changes": [
							{
								"type": "bug_fix",
								"description": "Fixed memory leak in parser",
								"impact": "low",
								"files_affected": ["parser.js"]
							}
						],
						"compatibility_assessment": "fully_compatible",
						"migration_complexity": "none",
						"recommendations": ["Safe to update immediately"]
					}`,
				},
				Done: true,
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	config := &OllamaConfig{
		BaseURL: server.URL,
		Model:   "llama2",
	}

	provider, err := NewOllamaProvider(config)
	require.NoError(t, err)

	request := &types.VersionDiffAnalysisRequest{
		PackageName:    "test-package",
		FromVersion:    "1.0.0",
		ToVersion:      "1.0.1",
		DiffText:       "- Fixed memory leak\n+ Improved performance",
		PackageManager: "npm",
		Language:       "javascript",
	}

	ctx := context.Background()
	response, err := provider.AnalyzeVersionDiff(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, types.RiskLevelLow, response.RiskLevel)
	assert.Equal(t, 0.92, response.Confidence)
	assert.Contains(t, response.Summary, "bug fixes")
	assert.Len(t, response.APIChanges, 1)
	assert.NotEmpty(t, response.Summary)
}

func TestOllamaProvider_PredictCompatibility(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			response := ChatResponse{
				Message: Message{
					Role: "assistant",
					Content: `{
						"compatibility_score": 0.88,
						"risk_level": "medium",
						"confidence": 0.91,
						"assessment": "mostly_compatible",
						"potential_issues": [
							{
								"type": "api_deprecation",
								"description": "Some methods may be deprecated",
								"severity": "medium",
								"likelihood": 0.6
							}
						],
						"recommendations": [
							"Test in development environment first",
							"Review deprecation warnings"
						]
					}`,
				},
				Done: true,
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	config := &OllamaConfig{
		BaseURL: server.URL,
		Model:   "llama2",
	}

	provider, err := NewOllamaProvider(config)
	require.NoError(t, err)

	request := &types.CompatibilityPredictionRequest{
		PackageName:    "test-package",
		FromVersion:    "1.0.0",
		ToVersion:      "2.0.0",
		PackageManager: "npm",
		Language:       "javascript",
	}

	ctx := context.Background()
	response, err := provider.PredictCompatibility(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 0.88, response.CompatibilityScore)
	assert.Equal(t, types.RiskLevelMedium, response.RiskLevel)
	assert.NotEmpty(t, response.Summary)
	assert.Len(t, response.PotentialIssues, 1)
}

func TestOllamaProvider_ClassifyUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			response := ChatResponse{
				Message: Message{
					Role: "assistant",
					Content: `{
						"update_type": "major",
						"priority": "high",
						"urgency": "medium",
						"risk_level": "high",
						"confidence": 0.94,
						"reasoning": "Breaking changes require immediate attention",
						"impact_assessment": {
							"breaking_changes": true,
							"security_fixes": false,
							"performance_improvements": true,
							"new_features": true
						},
						"recommended_action": "plan_carefully",
						"timeline_recommendation": "within_month"
					}`,
				},
				Done: true,
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	config := &OllamaConfig{
		BaseURL: server.URL,
		Model:   "llama2",
	}

	provider, err := NewOllamaProvider(config)
	require.NoError(t, err)

	request := &types.UpdateClassificationRequest{
		PackageName:    "test-package",
		FromVersion:    "1.0.0",
		ToVersion:      "2.0.0",
		ChangelogText:  "Breaking: Major API overhaul",
		ReleaseNotes:   "Major version update with breaking changes",
		PackageManager: "npm",
		Language:       "javascript",
	}

	ctx := context.Background()
	response, err := provider.ClassifyUpdate(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, types.UpdateTypeMajor, response.UpdateType)
	assert.Equal(t, types.PriorityHigh, response.Priority)
	assert.Equal(t, types.UrgencyMedium, response.Urgency)
	assert.Equal(t, types.RiskLevelHigh, response.RiskLevel)
	assert.NotEmpty(t, response.Summary)
}

func TestOllamaProvider_IsAvailable(t *testing.T) {
	// Test with available server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			response := ModelsResponse{
				Models: []ModelInfo{
					{Name: "llama2", Size: 3825819519},
				},
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	config := &OllamaConfig{
		BaseURL: server.URL,
		Model:   "llama2",
	}

	provider, err := NewOllamaProvider(config)
	require.NoError(t, err)

	ctx := context.Background()
	available := provider.IsAvailable(ctx)
	assert.True(t, available)

	// Test with unavailable server
	config.BaseURL = "http://localhost:99999"
	provider, err = NewOllamaProvider(config)
	require.NoError(t, err)

	available = provider.IsAvailable(ctx)
	assert.False(t, available)
}

func TestOllamaProvider_ListAvailableModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			response := ModelsResponse{
				Models: []ModelInfo{
					{Name: "llama2", Size: 3825819519},
					{Name: "codellama", Size: 3825819519},
					{Name: "mistral", Size: 4109016519},
				},
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	config := &OllamaConfig{
		BaseURL: server.URL,
		Model:   "llama2",
	}

	provider, err := NewOllamaProvider(config)
	require.NoError(t, err)

	ctx := context.Background()
	models, err := provider.ListAvailableModels(ctx)

	assert.NoError(t, err)
	assert.Len(t, models, 3)
	assert.Contains(t, models, "llama2")
	assert.Contains(t, models, "codellama")
	assert.Contains(t, models, "mistral")
}

func TestOllamaProvider_TestConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			response := ModelsResponse{
				Models: []ModelInfo{
					{Name: "llama2", Size: 3825819519},
				},
			}
			json.NewEncoder(w).Encode(response)
		} else if r.URL.Path == "/api/chat" {
			response := ChatResponse{
				Message: Message{
					Role:    "assistant",
					Content: "OK",
				},
				Done: true,
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	config := &OllamaConfig{
		BaseURL: server.URL,
		Model:   "llama2",
	}

	provider, err := NewOllamaProvider(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = provider.TestConnection(ctx)
	assert.NoError(t, err)
}

func TestOllamaProvider_ExtractJSON(t *testing.T) {
	provider := &OllamaProvider{}

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "json in code block",
			content:  "Here's the analysis:\n```json\n{\"key\": \"value\"}\n```\nDone!",
			expected: "{\"key\": \"value\"}",
		},
		{
			name:     "plain json object",
			content:  "Analysis: {\"result\": \"success\", \"score\": 0.85} - complete",
			expected: "{\"result\": \"success\", \"score\": 0.85}",
		},
		{
			name:     "nested json object",
			content:  "Result: {\"data\": {\"nested\": true}, \"count\": 1}",
			expected: "{\"data\": {\"nested\": true}, \"count\": 1}",
		},
		{
			name:     "no json found",
			content:  "This is just plain text without any JSON",
			expected: "",
		},
		{
			name:     "malformed json",
			content:  "Broken: {\"incomplete\": ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.extractJSON(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOllamaProvider_ErrorHandling(t *testing.T) {
	// Test server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	config := &OllamaConfig{
		BaseURL: server.URL,
		Model:   "llama2",
	}

	provider, err := NewOllamaProvider(config)
	require.NoError(t, err)

	request := &types.ChangelogAnalysisRequest{
		PackageName: "test-package",
		FromVersion: "1.0.0",
		ToVersion:   "2.0.0",
	}

	ctx := context.Background()
	_, err = provider.AnalyzeChangelog(ctx, request)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ollama API error")
}

func TestOllamaProvider_Timeout(t *testing.T) {
	// Test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		response := ChatResponse{
			Message: Message{
				Role:    "assistant",
				Content: `{"risk_level": "low"}`,
			},
			Done: true,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &OllamaConfig{
		BaseURL: server.URL,
		Model:   "llama2",
	}

	provider, err := NewOllamaProvider(config)
	require.NoError(t, err)

	request := &types.ChangelogAnalysisRequest{
		PackageName: "test-package",
		FromVersion: "1.0.0",
		ToVersion:   "2.0.0",
	}

	// Test with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err = provider.AnalyzeChangelog(ctx, request)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "context deadline exceeded") || 
		strings.Contains(err.Error(), "timeout"))
}
