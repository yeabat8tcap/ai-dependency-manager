package github

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/ai"
)

// PatchApplicator handles intelligent patch application with conflict resolution
type PatchApplicator struct {
	client          *Client
	aiManager       AIManager
	validator       PatchValidator
	conflictResolver *ConflictResolver
	rollbackManager *RollbackManager
}

// NewPatchApplicator creates a new patch applicator
func NewPatchApplicator(client *Client, aiManager *ai.Manager) *PatchApplicator {
	validator := NewPatchValidator()
	conflictResolver := NewConflictResolver(aiManager)
	rollbackManager := NewRollbackManager()
	
	return &PatchApplicator{
		client:          client,
		aiManager:       aiManager,
		validator:       validator,
		conflictResolver: conflictResolver,
		rollbackManager: rollbackManager,
	}
}

// ApplicationRequest represents a patch application request
type ApplicationRequest struct {
	Repository    string                 `json:"repository"`
	Branch        string                 `json:"branch"`
	Patches           []*Patch      `json:"patches"`
	Strategy      ApplicationStrategy    `json:"strategy"`
	Options       *ApplicationOptions    `json:"options"`
	Context       map[string]interface{} `json:"context"`
}

// ApplicationStrategy defines how patches should be applied
type ApplicationStrategy string

const (
	StrategySequential ApplicationStrategy = "sequential"
	StrategyParallel   ApplicationStrategy = "parallel"
	StrategyOptimized  ApplicationStrategy = "optimized"
	StrategyConservative ApplicationStrategy = "conservative"
)

// ApplicationOptions configures patch application behavior
type ApplicationOptions struct {
	DryRun              bool                    `json:"dry_run"`
	AutoResolveConflicts bool                   `json:"auto_resolve_conflicts"`
	CreateBackup        bool                   `json:"create_backup"`
	ValidateAfterApply  bool                   `json:"validate_after_apply"`
	RollbackOnFailure   bool                   `json:"rollback_on_failure"`
	MaxRetries          int                    `json:"max_retries"`
	RetryDelay          time.Duration          `json:"retry_delay"`
	ConflictResolution  ConflictResolutionMode `json:"conflict_resolution"`
	NotifyOnCompletion  bool                   `json:"notify_on_completion"`
}

// ConflictResolutionMode defines how conflicts should be resolved
type ConflictResolutionMode string

const (
	ConflictResolveAuto   ConflictResolutionMode = "auto"
	ConflictResolveManual ConflictResolutionMode = "manual"
	ConflictResolveAI     ConflictResolutionMode = "ai"
	ConflictResolveAbort  ConflictResolutionMode = "abort"
)

// ApplicationResult represents the result of patch application
type ApplicationResult struct {
	Success         bool                     `json:"success"`
	AppliedPatches  []*AppliedPatch         `json:"applied_patches"`
	FailedPatches   []*FailedPatch          `json:"failed_patches"`
	Conflicts       []*Conflict             `json:"conflicts"`
	ValidationResults []*ValidationResult   `json:"validation_results"`
	RollbackPoint   *RollbackPoint          `json:"rollback_point,omitempty"`
	Duration        time.Duration           `json:"duration"`
	Summary         *ApplicationSummary     `json:"summary"`
	Recommendations []*Recommendation       `json:"recommendations"`
}

// AppliedPatch represents a successfully applied patch
type AppliedPatch struct {
	Patch       *Patch        `json:"patch"`
	AppliedAt   time.Time     `json:"applied_at"`
	Duration    time.Duration `json:"duration"`
	Conflicts   []*Conflict   `json:"conflicts"`
	Resolution  string        `json:"resolution"`
	Confidence  float64       `json:"confidence"`
}

// FailedPatch represents a patch that failed to apply
type FailedPatch struct {
	Patch     *Patch    `json:"patch"`
	Error     string    `json:"error"`
	Reason    string    `json:"reason"`
	FailedAt  time.Time `json:"failed_at"`
	Retries   int       `json:"retries"`
	Conflicts []*Conflict `json:"conflicts"`
}

// Conflict represents a merge conflict during patch application
type Conflict struct {
	File        string              `json:"file"`
	Type        ConflictType        `json:"type"`
	Line        int                 `json:"line"`
	Current     string              `json:"current"`
	Incoming    string              `json:"incoming"`
	Context     string              `json:"context"`
	Severity    ConflictSeverity    `json:"severity"`
	Resolution  *ConflictResolution `json:"resolution,omitempty"`
	Confidence  float64             `json:"confidence"`
}

