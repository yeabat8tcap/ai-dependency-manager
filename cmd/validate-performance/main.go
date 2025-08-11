package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/ai"
	"github.com/8tcapital/ai-dep-manager/internal/ai/types"
	"github.com/8tcapital/ai-dep-manager/internal/logger"
)

func main() {
	fmt.Println("🚀 AI Provider Performance Validation")
	fmt.Println("=====================================")

	// Initialize logger
	logger.Init("info", "text")

	// Initialize AI system
	if err := ai.Initialize(); err != nil {
		fmt.Printf("❌ Failed to initialize AI system: %v\n", err)
		os.Exit(1)
	}

	// Test data for performance comparison
	testCases := []struct {
		name        string
		packageName string
		fromVersion string
		toVersion   string
		changelog   string
	}{
		{
			name:        "React Major Update",
			packageName: "react",
			fromVersion: "17.0.0",
			toVersion:   "18.0.0",
			changelog: `# React 18.0.0

## Breaking Changes
- Automatic batching changes
- Stricter StrictMode
- Consistent useEffect timing

## New Features  
- Concurrent rendering
- Suspense improvements
- New hooks: useId, useDeferredValue, useTransition

## Bug Fixes
- Fixed memory leaks in development
- Improved error boundaries`,
		},
		{
			name:        "Express Security Update",
			packageName: "express",
			fromVersion: "4.17.1",
			toVersion:   "4.18.2",
			changelog: `# Express 4.18.2

## Security Fixes
- Fixed prototype pollution vulnerability
- Updated dependencies with security patches
- Improved input validation

## Bug Fixes
- Fixed route parameter parsing
- Improved error handling`,
		},
		{
			name:        "Lodash Minor Update",
			packageName: "lodash",
			fromVersion: "4.17.20",
			toVersion:   "4.17.21",
			changelog: `# Lodash 4.17.21

## Bug Fixes
- Fixed template injection vulnerability
- Improved type definitions
- Performance optimizations`,
		},
	}

	// Get available providers
	providers := []string{"openai", "claude", "ollama", "heuristic"}
	results := make(map[string][]PerformanceResult)

	fmt.Println("\n🔍 Testing AI Providers...")
	
	for _, provider := range providers {
		fmt.Printf("\n📊 Testing %s provider:\n", provider)
		
		if !isProviderAvailable(provider) {
			fmt.Printf("   ⚠️  Provider %s not available, skipping...\n", provider)
			continue
		}

		var providerResults []PerformanceResult
		
		for _, testCase := range testCases {
			fmt.Printf("   🧪 %s... ", testCase.name)
			
			result := testProviderPerformance(provider, testCase)
			providerResults = append(providerResults, result)
			
			if result.Success {
				fmt.Printf("✅ %.2fs (confidence: %.1f%%)\n", result.Duration.Seconds(), result.Confidence*100)
			} else {
				fmt.Printf("❌ Failed: %v\n", result.Error)
			}
		}
		
		results[provider] = providerResults
	}

	// Display performance comparison
	displayPerformanceComparison(results)
}

type PerformanceResult struct {
	TestCase   string
	Provider   string
	Duration   time.Duration
	Success    bool
	Error      error
	Confidence float64
	RiskLevel  string
}

func isProviderAvailable(provider string) bool {
	switch provider {
	case "openai":
		return os.Getenv("OPENAI_API_KEY") != ""
	case "claude":
		return os.Getenv("CLAUDE_API_KEY") != ""
	case "ollama":
		// Check if Ollama is running
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		// Try to get a provider instance to test availability
		if provider, exists := ai.GetProvider("ollama"); exists {
			return provider.IsAvailable(ctx)
		}
		return false
	case "heuristic":
		return true // Always available
	default:
		return false
	}
}

