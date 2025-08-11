package testing

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/config"
	"github.com/8tcapital/ai-dep-manager/internal/database"
	"github.com/8tcapital/ai-dep-manager/internal/logger"
	"github.com/8tcapital/ai-dep-manager/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestContext provides a complete testing environment
type TestContext struct {
	DB        *gorm.DB
	Config    *config.Config
	TempDir   string
	Projects  []models.Project
	CleanupFn func()
}

// SetupTestEnvironment creates a complete test environment
func SetupTestEnvironment(t *testing.T) *TestContext {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "ai-dep-manager-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create test config
	cfg := &config.Config{
		DataDir:   tempDir,
		LogLevel:  "error", // Reduce log noise in tests
		LogFormat: "text",
		Database: config.DatabaseConfig{
			Type: "sqlite",
			Path: filepath.Join(tempDir, "test.db"),
		},
		Security: config.SecurityConfig{
			VerifyChecksums:        true,
			WhitelistEnabled:       false,
			VulnerabilityScanning:  true,
			MasterKey:              "dGVzdC1tYXN0ZXIta2V5LWZvci10ZXN0aW5nLW9ubHk=", // base64 encoded test key
			UpdateVulnDB:           false,
			VulnDBUpdateInterval:   "24h",
		},
		Agent: config.AgentConfig{
			Enabled:          true,
			ScanInterval:     "1h",
			MaxConcurrency:   2,
			AutoUpdateLevel:  "none",
			NotificationMode: "console",
		},
	}

	// Initialize logger
	logger.Init(cfg.LogLevel, cfg.LogFormat)

	// Create test database
	db, err := gorm.Open(sqlite.Open(cfg.Database.Path), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Run migrations
	if err := db.AutoMigrate(
		&models.Project{},
		&models.Dependency{},
		&models.ScanResult{},
		&models.Update{},
		&models.AIPrediction{},
		&models.AuditLog{},
		&models.RollbackPlan{},
		&models.RollbackItem{},
		&models.SecurityCheck{},
		&models.SecurityRule{},
		&models.Credential{},
		&models.VulnerabilityEntry{},
	); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Set global database for other components
	database.SetTestDB(db)

	ctx := &TestContext{
		DB:      db,
		Config:  cfg,
		TempDir: tempDir,
		CleanupFn: func() {
			os.RemoveAll(tempDir)
		},
	}

	// Create test data
	ctx.createTestData(t)

	return ctx
}

// Cleanup cleans up the test environment
func (ctx *TestContext) Cleanup() {
	if ctx.CleanupFn != nil {
		ctx.CleanupFn()
	}
}

// createTestData creates sample test data
func (ctx *TestContext) createTestData(t *testing.T) {
	// Create test projects
	projects := []models.Project{
		{
			Name:       "test-node-project",
			Path:       filepath.Join(ctx.TempDir, "node-project"),
			Type:       "npm",
			ConfigFile: "package.json",
			Enabled:    true,
		},
		{
			Name:       "test-python-project",
			Path:       filepath.Join(ctx.TempDir, "python-project"),
			Type:       "pip",
			ConfigFile: "requirements.txt",
			Enabled:    true,
		},
		{
			Name:       "test-java-project",
			Path:       filepath.Join(ctx.TempDir, "java-project"),
			Type:       "maven",
			ConfigFile: "pom.xml",
			Enabled:    true,
		},
	}

	for i := range projects {
		if err := ctx.DB.Create(&projects[i]).Error; err != nil {
			t.Fatalf("Failed to create test project: %v", err)
		}
		
		// Create project directory
		if err := os.MkdirAll(projects[i].Path, 0755); err != nil {
			t.Fatalf("Failed to create project directory: %v", err)
		}
	}

	ctx.Projects = projects

	// Create test dependencies
	dependencies := []models.Dependency{
		{
			ProjectID:      projects[0].ID,
			Name:           "express",
			CurrentVersion: "4.18.0",
			LatestVersion:  "4.18.2",
			Type:           "direct",
			Status:         "outdated",
		},
		{
			ProjectID:      projects[0].ID,
			Name:           "lodash",
			CurrentVersion: "4.17.20",
			LatestVersion:  "4.17.21",
			Type:           "direct",
			Status:         "outdated",
		},
		{
			ProjectID:      projects[1].ID,
			Name:           "requests",
			CurrentVersion: "2.28.0",
			LatestVersion:  "2.31.0",
			Type:           "direct",
			Status:         "outdated",
		},
		{
			ProjectID:      projects[1].ID,
			Name:           "numpy",
			CurrentVersion: "1.21.0",
			LatestVersion:  "1.24.3",
			Type:           "direct",
			Status:         "outdated",
		},
		{
			ProjectID:      projects[2].ID,
			Name:           "junit",
			CurrentVersion: "4.13.1",
			LatestVersion:  "4.13.2",
			Type:           "direct",
			Status:         "outdated",
		},
	}

	for i := range dependencies {
		if err := ctx.DB.Create(&dependencies[i]).Error; err != nil {
			t.Fatalf("Failed to create test dependency: %v", err)
		}
	}

	// Create test package manager files
	ctx.createPackageManagerFiles(t)
}

// createPackageManagerFiles creates realistic package manager files
func (ctx *TestContext) createPackageManagerFiles(t *testing.T) {
	// Create package.json for Node.js project
	packageJSON := `{
  "name": "test-node-project",
  "version": "1.0.0",
  "description": "Test Node.js project",
  "dependencies": {
    "express": "^4.18.0",
    "lodash": "^4.17.20"
  },
  "devDependencies": {
    "jest": "^29.0.0"
  }
}`
	packageJSONPath := filepath.Join(ctx.Projects[0].Path, "package.json")
	if err := os.WriteFile(packageJSONPath, []byte(packageJSON), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	// Create requirements.txt for Python project
	requirementsTxt := `requests==2.28.0
numpy==1.21.0
pytest==7.1.0
`
	requirementsPath := filepath.Join(ctx.Projects[1].Path, "requirements.txt")
	if err := os.WriteFile(requirementsPath, []byte(requirementsTxt), 0644); err != nil {
		t.Fatalf("Failed to create requirements.txt: %v", err)
	}

	// Create pom.xml for Java project
	pomXML := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 
         http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    
    <groupId>com.example</groupId>
    <artifactId>test-java-project</artifactId>
    <version>1.0.0</version>
    
    <dependencies>
        <dependency>
            <groupId>junit</groupId>
            <artifactId>junit</artifactId>
            <version>4.13.1</version>
            <scope>test</scope>
        </dependency>
    </dependencies>
</project>`
	pomPath := filepath.Join(ctx.Projects[2].Path, "pom.xml")
	if err := os.WriteFile(pomPath, []byte(pomXML), 0644); err != nil {
		t.Fatalf("Failed to create pom.xml: %v", err)
	}
}

// CreateTestProject creates a single test project
func (ctx *TestContext) CreateTestProject(t *testing.T, name, projectType string) models.Project {
	project := models.Project{
		Name:       name,
		Path:       filepath.Join(ctx.TempDir, name),
		Type:       projectType,
		ConfigFile: fmt.Sprintf("%s.json", projectType),
		Enabled:    true,
	}

	if err := ctx.DB.Create(&project).Error; err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	if err := os.MkdirAll(project.Path, 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	return project
}

// CreateTestDependency creates a test dependency
func (ctx *TestContext) CreateTestDependency(t *testing.T, projectID uint, name, currentVer, latestVer, depType string) models.Dependency {
	dependency := models.Dependency{
		ProjectID:      projectID,
		Name:           name,
		CurrentVersion: currentVer,
		LatestVersion:  latestVer,
		Type:           depType,
		Status:         "outdated",
	}

	if err := ctx.DB.Create(&dependency).Error; err != nil {
		t.Fatalf("Failed to create test dependency: %v", err)
	}

	return dependency
}

// CreateTestUpdate creates a test update
func (ctx *TestContext) CreateTestUpdate(t *testing.T, projectID uint, packageName, fromVer, toVer, updateType string) models.Update {
	update := models.Update{
		DependencyID:  1,
		FromVersion:   fromVer,
		ToVersion:     toVer,
		UpdateType:    updateType,
		Status:        "pending",
		Severity:      "medium",
		CreatedAt:     time.Now(),
	}

	if err := ctx.DB.Create(&update).Error; err != nil {
		t.Fatalf("Failed to create test update: %v", err)
	}

	return update
}

// CreateTestSecurityCheck creates a test security check
func (ctx *TestContext) CreateTestSecurityCheck(t *testing.T, packageName, version, checkType, status string) models.SecurityCheck {
	check := models.SecurityCheck{
		DependencyID: 1,
		PackageName:  packageName,
		Version:      version,
		Type:         checkType,
		CheckType:    checkType,
		Status:       status,
		Severity:     "medium",
		Details:      fmt.Sprintf("Test security check for %s", packageName),
		CheckedAt:    time.Now(),
		CreatedAt:    time.Now(),
	}

	if err := ctx.DB.Create(&check).Error; err != nil {
		t.Fatalf("Failed to create test security check: %v", err)
	}

	return check
}

// AssertNoError is a helper to assert no error occurred
func AssertNoError(t *testing.T, err error, msg string) {
	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
}

// AssertError is a helper to assert an error occurred
func AssertError(t *testing.T, err error, msg string) {
	if err == nil {
		t.Fatalf("%s: expected error but got none", msg)
	}
}

// AssertEqual is a helper to assert equality
func AssertEqual(t *testing.T, expected, actual interface{}, msg string) {
	if expected != actual {
		t.Fatalf("%s: expected %v, got %v", msg, expected, actual)
	}
}

// AssertNotEqual is a helper to assert inequality
func AssertNotEqual(t *testing.T, expected, actual interface{}, msg string) {
	if expected == actual {
		t.Fatalf("%s: expected values to be different, but both were %v", msg, expected)
	}
}

// AssertTrue is a helper to assert true
func AssertTrue(t *testing.T, condition bool, msg string) {
	if !condition {
		t.Fatalf("%s: expected true but got false", msg)
	}
}

// AssertFalse is a helper to assert false
func AssertFalse(t *testing.T, condition bool, msg string) {
	if condition {
		t.Fatalf("%s: expected false but got true", msg)
	}
}

// AssertContains is a helper to assert a slice contains an element
func AssertContains(t *testing.T, slice []string, element string, msg string) {
	for _, item := range slice {
		if item == element {
			return
		}
	}
	t.Fatalf("%s: slice %v does not contain %s", msg, slice, element)
}

// AssertLen is a helper to assert slice length
func AssertLen(t *testing.T, slice interface{}, expectedLen int, msg string) {
	var actualLen int
	
	switch s := slice.(type) {
	case []string:
		actualLen = len(s)
	case []models.Project:
		actualLen = len(s)
	case []models.Dependency:
		actualLen = len(s)
	case []models.Update:
		actualLen = len(s)
	default:
		t.Fatalf("%s: unsupported slice type for length assertion", msg)
	}
	
	if actualLen != expectedLen {
		t.Fatalf("%s: expected length %d, got %d", msg, expectedLen, actualLen)
	}
}

// MockHTTPServer provides a mock HTTP server for testing external API calls
type MockHTTPServer struct {
	Responses map[string]string
	StatusCodes map[string]int
}

// NewMockHTTPServer creates a new mock HTTP server
func NewMockHTTPServer() *MockHTTPServer {
	return &MockHTTPServer{
		Responses:   make(map[string]string),
		StatusCodes: make(map[string]int),
	}
}

// AddResponse adds a mock response for a URL
func (m *MockHTTPServer) AddResponse(url, response string, statusCode int) {
	m.Responses[url] = response
	m.StatusCodes[url] = statusCode
}

// WithTimeout creates a context with timeout for testing
func WithTimeout(t *testing.T, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// SkipIfShort skips the test if running in short mode
func SkipIfShort(t *testing.T, reason string) {
	if testing.Short() {
		t.Skipf("Skipping test in short mode: %s", reason)
	}
}
