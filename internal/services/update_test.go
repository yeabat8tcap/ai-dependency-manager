package services

import (
	"context"
	"testing"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/models"
	testingPkg "github.com/8tcapital/ai-dep-manager/internal/testing"
)

func TestUpdateService_CreateUpdatePlan(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	updateService := NewUpdateService()

	// Create some test updates
	update1 := ctx.CreateTestUpdate(t, ctx.Projects[0].ID, "express", "4.18.0", "4.18.2", "patch")
	update2 := ctx.CreateTestUpdate(t, ctx.Projects[0].ID, "lodash", "4.17.20", "4.17.21", "security")

	plan, err := updateService.CreateUpdatePlan(context.Background(), ctx.Projects[0].ID, UpdatePlanOptions{
		Strategy: "balanced",
	})

	testingPkg.AssertNoError(t, err, "CreateUpdatePlan should not return error")
	testingPkg.AssertNotEqual(t, uint(0), plan.ID, "Plan should have ID")
	testingPkg.AssertEqual(t, ctx.Projects[0].ID, plan.ProjectID, "Project ID should match")
	testingPkg.AssertTrue(t, len(plan.Groups) > 0, "Plan should have update groups")

	// Verify security updates are prioritized
	securityGroup := findGroupByRisk(plan.Groups, "security")
	testingPkg.AssertTrue(t, securityGroup != nil, "Should have security group")
	testingPkg.AssertTrue(t, len(securityGroup.Updates) > 0, "Security group should have updates")
}

func TestUpdateService_ApplyUpdates(t *testing.T) {
	testingPkg.SkipIfShort(t, "requires package manager operations")

	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	updateService := NewUpdateService()

	// Create a test update
	update := ctx.CreateTestUpdate(t, ctx.Projects[0].ID, "lodash", "4.17.20", "4.17.21", "patch")

	result, err := updateService.ApplyUpdates(context.Background(), []uint{update.ID}, ApplyOptions{
		DryRun: true, // Use dry run to avoid actual package manager calls
	})

	testingPkg.AssertNoError(t, err, "ApplyUpdates should not return error")
	testingPkg.AssertNotEqual(t, uint(0), result.ID, "Result should have ID")
	testingPkg.AssertEqual(t, 1, result.TotalUpdates, "Should have 1 total update")
	testingPkg.AssertEqual(t, 1, result.SuccessfulUpdates, "Should have 1 successful update in dry run")
}

func TestUpdateService_GetUpdateRecommendations(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	updateService := NewUpdateService()

	// Create test updates with different risk levels
	ctx.CreateTestUpdate(t, ctx.Projects[0].ID, "express", "4.18.0", "4.18.2", "patch")
	ctx.CreateTestUpdate(t, ctx.Projects[0].ID, "lodash", "4.17.20", "4.17.21", "security")
	ctx.CreateTestUpdate(t, ctx.Projects[0].ID, "react", "17.0.0", "18.0.0", "major")

	recommendations, err := updateService.GetUpdateRecommendations(context.Background(), ctx.Projects[0].ID)
	testingPkg.AssertNoError(t, err, "GetUpdateRecommendations should not return error")
	testingPkg.AssertTrue(t, len(recommendations) > 0, "Should have recommendations")

	// Verify recommendations are sorted by priority
	for i := 1; i < len(recommendations); i++ {
		testingPkg.AssertTrue(t, recommendations[i-1].Priority >= recommendations[i].Priority,
			"Recommendations should be sorted by priority")
	}
}

