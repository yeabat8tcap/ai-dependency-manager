package scanner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/ai"
	"github.com/8tcapital/ai-dep-manager/internal/ai/types"
	"github.com/8tcapital/ai-dep-manager/internal/database"
	"github.com/8tcapital/ai-dep-manager/internal/logger"
	"github.com/8tcapital/ai-dep-manager/internal/models"
	"github.com/8tcapital/ai-dep-manager/internal/packagemanager"
	pmtypes "github.com/8tcapital/ai-dep-manager/internal/packagemanager/types"
	"gorm.io/gorm"
)

// Scanner handles dependency scanning operations
type Scanner struct {
	db             *gorm.DB
	packageManager *packagemanager.Manager
	maxConcurrency int
}

// NewScanner creates a new dependency scanner
func NewScanner(maxConcurrency int) *Scanner {
	return &Scanner{
		db:             database.GetDB(),
		packageManager: packagemanager.GetManager(),
		maxConcurrency: maxConcurrency,
	}
}

// ScanOptions contains options for scanning operations
type ScanOptions struct {
	ProjectID    uint
	ScanType     string // full, incremental, security
	ForceRefresh bool
	Timeout      time.Duration
}

// ScanResult contains the results of a scan operation
type ScanResult struct {
	ProjectID         uint
	DependenciesFound int
	UpdatesFound      int
	NewDependencies   []models.Dependency
	UpdatedDependencies []models.Dependency
	AvailableUpdates  []models.Update
	Errors            []error
}

// ScanProject scans a single project for dependency updates
func (s *Scanner) ScanProject(ctx context.Context, projectID uint, options *ScanOptions) (*ScanResult, error) {
	logger.Info("Starting dependency scan for project ID: %d", projectID)
	
	// Get project from database
	var project models.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		return nil, fmt.Errorf("failed to find project: %w", err)
	}
	
	// Create scan result record
	scanResult := &models.ScanResult{
		ProjectID: projectID,
		ScanType:  options.ScanType,
		Status:    "running",
		StartedAt: time.Now(),
	}
	
	if err := s.db.Create(scanResult).Error; err != nil {
		logger.Error("Failed to create scan result record: %v", err)
	}
	
	// Get package manager for this project
	pm, exists := s.packageManager.Get(project.Type)
	if !exists {
		return nil, fmt.Errorf("unsupported package manager: %s", project.Type)
	}
	
	// Validate project
	if err := pm.ValidateProject(ctx, project.Path); err != nil {
		return nil, fmt.Errorf("project validation failed: %w", err)
	}
	
	// Parse current dependencies
	depInfo, err := pm.ParseDependencies(ctx, project.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dependencies: %w", err)
	}
	
	result := &ScanResult{
		ProjectID: projectID,
	}
	
	// Process dependencies
	if err := s.processDependencies(ctx, &project, depInfo, result); err != nil {
		logger.Error("Failed to process dependencies: %v", err)
		result.Errors = append(result.Errors, err)
	}
	
	// Update scan result
	scanResult.Status = "completed"
	scanResult.DependenciesFound = result.DependenciesFound
	scanResult.UpdatesFound = result.UpdatesFound
	scanResult.CompletedAt = &[]time.Time{time.Now()}[0]
	scanResult.Duration = time.Since(scanResult.StartedAt).Milliseconds()
	
	if len(result.Errors) > 0 {
		scanResult.Status = "failed"
		scanResult.ErrorMessage = fmt.Sprintf("%d errors occurred during scan", len(result.Errors))
	}
	
	s.db.Save(scanResult)
	
	// Update project last scan time
	project.LastScan = &[]time.Time{time.Now()}[0]
	s.db.Save(&project)
	
	logger.Info("Completed dependency scan for project ID: %d (found %d dependencies, %d updates)", 
		projectID, result.DependenciesFound, result.UpdatesFound)
	
	return result, nil
}

// processDependencies processes the parsed dependencies and checks for updates
func (s *Scanner) processDependencies(ctx context.Context, project *models.Project, depInfo *pmtypes.DependencyInfo, result *ScanResult) error {
	pm, _ := s.packageManager.Get(project.Type)
	
	// Create a channel for dependency processing
	depChan := make(chan pmtypes.DependencyEntry, len(depInfo.Dependencies))
	resultChan := make(chan *dependencyResult, len(depInfo.Dependencies))
	
	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < s.maxConcurrency; i++ {
		wg.Add(1)
		go s.dependencyWorker(ctx, &wg, project, pm, depChan, resultChan)
	}
	
	// Send dependencies to workers
	go func() {
		defer close(depChan)
		for _, dep := range depInfo.Dependencies {
			depChan <- dep
		}
	}()
	
	// Wait for workers to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	
	// Collect results
	for depResult := range resultChan {
		if depResult.Error != nil {
			result.Errors = append(result.Errors, depResult.Error)
			continue
		}
		
		result.DependenciesFound++
		
		if depResult.Dependency != nil {
			if depResult.IsNew {
				result.NewDependencies = append(result.NewDependencies, *depResult.Dependency)
			} else {
				result.UpdatedDependencies = append(result.UpdatedDependencies, *depResult.Dependency)
			}
		}
		
		if depResult.Update != nil {
			result.UpdatesFound++
			result.AvailableUpdates = append(result.AvailableUpdates, *depResult.Update)
		}
	}
	
	return nil
}

