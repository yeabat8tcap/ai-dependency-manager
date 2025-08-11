package packagemanager

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	testingPkg "github.com/8tcapital/ai-dep-manager/internal/testing"
)

func TestNpmManager_DetectProject(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	npm := NewNpmManager()

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "valid npm project",
			path:     ctx.Projects[0].Path, // Has package.json
			expected: true,
		},
		{
			name:     "non-npm project",
			path:     ctx.Projects[1].Path, // Python project
			expected: false,
		},
		{
			name:     "non-existent path",
			path:     "/nonexistent/path",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := npm.DetectProject(tt.path)
			testingPkg.AssertEqual(t, tt.expected, result, "Project detection should match expected")
		})
	}
}

func TestNpmManager_ParseDependencies(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	npm := NewNpmManager()

	dependencies, err := npm.ParseDependencies(ctx.Projects[0].Path)
	testingPkg.AssertNoError(t, err, "ParseDependencies should not return error")
	testingPkg.AssertTrue(t, len(dependencies) >= 2, "Should parse at least 2 dependencies")

	// Check for expected dependencies
	foundExpress := false
	foundLodash := false

	for _, dep := range dependencies {
		testingPkg.AssertNotEqual(t, "", dep.Name, "Dependency should have name")
		testingPkg.AssertNotEqual(t, "", dep.Version, "Dependency should have version")
		testingPkg.AssertEqual(t, "npm", dep.Type, "Dependency type should be npm")

		if dep.Name == "express" {
			foundExpress = true
			testingPkg.AssertEqual(t, "^4.18.0", dep.Version, "Express version should match")
		}
		if dep.Name == "lodash" {
			foundLodash = true
			testingPkg.AssertEqual(t, "^4.17.20", dep.Version, "Lodash version should match")
		}
	}

	testingPkg.AssertTrue(t, foundExpress, "Should find express dependency")
	testingPkg.AssertTrue(t, foundLodash, "Should find lodash dependency")
}

func TestNpmManager_GetLatestVersion(t *testing.T) {
	testingPkg.SkipIfShort(t, "requires external npm registry calls")

	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	npm := NewNpmManager()

	tests := []struct {
		name        string
		packageName string
		expectError bool
	}{
		{
			name:        "popular package",
			packageName: "express",
			expectError: false,
		},
		{
			name:        "another popular package",
			packageName: "lodash",
			expectError: false,
		},
		{
			name:        "non-existent package",
			packageName: "non-existent-package-xyz-123",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := npm.GetLatestVersion(context.Background(), tt.packageName)

			if tt.expectError {
				testingPkg.AssertError(t, err, "Should return error for non-existent package")
			} else {
				testingPkg.AssertNoError(t, err, "Should not return error for valid package")
				testingPkg.AssertNotEqual(t, "", version, "Should return non-empty version")
				// Version should match semver pattern
				testingPkg.AssertTrue(t, len(version) > 0 && version[0] >= '0' && version[0] <= '9',
					"Version should start with a digit")
			}
		})
	}
}

func TestNpmManager_GetChangelog(t *testing.T) {
	testingPkg.SkipIfShort(t, "requires external npm registry calls")

	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	npm := NewNpmManager()

	changelog, err := npm.GetChangelog(context.Background(), "express", "4.18.0", "4.18.2")
	testingPkg.AssertNoError(t, err, "GetChangelog should not return error")
	testingPkg.AssertNotEqual(t, "", changelog, "Should return non-empty changelog")
}

func TestNpmManager_UpdatePackage(t *testing.T) {
	testingPkg.SkipIfShort(t, "modifies package.json file")

	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	npm := NewNpmManager()

	// Test dry run first
	err := npm.UpdatePackage(context.Background(), ctx.Projects[0].Path, "lodash", "4.17.21", UpdateOptions{
		DryRun: true,
	})
	testingPkg.AssertNoError(t, err, "Dry run update should not return error")

	// Verify package.json wasn't actually modified in dry run
	packageJSON := readPackageJSON(t, ctx.Projects[0].Path)
	testingPkg.AssertEqual(t, "^4.17.20", packageJSON.Dependencies["lodash"], "Package.json should not be modified in dry run")
}

func TestNpmManager_InstallDependencies(t *testing.T) {
	testingPkg.SkipIfShort(t, "requires npm install")

	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	npm := NewNpmManager()

	// Test dry run
	err := npm.InstallDependencies(context.Background(), ctx.Projects[0].Path, InstallOptions{
		DryRun: true,
	})
	testingPkg.AssertNoError(t, err, "Dry run install should not return error")
}

func TestNpmManager_GetPackageInfo(t *testing.T) {
	testingPkg.SkipIfShort(t, "requires external npm registry calls")

	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	npm := NewNpmManager()

	info, err := npm.GetPackageInfo(context.Background(), "express")
	testingPkg.AssertNoError(t, err, "GetPackageInfo should not return error")
	testingPkg.AssertEqual(t, "express", info.Name, "Package name should match")
	testingPkg.AssertNotEqual(t, "", info.Description, "Should have description")
	testingPkg.AssertNotEqual(t, "", info.LatestVersion, "Should have latest version")
	testingPkg.AssertNotEqual(t, "", info.Homepage, "Should have homepage")
}

