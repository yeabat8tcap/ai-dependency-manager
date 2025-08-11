package pip

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/logger"
	"github.com/8tcapital/ai-dep-manager/internal/packagemanager/types"
)

// PipManager implements PackageManager interface for pip/Python
type PipManager struct {
	httpClient *http.Client
}

// NewPipManager creates a new pip package manager
func NewPipManager() *PipManager {
	return &PipManager{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName returns the display name of the package manager
func (p *PipManager) GetName() string {
	return "Python Package Installer"
}

// GetType returns the type identifier
func (p *PipManager) GetType() string {
	return "pip"
}

// GetVersion returns the version of the package manager
func (p *PipManager) GetVersion() string {
	return "1.0.0"
}

// GetVersions returns available versions for a package
func (p *PipManager) GetVersions(ctx context.Context, packageName string) ([]string, error) {
	// This is a placeholder implementation
	// In a real implementation, this would query PyPI
	return []string{}, nil
}

// IsAvailable checks if pip is available on the system
func (p *PipManager) IsAvailable(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "pip", "--version")
	err := cmd.Run()
	if err != nil {
		// Try pip3 as fallback
		cmd = exec.CommandContext(ctx, "pip3", "--version")
		err = cmd.Run()
	}
	return err == nil
}

// DetectProjects scans for Python projects in the given directory
func (p *PipManager) DetectProjects(ctx context.Context, rootPath string) ([]types.Project, error) {
	var projects []types.Project
	
	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		// Skip common Python cache directories
		if d.IsDir() && (d.Name() == "__pycache__" || d.Name() == ".venv" || d.Name() == "venv" || d.Name() == ".env") {
			return filepath.SkipDir
		}
		
		// Look for Python project files
		if !d.IsDir() {
			switch d.Name() {
			case "requirements.txt", "setup.py", "pyproject.toml", "Pipfile":
				projectInfo, err := p.parseProjectInfo(path, d.Name())
				if err != nil {
					logger.Warn("Failed to parse Python project file at %s: %v", path, err)
					return nil
				}
				
				projects = append(projects, *projectInfo)
			}
		}
		
		return nil
	})
	
	return projects, err
}

// parseProjectInfo parses Python project files to extract project information
func (p *PipManager) parseProjectInfo(filePath, fileName string) (*types.Project, error) {
	projectPath := filepath.Dir(filePath)
	projectName := filepath.Base(projectPath)
	
	// Try to extract project name from setup.py or pyproject.toml
	if fileName == "setup.py" {
		if name := p.extractNameFromSetupPy(filePath); name != "" {
			projectName = name
		}
	} else if fileName == "pyproject.toml" {
		if name := p.extractNameFromPyprojectToml(filePath); name != "" {
			projectName = name
		}
	}
	
	return &types.Project{
		Name:           projectName,
		Path:           projectPath,
		ConfigFile:     filePath,
		PackageManager: "pip",
		Metadata: map[string]string{
			"config_file": filePath,
			"file_type":   fileName,
		},
	}, nil
}

// extractNameFromSetupPy attempts to extract project name from setup.py
func (p *PipManager) extractNameFromSetupPy(filePath string) string {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}
	
	// Simple regex to find name parameter in setup()
	nameRegex := regexp.MustCompile(`name\s*=\s*['"](.*?)['"]`)
	matches := nameRegex.FindStringSubmatch(string(data))
	if len(matches) > 1 {
		return matches[1]
	}
	
	return ""
}

// extractNameFromPyprojectToml attempts to extract project name from pyproject.toml
func (p *PipManager) extractNameFromPyprojectToml(filePath string) string {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}
	
	// Simple regex to find name in [project] section
	nameRegex := regexp.MustCompile(`(?m)^\s*name\s*=\s*['"](.*?)['"]`)
	matches := nameRegex.FindStringSubmatch(string(data))
	if len(matches) > 1 {
		return matches[1]
	}
	
	return ""
}

// ParseDependencies reads and parses Python dependency files
func (p *PipManager) ParseDependencies(ctx context.Context, projectPath string) (*types.DependencyInfo, error) {
	depInfo := &types.DependencyInfo{
		ProjectName: filepath.Base(projectPath),
	}
	
	// Check for different dependency files in order of preference
	dependencyFiles := []string{
		"requirements.txt",
		"setup.py",
		"pyproject.toml",
		"Pipfile",
	}
	
	var foundFile string
	var dependencies []types.DependencyEntry
	
	for _, fileName := range dependencyFiles {
		filePath := filepath.Join(projectPath, fileName)
		if _, err := os.Stat(filePath); err == nil {
			foundFile = fileName
			var err error
			
			switch fileName {
			case "requirements.txt":
				dependencies, err = p.parseRequirementsTxt(filePath)
			case "setup.py":
				dependencies, err = p.parseSetupPy(filePath)
			case "pyproject.toml":
				dependencies, err = p.parsePyprojectToml(filePath)
			case "Pipfile":
				dependencies, err = p.parsePipfile(filePath)
			}
			
			if err != nil {
				logger.Warn("Failed to parse %s: %v", fileName, err)
				continue
			}
			
			break
		}
	}
	
	if foundFile == "" {
		return nil, fmt.Errorf("no Python dependency files found in %s", projectPath)
	}
	
	depInfo.Dependencies = dependencies
	depInfo.Metadata = map[string]string{
		"config_file": foundFile,
		"python_version": "unknown",
	}
	
	return depInfo, nil
}

