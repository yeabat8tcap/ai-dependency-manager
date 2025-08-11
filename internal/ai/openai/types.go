package openai

import (
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/ai/types"
)

// ChangelogAnalysisResult represents the structured response from OpenAI for changelog analysis
type ChangelogAnalysisResult struct {
	PackageName       string                    `json:"package_name"`
	FromVersion       string                    `json:"from_version"`
	ToVersion         string                    `json:"to_version"`
	HasBreakingChange bool                      `json:"has_breaking_change"`
	BreakingChanges   []BreakingChangeAnalysis  `json:"breaking_changes"`
	NewFeatures       []FeatureAnalysis         `json:"new_features"`
	BugFixes          []BugFixAnalysis          `json:"bug_fixes"`
	SecurityFixes     []SecurityFixAnalysis     `json:"security_fixes"`
	Deprecations      []DeprecationAnalysis     `json:"deprecations"`
	RiskLevel         string                    `json:"risk_level"`
	RiskScore         float64                   `json:"risk_score"`
	Confidence        float64                   `json:"confidence"`
	Summary           string                    `json:"summary"`
	Recommendations   []string                  `json:"recommendations"`
	MigrationSteps    []string                  `json:"migration_steps"`
	TestingAdvice     []string                  `json:"testing_advice"`
	Timeline          string                    `json:"recommended_timeline"`
	BusinessImpact    string                    `json:"business_impact"`
}

// BreakingChangeAnalysis represents detailed breaking change analysis
type BreakingChangeAnalysis struct {
	Type         string  `json:"type"`
	Description  string  `json:"description"`
	Impact       string  `json:"impact"`
	Severity     string  `json:"severity"`
	Confidence   float64 `json:"confidence"`
	Mitigation   string  `json:"mitigation"`
	AffectedAPIs []string `json:"affected_apis"`
}

// FeatureAnalysis represents detailed feature analysis
type FeatureAnalysis struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Type        string  `json:"type"`
	Impact      string  `json:"impact"`
	Confidence  float64 `json:"confidence"`
	Benefits    []string `json:"benefits"`
	UsageExample string  `json:"usage_example"`
}

// BugFixAnalysis represents detailed bug fix analysis
type BugFixAnalysis struct {
	Description string  `json:"description"`
	Impact      string  `json:"impact"`
	Severity    string  `json:"severity"`
	Confidence  float64 `json:"confidence"`
	AffectedComponents []string `json:"affected_components"`
}

// SecurityFixAnalysis represents detailed security fix analysis
type SecurityFixAnalysis struct {
	Description string   `json:"description"`
	Severity    string   `json:"severity"`
	CVE         string   `json:"cve"`
	CVSS        float64  `json:"cvss"`
	Impact      string   `json:"impact"`
	Confidence  float64  `json:"confidence"`
	References  []string `json:"references"`
	Urgency     string   `json:"urgency"`
}

// DeprecationAnalysis represents detailed deprecation analysis
type DeprecationAnalysis struct {
	API          string `json:"api"`
	Replacement  string `json:"replacement"`
	Timeline     string `json:"timeline"`
	Impact       string `json:"impact"`
	MigrationGuide string `json:"migration_guide"`
}

// VersionDiffAnalysisResult represents the structured response for version diff analysis
type VersionDiffAnalysisResult struct {
	PackageName     string                    `json:"package_name"`
	FromVersion     string                    `json:"from_version"`
	ToVersion       string                    `json:"to_version"`
	UpdateType      string                    `json:"update_type"`
	SemanticImpact  string                    `json:"semantic_impact"`
	APIChanges      []APIChangeAnalysis       `json:"api_changes"`
	BehaviorChanges []BehaviorChangeAnalysis  `json:"behavior_changes"`
	RiskLevel       string                    `json:"risk_level"`
	RiskScore       float64                   `json:"risk_score"`
	Confidence      float64                   `json:"confidence"`
	Summary         string                    `json:"summary"`
	Recommendations []string                  `json:"recommendations"`
	MigrationEffort string                    `json:"migration_effort"`
	BackwardCompatibility bool                `json:"backward_compatibility"`
}

// APIChangeAnalysis represents detailed API change analysis
type APIChangeAnalysis struct {
	Type        string   `json:"type"`
	API         string   `json:"api"`
	Description string   `json:"description"`
	Impact      string   `json:"impact"`
	Severity    string   `json:"severity"`
	Examples    []string `json:"examples"`
	Migration   string   `json:"migration"`
}

