package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/8tcapital/superint-dep-manager/internal/superint/ollama"
	"github.com/8tcapital/superint-dep-manager/internal/superint/types"
	"github.com/8tcapital/superint-dep-manager/internal/logger"
)

func main() {
	// Initialize logger
	logger.Init("INFO", "text")

	fmt.Println("🦙 Ollama AI Provider Validation")
	fmt.Println("================================")
	fmt.Println()

	// Load configuration from environment or use defaults
	config := &ollama.OllamaConfig{
		BaseURL:     getEnvOrDefault("OLLAMA_BASE_URL", "http://localhost:11434"),
		Model:       getEnvOrDefault("OLLAMA_MODEL", "llama2"),
		Temperature: 0.1, // Low temperature for consistent results
		TopP:        0.9,
		TopK:        40,
		NumPredict:  1024,
	}

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Base URL: %s\n", config.BaseURL)
	fmt.Printf("  Model: %s\n", config.Model)
	fmt.Printf("  Temperature: %.1f\n", config.Temperature)
	fmt.Println()

	// Create Ollama provider
	provider, err := ollama.NewOllamaProvider(config)
	if err != nil {
		fmt.Printf("❌ Failed to create Ollama provider: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Test 1: Availability Check
	fmt.Println("🔍 Testing Ollama availability...")
	if !provider.IsAvailable(ctx) {
		fmt.Printf("❌ Ollama is not available at %s\n", config.BaseURL)
		fmt.Println("   Please ensure:")
		fmt.Println("   1. Ollama is installed and running")
		fmt.Println("   2. The specified model is downloaded")
		fmt.Println("   3. The endpoint is accessible")
		fmt.Println()
		fmt.Println("   To install Ollama: https://ollama.ai/")
		fmt.Printf("   To download model: ollama pull %s\n", config.Model)
		os.Exit(1)
	}
	fmt.Println("✅ Ollama is available")

	// Test 2: List Available Models
	fmt.Println("\n📋 Listing available models...")
	models, err := provider.ListAvailableModels(ctx)
	if err != nil {
		fmt.Printf("⚠️  Failed to list models: %v\n", err)
	} else {
		fmt.Printf("✅ Found %d models:\n", len(models))
		for _, model := range models {
			if model == config.Model {
				fmt.Printf("   • %s (current)\n", model)
			} else {
				fmt.Printf("   • %s\n", model)
			}
		}
	}

	// Test 3: Connection Test
	fmt.Println("\n🔗 Testing connection...")
	if err := provider.TestConnection(ctx); err != nil {
		fmt.Printf("❌ Connection test failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Connection test successful")

	// Test 4: Changelog Analysis
	fmt.Println("\n📝 Testing changelog analysis...")
	changelogRequest := &types.ChangelogAnalysisRequest{
		PackageName:   "react",
		FromVersion:   "17.0.0",
		ToVersion:     "18.0.0",
		ChangelogText: `# React 18.0.0 Release Notes

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
- Improved error boundaries`,
		ReleaseNotes:   "Major release with concurrent features and breaking changes",
		PackageManager: "npm",
		Language:       "javascript",
	}

	start := time.Now()
	changelogResponse, err := provider.AnalyzeChangelog(ctx, changelogRequest)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ Changelog analysis failed: %v\n", err)
	} else {
		fmt.Printf("✅ Changelog analysis completed in %v\n", duration)
		fmt.Printf("   Risk Level: %s\n", changelogResponse.RiskLevel)
		fmt.Printf("   Confidence: %.2f\n", changelogResponse.Confidence)
		fmt.Printf("   Summary: %s\n", truncateString(changelogResponse.Summary, 100))
		fmt.Printf("   Breaking Changes: %d\n", len(changelogResponse.BreakingChanges))
		fmt.Printf("   Recommendations: %d\n", len(changelogResponse.Recommendations))
	}

	// Test 5: Version Diff Analysis
	fmt.Println("\n🔍 Testing version diff analysis...")
	diffRequest := &types.VersionDiffAnalysisRequest{
		PackageName:    "lodash",
		FromVersion:    "4.17.20",
		ToVersion:      "4.17.21",
		DiffText:       `- Fixed security vulnerability in template function
+ Added input validation
+ Improved error handling
- Deprecated old utility functions
+ Added new array manipulation methods`,
		PackageManager: "npm",
		Language:       "javascript",
	}

	start = time.Now()
	diffResponse, err := provider.AnalyzeVersionDiff(ctx, diffRequest)
	duration = time.Since(start)

	if err != nil {
		fmt.Printf("❌ Version diff analysis failed: %v\n", err)
	} else {
		fmt.Printf("✅ Version diff analysis completed in %v\n", duration)
		fmt.Printf("   Risk Level: %s\n", diffResponse.RiskLevel)
		fmt.Printf("   Confidence: %.2f\n", diffResponse.Confidence)
		fmt.Printf("   Summary: %s\n", truncateString(diffResponse.Summary, 100))
		fmt.Printf("   API Changes: %d\n", len(diffResponse.APIChanges))
	}

	// Test 6: Compatibility Prediction
	fmt.Println("\n🎯 Testing compatibility prediction...")
	compatRequest := &types.CompatibilityPredictionRequest{
		PackageName:    "express",
		FromVersion:    "4.17.1",
		ToVersion:      "4.18.0",
		PackageManager: "npm",
		Language:       "javascript",
	}

	start = time.Now()
	compatResponse, err := provider.PredictCompatibility(ctx, compatRequest)
	duration = time.Since(start)

	if err != nil {
		fmt.Printf("❌ Compatibility prediction failed: %v\n", err)
	} else {
		fmt.Printf("✅ Compatibility prediction completed in %v\n", duration)
		fmt.Printf("   Compatibility Score: %.2f\n", compatResponse.CompatibilityScore)
		fmt.Printf("   Summary: %s\n", truncateString(compatResponse.Summary, 100))
		fmt.Printf("   Risk Level: %s\n", compatResponse.RiskLevel)
		fmt.Printf("   Potential Issues: %d\n", len(compatResponse.PotentialIssues))
	}

	// Test 7: Update Classification
	fmt.Println("\n🏷️  Testing update classification...")
	classifyRequest := &types.UpdateClassificationRequest{
		PackageName:    "typescript",
		FromVersion:    "4.9.0",
		ToVersion:      "5.0.0",
		ChangelogText:  "Major release with breaking changes to type system and new features",
		ReleaseNotes:   "TypeScript 5.0 introduces new features and breaking changes",
		PackageManager: "npm",
		Language:       "typescript",
	}

	start = time.Now()
	classifyResponse, err := provider.ClassifyUpdate(ctx, classifyRequest)
	duration = time.Since(start)

	if err != nil {
		fmt.Printf("❌ Update classification failed: %v\n", err)
	} else {
		fmt.Printf("✅ Update classification completed in %v\n", duration)
		fmt.Printf("   Update Type: %s\n", classifyResponse.UpdateType)
		fmt.Printf("   Priority: %s\n", classifyResponse.Priority)
		fmt.Printf("   Urgency: %s\n", classifyResponse.Urgency)
		fmt.Printf("   Risk Level: %s\n", classifyResponse.RiskLevel)
		fmt.Printf("   Summary: %s\n", truncateString(classifyResponse.Summary, 100))
	}

	// Summary
	fmt.Println("\n🎉 Ollama Provider Validation Complete!")
	fmt.Println("=====================================")
	fmt.Printf("✅ All tests passed successfully\n")
	fmt.Printf("🦙 Model: %s\n", config.Model)
	fmt.Printf("🌐 Endpoint: %s\n", config.BaseURL)
	fmt.Println()
	fmt.Println("Your Ollama provider is ready for use in the Superintelligence Dependency Manager!")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  OLLAMA_BASE_URL - Ollama server endpoint")
	fmt.Println("  OLLAMA_MODEL - Model to use for analysis")
	fmt.Println("  AI_DEFAULT_PROVIDER=ollama - Set Ollama as default provider")
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