// parseRequirementsTxt parses requirements.txt file
func (p *PipManager) parseRequirementsTxt(filePath string) ([]types.DependencyEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open requirements.txt: %w", err)
	}
	defer file.Close()
	
	var dependencies []types.DependencyEntry
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Skip -e (editable) and -r (recursive) flags for now
		if strings.HasPrefix(line, "-") {
			continue
		}
		
		// Parse package name and version constraint
		dep := p.parseRequirementLine(line)
		if dep != nil {
			dependencies = append(dependencies, *dep)
		}
	}
	
	return dependencies, scanner.Err()
}

// parseRequirementLine parses a single requirement line
func (p *PipManager) parseRequirementLine(line string) *types.DependencyEntry {
	// Handle various formats: package==1.0.0, package>=1.0.0, package~=1.0.0, etc.
	operatorRegex := regexp.MustCompile(`^([a-zA-Z0-9\-_\.]+)([><=~!]*)(.*)$`)
	matches := operatorRegex.FindStringSubmatch(line)
	
	if len(matches) < 2 {
		return nil
	}
	
	name := matches[1]
	version := ""
	if len(matches) > 2 {
		version = matches[2] + matches[3]
	}
	
	return &types.DependencyEntry{
		Name:    name,
		Version: version,
		Type:    "direct",
	}
}

// parseSetupPy parses setup.py file (simplified implementation)
func (p *PipManager) parseSetupPy(filePath string) ([]types.DependencyEntry, error) {
	// This is a simplified implementation
	// In practice, you'd want to execute setup.py or use AST parsing
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read setup.py: %w", err)
	}
	
	content := string(data)
	var dependencies []types.DependencyEntry
	
	// Look for install_requires
	installRequiresRegex := regexp.MustCompile(`install_requires\s*=\s*\[(.*?)\]`)
	matches := installRequiresRegex.FindStringSubmatch(content)
	
	if len(matches) > 1 {
		reqsStr := matches[1]
		// Extract individual requirements
		reqRegex := regexp.MustCompile(`['"]([^'"]+)['"]`)
		reqMatches := reqRegex.FindAllStringSubmatch(reqsStr, -1)
		
		for _, reqMatch := range reqMatches {
			if len(reqMatch) > 1 {
				dep := p.parseRequirementLine(reqMatch[1])
				if dep != nil {
					dependencies = append(dependencies, *dep)
				}
			}
		}
	}
	
	return dependencies, nil
}

// parsePyprojectToml parses pyproject.toml file (simplified implementation)
func (p *PipManager) parsePyprojectToml(filePath string) ([]types.DependencyEntry, error) {
	// This is a very simplified implementation
	// In practice, you'd want to use a proper TOML parser
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read pyproject.toml: %w", err)
	}
	
	content := string(data)
	var dependencies []types.DependencyEntry
	
	// Look for dependencies in [project] section
	dependenciesRegex := regexp.MustCompile(`(?s)dependencies\s*=\s*\[(.*?)\]`)
	matches := dependenciesRegex.FindStringSubmatch(content)
	
	if len(matches) > 1 {
		depsStr := matches[1]
		// Extract individual dependencies
		depRegex := regexp.MustCompile(`['"]([^'"]+)['"]`)
		depMatches := depRegex.FindAllStringSubmatch(depsStr, -1)
		
		for _, depMatch := range depMatches {
			if len(depMatch) > 1 {
				dep := p.parseRequirementLine(depMatch[1])
				if dep != nil {
					dependencies = append(dependencies, *dep)
				}
			}
		}
	}
	
	return dependencies, nil
}

// parsePipfile parses Pipfile (simplified implementation)
func (p *PipManager) parsePipfile(filePath string) ([]types.DependencyEntry, error) {
	// This is a simplified implementation
	// In practice, you'd want to use a proper TOML parser for Pipfile
	return []types.DependencyEntry{}, nil
}