// BehaviorChangeAnalysis represents detailed behavior change analysis
type BehaviorChangeAnalysis struct {
	Component   string `json:"component"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
	Likelihood  float64 `json:"likelihood"`
	TestingAdvice string `json:"testing_advice"`
}

// CompatibilityPredictionResult represents the structured response for compatibility prediction
type CompatibilityPredictionResult struct {
	PackageName             string                      `json:"package_name"`
	FromVersion             string                      `json:"from_version"`
	ToVersion               string                      `json:"to_version"`
	CompatibilityScore      float64                     `json:"compatibility_score"`
	RiskLevel               string                      `json:"risk_level"`
	RiskScore               float64                     `json:"risk_score"`
	Confidence              float64                     `json:"confidence"`
	PotentialIssues         []CompatibilityIssueAnalysis `json:"potential_issues"`
	MigrationSteps          []MigrationStepAnalysis     `json:"migration_steps"`
	TestingRecommendations  []TestingRecommendation     `json:"testing_recommendations"`
	Summary                 string                      `json:"summary"`
	Recommendations         []string                    `json:"recommendations"`
	EstimatedEffort         string                      `json:"estimated_effort"`
	RollbackComplexity      string                      `json:"rollback_complexity"`
}

// CompatibilityIssueAnalysis represents detailed compatibility issue analysis
type CompatibilityIssueAnalysis struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Severity    string  `json:"severity"`
	Likelihood  float64 `json:"likelihood"`
	Impact      string  `json:"impact"`
	Mitigation  string  `json:"mitigation"`
	Detection   string  `json:"detection"`
}

// MigrationStepAnalysis represents detailed migration step analysis
type MigrationStepAnalysis struct {
	Step        string `json:"step"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Effort      string `json:"effort"`
	Risk        string `json:"risk"`
	Validation  string `json:"validation"`
}

// TestingRecommendation represents testing recommendations
type TestingRecommendation struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Priority    string   `json:"priority"`
	TestCases   []string `json:"test_cases"`
	Tools       []string `json:"tools"`
}

// UpdateClassificationResult represents the structured response for update classification
type UpdateClassificationResult struct {
	PackageName      string                    `json:"package_name"`
	FromVersion      string                    `json:"from_version"`
	ToVersion        string                    `json:"to_version"`
	UpdateType       string                    `json:"update_type"`
	Priority         string                    `json:"priority"`
	Categories       []UpdateCategoryAnalysis  `json:"categories"`
	Urgency          string                    `json:"urgency"`
	Timeline         string                    `json:"recommended_timeline"`
	BusinessImpact   string                    `json:"business_impact"`
	TechnicalImpact  string                    `json:"technical_impact"`
	RiskAssessment   RiskAssessmentAnalysis    `json:"risk_assessment"`
	Summary          string                    `json:"summary"`
	Recommendations  []string                  `json:"recommendations"`
	Dependencies     []DependencyImpactAnalysis `json:"dependency_impacts"`
}

// UpdateCategoryAnalysis represents detailed update category analysis
type UpdateCategoryAnalysis struct {
	Name        string  `json:"name"`
	Weight      float64 `json:"weight"`
	Description string  `json:"description"`
	Impact      string  `json:"impact"`
	Examples    []string `json:"examples"`
}

// RiskAssessmentAnalysis represents detailed risk assessment
type RiskAssessmentAnalysis struct {
	Level       string  `json:"level"`
	Score       float64 `json:"score"`
	Factors     []string `json:"factors"`
	Mitigation  []string `json:"mitigation"`
	Monitoring  []string `json:"monitoring"`
}

// DependencyImpactAnalysis represents dependency impact analysis
type DependencyImpactAnalysis struct {
	Dependency  string  `json:"dependency"`
	Impact      string  `json:"impact"`
	Likelihood  float64 `json:"likelihood"`
	Mitigation  string  `json:"mitigation"`
}

// Conversion functions to transform OpenAI results to our standard types

