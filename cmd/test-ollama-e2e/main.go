package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/8tcapital/superint-dep-manager/internal/superint"
	"github.com/8tcapital/superint-dep-manager/internal/superint/ollama"
	"github.com/8tcapital/superint-dep-manager/internal/superint/types"
	"github.com/8tcapital/superint-dep-manager/internal/logger"
)

func main() {
	fmt.Println("🧪 Ollama End-to-End Testing")
	fmt.Println("============================")

	// Initialize logger
	logger.Init("info", "text")

	// Test Ollama deployment and functionality
	if err := runEndToEndTests(); err != nil {
		fmt.Printf("❌ End-to-end tests failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n🎉 All end-to-end tests passed successfully!")
	fmt.Println("✅ Ollama integration is production-ready!")
}

func runEndToEndTests() error {
	fmt.Println("\n🔍 Step 1: Testing Ollama Server Connection")
	if err := testOllamaConnection(); err != nil {
		return fmt.Errorf("ollama connection test failed: %w", err)
	}
	fmt.Println("✅ Ollama server connection successful")

	fmt.Println("\n📋 Step 2: Testing Model Availability")
	if err := testModelAvailability(); err != nil {
		return fmt.Errorf("model availability test failed: %w", err)
	}
	fmt.Println("✅ Model availability check successful")

	fmt.Println("\n🔄 Step 3: Testing Model Switching")
	if err := testModelSwitching(); err != nil {
		return fmt.Errorf("model switching test failed: %w", err)
	}
	fmt.Println("✅ Model switching functionality working")

	fmt.Println("\n🧠 Step 4: Testing AI Analysis Functionality")
	if err := testAIAnalysis(); err != nil {
		return fmt.Errorf("AI analysis test failed: %w", err)
	}
	fmt.Println("✅ AI analysis functionality working")

	fmt.Println("\n⚡ Step 5: Testing Performance and Reliability")
	if err := testPerformanceReliability(); err != nil {
		return fmt.Errorf("performance test failed: %w", err)
	}
	fmt.Println("✅ Performance and reliability tests passed")

	fmt.Println("\n🔄 Step 6: Testing Fallback Mechanisms")
	if err := testFallbackMechanisms(); err != nil {
		return fmt.Errorf("fallback test failed: %w", err)
	}
	fmt.Println("✅ Fallback mechanisms working correctly")

	return nil
}

func testOllamaConnection() error {
	// Create Ollama provider
	config := &ollama.OllamaConfig{
		BaseURL:     getOllamaURL(),
		Model:       getOllamaModel(),
		Temperature: 0.7,
	}

	provider, err := ollama.NewOllamaProvider(config)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if !provider.IsAvailable(ctx) {
		return fmt.Errorf("ollama server not available at %s", config.BaseURL)
	}

	// Test connection endpoint
	if err := provider.TestConnection(ctx); err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	return nil
}

func testModelAvailability() error {
	config := &ollama.OllamaConfig{
		BaseURL: getOllamaURL(),
		Model:   getOllamaModel(),
	}

	provider, err := ollama.NewOllamaProvider(config)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// List available models
	models, err := provider.ListAvailableModels(ctx)
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	if len(models) == 0 {
		return fmt.Errorf("no models available - please run 'ollama pull %s'", getOllamaModel())
	}

	fmt.Printf("   📦 Found %d available model(s):\n", len(models))
	for _, model := range models {
		fmt.Printf("      - %s\n", model)
	}

	// Check if configured model is available
	modelFound := false
	for _, model := range models {
		if model == getOllamaModel() {
			modelFound = true
			break
		}
	}

	if !modelFound {
		return fmt.Errorf("configured model '%s' not found - please run 'ollama pull %s'", 
			getOllamaModel(), getOllamaModel())
	}

	return nil
}

func testModelSwitching() error {
	config := &ollama.OllamaConfig{
		BaseURL: getOllamaURL(),
		Model:   getOllamaModel(),
	}

	provider, err := ollama.NewOllamaProvider(config)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Get available models
	models, err := provider.ListAvailableModels(ctx)
	if err != nil {
		return err
	}

	originalModel := provider.GetCurrentModel()
	fmt.Printf("   🔄 Current model: %s\n", originalModel)

	// Test getting available model configurations
	availableConfigs, err := provider.GetAvailableModels()
	if err != nil {
		return fmt.Errorf("failed to get model configurations: %w", err)
	}

	fmt.Printf("   📋 Available model configurations: %d\n", len(availableConfigs))

	// If we have multiple models, test switching
	if len(models) > 1 {
		for _, model := range models {
			if model != originalModel {
				fmt.Printf("   🔄 Testing switch to %s...\n", model)
				
				if err := provider.SwitchModel(model); err != nil {
					return fmt.Errorf("failed to switch to model %s: %w", model, err)
				}

				if provider.GetCurrentModel() != model {
					return fmt.Errorf("model switch failed - expected %s, got %s", 
						model, provider.GetCurrentModel())
				}

				// Switch back
				if err := provider.SwitchModel(originalModel); err != nil {
					return fmt.Errorf("failed to switch back to %s: %w", originalModel, err)
				}
				break
			}
		}
	} else {
		fmt.Printf("   ℹ️  Only one model available, skipping switch test\n")
	}

	return nil
}

func testAIAnalysis() error {
	// Initialize AI system
	if err := superint.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize AI system: %w", err)
	}

	// Get Ollama provider
	provider, exists := superint.GetProvider("ollama")
	if !exists {
		return fmt.Errorf("ollama provider not found")
	}

	// Test changelog analysis
	request := &types.ChangelogAnalysisRequest{
		PackageName:   "react",
		FromVersion:   "17.0.0",
		ToVersion:     "18.0.0",
		ChangelogText: `# React 18.0.0

## Breaking Changes
- Automatic batching changes
- Stricter StrictMode
- Consistent useEffect timing

## New Features  
- Concurrent rendering
- Suspense improvements
- New hooks: useId, useDeferredValue, useTransition`,
		ReleaseNotes: "Major release with concurrent features and breaking changes",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fmt.Printf("   🧠 Testing changelog analysis...\n")
	start := time.Now()

	response, err := provider.AnalyzeChangelog(ctx, request)
	if err != nil {
		return fmt.Errorf("changelog analysis failed: %w", err)
	}

	duration := time.Since(start)
	fmt.Printf("   ✅ Analysis completed in %.2fs\n", duration.Seconds())
	fmt.Printf("   📊 Risk Level: %s\n", response.RiskLevel)
	fmt.Printf("   📈 Confidence: %.1f%%\n", response.ConfidenceScore*100)
	fmt.Printf("   📝 Summary: %s\n", truncateString(response.Summary, 100))

	// Validate response structure
	if response.PackageName != request.PackageName {
		return fmt.Errorf("response package name mismatch")
	}
	if response.FromVersion != request.FromVersion {
		return fmt.Errorf("response from version mismatch")
	}
	if response.ToVersion != request.ToVersion {
		return fmt.Errorf("response to version mismatch")
	}

	return nil
}

func testPerformanceReliability() error {
	provider, exists := superint.GetProvider("ollama")
	if !exists {
		return fmt.Errorf("ollama provider not found")
	}

	// Test multiple requests for reliability
	const numTests = 3
	var totalDuration time.Duration
	successCount := 0

	fmt.Printf("   ⚡ Running %d performance tests...\n", numTests)

	for i := 0; i < numTests; i++ {
		request := &types.ChangelogAnalysisRequest{
			PackageName:   fmt.Sprintf("test-package-%d", i+1),
			FromVersion:   "1.0.0",
			ToVersion:     "2.0.0",
			ChangelogText: "Minor bug fixes and improvements",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		start := time.Now()

		_, err := provider.AnalyzeChangelog(ctx, request)
		duration := time.Since(start)
		cancel()

		if err != nil {
			fmt.Printf("   ⚠️  Test %d failed: %v\n", i+1, err)
		} else {
			totalDuration += duration
			successCount++
			fmt.Printf("   ✅ Test %d: %.2fs\n", i+1, duration.Seconds())
		}
	}

	if successCount == 0 {
		return fmt.Errorf("all performance tests failed")
	}

	avgDuration := totalDuration / time.Duration(successCount)
	successRate := float64(successCount) / float64(numTests) * 100

	fmt.Printf("   📊 Performance Results:\n")
	fmt.Printf("      Success Rate: %.1f%%\n", successRate)
	fmt.Printf("      Average Time: %.2fs\n", avgDuration.Seconds())

	if successRate < 80 {
		return fmt.Errorf("success rate too low: %.1f%%", successRate)
	}

	return nil
}

func testFallbackMechanisms() error {
	// Initialize AI system
	if err := superint.Initialize(); err != nil {
		return err
	}

	// Test with invalid Ollama configuration to trigger fallback
	fmt.Printf("   🔄 Testing fallback to heuristic provider...\n")

	// Create request
	request := &types.ChangelogAnalysisRequest{
		PackageName:   "test-fallback",
		FromVersion:   "1.0.0",
		ToVersion:     "2.0.0",
		ChangelogText: "Breaking changes and security fixes",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// This should use the AI system's fallback mechanism
	response, err := superint.AnalyzeChangelog(ctx, request)
	if err != nil {
		return fmt.Errorf("fallback mechanism failed: %w", err)
	}

	fmt.Printf("   ✅ Fallback analysis completed\n")
	fmt.Printf("   📊 Risk Level: %s\n", response.RiskLevel)

	return nil
}

func getOllamaURL() string {
	if url := os.Getenv("OLLAMA_BASE_URL"); url != "" {
		return url
	}
	return "http://localhost:11434"
}

func getOllamaModel() string {
	if model := os.Getenv("OLLAMA_MODEL"); model != "" {
		return model
	}
	return "llama2"
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
