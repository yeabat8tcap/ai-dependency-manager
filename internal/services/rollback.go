package services

import (
	"context"
	"fmt"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/database"
	"github.com/8tcapital/ai-dep-manager/internal/logger"
	"github.com/8tcapital/ai-dep-manager/internal/models"
	"github.com/8tcapital/ai-dep-manager/internal/packagemanager"
	pmtypes "github.com/8tcapital/ai-dep-manager/internal/packagemanager/types"
	"gorm.io/gorm"
)

// RollbackService handles rollback operations
type RollbackService struct {
	db *gorm.DB
}

// NewRollbackService creates a new rollback service
func NewRollbackService() *RollbackService {
	return &RollbackService{
		db: database.GetDB(),
	}
}

// RollbackOptions contains options for rollback operations
type RollbackOptions struct {
	DryRun bool
	Force  bool
}



// RollbackResult represents the result of a rollback operation
type RollbackResult struct {
	ProjectID      uint                    `json:"project_id"`
	ProjectName    string                  `json:"project_name"`
	TotalAttempted int                     `json:"total_attempted"`
	Successful     []RollbackItem          `json:"successful"`
	Failed         []RollbackFailure       `json:"failed"`
	Duration       time.Duration           `json:"duration"`
}

// RollbackFailure represents a failed rollback operation
type RollbackFailure struct {
	RollbackItem RollbackItem `json:"rollback_item"`
	Error        string       `json:"error"`
	Timestamp    time.Time    `json:"timestamp"`
}

// CreateRollbackPlan creates a new rollback plan
func (rs *RollbackService) CreateRollbackPlan(ctx context.Context, projectID uint, rollbacks []RollbackItem) (*RollbackPlan, error) {
	logger.Info("Creating rollback plan for project ID: %d", projectID)
	
	// Get project
	var project models.Project
	if err := rs.db.First(&project, projectID).Error; err != nil {
		return nil, fmt.Errorf("failed to find project: %w", err)
	}
	
	// Create rollback plan record
	rollbackPlan := &models.RollbackPlan{
		ProjectID: projectID,
		Status:    "created",
		CreatedAt: time.Now(),
	}
	
	if err := rs.db.Create(rollbackPlan).Error; err != nil {
		return nil, fmt.Errorf("failed to create rollback plan: %w", err)
	}
	
	// Create rollback items
	for _, rollback := range rollbacks {
		rollbackItem := &models.RollbackItem{
			RollbackPlanID: rollbackPlan.ID,
			DependencyName: rollback.DependencyName,
			FromVersion:    rollback.FromVersion,
			ToVersion:      rollback.ToVersion,
			Command:        rollback.Command,
		}
		
		if err := rs.db.Create(rollbackItem).Error; err != nil {
			return nil, fmt.Errorf("failed to create rollback item: %w", err)
		}
	}
	
	plan := &RollbackPlan{
		ProjectID:   projectID,
		CreatedAt:   rollbackPlan.CreatedAt,
		Rollbacks:   rollbacks,
	}
	
	logger.Info("Created rollback plan ID: %d with %d operations", rollbackPlan.ID, len(rollbacks))
	return plan, nil
}

// ListRollbackPlans lists all available rollback plans
func (rs *RollbackService) ListRollbackPlans(ctx context.Context) ([]*RollbackPlan, error) {
	var dbPlans []models.RollbackPlan
	
	err := rs.db.Joins("JOIN projects ON rollback_plans.project_id = projects.id").
		Select("rollback_plans.*, projects.name as project_name").
		Order("rollback_plans.created_at DESC").
		Find(&dbPlans).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to list rollback plans: %w", err)
	}
	
	var plans []*RollbackPlan
	for _, dbPlan := range dbPlans {
		// Get rollback items
		var dbItems []models.RollbackItem
		rs.db.Where("rollback_plan_id = ?", dbPlan.ID).Find(&dbItems)
		
		var rollbacks []RollbackItem
		for _, dbItem := range dbItems {
			rollbacks = append(rollbacks, RollbackItem{
				DependencyName: dbItem.DependencyName,
				FromVersion:    dbItem.FromVersion,
				ToVersion:      dbItem.ToVersion,
				Command:        dbItem.Command,
			})
		}
		
		// Get project name
		var project models.Project
		rs.db.First(&project, dbPlan.ProjectID)
		
		plan := &RollbackPlan{
			ProjectID:   dbPlan.ProjectID,
			CreatedAt:   dbPlan.CreatedAt,
			Rollbacks:   rollbacks,
		}
		
		plans = append(plans, plan)
	}
	
	return plans, nil
}

