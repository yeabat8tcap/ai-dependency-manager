package github

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/logger"
)

// ParsingService handles code parsing for different project types
type ParsingService struct {
	client       *Client
	repositories *RepositoriesService
}

// NewParsingService creates a new parsing service
func NewParsingService(client *Client) *ParsingService {
	return &ParsingService{
		client:       client,
		repositories: client.Repositories,
	}
}

// ProjectStructure represents the structure of a project
type ProjectStructure struct {
	Repository      string                 `json:"repository"`
	ProjectType     string                 `json:"project_type"`
	PackageManager  string                 `json:"package_manager"`
	ConfigFiles     []*ConfigFile          `json:"config_files"`
	Dependencies    []*ParsedDependency    `json:"dependencies"`
	DevDependencies []*ParsedDependency    `json:"dev_dependencies,omitempty"`
	Scripts         map[string]string      `json:"scripts,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	ParsedAt        time.Time              `json:"parsed_at"`
}

// ConfigFile represents a configuration file
type ConfigFile struct {
	Path        string                 `json:"path"`
	Type        string                 `json:"type"`        // "package", "lock", "config"
	Format      string                 `json:"format"`      // "json", "yaml", "toml", "txt"
	Content     string                 `json:"content"`
	Parsed      bool                   `json:"parsed"`
	ParseError  string                 `json:"parse_error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ParsedDependency represents a parsed dependency
type ParsedDependency struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	VersionRange    string            `json:"version_range,omitempty"`
	Type            string            `json:"type"`            // "runtime", "dev", "peer", "optional"
	Source          string            `json:"source"`          // "package.json", "requirements.txt", etc.
	Line            int               `json:"line,omitempty"`
	Constraints     []string          `json:"constraints,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

// ParseProject parses a project and extracts its structure
func (p *ParsingService) ParseProject(ctx context.Context, owner, repo string) (*ProjectStructure, error) {
	logger.Info("Parsing project structure for %s/%s", owner, repo)
	
	structure := &ProjectStructure{
		Repository:      fmt.Sprintf("%s/%s", owner, repo),
		ConfigFiles:     []*ConfigFile{},
		Dependencies:    []*ParsedDependency{},
		DevDependencies: []*ParsedDependency{},
		Scripts:         make(map[string]string),
		Metadata:        make(map[string]interface{}),
		ParsedAt:        time.Now(),
	}
	
	// Detect project type
	projectType, err := p.detectProjectType(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to detect project type: %w", err)
	}
	
	structure.ProjectType = projectType.Type
	structure.PackageManager = projectType.PackageManager
	
	// Parse based on project type
	switch projectType.Type {
	case "nodejs":
		if err := p.parseNodeJSProject(ctx, owner, repo, structure); err != nil {
			return nil, fmt.Errorf("failed to parse Node.js project: %w", err)
		}
	case "python":
		if err := p.parsePythonProject(ctx, owner, repo, structure); err != nil {
			return nil, fmt.Errorf("failed to parse Python project: %w", err)
		}
	case "java":
		if err := p.parseJavaProject(ctx, owner, repo, structure); err != nil {
			return nil, fmt.Errorf("failed to parse Java project: %w", err)
		}
	case "rust":
		if err := p.parseRustProject(ctx, owner, repo, structure); err != nil {
			return nil, fmt.Errorf("failed to parse Rust project: %w", err)
		}
	case "go":
		if err := p.parseGoProject(ctx, owner, repo, structure); err != nil {
			return nil, fmt.Errorf("failed to parse Go project: %w", err)
		}
	default:
		logger.Warn("Unsupported project type: %s", projectType.Type)
	}
	
	logger.Info("Parsed project %s/%s: %d dependencies, %d dev dependencies", 
		owner, repo, len(structure.Dependencies), len(structure.DevDependencies))
	
	return structure, nil
}

// parseNodeJSProject parses a Node.js project
func (p *ParsingService) parseNodeJSProject(ctx context.Context, owner, repo string, structure *ProjectStructure) error {
	// Parse package.json
	packageJSON, err := p.repositories.GetContents(ctx, owner, repo, "package.json", "")
	if err != nil {
		return fmt.Errorf("failed to get package.json: %w", err)
	}
	
	configFile := &ConfigFile{
		Path:    "package.json",
		Type:    "package",
		Format:  "json",
		Content: packageJSON.Content,
		Parsed:  false,
	}
	
	// Parse package.json content
	var packageData map[string]interface{}
	if err := json.Unmarshal([]byte(packageJSON.Content), &packageData); err != nil {
		configFile.ParseError = err.Error()
	} else {
		configFile.Parsed = true
		
		// Extract dependencies
		if deps, ok := packageData["dependencies"].(map[string]interface{}); ok {
			for name, version := range deps {
				if versionStr, ok := version.(string); ok {
					structure.Dependencies = append(structure.Dependencies, &ParsedDependency{
						Name:         name,
						Version:      versionStr,
						VersionRange: versionStr,
						Type:         "runtime",
						Source:       "package.json",
					})
				}
			}
		}
		
		// Extract dev dependencies
		if devDeps, ok := packageData["devDependencies"].(map[string]interface{}); ok {
			for name, version := range devDeps {
				if versionStr, ok := version.(string); ok {
					structure.DevDependencies = append(structure.DevDependencies, &ParsedDependency{
						Name:         name,
						Version:      versionStr,
						VersionRange: versionStr,
						Type:         "dev",
						Source:       "package.json",
					})
				}
			}
		}
		
		// Extract scripts
		if scripts, ok := packageData["scripts"].(map[string]interface{}); ok {
			for name, script := range scripts {
				if scriptStr, ok := script.(string); ok {
					structure.Scripts[name] = scriptStr
				}
			}
		}
		
		// Extract metadata
		if name, ok := packageData["name"].(string); ok {
			structure.Metadata["name"] = name
		}
		if version, ok := packageData["version"].(string); ok {
			structure.Metadata["version"] = version
		}
		if description, ok := packageData["description"].(string); ok {
			structure.Metadata["description"] = description
		}
	}
	
	structure.ConfigFiles = append(structure.ConfigFiles, configFile)
	
	// Parse package-lock.json if it exists
	if lockFile, err := p.repositories.GetContents(ctx, owner, repo, "package-lock.json", ""); err == nil {
		lockConfigFile := &ConfigFile{
			Path:    "package-lock.json",
			Type:    "lock",
			Format:  "json",
			Content: lockFile.Content,
			Parsed:  true,
		}
		structure.ConfigFiles = append(structure.ConfigFiles, lockConfigFile)
		
		// Parse lock file for exact versions
		var lockData map[string]interface{}
		if err := json.Unmarshal([]byte(lockFile.Content), &lockData); err == nil {
			if packages, ok := lockData["packages"].(map[string]interface{}); ok {
				p.updateDependencyVersionsFromLock(structure, packages)
			}
		}
	}
	
	return nil
}

// parsePythonProject parses a Python project
func (p *ParsingService) parsePythonProject(ctx context.Context, owner, repo string, structure *ProjectStructure) error {
	// Parse requirements.txt
	if reqFile, err := p.repositories.GetContents(ctx, owner, repo, "requirements.txt", ""); err == nil {
		configFile := &ConfigFile{
			Path:    "requirements.txt",
			Type:    "package",
			Format:  "txt",
			Content: reqFile.Content,
			Parsed:  true,
		}
		
		// Parse requirements
		lines := strings.Split(reqFile.Content, "\n")
		for i, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			
			dep := p.parsePythonRequirement(line, i+1)
			if dep != nil {
				dep.Source = "requirements.txt"
				structure.Dependencies = append(structure.Dependencies, dep)
			}
		}
		
		structure.ConfigFiles = append(structure.ConfigFiles, configFile)
	}
	
	// Parse setup.py if it exists
	if setupFile, err := p.repositories.GetContents(ctx, owner, repo, "setup.py", ""); err == nil {
		configFile := &ConfigFile{
			Path:    "setup.py",
			Type:    "config",
			Format:  "python",
			Content: setupFile.Content,
			Parsed:  false, // Complex parsing needed
		}
		structure.ConfigFiles = append(structure.ConfigFiles, configFile)
		
		// Extract dependencies from setup.py (basic regex parsing)
		deps := p.extractPythonSetupDependencies(setupFile.Content)
		for _, dep := range deps {
			dep.Source = "setup.py"
			structure.Dependencies = append(structure.Dependencies, dep)
		}
	}
	
	// Parse pyproject.toml if it exists
	if pyprojectFile, err := p.repositories.GetContents(ctx, owner, repo, "pyproject.toml", ""); err == nil {
		configFile := &ConfigFile{
			Path:    "pyproject.toml",
			Type:    "config",
			Format:  "toml",
			Content: pyprojectFile.Content,
			Parsed:  false, // TOML parsing needed
		}
		structure.ConfigFiles = append(structure.ConfigFiles, configFile)
	}
	
	return nil
}

// parseJavaProject parses a Java project
func (p *ParsingService) parseJavaProject(ctx context.Context, owner, repo string, structure *ProjectStructure) error {
	// Parse pom.xml (Maven)
	if pomFile, err := p.repositories.GetContents(ctx, owner, repo, "pom.xml", ""); err == nil {
		configFile := &ConfigFile{
			Path:    "pom.xml",
			Type:    "package",
			Format:  "xml",
			Content: pomFile.Content,
			Parsed:  false, // XML parsing needed
		}
		structure.ConfigFiles = append(structure.ConfigFiles, configFile)
		
		// Extract dependencies from pom.xml (basic regex parsing)
		deps := p.extractMavenDependencies(pomFile.Content)
		for _, dep := range deps {
			dep.Source = "pom.xml"
			structure.Dependencies = append(structure.Dependencies, dep)
		}
	}
	
	// Parse build.gradle (Gradle)
	if gradleFile, err := p.repositories.GetContents(ctx, owner, repo, "build.gradle", ""); err == nil {
		configFile := &ConfigFile{
			Path:    "build.gradle",
			Type:    "package",
			Format:  "gradle",
			Content: gradleFile.Content,
			Parsed:  false, // Gradle parsing needed
		}
		structure.ConfigFiles = append(structure.ConfigFiles, configFile)
		
		// Extract dependencies from build.gradle (basic regex parsing)
		deps := p.extractGradleDependencies(gradleFile.Content)
		for _, dep := range deps {
			dep.Source = "build.gradle"
			structure.Dependencies = append(structure.Dependencies, dep)
		}
	}
	
	return nil
}

// parseRustProject parses a Rust project
func (p *ParsingService) parseRustProject(ctx context.Context, owner, repo string, structure *ProjectStructure) error {
	// Parse Cargo.toml
	if cargoFile, err := p.repositories.GetContents(ctx, owner, repo, "Cargo.toml", ""); err == nil {
		configFile := &ConfigFile{
			Path:    "Cargo.toml",
			Type:    "package",
			Format:  "toml",
			Content: cargoFile.Content,
			Parsed:  false, // TOML parsing needed
		}
		structure.ConfigFiles = append(structure.ConfigFiles, configFile)
		
		// Extract dependencies from Cargo.toml (basic parsing)
		deps := p.extractCargoDependencies(cargoFile.Content)
		for _, dep := range deps {
			dep.Source = "Cargo.toml"
			structure.Dependencies = append(structure.Dependencies, dep)
		}
	}
	
	return nil
}

// parseGoProject parses a Go project
func (p *ParsingService) parseGoProject(ctx context.Context, owner, repo string, structure *ProjectStructure) error {
	// Parse go.mod
	if goModFile, err := p.repositories.GetContents(ctx, owner, repo, "go.mod", ""); err == nil {
		configFile := &ConfigFile{
			Path:    "go.mod",
			Type:    "package",
			Format:  "mod",
			Content: goModFile.Content,
			Parsed:  true,
		}
		structure.ConfigFiles = append(structure.ConfigFiles, configFile)
		
		// Extract dependencies from go.mod
		deps := p.extractGoModDependencies(goModFile.Content)
		for _, dep := range deps {
			dep.Source = "go.mod"
			structure.Dependencies = append(structure.Dependencies, dep)
		}
	}
	
	return nil
}

// parsePythonRequirement parses a Python requirement line
func (p *ParsingService) parsePythonRequirement(line string, lineNum int) *ParsedDependency {
	// Handle different requirement formats
	// package==1.0.0
	// package>=1.0.0
	// package~=1.0.0
	// package[extra]>=1.0.0
	
	re := regexp.MustCompile(`^([a-zA-Z0-9_-]+)(\[[^\]]+\])?(.*?)$`)
	matches := re.FindStringSubmatch(line)
	
	if len(matches) < 2 {
		return nil
	}
	
	name := matches[1]
	versionSpec := strings.TrimSpace(matches[3])
	
	return &ParsedDependency{
		Name:         name,
		Version:      versionSpec,
		VersionRange: versionSpec,
		Type:         "runtime",
		Line:         lineNum,
	}
}

// extractPythonSetupDependencies extracts dependencies from setup.py
func (p *ParsingService) extractPythonSetupDependencies(content string) []*ParsedDependency {
	var deps []*ParsedDependency
	
	// Look for install_requires
	re := regexp.MustCompile(`install_requires\s*=\s*\[(.*?)\]`)
	matches := re.FindStringSubmatch(content)
	
	if len(matches) > 1 {
		// Extract individual requirements
		reqRe := regexp.MustCompile(`['"]([^'"]+)['"]`)
		reqMatches := reqRe.FindAllStringSubmatch(matches[1], -1)
		
		for _, reqMatch := range reqMatches {
			if len(reqMatch) > 1 {
				dep := p.parsePythonRequirement(reqMatch[1], 0)
				if dep != nil {
					deps = append(deps, dep)
				}
			}
		}
	}
	
	return deps
}

// extractMavenDependencies extracts dependencies from pom.xml
func (p *ParsingService) extractMavenDependencies(content string) []*ParsedDependency {
	var deps []*ParsedDependency
	
	// Basic regex to extract Maven dependencies
	re := regexp.MustCompile(`<dependency>.*?<groupId>(.*?)</groupId>.*?<artifactId>(.*?)</artifactId>.*?<version>(.*?)</version>.*?</dependency>`)
	matches := re.FindAllStringSubmatch(content, -1)
	
	for _, match := range matches {
		if len(match) >= 4 {
			name := fmt.Sprintf("%s:%s", match[1], match[2])
			version := match[3]
			
			deps = append(deps, &ParsedDependency{
				Name:         name,
				Version:      version,
				VersionRange: version,
				Type:         "runtime",
				Metadata: map[string]string{
					"groupId":    match[1],
					"artifactId": match[2],
				},
			})
		}
	}
	
	return deps
}

// extractGradleDependencies extracts dependencies from build.gradle
func (p *ParsingService) extractGradleDependencies(content string) []*ParsedDependency {
	var deps []*ParsedDependency
	
	// Basic regex to extract Gradle dependencies
	re := regexp.MustCompile(`(?:implementation|compile|api)\s+['"]([^:]+):([^:]+):([^'"]+)['"]`)
	matches := re.FindAllStringSubmatch(content, -1)
	
	for _, match := range matches {
		if len(match) >= 4 {
			name := fmt.Sprintf("%s:%s", match[1], match[2])
			version := match[3]
			
			deps = append(deps, &ParsedDependency{
				Name:         name,
				Version:      version,
				VersionRange: version,
				Type:         "runtime",
				Metadata: map[string]string{
					"group":   match[1],
					"name":    match[2],
				},
			})
		}
	}
	
	return deps
}

// extractCargoDependencies extracts dependencies from Cargo.toml
func (p *ParsingService) extractCargoDependencies(content string) []*ParsedDependency {
	var deps []*ParsedDependency
	
	// Basic parsing for [dependencies] section
	lines := strings.Split(content, "\n")
	inDepsSection := false
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if line == "[dependencies]" {
			inDepsSection = true
			continue
		}
		
		if strings.HasPrefix(line, "[") && line != "[dependencies]" {
			inDepsSection = false
			continue
		}
		
		if inDepsSection && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				name := strings.TrimSpace(parts[0])
				version := strings.Trim(strings.TrimSpace(parts[1]), `"`)
				
				deps = append(deps, &ParsedDependency{
					Name:         name,
					Version:      version,
					VersionRange: version,
					Type:         "runtime",
				})
			}
		}
	}
	
	return deps
}

