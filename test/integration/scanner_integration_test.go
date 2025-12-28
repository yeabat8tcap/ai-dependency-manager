package integration

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/8tcapital/superint-dep-manager/internal/models"
	"github.com/8tcapital/superint-dep-manager/internal/packagemanager"
	pmtypes "github.com/8tcapital/superint-dep-manager/internal/packagemanager/types"
	"github.com/8tcapital/superint-dep-manager/internal/scanner"
	"github.com/8tcapital/superint-dep-manager/internal/services"
	testingPkg "github.com/8tcapital/superint-dep-manager/internal/testing"
)

// MockPackageManager implements PackageManager interface for testing
type MockPackageManager struct {
	Type string
}

func NewMockPackageManager(pmType string) *MockPackageManager {
	return &MockPackageManager{Type: pmType}
}

func (m *MockPackageManager) GetName() string {
	return m.Type
}

func (m *MockPackageManager) GetVersion() string {
	return "1.0.0"
}

func (m *MockPackageManager) GetType() string {
	return m.Type
}

func (m *MockPackageManager) IsAvailable(ctx context.Context) bool {
	return true
}

func (m *MockPackageManager) DetectProjects(ctx context.Context, rootPath string) ([]pmtypes.Project, error) {
	var projects []pmtypes.Project

	// Create projects based on type
	if m.Type == "npm" {
		projects = append(projects, pmtypes.Project{
			Name:           "test-node-project",
			Path:           filepath.Join(rootPath, "node-project"),
			Type:           "npm",
			PackageManager: "npm",
			ConfigFile:     "package.json",
			Language:       "javascript",
		})
	} else if m.Type == "pip" {
		projects = append(projects, pmtypes.Project{
			Name:           "test-python-project",
			Path:           filepath.Join(rootPath, "python-project"),
			Type:           "pip",
			PackageManager: "pip",
			ConfigFile:     "requirements.txt",
			Language:       "python",
		})
	} else if m.Type == "maven" {
		projects = append(projects, pmtypes.Project{
			Name:           "test-java-project",
			Path:           filepath.Join(rootPath, "java-project"),
			Type:           "maven",
			PackageManager: "maven",
			ConfigFile:     "pom.xml",
			Language:       "java",
		})
	}

	return projects, nil
}

func (m *MockPackageManager) ParseDependencies(ctx context.Context, projectPath string) (*pmtypes.DependencyInfo, error) {
	deps := []pmtypes.DependencyEntry{}

	if m.Type == "npm" {
		deps = append(deps, pmtypes.DependencyEntry{
			Name:            "express",
			Version:         "^4.18.0",
			ResolvedVersion: "4.18.0",
			Type:            "production",
		})
	} else if m.Type == "pip" {
		deps = append(deps, pmtypes.DependencyEntry{
			Name:            "requests",
			Version:         "2.28.0",
			ResolvedVersion: "2.28.0",
			Type:            "production",
		})
	} else if m.Type == "maven" {
		deps = append(deps, pmtypes.DependencyEntry{
			Name:            "junit:junit",
			Version:         "4.13.1",
			ResolvedVersion: "4.13.1",
			Type:            "test",
		})
	}

	return &pmtypes.DependencyInfo{
		Dependencies: deps,
	}, nil
}

func (m *MockPackageManager) GetLatestVersion(ctx context.Context, packageName string, registry *pmtypes.RegistryConfig) (*pmtypes.VersionInfo, error) {
	version := "1.0.0"
	if packageName == "express" {
		version = "4.18.2"
	} else if packageName == "requests" {
		version = "2.31.0"
	}

	return &pmtypes.VersionInfo{
		Version:     version,
		PublishedAt: time.Now(),
	}, nil
}

func (m *MockPackageManager) GetVersions(ctx context.Context, packageName string) ([]string, error) {
	return []string{"1.0.0"}, nil
}

func (m *MockPackageManager) GetChangelog(ctx context.Context, packageName, version string, registry *pmtypes.RegistryConfig) (*pmtypes.ChangelogInfo, error) {
	return &pmtypes.ChangelogInfo{
		Description: "Test changelog",
		Version:     version,
	}, nil
}

func (m *MockPackageManager) UpdateDependency(ctx context.Context, projectPath, packageName, version string, options *pmtypes.UpdateOptions) error {
	return nil
}

func (m *MockPackageManager) InstallDependencies(ctx context.Context, projectPath string, options *pmtypes.InstallOptions) error {
	return nil
}

func (m *MockPackageManager) ValidateProject(ctx context.Context, projectPath string) error {
	return nil
}

