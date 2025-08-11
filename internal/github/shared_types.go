package github

import (
	"time"
)

// Shared types for GitHub integration to resolve redeclaration conflicts

// DependencyUpdate represents a dependency that needs to be updated
type DependencyUpdate struct {
	Name           string    `json:"name"`
	CurrentVersion string    `json:"current_version"`
	LatestVersion  string    `json:"latest_version"`
	PackageManager string    `json:"package_manager"`
	Repository     string    `json:"repository"`
	ChangelogURL   string    `json:"changelog_url,omitempty"`
	SecurityFix    bool      `json:"security_fix"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// BreakingChange represents a breaking change in a dependency update
type BreakingChange struct {
	Type             string   `json:"type"`
	Description      string   `json:"description"`
	AffectedAPIs     []string `json:"affected_apis,omitempty"`
	Severity         string   `json:"severity"`
	MigrationPath    string   `json:"migration_path,omitempty"`
	ExampleBefore    string   `json:"example_before,omitempty"`
	ExampleAfter     string   `json:"example_after,omitempty"`
	DocumentationURL string   `json:"documentation_url,omitempty"`
}

// Recommendation represents a recommendation for handling a dependency update
type Recommendation struct {
	Action      string `json:"action"`
	Description string `json:"description"`
	Priority    string `json:"priority"` // "required", "recommended", "optional"
	Effort      string `json:"effort"`   // "low", "medium", "high"
}

// PatchSuggestion represents a suggested code patch
type PatchSuggestion struct {
	OldCode     string  `json:"old_code"`
	NewCode     string  `json:"new_code"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

// RiskAssessment represents the risk assessment for a patch or update
type RiskAssessment struct {
	Level       string   `json:"level"`
	Score       float64  `json:"score"`
	Factors     []string `json:"factors"`
	Confidence  float64  `json:"confidence"`
	Mitigations []string `json:"mitigations"`
}

// PullRequest represents a GitHub pull request
type PullRequest struct {
	ID          int64     `json:"id"`
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	State       string    `json:"state"`
	Repository  string    `json:"repository"`
	Branch      string    `json:"branch"`
	BaseBranch  string    `json:"base_branch"`
	Author      string    `json:"author"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	MergedAt    *time.Time `json:"merged_at,omitempty"`
	URL         string    `json:"url"`
}

// Change represents a code change in a patch
type Change struct {
	File        string `json:"file"`
	OldContent  string `json:"old_content"`
	NewContent  string `json:"new_content"`
	LineNumber  int    `json:"line_number"`
	ChangeType  string `json:"change_type"` // "add", "remove", "modify"
	Description string `json:"description"`
}

// Patch represents a code patch for dependency updates
type Patch struct {
	ID          string    `json:"id"`
	Repository  string    `json:"repository"`
	Dependency  string    `json:"dependency"`
	Changes     []Change  `json:"changes"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	Status      string    `json:"status"`
}

// PatchValidator interface for validating patches
type PatchValidator interface {
	ValidatePatch(patch *Patch) error
	ValidateChanges(changes []Change) error
}

// AIManager interface for AI-powered analysis
type AIManager interface {
	AnalyzeDependencyUpdate(dependency *DependencyUpdate) (*BreakingChangeAnalysis, error)
	GeneratePatchSuggestions(dependency *DependencyUpdate, breakingChanges []BreakingChange) ([]PatchSuggestion, error)
}

// BreakingChangeAnalysis represents the result of breaking change analysis
type BreakingChangeAnalysis struct {
	Dependency         *DependencyUpdate     `json:"dependency"`
	HasBreakingChanges bool                  `json:"has_breaking_changes"`
	BreakingChanges    []*BreakingChange     `json:"breaking_changes"`
	RiskLevel          string                `json:"risk_level"`
	Confidence         float64               `json:"confidence"`
	Recommendations    []*Recommendation     `json:"recommendations"`
	PatchSuggestions   []*PatchSuggestion    `json:"patch_suggestions"`
	AnalyzedAt         time.Time             `json:"analyzed_at"`
	AnalysisSource     string                `json:"analysis_source"` // "ai", "heuristic", "changelog"
}
