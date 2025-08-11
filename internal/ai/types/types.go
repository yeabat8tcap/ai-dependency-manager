package types

import (
	"context"
	"time"
)

// AIProvider defines the interface for AI analysis providers
type AIProvider interface {
	GetName() string
	GetVersion() string
	IsAvailable(ctx context.Context) bool
	AnalyzeChangelog(ctx context.Context, request *ChangelogAnalysisRequest) (*ChangelogAnalysisResponse, error)
	AnalyzeVersionDiff(ctx context.Context, request *VersionDiffAnalysisRequest) (*VersionDiffAnalysisResponse, error)
	PredictCompatibility(ctx context.Context, request *CompatibilityPredictionRequest) (*CompatibilityPredictionResponse, error)
	ClassifyUpdate(ctx context.Context, request *UpdateClassificationRequest) (*UpdateClassificationResponse, error)
}

// ChangelogAnalysisRequest represents a request to analyze changelog text
type ChangelogAnalysisRequest struct {
	PackageName    string
	FromVersion    string
	ToVersion      string
	ChangelogText  string
	ReleaseNotes   string
	PackageManager string
	Language       string
}

// ChangelogAnalysisResponse represents the result of changelog analysis
type ChangelogAnalysisResponse struct {
	PackageName       string
	FromVersion       string
	ToVersion         string
	HasBreakingChange bool
	BreakingChanges   []BreakingChange
	NewFeatures       []Feature
	BugFixes          []BugFix
	SecurityFixes     []SecurityFix
	DeprecatedAPIs    []DeprecatedAPI
	Deprecations      []Deprecation
	RiskLevel         RiskLevel
	RiskScore         float64
	Confidence        float64
	ConfidenceScore   float64
	Recommendations   []string
	Summary           string
	AnalyzedAt        time.Time
}

// VersionDiffAnalysisRequest represents a request to analyze version differences
type VersionDiffAnalysisRequest struct {
	PackageName    string
	FromVersion    string
	ToVersion      string
	DiffText       string
	FileChanges    []FileChange
	PackageManager string
	Language       string
}

// VersionDiffAnalysisResponse represents the result of version diff analysis
type VersionDiffAnalysisResponse struct {
	PackageName     string
	FromVersion     string
	ToVersion       string
	UpdateType      string
	SemanticImpact  string
	APIChanges      []APIChange
	BehaviorChanges []BehaviorChange
	RiskLevel       RiskLevel
	RiskScore       float64
	Confidence      float64
	ConfidenceScore float64
	Recommendations []string
	Summary         string
	ProcessedAt     time.Time
	AnalyzedAt      time.Time
}

// CompatibilityPredictionRequest represents a request for compatibility prediction
type CompatibilityPredictionRequest struct {
	PackageName     string
	FromVersion     string
	ToVersion       string
	ProjectContext  ProjectContext
	DependencyGraph []Dependency
	PackageManager  string
	Language        string
}

// CompatibilityPredictionResponse represents the result of compatibility prediction
type CompatibilityPredictionResponse struct {
	PackageName             string
	FromVersion             string
	ToVersion               string
	CompatibilityScore      float64
	RiskLevel               RiskLevel
	RiskScore               float64
	Confidence              float64
	ConfidenceScore         float64
	PotentialIssues         []CompatibilityIssue
	MigrationSteps          []string
	TestingRecommendations  []string
	Recommendations         []string
	Summary                 string
	ProcessedAt             time.Time
	PredictedAt             time.Time
}

// CompatibilityResponse is an alias for backward compatibility
type CompatibilityResponse = CompatibilityPredictionResponse

// UpdateClassificationRequest represents a request to classify an update
type UpdateClassificationRequest struct {
	PackageName    string
	FromVersion    string
	ToVersion      string
	ChangelogText  string
	ReleaseNotes   string
	PackageManager string
	Language       string
	ProjectContext ProjectContext
}

// UpdateClassificationResponse represents the result of update classification
type UpdateClassificationResponse struct {
	PackageName     string
	FromVersion     string
	ToVersion       string
	UpdateType      string
	Categories      []UpdateCategory
	Priority        Priority
	RiskLevel       RiskLevel
	Urgency         Urgency
	RiskScore       float64
	Confidence      float64
	ConfidenceScore float64
	Reasons         []string
	Recommendations []string
	Summary         string
	ProcessedAt     time.Time
	ClassifiedAt    time.Time
}

// Supporting types
type BreakingChange struct {
	Type        string
	Description string
	Impact      string
	Confidence  float64
	Mitigation  string
}

type Feature struct {
	Name        string
	Description string
	Type        string
	Impact      string
	Confidence  float64
}

type BugFix struct {
	Description string
	Impact      string
	Confidence  float64
	Severity    string
}

type SecurityFix struct {
	CVE         string
	Severity    string
	Description string
	Impact      string
	Confidence  float64
}

type DeprecatedAPI struct {
	API         string
	Replacement string
	Timeline    string
}

// Deprecation represents a deprecation in an update
type Deprecation struct {
	Item        string  `json:"item"`
	Replacement string  `json:"replacement,omitempty"`
	Timeline    string  `json:"timeline,omitempty"`
	Confidence  float64 `json:"confidence"`
}

// UpdateCategory represents a category of update
type UpdateCategory struct {
	Name        string  `json:"name"`
	Weight      float64 `json:"weight"`
	Description string  `json:"description,omitempty"`
}

// Urgency represents the urgency of an update
type Urgency string

const (
	UrgencyLow       Urgency = "low"
	UrgencyMedium    Urgency = "medium"
	UrgencyHigh      Urgency = "high"
	UrgencyImmediate Urgency = "immediate"
)

type FileChange struct {
	Path      string
	Type      string
	LinesAdded   int
	LinesRemoved int
}

type APIChange struct {
	Type        string
	API         string
	Description string
	Impact      string
}

type BehaviorChange struct {
	Component   string
	Description string
	Impact      string
}

type ProjectContext struct {
	Language     string
	Framework    string
	Dependencies []string
	TestCoverage float64
}

type Dependency struct {
	Name    string
	Version string
	Type    string
}

type CompatibilityIssue struct {
	Type        string
	Description string
	Severity    string
	Likelihood  float64
	Mitigation  string
}

// Enums
type UpdateType string

const (
	UpdateTypeMajor UpdateType = "major"
	UpdateTypeMinor UpdateType = "minor"
	UpdateTypePatch UpdateType = "patch"
)

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
	PriorityCritical Priority = "critical"
)

type RiskLevel string

const (
	RiskLevelLow    RiskLevel = "low"
	RiskLevelMedium RiskLevel = "medium"
	RiskLevelHigh   RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)