// dependencyResult holds the result of processing a single dependency
type dependencyResult struct {
	Dependency *models.Dependency
	Update     *models.Update
	IsNew      bool
	Error      error
}

// dependencyWorker processes dependencies concurrently
func (s *Scanner) dependencyWorker(ctx context.Context, wg *sync.WaitGroup, project *models.Project, pm packagemanager.PackageManager, depChan <-chan pmtypes.DependencyEntry, resultChan chan<- *dependencyResult) {
	defer wg.Done()
	
	for dep := range depChan {
		result := s.processSingleDependency(ctx, project, pm, dep)
		resultChan <- result
	}
}

// processSingleDependency processes a single dependency
func (s *Scanner) processSingleDependency(ctx context.Context, project *models.Project, pm packagemanager.PackageManager, depEntry pmtypes.DependencyEntry) *dependencyResult {
	result := &dependencyResult{}
	
	// Check if dependency already exists in database
	var existingDep models.Dependency
	err := s.db.Where("project_id = ? AND name = ?", project.ID, depEntry.Name).First(&existingDep).Error
	
	isNew := err == gorm.ErrRecordNotFound
	
	// Create or update dependency record
	dependency := &models.Dependency{
		ProjectID:       project.ID,
		Name:            depEntry.Name,
		CurrentVersion:  depEntry.ResolvedVersion,
		RequiredVersion: depEntry.Version,
		Type:            depEntry.Type,
		Registry:        depEntry.Source,
		Status:          "unknown",
		LastChecked:     &[]time.Time{time.Now()}[0],
	}
	
	if !isNew {
		dependency.ID = existingDep.ID
		dependency.CreatedAt = existingDep.CreatedAt
	}
	
	// Get latest version from registry
	latestVersion, err := pm.GetLatestVersion(ctx, depEntry.Name, nil)
	if err != nil {
		logger.Warn("Failed to get latest version for %s: %v", depEntry.Name, err)
		dependency.Status = "unknown"
	} else {
		dependency.LatestVersion = latestVersion.Version
		
		// Determine status
		if dependency.CurrentVersion == "" {
			dependency.Status = "unknown"
		} else if dependency.CurrentVersion == latestVersion.Version {
			dependency.Status = "up-to-date"
		} else {
			dependency.Status = "outdated"
			
			// Create update record
			update := &models.Update{
				DependencyID: dependency.ID,
				FromVersion:  dependency.CurrentVersion,
				ToVersion:    latestVersion.Version,
				UpdateType:   s.determineUpdateType(dependency.CurrentVersion, latestVersion.Version),
				Status:       "pending",
			}
			
			// Try to get changelog
			changelog, err := pm.GetChangelog(ctx, depEntry.Name, latestVersion.Version, nil)
			if err == nil {
				update.ChangelogURL = changelog.URL
				update.ReleaseNotes = changelog.Description
				update.BreakingChange = changelog.IsBreaking
				update.SecurityFix = changelog.SecurityFix
				
				// Perform AI analysis on changelog
				if changelog.Description != "" {
					aiAnalysis, aiErr := s.performAIAnalysis(ctx, dependency, update, changelog.Description)
					if aiErr != nil {
						logger.Warn("AI analysis failed for %s: %v", depEntry.Name, aiErr)
					} else {
						// Update fields based on AI analysis
						update.BreakingChange = aiAnalysis.HasBreakingChange || update.BreakingChange
						update.SecurityFix = len(aiAnalysis.SecurityFixes) > 0 || update.SecurityFix
						
						// Set severity based on AI risk level
						switch aiAnalysis.RiskLevel {
						case types.RiskLevelCritical:
							update.Severity = "critical"
						case types.RiskLevelHigh:
							update.Severity = "high"
						case types.RiskLevelMedium:
							update.Severity = "medium"
						default:
							update.Severity = "low"
						}
					}
				}
			}
			
			result.Update = update
		}
	}
	
	// Save dependency to database
	if isNew {
		if err := s.db.Create(dependency).Error; err != nil {
			result.Error = fmt.Errorf("failed to create dependency %s: %w", depEntry.Name, err)
			return result
		}
	} else {
		if err := s.db.Save(dependency).Error; err != nil {
			result.Error = fmt.Errorf("failed to update dependency %s: %w", depEntry.Name, err)
			return result
		}
	}
	
	// Save update if it exists
	if result.Update != nil {
		result.Update.DependencyID = dependency.ID
		if err := s.db.Create(result.Update).Error; err != nil {
			logger.Error("Failed to create update record for %s: %v", depEntry.Name, err)
		}
	}
	
	result.Dependency = dependency
	result.IsNew = isNew
	
	return result
}