// convertToChangelogResponse converts OpenAI result to standard changelog response
func (o *OpenAIProvider) convertToChangelogResponse(request *types.ChangelogAnalysisRequest, result *ChangelogAnalysisResult) *types.ChangelogAnalysisResponse {
	response := &types.ChangelogAnalysisResponse{
		PackageName:       result.PackageName,
		FromVersion:       result.FromVersion,
		ToVersion:         result.ToVersion,
		HasBreakingChange: result.HasBreakingChange,
		RiskScore:         result.RiskScore,
		Confidence:        result.Confidence,
		ConfidenceScore:   result.Confidence,
		Summary:           result.Summary,
		Recommendations:   result.Recommendations,
		AnalyzedAt:        time.Now(),
	}

	// Convert risk level
	response.RiskLevel = o.parseRiskLevel(result.RiskLevel)

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

	// Convert features
	for _, f := range result.NewFeatures {
		response.NewFeatures = append(response.NewFeatures, types.Feature{
			Name:        f.Name,
			Description: f.Description,
			Type:        f.Type,
			Impact:      f.Impact,
			Confidence:  f.Confidence,
		})
	}

	// Convert bug fixes
	for _, bf := range result.BugFixes {
		response.BugFixes = append(response.BugFixes, types.BugFix{
			Description: bf.Description,
			Impact:      bf.Impact,
			Confidence:  bf.Confidence,
			Severity:    bf.Severity,
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
	for _, d := range result.Deprecations {
		response.Deprecations = append(response.Deprecations, types.Deprecation{
			Item:        d.API,
			Replacement: d.Replacement,
			Timeline:    d.Timeline,
			Confidence:  1.0,
		})
	}

	return response
}

// convertToVersionDiffResponse converts OpenAI result to standard version diff response
func (o *OpenAIProvider) convertToVersionDiffResponse(request *types.VersionDiffAnalysisRequest, result *VersionDiffAnalysisResult) *types.VersionDiffAnalysisResponse {
	response := &types.VersionDiffAnalysisResponse{
		PackageName:     result.PackageName,
		FromVersion:     result.FromVersion,
		ToVersion:       result.ToVersion,
		UpdateType:      result.UpdateType,
		SemanticImpact:  result.SemanticImpact,
		RiskScore:       result.RiskScore,
		Confidence:      result.Confidence,
		ConfidenceScore: result.Confidence,
		Summary:         result.Summary,
		Recommendations: result.Recommendations,
		ProcessedAt:     time.Now(),
		AnalyzedAt:      time.Now(),
	}

	// Convert risk level
	response.RiskLevel = o.parseRiskLevel(result.RiskLevel)

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

// convertToCompatibilityResponse converts OpenAI result to standard compatibility response
func (o *OpenAIProvider) convertToCompatibilityResponse(request *types.CompatibilityPredictionRequest, result *CompatibilityPredictionResult) *types.CompatibilityPredictionResponse {
	response := &types.CompatibilityPredictionResponse{
		PackageName:        result.PackageName,
		FromVersion:        result.FromVersion,
		ToVersion:          result.ToVersion,
		CompatibilityScore: result.CompatibilityScore,
		RiskScore:          result.RiskScore,
		Confidence:         result.Confidence,
		ConfidenceScore:    result.Confidence,
		Summary:            result.Summary,
		Recommendations:    result.Recommendations,
		ProcessedAt:        time.Now(),
		PredictedAt:        time.Now(),
	}

	// Convert risk level
	response.RiskLevel = o.parseRiskLevel(result.RiskLevel)

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

	// Convert migration steps
	for _, step := range result.MigrationSteps {
		response.MigrationSteps = append(response.MigrationSteps, step.Description)
	}

	// Convert testing recommendations
	for _, test := range result.TestingRecommendations {
		response.TestingRecommendations = append(response.TestingRecommendations, test.Description)
	}

	return response
}

// convertToUpdateClassificationResponse converts OpenAI result to standard update classification response
func (o *OpenAIProvider) convertToUpdateClassificationResponse(request *types.UpdateClassificationRequest, result *UpdateClassificationResult) *types.UpdateClassificationResponse {
	response := &types.UpdateClassificationResponse{
		PackageName:     result.PackageName,
		FromVersion:     result.FromVersion,
		ToVersion:       result.ToVersion,
		UpdateType:      result.UpdateType,
		Priority:        o.parsePriority(result.Priority),
		Summary:         result.Summary,
		Recommendations: result.Recommendations,
		ProcessedAt:     time.Now(),
		ClassifiedAt:    time.Now(),
	}

	// Convert categories
	for _, cat := range result.Categories {
		response.Categories = append(response.Categories, types.UpdateCategory{
			Name:   cat.Name,
			Weight: cat.Weight,
		})
	}

	return response
}

// parseRiskLevel converts string risk level to our enum type
func (o *OpenAIProvider) parseRiskLevel(level string) types.RiskLevel {
	switch level {
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

// parsePriority converts string priority to our enum type
func (o *OpenAIProvider) parsePriority(priority string) types.Priority {
	switch priority {
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
