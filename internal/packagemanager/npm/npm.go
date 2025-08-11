package npm

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/logger"
	"github.com/8tcapital/ai-dep-manager/internal/packagemanager/types"
)

// NPMManager implements PackageManager interface for npm
type NPMManager struct {
	httpClient *http.Client
}

// NewNPMManager creates a new NPM package manager
func NewNPMManager() *NPMManager {
	return &NPMManager{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName returns the display name of the package manager
func (n *NPMManager) GetName() string {
	return "Node Package Manager"
}

// GetType returns the type identifier
func (n *NPMManager) GetType() string {
	return "npm"
}

// GetVersion returns the version of the package manager
func (n *NPMManager) GetVersion() string {
	return "1.0.0"
}

// GetVersions returns available versions for a package
func (n *NPMManager) GetVersions(ctx context.Context, packageName string) ([]string, error) {
	// This is a placeholder implementation
	// In a real implementation, this would query the npm registry
	return []string{}, nil
}

// IsAvailable checks if npm is available on the system
func (n *NPMManager) IsAvailable(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "npm", "--version")
	err := cmd.Run()
	return err == nil
}

// DetectProjects scans for npm projects in the given directory
func (n *NPMManager) DetectProjects(ctx context.Context, rootPath string) ([]types.Project, error) {
	var projects []types.Project
	
	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		// Skip node_modules directories
		if d.IsDir() && d.Name() == "node_modules" {
			return filepath.SkipDir
		}
		
		// Look for package.json files
		if !d.IsDir() && d.Name() == "package.json" {
			projectInfo, err := n.parseProjectInfo(path)
			if err != nil {
				logger.Warn("Failed to parse package.json at %s: %v", path, err)
				return nil
			}
			
			projects = append(projects, *projectInfo)
		}
		
		return nil
	})
	
	return projects, err
}

// parseProjectInfo parses a package.json file to extract project information
func (n *NPMManager) parseProjectInfo(packageJsonPath string) (*types.Project, error) {
	data, err := os.ReadFile(packageJsonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read package.json: %w", err)
	}
	
	var packageJson struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	
	if err := json.Unmarshal(data, &packageJson); err != nil {
		return nil, fmt.Errorf("failed to parse package.json: %w", err)
	}
	
	projectPath := filepath.Dir(packageJsonPath)
	
	return &types.Project{
		Name:           packageJson.Name,
		Path:           projectPath,
		Type:           "npm",
		PackageManager: "npm",
		ConfigFile:     packageJsonPath,
		Language:       "javascript",
		Metadata: map[string]string{
			"config_file": packageJsonPath,
			"version":     packageJson.Version,
		},
		DetectedAt: time.Now(),
	}, nil
}

// ParseDependencies reads and parses package.json and package-lock.json
func (n *NPMManager) ParseDependencies(ctx context.Context, projectPath string) (*types.DependencyInfo, error) {
	packageJsonPath := filepath.Join(projectPath, "package.json")
	
	data, err := os.ReadFile(packageJsonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read package.json: %w", err)
	}
	
	var packageJson struct {
		Name                 string            `json:"name"`
		Version              string            `json:"version"`
		Dependencies         map[string]string `json:"dependencies"`
		DevDependencies      map[string]string `json:"devDependencies"`
		PeerDependencies     map[string]string `json:"peerDependencies"`
		OptionalDependencies map[string]string `json:"optionalDependencies"`
	}
	
	if err := json.Unmarshal(data, &packageJson); err != nil {
		return nil, fmt.Errorf("failed to parse package.json: %w", err)
	}
	
	// Parse installed versions from package-lock.json if available
	installedVersions := make(map[string]string)
	lockfilePath := filepath.Join(projectPath, "package-lock.json")
	if lockData, err := os.ReadFile(lockfilePath); err == nil {
		var lockfile struct {
			Packages map[string]struct {
				Version string `json:"version"`
			} `json:"packages"`
		}
		
		if json.Unmarshal(lockData, &lockfile) == nil {
			for pkgPath, info := range lockfile.Packages {
				if pkgPath == "" {
					continue // Skip root package
				}
				// Extract package name from path (remove node_modules prefix)
				pkgName := strings.TrimPrefix(pkgPath, "node_modules/")
				installedVersions[pkgName] = info.Version
			}
		}
	}
	
	depInfo := &types.DependencyInfo{
		ProjectName:    packageJson.Name,
		ProjectVersion: packageJson.Version,
		Dependencies:   n.convertDependencies(packageJson.Dependencies, "direct", installedVersions),
		DevDependencies: n.convertDependencies(packageJson.DevDependencies, "dev", installedVersions),
		PeerDependencies: n.convertDependencies(packageJson.PeerDependencies, "peer", installedVersions),
		OptionalDependencies: n.convertDependencies(packageJson.OptionalDependencies, "optional", installedVersions),
	}
	
	return depInfo, nil
}

