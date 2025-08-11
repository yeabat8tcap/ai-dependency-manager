package github

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/ai"
)

// ReviewManager handles automated review request management
type ReviewManager struct {
	client    *Client
	aiManager AIManager
}

// NewReviewManager creates a new review manager
func NewReviewManager(client *Client, aiManager *ai.Manager) *ReviewManager {
	return &ReviewManager{
		client:    client,
		aiManager: aiManager,
	}
}

// RequestReviews automatically requests appropriate reviewers for a PR
func (rm *ReviewManager) RequestReviews(ctx context.Context, pr *PullRequest, options *PRCreationOptions) ([]*ReviewRequest, error) {
	var reviewRequests []*ReviewRequest

	// Analyze PR to determine appropriate reviewers
	reviewAnalysis, err := rm.analyzePRForReviewers(ctx, pr)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze PR for reviewers: %w", err)
	}

	// Get suggested reviewers
	suggestedReviewers := rm.getSuggestedReviewers(reviewAnalysis, options)

	// Request individual reviewers
	for _, reviewer := range suggestedReviewers.Users {
		request, err := rm.requestUserReview(ctx, pr, reviewer)
		if err != nil {
			continue // Log error but continue with other reviewers
		}
		reviewRequests = append(reviewRequests, request)
	}

	// Request team reviews
	for _, team := range suggestedReviewers.Teams {
		request, err := rm.requestTeamReview(ctx, pr, team)
		if err != nil {
			continue // Log error but continue with other teams
		}
		reviewRequests = append(reviewRequests, request)
	}

	return reviewRequests, nil
}

