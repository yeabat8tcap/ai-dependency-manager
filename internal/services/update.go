package services

import (
	"context"
	"fmt"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/ai"
	aitypes "github.com/8tcapital/ai-dep-manager/internal/ai/types"
	"github.com/8tcapital/ai-dep-manager/internal/database"
	"github.com/8tcapital/ai-dep-manager/internal/logger"
	"github.com/8tcapital/ai-dep-manager/internal/models"
	"github.com/8tcapital/ai-dep-manager/internal/packagemanager"
	pmtypes "github.com/8tcapital/ai-dep-manager/internal/packagemanager/types"
	"gorm.io/gorm"
)

// UpdateService handles dependency update operations
type UpdateService struct {
	db *gorm.DB
}

// NewUpdateService creates a new update service
func NewUpdateService() *UpdateService {
	return &UpdateService{
		db: database.GetDB(),
	}
}

// UpdateOptions contains options for update operations
type UpdateOptions struct {
	ProjectID         uint
	DependencyNames   []string
	UpdateTypes       []string // major, minor, patch, security
	RiskLevels        []string // low, medium, high, critical
	DryRun            bool
	Force             bool
	Interactive       bool
	AutoApprove       bool
	SkipBreaking      bool
	SecurityOnly      bool
	BatchSize         int
	Timeout           time.Duration
}

// UpdatePlan represents a plan for applying updates
type UpdatePlan struct {
	ProjectID     uint                    `json:"project_id"`
	ProjectName   string                  `json:"project_name"`
	TotalUpdates  int                     `json:"total_updates"`
	UpdateGroups  []UpdateGroup           `json:"update_groups"`
	Recommendations []string              `json:"recommendations"`
	Warnings      []string                `json:"warnings"`
	EstimatedTime time.Duration           `json:"estimated_time"`
	RiskSummary   UpdateRiskSummary       `json:"risk_summary"`
}

// UpdateGroup represents a group of related updates
type UpdateGroup struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Priority     ai.Priority            `json:"priority"`
	RiskLevel    ai.RiskLevel           `json:"risk_level"`
	Updates      []UpdateItem           `json:"updates"`
	Dependencies []string               `json:"dependencies"` // Other groups this depends on
	CanParallel  bool                   `json:"can_parallel"`
}

// UpdateItem represents a single update to be applied
type UpdateItem struct {
	UpdateID         uint                   `json:"update_id"`
	DependencyID     uint                   `json:"dependency_id"`
	DependencyName   string                 `json:"dependency_name"`
	FromVersion      string                 `json:"from_version"`
	ToVersion        string                 `json:"to_version"`
	UpdateType       string                 `json:"update_type"`
	RiskLevel        ai.RiskLevel           `json:"risk_level"`
	BreakingChange   bool                   `json:"breaking_change"`
	SecurityFix      bool                   `json:"security_fix"`
	Confidence       float64                `json:"confidence"`
	Recommendations  []string               `json:"recommendations"`
	AIPredictions    []models.AIPrediction  `json:"ai_predictions,omitempty"`
}

// UpdateRiskSummary provides an overview of update risks
type UpdateRiskSummary struct {
	TotalUpdates     int `json:"total_updates"`
	LowRisk          int `json:"low_risk"`
	MediumRisk       int `json:"medium_risk"`
	HighRisk         int `json:"high_risk"`
	CriticalRisk     int `json:"critical_risk"`
	BreakingChanges  int `json:"breaking_changes"`
	SecurityUpdates  int `json:"security_updates"`
	OverallRisk      ai.RiskLevel `json:"overall_risk"`
}

// UpdateResult represents the result of applying updates
type UpdateResult struct {
	ProjectID        uint                    `json:"project_id"`
	ProjectName      string                  `json:"project_name"`
	TotalAttempted   int                     `json:"total_attempted"`
	Successful       []UpdateItem            `json:"successful"`
	Failed           []UpdateFailure         `json:"failed"`
	Skipped          []UpdateItem            `json:"skipped"`
	Duration         time.Duration           `json:"duration"`
	RollbackPlan     *RollbackPlan           `json:"rollback_plan,omitempty"`
}

