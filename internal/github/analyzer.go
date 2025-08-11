package github

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/ai"
	"github.com/8tcapital/ai-dep-manager/internal/logger"
	"github.com/8tcapital/ai-dep-manager/internal/models"
)

// AnalysisService handles code analysis and breaking change detection
type AnalysisService struct {
	aiManager    interface{}
	client       *Client
	repositories *RepositoriesService
}

// NewAnalysisService creates a new analysis service
func NewAnalysisService(client *Client, aiManager interface{}) *AnalysisService {
	return &AnalysisService{
		aiManager:    aiManager,
		client:       client,
		repositories: client.Repositories,
	}
}

// AnalysisRecommendation represents a recommendation for handling a dependency update
type AnalysisRecommendation struct {
	Action      string `json:"action"`
	Description string `json:"description"`
	Priority    string `json:"priority"` // "required", "recommended", "optional"
	Effort      string `json:"effort"`   // "low", "medium", "high"
}

// ProjectAnalysis represents analysis of a project's dependencies
type ProjectAnalysis struct {
	Repository      string                    `json:"repository"`
	ProjectType     string                    `json:"project_type"`     // "nodejs", "python", "java", etc.
	PackageManager  string                    `json:"package_manager"`  // "npm", "pip", "maven", etc.
	Dependencies    []*BreakingChangeAnalysis `json:"dependencies"`
	OverallRisk     string          `json:"overall_risk"`
	TotalChanges    int                       `json:"total_changes"`
	BreakingChanges int                       `json:"breaking_changes"`
	AnalyzedAt      time.Time                 `json:"analyzed_at"`
}

// AnalyzeDependencyUpdate analyzes a dependency update for breaking changes
func (a *AnalysisService) AnalyzeDependencyUpdate(ctx context.Context, dependency *DependencyUpdate) (*BreakingChangeAnalysis, error) {
	logger.Info("Analyzing dependency update: %s %s -> %s", 
		dependency.Name, dependency.CurrentVersion, dependency.LatestVersion)
	
	analysis := &BreakingChangeAnalysis{
		Dependency:      dependency,
		BreakingChanges: []*BreakingChange{},
		Recommendations: []*Recommendation{},
		PatchSuggestions: []*PatchSuggestion{},
		AnalyzedAt:      time.Now(),
	}
	
	// Try AI analysis first
	if a.aiManager != nil {
		aiAnalysis, err := a.performAIAnalysis(ctx, dependency)
		if err != nil {
			logger.Warn("AI analysis failed for %s: %v", dependency.Name, err)
		} else {
			analysis = aiAnalysis
			analysis.AnalysisSource = "ai"
			logger.Info("AI analysis completed for %s with confidence %.2f", 
				dependency.Name, analysis.Confidence)
		}
	}
	
	// Fallback to heuristic analysis if AI failed or confidence is low
	if analysis.AnalysisSource == "" || analysis.Confidence < 0.7 {
		heuristicAnalysis := a.performHeuristicAnalysis(dependency)
		if analysis.AnalysisSource == "" {
			analysis = heuristicAnalysis
			analysis.AnalysisSource = "heuristic"
		} else {
			// Merge AI and heuristic results
			analysis = a.mergeAnalysis(analysis, heuristicAnalysis)
			analysis.AnalysisSource = "ai+heuristic"
		}
	}
	
	// Enhance with changelog analysis if available
	if changelogAnalysis, err := a.performChangelogAnalysis(ctx, dependency); err == nil {
		analysis = a.mergeAnalysis(analysis, changelogAnalysis)
		if strings.Contains(analysis.AnalysisSource, "changelog") == false {
			analysis.AnalysisSource += "+changelog"
		}
	}
	
	logger.Info("Analysis completed for %s: %d breaking changes, risk level %s", 
		dependency.Name, len(analysis.BreakingChanges), analysis.RiskLevel)
	
	return analysis, nil
}

