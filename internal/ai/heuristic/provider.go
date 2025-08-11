package heuristic

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/ai/types"
	"github.com/8tcapital/ai-dep-manager/internal/logger"
)

// HeuristicProvider implements AI analysis using keyword-based heuristics
type HeuristicProvider struct {
	name    string
	version string
}

// NewHeuristicProvider creates a new heuristic-based AI provider
func NewHeuristicProvider() *HeuristicProvider {
	return &HeuristicProvider{
		name:    "heuristic",
		version: "1.0.0",
	}
}

// GetName returns the name of the AI provider
func (h *HeuristicProvider) GetName() string {
	return h.name
}

// GetVersion returns the version of the AI provider
func (h *HeuristicProvider) GetVersion() string {
	return h.version
}

// IsAvailable checks if the provider is available (always true for heuristic)
func (h *HeuristicProvider) IsAvailable(ctx context.Context) bool {
	return true
}

// AnalyzeChangelog analyzes changelog text using keyword-based heuristics
func (h *HeuristicProvider) AnalyzeChangelog(ctx context.Context, request *types.ChangelogAnalysisRequest) (*types.ChangelogAnalysisResponse, error) {
	logger.Debug("Analyzing changelog for %s: %s -> %s", request.PackageName, request.FromVersion, request.ToVersion)
	
	text := strings.ToLower(request.ChangelogText + " " + request.ReleaseNotes)
	
	response := &types.ChangelogAnalysisResponse{
		PackageName: request.PackageName,
		FromVersion: request.FromVersion,
		ToVersion:   request.ToVersion,
		AnalyzedAt:  time.Now(),
	}
	
	// Analyze breaking changes
	breakingChanges := h.detectBreakingChanges(text)
	response.BreakingChanges = breakingChanges
	response.HasBreakingChange = len(breakingChanges) > 0
	
	// Analyze security fixes
	response.SecurityFixes = h.detectSecurityFixes(text)
	
	// Analyze new features
	response.NewFeatures = h.detectNewFeatures(text)
	
	// Analyze bug fixes
	response.BugFixes = h.detectBugFixes(text)
	
	// Analyze deprecations
	response.Deprecations = h.detectDeprecations(text)
	
	// Determine risk level
	response.RiskLevel = h.calculateRiskLevel(response)
	
	// Set confidence (heuristic-based is lower confidence)
	response.Confidence = h.calculateConfidence(text, response)
	
	// Generate summary
	response.Summary = h.generateSummary(response)
	
	// Generate recommendations
	response.Recommendations = h.generateRecommendations(response)
	
	logger.Debug("Changelog analysis complete for %s: risk=%s, confidence=%.2f", 
		request.PackageName, response.RiskLevel, response.Confidence)
	
	return response, nil
}

// AnalyzeVersionDiff analyzes version differences using semantic versioning heuristics
func (h *HeuristicProvider) AnalyzeVersionDiff(ctx context.Context, request *types.VersionDiffAnalysisRequest) (*types.VersionDiffAnalysisResponse, error) {
	logger.Debug("Analyzing version diff for %s: %s -> %s", request.PackageName, request.FromVersion, request.ToVersion)
	
	response := &types.VersionDiffAnalysisResponse{
		PackageName: request.PackageName,
		FromVersion: request.FromVersion,
		ToVersion:   request.ToVersion,
		ProcessedAt: time.Now(),
	}
	
	// Determine update type using semantic versioning
	updateType := h.determineUpdateType(request.FromVersion, request.ToVersion)
	response.UpdateType = updateType
	
	// Determine semantic impact
	response.SemanticImpact = h.determineSemanticImpact(updateType)
	
	// Calculate risk level based on update type
	response.RiskLevel = h.calculateVersionRiskLevel(updateType)
	
	// Set confidence
	response.Confidence = 0.8 // High confidence for version-based analysis
	
	// Generate summary
	response.Summary = fmt.Sprintf("%s update from %s to %s", 
		strings.Title(updateType), request.FromVersion, request.ToVersion)
	
	return response, nil
}

