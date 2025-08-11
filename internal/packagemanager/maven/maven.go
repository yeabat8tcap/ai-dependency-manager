package maven

import (
	"context"
	"encoding/xml"
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

// MavenManager implements PackageManager interface for Maven
type MavenManager struct {
	httpClient *http.Client
}

// NewMavenManager creates a new Maven package manager
func NewMavenManager() *MavenManager {
	return &MavenManager{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName returns the display name of the package manager
func (m *MavenManager) GetName() string {
	return "Apache Maven"
}

// GetType returns the type identifier
func (m *MavenManager) GetType() string {
	return "maven"
}

// GetVersion returns the version of the package manager
func (m *MavenManager) GetVersion() string {
	return "1.0.0"
}

// GetVersions returns available versions for a package
func (m *MavenManager) GetVersions(ctx context.Context, packageName string) ([]string, error) {
	// This is a placeholder implementation
	// In a real implementation, this would query Maven Central
	return []string{}, nil
}

// IsAvailable checks if Maven is available on the system
func (m *MavenManager) IsAvailable(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "mvn", "--version")
	err := cmd.Run()
	return err == nil
}

// DetectProjects scans for Maven projects in the given directory
func (m *MavenManager) DetectProjects(ctx context.Context, rootPath string) ([]types.Project, error) {
	var projects []types.Project
	
	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		// Skip target directories
		if d.IsDir() && d.Name() == "target" {
			return filepath.SkipDir
		}
		
		// Look for pom.xml files
		if !d.IsDir() && d.Name() == "pom.xml" {
			projectInfo, err := m.parseProjectInfo(path)
			if err != nil {
				logger.Warn("Failed to parse pom.xml at %s: %v", path, err)
				return nil
			}
			
			projects = append(projects, *projectInfo)
		}
		
		return nil
	})
	
	return projects, err
}

// POM represents a simplified Maven POM structure
type POM struct {
	XMLName    xml.Name `xml:"project"`
	GroupID    string   `xml:"groupId"`
	ArtifactID string   `xml:"artifactId"`
	Version    string   `xml:"version"`
	Parent     *Parent  `xml:"parent"`
	Dependencies struct {
		Dependency []Dependency `xml:"dependency"`
	} `xml:"dependencies"`
}

// Parent represents parent POM information
type Parent struct {
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
	Version    string `xml:"version"`
}

// Dependency represents a Maven dependency
type Dependency struct {
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
	Version    string `xml:"version"`
	Scope      string `xml:"scope"`
	Type       string `xml:"type"`
	Optional   string `xml:"optional"`
}

// parseProjectInfo parses a pom.xml file to extract project information
func (m *MavenManager) parseProjectInfo(pomPath string) (*types.Project, error) {
	data, err := os.ReadFile(pomPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read pom.xml: %w", err)
	}
	
	var pom POM
	if err := xml.Unmarshal(data, &pom); err != nil {
		return nil, fmt.Errorf("failed to parse pom.xml: %w", err)
	}
	
	projectPath := filepath.Dir(pomPath)
	projectName := pom.ArtifactID
	if projectName == "" {
		projectName = filepath.Base(projectPath)
	}
	
	return &types.Project{
		Name:           projectName,
		Path:           projectPath,
		Type:           "maven",
		PackageManager: "maven",
		ConfigFile:     pomPath,
		Language:       "java",
		Metadata: map[string]string{
			"group_id":    pom.GroupID,
			"artifact_id": pom.ArtifactID,
			"config_file": pomPath,
			"version":     pom.Version,
		},
		DetectedAt: time.Now(),
	}, nil
}

// ParseDependencies reads and parses pom.xml
func (m *MavenManager) ParseDependencies(ctx context.Context, projectPath string) (*types.DependencyInfo, error) {
	pomPath := filepath.Join(projectPath, "pom.xml")
	
	data, err := os.ReadFile(pomPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read pom.xml: %w", err)
	}
	
	var pom POM
	if err := xml.Unmarshal(data, &pom); err != nil {
		return nil, fmt.Errorf("failed to parse pom.xml: %w", err)
	}
	
	depInfo := &types.DependencyInfo{
		ProjectName:    pom.ArtifactID,
		ProjectVersion: pom.Version,
		Dependencies:   m.convertDependencies(pom.Dependencies.Dependency),
		ConfigFile:     pomPath,
		Metadata: map[string]string{
			"groupId":    pom.GroupID,
			"artifactId": pom.ArtifactID,
		},
		ParsedAt: time.Now(),
	}
	
	return depInfo, nil
}

// convertDependencies converts Maven dependencies to DependencyEntry slice
func (m *MavenManager) convertDependencies(deps []Dependency) []types.DependencyEntry {
	var entries []types.DependencyEntry
	
	for _, dep := range deps {
		depType := "direct"
		if dep.Scope == "test" {
			depType = "dev"
		} else if dep.Optional == "true" {
			depType = "optional"
		}
		
		entry := types.DependencyEntry{
			Name:    fmt.Sprintf("%s:%s", dep.GroupID, dep.ArtifactID),
			Version: dep.Version,
			Type:    depType,
			Metadata: map[string]string{
				"group_id":    dep.GroupID,
				"artifact_id": dep.ArtifactID,
				"scope":       dep.Scope,
				"type":        dep.Type,
			},
		}
		entries = append(entries, entry)
	}
	
	return entries
}

