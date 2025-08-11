package github

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// PolicyManager handles custom patch rules and organization policies
type PolicyManager struct {
	client   *Client
	config   *PolicyConfig
	policies []*Policy
	rules    []*PatchRule
}

// NewPolicyManager creates a new policy manager
func NewPolicyManager(client *Client, config *PolicyConfig) *PolicyManager {
	pm := &PolicyManager{
		client:   client,
		config:   config,
		policies: []*Policy{},
		rules:    []*PatchRule{},
	}

	// Load default policies
	pm.loadDefaultPolicies()
	
	return pm
}

// PolicyConfig defines policy management configuration
type PolicyConfig struct {
	EnablePolicyEnforcement bool          `json:"enable_policy_enforcement"`
	DefaultRiskTolerance    string        `json:"default_risk_tolerance"` // "low", "medium", "high"
	RequireApprovalFor      []string      `json:"require_approval_for"`   // risk levels requiring approval
	BlockedDependencies     []string      `json:"blocked_dependencies"`   // dependencies that are blocked
	AllowedLicenses         []string      `json:"allowed_licenses"`       // allowed license types
	SecurityScanRequired    bool          `json:"security_scan_required"`
	ComplianceMode          string        `json:"compliance_mode"`        // "strict", "moderate", "permissive"
	PolicyViolationAction   string        `json:"policy_violation_action"` // "block", "warn", "log"
}

// Policy represents an organization policy
type Policy struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`        // "security", "compliance", "quality", "performance"
	Scope       string                 `json:"scope"`       // "organization", "repository", "dependency"
	Priority    int                    `json:"priority"`    // 1-10, higher is more important
	Enabled     bool                   `json:"enabled"`
	Rules       []*PolicyRule          `json:"rules"`
	Actions     []*PolicyAction        `json:"actions"`
	Exceptions  []*PolicyException     `json:"exceptions"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CreatedBy   string                 `json:"created_by"`
}

// PolicyRule defines a rule within a policy
type PolicyRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Condition   *RuleCondition         `json:"condition"`
	Severity    string                 `json:"severity"`    // "critical", "high", "medium", "low"
	Message     string                 `json:"message"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// PolicyAction defines an action to take when a policy is violated
type PolicyAction struct {
	Type       string                 `json:"type"`       // "block", "warn", "require_approval", "notify"
	Target     string                 `json:"target"`     // who/what to target
	Parameters map[string]interface{} `json:"parameters"`
}

// PolicyException defines an exception to a policy
type PolicyException struct {
	ID          string                 `json:"id"`
	Repository  string                 `json:"repository"`
	Dependency  string                 `json:"dependency"`
	Reason      string                 `json:"reason"`
	ExpiresAt   *time.Time             `json:"expires_at"`
	CreatedBy   string                 `json:"created_by"`
	ApprovedBy  string                 `json:"approved_by"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// PatchRule represents a custom rule for patch generation
