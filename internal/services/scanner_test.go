package services

import (
	"context"
	"testing"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/models"
	testingPkg "github.com/8tcapital/ai-dep-manager/internal/testing"
)

func TestScannerService_ScanProject(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	scanner := NewScannerService()

	tests := []struct {
		name        string
		projectID   uint
		expectError bool
		expectCount int
	}{
		{
			name:        "scan existing project",
			projectID:   ctx.Projects[0].ID,
			expectError: false,
			expectCount: 2, // express and lodash dependencies
		},
		{
			name:        "scan non-existent project",
			projectID:   9999,
			expectError: true,
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := scanner.ScanProject(context.Background(), tt.projectID)

			if tt.expectError {
				testingPkg.AssertError(t, err, "Expected error for invalid project")
				return
			}

			testingPkg.AssertNoError(t, err, "ScanProject should not return error")
			testingPkg.AssertNotEqual(t, uint(0), result.ID, "Scan result should have ID")
			testingPkg.AssertEqual(t, tt.projectID, result.ProjectID, "Project ID should match")
			testingPkg.AssertEqual(t, "completed", result.Status, "Scan should complete successfully")
			testingPkg.AssertTrue(t, result.PackagesScanned >= tt.expectCount, "Should scan expected number of packages")
		})
	}
}

func TestScannerService_ScanAllProjects(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	scanner := NewScannerService()

	results, err := scanner.ScanAllProjects(context.Background())
	testingPkg.AssertNoError(t, err, "ScanAllProjects should not return error")
	testingPkg.AssertLen(t, results, 3, "Should scan all 3 test projects")

	for _, result := range results {
		testingPkg.AssertEqual(t, "completed", result.Status, "All scans should complete")
		testingPkg.AssertTrue(t, result.PackagesScanned > 0, "Should scan some packages")
	}
}

func TestScannerService_GetScanHistory(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	scanner := NewScannerService()

	// Create some scan history
	_, err := scanner.ScanProject(context.Background(), ctx.Projects[0].ID)
	testingPkg.AssertNoError(t, err, "First scan should succeed")

	time.Sleep(10 * time.Millisecond) // Ensure different timestamps

	_, err = scanner.ScanProject(context.Background(), ctx.Projects[0].ID)
	testingPkg.AssertNoError(t, err, "Second scan should succeed")

	// Get scan history
	history, err := scanner.GetScanHistory(context.Background(), &ctx.Projects[0].ID, 10)
	testingPkg.AssertNoError(t, err, "GetScanHistory should not return error")
	testingPkg.AssertTrue(t, len(history) >= 2, "Should have at least 2 scan results")

	// Verify ordering (most recent first)
	if len(history) >= 2 {
		testingPkg.AssertTrue(t, history[0].CreatedAt.After(history[1].CreatedAt), "Results should be ordered by date")
	}
}

func TestScannerService_GetLatestScan(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	scanner := NewScannerService()

	// No scans initially
	result, err := scanner.GetLatestScan(context.Background(), ctx.Projects[0].ID)
	testingPkg.AssertError(t, err, "Should return error when no scans exist")

	// Create a scan
	scanResult, err := scanner.ScanProject(context.Background(), ctx.Projects[0].ID)
	testingPkg.AssertNoError(t, err, "Scan should succeed")

	// Get latest scan
	result, err = scanner.GetLatestScan(context.Background(), ctx.Projects[0].ID)
	testingPkg.AssertNoError(t, err, "GetLatestScan should succeed")
	testingPkg.AssertEqual(t, scanResult.ID, result.ID, "Should return the correct scan result")
}

func TestScannerService_DiscoverProjects(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	scanner := NewScannerService()

	projects, err := scanner.DiscoverProjects(context.Background(), ctx.TempDir)
	testingPkg.AssertNoError(t, err, "DiscoverProjects should not return error")
	testingPkg.AssertTrue(t, len(projects) >= 3, "Should discover at least 3 projects")

	// Check that we found different project types
	foundTypes := make(map[string]bool)
	for _, project := range projects {
		foundTypes[project.Type] = true
	}

	testingPkg.AssertTrue(t, foundTypes["npm"], "Should discover npm project")
	testingPkg.AssertTrue(t, foundTypes["pip"], "Should discover pip project")
	testingPkg.AssertTrue(t, foundTypes["maven"], "Should discover maven project")
}

func TestScannerService_CheckForUpdates(t *testing.T) {
	testingPkg.SkipIfShort(t, "requires external API calls")

	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	scanner := NewScannerService()

	// Create a dependency that likely has updates
	dep := ctx.CreateTestDependency(t, ctx.Projects[0].ID, "lodash", "4.17.20", "4.17.21", "npm")

	updates, err := scanner.CheckForUpdates(context.Background(), ctx.Projects[0].ID)
	testingPkg.AssertNoError(t, err, "CheckForUpdates should not return error")

	// Should find at least one update
	testingPkg.AssertTrue(t, len(updates) > 0, "Should find some updates")

	// Verify update structure
	for _, update := range updates {
		testingPkg.AssertNotEqual(t, "", update.PackageName, "Update should have package name")
		testingPkg.AssertNotEqual(t, "", update.FromVersion, "Update should have from version")
		testingPkg.AssertNotEqual(t, "", update.ToVersion, "Update should have to version")
		testingPkg.AssertTrue(t, update.RiskScore >= 0, "Risk score should be non-negative")
		testingPkg.AssertTrue(t, update.Confidence >= 0 && update.Confidence <= 1, "Confidence should be between 0 and 1")
	}
}

