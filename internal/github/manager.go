package github

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/logger"
)

// Manager represents the GitHub integration manager
type Manager struct {
	config        *Config
	client        *Client
	webhookServer *WebhookServer
	
	// Services
	repositories  *RepositoriesService
	branches      *BranchesService
	pullRequests  *PullRequestsService
	webhooks      *WebhooksService
	
	// State
	mu            sync.RWMutex
	initialized   bool
	webhookActive bool
	
	// Event handlers
	eventHandlers map[string][]EventHandler
}

// EventHandler represents a function that handles GitHub events
type EventHandler func(ctx context.Context, event *Event) error

// Event represents a GitHub event
type Event struct {
	Type       string      `json:"type"`
	Repository string      `json:"repository"`
	Payload    interface{} `json:"payload"`
	Timestamp  time.Time   `json:"timestamp"`
}

// DependencyUpdateEvent represents a dependency update event
type DependencyUpdateEvent struct {
	Repository    string              `json:"repository"`
	Dependencies  []*DependencyUpdate `json:"dependencies"`
	UpdateType    string              `json:"update_type"`
	Severity      string              `json:"severity"`
	TriggeredBy   string              `json:"triggered_by"`
	Branch        string              `json:"branch,omitempty"`
	CommitSHA     string              `json:"commit_sha,omitempty"`
}

// PatchRequest represents a request to create a patch
type PatchRequest struct {
	Repository      string              `json:"repository"`
	Dependencies    []*DependencyUpdate `json:"dependencies"`
	UpdateType      string              `json:"update_type"`
	BaseBranch      string              `json:"base_branch"`
	PatchContent    string              `json:"patch_content"`
	CommitMessage   string              `json:"commit_message"`
	PRTitle         string              `json:"pr_title"`
	PRDescription   string              `json:"pr_description"`
	Reviewers       []string            `json:"reviewers,omitempty"`
	Labels          []string            `json:"labels,omitempty"`
	AutoMerge       bool                `json:"auto_merge"`
}

// PatchResult represents the result of a patch operation
type PatchResult struct {
	Repository    string       `json:"repository"`
	BranchName    string       `json:"branch_name"`
	PullRequest   *PullRequest `json:"pull_request"`
	Success       bool         `json:"success"`
	Error         string       `json:"error,omitempty"`
	CreatedAt     time.Time    `json:"created_at"`
}

// NewManager creates a new GitHub integration manager
func NewManager(config *Config) (*Manager, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	// Create authentication provider
	authConfig := config.CreateAuthConfig()
	authProvider, err := NewAuthProvider(authConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth provider: %w", err)
	}
	
	// Create GitHub client
	client, err := NewClient(authProvider, config.BaseURL, config.Timeout, config.RateLimitRetries)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub client: %w", err)
	}
	
	// Create webhook server
	webhookServer := NewWebhookServer(config.WebhookSecret)
	
	manager := &Manager{
		config:        config,
		client:        client,
		webhookServer: webhookServer,
		repositories:  client.Repositories,
		branches:      client.Branches,
		pullRequests:  client.PullRequests,
		webhooks:      client.Webhooks,
		eventHandlers: make(map[string][]EventHandler),
	}
	
	// Register default webhook handlers
	manager.registerDefaultHandlers()
	
	return manager, nil
}

// Initialize initializes the GitHub integration
func (m *Manager) Initialize(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.initialized {
		return nil
	}
	
	logger.Info("Initializing GitHub integration")
	
	// Test authentication
	if err := m.testAuthentication(ctx); err != nil {
		return fmt.Errorf("authentication test failed: %w", err)
	}
	
	// Setup webhooks for enabled repositories
	if err := m.setupWebhooks(ctx); err != nil {
		logger.Warn("Failed to setup webhooks: %v", err)
		// Don't fail initialization if webhooks fail
	}
	
	m.initialized = true
	logger.Info("GitHub integration initialized successfully")
	
	return nil
}

// testAuthentication tests the GitHub authentication
func (m *Manager) testAuthentication(ctx context.Context) error {
	logger.Info("Testing GitHub authentication")
	
	// Try to get the authenticated user
	user, err := m.client.GetAuthenticatedUser(ctx)
	if err != nil {
		return fmt.Errorf("failed to get authenticated user: %w", err)
	}
	
	logger.Info("Authenticated as GitHub user: %s", user.Login)
	return nil
}

