package services

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/database"
	"github.com/8tcapital/ai-dep-manager/internal/logger"
	"github.com/8tcapital/ai-dep-manager/internal/models"
	"gorm.io/gorm"
)

// LagService handles dependency lag analysis and resolution
type LagService struct {
	db *gorm.DB
}

// LagAnalysis represents the analysis of dependency lag
type LagAnalysis struct {
	ProjectID          uint                    `json:"project_id"`
	ProjectName        string                  `json:"project_name"`
	TotalDependencies  int                     `json:"total_dependencies"`
	LaggedDependencies int                     `json:"lagged_dependencies"`
	AverageLagDays     float64                 `json:"average_lag_days"`
	MaxLagDays         int                     `json:"max_lag_days"`
	LagDistribution    map[string]int          `json:"lag_distribution"`
	TopLaggedPackages  []LaggedPackage         `json:"top_lagged_packages"`
	RecommendedActions []LagRecommendation     `json:"recommended_actions"`
	LagTrend           []LagTrendPoint         `json:"lag_trend"`
}

// LaggedPackage represents a package with significant lag
type LaggedPackage struct {
	Name           string    `json:"name"`
	CurrentVersion string    `json:"current_version"`
	LatestVersion  string    `json:"latest_version"`
	LagDays        int       `json:"lag_days"`
	ReleasesBehind int       `json:"releases_behind"`
	SecurityRisk   bool      `json:"security_risk"`
	BreakingRisk   bool      `json:"breaking_risk"`
	LastUpdated    time.Time `json:"last_updated"`
	Reason         string    `json:"reason"`
}

// LagRecommendation represents a recommended action to reduce lag
type LagRecommendation struct {
	Priority    string   `json:"priority"`
	Action      string   `json:"action"`
	Description string   `json:"description"`
	Packages    []string `json:"packages"`
	Impact      string   `json:"impact"`
	Effort      string   `json:"effort"`
}

// LagTrendPoint represents a point in the lag trend analysis
type LagTrendPoint struct {
	Date       time.Time `json:"date"`
	AverageLag float64   `json:"average_lag"`
	MaxLag     int       `json:"max_lag"`
	Count      int       `json:"count"`
}

// LagResolutionPlan represents a plan to resolve dependency lag
type LagResolutionPlan struct {
	ProjectID     uint                    `json:"project_id"`
	GeneratedAt   time.Time               `json:"generated_at"`
	TotalPackages int                     `json:"total_packages"`
	Phases        []LagResolutionPhase    `json:"phases"`
	EstimatedTime string                  `json:"estimated_time"`
	RiskLevel     string                  `json:"risk_level"`
	Prerequisites []string                `json:"prerequisites"`
}