// PredictCompatibility predicts compatibility using basic heuristics
func (h *HeuristicProvider) PredictCompatibility(ctx context.Context, request *types.CompatibilityPredictionRequest) (*types.CompatibilityPredictionResponse, error) {
	logger.Debug("Predicting compatibility for %s: %s -> %s", request.PackageName, request.FromVersion, request.ToVersion)
	
	response := &types.CompatibilityResponse{
		PackageName: request.PackageName,
		FromVersion: request.FromVersion,
		ToVersion:   request.ToVersion,
		ProcessedAt: time.Now(),
	}
	
	// Determine update type
	updateType := h.determineUpdateType(request.FromVersion, request.ToVersion)
	
	// Calculate compatibility score based on update type
	response.CompatibilityScore = h.calculateCompatibilityScore(updateType)
	
	// Determine risk level
	response.RiskLevel = h.calculateVersionRiskLevel(updateType)
	
	// Generate potential issues
	response.PotentialIssues = h.generatePotentialIssues(updateType)
	
	// Generate migration steps
	response.MigrationSteps = h.generateMigrationSteps(updateType)
	
	// Generate testing recommendations
	response.TestingRecommendations = h.generateTestingRecommendations(updateType)
	
	// Set confidence
	response.Confidence = 0.6 // Medium confidence for compatibility prediction
	
	// Generate summary
	response.Summary = h.generateCompatibilitySummary(updateType, response.CompatibilityScore)
	
	return response, nil
}

// ClassifyUpdate classifies updates using heuristics
func (h *HeuristicProvider) ClassifyUpdate(ctx context.Context, request *types.UpdateClassificationRequest) (*types.UpdateClassificationResponse, error) {
	logger.Debug("Classifying update for %s: %s -> %s", request.PackageName, request.FromVersion, request.ToVersion)
	
	response := &types.UpdateClassificationResponse{
		PackageName: request.PackageName,
		FromVersion: request.FromVersion,
		ToVersion:   request.ToVersion,
		ProcessedAt: time.Now(),
	}
	
	// Determine update type
	updateType := h.determineUpdateType(request.FromVersion, request.ToVersion)
	response.UpdateType = updateType
	
	// Classify categories
	response.Categories = h.classifyUpdateCategories(request.ChangelogText, request.ReleaseNotes)
	
	// Determine priority
	response.Priority = h.determinePriority(updateType, response.Categories)
	
	// Determine risk level
	response.RiskLevel = h.calculateVersionRiskLevel(updateType)
	
	// Determine urgency
	response.Urgency = h.determineUrgency(response.Categories)
	
	// Set confidence
	response.Confidence = 0.7 // Good confidence for classification
	
	// Generate summary
	response.Summary = h.generateClassificationSummary(response)
	
	return response, nil
}

// Helper methods for heuristic analysis

func (h *HeuristicProvider) detectBreakingChanges(text string) []types.BreakingChange {
	var breakingChanges []types.BreakingChange
	
	// Breaking change keywords and patterns
	breakingPatterns := map[string]string{
		"breaking change":     "explicit_breaking_change",
		"breaking":           "potential_breaking_change",
		"removed":            "api_removal",
		"deprecated":         "deprecation",
		"incompatible":       "incompatibility",
		"major change":       "major_change",
		"api change":         "api_change",
		"signature change":   "signature_change",
		"behavior change":    "behavior_change",
		"no longer":          "removal",
		"not supported":      "support_removal",
	}
	
	for pattern, changeType := range breakingPatterns {
		if strings.Contains(text, pattern) {
			confidence := 0.8
			if pattern == "breaking change" {
				confidence = 0.95
			} else if pattern == "breaking" {
				confidence = 0.7
			}
			
			breakingChanges = append(breakingChanges, types.BreakingChange{
				Type:        changeType,
				Description: fmt.Sprintf("Detected '%s' in changelog", pattern),
				Impact:      h.determineImpact(pattern),
				Confidence:  confidence,
				Mitigation:  h.generateMitigation(changeType),
			})
		}
	}
	
	return breakingChanges
}

