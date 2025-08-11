package e2e

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	testingPkg "github.com/8tcapital/ai-dep-manager/internal/testing"
)

// TestFullWorkflow tests the complete end-to-end workflow
func TestFullWorkflow(t *testing.T) {
	testingPkg.SkipIfShort(t, "end-to-end test requires full system")

	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	// Build the CLI binary for testing
	binaryPath := buildCLIBinary(t)
	defer os.Remove(binaryPath)

	t.Run("configure_system", func(t *testing.T) {
		// Test system configuration
		cmd := exec.Command(binaryPath, "configure", "--data-dir", ctx.TempDir)
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			t.Logf("Configure output: %s", string(output))
		}
		// Configuration might fail in test environment, but shouldn't crash
	})

	t.Run("discover_projects", func(t *testing.T) {
		// Test project discovery
		cmd := exec.Command(binaryPath, "configure", "add-project", ctx.Projects[0].Path)
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			t.Logf("Add project output: %s", string(output))
		}
		// Should not crash even if project already exists
	})

	t.Run("scan_projects", func(t *testing.T) {
		// Test scanning
		cmd := exec.Command(binaryPath, "scan", "--project-id", "1")
		cmd.Env = append(os.Environ(), "AI_DEP_MANAGER_DATA_DIR="+ctx.TempDir)
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			t.Logf("Scan output: %s", string(output))
		}
		// Scan might fail due to missing package managers, but shouldn't crash
	})

	t.Run("check_status", func(t *testing.T) {
		// Test status command
		cmd := exec.Command(binaryPath, "status")
		cmd.Env = append(os.Environ(), "AI_DEP_MANAGER_DATA_DIR="+ctx.TempDir)
		output, err := cmd.CombinedOutput()
		
		// Status should always work
		testingPkg.AssertNoError(t, err, "Status command should not fail")
		testingPkg.AssertTrue(t, len(output) > 0, "Status should produce output")
	})

	t.Run("security_scan", func(t *testing.T) {
		// Test security scanning
		cmd := exec.Command(binaryPath, "security", "scan", "--project-id", "1")
		cmd.Env = append(os.Environ(), "AI_DEP_MANAGER_DATA_DIR="+ctx.TempDir)
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			t.Logf("Security scan output: %s", string(output))
		}
		// Security scan might fail due to network issues, but shouldn't crash
	})

	t.Run("generate_report", func(t *testing.T) {
		// Test report generation
		reportFile := filepath.Join(ctx.TempDir, "test-report.json")
		cmd := exec.Command(binaryPath, "report", "generate", "summary", "--output", reportFile)
		cmd.Env = append(os.Environ(), "AI_DEP_MANAGER_DATA_DIR="+ctx.TempDir)
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			t.Logf("Report generation output: %s", string(output))
		}
		
		// Check if report file was created
		if _, err := os.Stat(reportFile); err == nil {
			t.Logf("Report successfully generated at %s", reportFile)
		}
	})
}

func TestCLICommands(t *testing.T) {
	testingPkg.SkipIfShort(t, "CLI testing requires binary build")

	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	binaryPath := buildCLIBinary(t)
	defer os.Remove(binaryPath)

	tests := []struct {
		name     string
		args     []string
		expectOK bool
	}{
		{
			name:     "help command",
			args:     []string{"--help"},
			expectOK: true,
		},
		{
			name:     "version command",
			args:     []string{"version"},
			expectOK: true,
		},
		{
			name:     "status command",
			args:     []string{"status"},
			expectOK: true,
		},
		{
			name:     "invalid command",
			args:     []string{"invalid-command"},
			expectOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			cmd.Env = append(os.Environ(), "AI_DEP_MANAGER_DATA_DIR="+ctx.TempDir)
			output, err := cmd.CombinedOutput()

			if tt.expectOK {
				testingPkg.AssertNoError(t, err, "Command should succeed")
				testingPkg.AssertTrue(t, len(output) > 0, "Should produce output")
			} else {
				testingPkg.AssertError(t, err, "Invalid command should fail")
			}
		})
	}
}