// UpdateFailure represents a failed update
type UpdateFailure struct {
	UpdateItem UpdateItem `json:"update_item"`
	Error      string     `json:"error"`
	Timestamp  time.Time  `json:"timestamp"`
}

// RollbackPlan contains information for rolling back updates (service layer)
type RollbackPlan struct {
	ProjectID    uint           `json:"project_id"`
	ProjectName  string         `json:"project_name"`
	Status       string         `json:"status"`
	Rollbacks    []RollbackItem `json:"rollbacks"`
	Instructions []string       `json:"instructions"`
	CreatedAt    time.Time      `json:"created_at"`
}

// RollbackItem represents a single rollback operation (service layer)
type RollbackItem struct {
	DependencyName string `json:"dependency_name"`
	FromVersion    string `json:"from_version"`
	ToVersion      string `json:"to_version"`
	Command        string `json:"command"`
}


// GenerateUpdatePlan creates an update plan for a project
func (us *UpdateService) GenerateUpdatePlan(ctx context.Context, options *UpdateOptions) (*UpdatePlan, error) {
	logger.Info("Generating update plan for project ID: %d", options.ProjectID)
	
	// Get project
	var project models.Project
	if err := us.db.First(&project, options.ProjectID).Error; err != nil {
		return nil, fmt.Errorf("failed to find project: %w", err)
	}
	
	// Get pending updates
	updates, err := us.getPendingUpdates(options)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending updates: %w", err)
	}
	
	if len(updates) == 0 {
		return &UpdatePlan{
			ProjectID:   options.ProjectID,
			ProjectName: project.Name,
			UpdateGroups: []UpdateGroup{},
			Recommendations: []string{"No updates available"},
		}, nil
	}
	
	// Group updates by priority and risk
	updateGroups := us.groupUpdates(updates)
	
	// Calculate risk summary
	riskSummary := us.calculateRiskSummary(updates)
	
	// Generate recommendations and warnings
	recommendations := us.generateRecommendations(updates, riskSummary)
	warnings := us.generateWarnings(updates, riskSummary)
	
	// Estimate time
	estimatedTime := us.estimateUpdateTime(updates)
	
	plan := &UpdatePlan{
		ProjectID:       options.ProjectID,
		ProjectName:     project.Name,
		TotalUpdates:    len(updates),
		UpdateGroups:    updateGroups,
		Recommendations: recommendations,
		Warnings:        warnings,
		EstimatedTime:   estimatedTime,
		RiskSummary:     riskSummary,
	}
	
	logger.Info("Generated update plan: %d updates in %d groups", len(updates), len(updateGroups))
	return plan, nil
}