func (h *HeuristicProvider) detectSecurityFixes(text string) []types.SecurityFix {
	var securityFixes []types.SecurityFix
	
	securityPatterns := map[string]string{
		"security":           "medium",
		"vulnerability":      "high",
		"cve":               "high",
		"security fix":      "medium",
		"security update":   "medium",
		"security patch":    "medium",
		"exploit":           "critical",
		"injection":         "high",
		"xss":               "high",
		"csrf":              "medium",
	}
	
	for pattern, severity := range securityPatterns {
		if strings.Contains(text, pattern) {
			confidence := 0.8
			if pattern == "cve" || pattern == "vulnerability" {
				confidence = 0.9
			}
			
			securityFixes = append(securityFixes, types.SecurityFix{
				CVE:         "",
				Severity:    severity,
				Description: fmt.Sprintf("Detected security-related keyword: '%s'", pattern),
				Impact:      fmt.Sprintf("Potential %s severity security issue", severity),
				Confidence:  confidence,
			})
		}
	}
	
	return securityFixes
}

func (h *HeuristicProvider) detectNewFeatures(text string) []types.Feature {
	var features []types.Feature
	
	featurePatterns := []string{
		"new feature", "added", "introduce", "support for", "enhancement",
		"improvement", "new api", "new method", "new function",
	}
	
	for _, pattern := range featurePatterns {
		if strings.Contains(text, pattern) {
			features = append(features, types.Feature{
				Name:        pattern,
				Description: fmt.Sprintf("Detected new feature: '%s'", pattern),
				Type:        h.categorizeFeature(pattern),
				Confidence:  0.7,
			})
		}
	}
	
	return features
}

func (h *HeuristicProvider) detectBugFixes(text string) []types.BugFix {
	var bugFixes []types.BugFix
	
	bugPatterns := []string{
		"fix", "fixed", "bug", "issue", "problem", "resolve", "resolved",
		"correct", "corrected", "patch",
	}
	
	for _, pattern := range bugPatterns {
		if strings.Contains(text, pattern) {
			bugFixes = append(bugFixes, types.BugFix{
				Description: fmt.Sprintf("Detected bug fix: '%s'", pattern),
				Impact:      "medium",
				Confidence:  0.6,
			})
		}
	}
	
	return bugFixes
}

func (h *HeuristicProvider) detectDeprecations(text string) []types.Deprecation {
	var deprecations []types.Deprecation
	
	deprecationPatterns := []string{
		"deprecated", "deprecation", "will be removed", "legacy",
		"obsolete", "no longer recommended",
	}
	
	for _, pattern := range deprecationPatterns {
		if strings.Contains(text, pattern) {
			deprecations = append(deprecations, types.Deprecation{
				Item:       pattern,
				Timeline:   "unknown",
				Confidence: 0.8,
			})
		}
	}
	
	return deprecations
}

func (h *HeuristicProvider) determineUpdateType(fromVersion, toVersion string) string {
	// Simple semantic version parsing
	fromParts := h.parseVersion(fromVersion)
	toParts := h.parseVersion(toVersion)
	
	if len(fromParts) < 3 || len(toParts) < 3 {
		return "unknown"
	}
	
	if toParts[0] > fromParts[0] {
		return "major"
	} else if toParts[1] > fromParts[1] {
		return "minor"
	} else if toParts[2] > fromParts[2] {
		return "patch"
	}
	
	// Check for prerelease
	if strings.Contains(toVersion, "-") {
		return "prerelease"
	}
	
	return "patch"
}

