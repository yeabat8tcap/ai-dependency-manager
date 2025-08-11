package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/config"
	"github.com/8tcapital/ai-dep-manager/internal/logger"
	"github.com/8tcapital/ai-dep-manager/internal/models"
	"github.com/8tcapital/ai-dep-manager/internal/scanner"
	"github.com/8tcapital/ai-dep-manager/internal/services"
)

// Agent represents the background dependency management agent
type Agent struct {
	config         *config.Config
	projectService *services.ProjectService
	scanner        *scanner.Scanner
	updateService  *services.UpdateService
	
	// Agent state
	running        bool
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	
	// Scheduling
	scanTicker     *time.Ticker
	updateTicker   *time.Ticker
	
	// Statistics
	stats          *AgentStats
	statsMu        sync.RWMutex
}

// AgentStats tracks agent performance and activity
type AgentStats struct {
	StartTime           time.Time     `json:"start_time"`
	LastScanTime        *time.Time    `json:"last_scan_time,omitempty"`
	LastUpdateTime      *time.Time    `json:"last_update_time,omitempty"`
	TotalScans          int64         `json:"total_scans"`
	TotalUpdates        int64         `json:"total_updates"`
	TotalErrors         int64         `json:"total_errors"`
	ProjectsMonitored   int           `json:"projects_monitored"`
	PendingUpdates      int           `json:"pending_updates"`
	SecurityUpdates     int           `json:"security_updates"`
	Uptime              time.Duration `json:"uptime"`
}

// AgentConfig holds agent-specific configuration
type AgentConfig struct {
	ScanInterval    time.Duration `json:"scan_interval"`
	UpdateInterval  time.Duration `json:"update_interval"`
	AutoUpdate      bool          `json:"auto_update"`
	SecurityOnly    bool          `json:"security_only"`
	MaxConcurrency  int           `json:"max_concurrency"`
	RetryAttempts   int           `json:"retry_attempts"`
	RetryDelay      time.Duration `json:"retry_delay"`
}

// NewAgent creates a new background agent
func NewAgent(cfg *config.Config) *Agent {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Agent{
		config:         cfg,
		projectService: services.NewProjectService(),
		scanner:        scanner.NewScanner(10), // 10 concurrent workers
		updateService:  services.NewUpdateService(),
		ctx:            ctx,
		cancel:         cancel,
		stats: &AgentStats{
			StartTime: time.Now(),
		},
	}
}

// Start starts the background agent
func (a *Agent) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if a.running {
		return fmt.Errorf("agent is already running")
	}
	
	logger.Info("Starting AI Dependency Manager Agent")
	
	// Initialize agent configuration
	agentConfig := a.getAgentConfig()
	
	// Start scan ticker
	if agentConfig.ScanInterval > 0 {
		a.scanTicker = time.NewTicker(agentConfig.ScanInterval)
		go a.scanLoop()
		logger.Info("Scan scheduler started with interval: %s", agentConfig.ScanInterval)
	}
	
	// Start update ticker
	if agentConfig.UpdateInterval > 0 && agentConfig.AutoUpdate {
		a.updateTicker = time.NewTicker(agentConfig.UpdateInterval)
		go a.updateLoop()
		logger.Info("Update scheduler started with interval: %s", agentConfig.UpdateInterval)
	}
	
	// Start monitoring goroutine
	go a.monitoringLoop()
	
	a.running = true
	a.stats.StartTime = time.Now()
	
	logger.Info("AI Dependency Manager Agent started successfully")
	return nil
}

// Stop stops the background agent gracefully
func (a *Agent) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if !a.running {
		return fmt.Errorf("agent is not running")
	}
	
	logger.Info("Stopping AI Dependency Manager Agent")
	
	// Stop tickers
	if a.scanTicker != nil {
		a.scanTicker.Stop()
		a.scanTicker = nil
	}
	
	if a.updateTicker != nil {
		a.updateTicker.Stop()
		a.updateTicker = nil
	}
	
	// Cancel context to stop all goroutines
	a.cancel()
	
	// Wait a moment for graceful shutdown
	time.Sleep(1 * time.Second)
	
	a.running = false
	
	logger.Info("AI Dependency Manager Agent stopped")
	return nil
}

// Restart restarts the agent
func (a *Agent) Restart() error {
	logger.Info("Restarting AI Dependency Manager Agent")
	
	if err := a.Stop(); err != nil {
		logger.Warn("Error stopping agent during restart: %v", err)
	}
	
	// Wait a moment before restarting
	time.Sleep(2 * time.Second)
	
	return a.Start()
}