// ApplyUpdates applies the updates according to the plan
func (us *UpdateService) ApplyUpdates(ctx context.Context, plan *UpdatePlan, options *UpdateOptions) (*UpdateResult, error) {
	logger.Info("Applying updates for project: %s (%d updates)", plan.ProjectName, plan.TotalUpdates)
	
	startTime := time.Now()
	result := &UpdateResult{
		ProjectID:   plan.ProjectID,
		ProjectName: plan.ProjectName,
	}
	
	// Get project and package manager
	var project models.Project
	if err := us.db.First(&project, plan.ProjectID).Error; err != nil {
		return nil, fmt.Errorf("failed to find project: %w", err)
	}
	
	pm, exists := packagemanager.GetPackageManager(project.Type)
	if !exists {
		return nil, fmt.Errorf("unsupported package manager: %s", project.Type)
	}
	
	// Create rollback plan
	rollbackPlan := &RollbackPlan{
		ProjectID:   plan.ProjectID,
		ProjectName: plan.ProjectName,
		Status:      "pending",
		CreatedAt:   time.Now(),
	}
	
	// Apply updates group by group
	for _, group := range plan.UpdateGroups {
		logger.Info("Applying update group: %s (%d updates)", group.Name, len(group.Updates))
		
		for _, updateItem := range group.Updates {
			result.TotalAttempted++
			
			// Check if we should skip this update
			if us.shouldSkipUpdate(updateItem, options) {
				logger.Info("Skipping update: %s", updateItem.DependencyName)
				result.Skipped = append(result.Skipped, updateItem)
				continue
			}
			
			// Apply the update
			if options.DryRun {
				logger.Info("DRY RUN: Would update %s from %s to %s", 
					updateItem.DependencyName, updateItem.FromVersion, updateItem.ToVersion)
				result.Successful = append(result.Successful, updateItem)
			} else {
				err := us.applyUpdate(ctx, &project, pm, updateItem, rollbackPlan)
				if err != nil {
					logger.Error("Failed to update %s: %v", updateItem.DependencyName, err)
					result.Failed = append(result.Failed, UpdateFailure{
						UpdateItem: updateItem,
						Error:      err.Error(),
						Timestamp:  time.Now(),
					})
					
					// Stop on first failure unless force is enabled
					if !options.Force {
						break
					}
				} else {
					logger.Info("Successfully updated %s from %s to %s", 
						updateItem.DependencyName, updateItem.FromVersion, updateItem.ToVersion)
					result.Successful = append(result.Successful, updateItem)
				}
			}
		}
		
		// Stop if we have failures and force is not enabled
		if len(result.Failed) > 0 && !options.Force {
			break
		}
	}
	
	result.Duration = time.Since(startTime)
	
	// Set rollback plan if we made changes
	if len(result.Successful) > 0 && !options.DryRun {
		result.RollbackPlan = rollbackPlan
	}
	
	logger.Info("Update operation completed: %d successful, %d failed, %d skipped", 
		len(result.Successful), len(result.Failed), len(result.Skipped))
	
	return result, nil
}

// GetUpdateRecommendations provides AI-powered update recommendations
func (us *UpdateService) GetUpdateRecommendations(ctx context.Context, projectID uint) ([]string, error) {
	// Get pending updates with AI predictions
	updates, err := us.getPendingUpdatesWithPredictions(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updates: %w", err)
	}
	
	var recommendations []string
	
	// Analyze updates and generate recommendations
	securityUpdates := 0
	breakingChanges := 0
	highRiskUpdates := 0
	
	for _, update := range updates {
		if update.SecurityFix {
			securityUpdates++
		}
		if update.BreakingChange {
			breakingChanges++
		}
		if update.Severity == "high" || update.Severity == "critical" {
			highRiskUpdates++
		}
	}
	
	// Generate specific recommendations
	if securityUpdates > 0 {
		recommendations = append(recommendations, 
			fmt.Sprintf("🔒 Apply %d security update(s) immediately", securityUpdates))
	}
	
	if breakingChanges > 0 {
		recommendations = append(recommendations, 
			fmt.Sprintf("⚠️  Review %d breaking change(s) carefully before updating", breakingChanges))
		recommendations = append(recommendations, 
			"📋 Create comprehensive test plan for breaking changes")
	}
	
	if highRiskUpdates > 0 {
		recommendations = append(recommendations, 
			fmt.Sprintf("🚨 %d high-risk updates require careful evaluation", highRiskUpdates))
		recommendations = append(recommendations, 
			"🔄 Consider staging environment testing")
	}
	
	if len(updates) > 10 {
		recommendations = append(recommendations, 
			"📦 Consider batch updates to reduce complexity")
	}
	
	// Use AI for additional recommendations
	if len(updates) > 0 {
		aiRecommendations, err := us.generateAIRecommendations(ctx, updates)
		if err == nil {
			recommendations = append(recommendations, aiRecommendations...)
		}
	}
	
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "✅ All dependencies are up to date")
	}
	
	return recommendations, nil
}

