package types

import (
	"context"
	"time"
)

// PackageManager defines the interface for package managers
type PackageManager interface {
	GetName() string
	GetVersion() string
	IsAvailable(ctx context.Context) bool
	DetectProjects(ctx context.Context, rootPath string) ([]Project, error)
	ParseDependencies(ctx context.Context, projectPath string) (*DependencyInfo, error)
	GetLatestVersion(ctx context.Context, packageName string, registry *RegistryConfig) (*VersionInfo, error)
	GetVersions(ctx context.Context, packageName string) ([]string, error)
	GetChangelog(ctx context.Context, packageName, version string, registry *RegistryConfig) (*ChangelogInfo, error)
	UpdateDependency(ctx context.Context, projectPath, packageName, version string, options *UpdateOptions) error
	InstallDependencies(ctx context.Context, projectPath string, options *InstallOptions) error
	ValidateProject(ctx context.Context, projectPath string) error
}

// Project represents a detected project
type Project struct {
	Name           string            `json:"name"`
	Path           string            `json:"path"`
	Type           string            `json:"type"`
	PackageManager string            `json:"package_manager"`
	ConfigFile     string            `json:"config_file"`
	Language       string            `json:"language"`
	Framework      string            `json:"framework,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	DetectedAt     time.Time         `json:"detected_at"`
}

// Dependency represents a package dependency
type Dependency struct {
	Name            string            `json:"name"`
	CurrentVersion  string            `json:"current_version"`
	LatestVersion   string            `json:"latest_version,omitempty"`
	Type            DependencyType    `json:"type"`
	Scope           string            `json:"scope,omitempty"`
	Source          string            `json:"source,omitempty"`
	PackageManager  string            `json:"package_manager"`
	UpdateAvailable bool              `json:"update_available"`
	SecurityIssues  []SecurityIssue   `json:"security_issues,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	LastChecked     time.Time         `json:"last_checked"`
}

// DependencyType represents the type of dependency
type DependencyType string

const (
	DependencyTypeProduction    DependencyType = "production"
	DependencyTypeDevelopment   DependencyType = "development"
	DependencyTypeOptional      DependencyType = "optional"
	DependencyTypePeer          DependencyType = "peer"
	DependencyTypeTest          DependencyType = "test"
	DependencyTypeBuild         DependencyType = "build"
	DependencyTypeRuntime       DependencyType = "runtime"
)

// SecurityIssue represents a security vulnerability
type SecurityIssue struct {
	ID          string    `json:"id"`
	CVE         string    `json:"cve,omitempty"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	CVSS        float64   `json:"cvss,omitempty"`
	URL         string    `json:"url,omitempty"`
	FixedIn     string    `json:"fixed_in,omitempty"`
	PublishedAt time.Time `json:"published_at"`
}

// VersionInfo represents version information for a package
type VersionInfo struct {
	Version     string            `json:"version"`
	PublishedAt time.Time         `json:"published_at"`
	IsPrerelease bool             `json:"is_prerelease"`
	Tags        []string          `json:"tags,omitempty"`
	Checksums   map[string]string `json:"checksums,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// ChangelogInfo represents changelog information
type ChangelogInfo struct {
	PackageName   string    `json:"package_name"`
	FromVersion   string    `json:"from_version"`
	ToVersion     string    `json:"to_version"`
	Version       string    `json:"version"`
	Content       string    `json:"content"`
	Description   string    `json:"description"`
	URL           string    `json:"url,omitempty"`
	ReleaseNotes  string    `json:"release_notes,omitempty"`
	IsBreaking    bool      `json:"is_breaking"`
	SecurityFix   bool      `json:"security_fix"`
	RetrievedAt   time.Time `json:"retrieved_at"`
}

// UpdateRequest represents a dependency update request
type UpdateRequest struct {
	ProjectPath   string            `json:"project_path"`
	PackageName   string            `json:"package_name"`
	FromVersion   string            `json:"from_version"`
	ToVersion     string            `json:"to_version"`
	UpdateType    string            `json:"update_type"`
	Options       map[string]string `json:"options,omitempty"`
	DryRun        bool              `json:"dry_run"`
}

// UpdateResult represents the result of a dependency update
type UpdateResult struct {
	PackageName   string            `json:"package_name"`
	FromVersion   string            `json:"from_version"`
	ToVersion     string            `json:"to_version"`
	Success       bool              `json:"success"`
	Error         string            `json:"error,omitempty"`
	ChangedFiles  []string          `json:"changed_files,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// Additional types needed by existing code

// ProjectInfo contains information about a detected project
type ProjectInfo struct {
	Name           string            `json:"name"`
	Path           string            `json:"path"`
	Type           string            `json:"type"`
	PackageManager string            `json:"package_manager"`
	ConfigFile     string            `json:"config_file"`
	Language       string            `json:"language"`
	Framework      string            `json:"framework,omitempty"`
	Version        string            `json:"version,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	DetectedAt     time.Time         `json:"detected_at"`
}

// DependencyInfo contains parsed dependency information
type DependencyInfo struct {
	ProjectName          string            `json:"project_name"`
	ProjectVersion       string            `json:"project_version"`
	Dependencies         []DependencyEntry `json:"dependencies"`
	DevDependencies      []DependencyEntry `json:"dev_dependencies,omitempty"`
	OptionalDeps         []DependencyEntry `json:"optional_dependencies,omitempty"`
	OptionalDependencies []DependencyEntry `json:"optional_dependencies,omitempty"`
	PeerDependencies     []DependencyEntry `json:"peer_dependencies,omitempty"`
	ConfigFile           string            `json:"config_file"`
	Metadata             map[string]string `json:"metadata,omitempty"`
	ParsedAt             time.Time         `json:"parsed_at"`
}

// DependencyEntry represents a single dependency
type DependencyEntry struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	ResolvedVersion string            `json:"resolved_version,omitempty"`
	Type            string            `json:"type"`
	Scope           string            `json:"scope,omitempty"`
	Source          string            `json:"source,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

// RegistryConfig contains registry configuration
type RegistryConfig struct {
	URL         string            `json:"url"`
	Username    string            `json:"username,omitempty"`
	Password    string            `json:"password,omitempty"`
	Token       string            `json:"token,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Timeout     time.Duration     `json:"timeout,omitempty"`
	Insecure    bool              `json:"insecure,omitempty"`
}

// UpdateOptions contains options for updating dependencies
type UpdateOptions struct {
	DryRun      bool              `json:"dry_run"`
	SaveExact   bool              `json:"save_exact"`
	SaveDev     bool              `json:"save_dev"`
	Force       bool              `json:"force"`
	Registry    *RegistryConfig   `json:"registry,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
}

// InstallOptions contains options for installing dependencies
type InstallOptions struct {
	Production  bool              `json:"production"`
	Development bool              `json:"development"`
	Clean       bool              `json:"clean"`
	Force       bool              `json:"force"`
	Registry    *RegistryConfig   `json:"registry,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
}