type PatchRule struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	DependencyPattern string                `json:"dependency_pattern"` // regex pattern for dependency names
	VersionPattern   string                 `json:"version_pattern"`    // regex pattern for versions
	FilePatterns     []string               `json:"file_patterns"`      // file patterns to apply rule to
	Conditions       []*RuleCondition       `json:"conditions"`
	Transformations  []*PatchTransformation `json:"transformations"`
	Priority         int                    `json:"priority"`
	Enabled          bool                   `json:"enabled"`
	Metadata         map[string]interface{} `json:"metadata"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// PatchTransformation defines how to transform code during patching
type PatchTransformation struct {
	Type        string                 `json:"type"`        // "replace", "insert", "delete", "modify"
	Pattern     string                 `json:"pattern"`     // regex pattern to match
	Replacement string                 `json:"replacement"` // replacement text
	Scope       string                 `json:"scope"`       // "line", "function", "file"
	Parameters  map[string]interface{} `json:"parameters"`
}

// PolicyViolation represents a policy violation
type PolicyViolation struct {
	ID           string                 `json:"id"`
	PolicyID     string                 `json:"policy_id"`
	RuleID       string                 `json:"rule_id"`
	Repository   string                 `json:"repository"`
	Dependency   string                 `json:"dependency"`
	Version      string                 `json:"version"`
	Severity     string                 `json:"severity"`
	Message      string                 `json:"message"`
	Details      string                 `json:"details"`
	Status       string                 `json:"status"`       // "open", "resolved", "ignored"
	Resolution   string                 `json:"resolution"`
	Metadata     map[string]interface{} `json:"metadata"`
	DetectedAt   time.Time              `json:"detected_at"`
	ResolvedAt   *time.Time             `json:"resolved_at"`
}

// PolicyEvaluationResult represents the result of policy evaluation
type PolicyEvaluationResult struct {
	Allowed         bool                `json:"allowed"`
	Violations      []*PolicyViolation  `json:"violations"`
	Warnings        []string            `json:"warnings"`
	RequiredActions []*PolicyAction     `json:"required_actions"`
	ApprovalNeeded  bool                `json:"approval_needed"`
	BlockingIssues  []string            `json:"blocking_issues"`
	Recommendations []string            `json:"recommendations"`
	EvaluatedAt     time.Time           `json:"evaluated_at"`
}

// EvaluatePolicies evaluates policies for a dependency update
func (pm *PolicyManager) EvaluatePolicies(ctx context.Context, dependency *DependencyInfo, patches []*Patch) (*PolicyEvaluationResult, error) {
	result := &PolicyEvaluationResult{
		Allowed:         true,
		Violations:      []*PolicyViolation{},
		Warnings:        []string{},
		RequiredActions: []*PolicyAction{},
		ApprovalNeeded:  false,
		BlockingIssues:  []string{},
		Recommendations: []string{},
		EvaluatedAt:     time.Now(),
	}

	// Evaluate each policy
	for _, policy := range pm.policies {
		if !policy.Enabled {
			continue
		}

		policyResult, err := pm.evaluatePolicy(ctx, policy, dependency, patches)
		if err != nil {
			return result, fmt.Errorf("failed to evaluate policy %s: %w", policy.ID, err)
		}

		// Merge results
		result.Violations = append(result.Violations, policyResult.Violations...)
		result.Warnings = append(result.Warnings, policyResult.Warnings...)
		result.RequiredActions = append(result.RequiredActions, policyResult.RequiredActions...)

		if !policyResult.Allowed {
			result.Allowed = false
		}

		if policyResult.ApprovalNeeded {
			result.ApprovalNeeded = true
		}

		result.BlockingIssues = append(result.BlockingIssues, policyResult.BlockingIssues...)
		result.Recommendations = append(result.Recommendations, policyResult.Recommendations...)
	}

	// Apply final decision logic
	if len(result.Violations) > 0 {
		criticalViolations := pm.filterViolationsBySeverity(result.Violations, "critical")
		if len(criticalViolations) > 0 && pm.config.ComplianceMode == "strict" {
			result.Allowed = false
			result.BlockingIssues = append(result.BlockingIssues, "Critical policy violations detected")
		}
	}

	return result, nil
}

// evaluatePolicy evaluates a single policy
func (pm *PolicyManager) evaluatePolicy(ctx context.Context, policy *Policy, dependency *DependencyInfo, patches []*Patch) (*PolicyEvaluationResult, error) {
	result := &PolicyEvaluationResult{
		Allowed:         true,
		Violations:      []*PolicyViolation{},
		Warnings:        []string{},
		RequiredActions: []*PolicyAction{},
		ApprovalNeeded:  false,
		BlockingIssues:  []string{},
		Recommendations: []string{},
		EvaluatedAt:     time.Now(),
	}

	// Check if policy applies to this dependency
	if !pm.policyApplies(policy, dependency) {
		return result, nil
	}

	// Evaluate each rule in the policy
	for _, rule := range policy.Rules {
		violation := pm.evaluateRule(rule, dependency, patches)
		if violation != nil {
			violation.PolicyID = policy.ID
			result.Violations = append(result.Violations, violation)

			// Determine actions based on severity
			switch violation.Severity {
			case "critical":
				result.Allowed = false
				result.BlockingIssues = append(result.BlockingIssues, violation.Message)
			case "high":
				if pm.config.ComplianceMode == "strict" {
					result.Allowed = false
					result.BlockingIssues = append(result.BlockingIssues, violation.Message)
				} else {
					result.ApprovalNeeded = true
				}
			case "medium":
				if pm.config.ComplianceMode != "permissive" {
					result.ApprovalNeeded = true
				}
				result.Warnings = append(result.Warnings, violation.Message)
			case "low":
				result.Warnings = append(result.Warnings, violation.Message)
			}
		}
	}

	// Apply policy actions
	if len(result.Violations) > 0 {
		result.RequiredActions = append(result.RequiredActions, policy.Actions...)
	}

	return result, nil
}

// evaluateRule evaluates a single policy rule
func (pm *PolicyManager) evaluateRule(rule *PolicyRule, dependency *DependencyInfo, patches []*Patch) *PolicyViolation {
	// Check if rule condition is met
	conditionMet, err := pm.evaluateRuleCondition(rule.Condition, dependency, patches)
	if err != nil || !conditionMet {
		return nil
	}

	// Create violation
	violation := &PolicyViolation{
		ID:         fmt.Sprintf("violation_%d", time.Now().UnixNano()),
		RuleID:     rule.ID,
		Repository: dependency.Repository,
		Dependency: dependency.Name,
		Version:    dependency.LatestVersion,
		Severity:   rule.Severity,
		Message:    rule.Message,
		Status:     "open",
		DetectedAt: time.Now(),
		Metadata: map[string]interface{}{
			"rule_name": rule.Name,
		},
	}

	return violation
}

// evaluateRuleCondition evaluates a rule condition
func (pm *PolicyManager) evaluateRuleCondition(condition *RuleCondition, dependency *DependencyInfo, patches []*Patch) (bool, error) {
	switch condition.Type {
	case "dependency_name":
		pattern, ok := condition.Value.(string)
		if !ok {
			return false, fmt.Errorf("invalid dependency_name pattern: %v", condition.Value)
		}
		matched, err := regexp.MatchString(pattern, dependency.Name)
		return matched, err

	case "version_change":
		changeType, ok := condition.Value.(string)
		if !ok {
			return false, fmt.Errorf("invalid version_change type: %v", condition.Value)
		}
		return pm.matchesVersionChange(dependency, changeType), nil

	case "license_type":
		allowedLicenses, ok := condition.Value.([]string)
		if !ok {
			return false, fmt.Errorf("invalid license_type list: %v", condition.Value)
		}
		return pm.checkLicenseCompliance(dependency, allowedLicenses), nil

	case "security_vulnerability":
		hasVulnerability, ok := condition.Value.(bool)
		if !ok {
			return false, fmt.Errorf("invalid security_vulnerability flag: %v", condition.Value)
		}
		return pm.hasSecurityVulnerability(dependency) == hasVulnerability, nil

	case "breaking_changes":
		hasBreaking, ok := condition.Value.(bool)
		if !ok {
			return false, fmt.Errorf("invalid breaking_changes flag: %v", condition.Value)
		}
		return pm.hasBreakingChanges(patches) == hasBreaking, nil

	default:
		return false, fmt.Errorf("unsupported condition type: %s", condition.Type)
	}
}

// ApplyPatchRules applies custom patch rules to modify patches
func (pm *PolicyManager) ApplyPatchRules(ctx context.Context, patches []*Patch, dependency *DependencyInfo) ([]*Patch, error) {
	modifiedPatches := make([]*Patch, len(patches))
	copy(modifiedPatches, patches)

	// Apply each rule in priority order
	applicableRules := pm.getApplicableRules(dependency)
	
	for _, rule := range applicableRules {
		var err error
		modifiedPatches, err = pm.applyPatchRule(rule, modifiedPatches, dependency)
		if err != nil {
			return patches, fmt.Errorf("failed to apply patch rule %s: %w", rule.ID, err)
		}
	}

	return modifiedPatches, nil
}

// applyPatchRule applies a single patch rule
func (pm *PolicyManager) applyPatchRule(rule *PatchRule, patches []*Patch, dependency *DependencyInfo) ([]*Patch, error) {
	for i, patch := range patches {
		// Check if rule applies to this patch
		if !pm.ruleAppliesTo(rule, patch, dependency) {
			continue
		}

		// Apply transformations
		for _, transform := range rule.Transformations {
			modifiedPatch, err := pm.applyTransformation(transform, patch)
			if err != nil {
				return patches, fmt.Errorf("failed to apply transformation: %w", err)
			}
			patches[i] = modifiedPatch
		}
	}

	return patches, nil
}

// applyTransformation applies a single transformation to a patch
func (pm *PolicyManager) applyTransformation(transform *PatchTransformation, patch *Patch) (*Patch, error) {
	modifiedPatch := *patch // Copy the patch

	switch transform.Type {
	case "replace":
		regex, err := regexp.Compile(transform.Pattern)
		if err != nil {
			return patch, fmt.Errorf("invalid regex pattern: %w", err)
		}
		modifiedPatch.Content = regex.ReplaceAllString(patch.Content, transform.Replacement)

	case "insert":
		// Insert content at specified location
		modifiedPatch.Content = patch.Content + "\n" + transform.Replacement

	case "delete":
		// Remove content matching pattern
		regex, err := regexp.Compile(transform.Pattern)
		if err != nil {
			return patch, fmt.Errorf("invalid regex pattern: %w", err)
		}
		modifiedPatch.Content = regex.ReplaceAllString(patch.Content, "")

	case "modify":
		// Custom modification logic
		modifiedPatch.Content = pm.applyCustomModification(patch.Content, transform)

	default:
		return patch, fmt.Errorf("unsupported transformation type: %s", transform.Type)
	}

	return &modifiedPatch, nil
}

// CreatePolicy creates a new organization policy
func (pm *PolicyManager) CreatePolicy(ctx context.Context, policy *Policy) error {
	policy.ID = fmt.Sprintf("policy_%d", time.Now().Unix())
	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()

	// Validate policy
	if err := pm.validatePolicy(policy); err != nil {
		return fmt.Errorf("invalid policy: %w", err)
	}

	pm.policies = append(pm.policies, policy)
	return nil
}

// CreatePatchRule creates a new custom patch rule
func (pm *PolicyManager) CreatePatchRule(ctx context.Context, rule *PatchRule) error {
	rule.ID = fmt.Sprintf("rule_%d", time.Now().Unix())
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	// Validate rule
	if err := pm.validatePatchRule(rule); err != nil {
		return fmt.Errorf("invalid patch rule: %w", err)
	}

	pm.rules = append(pm.rules, rule)
	return nil
}

// Helper methods

func (pm *PolicyManager) loadDefaultPolicies() {
	// Security policy
	securityPolicy := &Policy{
		ID:          "security_default",
		Name:        "Security Policy",
		Description: "Default security policy for dependency updates",
		Type:        "security",
		Scope:       "organization",
		Priority:    10,
		Enabled:     true,
		Rules: []*PolicyRule{
			{
				ID:   "security_vuln_check",
				Name: "Security Vulnerability Check",
				Condition: &RuleCondition{
					Type:  "security_vulnerability",
					Value: true,
				},
				Severity: "high",
				Message:  "Dependency has known security vulnerabilities",
			},
		},
		Actions: []*PolicyAction{
			{
				Type:   "require_approval",
				Target: "security-team",
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: "system",
	}

	// License compliance policy
	licensePolicy := &Policy{
		ID:          "license_compliance",
		Name:        "License Compliance Policy",
		Description: "Ensures only approved licenses are used",
		Type:        "compliance",
		Scope:       "organization",
		Priority:    8,
		Enabled:     true,
		Rules: []*PolicyRule{
			{
				ID:   "license_check",
				Name: "License Compliance Check",
				Condition: &RuleCondition{
					Type:  "license_type",
					Value: pm.config.AllowedLicenses,
				},
				Severity: "medium",
				Message:  "Dependency uses non-approved license",
			},
		},
		Actions: []*PolicyAction{
			{
				Type:   "require_approval",
				Target: "legal-team",
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: "system",
	}

	pm.policies = append(pm.policies, securityPolicy, licensePolicy)
}

func (pm *PolicyManager) policyApplies(policy *Policy, dependency *DependencyInfo) bool {
	// Check scope
	switch policy.Scope {
	case "organization":
		return true
	case "repository":
		// Check if policy applies to this repository
		return true // Simplified
	case "dependency":
		// Check if policy applies to this specific dependency
		return strings.Contains(strings.ToLower(dependency.Name), strings.ToLower(policy.Name))
	}
	return false
}

func (pm *PolicyManager) filterViolationsBySeverity(violations []*PolicyViolation, severity string) []*PolicyViolation {
	var filtered []*PolicyViolation
	for _, violation := range violations {
		if violation.Severity == severity {
			filtered = append(filtered, violation)
		}
	}
	return filtered
}

func (pm *PolicyManager) matchesVersionChange(dependency *DependencyInfo, changeType string) bool {
	// Simplified version change detection
	switch changeType {
	case "major":
		return strings.Contains(dependency.LatestVersion, ".")
	case "minor":
		return true // Simplified
	case "patch":
		return true // Simplified
	}
	return false
}

func (pm *PolicyManager) checkLicenseCompliance(dependency *DependencyInfo, allowedLicenses []string) bool {
	// Simplified license check
	for _, license := range allowedLicenses {
		if license == "MIT" || license == "Apache-2.0" {
			return true
		}
	}
	return false
}

func (pm *PolicyManager) hasSecurityVulnerability(dependency *DependencyInfo) bool {
	// Simplified vulnerability check
	return strings.Contains(strings.ToLower(dependency.Name), "vulnerable")
}

func (pm *PolicyManager) hasBreakingChanges(patches []*Patch) bool {
	for _, patch := range patches {
		if patch.BreakingChange {
			return true
		}
	}
	return false
}

func (pm *PolicyManager) getApplicableRules(dependency *DependencyInfo) []*PatchRule {
	var applicable []*PatchRule
	
	for _, rule := range pm.rules {
		if !rule.Enabled {
			continue
		}
		
		// Check if rule applies to this dependency
		if rule.DependencyPattern != "" {
			matched, err := regexp.MatchString(rule.DependencyPattern, dependency.Name)
			if err != nil || !matched {
				continue
			}
		}
		
		applicable = append(applicable, rule)
	}
	
	return applicable
}

func (pm *PolicyManager) ruleAppliesTo(rule *PatchRule, patch *Patch, dependency *DependencyInfo) bool {
	// Check file patterns
	for _, pattern := range rule.FilePatterns {
		matched, err := regexp.MatchString(pattern, patch.File)
		if err == nil && matched {
			return true
		}
	}
	
	// If no file patterns specified, rule applies to all files
	return len(rule.FilePatterns) == 0
}

func (pm *PolicyManager) applyCustomModification(content string, transform *PatchTransformation) string {
	// Custom modification logic based on parameters
	return content // Simplified
}

func (pm *PolicyManager) validatePolicy(policy *Policy) error {
	if policy.Name == "" {
		return fmt.Errorf("policy name is required")
	}
	if policy.Type == "" {
		return fmt.Errorf("policy type is required")
	}
	return nil
}

func (pm *PolicyManager) validatePatchRule(rule *PatchRule) error {
	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}
	if rule.DependencyPattern != "" {
		_, err := regexp.Compile(rule.DependencyPattern)
		if err != nil {
			return fmt.Errorf("invalid dependency pattern: %w", err)
		}
	}
	return nil
}

// GetPolicies returns all policies
func (pm *PolicyManager) GetPolicies() []*Policy {
	return pm.policies
}

// GetPatchRules returns all patch rules
func (pm *PolicyManager) GetPatchRules() []*PatchRule {
	return pm.rules
}

// GetViolations returns all policy violations
func (pm *PolicyManager) GetViolations(ctx context.Context, repository string) ([]*PolicyViolation, error) {
	// In real implementation, would query database for violations
	return []*PolicyViolation{}, nil
}