func TestUpdateService_RiskGrouping(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	updateService := NewUpdateService()

	// Create updates with different risk levels
	securityUpdate := ctx.CreateTestUpdate(t, ctx.Projects[0].ID, "security-pkg", "1.0.0", "1.0.1", "security")
	securityUpdate.RiskScore = 8.5
	ctx.DB.Save(&securityUpdate)

	patchUpdate := ctx.CreateTestUpdate(t, ctx.Projects[0].ID, "patch-pkg", "1.0.0", "1.0.1", "patch")
	patchUpdate.RiskScore = 1.5
	ctx.DB.Save(&patchUpdate)

	majorUpdate := ctx.CreateTestUpdate(t, ctx.Projects[0].ID, "major-pkg", "1.0.0", "2.0.0", "major")
	majorUpdate.RiskScore = 6.0
	ctx.DB.Save(&majorUpdate)

	plan, err := updateService.CreateUpdatePlan(context.Background(), ctx.Projects[0].ID, UpdatePlanOptions{
		Strategy: "balanced",
	})

	testingPkg.AssertNoError(t, err, "CreateUpdatePlan should not return error")

	// Verify risk-based grouping
	securityGroup := findGroupByRisk(plan.Groups, "security")
	patchGroup := findGroupByRisk(plan.Groups, "low")
	majorGroup := findGroupByRisk(plan.Groups, "high")

	testingPkg.AssertTrue(t, securityGroup != nil, "Should have security group")
	testingPkg.AssertTrue(t, patchGroup != nil, "Should have low risk group")
	testingPkg.AssertTrue(t, majorGroup != nil, "Should have high risk group")

	// Verify execution order
	testingPkg.AssertTrue(t, securityGroup.ExecutionOrder < patchGroup.ExecutionOrder,
		"Security updates should execute before patch updates")
	testingPkg.AssertTrue(t, patchGroup.ExecutionOrder < majorGroup.ExecutionOrder,
		"Patch updates should execute before major updates")
}

func TestUpdateService_UpdateStrategies(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	updateService := NewUpdateService()

	// Create test updates
	ctx.CreateTestUpdate(t, ctx.Projects[0].ID, "pkg1", "1.0.0", "1.1.0", "minor")
	ctx.CreateTestUpdate(t, ctx.Projects[0].ID, "pkg2", "1.0.0", "2.0.0", "major")

	strategies := []string{"conservative", "balanced", "aggressive"}

	for _, strategy := range strategies {
		t.Run(strategy, func(t *testing.T) {
			plan, err := updateService.CreateUpdatePlan(context.Background(), ctx.Projects[0].ID, UpdatePlanOptions{
				Strategy: strategy,
			})

			testingPkg.AssertNoError(t, err, "CreateUpdatePlan should not return error for "+strategy)
			testingPkg.AssertTrue(t, len(plan.Groups) > 0, "Plan should have groups for "+strategy)

			// Verify strategy-specific behavior
			switch strategy {
			case "conservative":
				// Conservative should have more groups with smaller batch sizes
				testingPkg.AssertTrue(t, len(plan.Groups) >= 2, "Conservative should have multiple groups")
			case "aggressive":
				// Aggressive might have fewer groups with larger batch sizes
				testingPkg.AssertTrue(t, plan.EstimatedDuration < time.Hour, "Aggressive should be faster")
			}
		})
	}
}

func TestUpdateService_FilterUpdates(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	updateService := NewUpdateService()

	// Create various test updates
	ctx.CreateTestUpdate(t, ctx.Projects[0].ID, "security-pkg", "1.0.0", "1.0.1", "security")
	ctx.CreateTestUpdate(t, ctx.Projects[0].ID, "patch-pkg", "1.0.0", "1.0.1", "patch")
	ctx.CreateTestUpdate(t, ctx.Projects[0].ID, "major-pkg", "1.0.0", "2.0.0", "major")

	tests := []struct {
		name     string
		filter   UpdateFilter
		expected int
	}{
		{
			name:     "security only",
			filter:   UpdateFilter{Types: []string{"security"}},
			expected: 1,
		},
		{
			name:     "patch and security",
			filter:   UpdateFilter{Types: []string{"patch", "security"}},
			expected: 2,
		},
		{
			name:     "all types",
			filter:   UpdateFilter{},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updates, err := updateService.GetAvailableUpdates(context.Background(), ctx.Projects[0].ID, tt.filter)
			testingPkg.AssertNoError(t, err, "GetAvailableUpdates should not return error")
			testingPkg.AssertLen(t, updates, tt.expected, "Should return expected number of updates")
		})
	}
}