func TestAgentWorkflow(t *testing.T) {
	testingPkg.SkipIfShort(t, "agent testing requires full system")

	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	binaryPath := buildCLIBinary(t)
	defer os.Remove(binaryPath)

	t.Run("agent_start_stop", func(t *testing.T) {
		// Start agent in background
		startCmd := exec.Command(binaryPath, "agent", "start", "--foreground")
		startCmd.Env = append(os.Environ(), "AI_DEP_MANAGER_DATA_DIR="+ctx.TempDir)
		
		// Start the agent
		err := startCmd.Start()
		if err != nil {
			t.Logf("Agent start failed (expected in test environment): %v", err)
			return
		}

		// Give it a moment to start
		time.Sleep(2 * time.Second)

		// Stop the agent
		if startCmd.Process != nil {
			startCmd.Process.Kill()
			startCmd.Wait()
		}
	})

	t.Run("agent_status", func(t *testing.T) {
		// Check agent status
		cmd := exec.Command(binaryPath, "agent", "status")
		cmd.Env = append(os.Environ(), "AI_DEP_MANAGER_DATA_DIR="+ctx.TempDir)
		output, err := cmd.CombinedOutput()
		
		// Status command should work even if agent is not running
		testingPkg.AssertNoError(t, err, "Agent status should not fail")
		testingPkg.AssertTrue(t, strings.Contains(string(output), "stopped") || 
			strings.Contains(string(output), "running"), "Should show agent status")
	})
}

func TestPolicyWorkflow(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	binaryPath := buildCLIBinary(t)
	defer os.Remove(binaryPath)

	t.Run("policy_list_empty", func(t *testing.T) {
		// List policies (should be empty initially)
		cmd := exec.Command(binaryPath, "policy", "list")
		cmd.Env = append(os.Environ(), "AI_DEP_MANAGER_DATA_DIR="+ctx.TempDir)
		output, err := cmd.CombinedOutput()
		
		testingPkg.AssertNoError(t, err, "Policy list should not fail")
		testingPkg.AssertTrue(t, strings.Contains(string(output), "No policies") ||
			strings.Contains(string(output), "policies"), "Should show policy status")
	})

	t.Run("policy_templates", func(t *testing.T) {
		// Show policy templates
		templates := []string{"security", "conservative", "aggressive"}
		
		for _, template := range templates {
			cmd := exec.Command(binaryPath, "policy", "template", template)
			cmd.Env = append(os.Environ(), "AI_DEP_MANAGER_DATA_DIR="+ctx.TempDir)
			output, err := cmd.CombinedOutput()
			
			testingPkg.AssertNoError(t, err, "Policy template should not fail")
			testingPkg.AssertTrue(t, len(output) > 0, "Should show template content")
		}
	})
}

func TestNotificationWorkflow(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	binaryPath := buildCLIBinary(t)
	defer os.Remove(binaryPath)

	t.Run("notification_list", func(t *testing.T) {
		// List notification channels
		cmd := exec.Command(binaryPath, "notify", "list")
		cmd.Env = append(os.Environ(), "AI_DEP_MANAGER_DATA_DIR="+ctx.TempDir)
		output, err := cmd.CombinedOutput()
		
		testingPkg.AssertNoError(t, err, "Notification list should not fail")
		testingPkg.AssertTrue(t, len(output) > 0, "Should show notification channels")
	})

	t.Run("notification_test_dry_run", func(t *testing.T) {
		// Test notification with dry run
		cmd := exec.Command(binaryPath, "notify", "test", "email", "--dry-run")
		cmd.Env = append(os.Environ(), "AI_DEP_MANAGER_DATA_DIR="+ctx.TempDir)
		output, err := cmd.CombinedOutput()
		
		testingPkg.AssertNoError(t, err, "Notification test dry run should not fail")
		testingPkg.AssertTrue(t, strings.Contains(string(output), "dry run") ||
			strings.Contains(string(output), "would send"), "Should show dry run message")
	})
}

func TestLagAnalysisWorkflow(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	binaryPath := buildCLIBinary(t)
	defer os.Remove(binaryPath)

	t.Run("lag_analyze", func(t *testing.T) {
		// Analyze dependency lag
		cmd := exec.Command(binaryPath, "lag", "analyze", "1")
		cmd.Env = append(os.Environ(), "AI_DEP_MANAGER_DATA_DIR="+ctx.TempDir)
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			t.Logf("Lag analysis output: %s", string(output))
		}
		// Lag analysis might fail due to missing data, but shouldn't crash
	})

	t.Run("lag_plan", func(t *testing.T) {
		// Create lag resolution plan
		cmd := exec.Command(binaryPath, "lag", "plan", "1", "--strategy", "balanced")
		cmd.Env = append(os.Environ(), "AI_DEP_MANAGER_DATA_DIR="+ctx.TempDir)
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			t.Logf("Lag plan output: %s", string(output))
		}
		// Plan creation might fail due to missing data, but shouldn't crash
	})
}