// GetLatestVersion fetches the latest version from Maven Central
func (m *MavenManager) GetLatestVersion(ctx context.Context, packageName string, registry *types.RegistryConfig) (*types.VersionInfo, error) {
	parts := strings.Split(packageName, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid Maven package name format, expected groupId:artifactId")
	}
	
	groupID := parts[0]
	artifactID := parts[1]
	
	registryURL := "https://repo1.maven.org/maven2"
	if registry != nil && registry.URL != "" {
		registryURL = registry.URL
	}
	
	// Maven Central metadata URL
	url := fmt.Sprintf("%s/%s/%s/maven-metadata.xml", 
		registryURL, 
		strings.ReplaceAll(groupID, ".", "/"), 
		artifactID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metadata: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Maven repository returned status %d", resp.StatusCode)
	}
	
	var metadata struct {
		XMLName    xml.Name `xml:"metadata"`
		Versioning struct {
			Latest  string `xml:"latest"`
			Release string `xml:"release"`
		} `xml:"versioning"`
	}
	
	if err := xml.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %w", err)
	}
	
	version := metadata.Versioning.Release
	if version == "" {
		version = metadata.Versioning.Latest
	}
	
	return &types.VersionInfo{
		Version:     version,
		PublishedAt: time.Now(), // Maven metadata doesn't include publish date
	}, nil
}

// GetVersionHistory fetches version history for a package
func (m *MavenManager) GetVersionHistory(ctx context.Context, packageName string, registry *types.RegistryConfig) ([]types.VersionInfo, error) {
	parts := strings.Split(packageName, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid Maven package name format, expected groupId:artifactId")
	}
	
	groupID := parts[0]
	artifactID := parts[1]
	
	registryURL := "https://repo1.maven.org/maven2"
	if registry != nil && registry.URL != "" {
		registryURL = registry.URL
	}
	
	// Maven Central metadata URL
	url := fmt.Sprintf("%s/%s/%s/maven-metadata.xml", 
		registryURL, 
		strings.ReplaceAll(groupID, ".", "/"), 
		artifactID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metadata: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Maven repository returned status %d", resp.StatusCode)
	}
	
	var metadata struct {
		XMLName    xml.Name `xml:"metadata"`
		Versioning struct {
			Versions struct {
				Version []string `xml:"version"`
			} `xml:"versions"`
		} `xml:"versioning"`
	}
	
	if err := xml.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %w", err)
	}
	
	var versions []types.VersionInfo
	for _, version := range metadata.Versioning.Versions.Version {
		versionInfo := types.VersionInfo{
			Version:     version,
			PublishedAt: time.Now(), // Maven metadata doesn't include publish dates
		}
		versions = append(versions, versionInfo)
	}
	
	return versions, nil
}

// GetChangelog fetches changelog/release notes for a version
func (m *MavenManager) GetChangelog(ctx context.Context, packageName, version string, registry *types.RegistryConfig) (*types.ChangelogInfo, error) {
	// Maven doesn't have built-in changelog support
	// This would typically require checking project websites, GitHub releases, etc.
	return &types.ChangelogInfo{
		Version:     version,
		Description: fmt.Sprintf("Release %s of %s", version, packageName),
		IsBreaking:  false,
		SecurityFix: false,
	}, nil
}

// UpdateDependency updates a specific dependency
func (m *MavenManager) UpdateDependency(ctx context.Context, projectPath, packageName, newVersion string, options *types.UpdateOptions) error {
	parts := strings.Split(packageName, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid Maven package name format, expected groupId:artifactId")
	}
	
	groupID := parts[0]
	artifactID := parts[1]
	
	// Use Maven versions plugin to update dependency
	args := []string{
		"versions:use-dep-version",
		fmt.Sprintf("-Dincludes=%s:%s", groupID, artifactID),
		fmt.Sprintf("-DdepVersion=%s", newVersion),
	}
	
	if options != nil && options.DryRun {
		args = append(args, "-DdryRun=true")
	}
	
	cmd := exec.CommandContext(ctx, "mvn", args...)
	cmd.Dir = projectPath
	
	if options != nil && options.Environment != nil {
		cmd.Env = append(os.Environ())
		for key, value := range options.Environment {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Maven update failed: %w\nOutput: %s", err, string(output))
	}
	
	logger.Info("Successfully updated %s to %s", packageName, newVersion)
	return nil
}

// InstallDependencies installs all dependencies
func (m *MavenManager) InstallDependencies(ctx context.Context, projectPath string, options *types.InstallOptions) error {
	args := []string{"install"}
	
	if options != nil {
		if options.Clean {
			args = []string{"clean", "install"}
		}
	}
	
	cmd := exec.CommandContext(ctx, "mvn", args...)
	cmd.Dir = projectPath
	
	if options != nil && options.Environment != nil {
		cmd.Env = append(os.Environ())
		for key, value := range options.Environment {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Maven install failed: %w\nOutput: %s", err, string(output))
	}
	
	logger.Info("Successfully installed dependencies for project at %s", projectPath)
	return nil
}

// ValidateProject checks if the project is a valid Maven project
func (m *MavenManager) ValidateProject(ctx context.Context, projectPath string) error {
	pomPath := filepath.Join(projectPath, "pom.xml")
	
	if _, err := os.Stat(pomPath); os.IsNotExist(err) {
		return fmt.Errorf("pom.xml not found in %s", projectPath)
	}
	
	// Try to parse the pom.xml
	data, err := os.ReadFile(pomPath)
	if err != nil {
		return fmt.Errorf("failed to read pom.xml: %w", err)
	}
	
	var pom POM
	if err := xml.Unmarshal(data, &pom); err != nil {
		return fmt.Errorf("invalid pom.xml format: %w", err)
	}
	
	// Check for required fields
	if pom.ArtifactID == "" {
		return fmt.Errorf("pom.xml missing required 'artifactId' field")
	}
	
	return nil
}
