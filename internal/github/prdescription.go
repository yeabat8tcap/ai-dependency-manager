package github

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// PRDescriptionGenerator handles automated PR description generation
type PRDescriptionGenerator struct {
	aiManager AIManager
}

// NewPRDescriptionGenerator creates a new PR description generator
func NewPRDescriptionGenerator(aiManager AIManager) *PRDescriptionGenerator {
	return &PRDescriptionGenerator{
		aiManager: aiManager,
	}
}

// GenerateDescription generates a comprehensive PR description
func (pdg *PRDescriptionGenerator) GenerateDescription(ctx context.Context, request *PRCreationRequest) (*PRDescription, error) {
	// Analyze patches and dependencies
	analysis, err := pdg.analyzeChanges(ctx, request.Patches, request.Dependencies)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze changes: %w", err)
	}

	// Generate summary
	summary, err := pdg.generateSummary(ctx, analysis)
	if err != nil {
		return nil, fmt.Errorf("failed to generate summary: %w", err)
	}

	// Generate change descriptions
	changes := pdg.generateChangeDescriptions(request.Patches)

	// Generate dependency changes
	depChanges := pdg.generateDependencyChanges(request.Dependencies)

	// Perform risk assessment
	riskAssessment := pdg.performRiskAssessment(analysis, request.Dependencies)

	// Generate testing notes
	testingNotes := pdg.generateTestingNotes(analysis, riskAssessment)

	// Generate review guidance
	reviewGuidance := pdg.generateReviewGuidance(analysis, riskAssessment)

	// Identify breaking changes
	breakingChanges := pdg.identifyBreakingChanges(analysis, request.Dependencies)

	// Generate checklist
	checklist := pdg.generateChecklist(riskAssessment, breakingChanges)

	description := &PRDescription{
		Summary:         summary,
		Changes:         changes,
		Dependencies:    depChanges,
		RiskAssessment:  riskAssessment,
		TestingNotes:    testingNotes,
		ReviewGuidance:  reviewGuidance,
		BreakingChanges: breakingChanges,
		Checklist:       checklist,
		Metadata: map[string]interface{}{
			"generated_at": time.Now(),
			"generator":    "ai_powered",
			"version":      "1.0",
		},
	}

	return description, nil
}

// ChangeAnalysis represents the analysis of changes in patches
type ChangeAnalysis struct {
	TotalFiles           int                    `json:"total_files"`
	ModifiedFiles        []string               `json:"modified_files"`
	AddedFiles           []string               `json:"added_files"`
	DeletedFiles         []string               `json:"deleted_files"`
	CodeChanges          []*CodeChange          `json:"code_changes"`
	ConfigChanges        []*ConfigChange        `json:"config_changes"`
	TestChanges          []*TestChange          `json:"test_changes"`
	DocumentationChanges []*DocumentationChange `json:"documentation_changes"`
	Complexity           *ComplexityAnalysis    `json:"complexity"`
	Impact               *ImpactAnalysis        `json:"impact"`
	Dependencies         []*DependencyUpdate    `json:"dependencies"`
}

// CodeChange represents a code change
type CodeChange struct {
	File         string  `json:"file"`
	Type         string  `json:"type"` // "function", "class", "variable", "import"
	Name         string  `json:"name"`
	Action       string  `json:"action"` // "added", "modified", "deleted"
	LinesAdded   int     `json:"lines_added"`
	LinesRemoved int     `json:"lines_removed"`
	Complexity   int     `json:"complexity"`
	Confidence   float64 `json:"confidence"`
}

// ConfigChange represents a configuration change
type ConfigChange struct {
	File    string                 `json:"file"`
	Type    string                 `json:"type"` // "package.json", "requirements.txt", etc.
	Changes map[string]interface{} `json:"changes"`
	Impact  string                 `json:"impact"`
}

// TestChange represents a test-related change
type TestChange struct {
	File          string `json:"file"`
	Type          string `json:"type"` // "unit", "integration", "e2e"
	Action        string `json:"action"`
	TestsAdded    int    `json:"tests_added"`
	TestsModified int    `json:"tests_modified"`
	Coverage      string `json:"coverage"`
}

