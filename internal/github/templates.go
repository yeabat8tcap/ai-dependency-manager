package github

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// TemplateService handles template-based patch generation
type TemplateService struct {
	templates map[string]*PatchTemplate
}

// NewTemplateService creates a new template service
func NewTemplateService() *TemplateService {
	service := &TemplateService{
		templates: make(map[string]*PatchTemplate),
	}
	
	// Load built-in templates
	service.loadBuiltinTemplates()
	
	return service
}

// PatchTemplate represents a template for generating patches
type PatchTemplate struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	PackageName     string                 `json:"package_name"`
	VersionPattern  string                 `json:"version_pattern"`
	ProjectTypes    []string               `json:"project_types"`
	Languages       []string               `json:"languages"`
	Conditions      []*TemplateCondition   `json:"conditions"`
	Transformations []*TemplateTransformation `json:"transformations"`
	FilePatterns    []*FilePattern         `json:"file_patterns"`
	ConfigChanges   []*TemplateConfigChange `json:"config_changes"`
	ValidationSteps []*TemplateValidation  `json:"validation_steps"`
	Priority        int                    `json:"priority"`
	Confidence      float64                `json:"confidence"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// TemplateCondition represents a condition for template application
type TemplateCondition struct {
	Type      string      `json:"type"`      // "version_range", "file_exists", "content_match", "dependency_exists"
	Field     string      `json:"field"`
	Operator  string      `json:"operator"`  // "eq", "ne", "gt", "lt", "contains", "matches"
	Value     interface{} `json:"value"`
	Required  bool        `json:"required"`
}

// TemplateTransformation represents a code transformation
type TemplateTransformation struct {
	Type        string            `json:"type"`        // "replace", "insert", "delete", "rename"
	Target      string            `json:"target"`      // "import", "function_call", "class_method", "variable"
	Pattern     string            `json:"pattern"`     // Regex pattern to match
	Replacement string            `json:"replacement"` // Replacement text
	Description string            `json:"description"`
	Examples    []*TransformExample `json:"examples"`
}

// TransformExample represents an example of a transformation
type TransformExample struct {
	Before string `json:"before"`
	After  string `json:"after"`
}

// FilePattern represents a file pattern for template application
type FilePattern struct {
	Pattern     string   `json:"pattern"`     // Glob pattern for file matching
	Languages   []string `json:"languages"`
	Exclude     []string `json:"exclude"`
	Required    bool     `json:"required"`
	Description string   `json:"description"`
}

// TemplateConfigChange represents a configuration change template
type TemplateConfigChange struct {
	File        string                 `json:"file"`
	Type        string                 `json:"type"`
	Operations  []*ConfigOperation     `json:"operations"`
	Description string                 `json:"description"`
}

// ConfigOperation represents a configuration operation
type ConfigOperation struct {
	Type  string      `json:"type"`  // "set", "delete", "rename", "merge"
	Path  string      `json:"path"`  // JSON path or key
	Value interface{} `json:"value"`
}

// TemplateValidation represents validation steps for a template
type TemplateValidation struct {
	Type        string `json:"type"`
	Command     string `json:"command,omitempty"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// GetTemplate retrieves a template by ID
func (ts *TemplateService) GetTemplate(id string) *PatchTemplate {
	return ts.templates[id]
}

// FindTemplates finds templates matching the given criteria
func (ts *TemplateService) FindTemplates(packageName, projectType string, currentVersion, targetVersion string) []*PatchTemplate {
	var matches []*PatchTemplate
	
	for _, template := range ts.templates {
		if ts.matchesTemplate(template, packageName, projectType, currentVersion, targetVersion) {
			matches = append(matches, template)
		}
	}
	
	// Sort by priority (higher priority first)
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[i].Priority < matches[j].Priority {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}
	
	return matches
}

// matchesTemplate checks if a template matches the given criteria
func (ts *TemplateService) matchesTemplate(template *PatchTemplate, packageName, projectType, currentVersion, targetVersion string) bool {
	// Check package name
	if template.PackageName != "" && template.PackageName != packageName {
		return false
	}
	
	// Check project type
	if len(template.ProjectTypes) > 0 {
		found := false
		for _, pt := range template.ProjectTypes {
			if pt == projectType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	// Check version pattern
	if template.VersionPattern != "" {
		matched, err := regexp.MatchString(template.VersionPattern, targetVersion)
		if err != nil || !matched {
			return false
		}
	}
	
	return true
}

// ApplyTemplate applies a template to generate patches
func (ts *TemplateService) ApplyTemplate(template *PatchTemplate, dependency *DependencyUpdate, projectStructure *ProjectStructure) ([]*FilePatch, []*ConfigPatch, error) {
	var filePatches []*FilePatch
	var configPatches []*ConfigPatch
	
	// Apply file transformations
	for _, pattern := range template.FilePatterns {
		patches, err := ts.applyFileTransformations(template, pattern, dependency, projectStructure)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to apply file transformations: %w", err)
		}
		filePatches = append(filePatches, patches...)
	}
	
	// Apply configuration changes
	for _, configChange := range template.ConfigChanges {
		patch, err := ts.applyConfigChange(configChange, dependency)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to apply config change: %w", err)
		}
		configPatches = append(configPatches, patch)
	}
	
	return filePatches, configPatches, nil
}