// ConflictType defines the type of conflict
type ConflictType string

const (
	ConflictTypeContent    ConflictType = "content"
	ConflictTypeStructural ConflictType = "structural"
	ConflictTypeSemantic   ConflictType = "semantic"
	ConflictTypeSyntactic  ConflictType = "syntactic"
)

// ConflictSeverity defines the severity of a conflict
type ConflictSeverity string

const (
	SeverityLow      ConflictSeverity = "low"
	SeverityMedium   ConflictSeverity = "medium"
	SeverityHigh     ConflictSeverity = "high"
	SeverityCritical ConflictSeverity = "critical"
)

// ConflictResolution represents a resolved conflict
type ConflictResolution struct {
	Strategy    string    `json:"strategy"`
	Resolution  string    `json:"resolution"`
	Reasoning   string    `json:"reasoning"`
	ResolvedBy  string    `json:"resolved_by"`
	ResolvedAt  time.Time `json:"resolved_at"`
	Confidence  float64   `json:"confidence"`
}

// ApplicationSummary provides a summary of the application process
type ApplicationSummary struct {
	TotalPatches      int           `json:"total_patches"`
	AppliedPatches    int           `json:"applied_patches"`
	FailedPatches     int           `json:"failed_patches"`
	ConflictsFound    int           `json:"conflicts_found"`
	ConflictsResolved int           `json:"conflicts_resolved"`
	Duration          time.Duration `json:"duration"`
	SuccessRate       float64       `json:"success_rate"`
}

// ApplyPatches applies a set of patches with intelligent conflict resolution
func (pa *PatchApplicator) ApplyPatches(ctx context.Context, request *ApplicationRequest) (*ApplicationResult, error) {
	startTime := time.Now()
	
	// Initialize result
	result := &ApplicationResult{
		AppliedPatches:    []*AppliedPatch{},
		FailedPatches:     []*FailedPatch{},
		Conflicts:         []*Conflict{},
		ValidationResults: []*ValidationResult{},
		Recommendations:   []*Recommendation{},
	}
	
	// Create rollback point if requested
	if request.Options.CreateBackup {
		rollbackPoint, err := pa.rollbackManager.CreateRollbackPoint(request.Repository, request.Branch)
		if err != nil {
			return nil, fmt.Errorf("failed to create rollback point: %w", err)
		}
		result.RollbackPoint = rollbackPoint
	}
	
	// Apply patches based on strategy
	switch request.Strategy {
	case StrategySequential:
		err := pa.applySequential(ctx, request, result)
		if err != nil {
			return result, err
		}
	case StrategyParallel:
		err := pa.applyParallel(ctx, request, result)
		if err != nil {
			return result, err
		}
	case StrategyOptimized:
		err := pa.applyOptimized(ctx, request, result)
		if err != nil {
			return result, err
		}
	case StrategyConservative:
		err := pa.applyConservative(ctx, request, result)
		if err != nil {
			return result, err
		}
	default:
		return nil, fmt.Errorf("unsupported application strategy: %s", request.Strategy)
	}
	
	// Validate results if requested
	if request.Options.ValidateAfterApply {
		validationResults, err := pa.validateApplication(ctx, request, result)
		if err != nil {
			return result, fmt.Errorf("validation failed: %w", err)
		}
		result.ValidationResults = validationResults
	}
	
	// Calculate summary
	result.Duration = time.Since(startTime)
	result.Summary = pa.calculateSummary(result)
	result.Success = result.Summary.SuccessRate > 0.8 // 80% success threshold
	
	// Generate recommendations
	result.Recommendations = pa.generateRecommendations(result)
	
	// Rollback on failure if requested
	if !result.Success && request.Options.RollbackOnFailure && result.RollbackPoint != nil {
		rollbackErr := pa.rollbackManager.Rollback(result.RollbackPoint)
		if rollbackErr != nil {
			return result, fmt.Errorf("patch application failed and rollback failed: %w", rollbackErr)
		}
	}
	
	return result, nil
}