// DocumentationChange represents a documentation change
type DocumentationChange struct {
	File        string `json:"file"`
	Type        string `json:"type"` // "readme", "api", "changelog"
	Action      string `json:"action"`
	Description string `json:"description"`
}

// ComplexityAnalysis analyzes the complexity of changes
type ComplexityAnalysis struct {
	OverallComplexity    string  `json:"overall_complexity"` // "low", "medium", "high"
	CyclomaticComplexity int     `json:"cyclomatic_complexity"`
	CognitiveComplexity  int     `json:"cognitive_complexity"`
	LinesOfCode          int     `json:"lines_of_code"`
	FilesAffected        int     `json:"files_affected"`
	Score                float64 `json:"score"`
}

// ImpactAnalysis analyzes the impact of changes
type ImpactAnalysis struct {
	Scope             string   `json:"scope"` // "local", "module", "system", "global"
	AffectedModules   []string `json:"affected_modules"`
	UserFacing        bool     `json:"user_facing"`
	APIChanges        bool     `json:"api_changes"`
	DatabaseChanges   bool     `json:"database_changes"`
	ConfigChanges     bool     `json:"config_changes"`
	SecurityImpact    string   `json:"security_impact"`
	PerformanceImpact string   `json:"performance_impact"`
}

// analyzeChanges analyzes patches and dependencies to understand the changes
func (pdg *PRDescriptionGenerator) analyzeChanges(ctx context.Context, patches []*Patch, dependencies []*DependencyUpdate) (*ChangeAnalysis, error) {
	analysis := &ChangeAnalysis{
		TotalFiles:           len(patches), // Approximation
		ModifiedFiles:        []string{},
		AddedFiles:           []string{},
		DeletedFiles:         []string{},
		CodeChanges:          []*CodeChange{},
		ConfigChanges:        []*ConfigChange{},
		TestChanges:          []*TestChange{},
		DocumentationChanges: []*DocumentationChange{},
		Dependencies:         dependencies,
		Complexity: &ComplexityAnalysis{
			OverallComplexity: "low",
			Score:             1.0,
		},
		Impact: &ImpactAnalysis{
			Scope:      "minor",
			APIChanges: false,
		},
	}

	// Analyze file patches
	for _, patch := range patches {
		// analysis.TotalFiles++ // This is now handled by len(patches) above if patch represents a file.
		// If patch represents a group of file patches, this logic needs adjustment.
		// Assuming patch.FilePatches are the actual files.

		for _, filePatch := range patch.FilePatches {
			analysis.TotalFiles++ // Increment for each actual file patch
			switch filePatch.Type {
			case "create":
				analysis.AddedFiles = append(analysis.AddedFiles, filePatch.Path)
			case "delete":
				analysis.DeletedFiles = append(analysis.DeletedFiles, filePatch.Path)
			default:
				analysis.ModifiedFiles = append(analysis.ModifiedFiles, filePatch.Path)
			}
		}

		// Analyze file changes
		for i := range patch.FilePatches {
			filePatch := &patch.FilePatches[i]
			if strings.Contains(filePatch.Path, "test") || strings.Contains(filePatch.Path, "_test.go") {
				pdg.analyzeTestChange(filePatch, analysis)
			} else if strings.Contains(filePatch.Path, "config") || strings.Contains(filePatch.Path, ".json") || strings.Contains(filePatch.Path, ".yaml") {
				pdg.analyzeConfigChange(filePatch, analysis)
			} else if strings.Contains(filePatch.Path, "README") || strings.Contains(filePatch.Path, "docs/") {
				pdg.analyzeDocumentationChange(filePatch, analysis)
			} else {
				pdg.analyzeCodeChange(filePatch, analysis)
			}
		}
	}

	// Analyze complexity
	analysis.Complexity = pdg.analyzeComplexity(analysis)

	// Analyze impact
	analysis.Impact = pdg.analyzeImpact(analysis, dependencies)

	return analysis, nil
}