func testProviderPerformance(providerName string, testCase struct {
	name        string
	packageName string
	fromVersion string
	toVersion   string
	changelog   string
}) PerformanceResult {
	
	start := time.Now()
	
	// Get provider
	provider, exists := ai.GetProvider(providerName)
	if !exists {
		return PerformanceResult{
			TestCase: testCase.name,
			Provider: providerName,
			Duration: time.Since(start),
			Success:  false,
			Error:    fmt.Errorf("provider %s not found", providerName),
		}
	}

	// Create analysis request
	request := &types.ChangelogAnalysisRequest{
		PackageName:   testCase.packageName,
		FromVersion:   testCase.fromVersion,
		ToVersion:     testCase.toVersion,
		ChangelogText: testCase.changelog,
		ReleaseNotes:  testCase.changelog,
	}

	// Perform analysis with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	response, err := provider.AnalyzeChangelog(ctx, request)
	duration := time.Since(start)

	if err != nil {
		return PerformanceResult{
			TestCase: testCase.name,
			Provider: providerName,
			Duration: duration,
			Success:  false,
			Error:    err,
		}
	}

	return PerformanceResult{
		TestCase:   testCase.name,
		Provider:   providerName,
		Duration:   duration,
		Success:    true,
		Confidence: response.ConfidenceScore,
		RiskLevel:  string(response.RiskLevel),
	}
}

func displayPerformanceComparison(results map[string][]PerformanceResult) {
	fmt.Println("\n📈 Performance Comparison Results")
	fmt.Println("=================================")

	// Calculate averages for each provider
	fmt.Printf("%-12s %-12s %-12s %-12s %-12s\n", "Provider", "Avg Time", "Success Rate", "Avg Confidence", "Status")
	fmt.Println("------------------------------------------------------------------------")

	for provider, providerResults := range results {
		if len(providerResults) == 0 {
			continue
		}

		var totalDuration time.Duration
		var totalConfidence float64
		successCount := 0

		for _, result := range providerResults {
			totalDuration += result.Duration
			if result.Success {
				successCount++
				totalConfidence += result.Confidence
			}
		}

		avgDuration := totalDuration / time.Duration(len(providerResults))
		successRate := float64(successCount) / float64(len(providerResults)) * 100
		avgConfidence := float64(0)
		if successCount > 0 {
			avgConfidence = totalConfidence / float64(successCount) * 100
		}

		status := "✅ Good"
		if successRate < 100 {
			status = "⚠️  Issues"
		}
		if successRate == 0 {
			status = "❌ Failed"
		}

		fmt.Printf("%-12s %-12s %-12.1f%% %-12.1f%% %-12s\n", 
			provider, 
			formatDuration(avgDuration), 
			successRate, 
			avgConfidence,
			status)
	}

	// Detailed test results
	fmt.Println("\n📋 Detailed Test Results")
	fmt.Println("========================")

	for provider, providerResults := range results {
		if len(providerResults) == 0 {
			continue
		}

		fmt.Printf("\n🔧 %s Provider:\n", provider)
		for _, result := range providerResults {
			status := "✅"
			details := fmt.Sprintf("%.2fs, %.1f%% confidence, %s risk", 
				result.Duration.Seconds(), 
				result.Confidence*100, 
				result.RiskLevel)
			
			if !result.Success {
				status = "❌"
				details = fmt.Sprintf("Failed: %v", result.Error)
			}

			fmt.Printf("   %s %-20s %s\n", status, result.TestCase, details)
		}
	}

	// Recommendations
	fmt.Println("\n💡 Recommendations")
	fmt.Println("==================")

	// Find fastest provider
	var fastestProvider string
	var fastestTime time.Duration = time.Hour
	
	for provider, providerResults := range results {
		if len(providerResults) == 0 {
			continue
		}
		
		var totalDuration time.Duration
		successCount := 0
		
		for _, result := range providerResults {
			if result.Success {
				totalDuration += result.Duration
				successCount++
			}
		}
		
		if successCount > 0 {
			avgDuration := totalDuration / time.Duration(successCount)
			if avgDuration < fastestTime {
				fastestTime = avgDuration
				fastestProvider = provider
			}
		}
	}

	if fastestProvider != "" {
		fmt.Printf("🚀 Fastest Provider: %s (avg %.2fs)\n", fastestProvider, fastestTime.Seconds())
	}

	// Provider-specific recommendations
	for provider := range results {
		switch provider {
		case "ollama":
			fmt.Printf("🦙 Ollama: Best for privacy, offline usage, and cost-effectiveness\n")
		case "openai":
			fmt.Printf("🤖 OpenAI: Best for accuracy and advanced reasoning\n")
		case "claude":
			fmt.Printf("🧠 Claude: Best for nuanced analysis and safety\n")
		case "heuristic":
			fmt.Printf("⚡ Heuristic: Best for speed and reliability as fallback\n")
		}
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}
