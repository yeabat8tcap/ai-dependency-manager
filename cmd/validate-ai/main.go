package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/ai"
	"github.com/8tcapital/ai-dep-manager/internal/ai/types"
	"github.com/8tcapital/ai-dep-manager/internal/logger"
)

func main() {
	fmt.Println("🚀 AI Dependency Manager - AI Provider Runtime Validation")
	fmt.Println("=========================================================")
	
	// Initialize logger
	logger.Init("info", "text")
	
	// Initialize AI system
	fmt.Println("\n📋 Initializing AI Provider System...")
	if err := ai.Initialize(); err != nil {
		log.Fatalf("Failed to initialize AI system: %v", err)
	}
	
	// Get available providers
	providers := ai.GetAvailableProviders()
	fmt.Printf("✅ Available AI Providers: %v\n", providers)
	
	// Check API key availability
	fmt.Println("\n🔑 Checking API Key Configuration...")
	openaiKey := os.Getenv("OPENAI_API_KEY")
	claudeKey := os.Getenv("CLAUDE_API_KEY")
	
	if openaiKey != "" {
		fmt.Printf("✅ OpenAI API Key: Configured (length: %d)\n", len(openaiKey))
	} else {
		fmt.Println("⚠️  OpenAI API Key: Not configured")
	}
	
	if claudeKey != "" {
		fmt.Printf("✅ Claude API Key: Configured (length: %d)\n", len(claudeKey))
	} else {
		fmt.Println("⚠️  Claude API Key: Not configured")
	}
	
	// Test changelog analysis
	fmt.Println("\n📝 Testing Changelog Analysis...")
	testChangelogAnalysis()
	
	// Test version diff analysis
	fmt.Println("\n🔍 Testing Version Diff Analysis...")
	testVersionDiffAnalysis()
	
	// Test compatibility prediction
	fmt.Println("\n🎯 Testing Compatibility Prediction...")
	testCompatibilityPrediction()
	
	// Test update classification
	fmt.Println("\n📊 Testing Update Classification...")
	testUpdateClassification()
	
	// Test fallback mechanism
	fmt.Println("\n🔄 Testing Fallback Mechanism...")
	testFallbackMechanism()
	
	fmt.Println("\n🎉 AI Provider Runtime Validation Complete!")
}

func testChangelogAnalysis() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	request := &types.ChangelogAnalysisRequest{
		PackageName:    "react",
		FromVersion:    "17.0.2",
		ToVersion:      "18.2.0",
		ChangelogText:  "React 18.2.0 introduces concurrent features, automatic batching, and new hooks. Breaking changes include removal of deprecated APIs and changes to event handling.",
		ReleaseNotes:   "This major release includes significant architectural improvements and new concurrent rendering capabilities.",
		PackageManager: "npm",
		Language:       "javascript",
	}
	
	fmt.Printf("  📦 Analyzing changelog for %s: %s -> %s\n", request.PackageName, request.FromVersion, request.ToVersion)
	
	response, err := ai.AnalyzeChangelog(ctx, request)
	if err != nil {
		fmt.Printf("  ❌ Changelog analysis failed: %v\n", err)
		return
	}
	
	fmt.Printf("  ✅ Analysis complete!\n")
	fmt.Printf("     Risk Level: %s (Score: %.2f)\n", response.RiskLevel, response.RiskScore)
	fmt.Printf("     Confidence: %.2f\n", response.Confidence)
	fmt.Printf("     Breaking Changes: %d\n", len(response.BreakingChanges))
	fmt.Printf("     New Features: %d\n", len(response.NewFeatures))
	fmt.Printf("     Security Fixes: %d\n", len(response.SecurityFixes))
	fmt.Printf("     Summary: %s\n", response.Summary)
	
	if len(response.BreakingChanges) > 0 {
		fmt.Printf("     Breaking Change Example: %s\n", response.BreakingChanges[0].Description)
	}
}

func testVersionDiffAnalysis() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	request := &types.VersionDiffAnalysisRequest{
		PackageName:    "express",
		FromVersion:    "4.17.3",
		ToVersion:      "4.18.2",
		DiffText:       "Updated dependencies, security patches, and minor API improvements. Deprecated middleware functions removed.",
		PackageManager: "npm",
		Language:       "javascript",
	}
	
	fmt.Printf("  📦 Analyzing version diff for %s: %s -> %s\n", request.PackageName, request.FromVersion, request.ToVersion)
	
	response, err := ai.AnalyzeVersionDiff(ctx, request)
	if err != nil {
		fmt.Printf("  ❌ Version diff analysis failed: %v\n", err)
		return
	}
	
	fmt.Printf("  ✅ Analysis complete!\n")
	fmt.Printf("     Update Type: %s\n", response.UpdateType)
	fmt.Printf("     Semantic Impact: %s\n", response.SemanticImpact)
	fmt.Printf("     Risk Level: %s (Score: %.2f)\n", response.RiskLevel, response.RiskScore)
	fmt.Printf("     API Changes: %d\n", len(response.APIChanges))
	fmt.Printf("     Behavior Changes: %d\n", len(response.BehaviorChanges))
	fmt.Printf("     Summary: %s\n", response.Summary)
}

