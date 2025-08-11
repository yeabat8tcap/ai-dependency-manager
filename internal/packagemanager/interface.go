package packagemanager

import (
	"context"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/packagemanager/types"
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
	ProjectName    string                 `json:"project_name"`
	ProjectVersion string                 `json:"project_version"`
	Dependencies   []DependencyEntry      `json:"dependencies"`
	DevDependencies []DependencyEntry     `json:"dev_dependencies,omitempty"`
	PeerDependencies []DependencyEntry    `json:"peer_dependencies,omitempty"`
	OptionalDependencies []DependencyEntry `json:"optional_dependencies,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// DependencyEntry represents a single dependency
type DependencyEntry struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`          // Version constraint (e.g., "^1.2.3", ">=2.0.0")
	ResolvedVersion string            `json:"resolved_version"` // Actual installed version
	Type            string            `json:"type"`             // direct, dev, peer, optional
	Registry        string            `json:"registry,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	// Added a new field to DependencyEntry
	License        string            `json:"license,omitempty"`
}



// RegistryConfig contains registry configuration
type RegistryConfig struct {
	URL         string            `json:"url"`
	Name        string            `json:"name,omitempty"`
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
	Force       bool              `json:"force"`
	SaveExact   bool              `json:"save_exact"`
	Registry    *RegistryConfig   `json:"registry,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Timeout     time.Duration     `json:"timeout,omitempty"`
}

// InstallOptions contains options for installing dependencies
type InstallOptions struct {
	Clean       bool              `json:"clean"`        // Clean install (remove node_modules, etc.)
	Production  bool              `json:"production"`   // Install only production dependencies
	Registry    *RegistryConfig   `json:"registry,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Timeout     time.Duration     `json:"timeout,omitempty"`
}

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