// Helper methods

func (us *UpdateService) getPendingUpdates(options *UpdateOptions) ([]UpdateItem, error) {
	var updates []models.Update
	query := us.db.Where("updates.status = ?", "pending")
	
	// Filter by project
	if options.ProjectID != 0 {
		query = query.Joins("JOIN dependencies ON updates.dependency_id = dependencies.id").
			Where("dependencies.project_id = ?", options.ProjectID)
	}
	
	// Filter by dependency names
	if len(options.DependencyNames) > 0 {
		query = query.Joins("JOIN dependencies ON updates.dependency_id = dependencies.id").
			Where("dependencies.name IN ?", options.DependencyNames)
	}
	
	// Filter by update types
	if len(options.UpdateTypes) > 0 {
		query = query.Where("update_type IN ?", options.UpdateTypes)
	}
	
	// Filter by security only
	if options.SecurityOnly {
		query = query.Where("security_fix = ?", true)
	}
	
	// Filter by risk levels
	if len(options.RiskLevels) > 0 {
		query = query.Where("severity IN ?", options.RiskLevels)
	}
	
	// Skip breaking changes if requested
	if options.SkipBreaking {
		query = query.Where("breaking_change = ?", false)
	}
	
	if err := query.Preload("Dependency").Find(&updates).Error; err != nil {
		return nil, fmt.Errorf("failed to query updates: %w", err)
	}
	
	// Convert to UpdateItems
	var updateItems []UpdateItem
	for _, update := range updates {
		// Get AI predictions
		var predictions []models.AIPrediction
		us.db.Where("update_id = ?", update.ID).Find(&predictions)
		
		// Calculate confidence from predictions
		confidence := 0.5 // Default confidence
		for _, pred := range predictions {
			if pred.PredictionType == "breaking_change" {
				confidence = pred.Confidence
				break
			}
		}
		
		// Determine risk level
		riskLevel := aitypes.RiskLevelLow
		switch update.Severity {
		case "critical":
			riskLevel = aitypes.RiskLevelCritical
		case "high":
			riskLevel = aitypes.RiskLevelHigh
		case "medium":
			riskLevel = aitypes.RiskLevelMedium
		}
		
		updateItem := UpdateItem{
			UpdateID:       update.ID,
			DependencyID:   update.DependencyID,
			DependencyName: update.Dependency.Name,
			FromVersion:    update.FromVersion,
			ToVersion:      update.ToVersion,
			UpdateType:     update.UpdateType,
			RiskLevel:      riskLevel,
			BreakingChange: update.BreakingChange,
			SecurityFix:    update.SecurityFix,
			Confidence:     confidence,
			AIPredictions:  predictions,
		}
		
		updateItems = append(updateItems, updateItem)
	}
	
	return updateItems, nil
}

func (us *UpdateService) getPendingUpdatesWithPredictions(projectID uint) ([]models.Update, error) {
	var updates []models.Update
	err := us.db.Joins("JOIN dependencies ON updates.dependency_id = dependencies.id").
		Where("dependencies.project_id = ? AND updates.status = ?", projectID, "pending").
		Preload("Dependency").
		Preload("Predictions").
		Find(&updates).Error
	
	return updates, err
}