// IsRunning returns whether the agent is currently running
func (a *Agent) IsRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.running
}

// GetStats returns current agent statistics
func (a *Agent) GetStats() *AgentStats {
	a.statsMu.RLock()
	defer a.statsMu.RUnlock()
	
	stats := *a.stats
	stats.Uptime = time.Since(a.stats.StartTime)
	
	return &stats
}

// TriggerScan manually triggers a scan of all projects
func (a *Agent) TriggerScan() error {
	if !a.IsRunning() {
		return fmt.Errorf("agent is not running")
	}
	
	logger.Info("Manual scan triggered")
	go a.performScan()
	
	return nil
}

// TriggerUpdate manually triggers updates for all projects
func (a *Agent) TriggerUpdate() error {
	if !a.IsRunning() {
		return fmt.Errorf("agent is not running")
	}
	
	logger.Info("Manual update triggered")
	go a.performUpdate()
	
	return nil
}

// Private methods

func (a *Agent) getAgentConfig() *AgentConfig {
	// Get configuration from main config with defaults
	scanInterval := 1 * time.Hour
	updateInterval := 24 * time.Hour
	
	if a.config.Agent.ScanInterval != "" {
		if duration, err := time.ParseDuration(a.config.Agent.ScanInterval); err == nil {
			scanInterval = duration
		}
	}
	
	if a.config.Agent.ScanInterval != "" {
		if duration, err := time.ParseDuration(a.config.Agent.ScanInterval); err == nil {
			updateInterval = duration
		}
	}
	
	return &AgentConfig{
		ScanInterval:    scanInterval,
		UpdateInterval:  updateInterval,
		AutoUpdate:      a.config.Agent.AutoUpdateLevel != "none",
		SecurityOnly:    a.config.Agent.AutoUpdateLevel == "security",
		MaxConcurrency:  a.config.Agent.MaxConcurrency,
		RetryAttempts:   3,
		RetryDelay:      30 * time.Second,
	}
}

func (a *Agent) scanLoop() {
	logger.Debug("Starting scan loop")
	
	for {
		select {
		case <-a.ctx.Done():
			logger.Debug("Scan loop stopped")
			return
		case <-a.scanTicker.C:
			logger.Debug("Scheduled scan triggered")
			a.performScan()
		}
	}
}

func (a *Agent) updateLoop() {
	logger.Debug("Starting update loop")
	
	for {
		select {
		case <-a.ctx.Done():
			logger.Debug("Update loop stopped")
			return
		case <-a.updateTicker.C:
			logger.Debug("Scheduled update triggered")
			a.performUpdate()
		}
	}
}

func (a *Agent) monitoringLoop() {
	logger.Debug("Starting monitoring loop")
	
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-a.ctx.Done():
			logger.Debug("Monitoring loop stopped")
			return
		case <-ticker.C:
			a.updateStats()
		}
	}
}

func (a *Agent) performScan() {
	logger.Info("Performing scheduled dependency scan")
	
	defer func() {
		a.statsMu.Lock()
		now := time.Now()
		a.stats.LastScanTime = &now
		a.stats.TotalScans++
		a.statsMu.Unlock()
	}()
	
	// Get all enabled projects
	enabled := true
	projects, err := a.projectService.ListProjects(a.ctx, &enabled)
	if err != nil {
		logger.Error("Failed to list projects for scanning: %v", err)
		a.incrementErrorCount()
		return
	}
	
	if len(projects) == 0 {
		logger.Debug("No enabled projects found for scanning")
		return
	}
	
	logger.Info("Scanning %d enabled projects", len(projects))
	
	// Scan each project
	agentConfig := a.getAgentConfig()
	semaphore := make(chan struct{}, agentConfig.MaxConcurrency)
	var wg sync.WaitGroup
	
	for _, project := range projects {
		wg.Add(1)
		go func(proj models.Project) {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			a.scanProject(proj)
		}(project)
	}
	
	wg.Wait()
	logger.Info("Scheduled scan completed")
}