// LagResolutionPhase represents a phase in the lag resolution plan
type LagResolutionPhase struct {
	Phase       int                      `json:"phase"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Packages    []LagResolutionPackage   `json:"packages"`
	Order       string                   `json:"order"` // sequential, parallel
	RiskLevel   string                   `json:"risk_level"`
	Duration    string                   `json:"duration"`
}

// LagResolutionPackage represents a package in a resolution phase
type LagResolutionPackage struct {
	Name           string `json:"name"`
	CurrentVersion string `json:"current_version"`
	TargetVersion  string `json:"target_version"`
	UpdateType     string `json:"update_type"`
	Reason         string `json:"reason"`
	Dependencies   []string `json:"dependencies"`
}

// NewLagService creates a new lag service
func NewLagService() *LagService {
	return &LagService{
		db: database.GetDB(),
	}
}

// AnalyzeLag performs comprehensive lag analysis for a project or all projects
func (ls *LagService) AnalyzeLag(ctx context.Context, projectID *uint) (*LagAnalysis, error) {
	logger.Info("Analyzing dependency lag for project: %v", projectID)
	
	query := ls.db.Model(&models.Dependency{})
	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}
	
	var dependencies []models.Dependency
	if err := query.Preload("Project").Find(&dependencies).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch dependencies: %w", err)
	}
	
	if len(dependencies) == 0 {
		return &LagAnalysis{}, nil
	}
	
	analysis := &LagAnalysis{
		LagDistribution: make(map[string]int),
	}
	
	// Set project info if analyzing single project
	if projectID != nil && len(dependencies) > 0 {
		analysis.ProjectID = *projectID
		analysis.ProjectName = dependencies[0].Project.Name
	}
	
	analysis.TotalDependencies = len(dependencies)
	
	var totalLagDays int
	var maxLag int
	var laggedPackages []LaggedPackage
	
	for _, dep := range dependencies {
		lagDays := ls.calculateLagDays(dep)
		
		if lagDays > 0 {
			analysis.LaggedDependencies++
			totalLagDays += lagDays
			
			if lagDays > maxLag {
				maxLag = lagDays
			}
			
			// Categorize lag
			lagCategory := ls.categorizeLag(lagDays)
			analysis.LagDistribution[lagCategory]++
			
			// Add to lagged packages if significant
			if lagDays >= 7 { // More than a week
				laggedPkg := LaggedPackage{
					Name:           dep.Name,
					CurrentVersion: dep.CurrentVersion,
					LatestVersion:  dep.LatestVersion,
					LagDays:        lagDays,
					ReleasesBehind: ls.calculateReleasesBehind(dep),
					SecurityRisk:   ls.hasSecurityRisk(dep),
					BreakingRisk:   ls.hasBreakingRisk(dep),
					LastUpdated:    dep.UpdatedAt,
					Reason:         ls.determineLagReason(dep, lagDays),
				}
				laggedPackages = append(laggedPackages, laggedPkg)
			}
		} else {
			analysis.LagDistribution["current"]++
		}
	}
	
	if analysis.LaggedDependencies > 0 {
		analysis.AverageLagDays = float64(totalLagDays) / float64(analysis.LaggedDependencies)
	}
	analysis.MaxLagDays = maxLag
	
	// Sort lagged packages by lag days (descending)
	sort.Slice(laggedPackages, func(i, j int) bool {
		return laggedPackages[i].LagDays > laggedPackages[j].LagDays
	})
	
	// Keep top 20 most lagged packages
	if len(laggedPackages) > 20 {
		laggedPackages = laggedPackages[:20]
	}
	analysis.TopLaggedPackages = laggedPackages
	
	// Generate recommendations
	analysis.RecommendedActions = ls.generateLagRecommendations(analysis)
	
	// Generate lag trend (simplified - would use historical data)
	analysis.LagTrend = ls.generateLagTrend(projectID)
	
	return analysis, nil
}

// CreateResolutionPlan creates a plan to resolve dependency lag
func (ls *LagService) CreateResolutionPlan(ctx context.Context, projectID uint, strategy string) (*LagResolutionPlan, error) {
	logger.Info("Creating lag resolution plan for project %d with strategy: %s", projectID, strategy)
	
	// Get project dependencies
	var dependencies []models.Dependency
	if err := ls.db.Where("project_id = ?", projectID).Preload("Project").Find(&dependencies).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch dependencies: %w", err)
	}
	
	if len(dependencies) == 0 {
		return nil, fmt.Errorf("no dependencies found for project %d", projectID)
	}
	
	// Filter lagged dependencies
	var laggedDeps []models.Dependency
	for _, dep := range dependencies {
		if ls.calculateLagDays(dep) > 0 {
			laggedDeps = append(laggedDeps, dep)
		}
	}
	
	if len(laggedDeps) == 0 {
		return &LagResolutionPlan{
			ProjectID:     projectID,
			GeneratedAt:   time.Now(),
			TotalPackages: 0,
			Phases:        []LagResolutionPhase{},
			EstimatedTime: "0 minutes",
			RiskLevel:     "none",
		}, nil
	}
	
	plan := &LagResolutionPlan{
		ProjectID:     projectID,
		GeneratedAt:   time.Now(),
		TotalPackages: len(laggedDeps),
	}
	
	// Create phases based on strategy
	switch strategy {
	case "conservative":
		plan.Phases = ls.createConservativePhases(laggedDeps)
		plan.RiskLevel = "low"
	case "aggressive":
		plan.Phases = ls.createAggressivePhases(laggedDeps)
		plan.RiskLevel = "high"
	case "balanced":
		fallthrough
	default:
		plan.Phases = ls.createBalancedPhases(laggedDeps)
		plan.RiskLevel = "medium"
	}
	
	// Calculate estimated time
	plan.EstimatedTime = ls.calculateEstimatedTime(plan.Phases)
	
	// Add prerequisites
	plan.Prerequisites = ls.generatePrerequisites(laggedDeps)
	
	return plan, nil
}

// ExecuteResolutionPlan executes a lag resolution plan
func (ls *LagService) ExecuteResolutionPlan(ctx context.Context, plan *LagResolutionPlan, dryRun bool) error {
	logger.Info("Executing lag resolution plan for project %d (dry-run: %v)", plan.ProjectID, dryRun)
	
	if dryRun {
		logger.Info("Dry run mode - no actual updates will be performed")
		return ls.simulateExecution(plan)
	}
	
	// Execute phases sequentially
	for i, phase := range plan.Phases {
		logger.Info("Executing phase %d: %s", i+1, phase.Name)
		
		if err := ls.executePhase(ctx, plan.ProjectID, phase); err != nil {
			return fmt.Errorf("failed to execute phase %d: %w", i+1, err)
		}
		
		logger.Info("Phase %d completed successfully", i+1)
	}
	
	logger.Info("Lag resolution plan executed successfully")
	return nil
}

// Private helper methods

func (ls *LagService) calculateLagDays(dep models.Dependency) int {
	// Simplified calculation - in reality would compare version release dates
	if dep.CurrentVersion == dep.LatestVersion {
		return 0
	}
	
	// Simulate lag based on version difference
	// This is a placeholder - real implementation would fetch release dates
	return 15 // Placeholder
}

func (ls *LagService) categorizeLag(lagDays int) string {
	switch {
	case lagDays == 0:
		return "current"
	case lagDays <= 7:
		return "minor"
	case lagDays <= 30:
		return "moderate"
	case lagDays <= 90:
		return "significant"
	default:
		return "critical"
	}
}

func (ls *LagService) calculateReleasesBehind(dep models.Dependency) int {
	// Simplified calculation - would parse version numbers and count releases
	return 3 // Placeholder
}

func (ls *LagService) hasSecurityRisk(dep models.Dependency) bool {
	// Check if there are known security issues
	var count int64
	ls.db.Model(&models.SecurityCheck{}).Where("package_name = ? AND version = ? AND status = ?", 
		dep.Name, dep.CurrentVersion, "detected").Count(&count)
	return count > 0
}

func (ls *LagService) hasBreakingRisk(dep models.Dependency) bool {
	// Simplified check - would analyze changelog and version semantics
	return false // Placeholder
}

func (ls *LagService) determineLagReason(dep models.Dependency, lagDays int) string {
	if ls.hasSecurityRisk(dep) {
		return "Security vulnerabilities in current version"
	}
	
	if lagDays > 90 {
		return "Very outdated - multiple major releases behind"
	} else if lagDays > 30 {
		return "Outdated - missing recent improvements and fixes"
	} else if lagDays > 7 {
		return "Minor lag - recent updates available"
	}
	
	return "Up to date"
}

func (ls *LagService) generateLagRecommendations(analysis *LagAnalysis) []LagRecommendation {
	var recommendations []LagRecommendation
	
	// High priority: Security risks
	var securityPackages []string
	for _, pkg := range analysis.TopLaggedPackages {
		if pkg.SecurityRisk {
			securityPackages = append(securityPackages, pkg.Name)
		}
	}
	
	if len(securityPackages) > 0 {
		recommendations = append(recommendations, LagRecommendation{
			Priority:    "critical",
			Action:      "immediate_security_update",
			Description: "Update packages with known security vulnerabilities immediately",
			Packages:    securityPackages,
			Impact:      "high",
			Effort:      "low",
		})
	}
	
	// Medium priority: Significant lag
	var significantLagPackages []string
	for _, pkg := range analysis.TopLaggedPackages {
		if pkg.LagDays > 30 && !pkg.SecurityRisk {
			significantLagPackages = append(significantLagPackages, pkg.Name)
		}
	}
	
	if len(significantLagPackages) > 0 {
		recommendations = append(recommendations, LagRecommendation{
			Priority:    "high",
			Action:      "scheduled_update",
			Description: "Schedule updates for significantly outdated packages",
			Packages:    significantLagPackages,
			Impact:      "medium",
			Effort:      "medium",
		})
	}
	
	// Low priority: Minor lag
	if analysis.LaggedDependencies > analysis.TotalDependencies/2 {
		recommendations = append(recommendations, LagRecommendation{
			Priority:    "medium",
			Action:      "bulk_update_review",
			Description: "Review and plan bulk updates for multiple outdated packages",
			Packages:    []string{"multiple"},
			Impact:      "low",
			Effort:      "high",
		})
	}
	
	return recommendations
}

func (ls *LagService) generateLagTrend(projectID *uint) []LagTrendPoint {
	// Simplified trend generation - would use historical data
	now := time.Now()
	trend := []LagTrendPoint{
		{Date: now.AddDate(0, 0, -30), AverageLag: 20.5, MaxLag: 45, Count: 12},
		{Date: now.AddDate(0, 0, -21), AverageLag: 18.2, MaxLag: 42, Count: 11},
		{Date: now.AddDate(0, 0, -14), AverageLag: 16.8, MaxLag: 38, Count: 10},
		{Date: now.AddDate(0, 0, -7), AverageLag: 15.3, MaxLag: 35, Count: 9},
		{Date: now, AverageLag: 14.1, MaxLag: 32, Count: 8},
	}
	
	return trend
}

func (ls *LagService) createConservativePhases(dependencies []models.Dependency) []LagResolutionPhase {
	var phases []LagResolutionPhase
	
	// Phase 1: Security updates only
	var securityPackages []LagResolutionPackage
	for _, dep := range dependencies {
		if ls.hasSecurityRisk(dep) {
			securityPackages = append(securityPackages, LagResolutionPackage{
				Name:           dep.Name,
				CurrentVersion: dep.CurrentVersion,
				TargetVersion:  dep.LatestVersion,
				UpdateType:     "security",
				Reason:         "Security vulnerability",
			})
		}
	}
	
	if len(securityPackages) > 0 {
		phases = append(phases, LagResolutionPhase{
			Phase:       1,
			Name:        "Security Updates",
			Description: "Update packages with known security vulnerabilities",
			Packages:    securityPackages,
			Order:       "sequential",
			RiskLevel:   "low",
			Duration:    "30 minutes",
		})
	}
	
	// Phase 2: Patch updates
	var patchPackages []LagResolutionPackage
	for _, dep := range dependencies {
		if !ls.hasSecurityRisk(dep) && ls.calculateLagDays(dep) <= 30 {
			patchPackages = append(patchPackages, LagResolutionPackage{
				Name:           dep.Name,
				CurrentVersion: dep.CurrentVersion,
				TargetVersion:  dep.LatestVersion,
				UpdateType:     "patch",
				Reason:         "Minor updates and bug fixes",
			})
		}
	}
	
	if len(patchPackages) > 0 {
		phases = append(phases, LagResolutionPhase{
			Phase:       2,
			Name:        "Patch Updates",
			Description: "Update packages with minor version changes",
			Packages:    patchPackages,
			Order:       "sequential",
			RiskLevel:   "low",
			Duration:    "45 minutes",
		})
	}
	
	return phases
}

func (ls *LagService) createAggressivePhases(dependencies []models.Dependency) []LagResolutionPhase {
	var phases []LagResolutionPhase
	
	// Single phase: Update everything in parallel
	var allPackages []LagResolutionPackage
	for _, dep := range dependencies {
		if ls.calculateLagDays(dep) > 0 {
			updateType := "major"
			if ls.hasSecurityRisk(dep) {
				updateType = "security"
			}
			
			allPackages = append(allPackages, LagResolutionPackage{
				Name:           dep.Name,
				CurrentVersion: dep.CurrentVersion,
				TargetVersion:  dep.LatestVersion,
				UpdateType:     updateType,
				Reason:         "Aggressive update to latest version",
			})
		}
	}
	
	if len(allPackages) > 0 {
		phases = append(phases, LagResolutionPhase{
			Phase:       1,
			Name:        "Bulk Update",
			Description: "Update all outdated packages to latest versions",
			Packages:    allPackages,
			Order:       "parallel",
			RiskLevel:   "high",
			Duration:    "20 minutes",
		})
	}
	
	return phases
}

func (ls *LagService) createBalancedPhases(dependencies []models.Dependency) []LagResolutionPhase {
	// Combine conservative and aggressive approaches
	conservativePhases := ls.createConservativePhases(dependencies)
	
	// Add a final phase for remaining packages
	var remainingPackages []LagResolutionPackage
	for _, dep := range dependencies {
		if !ls.hasSecurityRisk(dep) && ls.calculateLagDays(dep) > 30 {
			remainingPackages = append(remainingPackages, LagResolutionPackage{
				Name:           dep.Name,
				CurrentVersion: dep.CurrentVersion,
				TargetVersion:  dep.LatestVersion,
				UpdateType:     "minor",
				Reason:         "Significant lag - balanced update",
			})
		}
	}
	
	if len(remainingPackages) > 0 {
		conservativePhases = append(conservativePhases, LagResolutionPhase{
			Phase:       len(conservativePhases) + 1,
			Name:        "Remaining Updates",
			Description: "Update remaining outdated packages with careful monitoring",
			Packages:    remainingPackages,
			Order:       "sequential",
			RiskLevel:   "medium",
			Duration:    "60 minutes",
		})
	}
	
	return conservativePhases
}

func (ls *LagService) calculateEstimatedTime(phases []LagResolutionPhase) string {
	totalMinutes := 0
	
	for _, phase := range phases {
		// Parse duration (simplified)
		switch phase.Duration {
		case "30 minutes":
			totalMinutes += 30
		case "45 minutes":
			totalMinutes += 45
		case "60 minutes":
			totalMinutes += 60
		case "20 minutes":
			totalMinutes += 20
		default:
			totalMinutes += 30 // Default
		}
	}
	
	if totalMinutes < 60 {
		return fmt.Sprintf("%d minutes", totalMinutes)
	}
	
	hours := totalMinutes / 60
	minutes := totalMinutes % 60
	
	if minutes == 0 {
		return fmt.Sprintf("%d hours", hours)
	}
	
	return fmt.Sprintf("%d hours %d minutes", hours, minutes)
}

func (ls *LagService) generatePrerequisites(dependencies []models.Dependency) []string {
	prerequisites := []string{
		"Ensure all tests pass before starting updates",
		"Create backup of current dependency state",
		"Verify CI/CD pipeline is functional",
	}
	
	// Add specific prerequisites based on dependencies
	hasNodePackages := false
	hasPythonPackages := false
	
	for range dependencies {
		// This would be determined by project type
		// For now, we'll add generic prerequisites
		hasNodePackages = true
		break
	}
	
	if hasNodePackages {
		prerequisites = append(prerequisites, "Ensure Node.js and npm are up to date")
	}
	
	if hasPythonPackages {
		prerequisites = append(prerequisites, "Ensure Python and pip are up to date")
	}
	
	return prerequisites
}

func (ls *LagService) simulateExecution(plan *LagResolutionPlan) error {
	logger.Info("Simulating execution of lag resolution plan")
	
	for i, phase := range plan.Phases {
		logger.Info("Phase %d: %s", i+1, phase.Name)
		logger.Info("  Packages: %d", len(phase.Packages))
		logger.Info("  Order: %s", phase.Order)
		logger.Info("  Risk Level: %s", phase.RiskLevel)
		logger.Info("  Duration: %s", phase.Duration)
		
		for _, pkg := range phase.Packages {
			logger.Info("    - %s: %s -> %s (%s)", 
				pkg.Name, pkg.CurrentVersion, pkg.TargetVersion, pkg.UpdateType)
		}
	}
	
	logger.Info("Simulation completed - estimated time: %s", plan.EstimatedTime)
	return nil
}

func (ls *LagService) executePhase(ctx context.Context, projectID uint, phase LagResolutionPhase) error {
	// This would integrate with the update service to actually perform updates
	// For now, we'll just log the execution
	
	logger.Info("Executing phase: %s", phase.Name)
	
	for _, pkg := range phase.Packages {
		logger.Info("Updating %s from %s to %s", 
			pkg.Name, pkg.CurrentVersion, pkg.TargetVersion)
		
		// Simulate update time
		time.Sleep(100 * time.Millisecond)
	}
	
	return nil
}