// applySequential applies patches one by one in sequence
func (pa *PatchApplicator) applySequential(ctx context.Context, request *ApplicationRequest, result *ApplicationResult) error {
	for i, patch := range request.Patches {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		appliedPatch, err := pa.applySinglePatch(ctx, request, patch, i)
		if err != nil {
			failedPatch := &FailedPatch{
				Patch:    patch,
				Error:    err.Error(),
				Reason:   "application_failed",
				FailedAt: time.Now(),
			}
			result.FailedPatches = append(result.FailedPatches, failedPatch)
			
			if !request.Options.RollbackOnFailure {
				continue // Continue with next patch
			}
			return err
		}
		
		result.AppliedPatches = append(result.AppliedPatches, appliedPatch)
	}
	
	return nil
}

// applyParallel applies patches in parallel (for non-conflicting patches)
func (pa *PatchApplicator) applyParallel(ctx context.Context, request *ApplicationRequest, result *ApplicationResult) error {
	// Analyze patch dependencies first
	groups := pa.groupNonConflictingPatches(request.Patches)
	
	for _, group := range groups {
		// Apply each group sequentially, but patches within group in parallel
		err := pa.applyPatchGroup(ctx, request, group, result)
		if err != nil {
			return err
		}
	}
	
	return nil
}

// applyOptimized applies patches using an optimized strategy
func (pa *PatchApplicator) applyOptimized(ctx context.Context, request *ApplicationRequest, result *ApplicationResult) error {
	// Sort patches by confidence and impact
	optimizedOrder := pa.optimizePatchOrder(request.Patches)
	
	// Create a new request with optimized order
	optimizedRequest := *request
	optimizedRequest.Patches = optimizedOrder
	optimizedRequest.Strategy = StrategySequential
	
	return pa.applySequential(ctx, &optimizedRequest, result)
}

// applyConservative applies patches with maximum safety checks
func (pa *PatchApplicator) applyConservative(ctx context.Context, request *ApplicationRequest, result *ApplicationResult) error {
	for i, patch := range request.Patches {
		// Pre-validate each patch
		validationResult, err := pa.validator.ValidatePatch(patch, &ValidationOptions{
			CheckSafety:     true,
			CheckSyntax:     true,
			CheckSemantics:  true,
			CheckBuild:      true,
			CheckTests:      true,
		})
		if err != nil || !validationResult.Valid {
			failedPatch := &FailedPatch{
				Patch:    patch,
				Error:    "pre-validation failed",
				Reason:   "validation_failed",
				FailedAt: time.Now(),
			}
			result.FailedPatches = append(result.FailedPatches, failedPatch)
			continue
		}
		
		// Apply with extra caution
		appliedPatch, err := pa.applySinglePatch(ctx, request, patch, i)
		if err != nil {
			failedPatch := &FailedPatch{
				Patch:    patch,
				Error:    err.Error(),
				Reason:   "conservative_application_failed",
				FailedAt: time.Now(),
			}
			result.FailedPatches = append(result.FailedPatches, failedPatch)
			continue
		}
		
		result.AppliedPatches = append(result.AppliedPatches, appliedPatch)
	}
	
	return nil
}

// applySinglePatch applies a single patch with conflict resolution
func (pa *PatchApplicator) applySinglePatch(ctx context.Context, request *ApplicationRequest, patch *Patch, index int) (*AppliedPatch, error) {
	startTime := time.Now()
	
	// Check for potential conflicts
	conflicts, err := pa.detectConflicts(patch, request.Repository, request.Branch)
	if err != nil {
		return nil, fmt.Errorf("conflict detection failed: %w", err)
	}
	
	// Resolve conflicts if found
	if len(conflicts) > 0 {
		resolvedConflicts, err := pa.conflictResolver.ResolveConflicts(ctx, conflicts, request.Options.ConflictResolution)
		if err != nil {
			return nil, fmt.Errorf("conflict resolution failed: %w", err)
		}
		conflicts = resolvedConflicts
	}
	
	// Apply the patch
	err = pa.applyPatchToFiles(patch, request.Repository)
	if err != nil {
		return nil, fmt.Errorf("patch application failed: %w", err)
	}
	
	appliedPatch := &AppliedPatch{
		Patch:      patch,
		AppliedAt:  time.Now(),
		Duration:   time.Since(startTime),
		Conflicts:  conflicts,
		Resolution: "success",
		Confidence: patch.Confidence,
	}
	
	return appliedPatch, nil
}