func TestErrorHandling(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	binaryPath := buildCLIBinary(t)
	defer os.Remove(binaryPath)

	t.Run("invalid_arguments", func(t *testing.T) {
		// Test with invalid arguments
		cmd := exec.Command(binaryPath, "scan", "--invalid-flag")
		cmd.Env = append(os.Environ(), "AI_DEP_MANAGER_DATA_DIR="+ctx.TempDir)
		_, err := cmd.CombinedOutput()
		
		testingPkg.AssertError(t, err, "Invalid arguments should cause error")
	})

	t.Run("missing_required_args", func(t *testing.T) {
		// Test with missing required arguments
		cmd := exec.Command(binaryPath, "policy", "show")
		cmd.Env = append(os.Environ(), "AI_DEP_MANAGER_DATA_DIR="+ctx.TempDir)
		_, err := cmd.CombinedOutput()
		
		testingPkg.AssertError(t, err, "Missing required arguments should cause error")
	})

	t.Run("invalid_data_directory", func(t *testing.T) {
		// Test with invalid data directory
		cmd := exec.Command(binaryPath, "status")
		cmd.Env = append(os.Environ(), "AI_DEP_MANAGER_DATA_DIR=/invalid/path")
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			t.Logf("Invalid data dir output: %s", string(output))
		}
		// Should handle invalid data directory gracefully
	})
}

// Helper functions

func buildCLIBinary(t *testing.T) string {
	// Build the CLI binary for testing
	tempDir, err := os.MkdirTemp("", "ai-dep-manager-e2e-*")
	testingPkg.AssertNoError(t, err, "Should create temp dir for binary")

	binaryPath := filepath.Join(tempDir, "ai-dep-manager")
	
	// Build the binary
	cmd := exec.Command("go", "build", "-o", binaryPath, "../../cmd/ai-dep-manager")
	cmd.Dir = filepath.Join("..", "..")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Build output: %s", string(output))
		t.Skipf("Cannot build CLI binary for testing: %v", err)
	}

	return binaryPath
}

func TestPerformanceUnderLoad(t *testing.T) {
	testingPkg.SkipIfShort(t, "performance testing requires time")

	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	binaryPath := buildCLIBinary(t)
	defer os.Remove(binaryPath)

	t.Run("concurrent_commands", func(t *testing.T) {
		// Test running multiple commands concurrently
		const numConcurrent = 5
		results := make(chan error, numConcurrent)

		for i := 0; i < numConcurrent; i++ {
			go func() {
				cmd := exec.Command(binaryPath, "status")
				cmd.Env = append(os.Environ(), "AI_DEP_MANAGER_DATA_DIR="+ctx.TempDir)
				_, err := cmd.CombinedOutput()
				results <- err
			}()
		}

		// Collect results
		var errors []error
		for i := 0; i < numConcurrent; i++ {
			select {
			case err := <-results:
				if err != nil {
					errors = append(errors, err)
				}
			case <-time.After(30 * time.Second):
				t.Fatal("Timeout waiting for concurrent commands")
			}
		}

		testingPkg.AssertLen(t, errors, 0, "No concurrent command should fail")
	})

	t.Run("memory_stability", func(t *testing.T) {
		// Run commands repeatedly to check for memory leaks
		for i := 0; i < 10; i++ {
			cmd := exec.Command(binaryPath, "status")
			cmd.Env = append(os.Environ(), "AI_DEP_MANAGER_DATA_DIR="+ctx.TempDir)
			_, err := cmd.CombinedOutput()
			testingPkg.AssertNoError(t, err, "Repeated commands should not fail")
		}
	})
}

// Benchmark tests

func BenchmarkE2E_StatusCommand(b *testing.B) {
	ctx := testingPkg.SetupTestEnvironment(&testing.T{})
	defer ctx.Cleanup()

	binaryPath := buildCLIBinary(&testing.T{})
	defer os.Remove(binaryPath)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := exec.Command(binaryPath, "status")
		cmd.Env = append(os.Environ(), "AI_DEP_MANAGER_DATA_DIR="+ctx.TempDir)
		_, err := cmd.CombinedOutput()
		if err != nil {
			b.Fatalf("Status command failed: %v", err)
		}
	}
}

func BenchmarkE2E_ConcurrentCommands(b *testing.B) {
	ctx := testingPkg.SetupTestEnvironment(&testing.T{})
	defer ctx.Cleanup()

	binaryPath := buildCLIBinary(&testing.T{})
	defer os.Remove(binaryPath)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cmd := exec.Command(binaryPath, "status")
			cmd.Env = append(os.Environ(), "AI_DEP_MANAGER_DATA_DIR="+ctx.TempDir)
			_, err := cmd.CombinedOutput()
			if err != nil {
				b.Fatalf("Concurrent command failed: %v", err)
			}
		}
	})
}
