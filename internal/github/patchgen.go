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

// PatchGenerator handles AI-powered patch generation
type PatchGenerator struct {
	aiManager       AIManager
	templateService *TemplateService
	parser          *ParsingService
	analyzer        *AnalysisService
	client          *Client
	repositories    *RepositoriesService
}

// NewPatchGenerator creates a new patch generator
func NewPatchGenerator(aiManager *ai.Manager, client *Client) *PatchGenerator {
	return &PatchGenerator{
		aiManager:       aiManager,
		templateService: NewTemplateService(),
		parser:          NewParsingService(client),
		analyzer:        NewAnalysisService(client, aiManager),
		client:          client,
		repositories:    client.Repositories,
	}
}

// GeneratedPatch represents a generated patch
type GeneratedPatch struct {
	Repository      string              `json:"repository"`
	Dependencies    []*DependencyUpdate `json:"dependencies"`
	Files           []*FilePatch        `json:"files"`
	ConfigChanges   []*ConfigPatch      `json:"config_changes"`
	CommitMessage   string              `json:"commit_message"`
	PRTitle         string              `json:"pr_title"`
	PRDescription   string              `json:"pr_description"`
	ValidationSteps []*ValidationStep   `json:"validation_steps"`
	RiskAssessment  *RiskAssessment     `json:"risk_assessment"`
	GeneratedAt     time.Time           `json:"generated_at"`
	GeneratedBy     string              `json:"generated_by"` // "ai", "template", "heuristic"
}

// FilePatch represents changes to a specific file
type FilePatch struct {
	Path        string       `json:"path"`
	Type        string       `json:"type"`        // "modify", "create", "delete"
	Language    string       `json:"language"`
	Changes     []*Change    `json:"changes"`
	NewContent  string       `json:"new_content,omitempty"`
	Confidence  float64      `json:"confidence"`
	Description string       `json:"description"`
}

// Change is now defined in shared_types.go
// ConfigPatch represents changes to configuration files
type ConfigPatch struct {
	File        string                 `json:"file"`
	Type        string                 `json:"type"`        // "package.json", "requirements.txt", etc.
	Changes     map[string]interface{} `json:"changes"`
	Description string                 `json:"description"`
}

