package packagemanager

import (
	"context"

	"github.com/8tcapital/ai-dep-manager/internal/packagemanager/maven"
	"github.com/8tcapital/ai-dep-manager/internal/packagemanager/npm"
	"github.com/8tcapital/ai-dep-manager/internal/packagemanager/pip"
	"github.com/8tcapital/ai-dep-manager/internal/packagemanager/types"
)

// DefaultManager is the global package manager registry
var DefaultManager *Manager

// init initializes the default package manager registry
func init() {
	DefaultManager = NewManager()
	
	// Register all supported package managers
	DefaultManager.Register(npm.NewNPMManager())
	DefaultManager.Register(pip.NewPipManager())
	DefaultManager.Register(maven.NewMavenManager())
}

// GetManager returns the default package manager registry
func GetManager() *Manager {
	return DefaultManager
}

// GetPackageManager returns a specific package manager by type
func GetPackageManager(pmType string) (PackageManager, bool) {
	return DefaultManager.Get(pmType)
}

// GetAvailablePackageManagers returns all available package managers on the system
func GetAvailablePackageManagers(ctx context.Context) map[string]PackageManager {
	return DefaultManager.GetAvailable(ctx)
}

// DetectAllProjects detects all projects in a directory using all available package managers
func DetectAllProjects(ctx context.Context, rootPath string) ([]types.Project, error) {
	return DefaultManager.DetectProjectTypes(ctx, rootPath)
}
