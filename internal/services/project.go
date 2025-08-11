package services

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/8tcapital/ai-dep-manager/internal/database"
	"github.com/8tcapital/ai-dep-manager/internal/logger"
	"github.com/8tcapital/ai-dep-manager/internal/models"
	"github.com/8tcapital/ai-dep-manager/internal/packagemanager"
	"gorm.io/gorm"
)

// ProjectService handles project-related operations
type ProjectService struct {
	db *gorm.DB
}

// NewProjectService creates a new project service
func NewProjectService() *ProjectService {
	return &ProjectService{
		db: database.GetDB(),
	}
}

// CreateProject creates a new project
func (ps *ProjectService) CreateProject(ctx context.Context, name, path, pmType string) (*models.Project, error) {
	// Validate package manager type
	pm, exists := packagemanager.GetPackageManager(pmType)
	if !exists {
		return nil, fmt.Errorf("unsupported package manager: %s", pmType)
	}
	
	// Validate project path
	if err := pm.ValidateProject(ctx, path); err != nil {
		return nil, fmt.Errorf("project validation failed: %w", err)
	}
	
	// Detect config file
	projects, err := pm.DetectProjects(ctx, path)
	if err != nil || len(projects) == 0 {
		return nil, fmt.Errorf("no valid %s project found at %s", pmType, path)
	}
	
	configFile := projects[0].ConfigFile
	
	// Check if project already exists
	var existingProject models.Project
	err = ps.db.Where("path = ? OR (name = ? AND type = ?)", path, name, pmType).First(&existingProject).Error
	if err == nil {
		return nil, fmt.Errorf("project already exists: %s", existingProject.Name)
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing project: %w", err)
	}
	
	// Create project
	project := &models.Project{
		Name:       name,
		Path:       path,
		Type:       pmType,
		ConfigFile: configFile,
		Enabled:    true,
	}
	
	if err := ps.db.Create(project).Error; err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}
	
	// Create default settings
	settings := &models.ProjectSettings{
		ProjectID:           project.ID,
		AutoUpdateLevel:     "none",
		RequireConfirmation: true,
		NotificationEnabled: true,
	}
	
	if err := ps.db.Create(settings).Error; err != nil {
		logger.Error("Failed to create project settings: %v", err)
	}
	
	logger.Info("Created project: %s (%s) at %s", name, pmType, path)
	return project, nil
}

// GetProject retrieves a project by ID
func (ps *ProjectService) GetProject(ctx context.Context, id uint) (*models.Project, error) {
	var project models.Project
	err := ps.db.Preload("Settings").Preload("Dependencies").First(&project, id).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	return &project, nil
}

// GetProjectByName retrieves a project by name
func (ps *ProjectService) GetProjectByName(ctx context.Context, name string) (*models.Project, error) {
	var project models.Project
	err := ps.db.Preload("Settings").Preload("Dependencies").Where("name = ?", name).First(&project).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	return &project, nil
}

// ListProjects lists all projects
func (ps *ProjectService) ListProjects(ctx context.Context, enabled *bool) ([]models.Project, error) {
	var projects []models.Project
	query := ps.db.Preload("Settings")
	
	if enabled != nil {
		query = query.Where("enabled = ?", *enabled)
	}
	
	err := query.Find(&projects).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	
	return projects, nil
}

// UpdateProject updates a project
func (ps *ProjectService) UpdateProject(ctx context.Context, id uint, updates map[string]interface{}) (*models.Project, error) {
	var project models.Project
	if err := ps.db.First(&project, id).Error; err != nil {
		return nil, fmt.Errorf("failed to find project: %w", err)
	}
	
	if err := ps.db.Model(&project).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}
	
	logger.Info("Updated project: %s", project.Name)
	return &project, nil
}

// DeleteProject deletes a project
func (ps *ProjectService) DeleteProject(ctx context.Context, id uint) error {
	var project models.Project
	if err := ps.db.First(&project, id).Error; err != nil {
		return fmt.Errorf("failed to find project: %w", err)
	}
	
	// Delete related records
	ps.db.Where("project_id = ?", id).Delete(&models.Dependency{})
	ps.db.Where("project_id = ?", id).Delete(&models.ProjectSettings{})
	ps.db.Where("project_id = ?", id).Delete(&models.ScanResult{})
	
	// Delete project
	if err := ps.db.Delete(&project).Error; err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	
	logger.Info("Deleted project: %s", project.Name)
	return nil
}

// AutoDiscoverProjects automatically discovers projects in a directory
func (ps *ProjectService) AutoDiscoverProjects(ctx context.Context, rootPath string) ([]models.Project, error) {
	logger.Info("Auto-discovering projects in: %s", rootPath)
	
	// Detect all projects using available package managers
	detectedProjects, err := packagemanager.DetectAllProjects(ctx, rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect projects: %w", err)
	}
	
	var createdProjects []models.Project
	
	for _, detected := range detectedProjects {
		// Check if project already exists
		var existingProject models.Project
		err := ps.db.Where("path = ?", detected.Path).First(&existingProject).Error
		if err == nil {
			logger.Info("Project already exists: %s at %s", detected.Name, detected.Path)
			continue
		} else if err != gorm.ErrRecordNotFound {
			logger.Error("Failed to check existing project: %v", err)
			continue
		}
		
		// Create project name if not provided
		projectName := detected.Name
		if projectName == "" {
			projectName = filepath.Base(detected.Path)
		}
		
		// Create project
		project, err := ps.CreateProject(ctx, projectName, detected.Path, detected.PackageManager)
		if err != nil {
			logger.Error("Failed to create discovered project %s: %v", projectName, err)
			continue
		}
		
		createdProjects = append(createdProjects, *project)
	}
	
	logger.Info("Auto-discovered %d new projects", len(createdProjects))
	return createdProjects, nil
}

// UpdateProjectSettings updates project-specific settings
func (ps *ProjectService) UpdateProjectSettings(ctx context.Context, projectID uint, settings map[string]interface{}) error {
	var projectSettings models.ProjectSettings
	err := ps.db.Where("project_id = ?", projectID).First(&projectSettings).Error
	
	if err == gorm.ErrRecordNotFound {
		// Create new settings
		projectSettings = models.ProjectSettings{
			ProjectID:           projectID,
			AutoUpdateLevel:     "none",
			RequireConfirmation: true,
			NotificationEnabled: true,
		}
		
		if err := ps.db.Create(&projectSettings).Error; err != nil {
			return fmt.Errorf("failed to create project settings: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to get project settings: %w", err)
	}
	
	// Update settings
	if err := ps.db.Model(&projectSettings).Updates(settings).Error; err != nil {
		return fmt.Errorf("failed to update project settings: %w", err)
	}
	
	logger.Info("Updated settings for project ID: %d", projectID)
	return nil
}