func (us *UpdateService) groupUpdates(updates []UpdateItem) []UpdateGroup {
	groups := make(map[string][]UpdateItem)
	
	// Group by priority and risk
	for _, update := range updates {
		var groupKey string
		
		if update.SecurityFix {
			groupKey = "security"
		} else if update.BreakingChange {
			groupKey = "breaking"
		} else if update.RiskLevel == aitypes.RiskLevelHigh || update.RiskLevel == aitypes.RiskLevelCritical {
			groupKey = "high-risk"
		} else if update.UpdateType == "major" {
			groupKey = "major"
		} else if update.UpdateType == "minor" {
			groupKey = "minor"
		} else {
			groupKey = "patch"
		}
		
		groups[groupKey] = append(groups[groupKey], update)
	}
	
	// Convert to UpdateGroups with proper ordering
	var updateGroups []UpdateGroup
	groupOrder := []string{"security", "breaking", "high-risk", "major", "minor", "patch"}
	
	for _, groupName := range groupOrder {
		if groupUpdates, exists := groups[groupName]; exists {
			group := UpdateGroup{
				Name:        groupName,
				Description: us.getGroupDescription(groupName),
				Priority:    us.getGroupPriority(groupName),
				RiskLevel:   us.getGroupRiskLevel(groupName),
				Updates:     groupUpdates,
				CanParallel: groupName == "patch" || groupName == "minor",
			}
			updateGroups = append(updateGroups, group)
		}
	}
	
	return updateGroups
}

func (us *UpdateService) calculateRiskSummary(updates []UpdateItem) UpdateRiskSummary {
	summary := UpdateRiskSummary{
		TotalUpdates: len(updates),
	}
	
	for _, update := range updates {
		switch update.RiskLevel {
		case aitypes.RiskLevelLow:
			summary.LowRisk++
		case aitypes.RiskLevelMedium:
			summary.MediumRisk++
		case aitypes.RiskLevelHigh:
			summary.HighRisk++
		case aitypes.RiskLevelCritical:
			summary.CriticalRisk++
		}
		
		if update.BreakingChange {
			summary.BreakingChanges++
		}
		
		if update.SecurityFix {
			summary.SecurityUpdates++
		}
	}
	
	// Determine overall risk
	if summary.CriticalRisk > 0 {
		summary.OverallRisk = aitypes.RiskLevelCritical
	} else if summary.HighRisk > 0 || summary.BreakingChanges > 0 {
		summary.OverallRisk = aitypes.RiskLevelHigh
	} else if summary.MediumRisk > 0 {
		summary.OverallRisk = aitypes.RiskLevelMedium
	} else {
		summary.OverallRisk = aitypes.RiskLevelLow
	}
	
	return summary
}

func (us *UpdateService) generateRecommendations(updates []UpdateItem, riskSummary UpdateRiskSummary) []string {
	var recommendations []string
	
	if riskSummary.SecurityUpdates > 0 {
		recommendations = append(recommendations, "Apply security updates immediately")
	}
	
	if riskSummary.BreakingChanges > 0 {
		recommendations = append(recommendations, "Test breaking changes in development environment")
		recommendations = append(recommendations, "Review migration guides for breaking changes")
	}
	
	if riskSummary.OverallRisk == aitypes.RiskLevelHigh || riskSummary.OverallRisk == aitypes.RiskLevelCritical {
		recommendations = append(recommendations, "Schedule updates during maintenance window")
		recommendations = append(recommendations, "Prepare rollback plan")
	}
	
	if len(updates) > 5 {
		recommendations = append(recommendations, "Consider applying updates in batches")
	}
	
	return recommendations
}

func (us *UpdateService) generateWarnings(updates []UpdateItem, riskSummary UpdateRiskSummary) []string {
	var warnings []string
	
	if riskSummary.CriticalRisk > 0 {
		warnings = append(warnings, "Critical risk updates detected - proceed with extreme caution")
	}
	
	if riskSummary.BreakingChanges > 3 {
		warnings = append(warnings, "Multiple breaking changes may require significant code changes")
	}
	
	if riskSummary.TotalUpdates > 20 {
		warnings = append(warnings, "Large number of updates increases complexity and risk")
	}
	
	return warnings
}

func (us *UpdateService) estimateUpdateTime(updates []UpdateItem) time.Duration {
	baseTime := 30 * time.Second // Base time per update
	
	totalTime := time.Duration(len(updates)) * baseTime
	
	// Add extra time for complex updates
	for _, update := range updates {
		if update.BreakingChange {
			totalTime += 2 * time.Minute
		}
		if update.UpdateType == "major" {
			totalTime += 1 * time.Minute
		}
	}
	
	return totalTime
}

