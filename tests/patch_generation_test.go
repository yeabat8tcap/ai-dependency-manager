package tests

import (
	"context"
	"testing"
	"time"

	"github.com/8tcapital/superint-dep-manager/internal/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestPatchGeneration tests the patch generation functionality
func TestPatchGeneration(t *testing.T) {
	ctx := context.Background()
	mockClient := &github.Client{}
	mockAIManager := &MockAIManager{}

	t.Run("GeneratePatch", func(t *testing.T) {
		patchGenerator := github.NewPatchGenerator(mockAIManager, mockClient)
		
		dependency := &github.DependencyUpdate{
			Name:           "react",
			CurrentVersion: "17.0.0",
			TargetVersion:  "18.0.0",
			PackageManager: "npm",
		}

		breakingChanges := []*github.BreakingChange{
			{
				Type:        "api_change",
				Description: "ReactDOM.render is deprecated",
				Severity:    "high",
				ExampleBefore: `ReactDOM.render(<App />, document.getElementById('root'));`,
				ExampleAfter:  `const root = ReactDOM.createRoot(document.getElementById('root')); root.render(<App />);`,
			},
		}

		// Mock AI manager response
		mockAIManager.On("GeneratePatchSuggestions", dependency, mock.AnythingOfType("[]github.BreakingChange")).Return(
			[]github.PatchSuggestion{
				{
					OldCode:     `ReactDOM.render(<App />, document.getElementById('root'));`,
					NewCode:     `const root = ReactDOM.createRoot(document.getElementById('root')); root.render(<App />);`,
					Description: "Update to new React 18 createRoot API",
					Confidence:  0.9,
				},
			}, nil)

		patches, err := patchGenerator.GeneratePatch(ctx, dependency, breakingChanges)
		
		assert.NoError(t, err)
		assert.NotEmpty(t, patches)
		assert.Equal(t, "react", patches[0].Dependency)
		assert.NotEmpty(t, patches[0].Changes)
		
		mockAIManager.AssertExpectations(t)
	})

	t.Run("GenerateTemplateBasedPatch", func(t *testing.T) {
		templateService := github.NewTemplateService()
		
		// Test React template
		template, err := templateService.GetTemplate("react", "17.0.0", "18.0.0")
		assert.NoError(t, err)
		assert.NotNil(t, template)
		assert.Equal(t, "react", template.PackageName)
		assert.Contains(t, template.Transformations, "ReactDOM.render")

		// Apply template
		patch, err := templateService.ApplyTemplate(template, &github.ProjectInfo{
			Type:           "react",
			PackageManager: "npm",
			Files:          []string{"src/index.js", "package.json"},
		})
		
		assert.NoError(t, err)
		assert.NotNil(t, patch)
		assert.NotEmpty(t, patch.Changes)
	})

	t.Run("RiskAssessment", func(t *testing.T) {
		riskAssessor := github.NewRiskAssessor()
		
		patch := &github.Patch{
			ID:         "patch-1",
			Repository: "test/repo",
			Dependency: "react",
			Changes: []github.Change{
				{
					File:        "src/index.js",
					ChangeType:  "modify",
					Description: "Update ReactDOM.render to createRoot",
				},
			},
		}

		assessment, err := riskAssessor.AssessRisk(patch)
		assert.NoError(t, err)
		assert.NotNil(t, assessment)
		assert.NotEmpty(t, assessment.OverallRisk)
		assert.GreaterOrEqual(t, assessment.Confidence, 0.0)
		assert.LessOrEqual(t, assessment.Confidence, 1.0)
	})
}

// TestPatchValidation tests the patch validation functionality
func TestPatchValidation(t *testing.T) {
	ctx := context.Background()
	mockClient := &github.Client{}

	t.Run("ValidatePatch", func(t *testing.T) {
		validator := github.NewPatchValidator()
		
		patch := &github.Patch{
			ID:          "patch-1",
			Repository:  "test/repo",
			Dependency:  "lodash",
			Description: "Update lodash to fix security vulnerability",
			Changes: []github.Change{
				{
					File:        "package.json",
					OldContent:  `"lodash": "^4.17.20"`,
					NewContent:  `"lodash": "^4.17.21"`,
					ChangeType:  "modify",
					Description: "Update lodash version",
				},
			},
		}

		err := validator.ValidatePatch(patch)
		assert.NoError(t, err)
	})

	t.Run("ValidateChanges", func(t *testing.T) {
		validator := github.NewPatchValidator()
		
		changes := []github.Change{
			{
				File:        "package.json",
				ChangeType:  "modify",
				Description: "Safe dependency update",
			},
		}

		err := validator.ValidateChanges(changes)
		assert.NoError(t, err)

		// Test dangerous change
		dangerousChanges := []github.Change{
			{
				File:        "config.js",
				ChangeType:  "modify",
				Description: "Dangerous configuration change",
			},
		}

		err = validator.ValidateChanges(dangerousChanges)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "dangerous change detected")
	})

	t.Run("ValidateGeneratedPatch", func(t *testing.T) {
		validationService := github.NewValidationService(mockClient)
		
		generatedPatch := &github.GeneratedPatch{
			Repository:  "test/repo",
			Dependency:  "express",
			GeneratedAt: time.Now(),
			Files: []*github.FilePatch{
				{
					Path:        "package.json",
					Type:        "modify",
					Description: "Update Express version",
				},
			},
		}

		options := &github.ValidationOptions{
			SkipBuild: false,
			SkipTests: false,
			Timeout:   5 * time.Minute,
		}

		result, err := validationService.ValidateGeneratedPatch(ctx, generatedPatch, options)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test/repo", result.Repository)
		assert.NotNil(t, result.ValidatedAt)
	})

	t.Run("SafetyValidation", func(t *testing.T) {
		validationService := github.NewValidationService(mockClient)
		
		// Test safe patch
		safePatch := &github.GeneratedPatch{
			Repository: "test/repo",
			Files: []*github.FilePatch{
				{
					Path: "package.json",
					Type: "modify",
				},
			},
		}

		result, err := validationService.ValidatePatchSafety(ctx, safePatch)
		assert.NoError(t, err)
		assert.Equal(t, github.ValidationStatusPassed, result.OverallStatus)
		assert.Empty(t, result.Errors)

		// Test dangerous patch
		dangerousPatch := &github.GeneratedPatch{
			Repository: "test/repo",
			Files: []*github.FilePatch{
				{
					Path:        "config/database.json",
					Type:        "delete",
					Description: "Delete database configuration",
				},
			},
		}

		result, err = validationService.ValidatePatchSafety(ctx, dangerousPatch)
		assert.NoError(t, err)
		assert.Equal(t, github.ValidationStatusFailed, result.OverallStatus)
		assert.NotEmpty(t, result.Errors)
	})
}