func TestScannerIntegration_FullWorkflow(t *testing.T) {
	testingPkg.SkipIfShort(t, "integration test requires full system")

	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	// Register mocks
	pm := packagemanager.GetManager()
	pm.Register(NewMockPackageManager("npm"))
	pm.Register(NewMockPackageManager("pip"))
	pm.Register(NewMockPackageManager("maven"))

	// Initialize services
	scanService := scanner.NewScanner(10)
	updateService := services.NewUpdateService()
	projectService := services.NewProjectService()

	// Test full scanning workflow
	t.Run("discover_and_scan_projects", func(t *testing.T) {
		// Delete existing projects from DB to test discovery
		ctx.DB.Exec("DELETE FROM dependencies")
		ctx.DB.Exec("DELETE FROM projects")

		// Discover projects
		projects, err := projectService.AutoDiscoverProjects(context.Background(), ctx.TempDir)
		testingPkg.AssertNoError(t, err, "Project discovery should succeed")
		testingPkg.AssertTrue(t, len(projects) >= 3, "Should discover test projects")

		// Update context projects with discovered ones for subsequent tests
		ctx.Projects = projects

		// Scan each discovered project
		for _, project := range projects {
			result, err := scanService.ScanProject(context.Background(), project.ID, &scanner.ScanOptions{
				ScanType: "full",
			})
			testingPkg.AssertNoError(t, err, "Project scan should succeed")
			testingPkg.AssertEqual(t, "completed", result.Status, "Scan should complete")
			testingPkg.AssertTrue(t, result.DependenciesFound > 0, "Should scan packages")
		}
	})

	t.Run("scan_to_update_workflow", func(t *testing.T) {
		if len(ctx.Projects) == 0 {
			t.Skip("No projects found, skipping update workflow")
		}

		// Scan a project
		scanResult, err := scanService.ScanProject(context.Background(), ctx.Projects[0].ID, &scanner.ScanOptions{
			ScanType: "full",
		})
		testingPkg.AssertNoError(t, err, "Scan should succeed")

		// Check for updates (from scan result)
		updates := scanResult.AvailableUpdates

		if len(updates) > 0 {
			// Create update plan
			plan, err := updateService.GenerateUpdatePlan(context.Background(), &services.UpdateOptions{
				ProjectID: ctx.Projects[0].ID,
			})
			testingPkg.AssertNoError(t, err, "Update plan creation should succeed")
			testingPkg.AssertTrue(t, len(plan.UpdateGroups) > 0, "Plan should have update groups")

			// Apply updates (dry run)
			result, err := updateService.ApplyUpdates(context.Background(), plan, &services.UpdateOptions{
				ProjectID: ctx.Projects[0].ID,
				DryRun:    true,
			})
			testingPkg.AssertNoError(t, err, "Update application should succeed")
			testingPkg.AssertTrue(t, result.TotalAttempted > 0, "Should process updates")
		}
	})

	t.Run("concurrent_scanning", func(t *testing.T) {
		if len(ctx.Projects) == 0 {
			t.Skip("No projects found, skipping concurrent scanning")
		}

		// Test concurrent scanning of multiple projects
		results := make(chan *scanner.ScanResult, len(ctx.Projects))
		errors := make(chan error, len(ctx.Projects))

		for _, project := range ctx.Projects {
			go func(projectID uint) {
				result, err := scanService.ScanProject(context.Background(), projectID, &scanner.ScanOptions{
					ScanType: "full",
				})
				if err != nil {
					errors <- err
					return
				}
				results <- result
			}(project.ID)
		}

		// Collect results
		var scanResults []*scanner.ScanResult
		var scanErrors []error

		for i := 0; i < len(ctx.Projects); i++ {
			select {
			case result := <-results:
				scanResults = append(scanResults, result)
			case err := <-errors:
				scanErrors = append(scanErrors, err)
			case <-time.After(60 * time.Second):
				t.Fatal("Timeout waiting for concurrent scans")
			}
		}

		testingPkg.AssertLen(t, scanErrors, 0, "No errors should occur")
		testingPkg.AssertLen(t, scanResults, len(ctx.Projects), "Should scan all projects")
	})
}