// performAIAnalysis performs AI-powered breaking change analysis
func (a *AnalysisService) performAIAnalysis(ctx context.Context, dependency *DependencyUpdate) (*BreakingChangeAnalysis, error) {
	// Create analysis request for AI
	request := &models.AnalysisRequest{
		PackageName:    dependency.Name,
		CurrentVersion: dependency.CurrentVersion,
		TargetVersion:  dependency.LatestVersion,
		PackageManager: "npm", // TODO: Detect from context
		ChangelogURL:   dependency.ChangelogURL,
	}
	
	// Get AI analysis
	result, err := a.aiManager.AnalyzeDependencyUpdate(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("AI analysis failed: %w", err)
	}
	
	// Convert AI result to our format
	analysis := &BreakingChangeAnalysis{
		Dependency:         dependency,
		HasBreakingChanges: len(result.BreakingChanges) > 0,
		RiskLevel:          result.RiskLevel,
		Confidence:         result.Confidence,
		AnalyzedAt:         time.Now(),
	}
	
	// Convert breaking changes
	for _, bc := range result.BreakingChanges {
		analysis.BreakingChanges = append(analysis.BreakingChanges, &BreakingChange{
			Type:             bc.Type,
			Description:      bc.Description,
			AffectedAPIs:     bc.AffectedAPIs,
			Severity:         bc.Severity,
			MigrationPath:    bc.MigrationPath,
			ExampleBefore:    bc.ExampleBefore,
			ExampleAfter:     bc.ExampleAfter,
			DocumentationURL: bc.DocumentationURL,
		})
	}
	
	// Convert recommendations
	for _, rec := range result.Recommendations {
		analysis.Recommendations = append(analysis.Recommendations, &AnalysisRecommendation{
			Action:      rec.Action,
			Description: rec.Description,
			Priority:    rec.Priority,
			Effort:      rec.Effort,
		})
	}
	
	// Generate patch suggestions based on AI analysis
	patchSuggestions := a.generatePatchSuggestions(dependency, analysis.BreakingChanges)
	analysis.PatchSuggestions = patchSuggestions
	
	return analysis, nil
}

// performHeuristicAnalysis performs rule-based breaking change analysis
func (a *AnalysisService) performHeuristicAnalysis(dependency *DependencyUpdate) *BreakingChangeAnalysis {
	analysis := &BreakingChangeAnalysis{
		Dependency:         dependency,
		BreakingChanges:    []*BreakingChange{},
		Recommendations:    []*Recommendation{},
		PatchSuggestions:   []*PatchSuggestion{},
		AnalyzedAt:         time.Now(),
		Confidence:         0.6, // Moderate confidence for heuristics
	}
	
	// Analyze version change pattern
	versionAnalysis := a.analyzeVersionChange(dependency.CurrentVersion, dependency.LatestVersion)
	
	// Check for major version changes
	if versionAnalysis.IsMajorChange {
		analysis.HasBreakingChanges = true
		analysis.RiskLevel = "high"
		
		analysis.BreakingChanges = append(analysis.BreakingChanges, &BreakingChange{
			Type:        "major_version_change",
			Description: fmt.Sprintf("Major version change from %s to %s likely contains breaking changes", 
				dependency.CurrentVersion, dependency.LatestVersion),
			Severity:    "high",
			MigrationPath: "Review changelog and migration guide for breaking changes",
		})
		
		analysis.Recommendations = append(analysis.Recommendations, &AnalysisRecommendation{
			Action:      "review_changelog",
			Description: "Review the changelog and migration guide before updating",
			Priority:    "required",
			Effort:      "medium",
		})
	} else if versionAnalysis.IsMinorChange {
		analysis.RiskLevel = "medium"
		analysis.Recommendations = append(analysis.Recommendations, &AnalysisRecommendation{
			Action:      "test_thoroughly",
			Description: "Test thoroughly as minor versions may introduce new features",
			Priority:    "recommended",
			Effort:      "low",
		})
	} else {
		analysis.RiskLevel = "low"
	}
	
	// Check for security updates
	if dependency.SecurityFix {
		analysis.Recommendations = append(analysis.Recommendations, &AnalysisRecommendation{
			Action:      "update_immediately",
			Description: "Security fix - update as soon as possible",
			Priority:    "required",
			Effort:      "low",
		})
	}
	
	// Package-specific heuristics
	packageHeuristics := a.getPackageSpecificHeuristics(dependency.Name)
	if packageHeuristics != nil {
		analysis.BreakingChanges = append(analysis.BreakingChanges, packageHeuristics.BreakingChanges...)
		analysis.Recommendations = append(analysis.Recommendations, packageHeuristics.Recommendations...)
	}
	
	return analysis
}

// performChangelogAnalysis analyzes changelog for breaking changes
func (a *AnalysisService) performChangelogAnalysis(ctx context.Context, dependency *DependencyUpdate) (*BreakingChangeAnalysis, error) {
	if dependency.ChangelogURL == "" {
		return nil, fmt.Errorf("no changelog URL available")
	}
	
	// This would fetch and parse the changelog
	// For now, return a placeholder analysis
	analysis := &BreakingChangeAnalysis{
		Dependency:      dependency,
		BreakingChanges: []*BreakingChange{},
		Recommendations: []*Recommendation{},
		PatchSuggestions: []*PatchSuggestion{},
		AnalyzedAt:      time.Now(),
		Confidence:      0.8,
	}
	
	// TODO: Implement actual changelog parsing
	logger.Debug("Changelog analysis not yet implemented for %s", dependency.Name)
	
	return analysis, nil
}

