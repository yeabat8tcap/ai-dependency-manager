package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"


	"github.com/8tcapital/ai-dep-manager/internal/services"
	"github.com/spf13/cobra"
)

// rollbackCmd represents the rollback command
var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback previous dependency updates",
	Long: `Rollback dependency updates to previous versions using stored rollback plans.
This command helps you safely revert changes if updates cause issues.

Examples:
  ai-dep-manager rollback --project my-app        # Rollback latest updates for project
  ai-dep-manager rollback --list                  # List available rollback plans
  ai-dep-manager rollback --plan-id 123           # Execute specific rollback plan`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRollback(cmd, args)
	},
}

var (
	rollbackProject   string
	rollbackProjectID uint
	rollbackPlanID    uint
	listPlans         bool
	rollbackForce     bool
	rollbackDryRun    bool
)

func runRollback(cmd *cobra.Command, args []string) error {
	rollbackService := services.NewRollbackService()
	
	// List rollback plans if requested
	if listPlans {
		return listRollbackPlans(cmd, rollbackService)
	}
	
	// Execute specific rollback plan
	if rollbackPlanID != 0 {
		return rollbackSpecificPlan(cmd, rollbackService, rollbackPlanID)
	}
	
	// Rollback latest updates for project
	projectService := services.NewProjectService()
	
	// Determine project ID
	projectID := rollbackProjectID
	if rollbackProject != "" {
		project, err := projectService.GetProjectByName(cmd.Context(), rollbackProject)
		if err != nil {
			return fmt.Errorf("failed to find project '%s': %w", rollbackProject, err)
		}
		projectID = project.ID
	}
	
	if projectID == 0 {
		return fmt.Errorf("project ID or name is required")
	}
	
	return rollbackLatestUpdates(cmd, rollbackService, projectID)
}