// ReviewAnalysis represents analysis of a PR for reviewer selection
type ReviewAnalysis struct {
	Complexity      string                 `json:"complexity"`
	RiskLevel       string                 `json:"risk_level"`
	AffectedAreas   []string               `json:"affected_areas"`
	RequiredSkills  []string               `json:"required_skills"`
	BreakingChanges bool                   `json:"breaking_changes"`
	SecurityImpact  bool                   `json:"security_impact"`
	Dependencies    []*DependencyChange    `json:"dependencies"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// SuggestedReviewers represents suggested reviewers for a PR
type SuggestedReviewers struct {
	Users []*ReviewerSuggestion `json:"users"`
	Teams []*TeamSuggestion     `json:"teams"`
}

// ReviewerSuggestion represents a suggested individual reviewer
type ReviewerSuggestion struct {
	Username   string                 `json:"username"`
	Reason     string                 `json:"reason"`
	Expertise  []string               `json:"expertise"`
	Priority   string                 `json:"priority"`
	Confidence float64                `json:"confidence"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// TeamSuggestion represents a suggested team reviewer
type TeamSuggestion struct {
	TeamName   string                 `json:"team_name"`
	Reason     string                 `json:"reason"`
	Expertise  []string               `json:"expertise"`
	Priority   string                 `json:"priority"`
	Confidence float64                `json:"confidence"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// analyzePRForReviewers analyzes a PR to determine appropriate reviewers
func (rm *ReviewManager) analyzePRForReviewers(ctx context.Context, pr *PullRequest) (*ReviewAnalysis, error) {
	analysis := &ReviewAnalysis{
		Complexity:      "medium",
		RiskLevel:       "medium",
		AffectedAreas:   []string{},
		RequiredSkills:  []string{},
		BreakingChanges: false,
		SecurityImpact:  false,
		Dependencies:    []*DependencyChange{},
		Metadata:        make(map[string]interface{}),
	}

	// Analyze PR title and description for keywords
	analysis.AffectedAreas = rm.extractAffectedAreas(pr.Title, pr.Body)
	analysis.RequiredSkills = rm.extractRequiredSkills(pr.Title, pr.Body)

	// Check for breaking changes
	if strings.Contains(strings.ToLower(pr.Body), "breaking") ||
		strings.Contains(strings.ToLower(pr.Body), "major") {
		analysis.BreakingChanges = true
		analysis.RiskLevel = "high"
	}

	// Check for security impact
	if rm.hasSecurityKeywords(pr.Title, pr.Body) {
		analysis.SecurityImpact = true
		analysis.RequiredSkills = append(analysis.RequiredSkills, "security")
	}

	// Use AI to enhance analysis if available
	if rm.aiManager != nil {
		aiAnalysis, err := rm.getAIReviewAnalysis(ctx, pr)
		if err == nil {
			analysis = rm.mergeAIAnalysis(analysis, aiAnalysis)
		}
	}

	return analysis, nil
}

// getSuggestedReviewers gets suggested reviewers based on analysis
func (rm *ReviewManager) getSuggestedReviewers(analysis *ReviewAnalysis, options *PRCreationOptions) *SuggestedReviewers {
	suggestions := &SuggestedReviewers{
		Users: []*ReviewerSuggestion{},
		Teams: []*TeamSuggestion{},
	}

	// Add explicitly requested reviewers
	for _, reviewer := range options.Reviewers {
		suggestions.Users = append(suggestions.Users, &ReviewerSuggestion{
			Username:   reviewer,
			Reason:     "Explicitly requested",
			Priority:   "high",
			Confidence: 1.0,
		})
	}

	// Add explicitly requested teams
	for _, team := range options.TeamReviewers {
		suggestions.Teams = append(suggestions.Teams, &TeamSuggestion{
			TeamName:   team,
			Reason:     "Explicitly requested",
			Priority:   "high",
			Confidence: 1.0,
		})
	}

	// Add suggested reviewers based on analysis
	suggestions.Users = append(suggestions.Users, rm.getSkillBasedReviewers(analysis)...)
	suggestions.Teams = append(suggestions.Teams, rm.getAreaBasedTeams(analysis)...)

	// Add reviewers based on risk level
	if analysis.RiskLevel == "high" || analysis.BreakingChanges {
		suggestions.Users = append(suggestions.Users, &ReviewerSuggestion{
			Username:   "senior-engineer",
			Reason:     "High risk changes require senior review",
			Expertise:  []string{"architecture", "senior"},
			Priority:   "high",
			Confidence: 0.9,
		})
	}

	// Add security team for security-related changes
	if analysis.SecurityImpact {
		suggestions.Teams = append(suggestions.Teams, &TeamSuggestion{
			TeamName:   "security-team",
			Reason:     "Security impact detected",
			Expertise:  []string{"security", "compliance"},
			Priority:   "critical",
			Confidence: 0.95,
		})
	}

	return suggestions
}

// getSkillBasedReviewers gets reviewers based on required skills
func (rm *ReviewManager) getSkillBasedReviewers(analysis *ReviewAnalysis) []*ReviewerSuggestion {
	var suggestions []*ReviewerSuggestion

	skillToReviewers := map[string][]string{
		"frontend":   {"frontend-lead", "ui-expert"},
		"backend":    {"backend-lead", "api-expert"},
		"database":   {"db-admin", "data-engineer"},
		"security":   {"security-engineer", "compliance-officer"},
		"devops":     {"devops-engineer", "infrastructure-lead"},
		"testing":    {"qa-lead", "test-automation-expert"},
		"javascript": {"js-expert", "frontend-lead"},
		"python":     {"python-expert", "backend-lead"},
		"go":         {"go-expert", "backend-lead"},
		"java":       {"java-expert", "enterprise-architect"},
	}

	for _, skill := range analysis.RequiredSkills {
		if reviewers, exists := skillToReviewers[strings.ToLower(skill)]; exists {
			for _, reviewer := range reviewers {
				suggestions = append(suggestions, &ReviewerSuggestion{
					Username:   reviewer,
					Reason:     fmt.Sprintf("Expert in %s", skill),
					Expertise:  []string{skill},
					Priority:   "medium",
					Confidence: 0.8,
				})
			}
		}
	}

	return suggestions
}

// getAreaBasedTeams gets team reviewers based on affected areas
func (rm *ReviewManager) getAreaBasedTeams(analysis *ReviewAnalysis) []*TeamSuggestion {
	var suggestions []*TeamSuggestion

	areaToTeams := map[string][]string{
		"api":           {"backend-team", "api-team"},
		"ui":            {"frontend-team", "design-team"},
		"database":      {"data-team", "backend-team"},
		"infrastructure": {"devops-team", "platform-team"},
		"auth":          {"security-team", "identity-team"},
		"payment":       {"payments-team", "security-team"},
		"integration":   {"integrations-team", "api-team"},
	}

	for _, area := range analysis.AffectedAreas {
		if teams, exists := areaToTeams[strings.ToLower(area)]; exists {
			for _, team := range teams {
				suggestions = append(suggestions, &TeamSuggestion{
					TeamName:   team,
					Reason:     fmt.Sprintf("Responsible for %s area", area),
					Expertise:  []string{area},
					Priority:   "medium",
					Confidence: 0.7,
				})
			}
		}
	}

	return suggestions
}

// requestUserReview requests a review from a specific user
func (rm *ReviewManager) requestUserReview(ctx context.Context, pr *PullRequest, suggestion *ReviewerSuggestion) (*ReviewRequest, error) {
	// In real implementation, would call GitHub API to request review
	request := &ReviewRequest{
		Reviewer:    suggestion.Username,
		Type:        "user",
		Status:      "requested",
		RequestedAt: time.Now(),
		Message:     rm.generateReviewMessage(suggestion),
	}

	return request, nil
}

// requestTeamReview requests a review from a specific team
func (rm *ReviewManager) requestTeamReview(ctx context.Context, pr *PullRequest, suggestion *TeamSuggestion) (*ReviewRequest, error) {
	// In real implementation, would call GitHub API to request team review
	request := &ReviewRequest{
		Reviewer:    suggestion.TeamName,
		Type:        "team",
		Status:      "requested",
		RequestedAt: time.Now(),
		Message:     rm.generateTeamReviewMessage(suggestion),
	}

	return request, nil
}

// generateReviewMessage generates a personalized review request message
func (rm *ReviewManager) generateReviewMessage(suggestion *ReviewerSuggestion) string {
	return fmt.Sprintf("Review requested: %s", suggestion.Reason)
}

// generateTeamReviewMessage generates a team review request message
func (rm *ReviewManager) generateTeamReviewMessage(suggestion *TeamSuggestion) string {
	return fmt.Sprintf("Team review requested: %s", suggestion.Reason)
}

// getAIReviewAnalysis gets AI-powered review analysis
func (rm *ReviewManager) getAIReviewAnalysis(ctx context.Context, pr *PullRequest) (*ReviewAnalysis, error) {
	prompt := fmt.Sprintf(`
Analyze this pull request for reviewer assignment:

Title: %s
Description: %s

Please provide:
1. Complexity level (low/medium/high)
2. Risk level (low/medium/high/critical)
3. Affected areas (frontend, backend, database, etc.)
4. Required skills for review
5. Whether it contains breaking changes
6. Security impact assessment

Respond in a structured format.
`, pr.Title, pr.Body)

	response, err := rm.aiManager.AnalyzeChangelog(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return rm.parseAIAnalysis(response), nil
}

// parseAIAnalysis parses AI analysis response
func (rm *ReviewManager) parseAIAnalysis(response string) *ReviewAnalysis {
	// Simplified parsing - in real implementation would parse structured response
	analysis := &ReviewAnalysis{
		Complexity:      "medium",
		RiskLevel:       "medium",
		AffectedAreas:   []string{},
		RequiredSkills:  []string{},
		BreakingChanges: false,
		SecurityImpact:  false,
		Metadata:        map[string]interface{}{"source": "ai"},
	}

	lines := strings.Split(strings.ToLower(response), "\n")
	for _, line := range lines {
		if strings.Contains(line, "high complexity") {
			analysis.Complexity = "high"
		}
		if strings.Contains(line, "high risk") {
			analysis.RiskLevel = "high"
		}
		if strings.Contains(line, "breaking") {
			analysis.BreakingChanges = true
		}
		if strings.Contains(line, "security") {
			analysis.SecurityImpact = true
		}
	}

	return analysis
}

// mergeAIAnalysis merges AI analysis with existing analysis
func (rm *ReviewManager) mergeAIAnalysis(existing, ai *ReviewAnalysis) *ReviewAnalysis {
	// Merge complexity (take higher)
	if ai.Complexity == "high" || (ai.Complexity == "medium" && existing.Complexity == "low") {
		existing.Complexity = ai.Complexity
	}

	// Merge risk level (take higher)
	riskLevels := map[string]int{"low": 1, "medium": 2, "high": 3, "critical": 4}
	if riskLevels[ai.RiskLevel] > riskLevels[existing.RiskLevel] {
		existing.RiskLevel = ai.RiskLevel
	}

	// Merge boolean flags (OR operation)
	existing.BreakingChanges = existing.BreakingChanges || ai.BreakingChanges
	existing.SecurityImpact = existing.SecurityImpact || ai.SecurityImpact

	// Merge arrays (unique values)
	existing.AffectedAreas = rm.mergeStringSlices(existing.AffectedAreas, ai.AffectedAreas)
	existing.RequiredSkills = rm.mergeStringSlices(existing.RequiredSkills, ai.RequiredSkills)

	return existing
}

// MonitorReviewProgress monitors the progress of reviews
func (rm *ReviewManager) MonitorReviewProgress(ctx context.Context, pr *PullRequest) (*ReviewProgressReport, error) {
	report := &ReviewProgressReport{
		PullRequestNumber: pr.Number,
		OverallStatus:     "pending",
		ReviewStatuses:    []*ReviewStatus{},
		Summary:           &ReviewSummary{},
		LastUpdated:       time.Now(),
	}

	// Get current review statuses (simplified)
	reviewStatuses := []*ReviewStatus{
		{
			Reviewer:     "frontend-lead",
			Type:         "user",
			Status:       "approved",
			SubmittedAt:  time.Now().Add(-1 * time.Hour),
			Comments:     3,
			ChangesRequested: false,
		},
		{
			Reviewer:     "backend-team",
			Type:         "team",
			Status:       "pending",
			RequestedAt:  time.Now().Add(-30 * time.Minute),
			Comments:     0,
			ChangesRequested: false,
		},
	}
	report.ReviewStatuses = reviewStatuses

	// Calculate summary
	report.Summary = rm.calculateReviewSummary(reviewStatuses)

	// Determine overall status
	report.OverallStatus = rm.determineReviewStatus(reviewStatuses)

	return report, nil
}

// ReviewProgressReport represents the progress of reviews
type ReviewProgressReport struct {
	PullRequestNumber int             `json:"pull_request_number"`
	OverallStatus     string          `json:"overall_status"`
	ReviewStatuses    []*ReviewStatus `json:"review_statuses"`
	Summary           *ReviewSummary  `json:"summary"`
	LastUpdated       time.Time       `json:"last_updated"`
}

// ReviewStatus represents the status of a single review
type ReviewStatus struct {
	Reviewer         string     `json:"reviewer"`
	Type             string     `json:"type"`
	Status           string     `json:"status"`
	RequestedAt      time.Time  `json:"requested_at,omitempty"`
	SubmittedAt      time.Time  `json:"submitted_at,omitempty"`
	Comments         int        `json:"comments"`
	ChangesRequested bool       `json:"changes_requested"`
	Message          string     `json:"message,omitempty"`
}

// ReviewSummary provides a summary of review progress
type ReviewSummary struct {
	TotalReviewers      int `json:"total_reviewers"`
	ApprovedReviewers   int `json:"approved_reviewers"`
	PendingReviewers    int `json:"pending_reviewers"`
	ChangesRequested    int `json:"changes_requested"`
	TotalComments       int `json:"total_comments"`
	ReadyToMerge        bool `json:"ready_to_merge"`
}

// calculateReviewSummary calculates review summary
func (rm *ReviewManager) calculateReviewSummary(statuses []*ReviewStatus) *ReviewSummary {
	summary := &ReviewSummary{
		TotalReviewers: len(statuses),
	}

	for _, status := range statuses {
		switch status.Status {
		case "approved":
			summary.ApprovedReviewers++
		case "pending":
			summary.PendingReviewers++
		case "changes_requested":
			summary.ChangesRequested++
		}

		summary.TotalComments += status.Comments
	}

	// Determine if ready to merge (simplified logic)
	summary.ReadyToMerge = summary.ApprovedReviewers > 0 && summary.ChangesRequested == 0 && summary.PendingReviewers == 0

	return summary
}

// determineReviewStatus determines overall review status
func (rm *ReviewManager) determineReviewStatus(statuses []*ReviewStatus) string {
	hasChangesRequested := false
	hasPending := false
	hasApproved := false

	for _, status := range statuses {
		switch status.Status {
		case "changes_requested":
			hasChangesRequested = true
		case "pending":
			hasPending = true
		case "approved":
			hasApproved = true
		}
	}

	if hasChangesRequested {
		return "changes_requested"
	}
	if hasPending {
		return "pending"
	}
	if hasApproved {
		return "approved"
	}

	return "pending"
}

// Helper methods
func (rm *ReviewManager) extractAffectedAreas(title, body string) []string {
	areas := []string{}
	text := strings.ToLower(title + " " + body)

	areaKeywords := map[string][]string{
		"frontend": {"frontend", "ui", "react", "vue", "angular", "css", "html", "javascript"},
		"backend":  {"backend", "api", "server", "database", "sql", "rest", "graphql"},
		"devops":   {"devops", "docker", "kubernetes", "ci/cd", "deployment", "infrastructure"},
		"security": {"security", "auth", "authentication", "authorization", "encryption", "vulnerability"},
		"database": {"database", "sql", "nosql", "migration", "schema", "query"},
		"testing":  {"test", "testing", "unit", "integration", "e2e", "qa"},
	}

	for area, keywords := range areaKeywords {
		for _, keyword := range keywords {
			if strings.Contains(text, keyword) {
				areas = append(areas, area)
				break
			}
		}
	}

	return rm.uniqueStrings(areas)
}

func (rm *ReviewManager) extractRequiredSkills(title, body string) []string {
	skills := []string{}
	text := strings.ToLower(title + " " + body)

	skillKeywords := map[string][]string{
		"javascript": {"javascript", "js", "node", "npm", "react", "vue", "angular"},
		"python":     {"python", "django", "flask", "fastapi", "pip"},
		"go":         {"golang", "go"},
		"java":       {"java", "spring", "maven", "gradle"},
		"rust":       {"rust", "cargo"},
		"docker":     {"docker", "container", "dockerfile"},
		"kubernetes": {"kubernetes", "k8s", "helm"},
		"aws":        {"aws", "amazon", "s3", "ec2", "lambda"},
		"database":   {"sql", "postgresql", "mysql", "mongodb", "redis"},
	}

	for skill, keywords := range skillKeywords {
		for _, keyword := range keywords {
			if strings.Contains(text, keyword) {
				skills = append(skills, skill)
				break
			}
		}
	}

	return rm.uniqueStrings(skills)
}

func (rm *ReviewManager) hasSecurityKeywords(title, body string) bool {
	text := strings.ToLower(title + " " + body)
	securityKeywords := []string{
		"security", "vulnerability", "auth", "authentication", "authorization",
		"encryption", "decrypt", "password", "token", "jwt", "oauth", "ssl", "tls",
		"xss", "csrf", "sql injection", "sanitize", "validate",
	}

	for _, keyword := range securityKeywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}

	return false
}

func (rm *ReviewManager) mergeStringSlices(slice1, slice2 []string) []string {
	merged := append([]string{}, slice1...)
	
	for _, item := range slice2 {
		found := false
		for _, existing := range merged {
			if existing == item {
				found = true
				break
			}
		}
		if !found {
			merged = append(merged, item)
		}
	}

	return merged
}

func (rm *ReviewManager) uniqueStrings(slice []string) []string {
	keys := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}