// generateSummary generates a summary using AI
func (pdg *PRDescriptionGenerator) generateSummary(ctx context.Context, analysis *ChangeAnalysis) (string, error) {
	prompt := pdg.buildSummaryPrompt(analysis)

	response, err := pdg.aiManager.AnalyzeChangelog(ctx, prompt)
	if err != nil {
		// Fallback to heuristic summary
		return pdg.generateHeuristicSummary(analysis), nil
	}

	return pdg.extractSummaryFromResponse(response), nil
}

// buildSummaryPrompt builds a prompt for AI summary generation
func (pdg *PRDescriptionGenerator) buildSummaryPrompt(analysis *ChangeAnalysis) string {
	return fmt.Sprintf(`
You are an expert software engineer writing a pull request summary.

Change Analysis:
- Total files changed: %d
- Files modified: %v
- Files added: %v
- Files deleted: %v
- Code changes: %d
- Config changes: %d
- Test changes: %d
- Documentation changes: %d
- Overall complexity: %s
- Impact scope: %s

Please write a concise, professional summary (2-3 sentences) that explains:
1. What this PR does
2. Why it's needed
3. The main impact

Focus on the business value and technical significance.
`, analysis.TotalFiles, analysis.ModifiedFiles, analysis.AddedFiles, analysis.DeletedFiles,
		len(analysis.CodeChanges), len(analysis.ConfigChanges), len(analysis.TestChanges),
		len(analysis.DocumentationChanges), analysis.Complexity.OverallComplexity, analysis.Impact.Scope)
}

// generateHeuristicSummary generates a summary using heuristics
func (pdg *PRDescriptionGenerator) generateHeuristicSummary(analysis *ChangeAnalysis) string {
	if len(analysis.Dependencies) > 0 {
		return fmt.Sprintf("Updates %d dependencies and modifies %d files to maintain compatibility and security.",
			len(analysis.Dependencies), analysis.TotalFiles)
	}

	if len(analysis.CodeChanges) > 0 {
		return fmt.Sprintf("Implements code changes across %d files with %s complexity impact.",
			len(analysis.ModifiedFiles), analysis.Complexity.OverallComplexity)
	}

	return fmt.Sprintf("Updates %d files with various improvements and maintenance changes.", analysis.TotalFiles)
}

// extractSummaryFromResponse extracts summary from AI response
func (pdg *PRDescriptionGenerator) extractSummaryFromResponse(response string) string {
	// Simple extraction - in real implementation would parse structured response
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 20 && !strings.HasPrefix(line, "##") && !strings.HasPrefix(line, "-") {
			return line
		}
	}
	return response
}

// generateChangeDescriptions generates descriptions for individual changes
func (pdg *PRDescriptionGenerator) generateChangeDescriptions(patches []*Patch) []*ChangeDescription {
	var changes []*ChangeDescription

	for _, patch := range patches {
		for i := range patch.FilePatches {
			filePatch := &patch.FilePatches[i]
			changeType := pdg.determineChangeType(filePatch)
			description := pdg.generateChangeDescription(filePatch)
			impact := pdg.assessChangeImpact(filePatch)

			change := &ChangeDescription{
				Type:        changeType,
				File:        filePatch.Path,
				Description: description,
				Impact:      impact,
				Confidence:  filePatch.Confidence,
			}
			changes = append(changes, change)
		}
	}

	return changes
}

// generateDependencyChanges generates descriptions for dependency changes
func (pdg *PRDescriptionGenerator) generateDependencyChanges(dependencies []*DependencyUpdate) []*DependencyChange {
	var changes []*DependencyChange

	for _, dep := range dependencies {
		change := &DependencyChange{
			Name:           dep.Name,
			CurrentVersion: dep.CurrentVersion,
			NewVersion:     dep.LatestVersion,
			ChangeType:     pdg.determineVersionChangeType(dep.CurrentVersion, dep.LatestVersion),
			RiskLevel:      string(dep.RiskLevel),
			Justification:  pdg.generateDependencyJustification(dep),
		}
		changes = append(changes, change)
	}

	return changes
}