func (a *Agent) scanProject(project models.Project) {
	logger.Debug("Scanning project: %s", project.Name)
	
	options := &scanner.ScanOptions{
		ProjectID:    project.ID,
		ScanType:     "scheduled",
		ForceRefresh: false,
		Timeout:      5 * time.Minute,
	}
	
	result, err := a.scanner.ScanProject(a.ctx, project.ID, options)
	if err != nil {
		logger.Error("Failed to scan project %s: %v", project.Name, err)
		a.incrementErrorCount()
		return
	}
	
	logger.Debug("Scan completed for %s: %d dependencies, %d updates available", 
		project.Name, result.DependenciesFound, result.UpdatesFound)
}

func (a *Agent) performUpdate() {
	logger.Info("Performing scheduled dependency updates")
	
	defer func() {
		a.statsMu.Lock()
		now := time.Now()
		a.stats.LastUpdateTime = &now
		a.stats.TotalUpdates++
		a.statsMu.Unlock()
	}()
	
	// Get all enabled projects
	enabled := true
	projects, err := a.projectService.ListProjects(a.ctx, &enabled)
	if err != nil {
		logger.Error("Failed to list projects for updating: %v", err)
		a.incrementErrorCount()
		return
	}
	
	if len(projects) == 0 {
		logger.Debug("No enabled projects found for updating")
		return
	}
	
	logger.Info("Checking updates for %d enabled projects", len(projects))
	
	agentConfig := a.getAgentConfig()
	
	// Process each project
	for _, project := range projects {
		a.updateProject(project, agentConfig)
	}
	
	logger.Info("Scheduled update check completed")
}

func (a *Agent) updateProject(project models.Project, agentConfig *AgentConfig) {
	logger.Debug("Checking updates for project: %s", project.Name)
	
	// Create update options based on agent configuration
	options := &services.UpdateOptions{
		ProjectID:    project.ID,
		SecurityOnly: agentConfig.SecurityOnly,
		AutoApprove:  true, // Agent mode auto-approves based on configuration
		SkipBreaking: true, // Agent skips breaking changes by default
		DryRun:       false,
	}
	
	// Generate update plan
	plan, err := a.updateService.GenerateUpdatePlan(a.ctx, options)
	if err != nil {
		logger.Error("Failed to generate update plan for %s: %v", project.Name, err)
		a.incrementErrorCount()
		return
	}
	
	if plan.TotalUpdates == 0 {
		logger.Debug("No updates available for project: %s", project.Name)
		return
	}
	
	// Filter updates based on agent policy
	if a.shouldApplyUpdates(plan, agentConfig) {
		logger.Info("Applying %d updates for project: %s", plan.TotalUpdates, project.Name)
		
		result, err := a.updateService.ApplyUpdates(a.ctx, plan, options)
		if err != nil {
			logger.Error("Failed to apply updates for %s: %v", project.Name, err)
			a.incrementErrorCount()
			return
		}
		
		logger.Info("Applied %d updates successfully for %s", 
			len(result.Successful), project.Name)
		
		if len(result.Failed) > 0 {
			logger.Warn("Failed to apply %d updates for %s", 
				len(result.Failed), project.Name)
		}
	} else {
		logger.Info("Updates available for %s but not applying due to agent policy", 
			project.Name)
	}
}

func (a *Agent) shouldApplyUpdates(plan *services.UpdatePlan, agentConfig *AgentConfig) bool {
	// Only apply security updates if security-only mode
	if agentConfig.SecurityOnly {
		return plan.RiskSummary.SecurityUpdates > 0
	}
	
	// Don't apply if there are critical risk updates
	if plan.RiskSummary.CriticalRisk > 0 {
		logger.Debug("Skipping updates due to critical risk level")
		return false
	}
	
	// Don't apply if there are too many breaking changes
	if plan.RiskSummary.BreakingChanges > 2 {
		logger.Debug("Skipping updates due to too many breaking changes")
		return false
	}
	
	// Apply low and medium risk updates
	return plan.RiskSummary.OverallRisk == "low" || plan.RiskSummary.OverallRisk == "medium"
}

func (a *Agent) updateStats() {
	a.statsMu.Lock()
	defer a.statsMu.Unlock()
	
	// Update project count
	enabled := true
	projects, err := a.projectService.ListProjects(a.ctx, &enabled)
	if err == nil {
		a.stats.ProjectsMonitored = len(projects)
	}
	
	// TODO: Update pending updates count and security updates count
	// This would require querying the database for pending updates
}

func (a *Agent) incrementErrorCount() {
	a.statsMu.Lock()
	defer a.statsMu.Unlock()
	a.stats.TotalErrors++
}