// extractGoModDependencies extracts dependencies from go.mod
func (p *ParsingService) extractGoModDependencies(content string) []*ParsedDependency {
	var deps []*ParsedDependency
	
	lines := strings.Split(content, "\n")
	inRequireSection := false
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if line == "require (" {
			inRequireSection = true
			continue
		}
		
		if inRequireSection && line == ")" {
			inRequireSection = false
			continue
		}
		
		if strings.HasPrefix(line, "require ") && !strings.Contains(line, "(") {
			// Single require line
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				name := parts[1]
				version := parts[2]
				
				deps = append(deps, &ParsedDependency{
					Name:         name,
					Version:      version,
					VersionRange: version,
					Type:         "runtime",
				})
			}
		} else if inRequireSection && line != "" {
			// Multi-line require
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				name := parts[0]
				version := parts[1]
				
				deps = append(deps, &ParsedDependency{
					Name:         name,
					Version:      version,
					VersionRange: version,
					Type:         "runtime",
				})
			}
		}
	}
	
	return deps
}

// updateDependencyVersionsFromLock updates dependency versions from lock file
func (p *ParsingService) updateDependencyVersionsFromLock(structure *ProjectStructure, packages map[string]interface{}) {
	for _, dep := range structure.Dependencies {
		if pkg, ok := packages["node_modules/"+dep.Name].(map[string]interface{}); ok {
			if version, ok := pkg["version"].(string); ok {
				dep.Version = version
			}
		}
	}
	
	for _, dep := range structure.DevDependencies {
		if pkg, ok := packages["node_modules/"+dep.Name].(map[string]interface{}); ok {
			if version, ok := pkg["version"].(string); ok {
				dep.Version = version
			}
		}
	}
}