// convertDependencies converts npm dependency map to DependencyEntry slice
func (n *NPMManager) convertDependencies(deps map[string]string, depType string, installedVersions map[string]string) []types.DependencyEntry {
	var entries []types.DependencyEntry
	
	for name, version := range deps {
		entry := types.DependencyEntry{
			Name:            name,
			Version:         version,
			ResolvedVersion: installedVersions[name],
			Type:            depType,
		}
		entries = append(entries, entry)
	}
	
	return entries
}

// GetLatestVersion fetches the latest version from npm registry
func (n *NPMManager) GetLatestVersion(ctx context.Context, packageName string, registry *types.RegistryConfig) (*types.VersionInfo, error) {
	registryURL := "https://registry.npmjs.org"
	if registry != nil && registry.URL != "" {
		registryURL = registry.URL
	}
	
	url := fmt.Sprintf("%s/%s/latest", registryURL, packageName)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Add authentication if provided
	if registry != nil {
		if registry.Token != "" {
			req.Header.Set("Authorization", "Bearer "+registry.Token)
		}
		for key, value := range registry.Headers {
			req.Header.Set(key, value)
		}
	}
	
	resp, err := n.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch package info: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned status %d", resp.StatusCode)
	}
	
	var packageInfo struct {
		Version string `json:"version"`
		Time    string `json:"time"`
		Dist    struct {
			Shasum string `json:"shasum"`
		} `json:"dist"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&packageInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	publishedAt, _ := time.Parse(time.RFC3339, packageInfo.Time)
	
	return &types.VersionInfo{
		Version:     packageInfo.Version,
		PublishedAt: publishedAt,
		Checksums: map[string]string{
			"sha1": packageInfo.Dist.Shasum,
		},
	}, nil
}

// GetVersionHistory fetches version history for a package
func (n *NPMManager) GetVersionHistory(ctx context.Context, packageName string, registry *types.RegistryConfig) ([]types.VersionInfo, error) {
	registryURL := "https://registry.npmjs.org"
	if registry != nil && registry.URL != "" {
		registryURL = registry.URL
	}
	
	url := fmt.Sprintf("%s/%s", registryURL, packageName)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Add authentication if provided
	if registry != nil {
		if registry.Token != "" {
			req.Header.Set("Authorization", "Bearer "+registry.Token)
		}
		for key, value := range registry.Headers {
			req.Header.Set(key, value)
		}
	}
	
	resp, err := n.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch package info: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned status %d", resp.StatusCode)
	}
	
	var packageInfo struct {
		Versions map[string]struct {
			Version string `json:"version"`
		} `json:"versions"`
		Time map[string]string `json:"time"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&packageInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	var versions []types.VersionInfo
	for version, timeStr := range packageInfo.Time {
		if version == "created" || version == "modified" {
			continue
		}
		
		publishedAt, _ := time.Parse(time.RFC3339, timeStr)
		
		versionInfo := types.VersionInfo{
			Version:     version,
			PublishedAt: publishedAt,
		}
		
		versions = append(versions, versionInfo)
	}
	
	return versions, nil
}