// performRiskAssessment performs comprehensive risk assessment
func (pdg *PRDescriptionGenerator) performRiskAssessment(analysis *ChangeAnalysis, dependencies []*DependencyUpdate) *RiskAssessment {
	riskFactors := []*RiskFactor{}
	mitigations := []string{}
	overallRisk := "low"

	// Assess complexity risk
	if analysis.Complexity.OverallComplexity == "high" {
		riskFactors = append(riskFactors, &RiskFactor{
			Type:        "complexity",
			Description: "High code complexity detected",
			Severity:    "medium",
			Likelihood:  "high",
			Impact:      "medium",
		})
		mitigations = append(mitigations, "Thorough code review recommended")
		overallRisk = "medium"
	}

	// Assess dependency risk
	for _, dep := range dependencies {
		if dep.RiskLevel == "high" || dep.RiskLevel == "critical" {
			riskFactors = append(riskFactors, &RiskFactor{
				Type:        "dependency",
				Description: fmt.Sprintf("High-risk dependency update: %s", dep.Name),
				Severity:    string(dep.RiskLevel),
				Likelihood:  "medium",
				Impact:      "high",
			})
			mitigations = append(mitigations, "Comprehensive testing of affected functionality")
			overallRisk = "high"
		}
	}

	// Assess impact risk
	if analysis.Impact.APIChanges {
		riskFactors = append(riskFactors, &RiskFactor{
			Type:        "api_changes",
			Description: "API changes detected",
			Severity:    "high",
			Likelihood:  "high",
			Impact:      "high",
		})
		mitigations = append(mitigations, "API compatibility testing required")
		overallRisk = "high"
	}

	testingRequired := overallRisk != "low"
	reviewRequired := overallRisk == "high" || overallRisk == "critical"

	return &RiskAssessment{
		OverallRisk:     overallRisk,
		RiskFactors:     riskFactors,
		Mitigations:     mitigations,
		TestingRequired: testingRequired,
		ReviewRequired:  reviewRequired,
		Confidence:      0.8,
	}
}

// generateTestingNotes generates testing guidance
func (pdg *PRDescriptionGenerator) generateTestingNotes(analysis *ChangeAnalysis, riskAssessment *RiskAssessment) string {
	var notes strings.Builder

	if riskAssessment.TestingRequired {
		notes.WriteString("**Testing Required**\n\n")

		if len(analysis.TestChanges) > 0 {
			notes.WriteString("- Review updated test cases\n")
		}

		if analysis.Impact.APIChanges {
			notes.WriteString("- Test API compatibility\n")
		}

		if analysis.Impact.DatabaseChanges {
			notes.WriteString("- Test database migrations\n")
		}

		if riskAssessment.OverallRisk == "high" {
			notes.WriteString("- Perform integration testing\n")
			notes.WriteString("- Consider staging environment testing\n")
		}
	} else {
		notes.WriteString("Standard testing procedures apply.")
	}

	return notes.String()
}

// generateReviewGuidance generates review guidance
func (pdg *PRDescriptionGenerator) generateReviewGuidance(analysis *ChangeAnalysis, riskAssessment *RiskAssessment) string {
	var guidance strings.Builder

	guidance.WriteString("**Review Focus Areas**\n\n")

	if len(analysis.CodeChanges) > 0 {
		guidance.WriteString("- Code quality and maintainability\n")
	}

	if analysis.Complexity.OverallComplexity == "high" {
		guidance.WriteString("- Logic correctness and edge cases\n")
	}

	if riskAssessment.OverallRisk == "high" {
		guidance.WriteString("- Security implications\n")
		guidance.WriteString("- Performance impact\n")
	}

	if analysis.Impact.APIChanges {
		guidance.WriteString("- API design and backward compatibility\n")
	}

	return guidance.String()
}

// identifyBreakingChanges identifies breaking changes
func (pdg *PRDescriptionGenerator) identifyBreakingChanges(analysis *ChangeAnalysis, dependencies []*DependencyUpdate) []*BreakingChange {
	var breakingChanges []*BreakingChange

	// Check for API breaking changes
	if analysis.Impact.APIChanges {
		breakingChanges = append(breakingChanges, &BreakingChange{
			Type:          "api",
			Description:   "API changes detected that may affect existing integrations",
			MigrationPath: "Review API documentation and update client code accordingly",
			AffectedAPIs:  []string{"API clients", "integrations"},
		})
	}

	// Check for dependency breaking changes
	for _, dep := range dependencies {
		if pdg.isMajorVersionChange(dep.CurrentVersion, dep.LatestVersion) {
			breakingChanges = append(breakingChanges, &BreakingChange{
				Type:          "dependency",
				Description:   fmt.Sprintf("Major version update for %s may introduce breaking changes", dep.Name),
				MigrationPath: "Review changelog and update code to handle API changes",
				AffectedAPIs:  []string{fmt.Sprintf("Code using %s", dep.Name)},
			})
		}
	}

	return breakingChanges
}

