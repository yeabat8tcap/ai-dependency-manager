package ollama

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

// OllamaChangelogAnalysisResult represents Ollama's response for changelog analysis
type OllamaChangelogAnalysisResult struct {
	PackageName       string                        `json:"package_name"`
	FromVersion       string                        `json:"from_version"`
	ToVersion         string                        `json:"to_version"`
	RiskLevel         string                        `json:"risk_level"`
	RiskScore         float64                       `json:"risk_score"`
	Confidence        float64                       `json:"confidence"`
	Summary           string                        `json:"summary"`
	Recommendations   []string                      `json:"recommendations"`
	BreakingChanges   []OllamaBreakingChange        `json:"breaking_changes"`
	NewFeatures       []OllamaFeature               `json:"new_features"`
	BugFixes          []OllamaBugFix                `json:"bug_fixes"`
	SecurityFixes     []OllamaSecurityFix           `json:"security_fixes"`
	Deprecations      []OllamaDeprecation           `json:"deprecations"`
}

type OllamaBreakingChange struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Impact      string  `json:"impact"`
	Confidence  float64 `json:"confidence"`
	Mitigation  string  `json:"mitigation"`
}

type OllamaFeature struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Type        string  `json:"type"`
	Impact      string  `json:"impact"`
	Confidence  float64 `json:"confidence"`
}

type OllamaBugFix struct {
	Description string  `json:"description"`
	Impact      string  `json:"impact"`
	Severity    string  `json:"severity"`
	Confidence  float64 `json:"confidence"`
}

type OllamaSecurityFix struct {
	CVE         string  `json:"cve"`
	Severity    string  `json:"severity"`
	Description string  `json:"description"`
	Impact      string  `json:"impact"`
	Confidence  float64 `json:"confidence"`
}

type OllamaDeprecation struct {
	API         string `json:"api"`
	Replacement string `json:"replacement"`
	Timeline    string `json:"timeline"`
}

// OllamaVersionDiffAnalysisResult represents Ollama's response for version diff analysis
type OllamaVersionDiffAnalysisResult struct {
	PackageName     string                    `json:"package_name"`
	FromVersion     string                    `json:"from_version"`
	ToVersion       string                    `json:"to_version"`
	UpdateType      string                    `json:"update_type"`
	SemanticImpact  string                    `json:"semantic_impact"`
	RiskLevel       string                    `json:"risk_level"`
	RiskScore       float64                   `json:"risk_score"`
	Confidence      float64                   `json:"confidence"`
	Summary         string                    `json:"summary"`
	Recommendations []string                  `json:"recommendations"`
	APIChanges      []OllamaAPIChange         `json:"api_changes"`
	BehaviorChanges []OllamaBehaviorChange    `json:"behavior_changes"`
}

type OllamaAPIChange struct {
	Type        string `json:"type"`
	API         string `json:"api"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
}

type OllamaBehaviorChange struct {
	Component   string `json:"component"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
}

// OllamaCompatibilityPredictionResult represents Ollama's response for compatibility prediction
type OllamaCompatibilityPredictionResult struct {
	PackageName        string                        `json:"package_name"`
	FromVersion        string                        `json:"from_version"`
	ToVersion          string                        `json:"to_version"`
	CompatibilityScore float64                       `json:"compatibility_score"`
	RiskLevel          string                        `json:"risk_level"`
	RiskScore          float64                       `json:"risk_score"`
	Confidence         float64                       `json:"confidence"`
	Summary            string                        `json:"summary"`
	Recommendations    []string                      `json:"recommendations"`
	PotentialIssues    []OllamaCompatibilityIssue    `json:"potential_issues"`
}

type OllamaCompatibilityIssue struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Severity    string  `json:"severity"`
	Likelihood  float64 `json:"likelihood"`
	Mitigation  string  `json:"mitigation"`
}

// OllamaUpdateClassificationResult represents Ollama's response for update classification
type OllamaUpdateClassificationResult struct {
	PackageName     string                    `json:"package_name"`
	FromVersion     string                    `json:"from_version"`
	ToVersion       string                    `json:"to_version"`
	UpdateType      string                    `json:"update_type"`
	Priority        string                    `json:"priority"`
	Urgency         string                    `json:"urgency"`
	Summary         string                    `json:"summary"`
	Recommendations []string                  `json:"recommendations"`
	Categories      []OllamaUpdateCategory    `json:"categories"`
}

type OllamaUpdateCategory struct {
	Name        string  `json:"name"`
	Weight      float64 `json:"weight"`
	Description string  `json:"description"`
}

// Conversion functions to internal types

func (o *OllamaProvider) convertToChangelogResponse(request *types.ChangelogAnalysisRequest, result *OllamaChangelogAnalysisResult) *types.ChangelogAnalysisResponse {
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

	// Convert features
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

func (o *OllamaProvider) convertToVersionDiffResponse(request *types.VersionDiffAnalysisRequest, result *OllamaVersionDiffAnalysisResult) *types.VersionDiffAnalysisResponse {
	response := &types.VersionDiffAnalysisResponse{
		PackageName:     result.PackageName,
		FromVersion:     result.FromVersion,
		ToVersion:       result.ToVersion,
		RiskLevel:       parseRiskLevel(result.RiskLevel),
		RiskScore:       result.RiskScore,
		Confidence:      result.Confidence,
		Summary:         result.Summary,
		Recommendations: result.Recommendations,
		ProcessedAt:     time.Now(),
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

func (o *OllamaProvider) convertToCompatibilityResponse(request *types.CompatibilityPredictionRequest, result *OllamaCompatibilityPredictionResult) *types.CompatibilityPredictionResponse {
	response := &types.CompatibilityPredictionResponse{
		PackageName:             request.PackageName,
		FromVersion:             request.FromVersion,
		ToVersion:               request.ToVersion,
		CompatibilityScore:      result.CompatibilityScore,
		RiskLevel:               parseRiskLevel(result.RiskLevel),
		Confidence:              result.Confidence,
		PotentialIssues:         []types.CompatibilityIssue{},
		Recommendations:         result.Recommendations,
		Summary:                 result.Summary,
		PredictedAt:             time.Now(),
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

func (o *OllamaProvider) convertToUpdateClassificationResponse(request *types.UpdateClassificationRequest, result *OllamaUpdateClassificationResult) *types.UpdateClassificationResponse {
	response := &types.UpdateClassificationResponse{
		PackageName:     result.PackageName,
		FromVersion:     result.FromVersion,
		ToVersion:       result.ToVersion,
		UpdateType:      result.UpdateType,
		Priority:        parsePriority(result.Priority),
		Urgency:         parseUrgency(result.Urgency),
		Summary:         result.Summary,
		Recommendations: result.Recommendations,
		ClassifiedAt:    time.Now(),
	}

	// Convert categories
	for _, cat := range result.Categories {
		response.Categories = append(response.Categories, types.UpdateCategory{
			Name:        cat.Name,
			Weight:      cat.Weight,
			Description: cat.Description,
		})
	}

	return response
}