// TestPatchApplication tests the patch application functionality
func TestPatchApplication(t *testing.T) {
	ctx := context.Background()
	mockClient := &github.Client{}
	mockAIManager := &MockAIManager{}

	t.Run("ApplyPatches", func(t *testing.T) {
		patchApplicator := github.NewPatchApplicator(mockClient, mockAIManager)
		
		patches := []*github.Patch{
			{
				ID:         "patch-1",
				Repository: "test/repo",
				Dependency: "lodash",
				Changes: []github.Change{
					{
						File:       "package.json",
						OldContent: `"lodash": "^4.17.20"`,
						NewContent: `"lodash": "^4.17.21"`,
						ChangeType: "modify",
					},
				},
			},
		}

		request := &github.ApplicationRequest{
			Repository: "test/repo",
			Branch:     "main",
			Strategy:   github.StrategySequential,
			Patches:    patches,
		}

		result, err := patchApplicator.ApplyPatches(ctx, request)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test/repo", result.Repository)
		assert.NotEmpty(t, result.AppliedPatches)
	})

	t.Run("ConflictDetection", func(t *testing.T) {
		patchApplicator := github.NewPatchApplicator(mockClient, mockAIManager)
		
		patch := &github.Patch{
			ID: "patch-1",
			Changes: []github.Change{
				{
					File:       "package.json",
					ChangeType: "modify",
					OldContent: `"react": "^17.0.0"`,
					NewContent: `"react": "^18.0.0"`,
				},
			},
		}

		conflicts, err := patchApplicator.DetectConflicts(patch, "test/repo", "main")
		assert.NoError(t, err)
		assert.NotNil(t, conflicts)
	})

	t.Run("RollbackManagement", func(t *testing.T) {
		rollbackManager := github.NewRollbackManager()
		
		applicationResult := &github.ApplicationResult{
			Repository: "test/repo",
			Branch:     "main",
			AppliedPatches: []*github.AppliedPatch{
				{
					PatchID:    "patch-1",
					Status:     "success",
					AppliedAt:  time.Now(),
					Confidence: 0.9,
				},
			},
		}

		rollbackPlan, err := rollbackManager.CreateRollbackPlan(applicationResult)
		assert.NoError(t, err)
		assert.NotNil(t, rollbackPlan)
		assert.Equal(t, "test/repo", rollbackPlan.Repository)
		assert.NotEmpty(t, rollbackPlan.Steps)

		// Test rollback execution
		rollbackResult, err := rollbackManager.ExecuteRollback(ctx, rollbackPlan)
		assert.NoError(t, err)
		assert.NotNil(t, rollbackResult)
		assert.Equal(t, "success", rollbackResult.Status)
	})
}

