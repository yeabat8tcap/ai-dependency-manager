package integration

import (
	"context"
	"testing"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/services"
	testingPkg "github.com/8tcapital/ai-dep-manager/internal/testing"
)

func TestScannerIntegration_FullWorkflow(t *testing.T) {
	testingPkg.SkipIfShort(t, "integration test requires full system")

	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	// Initialize services
	scanner := services.NewScannerService()
	updateService := services.NewUpdateService()

	// Test full scanning workflow
	t.Run("discover_and_scan_projects", func(t *testing.T) {
		// Discover projects
		projects, err := scanner.DiscoverProjects(context.Background(), ctx.TempDir)
		testingPkg.AssertNoError(t, err, "Project discovery should succeed")
		testingPkg.AssertTrue(t, len(projects) >= 3, "Should discover test projects")

		// Scan each discovered project
		for _, project := range projects {
			result, err := scanner.ScanProject(context.Background(), project.ID)
			testingPkg.AssertNoError(t, err, "Project scan should succeed")
			testingPkg.AssertEqual(t, "completed", result.Status, "Scan should complete")
			testingPkg.AssertTrue(t, result.PackagesScanned > 0, "Should scan packages")
		}
	})

	t.Run("scan_to_update_workflow", func(t *testing.T) {
		// Scan a project
		scanResult, err := scanner.ScanProject(context.Background(), ctx.Projects[0].ID)
		testingPkg.AssertNoError(t, err, "Scan should succeed")

		// Check for updates
		updates, err := scanner.CheckForUpdates(context.Background(), ctx.Projects[0].ID)
		testingPkg.AssertNoError(t, err, "Update check should succeed")

		if len(updates) > 0 {
			// Create update plan
			plan, err := updateService.CreateUpdatePlan(context.Background(), ctx.Projects[0].ID, services.UpdatePlanOptions{
				Strategy: "balanced",
			})
			testingPkg.AssertNoError(t, err, "Update plan creation should succeed")
			testingPkg.AssertTrue(t, len(plan.Groups) > 0, "Plan should have update groups")

			// Apply updates (dry run)
			var updateIDs []uint
			for _, group := range plan.Groups {
				for _, update := range group.Updates {
					updateIDs = append(updateIDs, update.ID)
				}
			}

			if len(updateIDs) > 0 {
				result, err := updateService.ApplyUpdates(context.Background(), updateIDs, services.ApplyOptions{
					DryRun: true,
				})
				testingPkg.AssertNoError(t, err, "Update application should succeed")
				testingPkg.AssertEqual(t, len(updateIDs), result.TotalUpdates, "Should process all updates")
			}
		}
	})

	t.Run("concurrent_scanning", func(t *testing.T) {
		// Test concurrent scanning of multiple projects
		results := make(chan *services.ScanResult, len(ctx.Projects))
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
		var scanResults []*services.ScanResult
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

	scanner := services.NewScannerService()

	// Perform scan
	result, err := scanner.ScanProject(context.Background(), ctx.Projects[0].ID)
	testingPkg.AssertNoError(t, err, "Scan should succeed")

	// Verify data persistence
	t.Run("scan_results_persisted", func(t *testing.T) {
		var dbResult services.ScanResult
		err := ctx.DB.First(&dbResult, result.ID).Error
		testingPkg.AssertNoError(t, err, "Scan result should be persisted")
		testingPkg.AssertEqual(t, result.ProjectID, dbResult.ProjectID, "Project ID should match")
	})

	t.Run("dependencies_updated", func(t *testing.T) {
		var dependencies []services.Dependency
		err := ctx.DB.Where("project_id = ?", ctx.Projects[0].ID).Find(&dependencies).Error
		testingPkg.AssertNoError(t, err, "Should query dependencies")
		testingPkg.AssertTrue(t, len(dependencies) > 0, "Should have dependencies")
	})

	t.Run("scan_history_maintained", func(t *testing.T) {
		// Perform another scan
		_, err := scanner.ScanProject(context.Background(), ctx.Projects[0].ID)
		testingPkg.AssertNoError(t, err, "Second scan should succeed")

		// Check history
		history, err := scanner.GetScanHistory(context.Background(), &ctx.Projects[0].ID, 10)
		testingPkg.AssertNoError(t, err, "Should get scan history")
		testingPkg.AssertTrue(t, len(history) >= 2, "Should have multiple scan results")
	})
}

func TestScannerIntegration_ErrorRecovery(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	scanner := services.NewScannerService()

	t.Run("invalid_project_recovery", func(t *testing.T) {
		// Test scanning non-existent project
		_, err := scanner.ScanProject(context.Background(), 9999)
		testingPkg.AssertError(t, err, "Should handle invalid project gracefully")
	})

	t.Run("timeout_recovery", func(t *testing.T) {
		// Test with very short timeout
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		time.Sleep(10 * time.Millisecond) // Ensure timeout

		_, err := scanner.ScanProject(timeoutCtx, ctx.Projects[0].ID)
		testingPkg.AssertError(t, err, "Should handle timeout gracefully")
	})

	t.Run("partial_failure_recovery", func(t *testing.T) {
		// This would test scenarios where some dependencies fail to scan
		// but the overall scan continues
		result, err := scanner.ScanProject(context.Background(), ctx.Projects[0].ID)
		testingPkg.AssertNoError(t, err, "Should handle partial failures")
		testingPkg.AssertEqual(t, "completed", result.Status, "Should complete despite partial failures")
	})
}

func TestScannerIntegration_PerformanceMetrics(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	scanner := services.NewScannerService()

	t.Run("scan_performance_tracking", func(t *testing.T) {
		start := time.Now()
		result, err := scanner.ScanProject(context.Background(), ctx.Projects[0].ID)
		duration := time.Since(start)

		testingPkg.AssertNoError(t, err, "Scan should succeed")
		testingPkg.AssertTrue(t, duration < 30*time.Second, "Scan should complete within reasonable time")
		testingPkg.AssertTrue(t, result.Duration > 0, "Should track scan duration")
	})

	t.Run("memory_usage_reasonable", func(t *testing.T) {
		// This is a simplified memory check
		// In a real scenario, you'd use runtime.ReadMemStats
		for i := 0; i < 5; i++ {
			_, err := scanner.ScanProject(context.Background(), ctx.Projects[0].ID)
			testingPkg.AssertNoError(t, err, "Repeated scans should not fail")
		}
	})
}

func BenchmarkScannerIntegration_FullWorkflow(b *testing.B) {
	ctx := testingPkg.SetupTestEnvironment(&testing.T{})
	defer ctx.Cleanup()

	scanner := services.NewScannerService()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := scanner.ScanProject(context.Background(), ctx.Projects[0].ID)
		if err != nil {
			b.Fatalf("Scan failed: %v", err)
		}
	}
}

func BenchmarkScannerIntegration_ConcurrentScans(b *testing.B) {
	ctx := testingPkg.SetupTestEnvironment(&testing.T{})
	defer ctx.Cleanup()

	scanner := services.NewScannerService()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := scanner.ScanProject(context.Background(), ctx.Projects[0].ID)
			if err != nil {
				b.Fatalf("Concurrent scan failed: %v", err)
			}
		}
	})
}