// generateChecklist generates a PR checklist
func (pdg *PRDescriptionGenerator) generateChecklist(riskAssessment *RiskAssessment, breakingChanges []*BreakingChange) []string {
	checklist := []string{
		"Code follows project style guidelines",
		"Self-review of the code completed",
		"Code is properly commented",
	}

	if riskAssessment.TestingRequired {
		checklist = append(checklist, "Tests added/updated for new functionality")
		checklist = append(checklist, "All tests pass locally")
	}

	if len(breakingChanges) > 0 {
		checklist = append(checklist, "Breaking changes documented")
		checklist = append(checklist, "Migration guide provided if needed")
	}

	if riskAssessment.OverallRisk == "high" {
		checklist = append(checklist, "Security review completed")
		checklist = append(checklist, "Performance impact assessed")
	}

	checklist = append(checklist, "Documentation updated if needed")

	return checklist
}

// Helper methods for analysis

func (pdg *PRDescriptionGenerator) isTestFile(path string) bool {
	return strings.Contains(path, "test") || strings.Contains(path, "spec") ||
		strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, ".test.js")
}

func (pdg *PRDescriptionGenerator) isConfigFile(path string) bool {
	configFiles := []string{"package.json", "requirements.txt", "pom.xml", "build.gradle", "Cargo.toml", "go.mod"}
	for _, config := range configFiles {
		if strings.Contains(path, config) {
			return true
		}
	}
	return false
}

func (pdg *PRDescriptionGenerator) isDocumentationFile(path string) bool {
	return strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".rst") ||
		strings.Contains(path, "doc") || strings.Contains(path, "README")
}

func (pdg *PRDescriptionGenerator) analyzeTestChange(filePatch *FilePatch, analysis *ChangeAnalysis) {
	testChange := &TestChange{
		File:          filePatch.Path,
		Type:          "unit", // Simplified
		Action:        filePatch.Type,
		TestsAdded:    len(filePatch.Changes),
		TestsModified: 0,
		Coverage:      "unknown",
	}
	analysis.TestChanges = append(analysis.TestChanges, testChange)
}

func (pdg *PRDescriptionGenerator) analyzeConfigChange(filePatch *FilePatch, analysis *ChangeAnalysis) {
	configChange := &ConfigChange{
		File:    filePatch.Path,
		Type:    pdg.getConfigType(filePatch.Path),
		Changes: map[string]interface{}{"changes": len(filePatch.Changes)},
		Impact:  "medium",
	}
	analysis.ConfigChanges = append(analysis.ConfigChanges, configChange)
}

func (pdg *PRDescriptionGenerator) analyzeDocumentationChange(filePatch *FilePatch, analysis *ChangeAnalysis) {
	docChange := &DocumentationChange{
		File:        filePatch.Path,
		Type:        "readme",
		Action:      filePatch.Type,
		Description: "Documentation updated",
	}
	analysis.DocumentationChanges = append(analysis.DocumentationChanges, docChange)
}

func (pdg *PRDescriptionGenerator) analyzeCodeChange(filePatch *FilePatch, analysis *ChangeAnalysis) {
	linesAdded := 0
	linesRemoved := 0

	for _, change := range filePatch.Changes {
		if change.Type == "add" {
			linesAdded++
		} else if change.Type == "delete" {
			linesRemoved++
		}
	}

	codeChange := &CodeChange{
		File:         filePatch.Path,
		Type:         "function", // Simplified
		Name:         "unknown",
		Action:       filePatch.Type,
		LinesAdded:   linesAdded,
		LinesRemoved: linesRemoved,
		Complexity:   len(filePatch.Changes),
		Confidence:   filePatch.Confidence,
	}
	analysis.CodeChanges = append(analysis.CodeChanges, codeChange)
}