// GetRollbackPlan gets a specific rollback plan by ID
func (rs *RollbackService) GetRollbackPlan(ctx context.Context, planID uint) (*RollbackPlan, error) {
	var dbPlan models.RollbackPlan
	
	err := rs.db.First(&dbPlan, planID).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find rollback plan: %w", err)
	}
	
	// Get rollback items
	var dbItems []models.RollbackItem
	rs.db.Where("rollback_plan_id = ?", planID).Find(&dbItems)
	
	var rollbacks []RollbackItem
	for _, dbItem := range dbItems {
		rollbacks = append(rollbacks, RollbackItem{
			DependencyName: dbItem.DependencyName,
			FromVersion:    dbItem.FromVersion,
			ToVersion:      dbItem.ToVersion,
			Command:        dbItem.Command,
		})
	}
	
	// Get project name
	var project models.Project
	rs.db.First(&project, dbPlan.ProjectID)
	
	plan := &RollbackPlan{
		ProjectID:   dbPlan.ProjectID,
		CreatedAt:   dbPlan.CreatedAt,
		Rollbacks:   rollbacks,
	}
	
	return plan, nil
}

// GetLatestRollbackPlan gets the latest rollback plan for a project
func (rs *RollbackService) GetLatestRollbackPlan(ctx context.Context, projectID uint) (*RollbackPlan, error) {
	var dbPlan models.RollbackPlan
	
	err := rs.db.Where("project_id = ? AND status = ?", projectID, "created").
		Order("created_at DESC").
		First(&dbPlan).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No rollback plan found
		}
		return nil, fmt.Errorf("failed to find latest rollback plan: %w", err)
	}
	
	return rs.GetRollbackPlan(ctx, dbPlan.ID)
}

// ExecuteRollback executes a rollback plan
func (rs *RollbackService) ExecuteRollback(ctx context.Context, planID uint, options *RollbackOptions) (*RollbackResult, error) {
	logger.Info("Executing rollback plan ID: %d", planID)
	
	startTime := time.Now()
	
	// Get rollback plan
	plan, err := rs.GetRollbackPlan(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rollback plan: %w", err)
	}
	
	if plan.Status == "executed" {
		return nil, fmt.Errorf("rollback plan has already been executed")
	}
	
	// Get project and package manager
	var project models.Project
	if err := rs.db.First(&project, plan.ProjectID).Error; err != nil {
		return nil, fmt.Errorf("failed to find project: %w", err)
	}
	
	pm, exists := packagemanager.GetPackageManager(project.Type)
	if !exists {
		return nil, fmt.Errorf("unsupported package manager: %s", project.Type)
	}
	
	result := &RollbackResult{
		ProjectID:   plan.ProjectID,
		ProjectName: plan.ProjectName,
	}
	
	// Execute rollback operations
	for _, rollback := range plan.Rollbacks {
		result.TotalAttempted++
		
		if options.DryRun {
			logger.Info("DRY RUN: Would rollback %s from %s to %s", 
				rollback.DependencyName, rollback.FromVersion, rollback.ToVersion)
			result.Successful = append(result.Successful, rollback)
		} else {
			err := rs.executeRollbackItem(ctx, &project, pm, rollback)
			if err != nil {
				logger.Error("Failed to rollback %s: %v", rollback.DependencyName, err)
				result.Failed = append(result.Failed, RollbackFailure{
					RollbackItem: rollback,
					Error:        err.Error(),
					Timestamp:    time.Now(),
				})
				
				// Stop on first failure unless force is enabled
				if !options.Force {
					break
				}
			} else {
				logger.Info("Successfully rolled back %s from %s to %s", 
					rollback.DependencyName, rollback.FromVersion, rollback.ToVersion)
				result.Successful = append(result.Successful, rollback)
			}
		}
	}
	
	result.Duration = time.Since(startTime)
	
	// Update rollback plan status if not dry run
	if !options.DryRun {
		status := "executed"
		if len(result.Failed) > 0 {
			status = "partially_executed"
		}
		
		now := time.Now()
		err := rs.db.Model(&models.RollbackPlan{}).
			Where("id = ?", planID).
			Updates(map[string]interface{}{
				"status":      status,
				"executed_at": &now,
			}).Error
		
		if err != nil {
			logger.Error("Failed to update rollback plan status: %v", err)
		}
	}
	
	logger.Info("Rollback operation completed: %d successful, %d failed", 
		len(result.Successful), len(result.Failed))
	
	return result, nil
}