// VersionAnalysis represents the result of version analysis
type VersionAnalysis struct {
	IsMajorChange bool
	IsMinorChange bool
	IsPatchChange bool
	IsPrerelease  bool
}

// analyzeVersionChange analyzes the type of version change
func (a *AnalysisService) analyzeVersionChange(current, target string) *VersionAnalysis {
	// Simple semantic version parsing
	currentParts := strings.Split(strings.TrimPrefix(current, "v"), ".")
	targetParts := strings.Split(strings.TrimPrefix(target, "v"), ".")
	
	analysis := &VersionAnalysis{}
	
	if len(currentParts) >= 1 && len(targetParts) >= 1 {
		if currentParts[0] != targetParts[0] {
			analysis.IsMajorChange = true
		} else if len(currentParts) >= 2 && len(targetParts) >= 2 && currentParts[1] != targetParts[1] {
			analysis.IsMinorChange = true
		} else {
			analysis.IsPatchChange = true
		}
	}
	
	// Check for prerelease versions
	if strings.Contains(target, "-") || strings.Contains(target, "alpha") || 
	   strings.Contains(target, "beta") || strings.Contains(target, "rc") {
		analysis.IsPrerelease = true
	}
	
	return analysis
}

// PackageHeuristics represents package-specific breaking change patterns
type PackageHeuristics struct {
	BreakingChanges []*BreakingChange
	Recommendations []*AnalysisRecommendation
}

// getPackageSpecificHeuristics returns package-specific heuristics
func (a *AnalysisService) getPackageSpecificHeuristics(packageName string) *PackageHeuristics {
	// Common packages with known breaking change patterns
	switch packageName {
	case "express":
		return &PackageHeuristics{
			BreakingChanges: []*BreakingChange{
				{
					Type:        "middleware_changes",
					Description: "Express major versions often change middleware behavior",
					Severity:    "medium",
					MigrationPath: "Review middleware configuration and error handling",
				},
			},
			Recommendations: []*AnalysisRecommendation{
				{
					Action:      "test_middleware",
					Description: "Test all middleware and route handlers",
					Priority:    "recommended",
					Effort:      "medium",
				},
			},
		}
	case "react":
		return &PackageHeuristics{
			BreakingChanges: []*BreakingChange{
				{
					Type:        "lifecycle_changes",
					Description: "React major versions often change component lifecycle methods",
					Severity:    "high",
					MigrationPath: "Update lifecycle methods and hooks usage",
				},
			},
			Recommendations: []*AnalysisRecommendation{
				{
					Action:      "update_components",
					Description: "Review and update component lifecycle methods",
					Priority:    "required",
					Effort:      "high",
				},
			},
		}
	case "lodash":
		return &PackageHeuristics{
			BreakingChanges: []*BreakingChange{
				{
					Type:        "method_removal",
					Description: "Lodash major versions may remove or rename utility methods",
					Severity:    "medium",
					MigrationPath: "Check for removed or renamed methods",
				},
			},
		}
	}
	
	return nil
}

// generatePatchSuggestions generates patch suggestions based on breaking changes
func (a *AnalysisService) generatePatchSuggestions(dependency *DependencyUpdate, breakingChanges []*BreakingChange) []*PatchSuggestion {
	var suggestions []*PatchSuggestion
	
	for _, bc := range breakingChanges {
		if bc.ExampleBefore != "" && bc.ExampleAfter != "" {
			suggestion := &PatchSuggestion{
				OldCode:     bc.ExampleBefore,
				NewCode:     bc.ExampleAfter,
				Description: fmt.Sprintf("Update for %s: %s", bc.Type, bc.Description),
				Confidence:  0.7,
			}
			suggestions = append(suggestions, suggestion)
		}
	}
	
	return suggestions
}