// GetAffectedFiles identifies files that might be affected by dependency changes
func (p *ParsingService) GetAffectedFiles(ctx context.Context, owner, repo string, dependencies []*DependencyUpdate) ([]string, error) {
	var affectedFiles []string
	
	// Get repository contents
	contents, err := p.repositories.GetContents(ctx, owner, repo, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to get repository contents: %w", err)
	}
	
	// Search for import/require statements
	for _, dep := range dependencies {
		files, err := p.searchForDependencyUsage(ctx, owner, repo, dep.Name, contents)
		if err != nil {
			logger.Warn("Failed to search for dependency usage %s: %v", dep.Name, err)
			continue
		}
		affectedFiles = append(affectedFiles, files...)
	}
	
	// Remove duplicates
	seen := make(map[string]bool)
	var uniqueFiles []string
	for _, file := range affectedFiles {
		if !seen[file] {
			seen[file] = true
			uniqueFiles = append(uniqueFiles, file)
		}
	}
	
	return uniqueFiles, nil
}

// searchForDependencyUsage searches for usage of a dependency in the codebase
func (p *ParsingService) searchForDependencyUsage(ctx context.Context, owner, repo, depName string, contents *RepositoryContent) ([]string, error) {
	var files []string
	
	// This is a simplified implementation
	// In a real implementation, you would recursively search through the repository
	// and look for import/require statements
	
	// For now, just return common file patterns
	commonFiles := []string{
		"src/index.js",
		"src/main.js",
		"src/app.js",
		"index.js",
		"main.py",
		"app.py",
		"src/main/java/Main.java",
	}
	
	return commonFiles, nil
}