// setupWebhooks sets up webhooks for all enabled repositories
func (m *Manager) setupWebhooks(ctx context.Context) error {
	if m.config.WebhookURL == "" {
		logger.Info("No webhook URL configured, skipping webhook setup")
		return nil
	}
	
	repos := m.config.GetWebhookEnabledRepositories()
	if len(repos) == 0 {
		logger.Info("No repositories configured for webhooks")
		return nil
	}
	
	logger.Info("Setting up webhooks for %d repositories", len(repos))
	
	for _, repo := range repos {
		if err := m.setupRepositoryWebhook(ctx, repo.Owner, repo.Name); err != nil {
			logger.Error("Failed to setup webhook for %s/%s: %v", repo.Owner, repo.Name, err)
			// Continue with other repositories
		}
	}
	
	return nil
}

// setupRepositoryWebhook sets up a webhook for a specific repository
func (m *Manager) setupRepositoryWebhook(ctx context.Context, owner, name string) error {
	logger.Info("Setting up webhook for repository %s/%s", owner, name)
	
	webhook, err := m.webhooks.EnsureDependencyWebhook(ctx, owner, name, m.config.WebhookURL, m.config.WebhookSecret)
	if err != nil {
		return fmt.Errorf("failed to ensure webhook: %w", err)
	}
	
	logger.Info("Webhook configured for %s/%s (ID: %d)", owner, name, webhook.ID)
	return nil
}

// StartWebhookServer starts the webhook server
func (m *Manager) StartWebhookServer() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.webhookActive {
		return fmt.Errorf("webhook server is already running")
	}
	
	addr := fmt.Sprintf(":%d", m.config.WebhookPort)
	
	go func() {
		if err := m.webhookServer.Start(addr); err != nil {
			logger.Error("Webhook server failed: %v", err)
		}
	}()
	
	m.webhookActive = true
	logger.Info("Webhook server started on %s", addr)
	
	return nil
}

// StopWebhookServer stops the webhook server
func (m *Manager) StopWebhookServer(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if !m.webhookActive {
		return nil
	}
	
	if err := m.webhookServer.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop webhook server: %w", err)
	}
	
	m.webhookActive = false
	logger.Info("Webhook server stopped")
	
	return nil
}

// RegisterEventHandler registers an event handler for a specific event type
func (m *Manager) RegisterEventHandler(eventType string, handler EventHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.eventHandlers[eventType] = append(m.eventHandlers[eventType], handler)
	logger.Debug("Registered event handler for event type: %s", eventType)
}

// registerDefaultHandlers registers default webhook handlers
func (m *Manager) registerDefaultHandlers() {
	// Push event handler
	m.webhookServer.RegisterHandler("push", func(eventType string, payload interface{}) error {
		pushPayload, ok := payload.(*PushPayload)
		if !ok {
			return fmt.Errorf("invalid push payload type")
		}
		
		event := &Event{
			Type:       eventType,
			Repository: pushPayload.Repository.FullName,
			Payload:    pushPayload,
			Timestamp:  time.Now(),
		}
		
		return m.handleEvent(context.Background(), event)
	})
	
	// Pull request event handler
	m.webhookServer.RegisterHandler("pull_request", func(eventType string, payload interface{}) error {
		prPayload, ok := payload.(*PullRequestPayload)
		if !ok {
			return fmt.Errorf("invalid pull request payload type")
		}
		
		event := &Event{
			Type:       eventType,
			Repository: prPayload.Repository.FullName,
			Payload:    prPayload,
			Timestamp:  time.Now(),
		}
		
		return m.handleEvent(context.Background(), event)
	})
	
	// Dependency update event handler
	m.webhookServer.RegisterHandler("repository_vulnerability_alert", func(eventType string, payload interface{}) error {
		depPayload, ok := payload.(*DependencyUpdatePayload)
		if !ok {
			return fmt.Errorf("invalid dependency payload type")
		}
		
		event := &Event{
			Type:       "dependency_update",
			Repository: depPayload.Repository.FullName,
			Payload:    depPayload,
			Timestamp:  time.Now(),
		}
		
		return m.handleEvent(context.Background(), event)
	})
}

// handleEvent handles a GitHub event
func (m *Manager) handleEvent(ctx context.Context, event *Event) error {
	logger.Debug("Handling GitHub event: %s for repository: %s", event.Type, event.Repository)
	
	// Get handlers for this event type
	m.mu.RLock()
	handlers := m.eventHandlers[event.Type]
	m.mu.RUnlock()
	
	// Execute all handlers
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			logger.Error("Event handler failed for %s: %v", event.Type, err)
			// Continue with other handlers
		}
	}
	
	return nil
}