func (h *HeuristicProvider) parseVersion(version string) []int {
	// Remove 'v' prefix if present
	version = strings.TrimPrefix(version, "v")
	
	// Extract numeric parts
	re := regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)
	matches := re.FindStringSubmatch(version)
	
	if len(matches) < 4 {
		return []int{0, 0, 0}
	}
	
	var parts []int
	for i := 1; i < 4; i++ {
		var num int
		fmt.Sscanf(matches[i], "%d", &num)
		parts = append(parts, num)
	}
	
	return parts
}

func (h *HeuristicProvider) calculateRiskLevel(response *types.ChangelogAnalysisResponse) types.RiskLevel {
	if len(response.SecurityFixes) > 0 {
		for _, fix := range response.SecurityFixes {
			if fix.Severity == "critical" {
				return types.RiskLevelCritical
			}
		}
		return types.RiskLevelHigh
	}
	
	if response.HasBreakingChange {
		return types.RiskLevelHigh
	}
	
	if len(response.Deprecations) > 0 {
		return types.RiskLevelMedium
	}
	
	return types.RiskLevelLow
}

func (h *HeuristicProvider) calculateVersionRiskLevel(updateType string) types.RiskLevel {
	switch updateType {
	case "major":
		return types.RiskLevelHigh
	case "minor":
		return types.RiskLevelMedium
	case "patch":
		return types.RiskLevelLow
	case "prerelease":
		return types.RiskLevelMedium
	default:
		return types.RiskLevelLow
	}
}

func (h *HeuristicProvider) calculateConfidence(text string, response *types.ChangelogAnalysisResponse) float64 {
	baseConfidence := 0.6
	
	// Increase confidence if explicit keywords are found
	if strings.Contains(text, "breaking change") {
		baseConfidence += 0.2
	}
	
	if strings.Contains(text, "security") {
		baseConfidence += 0.1
	}
	
	// Decrease confidence if text is very short
	if len(text) < 100 {
		baseConfidence -= 0.2
	}
	
	if baseConfidence > 1.0 {
		baseConfidence = 1.0
	}
	if baseConfidence < 0.1 {
		baseConfidence = 0.1
	}
	
	return baseConfidence
}

func (h *HeuristicProvider) generateSummary(response *types.ChangelogAnalysisResponse) string {
	var parts []string
	
	if response.HasBreakingChange {
		parts = append(parts, fmt.Sprintf("%d breaking change(s)", len(response.BreakingChanges)))
	}
	
	if len(response.SecurityFixes) > 0 {
		parts = append(parts, fmt.Sprintf("%d security fix(es)", len(response.SecurityFixes)))
	}
	
	if len(response.NewFeatures) > 0 {
		parts = append(parts, fmt.Sprintf("%d new feature(s)", len(response.NewFeatures)))
	}
	
	if len(response.BugFixes) > 0 {
		parts = append(parts, fmt.Sprintf("%d bug fix(es)", len(response.BugFixes)))
	}
	
	if len(parts) == 0 {
		return "Standard update with no significant changes detected"
	}
	
	return "Update contains: " + strings.Join(parts, ", ")
}

func (h *HeuristicProvider) generateRecommendations(response *types.ChangelogAnalysisResponse) []string {
	var recommendations []string
	
	if response.HasBreakingChange {
		recommendations = append(recommendations, "Review breaking changes carefully before updating")
		recommendations = append(recommendations, "Test thoroughly in a development environment")
	}
	
	if len(response.SecurityFixes) > 0 {
		recommendations = append(recommendations, "Apply security updates promptly")
	}
	
	if response.RiskLevel == types.RiskLevelHigh || response.RiskLevel == types.RiskLevelCritical {
		recommendations = append(recommendations, "Consider updating during maintenance window")
		recommendations = append(recommendations, "Have rollback plan ready")
	}
	
	return recommendations
}

// Additional helper methods for other analysis types...

