package ollama

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/ai/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOllamaIntegration tests integration with real Ollama instance
// This test requires Ollama to be running locally with llama2 model installed
func TestOllamaIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if OLLAMA_INTEGRATION_TEST is set
	if os.Getenv("OLLAMA_INTEGRATION_TEST") != "true" {
		t.Skip("Set OLLAMA_INTEGRATION_TEST=true to run integration tests")
	}

	baseURL := os.Getenv("OLLAMA_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "llama2"
	}

	config := &OllamaConfig{
		BaseURL:     baseURL,
		Model:       model,
		Temperature: 0.1, // Low temperature for consistent results
		TopP:        0.9,
		TopK:        40,
		NumPredict:  1024,
	}

	provider, err := NewOllamaProvider(config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	t.Run("availability_check", func(t *testing.T) {
		available := provider.IsAvailable(ctx)
		if !available {
			t.Skip("Ollama is not available - ensure Ollama is running and model is installed")
		}
		assert.True(t, available)
	})

	t.Run("list_models", func(t *testing.T) {
		if !provider.IsAvailable(ctx) {
			t.Skip("Ollama not available")
		}

		models, err := provider.ListAvailableModels(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, models)
		t.Logf("Available models: %v", models)
	})

	t.Run("test_connection", func(t *testing.T) {
		if !provider.IsAvailable(ctx) {
			t.Skip("Ollama not available")
		}

		err := provider.TestConnection(ctx)
		assert.NoError(t, err)
	})

	t.Run("changelog_analysis", func(t *testing.T) {
		if !provider.IsAvailable(ctx) {
			t.Skip("Ollama not available")
		}

		request := &types.ChangelogAnalysisRequest{
			PackageName:   "react",
			FromVersion:   "17.0.0",
			ToVersion:     "18.0.0",
			ChangelogText: `
# React 18.0.0 Release Notes

## Breaking Changes
- Automatic batching is now enabled by default
- Strict mode effects run twice in development
- Consistent useEffect timing

## New Features
- Concurrent features
- Suspense improvements
- New hooks: useId, useDeferredValue, useTransition

## Bug Fixes
- Fixed memory leaks in development mode
- Improved error boundaries
			`,
			ReleaseNotes:   "Major release with concurrent features and breaking changes",
			PackageManager: "npm",
			Language:       "javascript",
		}

		response, err := provider.AnalyzeChangelog(ctx, request)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "react", response.PackageName)
		assert.Equal(t, "17.0.0", response.FromVersion)
		assert.Equal(t, "18.0.0", response.ToVersion)
		assert.NotEmpty(t, response.Summary)
		assert.Greater(t, response.Confidence, 0.0)
		assert.LessOrEqual(t, response.Confidence, 1.0)

		t.Logf("Analysis Summary: %s", response.Summary)
		t.Logf("Risk Level: %s", response.RiskLevel)
		t.Logf("Confidence: %.2f", response.Confidence)
		t.Logf("Breaking Changes: %d", len(response.BreakingChanges))
		t.Logf("Recommendations: %d", len(response.Recommendations))
	})

	t.Run("version_diff_analysis", func(t *testing.T) {
		if !provider.IsAvailable(ctx) {
			t.Skip("Ollama not available")
		}

		request := &types.VersionDiffAnalysisRequest{
			PackageName:    "lodash",
			FromVersion:    "4.17.20",
			ToVersion:      "4.17.21",
			DiffText:       `
- Fixed security vulnerability in template function
+ Added input validation
+ Improved error handling
- Deprecated old utility functions
+ Added new array manipulation methods
			`,
			PackageManager: "npm",
			Language:       "javascript",
		}

		response, err := provider.AnalyzeVersionDiff(ctx, request)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.Summary)
		assert.Greater(t, response.Confidence, 0.0)

		t.Logf("Diff Analysis Summary: %s", response.Summary)
		t.Logf("Risk Level: %s", response.RiskLevel)
		t.Logf("API Changes: %d", len(response.APIChanges))
	})

	t.Run("compatibility_prediction", func(t *testing.T) {
		if !provider.IsAvailable(ctx) {
			t.Skip("Ollama not available")
		}

		request := &types.CompatibilityPredictionRequest{
			PackageName:    "express",
			FromVersion:    "4.17.1",
			ToVersion:      "4.18.0",
			PackageManager: "npm",
			Language:       "javascript",
		}

		response, err := provider.PredictCompatibility(ctx, request)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.GreaterOrEqual(t, response.CompatibilityScore, 0.0)
		assert.LessOrEqual(t, response.CompatibilityScore, 1.0)

		t.Logf("Compatibility Score: %.2f", response.CompatibilityScore)
		t.Logf("Summary: %s", response.Summary)
		t.Logf("Potential Issues: %d", len(response.PotentialIssues))
	})

	t.Run("update_classification", func(t *testing.T) {
		if !provider.IsAvailable(ctx) {
			t.Skip("Ollama not available")
		}

		request := &types.UpdateClassificationRequest{
			PackageName:    "typescript",
			FromVersion:    "4.9.0",
			ToVersion:      "5.0.0",
			ChangelogText:  "Major release with breaking changes to type system",
			ReleaseNotes:   "TypeScript 5.0 major release",
			PackageManager: "npm",
			Language:       "typescript",
		}

		response, err := provider.ClassifyUpdate(ctx, request)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.Summary)

		t.Logf("Update Type: %s", response.UpdateType)
		t.Logf("Priority: %s", response.Priority)
		t.Logf("Urgency: %s", response.Urgency)
		t.Logf("Summary: %s", response.Summary)
	})
}