// applyFileTransformations applies file transformations from a template
func (ts *TemplateService) applyFileTransformations(template *PatchTemplate, pattern *FilePattern, dependency *DependencyUpdate, projectStructure *ProjectStructure) ([]*FilePatch, error) {
	var patches []*FilePatch
	
	// Find matching files (simplified - in real implementation would use proper glob matching)
	matchingFiles := ts.findMatchingFiles(pattern, projectStructure)
	
	for _, filePath := range matchingFiles {
		patch := &FilePatch{
			Path:        filePath,
			Type:        "modify",
			Language:    detectLanguage(filePath),
			Changes:     []*Change{},
			Confidence:  template.Confidence,
			Description: fmt.Sprintf("Apply template %s for %s", template.Name, dependency.Name),
		}
		
		// Apply transformations
		for _, transform := range template.Transformations {
			change, err := ts.applyTransformation(transform, dependency, filePath)
			if err != nil {
				continue // Skip failed transformations
			}
			patch.Changes = append(patch.Changes, change)
		}
		
		if len(patch.Changes) > 0 {
			patches = append(patches, patch)
		}
	}
	
	return patches, nil
}

// applyTransformation applies a single transformation
func (ts *TemplateService) applyTransformation(transform *TemplateTransformation, dependency *DependencyUpdate, filePath string) (*Change, error) {
	// Replace placeholders in pattern and replacement
	pattern := ts.replacePlaceholders(transform.Pattern, dependency)
	replacement := ts.replacePlaceholders(transform.Replacement, dependency)
	
	change := &Change{
		OldContent:  pattern,
		NewContent:  replacement,
		Type:        transform.Type,
		Reason:      transform.Description,
		Confidence:  0.8,
	}
	
	return change, nil
}

// applyConfigChange applies a configuration change
func (ts *TemplateService) applyConfigChange(configChange *TemplateConfigChange, dependency *DependencyUpdate) (*ConfigPatch, error) {
	patch := &ConfigPatch{
		File:        configChange.File,
		Type:        configChange.File,
		Changes:     make(map[string]interface{}),
		Description: configChange.Description,
	}
	
	for _, op := range configChange.Operations {
		switch op.Type {
		case "set":
			value := ts.replacePlaceholderValue(op.Value, dependency)
			patch.Changes[op.Path] = value
		case "delete":
			patch.Changes[op.Path] = nil
		}
	}
	
	return patch, nil
}

// findMatchingFiles finds files matching a pattern
func (ts *TemplateService) findMatchingFiles(pattern *FilePattern, projectStructure *ProjectStructure) []string {
	var files []string
	
	// Simplified file matching - in real implementation would use proper glob
	commonFiles := []string{
		"src/index.js", "src/app.js", "src/main.js",
		"src/components/App.js", "src/components/App.tsx",
		"app.py", "main.py", "src/main.py",
		"src/main/java/Main.java", "src/App.java",
		"main.rs", "src/main.rs",
		"main.go", "cmd/main.go",
	}
	
	for _, file := range commonFiles {
		if ts.matchesPattern(file, pattern.Pattern) {
			files = append(files, file)
		}
	}
	
	return files
}

// matchesPattern checks if a file matches a pattern
func (ts *TemplateService) matchesPattern(file, pattern string) bool {
	// Simplified pattern matching
	if pattern == "*" {
		return true
	}
	if strings.Contains(pattern, "*") {
		// Basic glob matching
		prefix := strings.Split(pattern, "*")[0]
		return strings.HasPrefix(file, prefix)
	}
	return file == pattern
}

// replacePlaceholders replaces placeholders in a string
func (ts *TemplateService) replacePlaceholders(text string, dependency *DependencyUpdate) string {
	replacements := map[string]string{
		"{{package_name}}":     dependency.Name,
		"{{current_version}}":  dependency.CurrentVersion,
		"{{target_version}}":   dependency.LatestVersion,
		"{{package_var}}":      strings.ReplaceAll(dependency.Name, "-", "_"),
		"{{package_camel}}":    toCamelCase(dependency.Name),
		"{{package_pascal}}":   toPascalCase(dependency.Name),
	}
	
	result := text
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}
	
	return result
}