// detectConflicts detects potential conflicts before applying a patch
func (pa *PatchApplicator) detectConflicts(patch *Patch, repository, branch string) ([]*Conflict, error) {
	var conflicts []*Conflict
	
	// Check each file patch for conflicts
	for _, filePatch := range patch.FilePatches {
		fileConflicts, err := pa.detectFileConflicts(filePatch, repository)
		if err != nil {
			return nil, err
		}
		conflicts = append(conflicts, fileConflicts...)
	}
	
	return conflicts, nil
}

// detectFileConflicts detects conflicts in a single file
func (pa *PatchApplicator) detectFileConflicts(filePatch *FilePatch, repository string) ([]*Conflict, error) {
	var conflicts []*Conflict
	
	// Read current file content
	filePath := filepath.Join(repository, filePatch.Path)
	currentContent, err := os.ReadFile(filePath)
	if err != nil {
		// File doesn't exist - no conflict
		return conflicts, nil
	}
	
	currentLines := strings.Split(string(currentContent), "\n")
	
	// Check each change for conflicts
	for _, change := range filePatch.Changes {
		conflict := pa.checkChangeConflict(change, currentLines, filePatch.Path)
		if conflict != nil {
			conflicts = append(conflicts, conflict)
		}
	}
	
	return conflicts, nil
}

// checkChangeConflict checks if a change conflicts with current content
func (pa *PatchApplicator) checkChangeConflict(change *Change, currentLines []string, filePath string) *Conflict {
	// Simple conflict detection - in real implementation would be more sophisticated
	for i, line := range currentLines {
		if strings.Contains(line, change.OldContent) && !strings.Contains(line, change.NewContent) {
			return &Conflict{
				File:       filePath,
				Type:       ConflictTypeContent,
				Line:       i + 1,
				Current:    line,
				Incoming:   change.NewContent,
				Context:    getContextLines(currentLines, i, 3),
				Severity:   SeverityMedium,
				Confidence: 0.7,
			}
		}
	}
	
	return nil
}

// applyPatchToFiles applies a patch to the actual files
func (pa *PatchApplicator) applyPatchToFiles(patch *Patch, repository string) error {
	for _, filePatch := range patch.FilePatches {
		err := pa.applyFilePatch(filePatch, repository)
		if err != nil {
			return fmt.Errorf("failed to apply patch to %s: %w", filePatch.Path, err)
		}
	}
	
	for _, configPatch := range patch.ConfigPatches {
		err := pa.applyConfigPatch(configPatch, repository)
		if err != nil {
			return fmt.Errorf("failed to apply config patch to %s: %w", configPatch.File, err)
		}
	}
	
	return nil
}

// applyFilePatch applies a file patch
func (pa *PatchApplicator) applyFilePatch(filePatch *FilePatch, repository string) error {
	filePath := filepath.Join(repository, filePatch.Path)
	
	// Read current content
	content, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	
	currentContent := string(content)
	
	// Apply changes
	for _, change := range filePatch.Changes {
		currentContent = strings.ReplaceAll(currentContent, change.OldContent, change.NewContent)
	}
	
	// Write updated content
	err = os.WriteFile(filePath, []byte(currentContent), 0644)
	if err != nil {
		return err
	}
	
	return nil
}

// applyConfigPatch applies a configuration patch
func (pa *PatchApplicator) applyConfigPatch(configPatch *ConfigPatch, repository string) error {
	// Simplified config patch application
	// In real implementation would handle JSON, YAML, TOML, etc.
	return nil
}

// groupNonConflictingPatches groups patches that don't conflict with each other
func (pa *PatchApplicator) groupNonConflictingPatches(patches []*Patch) [][]*Patch {
	// Simplified grouping - in real implementation would analyze file dependencies
	var groups [][]*Patch
	
	// For now, just put each patch in its own group
	for _, patch := range patches {
		groups = append(groups, []*Patch{patch})
	}
	
	return groups
}

// applyPatchGroup applies a group of patches in parallel
func (pa *PatchApplicator) applyPatchGroup(ctx context.Context, request *ApplicationRequest, group []*Patch, result *ApplicationResult) error {
	// For now, apply sequentially - parallel implementation would use goroutines
	for i, patch := range group {
		appliedPatch, err := pa.applySinglePatch(ctx, request, patch, i)
		if err != nil {
			failedPatch := &FailedPatch{
				Patch:    patch,
				Error:    err.Error(),
				Reason:   "group_application_failed",
				FailedAt: time.Now(),
			}
			result.FailedPatches = append(result.FailedPatches, failedPatch)
			continue
		}
		
		result.AppliedPatches = append(result.AppliedPatches, appliedPatch)
	}
	
	return nil
}

