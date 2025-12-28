package github

import (
	"context"
	"time"

	"github.com/8tcapital/superint-dep-manager/internal/models"
)

// Shared types for GitHub integration to resolve redeclaration conflicts

// DependencyUpdate represents a dependency that needs to be updated
type DependencyUpdate struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	CurrentVersion string    `json:"current_version"`
	LatestVersion  string    `json:"latest_version"`
	PackageManager string    `json:"package_manager"`
	Repository     string    `json:"repository"`
	ChangelogURL   string    `json:"changelog_url,omitempty"`
	SecurityFix    bool      `json:"security_fix"`
	UpdatedAt      time.Time `json:"updated_at"`
	DependencyName string    `json:"dependency_name,omitempty"` // Alias for Name
	TargetVersion  string    `json:"target_version,omitempty"`  // Alias for LatestVersion
	Priority       string    `json:"priority,omitempty"`
	UpdateType     string    `json:"update_type,omitempty"`
	RiskLevel      string    `json:"risk_level,omitempty"`
	AffectedFiles  []string  `json:"affected_files,omitempty"`
	BreakingChange bool      `json:"breaking_change,omitempty"`
}

// DependencyInfo is an alias for DependencyUpdate
type DependencyInfo = DependencyUpdate

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
	ID          string    `json:"id,omitempty"`
	Type        string    `json:"type,omitempty"`
	Title       string    `json:"title,omitempty"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	Priority    string    `json:"priority"` // "required", "recommended", "optional"
	Effort      string    `json:"effort"`   // "low", "medium", "high"
	Impact      string    `json:"impact,omitempty"`
	Confidence  float64   `json:"confidence"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
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
	OverallRisk     string        `json:"overall_risk"`
	FilesModified   int           `json:"files_modified"`
	Reversible      bool          `json:"reversible"`
	Recommendations []string      `json:"recommendations"`
	BreakingChanges int           `json:"breaking_changes"`
	RiskFactors     []*RiskFactor `json:"risk_factors"`
	TestingRequired bool          `json:"testing_required"`
	ReviewRequired  bool          `json:"review_required"`
	Confidence      float64       `json:"confidence"`
	Mitigations     []string      `json:"mitigations"`
	Level           string        `json:"level,omitempty"`   // Deprecated: use OverallRisk
	Score           float64       `json:"score,omitempty"`   // Deprecated
	Factors         []string      `json:"factors,omitempty"` // Deprecated: use RiskFactors
}

// RiskFactor represents a specific risk factor identified in the assessment
type RiskFactor struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Likelihood  string `json:"likelihood"`
	Impact      string `json:"impact"`
}

// PullRequest represents a GitHub pull request
type PullRequest struct {
	ID         int64  `json:"id"`
	Number     int    `json:"number"`
	Title      string `json:"title"`
	Body       string `json:"body"`
	State      string `json:"state"`
	Repository string `json:"repository"` // This might be an object in API, but kept as string if used as such?
	// GitHub API returns "user" object for author
	User      *User              `json:"user"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
	MergedAt  *time.Time         `json:"merged_at,omitempty"`
	URL       string             `json:"url"`
	HTMLURL   string             `json:"html_url"`
	Head      *PullRequestBranch `json:"head"`
	Base      *PullRequestBranch `json:"base"`
	Labels    []string           `json:"labels,omitempty"`
	Assignees []string           `json:"assignees,omitempty"`
	Reviewers []string           `json:"reviewers,omitempty"`
	Milestone string             `json:"milestone,omitempty"`
}

// PullRequestBranch represents a branch in a pull request
type PullRequestBranch struct {
	Label string      `json:"label"`
	Ref   string      `json:"ref"`
	SHA   string      `json:"sha"`
	User  *User       `json:"user"`
	Repo  *Repository `json:"repo"`
}

// Change represents a code change in a patch
// Change represents a code change in a patch
type Change struct {
	File        string  `json:"file"`
	OldContent  string  `json:"old_content"`
	NewContent  string  `json:"new_content"`
	LineNumber  int     `json:"line_number"`
	LineStart   int     `json:"line_start,omitempty"`
	LineEnd     int     `json:"line_end,omitempty"`
	Type        string  `json:"type"` // "add", "remove", "modify"
	Description string  `json:"description"`
	Reason      string  `json:"reason,omitempty"`
	Confidence  float64 `json:"confidence,omitempty"`
}

// FilePatch represents a patch for a specific file
type FilePatch struct {
	Path        string    `json:"path"`
	Type        string    `json:"type"` // "modify", "create", "delete"
	Language    string    `json:"language"`
	Changes     []*Change `json:"changes"`
	NewContent  string    `json:"new_content,omitempty"`
	Confidence  float64   `json:"confidence"`
	Description string    `json:"description"`
}

// ConfigPatch represents a patch for a configuration file
type ConfigPatch struct {
	File        string                 `json:"file"`
	Type        string                 `json:"type"`
	Changes     map[string]interface{} `json:"changes"`
	Description string                 `json:"description"`
}

// Patch represents a code patch for dependency updates
type Patch struct {
	ID             string        `json:"id"`
	Repository     string        `json:"repository"`
	Dependency     string        `json:"dependency"`
	Changes        []*Change     `json:"changes"`
	FilePatches    []FilePatch   `json:"file_patches"`
	ConfigPatches  []ConfigPatch `json:"config_patches"`
	Description    string        `json:"description"`
	CreatedAt      time.Time     `json:"created_at"`
	Status         string        `json:"status"`
	Confidence     float64       `json:"confidence,omitempty"`
	RiskLevel      string        `json:"risk_level,omitempty"`
	Type           string        `json:"type,omitempty"` // "security", "bugfix", "feature", "chore"
	BreakingChange bool          `json:"breaking_change,omitempty"`
}

// PatchValidator interface for validating patches
type PatchValidator interface {
	ValidatePatch(ctx context.Context, patch *Patch, options *ValidationOptions) (*ValidationResult, error)
	ValidateChanges(changes []*Change) error
}

// AIManager interface for Superintelligence-powered analysis
type AIManager interface {
	AnalyzeDependencyUpdate(ctx context.Context, dependency *DependencyUpdate) (*BreakingChangeAnalysis, error)
	GeneratePatchSuggestions(ctx context.Context, dependency *DependencyUpdate, breakingChanges []BreakingChange) ([]PatchSuggestion, error)
	AnalyzeChangelog(ctx context.Context, prompt string) (string, error)
	GeneratePatches(ctx context.Context, request *models.PatchGenerationRequest) ([]*models.GeneratedPatch, error)
}

// BreakingChangeAnalysis represents the result of breaking change analysis
type BreakingChangeAnalysis struct {
	Dependency         *DependencyUpdate  `json:"dependency"`
	HasBreakingChanges bool               `json:"has_breaking_changes"`
	BreakingChanges    []*BreakingChange  `json:"breaking_changes"`
	RiskLevel          string             `json:"risk_level"`
	Confidence         float64            `json:"confidence"`
	Recommendations    []*Recommendation  `json:"recommendations"`
	PatchSuggestions   []*PatchSuggestion `json:"patch_suggestions"`
	AnalyzedAt         time.Time          `json:"analyzed_at"`
	AnalysisSource     string             `json:"analysis_source"` // "ai", "heuristic", "changelog"
}
