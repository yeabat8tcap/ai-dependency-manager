package packagemanager

import (
	"context"

	"github.com/8tcapital/superint-dep-manager/internal/packagemanager/types"
)

// PackageManager is an alias to the types.PackageManager interface
type PackageManager = types.PackageManager

// Type aliases for convenience
type Project = types.Project
type Dependency = types.Dependency
type DependencyType = types.DependencyType
type SecurityIssue = types.SecurityIssue
type VersionInfo = types.VersionInfo
type ChangelogInfo = types.ChangelogInfo
type UpdateRequest = types.UpdateRequest
type UpdateResult = types.UpdateResult

// ProjectInfo contains information about a detected project
type ProjectInfo struct {
	Name           string            `json:"name"`
	Path           string            `json:"path"`
	ConfigFile     string            `json:"config_file"`
	PackageManager string            `json:"package_manager"`
	Version        string            `json:"version,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// DependencyInfo contains parsed dependency information
type DependencyInfo struct {
	ProjectName          string                 `json:"project_name"`
	ProjectVersion       string                 `json:"project_version"`
	Dependencies         []DependencyEntry      `json:"dependencies"`
	DevDependencies      []DependencyEntry      `json:"dev_dependencies,omitempty"`
	PeerDependencies     []DependencyEntry      `json:"peer_dependencies,omitempty"`
	OptionalDependencies []DependencyEntry      `json:"optional_dependencies,omitempty"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
}

// Type aliases for convenience
type DependencyEntry = types.DependencyEntry
type RegistryConfig = types.RegistryConfig
type UpdateOptions = types.UpdateOptions
type InstallOptions = types.InstallOptions

// Manager is the main package manager registry
type Manager struct {
	managers map[string]PackageManager
}

// NewManager creates a new package manager registry
func NewManager() *Manager {
	return &Manager{
		managers: make(map[string]PackageManager),
	}
}

// Register registers a package manager
func (m *Manager) Register(pm PackageManager) {
	// Get the type from the package manager's GetType method
	pmType := ""
	switch p := pm.(type) {
	case interface{ GetType() string }:
		pmType = p.GetType()
	default:
		// Fallback - this shouldn't happen with proper implementations
		pmType = "unknown"
	}
	m.managers[pmType] = pm
}

// Get returns a package manager by type
func (m *Manager) Get(pmType string) (PackageManager, bool) {
	pm, exists := m.managers[pmType]
	return pm, exists
}

// GetAll returns all registered package managers
func (m *Manager) GetAll() map[string]PackageManager {
	result := make(map[string]PackageManager)
	for k, v := range m.managers {
		result[k] = v
	}
	return result
}

// GetAvailable returns all available package managers on the system
func (m *Manager) GetAvailable(ctx context.Context) map[string]PackageManager {
	available := make(map[string]PackageManager)
	for pmType, pm := range m.managers {
		if pm.IsAvailable(ctx) {
			available[pmType] = pm
		}
	}
	return available
}

// DetectProjectTypes detects which package managers are used in a directory
func (m *Manager) DetectProjectTypes(ctx context.Context, rootPath string) ([]types.Project, error) {
	var allProjects []types.Project

	for _, pm := range m.managers {
		if !pm.IsAvailable(ctx) {
			continue
		}

		projects, err := pm.DetectProjects(ctx, rootPath)
		if err != nil {
			// Log error but continue with other package managers
			continue
		}

		allProjects = append(allProjects, projects...)
	}

	return allProjects, nil
}