// GetLatestVersion fetches the latest version from PyPI
func (p *PipManager) GetLatestVersion(ctx context.Context, packageName string, registry *types.RegistryConfig) (*types.VersionInfo, error) {
	registryURL := "https://pypi.org"
	if registry != nil && registry.URL != "" {
		registryURL = registry.URL
	}
	
	url := fmt.Sprintf("%s/pypi/%s/json", registryURL, packageName)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch package info: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("PyPI returned status %d", resp.StatusCode)
	}
	
	var packageInfo struct {
		Info struct {
			Version string `json:"version"`
		} `json:"info"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&packageInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &types.VersionInfo{
		Version:     packageInfo.Info.Version,
		PublishedAt: time.Now(), // PyPI API doesn't provide publish date in this endpoint
	}, nil
}

// GetVersionHistory fetches version history for a package
func (p *PipManager) GetVersionHistory(ctx context.Context, packageName string, registry *types.RegistryConfig) ([]types.VersionInfo, error) {
	registryURL := "https://pypi.org"
	if registry != nil && registry.URL != "" {
		registryURL = registry.URL
	}
	
	url := fmt.Sprintf("%s/pypi/%s/json", registryURL, packageName)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch package info: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("PyPI returned status %d", resp.StatusCode)
	}
	
	var packageInfo struct {
		Releases map[string][]struct {
			UploadTime string `json:"upload_time"`
		} `json:"releases"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&packageInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	var versions []types.VersionInfo
	for version, releases := range packageInfo.Releases {
		if len(releases) > 0 {
			publishedAt, _ := time.Parse("2006-01-02T15:04:05", releases[0].UploadTime)
			
			versionInfo := types.VersionInfo{
				Version:     version,
				PublishedAt: publishedAt,
			}
			
			versions = append(versions, versionInfo)
		}
	}
	
	return versions, nil
}

// GetChangelog fetches changelog/release notes for a version
func (p *PipManager) GetChangelog(ctx context.Context, packageName, version string, registry *types.RegistryConfig) (*types.ChangelogInfo, error) {
	// For PyPI, changelog information is limited
	// This is a placeholder implementation
	return &types.ChangelogInfo{
		Version:     version,
		Description: fmt.Sprintf("Release %s of %s", version, packageName),
		IsBreaking:  false,
		SecurityFix: false,
	}, nil
}

// UpdateDependency updates a specific dependency
func (p *PipManager) UpdateDependency(ctx context.Context, projectPath, packageName, newVersion string, options *types.UpdateOptions) error {
	pipCmd := "pip"
	
	// Try pip3 if pip is not available
	if !p.commandExists("pip") && p.commandExists("pip3") {
		pipCmd = "pip3"
	}
	
	args := []string{"install", fmt.Sprintf("%s==%s", packageName, newVersion)}
	
	if options != nil {
		if options.Force {
			args = append(args, "--force-reinstall")
		}
		if options.Registry != nil && options.Registry.URL != "" {
			args = append(args, "--index-url", options.Registry.URL)
		}
	}
	
	cmd := exec.CommandContext(ctx, pipCmd, args...)
	cmd.Dir = projectPath
	
	if options != nil && options.Environment != nil {
		cmd.Env = append(os.Environ())
		for key, value := range options.Environment {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pip install failed: %w\nOutput: %s", err, string(output))
	}
	
	logger.Info("Successfully updated %s to %s", packageName, newVersion)
	return nil
}

// InstallDependencies installs all dependencies
func (p *PipManager) InstallDependencies(ctx context.Context, projectPath string, options *types.InstallOptions) error {
	pipCmd := "pip"
	
	// Try pip3 if pip is not available
	if !p.commandExists("pip") && p.commandExists("pip3") {
		pipCmd = "pip3"
	}
	
	// Look for requirements file
	requirementsFile := filepath.Join(projectPath, "requirements.txt")
	if _, err := os.Stat(requirementsFile); os.IsNotExist(err) {
		return fmt.Errorf("requirements.txt not found in %s", projectPath)
	}
	
	args := []string{"install", "-r", "requirements.txt"}
	
	if options != nil {
		if options.Registry != nil && options.Registry.URL != "" {
			args = append(args, "--index-url", options.Registry.URL)
		}
	}
	
	cmd := exec.CommandContext(ctx, pipCmd, args...)
	cmd.Dir = projectPath
	
	if options != nil && options.Environment != nil {
		cmd.Env = append(os.Environ())
		for key, value := range options.Environment {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pip install failed: %w\nOutput: %s", err, string(output))
	}
	
	logger.Info("Successfully installed dependencies for project at %s", projectPath)
	return nil
}

// ValidateProject checks if the project is a valid Python project
func (p *PipManager) ValidateProject(ctx context.Context, projectPath string) error {
	// Check for at least one Python dependency file
	dependencyFiles := []string{
		"requirements.txt",
		"setup.py",
		"pyproject.toml",
		"Pipfile",
	}
	
	for _, fileName := range dependencyFiles {
		filePath := filepath.Join(projectPath, fileName)
		if _, err := os.Stat(filePath); err == nil {
			return nil // Found at least one dependency file
		}
	}
	
	return fmt.Errorf("no Python dependency files found in %s", projectPath)
}

// commandExists checks if a command exists in PATH
func (p *PipManager) commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