func TestUpdateService_RollbackPlanCreation(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	updateService := NewUpdateService()

	// Create and apply an update
	update := ctx.CreateTestUpdate(t, ctx.Projects[0].ID, "test-pkg", "1.0.0", "1.1.0", "minor")

	result, err := updateService.ApplyUpdates(context.Background(), []uint{update.ID}, ApplyOptions{
		CreateRollbackPlan: true,
		DryRun:             true,
	})

	testingPkg.AssertNoError(t, err, "ApplyUpdates should not return error")
	testingPkg.AssertTrue(t, result.RollbackPlanID != nil, "Should create rollback plan")

	// Verify rollback plan exists
	var rollbackPlan models.RollbackPlan
	err = ctx.DB.First(&rollbackPlan, *result.RollbackPlanID).Error
	testingPkg.AssertNoError(t, err, "Rollback plan should exist in database")
	testingPkg.AssertEqual(t, ctx.Projects[0].ID, rollbackPlan.ProjectID, "Rollback plan should have correct project ID")
}

func TestUpdateService_BatchProcessing(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	updateService := NewUpdateService()

	// Create multiple updates
	var updateIDs []uint
	for i := 0; i < 5; i++ {
		update := ctx.CreateTestUpdate(t, ctx.Projects[0].ID, 
			fmt.Sprintf("pkg%d", i), "1.0.0", "1.1.0", "patch")
		updateIDs = append(updateIDs, update.ID)
	}

	result, err := updateService.ApplyUpdates(context.Background(), updateIDs, ApplyOptions{
		BatchSize: 2,
		DryRun:    true,
	})

	testingPkg.AssertNoError(t, err, "Batch processing should not return error")
	testingPkg.AssertEqual(t, 5, result.TotalUpdates, "Should process all updates")
	testingPkg.AssertTrue(t, result.ProcessingTime > 0, "Should have processing time")
}

func TestUpdateService_ConcurrencyControl(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	updateService := NewUpdateService()

	// Create updates for multiple projects
	update1 := ctx.CreateTestUpdate(t, ctx.Projects[0].ID, "pkg1", "1.0.0", "1.1.0", "patch")
	update2 := ctx.CreateTestUpdate(t, ctx.Projects[1].ID, "pkg2", "1.0.0", "1.1.0", "patch")

	// Test concurrent updates
	results := make(chan *ApplyResult, 2)
	errors := make(chan error, 2)

	for _, updateID := range []uint{update1.ID, update2.ID} {
		go func(id uint) {
			result, err := updateService.ApplyUpdates(context.Background(), []uint{id}, ApplyOptions{
				DryRun: true,
			})
			if err != nil {
				errors <- err
				return
			}
			results <- result
		}(updateID)
	}

	// Collect results
	var applyResults []*ApplyResult
	var applyErrors []error

	for i := 0; i < 2; i++ {
		select {
		case result := <-results:
			applyResults = append(applyResults, result)
		case err := <-errors:
			applyErrors = append(applyErrors, err)
		case <-time.After(30 * time.Second):
			t.Fatal("Timeout waiting for concurrent updates")
		}
	}

	testingPkg.AssertLen(t, applyErrors, 0, "No errors should occur during concurrent updates")
	testingPkg.AssertLen(t, applyResults, 2, "Should get results for both updates")
}

func TestUpdateService_ErrorHandling(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	updateService := NewUpdateService()

	// Test with non-existent project
	_, err := updateService.CreateUpdatePlan(context.Background(), 9999, UpdatePlanOptions{})
	testingPkg.AssertError(t, err, "Should return error for non-existent project")

	// Test with non-existent update
	_, err = updateService.ApplyUpdates(context.Background(), []uint{9999}, ApplyOptions{})
	testingPkg.AssertError(t, err, "Should return error for non-existent update")

	// Test with cancelled context
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = updateService.CreateUpdatePlan(cancelledCtx, ctx.Projects[0].ID, UpdatePlanOptions{})
	testingPkg.AssertError(t, err, "Should return error for cancelled context")
}

