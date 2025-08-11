package claude

import (
	"strings"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/ai/types"
)

// parseRiskLevel converts string risk level to types.RiskLevel
func parseRiskLevel(level string) types.RiskLevel {
	switch strings.ToLower(level) {
	case "low":
		return types.RiskLevelLow
	case "medium":
		return types.RiskLevelMedium
	case "high":
		return types.RiskLevelHigh
	case "critical":
		return types.RiskLevelCritical
	default:
		return types.RiskLevelMedium
	}
}

// ChangelogAnalysisResult represents Claude's response for changelog analysis
type ChangelogAnalysisResult struct {
	PackageName       string                    `json:"package_name"`
	FromVersion       string                    `json:"from_version"`
	ToVersion         string                    `json:"to_version"`
	HasBreakingChange bool                      `json:"has_breaking_change"`
	BreakingChanges   []ClaudeBreakingChange    `json:"breaking_changes"`
	NewFeatures       []ClaudeNewFeature        `json:"new_features"`
	BugFixes          []ClaudeBugFix            `json:"bug_fixes"`
	SecurityFixes     []ClaudeSecurityFix       `json:"security_fixes"`
	Deprecations      []ClaudeDeprecation       `json:"deprecations"`
	RiskLevel         string                    `json:"risk_level"`
	RiskScore         float64                   `json:"risk_score"`
	Confidence        float64                   `json:"confidence"`
	Summary           string                    `json:"summary"`
	Recommendations   []string                  `json:"recommendations"`
	MigrationSteps    []string                  `json:"migration_steps"`
	TestingAdvice     []string                  `json:"testing_advice"`
	RecommendedTimeline string                  `json:"recommended_timeline"`
	BusinessImpact    string                    `json:"business_impact"`
}

type ClaudeBreakingChange struct {
	Type          string   `json:"type"`
	Description   string   `json:"description"`
	Impact        string   `json:"impact"`
	Severity      string   `json:"severity"`
	Confidence    float64  `json:"confidence"`
	Mitigation    string   `json:"mitigation"`
	AffectedAPIs  []string `json:"affected_apis"`
}

type ClaudeNewFeature struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Type         string   `json:"type"`
	Impact       string   `json:"impact"`
	Confidence   float64  `json:"confidence"`
	Benefits     []string `json:"benefits"`
	UsageExample string   `json:"usage_example"`
}

type ClaudeBugFix struct {
	Description         string   `json:"description"`
	Impact              string   `json:"impact"`
	Severity            string   `json:"severity"`
	Confidence          float64  `json:"confidence"`
	AffectedComponents  []string `json:"affected_components"`
}

type ClaudeSecurityFix struct {
	Description string   `json:"description"`
	Severity    string   `json:"severity"`
	CVE         string   `json:"cve"`
	CVSS        float64  `json:"cvss"`
	Impact      string   `json:"impact"`
	Confidence  float64  `json:"confidence"`
	References  []string `json:"references"`
	Urgency     string   `json:"urgency"`
}

type ClaudeDeprecation struct {
	API           string `json:"api"`
	Replacement   string `json:"replacement"`
	Timeline      string `json:"timeline"`
	Impact        string `json:"impact"`
	MigrationGuide string `json:"migration_guide"`
}

// VersionDiffAnalysisResult represents Claude's response for version diff analysis
type VersionDiffAnalysisResult struct {
	PackageName          string                  `json:"package_name"`
	FromVersion          string                  `json:"from_version"`
	ToVersion            string                  `json:"to_version"`
	UpdateType           string                  `json:"update_type"`
	SemanticImpact       string                  `json:"semantic_impact"`
	APIChanges           []ClaudeAPIChange       `json:"api_changes"`
	BehaviorChanges      []ClaudeBehaviorChange  `json:"behavior_changes"`
	RiskLevel            string                  `json:"risk_level"`
	RiskScore            float64                 `json:"risk_score"`
	Confidence           float64                 `json:"confidence"`
	Summary              string                  `json:"summary"`
	Recommendations      []string                `json:"recommendations"`
	MigrationEffort      string                  `json:"migration_effort"`
	BackwardCompatibility bool                   `json:"backward_compatibility"`
}

type ClaudeAPIChange struct {
	Type        string   `json:"type"`
	API         string   `json:"api"`
	Description string   `json:"description"`
	Impact      string   `json:"impact"`
	Severity    string   `json:"severity"`
	Examples    []string `json:"examples"`
	Migration   string   `json:"migration"`
}

type ClaudeBehaviorChange struct {
	Component     string  `json:"component"`
	Description   string  `json:"description"`
	Impact        string  `json:"impact"`
	Likelihood    float64 `json:"likelihood"`
	TestingAdvice string  `json:"testing_advice"`
}