func (h *HeuristicProvider) determineSemanticImpact(updateType string) string {
	switch updateType {
	case "major":
		return "Potentially breaking changes, new features, and bug fixes"
	case "minor":
		return "New features and bug fixes, backward compatible"
	case "patch":
		return "Bug fixes and security patches, backward compatible"
	case "prerelease":
		return "Experimental features, use with caution"
	default:
		return "Unknown impact"
	}
}

func (h *HeuristicProvider) calculateCompatibilityScore(updateType string) float64 {
	switch updateType {
	case "major":
		return 0.3 // Low compatibility
	case "minor":
		return 0.8 // High compatibility
	case "patch":
		return 0.95 // Very high compatibility
	case "prerelease":
		return 0.5 // Medium compatibility
	default:
		return 0.7 // Default medium-high
	}
}

func (h *HeuristicProvider) generatePotentialIssues(updateType string) []types.CompatibilityIssue {
	var issues []types.CompatibilityIssue
	
	switch updateType {
	case "major":
		issues = append(issues, types.CompatibilityIssue{
			Type:        "breaking_change",
			Description: "Major version updates may contain breaking changes",
			Severity:    "high",
			Likelihood:  0.7,
			Mitigation:  "Review changelog and test thoroughly",
		})
	case "minor":
		issues = append(issues, types.CompatibilityIssue{
			Type:        "behavior_change",
			Description: "Minor updates may introduce subtle behavior changes",
			Severity:    "low",
			Likelihood:  0.2,
			Mitigation:  "Run existing tests to verify behavior",
		})
	}
	
	return issues
}

func (h *HeuristicProvider) generateMigrationSteps(updateType string) []string {
	switch updateType {
	case "major":
		return []string{
			"Review breaking changes in changelog",
			"Update code to handle API changes",
			"Update tests for new behavior",
			"Test thoroughly before deployment",
		}
	case "minor":
		return []string{
			"Review new features",
			"Consider adopting new functionality",
			"Run existing tests",
		}
	case "patch":
		return []string{
			"Review bug fixes",
			"Run regression tests",
		}
	default:
		return []string{"Review changelog and test"}
	}
}

func (h *HeuristicProvider) generateTestingRecommendations(updateType string) []string {
	switch updateType {
	case "major":
		return []string{
			"Run full test suite",
			"Perform integration testing",
			"Test in staging environment",
			"Consider canary deployment",
		}
	case "minor":
		return []string{
			"Run unit tests",
			"Test new features if adopted",
			"Perform smoke testing",
		}
	case "patch":
		return []string{
			"Run relevant unit tests",
			"Test affected functionality",
		}
	default:
		return []string{"Run basic tests"}
	}
}

func (h *HeuristicProvider) generateCompatibilitySummary(updateType string, score float64) string {
	scoreText := "medium"
	if score >= 0.8 {
		scoreText = "high"
	} else if score < 0.5 {
		scoreText = "low"
	}
	
	return fmt.Sprintf("%s update with %s compatibility score (%.1f)", 
		strings.Title(updateType), scoreText, score)
}