// GetChangelog fetches changelog/release notes for a version
func (n *NPMManager) GetChangelog(ctx context.Context, packageName, version string, registry *types.RegistryConfig) (*types.ChangelogInfo, error) {
	// For npm, we'll try to get changelog from the package metadata
	// This is a simplified implementation - in practice, you'd want to check
	// various sources like GitHub releases, CHANGELOG files, etc.
	
	registryURL := "https://registry.npmjs.org"
	if registry != nil && registry.URL != "" {
		registryURL = registry.URL
	}
	
	url := fmt.Sprintf("%s/%s/%s", registryURL, packageName, version)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := n.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch version info: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned status %d", resp.StatusCode)
	}
	
	var versionInfo struct {
		Version     string `json:"version"`
		Description string `json:"description"`
		Repository  struct {
			URL string `json:"url"`
		} `json:"repository"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&versionInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &types.ChangelogInfo{
		Version:     version,
		Description: versionInfo.Description,
		URL:         versionInfo.Repository.URL,
		// Note: Breaking change detection would be implemented in AI module
		IsBreaking:  false,
		SecurityFix: false,
	}, nil
}

// UpdateDependency updates a specific dependency
func (n *NPMManager) UpdateDependency(ctx context.Context, projectPath, packageName, newVersion string, options *types.UpdateOptions) error {
	args := []string{"install", fmt.Sprintf("%s@%s", packageName, newVersion)}
	
	if options != nil {
		if options.SaveExact {
			args = append(args, "--save-exact")
		}
		if options.DryRun {
			args = append(args, "--dry-run")
		}
		if options.Registry != nil && options.Registry.URL != "" {
			args = append(args, "--registry", options.Registry.URL)
		}
	}
	
	cmd := exec.CommandContext(ctx, "npm", args...)
	cmd.Dir = projectPath
	
	if options != nil && options.Environment != nil {
		cmd.Env = append(os.Environ())
		for key, value := range options.Environment {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("npm install failed: %w\nOutput: %s", err, string(output))
	}
	
	logger.Info("Successfully updated %s to %s", packageName, newVersion)
	return nil
}

// InstallDependencies installs all dependencies
func (n *NPMManager) InstallDependencies(ctx context.Context, projectPath string, options *types.InstallOptions) error {
	args := []string{"install"}
	
	if options != nil {
		if options.Clean {
			// Remove node_modules first
			nodeModulesPath := filepath.Join(projectPath, "node_modules")
			os.RemoveAll(nodeModulesPath)
		}
		if options.Production {
			args = append(args, "--production")
		}
		if options.Registry != nil && options.Registry.URL != "" {
			args = append(args, "--registry", options.Registry.URL)
		}
	}
	
	cmd := exec.CommandContext(ctx, "npm", args...)
	cmd.Dir = projectPath
	
	if options != nil && options.Environment != nil {
		cmd.Env = append(os.Environ())
		for key, value := range options.Environment {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("npm install failed: %w\nOutput: %s", err, string(output))
	}
	
	logger.Info("Successfully installed dependencies for project at %s", projectPath)
	return nil
}

// ValidateProject checks if the project is a valid npm project
func (n *NPMManager) ValidateProject(ctx context.Context, projectPath string) error {
	packageJsonPath := filepath.Join(projectPath, "package.json")
	
	if _, err := os.Stat(packageJsonPath); os.IsNotExist(err) {
		return fmt.Errorf("package.json not found in %s", projectPath)
	}
	
	// Try to parse the package.json
	data, err := os.ReadFile(packageJsonPath)
	if err != nil {
		return fmt.Errorf("failed to read package.json: %w", err)
	}
	
	var packageJson map[string]interface{}
	if err := json.Unmarshal(data, &packageJson); err != nil {
		return fmt.Errorf("invalid package.json format: %w", err)
	}
	
	// Check for required fields
	if _, exists := packageJson["name"]; !exists {
		return fmt.Errorf("package.json missing required 'name' field")
	}
	
	return nil
}