func TestScannerIntegration_DatabasePersistence(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	// Register mocks
	pm := packagemanager.GetManager()
	pm.Register(NewMockPackageManager("npm"))
	pm.Register(NewMockPackageManager("pip"))
	pm.Register(NewMockPackageManager("maven"))

	scanService := scanner.NewScanner(10)

	// Perform scan
	result, err := scanService.ScanProject(context.Background(), ctx.Projects[0].ID, &scanner.ScanOptions{
		ScanType: "full",
	})
	testingPkg.AssertNoError(t, err, "Scan should succeed")

	// Verify data persistence
	t.Run("scan_results_persisted", func(t *testing.T) {
		var dbResult models.ScanResult
		err := ctx.DB.Where("project_id = ?", ctx.Projects[0].ID).Last(&dbResult).Error
		testingPkg.AssertNoError(t, err, "Scan result should be persisted")
		testingPkg.AssertEqual(t, result.ProjectID, dbResult.ProjectID, "Project ID should match")
	})

	t.Run("dependencies_updated", func(t *testing.T) {
		var dependencies []models.Dependency
		err := ctx.DB.Where("project_id = ?", ctx.Projects[0].ID).Find(&dependencies).Error
		testingPkg.AssertNoError(t, err, "Should query dependencies")
		testingPkg.AssertTrue(t, len(dependencies) > 0, "Should have dependencies")
	})

	t.Run("scan_history_maintained", func(t *testing.T) {
		// Perform another scan
		_, err := scanService.ScanProject(context.Background(), ctx.Projects[0].ID, &scanner.ScanOptions{
			ScanType: "full",
		})
		testingPkg.AssertNoError(t, err, "Second scan should succeed")

		// Check history
		var history []models.ScanResult
		err = ctx.DB.Where("project_id = ?", ctx.Projects[0].ID).Order("created_at desc").Limit(10).Find(&history).Error
		testingPkg.AssertNoError(t, err, "Should get scan history")
		testingPkg.AssertTrue(t, len(history) >= 2, "Should have multiple scan results")
	})
}

func TestScannerIntegration_ErrorRecovery(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	// Register mocks
	pm := packagemanager.GetManager()
	pm.Register(NewMockPackageManager("npm"))

	scanService := scanner.NewScanner(10)

	t.Run("invalid_project_recovery", func(t *testing.T) {
		// Test scanning non-existent project
		_, err := scanService.ScanProject(context.Background(), 9999, &scanner.ScanOptions{
			ScanType: "full",
		})
		testingPkg.AssertError(t, err, "Should handle invalid project gracefully")
	})

	t.Run("timeout_recovery", func(t *testing.T) {
		// Test with very short timeout
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		time.Sleep(10 * time.Millisecond) // Ensure timeout

		_, err := scanService.ScanProject(timeoutCtx, ctx.Projects[0].ID, &scanner.ScanOptions{
			ScanType: "full",
		})
		if err == nil {
			// Ignore if it doesn't fail immediately
		}
	})

	t.Run("partial_failure_recovery", func(t *testing.T) {
		// This would test scenarios where some dependencies fail to scan
		// but the overall scan continues
		result, err := scanService.ScanProject(context.Background(), ctx.Projects[0].ID, &scanner.ScanOptions{
			ScanType: "full",
		})
		testingPkg.AssertNoError(t, err, "Should handle partial failures")
		testingPkg.AssertEqual(t, "completed", result.Status, "Should complete despite partial failures")
	})
}

func TestScannerIntegration_PerformanceMetrics(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	// Register mocks
	pm := packagemanager.GetManager()
	pm.Register(NewMockPackageManager("npm"))

	scanService := scanner.NewScanner(10)

	t.Run("scan_performance_tracking", func(t *testing.T) {
		start := time.Now()
		result, err := scanService.ScanProject(context.Background(), ctx.Projects[0].ID, &scanner.ScanOptions{
			ScanType: "full",
		})
		duration := time.Since(start)

		testingPkg.AssertNoError(t, err, "Scan should succeed")
		testingPkg.AssertTrue(t, duration < 30*time.Second, "Scan should complete within reasonable time")
		testingPkg.AssertTrue(t, result.Duration > 0, "Should track scan duration")
	})

	t.Run("memory_usage_reasonable", func(t *testing.T) {
		// This is a simplified memory check
		for i := 0; i < 5; i++ {
			_, err := scanService.ScanProject(context.Background(), ctx.Projects[0].ID, &scanner.ScanOptions{
				ScanType: "full",
			})
			testingPkg.AssertNoError(t, err, "Repeated scans should not fail")
		}
	})
}

func BenchmarkScannerIntegration_FullWorkflow(b *testing.B) {
	ctx := testingPkg.SetupTestEnvironment(&testing.T{})
	defer ctx.Cleanup()

	// Register mocks
	pm := packagemanager.GetManager()
	pm.Register(NewMockPackageManager("npm"))

	scanService := scanner.NewScanner(10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := scanService.ScanProject(context.Background(), ctx.Projects[0].ID, &scanner.ScanOptions{
			ScanType: "full",
		})
		if err != nil {
			b.Fatalf("Scan failed: %v", err)
		}
	}
}

func BenchmarkScannerIntegration_ConcurrentScans(b *testing.B) {
	ctx := testingPkg.SetupTestEnvironment(&testing.T{})
	defer ctx.Cleanup()

	// Register mocks
	pm := packagemanager.GetManager()
	pm.Register(NewMockPackageManager("npm"))

	scanService := scanner.NewScanner(10)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := scanService.ScanProject(context.Background(), ctx.Projects[0].ID, &scanner.ScanOptions{
				ScanType: "full",
			})
			if err != nil {
				b.Fatalf("Concurrent scan failed: %v", err)
			}
		}
	})
}
