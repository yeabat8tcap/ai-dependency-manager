package services

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/database"
	"github.com/8tcapital/ai-dep-manager/internal/logger"
	"github.com/8tcapital/ai-dep-manager/internal/models"
	"gorm.io/gorm"
)

// PolicyService handles custom update policies
type PolicyService struct {
	db *gorm.DB
}

// UpdatePolicy represents a custom update policy
type UpdatePolicy struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null"`
	Description string    `json:"description"`
	ProjectID   *uint     `json:"project_id,omitempty"` // nil for global policies
	Priority    int       `json:"priority"`             // Higher number = higher priority
	Enabled     bool      `json:"enabled" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Policy conditions
	Conditions PolicyConditions `json:"conditions" gorm:"type:text"`

	// Policy actions
	Actions PolicyActions `json:"actions" gorm:"type:text"`

	// Metadata
	Author   string `json:"author"`
	Version  string `json:"version" gorm:"default:'1.0'"`
	Tags     string `json:"tags"` // Comma-separated tags
}

// PolicyConditions defines when a policy should be applied
type PolicyConditions struct {
	// Package matching
	PackageNames    []string `json:"package_names,omitempty"`    // Exact package names
	PackagePatterns []string `json:"package_patterns,omitempty"` // Regex patterns
	PackageTypes    []string `json:"package_types,omitempty"`    // npm, pip, maven, etc.

	// Version matching
	CurrentVersionPattern string `json:"current_version_pattern,omitempty"` // Regex for current version
	TargetVersionPattern  string `json:"target_version_pattern,omitempty"`  // Regex for target version
	VersionChangeType     string `json:"version_change_type,omitempty"`     // major, minor, patch

	// Update characteristics
	UpdateTypes     []string `json:"update_types,omitempty"`     // security, breaking, feature, etc.
	SecurityRisk    *bool    `json:"security_risk,omitempty"`    // true/false/nil (any)
	BreakingChange  *bool    `json:"breaking_change,omitempty"`  // true/false/nil (any)
	RiskScoreMin    *float64 `json:"risk_score_min,omitempty"`   // Minimum risk score
	RiskScoreMax    *float64 `json:"risk_score_max,omitempty"`   // Maximum risk score
	ConfidenceMin   *float64 `json:"confidence_min,omitempty"`   // Minimum AI confidence
	ConfidenceMax   *float64 `json:"confidence_max,omitempty"`   // Maximum AI confidence

	// Timing conditions
	UpdateLagDaysMin *int       `json:"update_lag_days_min,omitempty"` // Minimum days since release
	UpdateLagDaysMax *int       `json:"update_lag_days_max,omitempty"` // Maximum days since release
	TimeWindow       TimeWindow `json:"time_window,omitempty"`         // When policy is active

	// Project conditions
	ProjectNames    []string `json:"project_names,omitempty"`    // Specific projects
	ProjectPatterns []string `json:"project_patterns,omitempty"` // Project name patterns
	ProjectTags     []string `json:"project_tags,omitempty"`     // Project tags

	// Custom conditions (advanced)
	CustomScript string `json:"custom_script,omitempty"` // JavaScript-like expression
}

// PolicyActions defines what actions to take when conditions are met
type PolicyActions struct {
	// Update behavior
	AutoUpdate      *bool   `json:"auto_update,omitempty"`       // Allow automatic updates
	RequireApproval *bool   `json:"require_approval,omitempty"`  // Require manual approval
	BlockUpdate     *bool   `json:"block_update,omitempty"`      // Block the update
	UpdateStrategy  string  `json:"update_strategy,omitempty"`   // conservative, balanced, aggressive
	MaxRiskScore    float64 `json:"max_risk_score,omitempty"`    // Maximum allowed risk score

	// Scheduling
	Schedule       string `json:"schedule,omitempty"`        // Cron expression for when to apply
	DelayDays      int    `json:"delay_days,omitempty"`      // Delay update by N days
	BatchSize      int    `json:"batch_size,omitempty"`      // Update in batches of N packages
	BatchInterval  string `json:"batch_interval,omitempty"`  // Time between batches

	// Notifications
	NotifyChannels []string `json:"notify_channels,omitempty"` // email, slack, webhook
	NotifyLevel    string   `json:"notify_level,omitempty"`    // info, warning, critical
	CustomMessage  string   `json:"custom_message,omitempty"`  // Custom notification message

	// Testing and validation
	RunTests       *bool    `json:"run_tests,omitempty"`       // Run tests before/after update
	TestCommands   []string `json:"test_commands,omitempty"`   // Custom test commands
	RollbackOnFail *bool    `json:"rollback_on_fail,omitempty"` // Auto-rollback on test failure

	// Custom actions
	PreUpdateHooks  []string `json:"pre_update_hooks,omitempty"`  // Commands to run before update
	PostUpdateHooks []string `json:"post_update_hooks,omitempty"` // Commands to run after update
	CustomScript    string   `json:"custom_script,omitempty"`     // Custom action script
}

// TimeWindow defines when a policy is active
type TimeWindow struct {
	DaysOfWeek []string `json:"days_of_week,omitempty"` // monday, tuesday, etc.
	StartTime  string   `json:"start_time,omitempty"`   // HH:MM format
	EndTime    string   `json:"end_time,omitempty"`     // HH:MM format
	Timezone   string   `json:"timezone,omitempty"`     // IANA timezone
}

// PolicyMatch represents a policy match result
type PolicyMatch struct {
	Policy     UpdatePolicy `json:"policy"`
	MatchScore float64      `json:"match_score"` // 0-1, how well conditions matched
	Reason     string       `json:"reason"`      // Why this policy matched
}

// PolicyEvaluation represents the result of policy evaluation
type PolicyEvaluation struct {
	PackageName    string        `json:"package_name"`
	ProjectID      uint          `json:"project_id"`
	MatchedPolicies []PolicyMatch `json:"matched_policies"`
	FinalAction    PolicyActions `json:"final_action"`
	Decision       string        `json:"decision"`       // allow, block, require_approval
	Explanation    string        `json:"explanation"`    // Human-readable explanation
}

// NewPolicyService creates a new policy service
func NewPolicyService() *PolicyService {
	return &PolicyService{
		db: database.GetDB(),
	}
}

// CreatePolicy creates a new update policy
func (ps *PolicyService) CreatePolicy(ctx context.Context, policy *UpdatePolicy) error {
	logger.Info("Creating update policy: %s", policy.Name)

	// Validate policy
	if err := ps.validatePolicy(policy); err != nil {
		return fmt.Errorf("policy validation failed: %w", err)
	}

	// Serialize conditions and actions
	conditionsJSON, err := json.Marshal(policy.Conditions)
	if err != nil {
		return fmt.Errorf("failed to serialize conditions: %w", err)
	}

	actionsJSON, err := json.Marshal(policy.Actions)
	if err != nil {
		return fmt.Errorf("failed to serialize actions: %w", err)
	}

	// Create database record
	dbPolicy := UpdatePolicy{
		Name:        policy.Name,
		Description: policy.Description,
		ProjectID:   policy.ProjectID,
		Priority:    policy.Priority,
		Enabled:     policy.Enabled,
		Author:      policy.Author,
		Version:     policy.Version,
		Tags:        policy.Tags,
	}

	if err := ps.db.Create(&dbPolicy).Error; err != nil {
		return fmt.Errorf("failed to create policy: %w", err)
	}

	// Update with serialized data
	if err := ps.db.Model(&dbPolicy).Updates(map[string]interface{}{
		"conditions": string(conditionsJSON),
		"actions":    string(actionsJSON),
	}).Error; err != nil {
		return fmt.Errorf("failed to update policy data: %w", err)
	}

	policy.ID = dbPolicy.ID
	policy.CreatedAt = dbPolicy.CreatedAt
	policy.UpdatedAt = dbPolicy.UpdatedAt

	logger.Info("Created policy %s with ID %d", policy.Name, policy.ID)
	return nil
}

// GetPolicy retrieves a policy by ID
func (ps *PolicyService) GetPolicy(ctx context.Context, id uint) (*UpdatePolicy, error) {
	var policy UpdatePolicy
	if err := ps.db.First(&policy, id).Error; err != nil {
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}

	// Deserialize conditions and actions
	if err := ps.deserializePolicyData(&policy); err != nil {
		return nil, fmt.Errorf("failed to deserialize policy data: %w", err)
	}

	return &policy, nil
}

// ListPolicies lists all policies, optionally filtered by project
func (ps *PolicyService) ListPolicies(ctx context.Context, projectID *uint) ([]UpdatePolicy, error) {
	query := ps.db.Model(&UpdatePolicy{})

	if projectID != nil {
		query = query.Where("project_id = ? OR project_id IS NULL", *projectID)
	}

	var policies []UpdatePolicy
	if err := query.Order("priority DESC, created_at ASC").Find(&policies).Error; err != nil {
		return nil, fmt.Errorf("failed to list policies: %w", err)
	}

	// Deserialize all policies
	for i := range policies {
		if err := ps.deserializePolicyData(&policies[i]); err != nil {
			logger.Warn("Failed to deserialize policy %d: %v", policies[i].ID, err)
		}
	}

	return policies, nil
}

// UpdatePolicy updates an existing policy
func (ps *PolicyService) UpdatePolicy(ctx context.Context, policy *UpdatePolicy) error {
	logger.Info("Updating policy %d: %s", policy.ID, policy.Name)

	// Validate policy
	if err := ps.validatePolicy(policy); err != nil {
		return fmt.Errorf("policy validation failed: %w", err)
	}

	// Serialize conditions and actions
	conditionsJSON, err := json.Marshal(policy.Conditions)
	if err != nil {
		return fmt.Errorf("failed to serialize conditions: %w", err)
	}

	actionsJSON, err := json.Marshal(policy.Actions)
	if err != nil {
		return fmt.Errorf("failed to serialize actions: %w", err)
	}

	// Update database record
	updates := map[string]interface{}{
		"name":        policy.Name,
		"description": policy.Description,
		"project_id":  policy.ProjectID,
		"priority":    policy.Priority,
		"enabled":     policy.Enabled,
		"author":      policy.Author,
		"version":     policy.Version,
		"tags":        policy.Tags,
		"conditions":  string(conditionsJSON),
		"actions":     string(actionsJSON),
	}

	if err := ps.db.Model(&UpdatePolicy{}).Where("id = ?", policy.ID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update policy: %w", err)
	}

	logger.Info("Updated policy %s", policy.Name)
	return nil
}

// DeletePolicy deletes a policy
func (ps *PolicyService) DeletePolicy(ctx context.Context, id uint) error {
	logger.Info("Deleting policy %d", id)

	if err := ps.db.Delete(&UpdatePolicy{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	logger.Info("Deleted policy %d", id)
	return nil
}

// EvaluateUpdate evaluates update policies for a specific package update
func (ps *PolicyService) EvaluateUpdate(ctx context.Context, dependency models.Dependency, update models.Update) (*PolicyEvaluation, error) {
	logger.Debug("Evaluating policies for %s update", dependency.Name)

	// Get applicable policies
	policies, err := ps.ListPolicies(ctx, &dependency.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get policies: %w", err)
	}

	evaluation := &PolicyEvaluation{
		PackageName:     dependency.Name,
		ProjectID:       dependency.ProjectID,
		MatchedPolicies: []PolicyMatch{},
		Decision:        "allow", // Default decision
	}

	// Evaluate each policy
	for _, policy := range policies {
		if !policy.Enabled {
			continue
		}

		match, score := ps.evaluatePolicyConditions(policy, dependency, update)
		if match {
			evaluation.MatchedPolicies = append(evaluation.MatchedPolicies, PolicyMatch{
				Policy:     policy,
				MatchScore: score,
				Reason:     fmt.Sprintf("Matched %d conditions", ps.countMatchedConditions(policy, dependency, update)),
			})
		}
	}

	// Apply policies in priority order
	evaluation.FinalAction = ps.combinePolicyActions(evaluation.MatchedPolicies)
	evaluation.Decision = ps.makeFinalDecision(evaluation.FinalAction)
	evaluation.Explanation = ps.generateExplanation(evaluation)

	return evaluation, nil
}

// Private helper methods

func (ps *PolicyService) validatePolicy(policy *UpdatePolicy) error {
	if policy.Name == "" {
		return fmt.Errorf("policy name is required")
	}

	if policy.Priority < 0 || policy.Priority > 100 {
		return fmt.Errorf("policy priority must be between 0 and 100")
	}

	// Validate conditions
	if err := ps.validateConditions(&policy.Conditions); err != nil {
		return fmt.Errorf("invalid conditions: %w", err)
	}

	// Validate actions
	if err := ps.validateActions(&policy.Actions); err != nil {
		return fmt.Errorf("invalid actions: %w", err)
	}

	return nil
}

func (ps *PolicyService) validateConditions(conditions *PolicyConditions) error {
	// Validate regex patterns
	for _, pattern := range conditions.PackagePatterns {
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("invalid package pattern '%s': %w", pattern, err)
		}
	}

	if conditions.CurrentVersionPattern != "" {
		if _, err := regexp.Compile(conditions.CurrentVersionPattern); err != nil {
			return fmt.Errorf("invalid current version pattern: %w", err)
		}
	}

	if conditions.TargetVersionPattern != "" {
		if _, err := regexp.Compile(conditions.TargetVersionPattern); err != nil {
			return fmt.Errorf("invalid target version pattern: %w", err)
		}
	}

	// Validate risk score ranges
	if conditions.RiskScoreMin != nil && (*conditions.RiskScoreMin < 0 || *conditions.RiskScoreMin > 10) {
		return fmt.Errorf("risk score min must be between 0 and 10")
	}

	if conditions.RiskScoreMax != nil && (*conditions.RiskScoreMax < 0 || *conditions.RiskScoreMax > 10) {
		return fmt.Errorf("risk score max must be between 0 and 10")
	}

	return nil
}

func (ps *PolicyService) validateActions(actions *PolicyActions) error {
	// Validate schedule if provided
	if actions.Schedule != "" {
		// This would validate cron expression - simplified for now
		if !strings.Contains(actions.Schedule, " ") {
			return fmt.Errorf("invalid schedule format")
		}
	}

	// Validate batch size
	if actions.BatchSize < 0 {
		return fmt.Errorf("batch size must be non-negative")
	}

	// Validate delay days
	if actions.DelayDays < 0 {
		return fmt.Errorf("delay days must be non-negative")
	}

	return nil
}

func (ps *PolicyService) deserializePolicyData(policy *UpdatePolicy) error {
	// Get the raw data from database
	var rawPolicy struct {
		Conditions string `gorm:"column:conditions"`
		Actions    string `gorm:"column:actions"`
	}

	if err := ps.db.Model(&UpdatePolicy{}).Select("conditions, actions").Where("id = ?", policy.ID).First(&rawPolicy).Error; err != nil {
		return fmt.Errorf("failed to get raw policy data: %w", err)
	}

	// Deserialize conditions
	if rawPolicy.Conditions != "" {
		if err := json.Unmarshal([]byte(rawPolicy.Conditions), &policy.Conditions); err != nil {
			return fmt.Errorf("failed to deserialize conditions: %w", err)
		}
	}

	// Deserialize actions
	if rawPolicy.Actions != "" {
		if err := json.Unmarshal([]byte(rawPolicy.Actions), &policy.Actions); err != nil {
			return fmt.Errorf("failed to deserialize actions: %w", err)
		}
	}

	return nil
}

func (ps *PolicyService) evaluatePolicyConditions(policy UpdatePolicy, dependency models.Dependency, update models.Update) (bool, float64) {
	conditions := policy.Conditions
	matchCount := 0
	totalConditions := 0
	
	// Package name matching
	if len(conditions.PackageNames) > 0 {
		totalConditions++
		for _, name := range conditions.PackageNames {
			if name == dependency.Name {
				matchCount++
				break
			}
		}
	}

	// Package pattern matching
	if len(conditions.PackagePatterns) > 0 {
		totalConditions++
		for _, pattern := range conditions.PackagePatterns {
			if matched, _ := regexp.MatchString(pattern, dependency.Name); matched {
				matchCount++
				break
			}
		}
	}

	// Update type matching
	if len(conditions.UpdateTypes) > 0 {
		totalConditions++
		for _, updateType := range conditions.UpdateTypes {
			if updateType == update.UpdateType {
				matchCount++
				break
			}
		}
	}

	// Security risk matching
	if conditions.SecurityRisk != nil {
		totalConditions++
		hasSecurityRisk := ps.hasSecurityRisk(dependency)
		if *conditions.SecurityRisk == hasSecurityRisk {
			matchCount++
		}
	}

	// Risk score range matching
	if conditions.RiskScoreMin != nil || conditions.RiskScoreMax != nil {
		totalConditions++
		riskScore := ps.calculateRiskScore(dependency, update)
		
		withinRange := true
		if conditions.RiskScoreMin != nil && riskScore < *conditions.RiskScoreMin {
			withinRange = false
		}
		if conditions.RiskScoreMax != nil && riskScore > *conditions.RiskScoreMax {
			withinRange = false
		}
		
		if withinRange {
			matchCount++
		}
	}

	// If no conditions specified, don't match
	if totalConditions == 0 {
		return false, 0.0
	}

	// Calculate match score
	score := float64(matchCount) / float64(totalConditions)
	
	// Policy matches if score is above threshold (e.g., 0.5)
	return score >= 0.5, score
}

func (ps *PolicyService) countMatchedConditions(policy UpdatePolicy, dependency models.Dependency, update models.Update) int {
	// Simplified count - would implement full condition counting
	return 3 // Placeholder
}

func (ps *PolicyService) combinePolicyActions(matches []PolicyMatch) PolicyActions {
	if len(matches) == 0 {
		return PolicyActions{}
	}

	// Sort by priority (highest first)
	// For simplicity, we'll just take the first match's actions
	return matches[0].Policy.Actions
}

func (ps *PolicyService) makeFinalDecision(actions PolicyActions) string {
	if actions.BlockUpdate != nil && *actions.BlockUpdate {
		return "block"
	}
	
	if actions.RequireApproval != nil && *actions.RequireApproval {
		return "require_approval"
	}
	
	if actions.AutoUpdate != nil && *actions.AutoUpdate {
		return "auto_update"
	}
	
	return "allow"
}

func (ps *PolicyService) generateExplanation(evaluation *PolicyEvaluation) string {
	if len(evaluation.MatchedPolicies) == 0 {
		return "No policies matched - using default behavior"
	}

	policy := evaluation.MatchedPolicies[0].Policy
	
	switch evaluation.Decision {
	case "block":
		return fmt.Sprintf("Update blocked by policy '%s'", policy.Name)
	case "require_approval":
		return fmt.Sprintf("Manual approval required by policy '%s'", policy.Name)
	case "auto_update":
		return fmt.Sprintf("Auto-update allowed by policy '%s'", policy.Name)
	default:
		return fmt.Sprintf("Update allowed by policy '%s'", policy.Name)
	}
}

func (ps *PolicyService) hasSecurityRisk(dependency models.Dependency) bool {
	var count int64
	ps.db.Model(&models.SecurityCheck{}).Where("package_name = ? AND version = ? AND status = ?", 
		dependency.Name, dependency.CurrentVersion, "detected").Count(&count)
	return count > 0
}

func (ps *PolicyService) calculateRiskScore(dependency models.Dependency, update models.Update) float64 {
	// Simplified risk calculation
	score := 0.0
	
	// Base score from update type
	switch update.UpdateType {
	case "security":
		score += 2.0
	case "major":
		score += 4.0
	case "minor":
		score += 2.0
	case "patch":
		score += 1.0
	}
	
	// Add security risk
	if ps.hasSecurityRisk(dependency) {
		score += 3.0
	}
	
	// Cap at 10.0
	if score > 10.0 {
		score = 10.0
	}
	
	return score
}
