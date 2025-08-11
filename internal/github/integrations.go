package github

import (
	"context"
	"fmt"
	"time"
)

// IntegrationManager handles integrations with project management tools
type IntegrationManager struct {
	client *Client
	config *IntegrationConfig
	jira   *JiraIntegration
	linear *LinearIntegration
	asana  *AsanaIntegration
}

// NewIntegrationManager creates a new integration manager
func NewIntegrationManager(client *Client, config *IntegrationConfig) *IntegrationManager {
	im := &IntegrationManager{
		client: client,
		config: config,
	}

	// Initialize integrations based on configuration
	if config.JiraConfig != nil && config.JiraConfig.Enabled {
		im.jira = NewJiraIntegration(config.JiraConfig)
	}
	if config.LinearConfig != nil && config.LinearConfig.Enabled {
		im.linear = NewLinearIntegration(config.LinearConfig)
	}
	if config.AsanaConfig != nil && config.AsanaConfig.Enabled {
		im.asana = NewAsanaIntegration(config.AsanaConfig)
	}

	return im
}

// IntegrationConfig defines configuration for project management integrations
type IntegrationConfig struct {
	JiraConfig   *JiraConfig   `json:"jira_config"`
	LinearConfig *LinearConfig `json:"linear_config"`
	AsanaConfig  *AsanaConfig  `json:"asana_config"`
	WebhookURL   string        `json:"webhook_url"`
	APITimeout   time.Duration `json:"api_timeout"`
}

// JiraConfig defines Jira integration configuration
type JiraConfig struct {
	Enabled     bool   `json:"enabled"`
	BaseURL     string `json:"base_url"`
	Username    string `json:"username"`
	APIToken    string `json:"api_token"`
	ProjectKey  string `json:"project_key"`
	IssueType   string `json:"issue_type"`
	Priority    string `json:"priority"`
	Labels      []string `json:"labels"`
	Components  []string `json:"components"`
	AutoAssign  bool   `json:"auto_assign"`
	DefaultAssignee string `json:"default_assignee"`
}

// LinearConfig defines Linear integration configuration
type LinearConfig struct {
	Enabled     bool   `json:"enabled"`
	APIKey      string `json:"api_key"`
	TeamID      string `json:"team_id"`
	ProjectID   string `json:"project_id"`
	Priority    int    `json:"priority"`
	Labels      []string `json:"labels"`
	AutoAssign  bool   `json:"auto_assign"`
	DefaultAssignee string `json:"default_assignee"`
}

// AsanaConfig defines Asana integration configuration
type AsanaConfig struct {
	Enabled     bool   `json:"enabled"`
	APIKey      string `json:"api_key"`
	WorkspaceID string `json:"workspace_id"`
	ProjectID   string `json:"project_id"`
	Priority    string `json:"priority"`
	Tags        []string `json:"tags"`
	AutoAssign  bool   `json:"auto_assign"`
	DefaultAssignee string `json:"default_assignee"`
}

// JiraIntegration handles Jira-specific operations
type JiraIntegration struct {
	config *JiraConfig
}

// NewJiraIntegration creates a new Jira integration
func NewJiraIntegration(config *JiraConfig) *JiraIntegration {
	return &JiraIntegration{config: config}
}

// LinearIntegration handles Linear-specific operations
type LinearIntegration struct {
	config *LinearConfig
}

// NewLinearIntegration creates a new Linear integration
func NewLinearIntegration(config *LinearConfig) *LinearIntegration {
	return &LinearIntegration{config: config}
}

// AsanaIntegration handles Asana-specific operations
type AsanaIntegration struct {
	config *AsanaConfig
}

// NewAsanaIntegration creates a new Asana integration
func NewAsanaIntegration(config *AsanaConfig) *AsanaIntegration {
	return &AsanaIntegration{config: config}
}

// IssueRequest represents a request to create an issue
type IssueRequest struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Priority    string                 `json:"priority"`
	Labels      []string               `json:"labels"`
	Assignee    string                 `json:"assignee"`
	DueDate     *time.Time             `json:"due_date"`
	Metadata    map[string]interface{} `json:"metadata"`
	Repository  string                 `json:"repository"`
	PRNumber    int                    `json:"pr_number"`
	IssueType   string                 `json:"issue_type"` // "dependency_update", "security_fix", "breaking_change"
}