// TestConflictResolution tests the conflict resolution functionality
func TestConflictResolution(t *testing.T) {
	ctx := context.Background()
	mockAIManager := &MockAIManager{}

	t.Run("ResolveConflicts", func(t *testing.T) {
		conflictResolver := github.NewConflictResolver(mockAIManager)
		
		conflicts := []*github.Conflict{
			{
				File:     "package.json",
				Type:     github.ConflictTypeContent,
				Line:     10,
				Current:  `"lodash": "^4.17.20"`,
				Incoming: `"lodash": "^4.17.21"`,
				Severity: github.SeverityLow,
			},
		}

		// Mock AI manager response
		mockAIManager.On("ResolveConflicts", conflicts).Return(
			[]*github.ConflictResolution{
				{
					ConflictID:   "conflict-1",
					Resolution:   "accept_incoming",
					ResolvedCode: `"lodash": "^4.17.21"`,
					Confidence:   0.95,
					Reasoning:    "Security update should be accepted",
				},
			}, nil)

		resolutions, err := conflictResolver.ResolveConflicts(ctx, conflicts)
		assert.NoError(t, err)
		assert.NotEmpty(t, resolutions)
		assert.Equal(t, "accept_incoming", resolutions[0].Resolution)
		assert.GreaterOrEqual(t, resolutions[0].Confidence, 0.9)
		
		mockAIManager.AssertExpectations(t)
	})

	t.Run("ConflictAnalysis", func(t *testing.T) {
		conflictAnalyzer := github.NewConflictAnalyzer()
		
		conflicts := []*github.Conflict{
			{Type: github.ConflictTypeContent, Severity: github.SeverityHigh},
			{Type: github.ConflictTypeStructural, Severity: github.SeverityMedium},
			{Type: github.ConflictTypeContent, Severity: github.SeverityLow},
		}

		analysis, err := conflictAnalyzer.AnalyzeConflicts(conflicts)
		assert.NoError(t, err)
		assert.NotNil(t, analysis)
		assert.Equal(t, 3, analysis.TotalConflicts)
		assert.Equal(t, 1, analysis.HighSeverityCount)
		assert.Equal(t, 1, analysis.MediumSeverityCount)
		assert.Equal(t, 1, analysis.LowSeverityCount)
	})
}

// MockAIManager implements the AIManager interface for testing
type MockAIManager struct {
	mock.Mock
}

func (m *MockAIManager) AnalyzeDependencyUpdate(dependency *github.DependencyUpdate) (*github.BreakingChangeAnalysis, error) {
	args := m.Called(dependency)
	return args.Get(0).(*github.BreakingChangeAnalysis), args.Error(1)
}

func (m *MockAIManager) GeneratePatchSuggestions(dependency *github.DependencyUpdate, breakingChanges []github.BreakingChange) ([]github.PatchSuggestion, error) {
	args := m.Called(dependency, breakingChanges)
	return args.Get(0).([]github.PatchSuggestion), args.Error(1)
}

func (m *MockAIManager) GeneratePRDescription(patches []*github.Patch) (string, error) {
	args := m.Called(patches)
	return args.String(0), args.Error(1)
}

func (m *MockAIManager) SelectReviewers(repository string, patches []*github.Patch) ([]string, error) {
	args := m.Called(repository, patches)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockAIManager) ResolveConflicts(conflicts []*github.Conflict) ([]*github.ConflictResolution, error) {
	args := m.Called(conflicts)
	return args.Get(0).([]*github.ConflictResolution), args.Error(1)
}

// BenchmarkPatchGeneration benchmarks patch generation performance
func BenchmarkPatchGeneration(b *testing.B) {
	ctx := context.Background()
	mockClient := &github.Client{}
	mockAIManager := &MockAIManager{}
	patchGenerator := github.NewPatchGenerator(mockAIManager, mockClient)

	dependency := &github.DependencyUpdate{
		Name:           "lodash",
		CurrentVersion: "4.17.20",
		LatestVersion:  "4.17.21",
		PackageManager: "npm",
	}

	breakingChanges := []*github.BreakingChange{
		{
			Type:        "security_fix",
			Description: "Fix prototype pollution vulnerability",
			Severity:    "high",
		},
	}

	// Mock AI manager response
	mockAIManager.On("GeneratePatchSuggestions", mock.Anything, mock.Anything).Return(
		[]github.PatchSuggestion{
			{
				OldCode:     `_.merge(target, source)`,
				NewCode:     `_.mergeWith(target, source, customizer)`,
				Description: "Use mergeWith for safer merging",
				Confidence:  0.8,
			},
		}, nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := patchGenerator.GeneratePatch(ctx, dependency, breakingChanges)
		if err != nil {
			b.Fatal(err)
		}
	}
}