// CreatePatch creates a patch for dependency updates
func (m *Manager) CreatePatch(ctx context.Context, request *PatchRequest) (*PatchResult, error) {
	logger.Info("Creating patch for repository: %s", request.Repository)
	
	// Parse repository name
	parts := splitRepositoryName(request.Repository)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository name: %s", request.Repository)
	}
	owner, repo := parts[0], parts[1]
	
	// Check if repository is enabled
	if !m.config.IsRepositoryEnabled(owner, repo) {
		return &PatchResult{
			Repository: request.Repository,
			Success:    false,
			Error:      "repository not enabled for GitHub integration",
			CreatedAt:  time.Now(),
		}, nil
	}
	
	// Create patch branch
	branchName := m.generatePatchBranchName(request.UpdateType, request.Dependencies)
	
	branch, err := m.branches.CreatePatchBranch(ctx, owner, repo, branchName, request.BaseBranch)
	if err != nil {
		return &PatchResult{
			Repository: request.Repository,
			Success:    false,
			Error:      fmt.Sprintf("failed to create branch: %v", err),
			CreatedAt:  time.Now(),
		}, nil
	}
	
	logger.Info("Created patch branch: %s", branch.Name)
	
	// Apply patch content (this would be implemented in Phase 2)
	// For now, we'll create an empty commit
	if err := m.createPatchCommit(ctx, owner, repo, branchName, request); err != nil {
		return &PatchResult{
			Repository: request.Repository,
			BranchName: branchName,
			Success:    false,
			Error:      fmt.Sprintf("failed to create patch commit: %v", err),
			CreatedAt:  time.Now(),
		}, nil
	}
	
	// Create pull request
	prRequest := &PullRequestRequest{
		Title:               request.PRTitle,
		Body:                request.PRDescription,
		Head:                branchName,
		Base:                request.BaseBranch,
		MaintainerCanModify: true,
		Draft:               false,
	}
	
	pr, err := m.pullRequests.Create(ctx, owner, repo, prRequest)
	if err != nil {
		return &PatchResult{
			Repository: request.Repository,
			BranchName: branchName,
			Success:    false,
			Error:      fmt.Sprintf("failed to create pull request: %v", err),
			CreatedAt:  time.Now(),
		}, nil
	}
	
	logger.Info("Created pull request: #%d", pr.Number)
	
	// Add labels
	labels := request.Labels
	if len(labels) == 0 {
		labels = m.config.GetLabels(owner, repo)
	}
	
	if len(labels) > 0 {
		if err := m.pullRequests.AddLabels(ctx, owner, repo, pr.Number, labels); err != nil {
			logger.Warn("Failed to add labels to PR #%d: %v", pr.Number, err)
		}
	}
	
	// Request reviewers
	reviewers := request.Reviewers
	if len(reviewers) == 0 {
		reviewers = m.config.GetReviewers(owner, repo)
	}
	
	if len(reviewers) > 0 {
		if err := m.pullRequests.RequestReviewers(ctx, owner, repo, pr.Number, reviewers, nil); err != nil {
			logger.Warn("Failed to request reviewers for PR #%d: %v", pr.Number, err)
		}
	}
	
	result := &PatchResult{
		Repository:  request.Repository,
		BranchName:  branchName,
		PullRequest: pr,
		Success:     true,
		CreatedAt:   time.Now(),
	}
	
	logger.Info("Successfully created patch for %s: PR #%d", request.Repository, pr.Number)
	
	return result, nil
}

// generatePatchBranchName generates a branch name for a patch
func (m *Manager) generatePatchBranchName(updateType string, dependencies []*DependencyUpdate) string {
	timestamp := time.Now().Format("20060102-150405")
	
	if len(dependencies) == 1 {
		dep := dependencies[0]
		return fmt.Sprintf("%s/%s/%s-to-%s-%s", 
			m.config.PatchBranchPrefix, updateType, dep.Name, dep.LatestVersion, timestamp)
	}
	
	return fmt.Sprintf("%s/%s/multi-dependency-%s", 
		m.config.PatchBranchPrefix, updateType, timestamp)
}

// createPatchCommit creates a commit with the patch content
func (m *Manager) createPatchCommit(ctx context.Context, owner, repo, branch string, request *PatchRequest) error {
	// This is a placeholder implementation
	// In Phase 2, this would apply the actual patch content
	logger.Info("Creating patch commit for %s/%s on branch %s", owner, repo, branch)
	
	// For now, just log that we would create a commit
	logger.Debug("Patch commit would contain: %s", request.CommitMessage)
	
	return nil
}