// IssueResponse represents the response from creating an issue
type IssueResponse struct {
	ID          string                 `json:"id"`
	Key         string                 `json:"key"`
	URL         string                 `json:"url"`
	Status      string                 `json:"status"`
	Platform    string                 `json:"platform"` // "jira", "linear", "asana"
	CreatedAt   time.Time              `json:"created_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// CreateIssueForPR creates an issue in the configured project management tool for a PR
func (im *IntegrationManager) CreateIssueForPR(ctx context.Context, pr *PullRequest, patches []*Patch) ([]*IssueResponse, error) {
	var responses []*IssueResponse

	// Generate issue request
	issueReq := im.generateIssueRequest(pr, patches)

	// Create issues in enabled integrations
	if im.jira != nil {
		response, err := im.jira.CreateIssue(ctx, issueReq)
		if err != nil {
			return responses, fmt.Errorf("failed to create Jira issue: %w", err)
		}
		responses = append(responses, response)
	}

	if im.linear != nil {
		response, err := im.linear.CreateIssue(ctx, issueReq)
		if err != nil {
			return responses, fmt.Errorf("failed to create Linear issue: %w", err)
		}
		responses = append(responses, response)
	}

	if im.asana != nil {
		response, err := im.asana.CreateTask(ctx, issueReq)
		if err != nil {
			return responses, fmt.Errorf("failed to create Asana task: %w", err)
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// generateIssueRequest generates an issue request based on PR and patches
func (im *IntegrationManager) generateIssueRequest(pr *PullRequest, patches []*Patch) *IssueRequest {
	// Determine issue type and priority based on patches
	issueType := "dependency_update"
	priority := "medium"
	
	hasSecurityFix := false
	hasBreakingChanges := false
	
	for _, patch := range patches {
		if patch.Type == "security" {
			hasSecurityFix = true
			priority = "high"
		}
		if patch.BreakingChange {
			hasBreakingChanges = true
			issueType = "breaking_change"
			priority = "high"
		}
	}
	
	if hasSecurityFix {
		issueType = "security_fix"
	}

	// Generate title and description
	title := fmt.Sprintf("Dependency Update: %s", pr.Title)
	description := im.generateIssueDescription(pr, patches, hasSecurityFix, hasBreakingChanges)

	// Generate labels
	labels := []string{"dependency-update", "automated"}
	if hasSecurityFix {
		labels = append(labels, "security")
	}
	if hasBreakingChanges {
		labels = append(labels, "breaking-change")
	}

	return &IssueRequest{
		Title:       title,
		Description: description,
		Priority:    priority,
		Labels:      labels,
		Repository:  pr.Repository,
		PRNumber:    pr.Number,
		IssueType:   issueType,
		Metadata: map[string]interface{}{
			"pr_url":           pr.URL,
			"patches_count":    len(patches),
			"security_fix":     hasSecurityFix,
			"breaking_changes": hasBreakingChanges,
		},
	}
}

// generateIssueDescription generates a detailed issue description
func (im *IntegrationManager) generateIssueDescription(pr *PullRequest, patches []*Patch, hasSecurityFix, hasBreakingChanges bool) string {
	description := fmt.Sprintf(`## Dependency Update Summary

**Pull Request:** [%s](%s)
**Repository:** %s
**Author:** %s
**Created:** %s

## Changes Overview
- **Total Patches:** %d
- **Security Fixes:** %t
- **Breaking Changes:** %t

## Patch Details
`, pr.Title, pr.URL, pr.Repository, pr.Author, pr.CreatedAt.Format("2006-01-02 15:04"), len(patches), hasSecurityFix, hasBreakingChanges)

	// Add patch details
	for i, patch := range patches {
		if i >= 5 { // Limit to first 5 patches to avoid overly long descriptions
			description += fmt.Sprintf("... and %d more patches\n", len(patches)-5)
			break
		}
		
		description += fmt.Sprintf(`
### Patch %d: %s
- **File:** %s
- **Type:** %s
- **Risk Level:** %s
- **Breaking Change:** %t
`, i+1, patch.Description, patch.File, patch.Type, patch.RiskLevel, patch.BreakingChange)
	}

	// Add action items
	description += `

## Action Items
- [ ] Review automated patches
- [ ] Validate functionality after merge
- [ ] Update documentation if needed
- [ ] Monitor for issues post-deployment
`

	if hasSecurityFix {
		description += "- [ ] Verify security vulnerabilities are resolved\n"
	}

	if hasBreakingChanges {
		description += "- [ ] Update dependent code for breaking changes\n"
		description += "- [ ] Communicate breaking changes to team\n"
	}

	return description
}

// CreateIssue creates a Jira issue
func (ji *JiraIntegration) CreateIssue(ctx context.Context, req *IssueRequest) (*IssueResponse, error) {
	// Simulate Jira API call
	issueKey := fmt.Sprintf("%s-%d", ji.config.ProjectKey, time.Now().Unix()%10000)
	
	response := &IssueResponse{
		ID:        fmt.Sprintf("jira_%d", time.Now().Unix()),
		Key:       issueKey,
		URL:       fmt.Sprintf("%s/browse/%s", ji.config.BaseURL, issueKey),
		Status:    "To Do",
		Platform:  "jira",
		CreatedAt: time.Now(),
		Metadata: map[string]interface{}{
			"project_key": ji.config.ProjectKey,
			"issue_type":  ji.config.IssueType,
		},
	}

	return response, nil
}

// UpdateIssue updates a Jira issue
func (ji *JiraIntegration) UpdateIssue(ctx context.Context, issueKey string, updates map[string]interface{}) error {
	// Simulate Jira API call to update issue
	return nil
}

// CreateIssue creates a Linear issue
func (li *LinearIntegration) CreateIssue(ctx context.Context, req *IssueRequest) (*IssueResponse, error) {
	// Simulate Linear API call
	issueID := fmt.Sprintf("LIN-%d", time.Now().Unix()%10000)
	
	response := &IssueResponse{
		ID:        fmt.Sprintf("linear_%d", time.Now().Unix()),
		Key:       issueID,
		URL:       fmt.Sprintf("https://linear.app/issue/%s", issueID),
		Status:    "Todo",
		Platform:  "linear",
		CreatedAt: time.Now(),
		Metadata: map[string]interface{}{
			"team_id":    li.config.TeamID,
			"project_id": li.config.ProjectID,
		},
	}

	return response, nil
}

// UpdateIssue updates a Linear issue
func (li *LinearIntegration) UpdateIssue(ctx context.Context, issueID string, updates map[string]interface{}) error {
	// Simulate Linear API call to update issue
	return nil
}

// CreateTask creates an Asana task
func (ai *AsanaIntegration) CreateTask(ctx context.Context, req *IssueRequest) (*IssueResponse, error) {
	// Simulate Asana API call
	taskID := fmt.Sprintf("task_%d", time.Now().Unix())
	
	response := &IssueResponse{
		ID:        taskID,
		Key:       taskID,
		URL:       fmt.Sprintf("https://app.asana.com/0/%s/%s", ai.config.ProjectID, taskID),
		Status:    "New",
		Platform:  "asana",
		CreatedAt: time.Now(),
		Metadata: map[string]interface{}{
			"workspace_id": ai.config.WorkspaceID,
			"project_id":   ai.config.ProjectID,
		},
	}

	return response, nil
}

// UpdateTask updates an Asana task
func (ai *AsanaIntegration) UpdateTask(ctx context.Context, taskID string, updates map[string]interface{}) error {
	// Simulate Asana API call to update task
	return nil
}

// SyncPRStatus syncs PR status with project management tools
func (im *IntegrationManager) SyncPRStatus(ctx context.Context, pr *PullRequest, issues []*IssueResponse) error {
	statusMap := map[string]string{
		"open":   "In Progress",
		"merged": "Done",
		"closed": "Cancelled",
	}

	newStatus, exists := statusMap[pr.State]
	if !exists {
		newStatus = "In Progress"
	}

	// Update status in all linked issues
	for _, issue := range issues {
		switch issue.Platform {
		case "jira":
			if im.jira != nil {
				err := im.jira.UpdateIssue(ctx, issue.Key, map[string]interface{}{
					"status": newStatus,
				})
				if err != nil {
					return fmt.Errorf("failed to update Jira issue %s: %w", issue.Key, err)
				}
			}
		case "linear":
			if im.linear != nil {
				err := im.linear.UpdateIssue(ctx, issue.ID, map[string]interface{}{
					"status": newStatus,
				})
				if err != nil {
					return fmt.Errorf("failed to update Linear issue %s: %w", issue.ID, err)
				}
			}
		case "asana":
			if im.asana != nil {
				err := im.asana.UpdateTask(ctx, issue.ID, map[string]interface{}{
					"status": newStatus,
				})
				if err != nil {
					return fmt.Errorf("failed to update Asana task %s: %w", issue.ID, err)
				}
			}
		}
	}

	return nil
}

// CreateEpicForBatchUpdate creates an epic/project for batch updates
func (im *IntegrationManager) CreateEpicForBatchUpdate(ctx context.Context, batchJob *BatchJob) ([]*IssueResponse, error) {
	var responses []*IssueResponse

	epicReq := &IssueRequest{
		Title:       fmt.Sprintf("Batch Dependency Update: %s", batchJob.Name),
		Description: im.generateEpicDescription(batchJob),
		Priority:    batchJob.Priority,
		Labels:      []string{"batch-update", "epic", "dependencies"},
		IssueType:   "epic",
		Metadata: map[string]interface{}{
			"batch_id":      batchJob.ID,
			"total_updates": len(batchJob.Updates),
		},
	}

	// Create epic in enabled integrations
	if im.jira != nil {
		response, err := im.jira.CreateIssue(ctx, epicReq)
		if err != nil {
			return responses, fmt.Errorf("failed to create Jira epic: %w", err)
		}
		responses = append(responses, response)
	}

	if im.linear != nil {
		response, err := im.linear.CreateIssue(ctx, epicReq)
		if err != nil {
			return responses, fmt.Errorf("failed to create Linear project: %w", err)
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// generateEpicDescription generates description for batch update epic
func (im *IntegrationManager) generateEpicDescription(batchJob *BatchJob) string {
	description := fmt.Sprintf(`## Batch Dependency Update

**Batch ID:** %s
**Total Updates:** %d
**Priority:** %s
**Status:** %s
**Created:** %s

## Update Summary
`, batchJob.ID, len(batchJob.Updates), batchJob.Priority, batchJob.Status, batchJob.CreatedAt.Format("2006-01-02 15:04"))

	// Group updates by type
	updateTypes := make(map[string]int)
	riskLevels := make(map[string]int)
	
	for _, update := range batchJob.Updates {
		updateTypes[update.UpdateType]++
		riskLevels[update.RiskLevel]++
	}

	description += "\n### By Update Type\n"
	for updateType, count := range updateTypes {
		description += fmt.Sprintf("- **%s:** %d updates\n", updateType, count)
	}

	description += "\n### By Risk Level\n"
	for riskLevel, count := range riskLevels {
		description += fmt.Sprintf("- **%s:** %d updates\n", riskLevel, count)
	}

	description += `

## Progress Tracking
- [ ] Batch job initiated
- [ ] Patches generated
- [ ] PRs created
- [ ] Reviews completed
- [ ] Updates merged
- [ ] Deployment verified

## Notes
This epic tracks the progress of a batch dependency update operation. Individual PRs will be created for each update and linked to this epic.
`

	return description
}

// NotifyCompletion sends completion notifications to project management tools
func (im *IntegrationManager) NotifyCompletion(ctx context.Context, batchJob *BatchJob, issues []*IssueResponse) error {
	completionSummary := fmt.Sprintf(`## Batch Update Completed

**Status:** %s
**Duration:** %s
**Success Rate:** %.1f%%

### Results Summary
- **Total Updates:** %d
- **Successful:** %d
- **Failed:** %d
- **Conflicts:** %d

The batch dependency update has been completed. Please review the individual PRs and verify the updates.
`, batchJob.Status, batchJob.ActualDuration, 
	float64(batchJob.Progress.SuccessfulUpdates)/float64(batchJob.Progress.TotalUpdates)*100,
	batchJob.Progress.TotalUpdates,
	batchJob.Progress.SuccessfulUpdates,
	batchJob.Progress.FailedUpdates,
	len(batchJob.Results))

	// Update all linked issues with completion summary
	for _, issue := range issues {
		updates := map[string]interface{}{
			"status":      "Done",
			"description": completionSummary,
		}

		switch issue.Platform {
		case "jira":
			if im.jira != nil {
				im.jira.UpdateIssue(ctx, issue.Key, updates)
			}
		case "linear":
			if im.linear != nil {
				im.linear.UpdateIssue(ctx, issue.ID, updates)
			}
		case "asana":
			if im.asana != nil {
				im.asana.UpdateTask(ctx, issue.ID, updates)
			}
		}
	}

	return nil
}

// GetIntegrationStatus returns the status of all integrations
func (im *IntegrationManager) GetIntegrationStatus() map[string]bool {
	status := make(map[string]bool)
	
	status["jira"] = im.jira != nil && im.jira.config.Enabled
	status["linear"] = im.linear != nil && im.linear.config.Enabled
	status["asana"] = im.asana != nil && im.asana.config.Enabled
	
	return status
}

// TestConnections tests connectivity to all enabled integrations
func (im *IntegrationManager) TestConnections(ctx context.Context) map[string]error {
	results := make(map[string]error)
	
	if im.jira != nil {
		results["jira"] = im.testJiraConnection(ctx)
	}
	
	if im.linear != nil {
		results["linear"] = im.testLinearConnection(ctx)
	}
	
	if im.asana != nil {
		results["asana"] = im.testAsanaConnection(ctx)
	}
	
	return results
}

// testJiraConnection tests Jira connectivity
func (im *IntegrationManager) testJiraConnection(ctx context.Context) error {
	// Simulate connection test
	return nil
}

// testLinearConnection tests Linear connectivity
func (im *IntegrationManager) testLinearConnection(ctx context.Context) error {
	// Simulate connection test
	return nil
}

// testAsanaConnection tests Asana connectivity
func (im *IntegrationManager) testAsanaConnection(ctx context.Context) error {
	// Simulate connection test
	return nil
}