// CompatibilityPredictionResult represents Claude's response for compatibility prediction
type CompatibilityPredictionResult struct {
	PackageName          string                        `json:"package_name"`
	FromVersion          string                        `json:"from_version"`
	ToVersion            string                        `json:"to_version"`
	CompatibilityScore   float64                       `json:"compatibility_score"`
	RiskLevel            string                        `json:"risk_level"`
	RiskScore            float64                       `json:"risk_score"`
	Confidence           float64                       `json:"confidence"`
	PotentialIssues      []ClaudeCompatibilityIssue    `json:"potential_issues"`
	MigrationSteps       []ClaudeMigrationStep         `json:"migration_steps"`
	TestingRecommendations []ClaudeTestingRecommendation `json:"testing_recommendations"`
	Summary              string                        `json:"summary"`
	Recommendations      []string                      `json:"recommendations"`
	EstimatedEffort      string                        `json:"estimated_effort"`
	RollbackComplexity   string                        `json:"rollback_complexity"`
	RiskFactors          []string                      `json:"risk_factors"`
}

type ClaudeCompatibilityIssue struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Severity    string  `json:"severity"`
	Likelihood  float64 `json:"likelihood"`
	Impact      string  `json:"impact"`
	Mitigation  string  `json:"mitigation"`
	Detection   string  `json:"detection"`
}