// mergeAnalysis merges two analysis results
func (a *AnalysisService) mergeAnalysis(primary, secondary *BreakingChangeAnalysis) *BreakingChangeAnalysis {
	merged := *primary // Copy primary analysis
	
	// Merge breaking changes (avoid duplicates)
	existingChanges := make(map[string]bool)
	for _, bc := range merged.BreakingChanges {
		existingChanges[bc.Type+":"+bc.Description] = true
	}
	
	for _, bc := range secondary.BreakingChanges {
		key := bc.Type + ":" + bc.Description
		if !existingChanges[key] {
			merged.BreakingChanges = append(merged.BreakingChanges, bc)
		}
	}
	
	// Merge recommendations
	existingRecs := make(map[string]bool)
	for _, rec := range merged.Recommendations {
		existingRecs[rec.Action+":"+rec.Description] = true
	}
	
	for _, rec := range secondary.Recommendations {
		key := rec.Action + ":" + rec.Description
		if !existingRecs[key] {
			merged.Recommendations = append(merged.Recommendations, rec)
		}
	}
	
	// Merge patch suggestions
	merged.PatchSuggestions = append(merged.PatchSuggestions, secondary.PatchSuggestions...)
	
	// Update flags
	merged.HasBreakingChanges = len(merged.BreakingChanges) > 0
	
	// Use higher risk level
	if secondary.RiskLevel > merged.RiskLevel {
		merged.RiskLevel = secondary.RiskLevel
	}
	
	// Average confidence scores
	merged.Confidence = (merged.Confidence + secondary.Confidence) / 2
	
	return &merged
}

// AnalyzeProject analyzes all dependencies in a project
func (a *AnalysisService) AnalyzeProject(ctx context.Context, owner, repo string, dependencies []*DependencyUpdate) (*ProjectAnalysis, error) {
	logger.Info("Analyzing project %s/%s with %d dependencies", owner, repo, len(dependencies))
	
	analysis := &ProjectAnalysis{
		Repository:      fmt.Sprintf("%s/%s", owner, repo),
		Dependencies:    []*BreakingChangeAnalysis{},
		AnalyzedAt:      time.Now(),
	}
	
	// Detect project type and package manager
	projectInfo, err := a.detectProjectType(ctx, owner, repo)
	if err != nil {
		logger.Warn("Failed to detect project type for %s/%s: %v", owner, repo, err)
	} else {
		analysis.ProjectType = projectInfo.Type
		analysis.PackageManager = projectInfo.PackageManager
	}
	
	// Analyze each dependency
	var totalBreakingChanges int
	var maxRiskLevel models.RiskLevel
	
	for _, dep := range dependencies {
		depAnalysis, err := a.AnalyzeDependencyUpdate(ctx, dep)
		if err != nil {
			logger.Error("Failed to analyze dependency %s: %v", dep.Name, err)
			continue
		}
		
		analysis.Dependencies = append(analysis.Dependencies, depAnalysis)
		
		if depAnalysis.HasBreakingChanges {
			totalBreakingChanges += len(depAnalysis.BreakingChanges)
		}
		
		if depAnalysis.RiskLevel > maxRiskLevel {
			maxRiskLevel = depAnalysis.RiskLevel
		}
	}
	
	analysis.TotalChanges = len(dependencies)
	analysis.BreakingChanges = totalBreakingChanges
	analysis.OverallRisk = maxRiskLevel
	
	logger.Info("Project analysis completed: %d total changes, %d breaking changes, risk level %s", 
		analysis.TotalChanges, analysis.BreakingChanges, analysis.OverallRisk)
	
	return analysis, nil
}

// ProjectInfo represents information about a project
type ProjectInfo struct {
	Type           string // "nodejs", "python", "java", etc.
	PackageManager string // "npm", "pip", "maven", etc.
}

// detectProjectType detects the project type and package manager
func (a *AnalysisService) detectProjectType(ctx context.Context, owner, repo string) (*ProjectInfo, error) {
	// Check for common project files
	files := []string{"package.json", "requirements.txt", "pom.xml", "build.gradle", "Cargo.toml", "go.mod"}
	
	for _, file := range files {
		_, err := a.repositories.GetContents(ctx, owner, repo, file, "")
		if err == nil {
			// File exists, determine project type
			switch file {
			case "package.json":
				return &ProjectInfo{Type: "nodejs", PackageManager: "npm"}, nil
			case "requirements.txt":
				return &ProjectInfo{Type: "python", PackageManager: "pip"}, nil
			case "pom.xml":
				return &ProjectInfo{Type: "java", PackageManager: "maven"}, nil
			case "build.gradle":
				return &ProjectInfo{Type: "java", PackageManager: "gradle"}, nil
			case "Cargo.toml":
				return &ProjectInfo{Type: "rust", PackageManager: "cargo"}, nil
			case "go.mod":
				return &ProjectInfo{Type: "go", PackageManager: "go"}, nil
			}
		}
	}
	
	return &ProjectInfo{Type: "unknown", PackageManager: "unknown"}, nil
}