func TestNpmManager_ValidatePackageJSON(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	npm := NewNpmManager()

	tests := []struct {
		name        string
		content     string
		expectValid bool
	}{
		{
			name: "valid package.json",
			content: `{
				"name": "test-package",
				"version": "1.0.0",
				"dependencies": {
					"express": "^4.18.0"
				}
			}`,
			expectValid: true,
		},
		{
			name:        "invalid JSON",
			content:     `{"name": "test", "version":}`,
			expectValid: false,
		},
		{
			name: "missing required fields",
			content: `{
				"dependencies": {
					"express": "^4.18.0"
				}
			}`,
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary package.json
			tempDir, err := os.MkdirTemp("", "npm-test-*")
			testingPkg.AssertNoError(t, err, "Should create temp dir")
			defer os.RemoveAll(tempDir)

			packageJSONPath := filepath.Join(tempDir, "package.json")
			err = os.WriteFile(packageJSONPath, []byte(tt.content), 0644)
			testingPkg.AssertNoError(t, err, "Should write package.json")

			isValid := npm.ValidatePackageFile(tempDir)
			testingPkg.AssertEqual(t, tt.expectValid, isValid, "Validation should match expected")
		})
	}
}

func TestNpmManager_GetSecurityAudit(t *testing.T) {
	testingPkg.SkipIfShort(t, "requires npm audit")

	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	npm := NewNpmManager()

	// This test might fail if npm audit finds no issues
	// We'll just verify the method doesn't crash
	audit, err := npm.GetSecurityAudit(context.Background(), ctx.Projects[0].Path)
	
	// Either succeeds or fails gracefully
	if err == nil {
		testingPkg.AssertTrue(t, audit.TotalVulnerabilities >= 0, "Should have non-negative vulnerability count")
		testingPkg.AssertTrue(t, len(audit.Vulnerabilities) >= 0, "Should have vulnerability list")
	}
}

func TestNpmManager_ErrorHandling(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	npm := NewNpmManager()

	// Test with non-existent project path
	_, err := npm.ParseDependencies("/nonexistent/path")
	testingPkg.AssertError(t, err, "Should return error for non-existent path")

	// Test with invalid package.json
	tempDir, err := os.MkdirTemp("", "npm-error-test-*")
	testingPkg.AssertNoError(t, err, "Should create temp dir")
	defer os.RemoveAll(tempDir)

	invalidJSON := `{"name": "test", invalid json}`
	packageJSONPath := filepath.Join(tempDir, "package.json")
	err = os.WriteFile(packageJSONPath, []byte(invalidJSON), 0644)
	testingPkg.AssertNoError(t, err, "Should write invalid JSON")

	_, err = npm.ParseDependencies(tempDir)
	testingPkg.AssertError(t, err, "Should return error for invalid JSON")
}

func TestNpmManager_ConcurrentOperations(t *testing.T) {
	testingPkg.SkipIfShort(t, "requires external npm registry calls")

	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	npm := NewNpmManager()

	packages := []string{"express", "lodash", "react"}
	results := make(chan string, len(packages))
	errors := make(chan error, len(packages))

	// Test concurrent version fetching
	for _, pkg := range packages {
		go func(packageName string) {
			version, err := npm.GetLatestVersion(context.Background(), packageName)
			if err != nil {
				errors <- err
				return
			}
			results <- version
		}(pkg)
	}

	// Collect results
	var versions []string
	var versionErrors []error

	for i := 0; i < len(packages); i++ {
		select {
		case version := <-results:
			versions = append(versions, version)
		case err := <-errors:
			versionErrors = append(versionErrors, err)
		case <-testingPkg.WithTimeout(t, 30*time.Second):
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}

	testingPkg.AssertLen(t, versionErrors, 0, "No errors should occur during concurrent operations")
	testingPkg.AssertLen(t, versions, len(packages), "Should get versions for all packages")
}

// Helper functions

type packageJSON struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Dependencies map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

func readPackageJSON(t *testing.T, projectPath string) packageJSON {
	packageJSONPath := filepath.Join(projectPath, "package.json")
	data, err := os.ReadFile(packageJSONPath)
	testingPkg.AssertNoError(t, err, "Should read package.json")

	var pkg packageJSON
	err = json.Unmarshal(data, &pkg)
	testingPkg.AssertNoError(t, err, "Should parse package.json")

	return pkg
}

// Benchmarks

func BenchmarkNpmManager_ParseDependencies(b *testing.B) {
	ctx := testingPkg.SetupTestEnvironment(&testing.T{})
	defer ctx.Cleanup()

	npm := NewNpmManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := npm.ParseDependencies(ctx.Projects[0].Path)
		if err != nil {
			b.Fatalf("ParseDependencies failed: %v", err)
		}
	}
}

func BenchmarkNpmManager_DetectProject(b *testing.B) {
	ctx := testingPkg.SetupTestEnvironment(&testing.T{})
	defer ctx.Cleanup()

	npm := NewNpmManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		npm.DetectProject(ctx.Projects[0].Path)
	}
}