type ClaudeMigrationStep struct {
	Step        string `json:"step"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Effort      string `json:"effort"`
	Risk        string `json:"risk"`
	Validation  string `json:"validation"`
}

type ClaudeTestingRecommendation struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Priority    string   `json:"priority"`
	TestCases   []string `json:"test_cases"`
	Tools       []string `json:"tools"`
}

// UpdateClassificationResult represents Claude's response for update classification
type UpdateClassificationResult struct {
	PackageName          string                      `json:"package_name"`
	FromVersion          string                      `json:"from_version"`
	ToVersion            string                      `json:"to_version"`
	UpdateType           string                      `json:"update_type"`
	Priority             string                      `json:"priority"`
	Categories           []ClaudeUpdateCategory      `json:"categories"`
	Urgency              string                      `json:"urgency"`
	RecommendedTimeline  string                      `json:"recommended_timeline"`
	BusinessImpact       string                      `json:"business_impact"`
	TechnicalImpact      string                      `json:"technical_impact"`
	RiskAssessment       ClaudeRiskAssessment        `json:"risk_assessment"`
	Summary              string                      `json:"summary"`
	Recommendations      []string                    `json:"recommendations"`
	DependencyImpacts    []ClaudeDependencyImpact    `json:"dependency_impacts"`
}

type ClaudeUpdateCategory struct {
	Name        string   `json:"name"`
	Weight      float64  `json:"weight"`
	Description string   `json:"description"`
	Impact      string   `json:"impact"`
	Examples    []string `json:"examples"`
}

type ClaudeRiskAssessment struct {
	Level       string   `json:"level"`
	Score       float64  `json:"score"`
	Factors     []string `json:"factors"`
	Mitigation  []string `json:"mitigation"`
	Monitoring  []string `json:"monitoring"`
}

type ClaudeDependencyImpact struct {
	Dependency string  `json:"dependency"`
	Impact     string  `json:"impact"`
	Likelihood float64 `json:"likelihood"`
	Mitigation string  `json:"mitigation"`
}

// Conversion functions to standard internal types

func (c *ClaudeProvider) convertToChangelogResponse(request *types.ChangelogAnalysisRequest, result *ChangelogAnalysisResult) *types.ChangelogAnalysisResponse {
	response := &types.ChangelogAnalysisResponse{
		PackageName: result.PackageName,
		FromVersion: result.FromVersion,
		ToVersion:   result.ToVersion,
		AnalyzedAt:  time.Now(),
		RiskLevel:   parseRiskLevel(result.RiskLevel),
		RiskScore:   result.RiskScore,
		Confidence:  result.Confidence,
		Summary:     result.Summary,
		Recommendations: result.Recommendations,
		HasBreakingChange: result.HasBreakingChange,
	}

	// Convert breaking changes
	for _, bc := range result.BreakingChanges {
		response.BreakingChanges = append(response.BreakingChanges, types.BreakingChange{
			Type:        bc.Type,
			Description: bc.Description,
			Impact:      bc.Impact,
			Confidence:  bc.Confidence,
			Mitigation:  bc.Mitigation,
		})
	}

	// Convert new features
	for _, nf := range result.NewFeatures {
		response.NewFeatures = append(response.NewFeatures, types.Feature{
			Name:        nf.Name,
			Description: nf.Description,
			Type:        nf.Type,
			Impact:      nf.Impact,
			Confidence:  nf.Confidence,
		})
	}

	// Convert bug fixes
	for _, bf := range result.BugFixes {
		response.BugFixes = append(response.BugFixes, types.BugFix{
			Description: bf.Description,
			Impact:      bf.Impact,
			Severity:    bf.Severity,
			Confidence:  bf.Confidence,
		})
	}

	// Convert security fixes
	for _, sf := range result.SecurityFixes {
		response.SecurityFixes = append(response.SecurityFixes, types.SecurityFix{
			CVE:         sf.CVE,
			Severity:    sf.Severity,
			Description: sf.Description,
			Impact:      sf.Impact,
			Confidence:  sf.Confidence,
		})
	}

	// Convert deprecations
	for _, dep := range result.Deprecations {
		response.Deprecations = append(response.Deprecations, types.Deprecation{
			Item:        dep.API,
			Replacement: dep.Replacement,
			Timeline:    dep.Timeline,
			Confidence:  1.0,
		})
	}

	return response
}

// parsePriority converts string priority to types.Priority
func parsePriority(priority string) types.Priority {
	switch strings.ToLower(priority) {
	case "low":
		return types.PriorityLow
	case "medium":
		return types.PriorityMedium
	case "high":
		return types.PriorityHigh
	case "critical":
		return types.PriorityCritical
	default:
		return types.PriorityMedium
	}
}

// parseUrgency converts string urgency to types.Urgency
func parseUrgency(urgency string) types.Urgency {
	switch strings.ToLower(urgency) {
	case "low":
		return types.UrgencyLow
	case "medium":
		return types.UrgencyMedium
	case "high":
		return types.UrgencyHigh
	case "critical", "immediate":
		return types.UrgencyImmediate
	default:
		return types.UrgencyMedium
	}
}

func (c *ClaudeProvider) convertToVersionDiffResponse(request *types.VersionDiffAnalysisRequest, result *VersionDiffAnalysisResult) *types.VersionDiffAnalysisResponse {
	response := &types.VersionDiffAnalysisResponse{
		PackageName:     result.PackageName,
		FromVersion:     result.FromVersion,
		ToVersion:       result.ToVersion,
		UpdateType:      result.UpdateType,
		SemanticImpact:  result.SemanticImpact,
		RiskLevel:       parseRiskLevel(result.RiskLevel),
		RiskScore:       result.RiskScore,
		Confidence:      result.Confidence,
		Summary:         result.Summary,
		Recommendations: result.Recommendations,
		AnalyzedAt:      time.Now(),
	}

	// Convert API changes
	for _, ac := range result.APIChanges {
		response.APIChanges = append(response.APIChanges, types.APIChange{
			Type:        ac.Type,
			API:         ac.API,
			Description: ac.Description,
			Impact:      ac.Impact,
		})
	}

	// Convert behavior changes
	for _, bc := range result.BehaviorChanges {
		response.BehaviorChanges = append(response.BehaviorChanges, types.BehaviorChange{
			Component:   bc.Component,
			Description: bc.Description,
			Impact:      bc.Impact,
		})
	}

	return response
}

func (c *ClaudeProvider) convertToCompatibilityResponse(request *types.CompatibilityPredictionRequest, result *CompatibilityPredictionResult) *types.CompatibilityPredictionResponse {
	response := &types.CompatibilityPredictionResponse{
		PackageName:        result.PackageName,
		FromVersion:        result.FromVersion,
		ToVersion:          result.ToVersion,
		CompatibilityScore: result.CompatibilityScore,
		RiskLevel:          parseRiskLevel(result.RiskLevel),
		RiskScore:          result.RiskScore,
		Confidence:         result.Confidence,
		Summary:            result.Summary,
		Recommendations:    result.Recommendations,
	}
	// Convert potential issues
	for _, issue := range result.PotentialIssues {
		response.PotentialIssues = append(response.PotentialIssues, types.CompatibilityIssue{
			Type:        issue.Type,
			Description: issue.Description,
			Severity:    issue.Severity,
			Likelihood:  issue.Likelihood,

			Mitigation:  issue.Mitigation,
		})
	}

	return response
}

func (c *ClaudeProvider) convertToUpdateClassificationResponse(request *types.UpdateClassificationRequest, result *UpdateClassificationResult) *types.UpdateClassificationResponse {
	response := &types.UpdateClassificationResponse{
		PackageName:         result.PackageName,
		FromVersion:         result.FromVersion,
		ToVersion:           result.ToVersion,
		UpdateType:          result.UpdateType,
		Priority:            parsePriority(result.Priority),
		Urgency:             parseUrgency(result.Urgency),

		Summary:             result.Summary,
		Recommendations:     result.Recommendations,
	}

	// Convert categories
	for _, cat := range result.Categories {
		response.Categories = append(response.Categories, types.UpdateCategory{
			Name:        cat.Name,
			Weight:      cat.Weight,
			Description: cat.Description,
		})
	}

	// Risk assessment and dependency impacts not supported in current types

	return response
}

// ... (rest of the code remains the same)