// GetRepositoryStatus gets the status of a repository
func (m *Manager) GetRepositoryStatus(ctx context.Context, owner, repo string) (*RepositoryStatus, error) {
	// Get repository info
	repository, err := m.repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}
	
	// Get open patch PRs
	patchPRs, err := m.pullRequests.ListPatchPRs(ctx, owner, repo, m.config.PatchBranchPrefix)
	if err != nil {
		logger.Warn("Failed to get patch PRs for %s/%s: %v", owner, repo, err)
		patchPRs = []*PullRequest{} // Continue with empty list
	}
	
	// Get patch branches
	patchBranches, err := m.branches.ListPatchBranches(ctx, owner, repo, m.config.PatchBranchPrefix)
	if err != nil {
		logger.Warn("Failed to get patch branches for %s/%s: %v", owner, repo, err)
		patchBranches = []*Branch{} // Continue with empty list
	}
	
	status := &RepositoryStatus{
		Repository:    repository,
		Enabled:       m.config.IsRepositoryEnabled(owner, repo),
		WebhookActive: m.config.IsWebhookEnabled(owner, repo),
		PatchPRs:      patchPRs,
		PatchBranches: patchBranches,
		LastChecked:   time.Now(),
	}
	
	return status, nil
}

// RepositoryStatus represents the status of a repository
type RepositoryStatus struct {
	Repository    *Repository    `json:"repository"`
	Enabled       bool           `json:"enabled"`
	WebhookActive bool           `json:"webhook_active"`
	PatchPRs      []*PullRequest `json:"patch_prs"`
	PatchBranches []*Branch      `json:"patch_branches"`
	LastChecked   time.Time      `json:"last_checked"`
}

// CleanupOldBranches cleans up old patch branches
func (m *Manager) CleanupOldBranches(ctx context.Context) error {
	if !m.config.CleanupOldBranches {
		return nil
	}
	
	logger.Info("Starting cleanup of old patch branches")
	
	repos := m.config.GetEnabledRepositories()
	for _, repo := range repos {
		if err := m.cleanupRepositoryBranches(ctx, repo.Owner, repo.Name); err != nil {
			logger.Error("Failed to cleanup branches for %s/%s: %v", repo.Owner, repo.Name, err)
			// Continue with other repositories
		}
	}
	
	logger.Info("Completed cleanup of old patch branches")
	return nil
}

// cleanupRepositoryBranches cleans up old branches for a specific repository
func (m *Manager) cleanupRepositoryBranches(ctx context.Context, owner, repo string) error {
	branches, err := m.branches.ListPatchBranches(ctx, owner, repo, m.config.PatchBranchPrefix)
	if err != nil {
		return fmt.Errorf("failed to list patch branches: %w", err)
	}
	
	cutoff := time.Now().Add(-m.config.BranchMaxAge)
	
	for _, branch := range branches {
		if branch.Commit.Commit.Author.Date.Before(cutoff) {
			// Check if branch has an open PR
			prs, err := m.pullRequests.ListForBranch(ctx, owner, repo, branch.Name)
			if err != nil {
				logger.Warn("Failed to check PRs for branch %s: %v", branch.Name, err)
				continue
			}
			
			hasOpenPR := false
			for _, pr := range prs {
				if pr.State == "open" {
					hasOpenPR = true
					break
				}
			}
			
			if !hasOpenPR {
				logger.Info("Deleting old patch branch: %s (age: %v)", branch.Name, time.Since(branch.Commit.Commit.Author.Date))
				if err := m.branches.Delete(ctx, owner, repo, branch.Name); err != nil {
					logger.Error("Failed to delete branch %s: %v", branch.Name, err)
				}
			}
		}
	}
	
	return nil
}

// Shutdown shuts down the GitHub integration manager
func (m *Manager) Shutdown(ctx context.Context) error {
	logger.Info("Shutting down GitHub integration manager")
	
	// Stop webhook server
	if err := m.StopWebhookServer(ctx); err != nil {
		logger.Error("Failed to stop webhook server: %v", err)
	}
	
	m.mu.Lock()
	m.initialized = false
	m.mu.Unlock()
	
	logger.Info("GitHub integration manager shut down")
	return nil
}

// splitRepositoryName splits a repository name into owner and name
func splitRepositoryName(fullName string) []string {
	parts := make([]string, 0, 2)
	for _, part := range []string{fullName} {
		if len(part) > 0 {
			subParts := []string{}
			current := ""
			for _, char := range part {
				if char == '/' {
					if current != "" {
						subParts = append(subParts, current)
						current = ""
					}
				} else {
					current += string(char)
				}
			}
			if current != "" {
				subParts = append(subParts, current)
			}
			parts = append(parts, subParts...)
		}
	}
	return parts
}