func TestUpdateService_AIIntegration(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	updateService := NewUpdateService()

	// Create an update that would benefit from AI analysis
	update := ctx.CreateTestUpdate(t, ctx.Projects[0].ID, "react", "17.0.0", "18.0.0", "major")

	recommendations, err := updateService.GetUpdateRecommendations(context.Background(), ctx.Projects[0].ID)
	testingPkg.AssertNoError(t, err, "GetUpdateRecommendations should not return error")

	// Verify AI analysis is included
	for _, rec := range recommendations {
		if rec.PackageName == "react" {
			testingPkg.AssertTrue(t, rec.RiskScore > 0, "Should have risk score from AI analysis")
			testingPkg.AssertTrue(t, rec.Confidence > 0, "Should have confidence score from AI analysis")
			testingPkg.AssertNotEqual(t, "", rec.Reasoning, "Should have AI reasoning")
			break
		}
	}
}

func TestUpdateService_UpdateHistory(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	updateService := NewUpdateService()

	// Apply some updates to create history
	update := ctx.CreateTestUpdate(t, ctx.Projects[0].ID, "test-pkg", "1.0.0", "1.1.0", "patch")
	
	_, err := updateService.ApplyUpdates(context.Background(), []uint{update.ID}, ApplyOptions{
		DryRun: true,
	})
	testingPkg.AssertNoError(t, err, "ApplyUpdates should not return error")

	// Get update history
	history, err := updateService.GetUpdateHistory(context.Background(), &ctx.Projects[0].ID, 10)
	testingPkg.AssertNoError(t, err, "GetUpdateHistory should not return error")
	testingPkg.AssertTrue(t, len(history) > 0, "Should have update history")

	// Verify history entry
	historyEntry := history[0]
	testingPkg.AssertEqual(t, ctx.Projects[0].ID, historyEntry.ProjectID, "History should have correct project ID")
	testingPkg.AssertTrue(t, historyEntry.TotalUpdates > 0, "History should show updates applied")
}

// Helper functions

func findGroupByRisk(groups []UpdateGroup, riskLevel string) *UpdateGroup {
	for _, group := range groups {
		if group.RiskLevel == riskLevel {
			return &group
		}
	}
	return nil
}

// Benchmarks

func BenchmarkUpdateService_CreateUpdatePlan(b *testing.B) {
	ctx := testingPkg.SetupTestEnvironment(&testing.T{})
	defer ctx.Cleanup()

	updateService := NewUpdateService()

	// Create test updates
	for i := 0; i < 10; i++ {
		ctx.CreateTestUpdate(&testing.T{}, ctx.Projects[0].ID, 
			fmt.Sprintf("pkg%d", i), "1.0.0", "1.1.0", "patch")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := updateService.CreateUpdatePlan(context.Background(), ctx.Projects[0].ID, UpdatePlanOptions{
			Strategy: "balanced",
		})
		if err != nil {
			b.Fatalf("CreateUpdatePlan failed: %v", err)
		}
	}
}

func BenchmarkUpdateService_GetUpdateRecommendations(b *testing.B) {
	ctx := testingPkg.SetupTestEnvironment(&testing.T{})
	defer ctx.Cleanup()

	updateService := NewUpdateService()

	// Create test updates
	for i := 0; i < 20; i++ {
		ctx.CreateTestUpdate(&testing.T{}, ctx.Projects[0].ID, 
			fmt.Sprintf("pkg%d", i), "1.0.0", "1.1.0", "patch")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := updateService.GetUpdateRecommendations(context.Background(), ctx.Projects[0].ID)
		if err != nil {
			b.Fatalf("GetUpdateRecommendations failed: %v", err)
		}
	}
}