// replacePlaceholderValue replaces placeholders in a value
func (ts *TemplateService) replacePlaceholderValue(value interface{}, dependency *DependencyUpdate) interface{} {
	if str, ok := value.(string); ok {
		return ts.replacePlaceholders(str, dependency)
	}
	return value
}

// loadBuiltinTemplates loads built-in templates
func (ts *TemplateService) loadBuiltinTemplates() {
	// Express.js templates
	ts.addExpressTemplates()
	
	// React templates
	ts.addReactTemplates()
	
	// Lodash templates
	ts.addLodashTemplates()
	
	// Python templates
	ts.addPythonTemplates()
	
	// Java templates
	ts.addJavaTemplates()
}

// addExpressTemplates adds Express.js specific templates
func (ts *TemplateService) addExpressTemplates() {
	// Express 4 to 5 migration
	ts.templates["express-4-to-5"] = &PatchTemplate{
		ID:             "express-4-to-5",
		Name:           "Express 4 to 5 Migration",
		Description:    "Migrates Express.js applications from version 4 to 5",
		PackageName:    "express",
		VersionPattern: "^5\\.",
		ProjectTypes:   []string{"nodejs"},
		Languages:      []string{"javascript", "typescript"},
		Transformations: []*TemplateTransformation{
			{
				Type:        "replace",
				Target:      "import",
				Pattern:     "const bodyParser = require\\('body-parser'\\)",
				Replacement: "// body-parser is now built into Express 5",
				Description: "Remove body-parser import as it's built into Express 5",
				Examples: []*TransformExample{
					{
						Before: "const bodyParser = require('body-parser');",
						After:  "// body-parser is now built into Express 5",
					},
				},
			},
			{
				Type:        "replace",
				Target:      "function_call",
				Pattern:     "app\\.use\\(bodyParser\\.json\\(\\)\\)",
				Replacement: "app.use(express.json())",
				Description: "Replace bodyParser.json() with express.json()",
				Examples: []*TransformExample{
					{
						Before: "app.use(bodyParser.json());",
						After:  "app.use(express.json());",
					},
				},
			},
			{
				Type:        "replace",
				Target:      "function_call",
				Pattern:     "app\\.use\\(bodyParser\\.urlencoded\\(([^)]*)\\)\\)",
				Replacement: "app.use(express.urlencoded($1))",
				Description: "Replace bodyParser.urlencoded() with express.urlencoded()",
				Examples: []*TransformExample{
					{
						Before: "app.use(bodyParser.urlencoded({ extended: true }));",
						After:  "app.use(express.urlencoded({ extended: true }));",
					},
				},
			},
		},
		FilePatterns: []*FilePattern{
			{
				Pattern:     "*.js",
				Languages:   []string{"javascript"},
				Description: "JavaScript files",
			},
			{
				Pattern:     "*.ts",
				Languages:   []string{"typescript"},
				Description: "TypeScript files",
			},
		},
		ValidationSteps: []*TemplateValidation{
			{
				Type:        "build",
				Command:     "npm run build",
				Description: "Build the application",
				Required:    true,
			},
			{
				Type:        "test",
				Command:     "npm test",
				Description: "Run tests",
				Required:    true,
			},
		},
		Priority:   10,
		Confidence: 0.9,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// addReactTemplates adds React specific templates
func (ts *TemplateService) addReactTemplates() {
	// React 16 to 17 migration
	ts.templates["react-16-to-17"] = &PatchTemplate{
		ID:             "react-16-to-17",
		Name:           "React 16 to 17 Migration",
		Description:    "Migrates React applications from version 16 to 17",
		PackageName:    "react",
		VersionPattern: "^17\\.",
		ProjectTypes:   []string{"nodejs"},
		Languages:      []string{"javascript", "typescript"},
		Transformations: []*TemplateTransformation{
			{
				Type:        "replace",
				Target:      "function_call",
				Pattern:     "componentWillMount\\(\\)",
				Replacement: "componentDidMount()",
				Description: "Replace deprecated componentWillMount with componentDidMount",
				Examples: []*TransformExample{
					{
						Before: "componentWillMount() { /* code */ }",
						After:  "componentDidMount() { /* code */ }",
					},
				},
			},
			{
				Type:        "replace",
				Target:      "function_call",
				Pattern:     "componentWillReceiveProps\\(([^)]*)\\)",
				Replacement: "componentDidUpdate(prevProps) { if (prevProps !== this.props) { /* migration needed */ } }",
				Description: "Replace deprecated componentWillReceiveProps",
				Examples: []*TransformExample{
					{
						Before: "componentWillReceiveProps(nextProps) { /* code */ }",
						After:  "componentDidUpdate(prevProps) { if (prevProps !== this.props) { /* migration needed */ } }",
					},
				},
			},
		},
		FilePatterns: []*FilePattern{
			{
				Pattern:     "src/components/*.js",
				Languages:   []string{"javascript"},
				Description: "React component files",
			},
			{
				Pattern:     "src/components/*.tsx",
				Languages:   []string{"typescript"},
				Description: "React TypeScript component files",
			},
		},
		Priority:   9,
		Confidence: 0.8,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// addLodashTemplates adds Lodash specific templates
func (ts *TemplateService) addLodashTemplates() {
	// Lodash 4 to 5 migration
	ts.templates["lodash-4-to-5"] = &PatchTemplate{
		ID:             "lodash-4-to-5",
		Name:           "Lodash 4 to 5 Migration",
		Description:    "Migrates Lodash usage from version 4 to 5",
		PackageName:    "lodash",
		VersionPattern: "^5\\.",
		ProjectTypes:   []string{"nodejs"},
		Languages:      []string{"javascript", "typescript"},
		Transformations: []*TemplateTransformation{
			{
				Type:        "replace",
				Target:      "function_call",
				Pattern:     "_\\.forEach\\(([^,]+),\\s*([^)]+)\\)",
				Replacement: "$1.forEach($2)",
				Description: "Replace _.forEach with native Array.forEach",
				Examples: []*TransformExample{
					{
						Before: "_.forEach(array, callback)",
						After:  "array.forEach(callback)",
					},
				},
			},
			{
				Type:        "replace",
				Target:      "function_call",
				Pattern:     "_\\.map\\(([^,]+),\\s*([^)]+)\\)",
				Replacement: "$1.map($2)",
				Description: "Replace _.map with native Array.map",
				Examples: []*TransformExample{
					{
						Before: "_.map(array, callback)",
						After:  "array.map(callback)",
					},
				},
			},
		},
		FilePatterns: []*FilePattern{
			{
				Pattern:     "*.js",
				Languages:   []string{"javascript"},
				Description: "JavaScript files",
			},
		},
		Priority:   7,
		Confidence: 0.7,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// addPythonTemplates adds Python specific templates
func (ts *TemplateService) addPythonTemplates() {
	// Django 3 to 4 migration
	ts.templates["django-3-to-4"] = &PatchTemplate{
		ID:             "django-3-to-4",
		Name:           "Django 3 to 4 Migration",
		Description:    "Migrates Django applications from version 3 to 4",
		PackageName:    "django",
		VersionPattern: "^4\\.",
		ProjectTypes:   []string{"python"},
		Languages:      []string{"python"},
		Transformations: []*TemplateTransformation{
			{
				Type:        "replace",
				Target:      "import",
				Pattern:     "from django\\.conf\\.urls import url",
				Replacement: "from django.urls import re_path as url",
				Description: "Replace deprecated url import with re_path",
				Examples: []*TransformExample{
					{
						Before: "from django.conf.urls import url",
						After:  "from django.urls import re_path as url",
					},
				},
			},
		},
		FilePatterns: []*FilePattern{
			{
				Pattern:     "*.py",
				Languages:   []string{"python"},
				Description: "Python files",
			},
		},
		Priority:   8,
		Confidence: 0.8,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// addJavaTemplates adds Java specific templates
func (ts *TemplateService) addJavaTemplates() {
	// Spring Boot 2 to 3 migration
	ts.templates["spring-boot-2-to-3"] = &PatchTemplate{
		ID:             "spring-boot-2-to-3",
		Name:           "Spring Boot 2 to 3 Migration",
		Description:    "Migrates Spring Boot applications from version 2 to 3",
		PackageName:    "org.springframework.boot:spring-boot-starter",
		VersionPattern: "^3\\.",
		ProjectTypes:   []string{"java"},
		Languages:      []string{"java"},
		Transformations: []*TemplateTransformation{
			{
				Type:        "replace",
				Target:      "import",
				Pattern:     "import javax\\.servlet\\.",
				Replacement: "import jakarta.servlet.",
				Description: "Replace javax.servlet with jakarta.servlet",
				Examples: []*TransformExample{
					{
						Before: "import javax.servlet.http.HttpServletRequest;",
						After:  "import jakarta.servlet.http.HttpServletRequest;",
					},
				},
			},
		},
		FilePatterns: []*FilePattern{
			{
				Pattern:     "*.java",
				Languages:   []string{"java"},
				Description: "Java files",
			},
		},
		Priority:   8,
		Confidence: 0.8,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// Helper functions
func toCamelCase(s string) string {
	parts := strings.Split(s, "-")
	if len(parts) == 0 {
		return s
	}
	
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result += strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return result
}

func toPascalCase(s string) string {
	camel := toCamelCase(s)
	if len(camel) > 0 {
		return strings.ToUpper(camel[:1]) + camel[1:]
	}
	return camel
}