func (us *UpdateService) shouldSkipUpdate(update UpdateItem, options *UpdateOptions) bool {
	if options.SkipBreaking && update.BreakingChange {
		return true
	}
	
	if options.SecurityOnly && !update.SecurityFix {
		return true
	}
	
	return false
}

func (us *UpdateService) applyUpdate(ctx context.Context, project *models.Project, pm packagemanager.PackageManager, update UpdateItem, rollbackPlan *RollbackPlan) error {
	// Create package manager update options
	pmOptions := &pmtypes.UpdateOptions{
		DryRun: false,
		Force:  false,
	}
	
	// Apply the update
	err := pm.UpdateDependency(ctx, project.Path, update.DependencyName, update.ToVersion, pmOptions)
	if err != nil {
		return fmt.Errorf("package manager update failed: %w", err)
	}
	
	// Add to rollback plan
	rollbackPlan.Rollbacks = append(rollbackPlan.Rollbacks, RollbackItem{
		DependencyName: update.DependencyName,
		FromVersion:    update.ToVersion,
		ToVersion:      update.FromVersion,
		Command:        fmt.Sprintf("Rollback %s from %s to %s", update.DependencyName, update.ToVersion, update.FromVersion),
	})
	
	// Update database record
	var dbUpdate models.Update
	if err := us.db.First(&dbUpdate, update.UpdateID).Error; err != nil {
		return fmt.Errorf("failed to find update record: %w", err)
	}
	
	dbUpdate.Status = "applied"
	dbUpdate.AppliedAt = &[]time.Time{time.Now()}[0]
	
	if err := us.db.Save(&dbUpdate).Error; err != nil {
		return fmt.Errorf("failed to update database record: %w", err)
	}
	
	return nil
}

func (us *UpdateService) generateAIRecommendations(ctx context.Context, updates []models.Update) ([]string, error) {
	// This would use AI to generate more sophisticated recommendations
	// For now, return basic recommendations
	var recommendations []string
	
	if len(updates) > 0 {
		recommendations = append(recommendations, "🤖 AI suggests reviewing changelog details before applying updates")
	}
	
	return recommendations, nil
}

// Helper methods for group properties
func (us *UpdateService) getGroupDescription(groupName string) string {
	descriptions := map[string]string{
		"security":  "Critical security updates that should be applied immediately",
		"breaking":  "Updates with breaking changes requiring code modifications",
		"high-risk": "High-risk updates that require careful testing",
		"major":     "Major version updates with potential compatibility issues",
		"minor":     "Minor version updates with new features",
		"patch":     "Patch updates with bug fixes and improvements",
	}
	
	if desc, exists := descriptions[groupName]; exists {
		return desc
	}
	return "Standard updates"
}

func (us *UpdateService) getGroupPriority(groupName string) aitypes.Priority {
	priorities := map[string]aitypes.Priority{
		"security":  aitypes.PriorityCritical,
		"breaking":  aitypes.PriorityHigh,
		"high-risk": aitypes.PriorityHigh,
		"major":     aitypes.PriorityMedium,
		"minor":     aitypes.PriorityMedium,
		"patch":     aitypes.PriorityLow,
	}
	
	if priority, exists := priorities[groupName]; exists {
		return priority
	}
	return aitypes.PriorityLow
}

func (us *UpdateService) getGroupRiskLevel(groupName string) aitypes.RiskLevel {
	riskLevels := map[string]aitypes.RiskLevel{
		"security":  aitypes.RiskLevelCritical,
		"breaking":  aitypes.RiskLevelHigh,
		"high-risk": aitypes.RiskLevelHigh,
		"major":     aitypes.RiskLevelMedium,
		"minor":     aitypes.RiskLevelLow,
		"patch":     aitypes.RiskLevelLow,
	}
	
	if risk, exists := riskLevels[groupName]; exists {
		return risk
	}
	return aitypes.RiskLevelLow
}