func (h *HeuristicProvider) classifyUpdateCategories(changelogText, releaseNotes string) []types.UpdateCategory {
	text := strings.ToLower(changelogText + " " + releaseNotes)
	var categories []types.UpdateCategory
	
	categoryWeights := map[string]float64{
		"security":    0.0,
		"feature":     0.0,
		"bugfix":      0.0,
		"maintenance": 0.0,
	}
	
	// Security keywords
	securityKeywords := []string{"security", "vulnerability", "cve", "exploit"}
	for _, keyword := range securityKeywords {
		if strings.Contains(text, keyword) {
			categoryWeights["security"] += 0.3
		}
	}
	
	// Feature keywords
	featureKeywords := []string{"feature", "added", "new", "enhancement"}
	for _, keyword := range featureKeywords {
		if strings.Contains(text, keyword) {
			categoryWeights["feature"] += 0.2
		}
	}
	
	// Bug fix keywords
	bugfixKeywords := []string{"fix", "bug", "issue", "resolve"}
	for _, keyword := range bugfixKeywords {
		if strings.Contains(text, keyword) {
			categoryWeights["bugfix"] += 0.2
		}
	}
	
	// Maintenance keywords
	maintenanceKeywords := []string{"refactor", "cleanup", "maintenance", "update"}
	for _, keyword := range maintenanceKeywords {
		if strings.Contains(text, keyword) {
			categoryWeights["maintenance"] += 0.1
		}
	}
	
	// Convert to categories
	for name, weight := range categoryWeights {
		if weight > 0 {
			categories = append(categories, types.UpdateCategory{
				Name:   name,
				Weight: weight,
			})
		}
	}
	
	// Default to maintenance if no categories detected
	if len(categories) == 0 {
		categories = append(categories, types.UpdateCategory{
			Name:   "maintenance",
			Weight: 0.5,
		})
	}
	
	return categories
}

func (h *HeuristicProvider) determinePriority(updateType string, categories []types.UpdateCategory) types.Priority {
	// Check for security updates
	for _, cat := range categories {
		if cat.Name == "security" && cat.Weight > 0.2 {
			return types.PriorityCritical
		}
	}
	
	// Check update type
	switch updateType {
	case "major":
		return types.PriorityMedium
	case "minor":
		return types.PriorityMedium
	case "patch":
		return types.PriorityLow
	default:
		return types.PriorityLow
	}
}

func (h *HeuristicProvider) determineUrgency(categories []types.UpdateCategory) types.Urgency {
	for _, cat := range categories {
		if cat.Name == "security" && cat.Weight > 0.3 {
			return types.UrgencyImmediate
		}
		if cat.Name == "security" && cat.Weight > 0.1 {
			return types.UrgencyHigh
		}
	}
	
	return types.UrgencyLow
}

func (h *HeuristicProvider) generateClassificationSummary(response *types.UpdateClassificationResponse) string {
	var parts []string
	
	parts = append(parts, fmt.Sprintf("%s update", strings.Title(response.UpdateType)))
	parts = append(parts, fmt.Sprintf("%s priority", response.Priority))
	parts = append(parts, fmt.Sprintf("%s risk", response.RiskLevel))
	
	if len(response.Categories) > 0 {
		var catNames []string
		for _, cat := range response.Categories {
			catNames = append(catNames, cat.Name)
		}
		parts = append(parts, fmt.Sprintf("categories: %s", strings.Join(catNames, ", ")))
	}
	
	return strings.Join(parts, ", ")
}

// Helper methods for impact and mitigation
func (h *HeuristicProvider) determineImpact(pattern string) string {
	highImpactPatterns := []string{"breaking change", "removed", "incompatible", "no longer"}
	for _, highPattern := range highImpactPatterns {
		if strings.Contains(pattern, highPattern) {
			return "high"
		}
	}
	return "medium"
}

func (h *HeuristicProvider) generateMitigation(changeType string) string {
	mitigations := map[string]string{
		"api_removal":      "Update code to use alternative APIs",
		"signature_change": "Update function calls to match new signature",
		"behavior_change":  "Review and test affected functionality",
		"deprecation":      "Plan migration to recommended alternative",
		"removal":          "Find alternative implementation",
	}
	
	if mitigation, exists := mitigations[changeType]; exists {
		return mitigation
	}
	
	return "Review changelog and update code accordingly"
}

func (h *HeuristicProvider) categorizeFeature(pattern string) string {
	if strings.Contains(pattern, "api") || strings.Contains(pattern, "method") || strings.Contains(pattern, "function") {
		return "new_api"
	}
	if strings.Contains(pattern, "performance") || strings.Contains(pattern, "optimization") {
		return "performance"
	}
	if strings.Contains(pattern, "enhancement") || strings.Contains(pattern, "improvement") {
		return "enhancement"
	}
	return "feature"
}