// TestOllamaPerformance benchmarks Ollama provider performance
func TestOllamaPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	if os.Getenv("OLLAMA_PERFORMANCE_TEST") != "true" {
		t.Skip("Set OLLAMA_PERFORMANCE_TEST=true to run performance tests")
	}

	config := &OllamaConfig{
		BaseURL:     "http://localhost:11434",
		Model:       "llama2",
		Temperature: 0.1,
		NumPredict:  512, // Smaller for faster responses
	}

	provider, err := NewOllamaProvider(config)
	require.NoError(t, err)

	ctx := context.Background()
	if !provider.IsAvailable(ctx) {
		t.Skip("Ollama not available")
	}

	request := &types.ChangelogAnalysisRequest{
		PackageName:    "test-package",
		FromVersion:    "1.0.0",
		ToVersion:      "2.0.0",
		ChangelogText:  "Minor bug fixes and improvements",
		PackageManager: "npm",
		Language:       "javascript",
	}

	// Warmup request
	_, err = provider.AnalyzeChangelog(ctx, request)
	require.NoError(t, err)

	// Performance test
	start := time.Now()
	numRequests := 5

	for i := 0; i < numRequests; i++ {
		_, err := provider.AnalyzeChangelog(ctx, request)
		assert.NoError(t, err)
	}

	duration := time.Since(start)
	avgDuration := duration / time.Duration(numRequests)

	t.Logf("Performance Results:")
	t.Logf("Total time for %d requests: %v", numRequests, duration)
	t.Logf("Average time per request: %v", avgDuration)
	t.Logf("Requests per second: %.2f", float64(numRequests)/duration.Seconds())

	// Assert reasonable performance (adjust thresholds as needed)
	assert.Less(t, avgDuration, 30*time.Second, "Average request should complete within 30 seconds")
}

// TestOllamaModelSwitching tests switching between different models
func TestOllamaModelSwitching(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping model switching test in short mode")
	}

	if os.Getenv("OLLAMA_MODEL_SWITCHING_TEST") != "true" {
		t.Skip("Set OLLAMA_MODEL_SWITCHING_TEST=true to run model switching tests")
	}

	baseURL := "http://localhost:11434"
	
	// Test with different models (if available)
	models := []string{"llama2", "codellama", "mistral"}

	for _, model := range models {
		t.Run("model_"+model, func(t *testing.T) {
			config := &OllamaConfig{
				BaseURL:     baseURL,
				Model:       model,
				Temperature: 0.1,
				NumPredict:  256,
			}

			provider, err := NewOllamaProvider(config)
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
			defer cancel()

			if !provider.IsAvailable(ctx) {
				t.Skipf("Model %s not available", model)
			}

			request := &types.ChangelogAnalysisRequest{
				PackageName:    "test-package",
				FromVersion:    "1.0.0",
				ToVersion:      "1.1.0",
				ChangelogText:  "Added new features and fixed bugs",
				PackageManager: "npm",
				Language:       "javascript",
			}

			start := time.Now()
			response, err := provider.AnalyzeChangelog(ctx, request)
			duration := time.Since(start)

			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.NotEmpty(t, response.Summary)

			t.Logf("Model %s - Response time: %v", model, duration)
			t.Logf("Model %s - Summary: %s", model, response.Summary)
		})
	}
}

// TestOllamaErrorRecovery tests error handling and recovery
func TestOllamaErrorRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping error recovery test in short mode")
	}

	if os.Getenv("OLLAMA_ERROR_RECOVERY_TEST") != "true" {
		t.Skip("Set OLLAMA_ERROR_RECOVERY_TEST=true to run error recovery tests")
	}

	t.Run("invalid_model", func(t *testing.T) {
		config := &OllamaConfig{
			BaseURL: "http://localhost:11434",
			Model:   "nonexistent-model",
		}

		provider, err := NewOllamaProvider(config)
		require.NoError(t, err)

		ctx := context.Background()
		available := provider.IsAvailable(ctx)
		assert.False(t, available)
	})

	t.Run("invalid_endpoint", func(t *testing.T) {
		config := &OllamaConfig{
			BaseURL: "http://localhost:99999",
			Model:   "llama2",
		}

		provider, err := NewOllamaProvider(config)
		require.NoError(t, err)

		ctx := context.Background()
		available := provider.IsAvailable(ctx)
		assert.False(t, available)
	})

	t.Run("timeout_handling", func(t *testing.T) {
		config := &OllamaConfig{
			BaseURL: "http://localhost:11434",
			Model:   "llama2",
		}

		provider, err := NewOllamaProvider(config)
		require.NoError(t, err)

		if !provider.IsAvailable(context.Background()) {
			t.Skip("Ollama not available")
		}

		request := &types.ChangelogAnalysisRequest{
			PackageName:   "test-package",
			FromVersion:   "1.0.0",
			ToVersion:     "2.0.0",
			ChangelogText: "Very long changelog that might take time to process...",
		}

		// Test with very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		_, err = provider.AnalyzeChangelog(ctx, request)
		// Should either succeed quickly or timeout
		if err != nil {
			assert.Contains(t, err.Error(), "context deadline exceeded")
		}
	})
}