// DeleteRollbackPlan deletes a rollback plan
func (rs *RollbackService) DeleteRollbackPlan(ctx context.Context, planID uint) error {
	logger.Info("Deleting rollback plan ID: %d", planID)
	
	// Delete rollback items first
	if err := rs.db.Where("rollback_plan_id = ?", planID).Delete(&models.RollbackItem{}).Error; err != nil {
		return fmt.Errorf("failed to delete rollback items: %w", err)
	}
	
	// Delete rollback plan
	if err := rs.db.Delete(&models.RollbackPlan{}, planID).Error; err != nil {
		return fmt.Errorf("failed to delete rollback plan: %w", err)
	}
	
	logger.Info("Deleted rollback plan ID: %d", planID)
	return nil
}

// CleanupOldRollbackPlans removes old executed rollback plans
func (rs *RollbackService) CleanupOldRollbackPlans(ctx context.Context, olderThan time.Duration) error {
	cutoffTime := time.Now().Add(-olderThan)
	
	// Find old executed plans
	var oldPlans []models.RollbackPlan
	err := rs.db.Where("status IN ? AND executed_at < ?", 
		[]string{"executed", "partially_executed"}, cutoffTime).
		Find(&oldPlans).Error
	
	if err != nil {
		return fmt.Errorf("failed to find old rollback plans: %w", err)
	}
	
	logger.Info("Cleaning up %d old rollback plans", len(oldPlans))
	
	// Delete each plan
	for _, plan := range oldPlans {
		if err := rs.DeleteRollbackPlan(ctx, plan.ID); err != nil {
			logger.Error("Failed to delete rollback plan %d: %v", plan.ID, err)
		}
	}
	
	return nil
}

// Helper methods

func (rs *RollbackService) executeRollbackItem(ctx context.Context, project *models.Project, pm packagemanager.PackageManager, rollback RollbackItem) error {
	// Create package manager update options
	pmOptions := &pmtypes.UpdateOptions{
		DryRun: false,
		Force:  false,
	}
	
	// Execute the rollback (which is essentially an update to the previous version)
	err := pm.UpdateDependency(ctx, project.Path, rollback.DependencyName, rollback.ToVersion, pmOptions)
	if err != nil {
		return fmt.Errorf("package manager rollback failed: %w", err)
	}
	
	// Update dependency record in database
	var dependency models.Dependency
	err = rs.db.Where("project_id = ? AND name = ?", project.ID, rollback.DependencyName).
		First(&dependency).Error
	
	if err == nil {
		dependency.CurrentVersion = rollback.ToVersion
		dependency.UpdatedAt = time.Now()
		rs.db.Save(&dependency)
	}
	
	return nil
}