func TestScannerService_GetProjectStats(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	scanner := NewScannerService()

	stats, err := scanner.GetProjectStats(context.Background(), ctx.Projects[0].ID)
	testingPkg.AssertNoError(t, err, "GetProjectStats should not return error")

	testingPkg.AssertEqual(t, ctx.Projects[0].ID, stats.ProjectID, "Project ID should match")
	testingPkg.AssertTrue(t, stats.TotalDependencies > 0, "Should have some dependencies")
	testingPkg.AssertTrue(t, stats.OutdatedDependencies >= 0, "Outdated count should be non-negative")
	testingPkg.AssertTrue(t, stats.SecurityIssues >= 0, "Security issues count should be non-negative")
}

func TestScannerService_ValidateProject(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	scanner := NewScannerService()

	tests := []struct {
		name        string
		projectPath string
		expectValid bool
		expectType  string
	}{
		{
			name:        "valid npm project",
			projectPath: ctx.Projects[0].Path,
			expectValid: true,
			expectType:  "npm",
		},
		{
			name:        "valid pip project",
			projectPath: ctx.Projects[1].Path,
			expectValid: true,
			expectType:  "pip",
		},
		{
			name:        "valid maven project",
			projectPath: ctx.Projects[2].Path,
			expectValid: true,
			expectType:  "maven",
		},
		{
			name:        "invalid project path",
			projectPath: "/nonexistent/path",
			expectValid: false,
			expectType:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid, projectType := scanner.ValidateProject(tt.projectPath)
			testingPkg.AssertEqual(t, tt.expectValid, isValid, "Project validity should match expected")
			if tt.expectValid {
				testingPkg.AssertEqual(t, tt.expectType, projectType, "Project type should match expected")
			}
		})
	}
}

func TestScannerService_ConcurrentScanning(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	scanner := NewScannerService()

	// Test concurrent scanning of multiple projects
	results := make(chan *models.ScanResult, len(ctx.Projects))
	errors := make(chan error, len(ctx.Projects))

	for _, project := range ctx.Projects {
		go func(projectID uint) {
			result, err := scanner.ScanProject(context.Background(), projectID)
			if err != nil {
				errors <- err
				return
			}
			results <- result
		}(project.ID)
	}

	// Collect results
	var scanResults []*models.ScanResult
	var scanErrors []error

	for i := 0; i < len(ctx.Projects); i++ {
		select {
		case result := <-results:
			scanResults = append(scanResults, result)
		case err := <-errors:
			scanErrors = append(scanErrors, err)
		case <-time.After(30 * time.Second):
			t.Fatal("Timeout waiting for concurrent scans to complete")
		}
	}

	testingPkg.AssertLen(t, scanErrors, 0, "No errors should occur during concurrent scanning")
	testingPkg.AssertLen(t, scanResults, len(ctx.Projects), "Should get results for all projects")

	for _, result := range scanResults {
		testingPkg.AssertEqual(t, "completed", result.Status, "All scans should complete successfully")
	}
}

func TestScannerService_ErrorHandling(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	scanner := NewScannerService()

	// Test with invalid project ID
	_, err := scanner.ScanProject(context.Background(), 9999)
	testingPkg.AssertError(t, err, "Should return error for invalid project ID")

	// Test with cancelled context
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = scanner.ScanProject(cancelledCtx, ctx.Projects[0].ID)
	testingPkg.AssertError(t, err, "Should return error for cancelled context")

	// Test with timeout context
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond) // Ensure timeout

	_, err = scanner.ScanProject(timeoutCtx, ctx.Projects[0].ID)
	testingPkg.AssertError(t, err, "Should return error for timed out context")
}

func TestScannerService_DatabaseIntegration(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	scanner := NewScannerService()

	// Verify scan results are persisted
	result, err := scanner.ScanProject(context.Background(), ctx.Projects[0].ID)
	testingPkg.AssertNoError(t, err, "Scan should succeed")

	// Verify in database
	var dbResult models.ScanResult
	err = ctx.DB.First(&dbResult, result.ID).Error
	testingPkg.AssertNoError(t, err, "Scan result should be in database")
	testingPkg.AssertEqual(t, result.ProjectID, dbResult.ProjectID, "Project ID should match")
	testingPkg.AssertEqual(t, result.Status, dbResult.Status, "Status should match")

	// Verify dependencies are updated
	var dependencies []models.Dependency
	err = ctx.DB.Where("project_id = ?", ctx.Projects[0].ID).Find(&dependencies).Error
	testingPkg.AssertNoError(t, err, "Should be able to query dependencies")
	testingPkg.AssertTrue(t, len(dependencies) > 0, "Should have dependencies")
}

func BenchmarkScannerService_ScanProject(b *testing.B) {
	ctx := testingPkg.SetupTestEnvironment(&testing.T{})
	defer ctx.Cleanup()

	scanner := NewScannerService()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := scanner.ScanProject(context.Background(), ctx.Projects[0].ID)
		if err != nil {
			b.Fatalf("Scan failed: %v", err)
		}
	}
}

func BenchmarkScannerService_DiscoverProjects(b *testing.B) {
	ctx := testingPkg.SetupTestEnvironment(&testing.T{})
	defer ctx.Cleanup()

	scanner := NewScannerService()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := scanner.DiscoverProjects(context.Background(), ctx.TempDir)
		if err != nil {
			b.Fatalf("Discovery failed: %v", err)
		}
	}
}