// determineUpdateType determines the type of update (major, minor, patch)
func (s *Scanner) determineUpdateType(currentVersion, latestVersion string) string {
	// This is a simplified implementation
	// In practice, you'd want to use proper semantic versioning parsing
	if currentVersion == "" || latestVersion == "" {
		return "unknown"
	}
	
	// For now, just return "minor" as a placeholder
	// TODO: Implement proper semantic version comparison
	return "minor"
}

// ScanAllProjects scans all enabled projects
func (s *Scanner) ScanAllProjects(ctx context.Context, options *ScanOptions) ([]*ScanResult, error) {
	logger.Info("Starting scan of all projects")
	
	var projects []models.Project
	query := s.db.Where("enabled = ?", true)
	
	if err := query.Find(&projects).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch projects: %w", err)
	}
	
	var results []*ScanResult
	var errors []error
	
	for _, project := range projects {
		projectOptions := *options
		projectOptions.ProjectID = project.ID
		
		result, err := s.ScanProject(ctx, project.ID, &projectOptions)
		if err != nil {
			logger.Error("Failed to scan project %s: %v", project.Name, err)
			errors = append(errors, fmt.Errorf("project %s: %w", project.Name, err))
			continue
		}
		
		results = append(results, result)
	}
	
	logger.Info("Completed scan of all projects (%d scanned, %d errors)", len(results), len(errors))
	
	if len(errors) > 0 {
		return results, fmt.Errorf("scan completed with %d errors", len(errors))
	}
	
	return results, nil
}

// performAIAnalysis performs AI analysis on changelog and stores predictions
func (s *Scanner) performAIAnalysis(ctx context.Context, dependency *models.Dependency, update *models.Update, changelogText string) (*ai.ChangelogAnalysisResponse, error) {
	// Create AI analysis request
	request := &ai.ChangelogAnalysisRequest{
		PackageName:    dependency.Name,
		FromVersion:    dependency.CurrentVersion,
		ToVersion:      update.ToVersion,
		ChangelogText:  changelogText,
		ReleaseNotes:   update.ReleaseNotes,
		PackageManager: "", // Will be determined from project context
		Language:       "", // Will be determined from project context
	}
	
	// Perform AI analysis
	response, err := ai.AnalyzeChangelog(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("AI changelog analysis failed: %w", err)
	}
	
	// Store AI predictions in database
	if err := s.storeAIPredictions(dependency.ID, update.ID, response); err != nil {
		logger.Error("Failed to store AI predictions: %v", err)
	}
	
	return response, nil
}

// storeAIPredictions stores AI analysis results as predictions in the database
func (s *Scanner) storeAIPredictions(dependencyID, updateID uint, analysis *ai.ChangelogAnalysisResponse) error {
	// Store breaking change prediction
	breakingPrediction := &models.AIPrediction{
		DependencyID:   dependencyID,
		UpdateID:       updateID,
		ModelName:      "heuristic",
		ModelVersion:   "1.0.0",
		PredictionType: "breaking_change",
		Confidence:     analysis.Confidence,
		Result:         fmt.Sprintf("%t", analysis.HasBreakingChange),
		Reasoning:      analysis.Summary,
		InputData:      fmt.Sprintf(`{"changelog_length": %d, "breaking_changes": %d}`, len(analysis.Summary), len(analysis.BreakingChanges)),
	}
	
	if err := s.db.Create(breakingPrediction).Error; err != nil {
		return fmt.Errorf("failed to store breaking change prediction: %w", err)
	}
	
	// Store security risk prediction
	hasSecurityFix := len(analysis.SecurityFixes) > 0
	securityPrediction := &models.AIPrediction{
		DependencyID:   dependencyID,
		UpdateID:       updateID,
		ModelName:      "heuristic",
		ModelVersion:   "1.0.0",
		PredictionType: "security_risk",
		Confidence:     analysis.Confidence,
		Result:         fmt.Sprintf("%t", hasSecurityFix),
		Reasoning:      fmt.Sprintf("Detected %d security fixes", len(analysis.SecurityFixes)),
		InputData:      fmt.Sprintf(`{"security_fixes": %d}`, len(analysis.SecurityFixes)),
	}
	
	if err := s.db.Create(securityPrediction).Error; err != nil {
		return fmt.Errorf("failed to store security prediction: %w", err)
	}
	
	// Store risk level prediction
	riskPrediction := &models.AIPrediction{
		DependencyID:   dependencyID,
		UpdateID:       updateID,
		ModelName:      "heuristic",
		ModelVersion:   "1.0.0",
		PredictionType: "risk_level",
		Confidence:     analysis.Confidence,
		Result:         string(analysis.RiskLevel),
		Reasoning:      analysis.Summary,
		InputData:      fmt.Sprintf(`{"features": %d, "bug_fixes": %d, "deprecations": %d}`, len(analysis.NewFeatures), len(analysis.BugFixes), len(analysis.Deprecations)),
	}
	
	if err := s.db.Create(riskPrediction).Error; err != nil {
		return fmt.Errorf("failed to store risk level prediction: %w", err)
	}
	
	logger.Debug("Stored %d AI predictions for dependency %d", 3, dependencyID)
	return nil
}
