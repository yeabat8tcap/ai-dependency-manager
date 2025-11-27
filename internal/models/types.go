package models

// BreakingChange represents a breaking change in an update
type BreakingChange struct {
	Type             string   `json:"type"`
	Description      string   `json:"description"`
	AffectedAPIs     []string `json:"affected_apis"`
	Severity         string   `json:"severity"`
	MigrationPath    string   `json:"migration_path"`
	ExampleBefore    string   `json:"example_before"`
	ExampleAfter     string   `json:"example_after"`
	DocumentationURL string   `json:"documentation_url"`
}

// Risk levels for dependency updates
const (
	RiskLow      = "low"
	RiskMedium   = "medium"
	RiskHigh     = "high"
	RiskCritical = "critical"
)

// DependencyInfo represents information about a dependency
type DependencyInfo struct {
	Name           string   `json:"name"`
	CurrentVersion string   `json:"current_version"`
	LatestVersion  string   `json:"latest_version"`
	Repository     string   `json:"repository"`
	Language       string   `json:"language"`
	PackageManager string   `json:"package_manager"`
	License        string   `json:"license"`
	SecurityFix    bool     `json:"security_fix"`
	BreakingChange bool     `json:"breaking_change"`
	AffectedFiles  []string `json:"affected_files"`
}

// PatchGenerationRequest represents a request to generate a patch
type PatchGenerationRequest struct {
	PackageName     string            `json:"package_name"`
	CurrentVersion  string            `json:"current_version"`
	TargetVersion   string            `json:"target_version"`
	BreakingChanges []*BreakingChange `json:"breaking_changes"`
	ProjectType     string            `json:"project_type"`
	AffectedFiles   []string          `json:"affected_files"`
}

// GeneratedPatch represents a generated patch for a file
type GeneratedPatch struct {
	FilePath    string    `json:"file_path"`
	Changes     []*Change `json:"changes"`
	Confidence  float64   `json:"confidence"`
	Description string    `json:"description"`
}

// Change represents a specific change in a file
type Change struct {
	LineStart  int     `json:"line_start"`
	LineEnd    int     `json:"line_end"`
	OldContent string  `json:"old_content"`
	NewContent string  `json:"new_content"`
	Type       string  `json:"type"` // replace, insert, delete
	Reason     string  `json:"reason"`
	Confidence float64 `json:"confidence"`
}

// AnalysisRequest represents a request for dependency analysis
type AnalysisRequest struct {
	PackageName    string `json:"package_name"`
	CurrentVersion string `json:"current_version"`
	TargetVersion  string `json:"target_version"`
	PackageManager string `json:"package_manager"`
	ChangelogURL   string `json:"changelog_url"`
}