func listRollbackPlans(cmd *cobra.Command, rollbackService *services.RollbackService) error {
	fmt.Println("📋 Available Rollback Plans")
	fmt.Println(strings.Repeat("=", 50))
	
	plans, err := rollbackService.ListRollbackPlans(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to list rollback plans: %w", err)
	}
	
	if len(plans) == 0 {
		fmt.Println("❌ No rollback plans available")
		return nil
	}
	
	for _, plan := range plans {
		fmt.Printf("\n🔄 Plan for Project: %s (ID: %d)\n", plan.ProjectName, plan.ProjectID)
		fmt.Printf("   Status: %s\n", plan.Status)
		fmt.Printf("   Created: %s\n", plan.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("   Rollbacks: %d operations\n", len(plan.Rollbacks))
		fmt.Printf("   Status: %s\n", plan.Status)
		
		if len(plan.Rollbacks) > 0 {
			fmt.Println("   Operations:")
			for i, rollback := range plan.Rollbacks {
				if i < 3 { // Show first 3
					fmt.Printf("     • %s: %s → %s\n", 
						rollback.DependencyName, rollback.FromVersion, rollback.ToVersion)
				} else if i == 3 {
					fmt.Printf("     ... and %d more\n", len(plan.Rollbacks)-3)
					break
				}
			}
		}
	}
	
	return nil
}

func rollbackSpecificPlan(cmd *cobra.Command, rollbackService *services.RollbackService, planID uint) error {
	fmt.Printf("🔄 Executing rollback plan ID: %d\n", planID)
	
	// Get rollback plan
	// Note: Using ProjectID as plan identifier since RollbackPlan doesn't have ID field
	plan, err := rollbackService.GetRollbackPlan(cmd.Context(), planID)
	if err != nil {
		return fmt.Errorf("failed to get rollback plan: %w", err)
	}
	
	// Display plan details
	fmt.Printf("📋 Rollback Plan Details:\n")
	fmt.Printf("   Project: %s (ID: %d)\n", plan.ProjectName, plan.ProjectID)
	fmt.Printf("   Created: %s\n", plan.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("   Operations: %d\n", len(plan.Rollbacks))
	
	if rollbackDryRun {
		fmt.Println("\n🔍 DRY RUN - Operations to be performed:")
	} else {
		fmt.Println("\n📦 Operations to be performed:")
	}
	
	for _, rollback := range plan.Rollbacks {
		fmt.Printf("   • %s: %s → %s\n", 
			rollback.DependencyName, rollback.FromVersion, rollback.ToVersion)
	}
	
	// Get user confirmation unless force is enabled
	if !rollbackForce && !rollbackDryRun {
		if !getRollbackConfirmation(len(plan.Rollbacks)) {
			fmt.Println("❌ Rollback cancelled by user")
			return nil
		}
	}
	
	// Execute rollback
	options := &services.RollbackOptions{
		DryRun: rollbackDryRun,
		Force:  rollbackForce,
	}
	
	result, err := rollbackService.ExecuteRollback(cmd.Context(), planID, options)
	if err != nil {
		return fmt.Errorf("failed to execute rollback: %w", err)
	}
	
	// Display results
	displayRollbackResult(result)
	
	return nil
}

func rollbackLatestUpdates(cmd *cobra.Command, rollbackService *services.RollbackService, projectID uint) error {
	fmt.Printf("🔄 Rolling back latest updates for project ID: %d\n", projectID)
	
	// Find latest rollback plan for project
	plan, err := rollbackService.GetLatestRollbackPlan(cmd.Context(), projectID)
	if err != nil {
		return fmt.Errorf("failed to find rollback plan: %w", err)
	}
	
	if plan == nil {
		fmt.Println("❌ No rollback plan available for this project")
		fmt.Println("💡 Rollback plans are created when updates are applied")
		return nil
	}
	
	return rollbackSpecificPlan(cmd, rollbackService, plan.ProjectID)
}

func displayRollbackResult(result *services.RollbackResult) {
	fmt.Printf("\n📊 Rollback Results for %s\n", result.ProjectName)
	fmt.Println(strings.Repeat("=", 50))
	
	fmt.Printf("⏱️  Duration: %s\n", result.Duration.Round(time.Second))
	fmt.Printf("📦 Total attempted: %d\n", result.TotalAttempted)
	fmt.Printf("✅ Successful: %d\n", len(result.Successful))
	fmt.Printf("❌ Failed: %d\n", len(result.Failed))
	
	// Show successful rollbacks
	if len(result.Successful) > 0 {
		fmt.Println("\n✅ Successfully Rolled Back:")
		for _, rollback := range result.Successful {
			fmt.Printf("   🔄 %s: %s → %s\n", 
				rollback.DependencyName, rollback.FromVersion, rollback.ToVersion)
		}
	}
	
	// Show failed rollbacks
	if len(result.Failed) > 0 {
		fmt.Println("\n❌ Failed Rollbacks:")
		for _, failure := range result.Failed {
			fmt.Printf("   🔄 %s: %s → %s\n", 
				failure.RollbackItem.DependencyName, 
				failure.RollbackItem.FromVersion, 
				failure.RollbackItem.ToVersion)
			fmt.Printf("      Error: %s\n", failure.Error)
		}
	}
	
	// Final status
	if len(result.Failed) == 0 {
		fmt.Println("\n🎉 All rollbacks completed successfully!")
		fmt.Println("💡 Your project has been restored to the previous state")
	} else if len(result.Successful) > 0 {
		fmt.Println("\n⚠️  Rollback completed with some failures")
		fmt.Println("💡 Review failed rollbacks and consider manual intervention")
	} else {
		fmt.Println("\n❌ No rollbacks were completed successfully")
		fmt.Println("💡 Your project state remains unchanged")
	}
}

func getRollbackConfirmation(operationCount int) bool {
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Printf("\n❓ Do you want to proceed with rolling back %d operation(s)? ", operationCount)
	fmt.Print("This will revert your dependencies to previous versions. [y/N]: ")
	
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

func init() {
	rootCmd.AddCommand(rollbackCmd)
	
	rollbackCmd.Flags().StringVar(&rollbackProject, "project", "", "Rollback updates for specific project by name")
	rollbackCmd.Flags().UintVar(&rollbackProjectID, "project-id", 0, "Rollback updates for specific project by ID")
	rollbackCmd.Flags().UintVar(&rollbackPlanID, "plan-id", 0, "Execute specific rollback plan by ID")
	rollbackCmd.Flags().BoolVar(&listPlans, "list", false, "List available rollback plans")
	rollbackCmd.Flags().BoolVar(&rollbackForce, "force", false, "Force rollback without confirmation")
	rollbackCmd.Flags().BoolVar(&rollbackDryRun, "dry-run", false, "Show what would be rolled back without making changes")
	
	// Add validation
	rollbackCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		// Count of mutually exclusive options
		optionCount := 0
		if listPlans {
			optionCount++
		}
		if rollbackPlanID != 0 {
			optionCount++
		}
		if rollbackProject != "" || rollbackProjectID != 0 {
			optionCount++
		}
		
		// If no specific options provided, require project specification
		if optionCount == 0 {
			if len(args) == 0 {
				return fmt.Errorf("must specify --list, --plan-id, or --project/--project-id")
			}
			// Try to parse first argument as project name or ID
			rollbackProject = args[0]
		}
		
		return nil
	}
}