func (pdg *PRDescriptionGenerator) analyzeComplexity(analysis *ChangeAnalysis) *ComplexityAnalysis {
	totalLines := 0
	for _, change := range analysis.CodeChanges {
		totalLines += change.LinesAdded + change.LinesRemoved
	}

	complexity := "low"
	if totalLines > 100 {
		complexity = "medium"
	}
	if totalLines > 500 {
		complexity = "high"
	}

	return &ComplexityAnalysis{
		OverallComplexity:    complexity,
		CyclomaticComplexity: totalLines / 10, // Simplified
		CognitiveComplexity:  totalLines / 15, // Simplified
		LinesOfCode:          totalLines,
		FilesAffected:        analysis.TotalFiles,
		Score:                float64(totalLines) / 100.0,
	}
}

func (pdg *PRDescriptionGenerator) analyzeImpact(analysis *ChangeAnalysis, dependencies []*DependencyUpdate) *ImpactAnalysis {
	scope := "local"
	if analysis.TotalFiles > 5 {
		scope = "module"
	}
	if analysis.TotalFiles > 20 {
		scope = "system"
	}

	return &ImpactAnalysis{
		Scope:             scope,
		AffectedModules:   analysis.ModifiedFiles,
		UserFacing:        len(analysis.CodeChanges) > 0,
		APIChanges:        pdg.hasAPIChanges(analysis),
		DatabaseChanges:   pdg.hasDatabaseChanges(analysis),
		ConfigChanges:     len(analysis.ConfigChanges) > 0,
		SecurityImpact:    "low",
		PerformanceImpact: "low",
	}
}

func (pdg *PRDescriptionGenerator) determineChangeType(filePatch *FilePatch) string {
	if pdg.isTestFile(filePatch.Path) {
		return "test"
	}
	if pdg.isConfigFile(filePatch.Path) {
		return "config"
	}
	if pdg.isDocumentationFile(filePatch.Path) {
		return "documentation"
	}
	return "feature"
}

func (pdg *PRDescriptionGenerator) generateChangeDescription(filePatch *FilePatch) string {
	return fmt.Sprintf("Updated %s with %d changes", filePatch.Path, len(filePatch.Changes))
}

func (pdg *PRDescriptionGenerator) assessChangeImpact(filePatch *FilePatch) string {
	if len(filePatch.Changes) > 10 {
		return "high"
	}
	if len(filePatch.Changes) > 3 {
		return "medium"
	}
	return "low"
}

func (pdg *PRDescriptionGenerator) determineVersionChangeType(current, new string) string {
	// Simplified version comparison
	if strings.Contains(new, "0.") && !strings.Contains(current, "0.") {
		return "major"
	}
	return "minor"
}

func (pdg *PRDescriptionGenerator) generateDependencyJustification(dep *DependencyUpdate) string {
	return fmt.Sprintf("Updated to address security vulnerabilities and improve compatibility")
}

func (pdg *PRDescriptionGenerator) isMajorVersionChange(current, new string) bool {
	// Simplified major version detection
	return strings.Split(current, ".")[0] != strings.Split(new, ".")[0]
}

func (pdg *PRDescriptionGenerator) getConfigType(path string) string {
	if strings.Contains(path, "package.json") {
		return "npm"
	}
	if strings.Contains(path, "requirements.txt") {
		return "pip"
	}
	return "unknown"
}

func (pdg *PRDescriptionGenerator) hasAPIChanges(analysis *ChangeAnalysis) bool {
	// Simplified API change detection
	for _, change := range analysis.CodeChanges {
		if strings.Contains(change.File, "api") || strings.Contains(change.File, "controller") {
			return true
		}
	}
	return false
}

func (pdg *PRDescriptionGenerator) hasDatabaseChanges(analysis *ChangeAnalysis) bool {
	// Simplified database change detection
	for _, change := range analysis.CodeChanges {
		if strings.Contains(change.File, "migration") || strings.Contains(change.File, "schema") {
			return true
		}
	}
	return false
}