// optimizePatchOrder optimizes the order of patch application
func (pa *PatchApplicator) optimizePatchOrder(patches []*Patch) []*Patch {
	// Sort by confidence (highest first)
	optimized := make([]*Patch, len(patches))
	copy(optimized, patches)
	
	for i := 0; i < len(optimized)-1; i++ {
		for j := i + 1; j < len(optimized); j++ {
			if optimized[i].Confidence < optimized[j].Confidence {
				optimized[i], optimized[j] = optimized[j], optimized[i]
			}
		}
	}
	
	return optimized
}

// validateApplication validates the result of patch application
func (pa *PatchApplicator) validateApplication(ctx context.Context, request *ApplicationRequest, result *ApplicationResult) ([]*ValidationResult, error) {
	var validationResults []*ValidationResult
	
	// Run build validation
	buildResult, err := pa.runBuildValidation(request.Repository)
	if err != nil {
		return nil, err
	}
	validationResults = append(validationResults, buildResult)
	
	// Run test validation
	testResult, err := pa.runTestValidation(request.Repository)
	if err != nil {
		return nil, err
	}
	validationResults = append(validationResults, testResult)
	
	return validationResults, nil
}

// runBuildValidation runs build validation
func (pa *PatchApplicator) runBuildValidation(repository string) (*ValidationResult, error) {
	cmd := exec.Command("npm", "run", "build")
	cmd.Dir = repository
	
	output, err := cmd.CombinedOutput()
	
	return &ValidationResult{
		Type:    "build",
		Success: err == nil,
		Output:  string(output),
		Error:   getErrorString(err),
	}, nil
}

// runTestValidation runs test validation
func (pa *PatchApplicator) runTestValidation(repository string) (*ValidationResult, error) {
	cmd := exec.Command("npm", "test")
	cmd.Dir = repository
	
	output, err := cmd.CombinedOutput()
	
	return &ValidationResult{
		Type:    "test",
		Success: err == nil,
		Output:  string(output),
		Error:   getErrorString(err),
	}, nil
}

// calculateSummary calculates application summary
func (pa *PatchApplicator) calculateSummary(result *ApplicationResult) *ApplicationSummary {
	totalPatches := len(result.AppliedPatches) + len(result.FailedPatches)
	successRate := 0.0
	if totalPatches > 0 {
		successRate = float64(len(result.AppliedPatches)) / float64(totalPatches)
	}
	
	return &ApplicationSummary{
		TotalPatches:      totalPatches,
		AppliedPatches:    len(result.AppliedPatches),
		FailedPatches:     len(result.FailedPatches),
		ConflictsFound:    len(result.Conflicts),
		ConflictsResolved: pa.countResolvedConflicts(result.Conflicts),
		Duration:          result.Duration,
		SuccessRate:       successRate,
	}
}

// countResolvedConflicts counts resolved conflicts
func (pa *PatchApplicator) countResolvedConflicts(conflicts []*Conflict) int {
	count := 0
	for _, conflict := range conflicts {
		if conflict.Resolution != nil {
			count++
		}
	}
	return count
}

// generateRecommendations generates recommendations based on application results
func (pa *PatchApplicator) generateRecommendations(result *ApplicationResult) []*Recommendation {
	var recommendations []*Recommendation
	
	// Recommend strategy changes based on success rate
	if result.Summary.SuccessRate < 0.5 {
		recommendations = append(recommendations, &Recommendation{
			Type:        "strategy",
			Priority:    "high",
			Description: "Consider using conservative strategy due to low success rate",
			Action:      "switch_to_conservative_strategy",
			Confidence:  0.8,
		})
	}
	
	// Recommend conflict resolution improvements
	if len(result.Conflicts) > result.Summary.ConflictsResolved {
		recommendations = append(recommendations, &Recommendation{
			Type:        "conflict_resolution",
			Priority:    "medium",
			Description: "Improve conflict resolution to handle more conflicts automatically",
			Action:      "enhance_conflict_resolution",
			Confidence:  0.7,
		})
	}
	
	return recommendations
}

// Helper functions
func getContextLines(lines []string, center, radius int) string {
	start := max(0, center-radius)
	end := min(len(lines), center+radius+1)
	
	contextLines := lines[start:end]
	return strings.Join(contextLines, "\n")
}

func getErrorString(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