// ValidationStep represents a step to validate the patch
type ValidationStep struct {
	Type        string `json:"type"`        // "build", "test", "lint", "manual"
	Command     string `json:"command,omitempty"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// RiskAssessment is now defined in shared_types.go

// PatchGenerationRequest represents a request to generate a patch
type PatchGenerationRequest struct {
	Repository      string              `json:"repository"`
	Dependencies    []*DependencyUpdate `json:"dependencies"`
	BaseBranch      string              `json:"base_branch"`
	ProjectStructure *ProjectStructure  `json:"project_structure,omitempty"`
	Analysis        []*BreakingChangeAnalysis `json:"analysis,omitempty"`
	Options         *PatchOptions       `json:"options,omitempty"`
}

// PatchOptions represents options for patch generation
type PatchOptions struct {
	UseAI           bool     `json:"use_ai"`
	UseTemplates    bool     `json:"use_templates"`
	IncludeTests    bool     `json:"include_tests"`
	UpdateLockFiles bool     `json:"update_lock_files"`
	ValidateChanges bool     `json:"validate_changes"`
	Languages       []string `json:"languages,omitempty"`
}

// GeneratePatch generates a comprehensive patch for dependency updates
func (pg *PatchGenerator) GeneratePatch(ctx context.Context, request *PatchGenerationRequest) (*GeneratedPatch, error) {
	logger.Info("Generating patch for repository %s with %d dependencies", 
		request.Repository, len(request.Dependencies))
	
	// Parse repository name
	parts := strings.Split(request.Repository, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format: %s", request.Repository)
	}
	owner, repo := parts[0], parts[1]
	
	// Get project structure if not provided
	if request.ProjectStructure == nil {
		structure, err := pg.parser.ParseProject(ctx, owner, repo)
		if err != nil {
			return nil, fmt.Errorf("failed to parse project structure: %w", err)
		}
		request.ProjectStructure = structure
	}
	
	// Perform analysis if not provided
	if request.Analysis == nil {
		analysis, err := pg.analyzer.AnalyzeProject(ctx, owner, repo, request.Dependencies)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze project: %w", err)
		}
		
		// Convert project analysis to breaking change analysis
		request.Analysis = analysis.Dependencies
	}
	
	// Set default options
	if request.Options == nil {
		request.Options = &PatchOptions{
			UseAI:           true,
			UseTemplates:    true,
			IncludeTests:    true,
			UpdateLockFiles: true,
			ValidateChanges: true,
		}
	}
	
	patch := &GeneratedPatch{
		Repository:      request.Repository,
		Dependencies:    request.Dependencies,
		Files:           []*FilePatch{},
		ConfigChanges:   []*ConfigPatch{},
		ValidationSteps: []*ValidationStep{},
		GeneratedAt:     time.Now(),
	}
	
	// Generate configuration file patches
	configPatches, err := pg.generateConfigPatches(request)
	if err != nil {
		logger.Warn("Failed to generate config patches: %v", err)
	} else {
		patch.ConfigChanges = configPatches
	}
	
	// Generate code patches
	if request.Options.UseAI && pg.aiManager != nil {
		aiPatches, err := pg.generateAIPatches(ctx, request)
		if err != nil {
			logger.Warn("AI patch generation failed: %v", err)
		} else {
			patch.Files = append(patch.Files, aiPatches...)
			patch.GeneratedBy = "ai"
		}
	}
	
	// Generate template-based patches
	if request.Options.UseTemplates {
		templatePatches, err := pg.generateTemplatePatches(request)
		if err != nil {
			logger.Warn("Template patch generation failed: %v", err)
		} else {
			patch.Files = append(patch.Files, templatePatches...)
			if patch.GeneratedBy == "" {
				patch.GeneratedBy = "template"
			} else {
				patch.GeneratedBy += "+template"
			}
		}
	}
	
	// Generate validation steps
	patch.ValidationSteps = pg.generateValidationSteps(request)
	
	// Assess risk
	patch.RiskAssessment = pg.assessPatchRisk(patch, request.Analysis)
	
	// Generate commit message and PR details
	patch.CommitMessage = pg.generateCommitMessage(request.Dependencies)
	patch.PRTitle = pg.generatePRTitle(request.Dependencies)
	patch.PRDescription = pg.generatePRDescription(patch, request.Analysis)
	
	logger.Info("Generated patch for %s: %d file changes, %d config changes, risk level %s", 
		request.Repository, len(patch.Files), len(patch.ConfigChanges), patch.RiskAssessment.OverallRisk)
	
	return patch, nil
}

// generateConfigPatches generates patches for configuration files
func (pg *PatchGenerator) generateConfigPatches(request *PatchGenerationRequest) ([]*ConfigPatch, error) {
	var patches []*ConfigPatch
	
	// Update package.json dependencies
	if request.ProjectStructure.ProjectType == "nodejs" {
		patch := &ConfigPatch{
			File:        "package.json",
			Type:        "package.json",
			Changes:     make(map[string]interface{}),
			Description: "Update dependency versions in package.json",
		}
		
		dependencies := make(map[string]string)
		for _, dep := range request.Dependencies {
			dependencies[dep.Name] = dep.LatestVersion
		}
		patch.Changes["dependencies"] = dependencies
		
		patches = append(patches, patch)
	}
	
	// Update requirements.txt
	if request.ProjectStructure.ProjectType == "python" {
		var newRequirements []string
		for _, dep := range request.Dependencies {
			newRequirements = append(newRequirements, fmt.Sprintf("%s==%s", dep.Name, dep.LatestVersion))
		}
		
		patch := &ConfigPatch{
			File:        "requirements.txt",
			Type:        "requirements.txt",
			Changes:     map[string]interface{}{"requirements": newRequirements},
			Description: "Update dependency versions in requirements.txt",
		}
		patches = append(patches, patch)
	}
	
	// Update pom.xml for Maven projects
	if request.ProjectStructure.PackageManager == "maven" {
		patch := &ConfigPatch{
			File:        "pom.xml",
			Type:        "pom.xml",
			Changes:     make(map[string]interface{}),
			Description: "Update dependency versions in pom.xml",
		}
		
		var dependencies []map[string]string
		for _, dep := range request.Dependencies {
			parts := strings.Split(dep.Name, ":")
			if len(parts) == 2 {
				dependencies = append(dependencies, map[string]string{
					"groupId":    parts[0],
					"artifactId": parts[1],
					"version":    dep.LatestVersion,
				})
			}
		}
		patch.Changes["dependencies"] = dependencies
		
		patches = append(patches, patch)
	}
	
	return patches, nil
}

// generateAIPatches generates patches using AI
func (pg *PatchGenerator) generateAIPatches(ctx context.Context, request *PatchGenerationRequest) ([]*FilePatch, error) {
	var patches []*FilePatch
	
	for _, analysis := range request.Analysis {
		if !analysis.HasBreakingChanges {
			continue
		}
		
		// Create AI request for patch generation
		aiRequest := &models.PatchGenerationRequest{
			PackageName:     analysis.Dependency.Name,
			CurrentVersion:  analysis.Dependency.CurrentVersion,
			TargetVersion:   analysis.Dependency.LatestVersion,
			BreakingChanges: convertBreakingChanges(analysis.BreakingChanges),
			ProjectType:     request.ProjectStructure.ProjectType,
			AffectedFiles:   analysis.Dependency.AffectedFiles,
		}
		
		// Get AI-generated patches
		aiPatches, err := pg.aiManager.GeneratePatches(ctx, aiRequest)
		if err != nil {
			logger.Warn("AI patch generation failed for %s: %v", analysis.Dependency.Name, err)
			continue
		}
		
		// Convert AI patches to our format
		for _, aiPatch := range aiPatches {
			patch := &FilePatch{
				Path:        aiPatch.FilePath,
				Type:        "modify",
				Language:    detectLanguage(aiPatch.FilePath),
				Changes:     []*Change{},
				Confidence:  aiPatch.Confidence,
				Description: aiPatch.Description,
			}
			
			// Convert AI changes
			for _, change := range aiPatch.Changes {
				patch.Changes = append(patch.Changes, &Change{
					LineStart:  change.LineStart,
					LineEnd:    change.LineEnd,
					OldContent: change.OldContent,
					NewContent: change.NewContent,
					Type:       change.Type,
					Reason:     change.Reason,
					Confidence: change.Confidence,
				})
			}
			
			patches = append(patches, patch)
		}
	}
	
	return patches, nil
}

// generateTemplatePatches generates patches using templates
func (pg *PatchGenerator) generateTemplatePatches(request *PatchGenerationRequest) ([]*FilePatch, error) {
	var patches []*FilePatch
	
	for _, analysis := range request.Analysis {
		templatePatches := pg.getTemplatePatches(analysis.Dependency, request.ProjectStructure.ProjectType)
		patches = append(patches, templatePatches...)
	}
	
	return patches, nil
}

// getTemplatePatches gets template-based patches for common scenarios
func (pg *PatchGenerator) getTemplatePatches(dependency *DependencyUpdate, projectType string) []*FilePatch {
	var patches []*FilePatch
	
	// Express.js template patches
	if dependency.Name == "express" && projectType == "nodejs" {
		patches = append(patches, &FilePatch{
			Path:        "src/app.js",
			Type:        "modify",
			Language:    "javascript",
			Confidence:  0.8,
			Description: "Update Express.js middleware configuration",
			Changes: []*Change{
				{
					OldContent: "app.use(bodyParser.json())",
					NewContent: "app.use(express.json())",
					Type:       "replace",
					Reason:     "bodyParser is now built into Express",
					Confidence: 0.9,
				},
			},
		})
	}
	
	// React template patches
	if dependency.Name == "react" && projectType == "nodejs" {
		patches = append(patches, &FilePatch{
			Path:        "src/components/App.js",
			Type:        "modify",
			Language:    "javascript",
			Confidence:  0.7,
			Description: "Update React component lifecycle methods",
			Changes: []*Change{
				{
					OldContent: "componentWillMount()",
					NewContent: "componentDidMount()",
					Type:       "replace",
					Reason:     "componentWillMount is deprecated",
					Confidence: 0.8,
				},
			},
		})
	}
	
	return patches
}

// generateValidationSteps generates validation steps for the patch
func (pg *PatchGenerator) generateValidationSteps(request *PatchGenerationRequest) []*ValidationStep {
	var steps []*ValidationStep
	
	// Build validation
	switch request.ProjectStructure.ProjectType {
	case "nodejs":
		steps = append(steps, &ValidationStep{
			Type:        "build",
			Command:     "npm run build",
			Description: "Build the project to check for compilation errors",
			Required:    true,
		})
		steps = append(steps, &ValidationStep{
			Type:        "test",
			Command:     "npm test",
			Description: "Run tests to ensure functionality is preserved",
			Required:    true,
		})
	case "python":
		steps = append(steps, &ValidationStep{
			Type:        "test",
			Command:     "python -m pytest",
			Description: "Run Python tests",
			Required:    true,
		})
	case "java":
		steps = append(steps, &ValidationStep{
			Type:        "build",
			Command:     "mvn compile",
			Description: "Compile Java project",
			Required:    true,
		})
		steps = append(steps, &ValidationStep{
			Type:        "test",
			Command:     "mvn test",
			Description: "Run Java tests",
			Required:    true,
		})
	}
	
	// Lint validation
	steps = append(steps, &ValidationStep{
		Type:        "lint",
		Command:     "npm run lint",
		Description: "Run linting to check code quality",
		Required:    false,
	})
	
	// Manual validation
	steps = append(steps, &ValidationStep{
		Type:        "manual",
		Description: "Manually test critical functionality",
		Required:    true,
	})
	
	return steps
}

// assessPatchRisk assesses the risk of applying the patch
func (pg *PatchGenerator) assessPatchRisk(patch *GeneratedPatch, analyses []*BreakingChangeAnalysis) *RiskAssessment {
	assessment := &RiskAssessment{
		OverallRisk:     models.RiskLow,
		FilesModified:   len(patch.Files),
		Reversible:      true,
		Recommendations: []string{},
	}
	
	// Count breaking changes
	for _, analysis := range analyses {
		if analysis.HasBreakingChanges {
			assessment.BreakingChanges += len(analysis.BreakingChanges)
		}
		
		// Use highest risk level
		if analysis.RiskLevel > assessment.OverallRisk {
			assessment.OverallRisk = analysis.RiskLevel
		}
	}
	
	// Assess based on files modified
	if assessment.FilesModified > 10 {
		assessment.OverallRisk = models.RiskHigh
		assessment.Recommendations = append(assessment.Recommendations, 
			"Large number of files modified - consider breaking into smaller patches")
	}
	
	// Assess based on breaking changes
	if assessment.BreakingChanges > 5 {
		assessment.OverallRisk = models.RiskHigh
		assessment.Recommendations = append(assessment.Recommendations, 
			"Multiple breaking changes detected - thorough testing required")
	}
	
	return assessment
}

// generateCommitMessage generates a commit message for the patch
func (pg *PatchGenerator) generateCommitMessage(dependencies []*DependencyUpdate) string {
	if len(dependencies) == 1 {
		dep := dependencies[0]
		return fmt.Sprintf("Update %s from %s to %s", dep.Name, dep.CurrentVersion, dep.LatestVersion)
	}
	
	return fmt.Sprintf("Update %d dependencies", len(dependencies))
}

// generatePRTitle generates a pull request title
func (pg *PatchGenerator) generatePRTitle(dependencies []*DependencyUpdate) string {
	if len(dependencies) == 1 {
		dep := dependencies[0]
		updateType := "patch"
		if dep.BreakingChange {
			updateType = "major"
		}
		return fmt.Sprintf("chore: %s update %s to %s", updateType, dep.Name, dep.LatestVersion)
	}
	
	return fmt.Sprintf("chore: update %d dependencies", len(dependencies))
}

// generatePRDescription generates a pull request description
func (pg *PatchGenerator) generatePRDescription(patch *GeneratedPatch, analyses []*BreakingChangeAnalysis) string {
	var description strings.Builder
	
	description.WriteString("## Dependency Updates\n\n")
	description.WriteString("This PR updates the following dependencies:\n\n")
	
	for _, dep := range patch.Dependencies {
		description.WriteString(fmt.Sprintf("- **%s**: %s → %s", dep.Name, dep.CurrentVersion, dep.LatestVersion))
		if dep.SecurityFix {
			description.WriteString(" 🔒 *Security Fix*")
		}
		if dep.BreakingChange {
			description.WriteString(" ⚠️ *Breaking Change*")
		}
		description.WriteString("\n")
	}
	
	// Add breaking changes section
	hasBreakingChanges := false
	for _, analysis := range analyses {
		if analysis.HasBreakingChanges {
			hasBreakingChanges = true
			break
		}
	}
	
	if hasBreakingChanges {
		description.WriteString("\n## Breaking Changes\n\n")
		for _, analysis := range analyses {
			if analysis.HasBreakingChanges {
				description.WriteString(fmt.Sprintf("### %s\n", analysis.Dependency.Name))
				for _, bc := range analysis.BreakingChanges {
					description.WriteString(fmt.Sprintf("- **%s**: %s\n", bc.Type, bc.Description))
				}
				description.WriteString("\n")
			}
		}
	}
	
	// Add changes section
	if len(patch.Files) > 0 {
		description.WriteString("## Changes Made\n\n")
		for _, file := range patch.Files {
			description.WriteString(fmt.Sprintf("- **%s**: %s\n", file.Path, file.Description))
		}
		description.WriteString("\n")
	}
	
	// Add validation section
	if len(patch.ValidationSteps) > 0 {
		description.WriteString("## Validation\n\n")
		description.WriteString("The following validation steps should be performed:\n\n")
		for _, step := range patch.ValidationSteps {
			required := ""
			if step.Required {
				required = " (Required)"
			}
			description.WriteString(fmt.Sprintf("- [ ] **%s**%s: %s", step.Type, required, step.Description))
			if step.Command != "" {
				description.WriteString(fmt.Sprintf(" (`%s`)", step.Command))
			}
			description.WriteString("\n")
		}
	}
	
	// Add risk assessment
	description.WriteString(fmt.Sprintf("\n## Risk Assessment\n\n"))
	description.WriteString(fmt.Sprintf("- **Overall Risk**: %s\n", patch.RiskAssessment.OverallRisk))
	description.WriteString(fmt.Sprintf("- **Files Modified**: %d\n", patch.RiskAssessment.FilesModified))
	description.WriteString(fmt.Sprintf("- **Breaking Changes**: %d\n", patch.RiskAssessment.BreakingChanges))
	description.WriteString(fmt.Sprintf("- **Reversible**: %t\n", patch.RiskAssessment.Reversible))
	
	if len(patch.RiskAssessment.Recommendations) > 0 {
		description.WriteString("\n### Recommendations\n\n")
		for _, rec := range patch.RiskAssessment.Recommendations {
			description.WriteString(fmt.Sprintf("- %s\n", rec))
		}
	}
	
	description.WriteString("\n---\n*This PR was generated automatically by AI Dependency Manager*")
	
	return description.String()
}

// Helper functions

func convertBreakingChanges(breakingChanges []*BreakingChange) []*models.BreakingChange {
	var converted []*models.BreakingChange
	for _, bc := range breakingChanges {
		converted = append(converted, &models.BreakingChange{
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
	return converted
}

func detectLanguage(filePath string) string {
	if strings.HasSuffix(filePath, ".js") || strings.HasSuffix(filePath, ".jsx") {
		return "javascript"
	}
	if strings.HasSuffix(filePath, ".ts") || strings.HasSuffix(filePath, ".tsx") {
		return "typescript"
	}
	if strings.HasSuffix(filePath, ".py") {
		return "python"
	}
	if strings.HasSuffix(filePath, ".java") {
		return "java"
	}
	if strings.HasSuffix(filePath, ".rs") {
		return "rust"
	}
	if strings.HasSuffix(filePath, ".go") {
		return "go"
	}
	return "text"
}