func testCompatibilityPrediction() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	request := &types.CompatibilityPredictionRequest{
		PackageName:    "lodash",
		FromVersion:    "4.17.20",
		ToVersion:      "4.17.21",
		ProjectContext: types.ProjectContext{
			Language:     "javascript",
			Framework:    "react",
			Dependencies: []string{"react@18.2.0", "express@4.18.2"},
		},
		PackageManager: "npm",
	}
	
	fmt.Printf("  📦 Predicting compatibility for %s: %s -> %s\n", request.PackageName, request.FromVersion, request.ToVersion)
	
	response, err := ai.PredictCompatibility(ctx, request)
	if err != nil {
		fmt.Printf("  ❌ Compatibility prediction failed: %v\n", err)
		return
	}
	
	fmt.Printf("  ✅ Prediction complete!\n")
	fmt.Printf("     Compatibility Score: %.2f\n", response.CompatibilityScore)
	fmt.Printf("     Risk Level: %s (Score: %.2f)\n", response.RiskLevel, response.RiskScore)
	fmt.Printf("     Potential Issues: %d\n", len(response.PotentialIssues))
	fmt.Printf("     Summary: %s\n", response.Summary)
	
	if len(response.PotentialIssues) > 0 {
		fmt.Printf("     Issue Example: %s\n", response.PotentialIssues[0].Description)
	}
}

func testUpdateClassification() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	request := &types.UpdateClassificationRequest{
		PackageName:    "typescript",
		FromVersion:    "4.9.5",
		ToVersion:      "5.0.2",

		ChangelogText:  "TypeScript 5.0 introduces decorators, const type parameters, and breaking changes to module resolution.",
		ProjectContext: types.ProjectContext{
			Language:     "typescript",
			Framework:    "angular",
			Dependencies: []string{"@angular/core@15.2.0", "@types/node@18.15.0"},
		},
		PackageManager: "npm",
	}
	
	fmt.Printf("  📦 Classifying update for %s: %s -> %s\n", request.PackageName, request.FromVersion, request.ToVersion)
	
	response, err := ai.ClassifyUpdate(ctx, request)
	if err != nil {
		fmt.Printf("  ❌ Update classification failed: %v\n", err)
		return
	}
	
	fmt.Printf("  ✅ Classification complete!\n")
	fmt.Printf("     Update Type: %s\n", response.UpdateType)
	fmt.Printf("     Priority: %s\n", response.Priority)
	fmt.Printf("     Urgency: %s\n", response.Urgency)
	fmt.Printf("     Categories: %d\n", len(response.Categories))
	fmt.Printf("     Summary: %s\n", response.Summary)
	
	if len(response.Categories) > 0 {
		fmt.Printf("     Category Example: %s (Weight: %.2f)\n", response.Categories[0].Name, response.Categories[0].Weight)
	}
}

func testFallbackMechanism() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Test with a simple request that should work with heuristic fallback
	request := &types.ChangelogAnalysisRequest{
		PackageName:    "test-package",
		FromVersion:    "1.0.0",
		ToVersion:      "2.0.0",
		ChangelogText:  "BREAKING CHANGE: Removed deprecated API. Added new security features.",
		ReleaseNotes:   "Major version with breaking changes and security improvements.",
		PackageManager: "npm",
		Language:       "javascript",
	}
	
	fmt.Printf("  📦 Testing fallback mechanism with test package\n")
	
	// This should work even if LLM providers are not available
	response, err := ai.AnalyzeChangelog(ctx, request)
	if err != nil {
		fmt.Printf("  ❌ Fallback mechanism failed: %v\n", err)
		return
	}
	
	fmt.Printf("  ✅ Fallback mechanism working!\n")
	fmt.Printf("     Risk Level: %s (Score: %.2f)\n", response.RiskLevel, response.RiskScore)
	fmt.Printf("     Breaking Changes Detected: %d\n", len(response.BreakingChanges))
	fmt.Printf("     Summary: %s\n", response.Summary)
	
	// Test provider availability
	fmt.Println("\n  🔍 Checking provider availability...")
	providers := ai.GetAvailableProviders()
	for _, provider := range providers {
		fmt.Printf("     Provider '%s': Available\n", provider)
	}
}
