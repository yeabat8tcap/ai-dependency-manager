package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/8tcapital/superint-dep-manager/internal/superint"
	aitypes "github.com/8tcapital/superint-dep-manager/internal/superint/types"
	"github.com/8tcapital/superint-dep-manager/internal/logger"
	"github.com/8tcapital/superint-dep-manager/internal/services"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update dependencies with Superintelligence-powered recommendations",
	Long: `Update project dependencies with intelligent analysis and safety checks. This command:
- Generates update plans with risk assessment
- Provides Superintelligence-powered recommendations
- Supports interactive and batch update modes
- Creates rollback plans for safety
- Handles breaking changes with user confirmation

Examples:
  superint-dep-manager update --preview                    # Preview all updates
  superint-dep-manager update --project my-app             # Update specific project
  superint-dep-manager update --security-only              # Apply only security updates
  superint-dep-manager update --interactive                # Interactive update mode`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUpdate(cmd, args)
	},
}

var (
	updateProject      string
	updateProjectID    uint
	updateDependencies []string
	updateTypes        []string
	riskLevels         []string
	preview            bool
	dryRun             bool
	force              bool
	updateInteractive  bool
	autoApprove        bool
	skipBreaking       bool
	updateSecurityOnly bool
	batchSize          int
	updateTimeout      time.Duration
)

func runUpdate(cmd *cobra.Command, args []string) error {
	projectService := services.NewProjectService()
	updateService := services.NewUpdateService()
	
	// Determine project ID
	projectID := updateProjectID
	if updateProject != "" {
		project, err := projectService.GetProjectByName(cmd.Context(), updateProject)
		if err != nil {
			return fmt.Errorf("failed to find project '%s': %w", updateProject, err)
		}
		projectID = project.ID
	}
	
	// If no specific project, handle all projects
	if projectID == 0 {
		return runUpdateAllProjects(cmd, projectService, updateService)
	}
	
	return runUpdateProject(cmd, projectService, updateService, projectID)
}

func runUpdateProject(cmd *cobra.Command, projectService *services.ProjectService, updateService *services.UpdateService, projectID uint) error {
	// Get project details
	project, err := projectService.GetProject(cmd.Context(), projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}
	
	fmt.Printf("🔄 Updating project: %s (%s)\n", project.Name, project.Type)
	fmt.Printf("Path: %s\n\n", project.Path)
	
	// Create update options
	options := &services.UpdateOptions{
		ProjectID:       projectID,
		DependencyNames: updateDependencies,
		UpdateTypes:     updateTypes,
		RiskLevels:      riskLevels,
		DryRun:          dryRun || preview,
		Force:           force,
		Interactive:     interactive,
		AutoApprove:     autoApprove,
		SkipBreaking:    skipBreaking,
		SecurityOnly:    securityOnly,
		BatchSize:       batchSize,
		Timeout:         updateTimeout,
	}
	
	// Generate update plan
	fmt.Println("📋 Generating update plan...")
	plan, err := updateService.GenerateUpdatePlan(cmd.Context(), options)
	if err != nil {
		return fmt.Errorf("failed to generate update plan: %w", err)
	}
	
	// Display plan
	displayUpdatePlan(plan)
	
	// If preview only, stop here
	if preview {
		return nil
	}
	
	// Get user confirmation if interactive
	if interactive && !autoApprove {
		if !getUserConfirmation(plan) {
			fmt.Println("❌ Update cancelled by user")
			return nil
		}
	}
	
	// Apply updates
	fmt.Println("\n🚀 Applying updates...")
	result, err := updateService.ApplyUpdates(cmd.Context(), plan, options)
	if err != nil {
		return fmt.Errorf("failed to apply updates: %w", err)
	}
	
	// Display results
	displayUpdateResult(result)
	
	return nil
}

func runUpdateAllProjects(cmd *cobra.Command, projectService *services.ProjectService, updateService *services.UpdateService) error {
	fmt.Println("🔄 Updating all enabled projects...")
	
	// Get all enabled projects
	enabled := true
	projects, err := projectService.ListProjects(cmd.Context(), &enabled)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}
	
	if len(projects) == 0 {
		fmt.Println("❌ No enabled projects found")
		return nil
	}
	
	fmt.Printf("Found %d enabled project(s)\n\n", len(projects))
	
	successCount := 0
	errorCount := 0
	
	for _, project := range projects {
		fmt.Printf("📦 Processing project: %s\n", project.Name)
		err := runUpdateProject(cmd, projectService, updateService, project.ID)
		if err != nil {
			fmt.Printf("❌ Failed to update %s: %v\n\n", project.Name, err)
			errorCount++
		} else {
			fmt.Printf("✅ Successfully processed %s\n\n", project.Name)
			successCount++
		}
	}
	
	// Summary
	fmt.Println("📊 Update Summary:")
	fmt.Printf("   Projects processed: %d\n", len(projects))
	fmt.Printf("   Successful: %d\n", successCount)
	fmt.Printf("   Failed: %d\n", errorCount)
	
	return nil
}

func displayUpdatePlan(plan *services.UpdatePlan) {
	fmt.Printf("📋 Update Plan for %s\n", plan.ProjectName)
	fmt.Println(strings.Repeat("=", 50))
	
	if plan.TotalUpdates == 0 {
		fmt.Println("✅ No updates available")
		return
	}
	
	// Risk summary
	fmt.Printf("📊 Risk Summary:\n")
	fmt.Printf("   Total updates: %d\n", plan.RiskSummary.TotalUpdates)
	fmt.Printf("   🟢 Low risk: %d\n", plan.RiskSummary.LowRisk)
	fmt.Printf("   🟡 Medium risk: %d\n", plan.RiskSummary.MediumRisk)
	fmt.Printf("   🟠 High risk: %d\n", plan.RiskSummary.HighRisk)
	fmt.Printf("   🔴 Critical risk: %d\n", plan.RiskSummary.CriticalRisk)
	
	if plan.RiskSummary.BreakingChanges > 0 {
		fmt.Printf("   ⚠️  Breaking changes: %d\n", plan.RiskSummary.BreakingChanges)
	}
	
	if plan.RiskSummary.SecurityUpdates > 0 {
		fmt.Printf("   🔒 Security updates: %d\n", plan.RiskSummary.SecurityUpdates)
	}
	
	fmt.Printf("   Overall risk: %s\n", getRiskIcon(plan.RiskSummary.OverallRisk))
	fmt.Printf("   Estimated time: %s\n\n", plan.EstimatedTime.Round(time.Second))
	
	// Update groups
	fmt.Println("📦 Update Groups:")
	for i, group := range plan.UpdateGroups {
		fmt.Printf("\n%d. %s (%s priority, %s risk)\n", 
			i+1, strings.Title(group.Name), group.Priority, group.RiskLevel)
		fmt.Printf("   %s\n", group.Description)
		fmt.Printf("   Updates: %d\n", len(group.Updates))
		
		if len(group.Updates) <= 5 {
			// Show all updates if 5 or fewer
			for _, update := range group.Updates {
				fmt.Printf("   %s %s: %s → %s", 
					getUpdateIcon(update), update.DependencyName, 
					update.FromVersion, update.ToVersion)
				
				if update.BreakingChange {
					fmt.Print(" ⚠️")
				}
				if update.SecurityFix {
					fmt.Print(" 🔒")
				}
				fmt.Printf(" (%.0f%% confidence)\n", update.Confidence*100)
			}
		} else {
			// Show first 3 and summary for large groups
			for _, update := range group.Updates[:3] {
				fmt.Printf("   %s %s: %s → %s", 
					getUpdateIcon(update), update.DependencyName, 
					update.FromVersion, update.ToVersion)
				
				if update.BreakingChange {
					fmt.Print(" ⚠️")
				}
				if update.SecurityFix {
					fmt.Print(" 🔒")
				}
				fmt.Printf(" (%.0f%% confidence)\n", update.Confidence*100)
			}
			fmt.Printf("   ... and %d more updates\n", len(group.Updates)-3)
		}
	}
	
	// Recommendations
	if len(plan.Recommendations) > 0 {
		fmt.Println("\n💡 Recommendations:")
		for _, rec := range plan.Recommendations {
			fmt.Printf("   • %s\n", rec)
		}
	}
	
	// Warnings
	if len(plan.Warnings) > 0 {
		fmt.Println("\n⚠️  Warnings:")
		for _, warning := range plan.Warnings {
			fmt.Printf("   • %s\n", warning)
		}
	}
}

func displayUpdateResult(result *services.UpdateResult) {
	fmt.Printf("\n📊 Update Results for %s\n", result.ProjectName)
	fmt.Println(strings.Repeat("=", 50))
	
	fmt.Printf("⏱️  Duration: %s\n", result.Duration.Round(time.Second))
	fmt.Printf("📦 Total attempted: %d\n", result.TotalAttempted)
	fmt.Printf("✅ Successful: %d\n", len(result.Successful))
	fmt.Printf("❌ Failed: %d\n", len(result.Failed))
	fmt.Printf("⏭️  Skipped: %d\n", len(result.Skipped))
	
	// Show successful updates
	if len(result.Successful) > 0 {
		fmt.Println("\n✅ Successfully Updated:")
		for _, update := range result.Successful {
			fmt.Printf("   %s %s: %s → %s\n", 
				getUpdateIcon(update), update.DependencyName, 
				update.FromVersion, update.ToVersion)
		}
	}
	
	// Show failed updates
	if len(result.Failed) > 0 {
		fmt.Println("\n❌ Failed Updates:")
		for _, failure := range result.Failed {
			fmt.Printf("   %s %s: %s → %s\n", 
				getUpdateIcon(failure.UpdateItem), failure.UpdateItem.DependencyName, 
				failure.UpdateItem.FromVersion, failure.UpdateItem.ToVersion)
			fmt.Printf("      Error: %s\n", failure.Error)
		}
	}
	
	// Show skipped updates
	if len(result.Skipped) > 0 {
		fmt.Println("\n⏭️  Skipped Updates:")
		for _, update := range result.Skipped {
			fmt.Printf("   %s %s: %s → %s\n", 
				getUpdateIcon(update), update.DependencyName, 
				update.FromVersion, update.ToVersion)
		}
	}
	
	// Show rollback plan if available
	if result.RollbackPlan != nil && len(result.RollbackPlan.Rollbacks) > 0 {
		fmt.Println("\n🔄 Rollback Plan Available:")
		fmt.Printf("   Created: %s\n", result.RollbackPlan.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("   Rollback operations: %d\n", len(result.RollbackPlan.Rollbacks))
		fmt.Println("   💡 Use 'superint-dep-manager rollback' to revert changes if needed")
	}
	
	// Final status
	if len(result.Failed) == 0 {
		fmt.Println("\n🎉 All updates completed successfully!")
	} else if len(result.Successful) > 0 {
		fmt.Println("\n⚠️  Updates completed with some failures")
		fmt.Println("💡 Review failed updates and consider manual intervention")
	} else {
		fmt.Println("\n❌ No updates were applied successfully")
	}
}

func getUserConfirmation(plan *services.UpdatePlan) bool {
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Printf("\n❓ Do you want to proceed with %d update(s)? ", plan.TotalUpdates)
	
	// Show risk warning if high risk
	if plan.RiskSummary.OverallRisk == aitypes.RiskLevelHigh || plan.RiskSummary.OverallRisk == aitypes.RiskLevelCritical {
		fmt.Printf("(%s risk) ", plan.RiskSummary.OverallRisk)
	}
	
	fmt.Print("[y/N]: ")
	
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

func getRiskIcon(risk superint.RiskLevel) string {
	switch risk {
	case aitypes.RiskLevelLow:
		return "🟢 Low"
	case aitypes.RiskLevelMedium:
		return "🟡 Medium"
	case aitypes.RiskLevelHigh:
		return "🟠 High"
	case aitypes.RiskLevelCritical:
		return "🔴 Critical"
	default:
		return "❓ Unknown"
	}
}

func getUpdateIcon(update services.UpdateItem) string {
	if update.SecurityFix {
		return "🔒"
	}
	if update.BreakingChange {
		return "⚠️"
	}
	
	switch update.RiskLevel {
	case aitypes.RiskLevelCritical:
		return "🔴"
	case aitypes.RiskLevelHigh:
		return "🟠"
	case aitypes.RiskLevelMedium:
		return "🟡"
	default:
		return "🟢"
	}
}

func init() {
	rootCmd.AddCommand(updateCmd)
	
	updateCmd.Flags().StringVar(&updateProject, "project", "", "Update specific project by name")
	updateCmd.Flags().UintVar(&updateProjectID, "project-id", 0, "Update specific project by ID")
	updateCmd.Flags().StringSliceVar(&updateDependencies, "dependencies", []string{}, "Update specific dependencies (comma-separated)")
	updateCmd.Flags().StringSliceVar(&updateTypes, "types", []string{}, "Update types to include: major, minor, patch, security")
	updateCmd.Flags().StringSliceVar(&riskLevels, "risk-levels", []string{}, "Risk levels to include: low, medium, high, critical")
	updateCmd.Flags().BoolVar(&preview, "preview", false, "Preview updates without applying them")
	updateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate updates without making changes")
	updateCmd.Flags().BoolVar(&force, "force", false, "Continue updating even if some updates fail")
	updateCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive update mode with confirmations")
	updateCmd.Flags().BoolVar(&autoApprove, "auto-approve", false, "Automatically approve all updates (use with caution)")
	updateCmd.Flags().BoolVar(&skipBreaking, "skip-breaking", false, "Skip updates with breaking changes")
	updateCmd.Flags().BoolVar(&securityOnly, "security-only", false, "Apply only security updates")
	updateCmd.Flags().IntVar(&batchSize, "batch-size", 10, "Number of updates to process in each batch")
	updateCmd.Flags().DurationVar(&updateTimeout, "timeout", 10*time.Minute, "Timeout for update operations")
	
	// Add validation
	updateCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		// Validate update types
		validTypes := map[string]bool{
			"major": true, "minor": true, "patch": true, "security": true,
		}
		for _, updateType := range updateTypes {
			if !validTypes[updateType] {
				return fmt.Errorf("invalid update type '%s'. Valid types: major, minor, patch, security", updateType)
			}
		}
		
		// Validate risk levels
		validRiskLevels := map[string]bool{
			"low": true, "medium": true, "high": true, "critical": true,
		}
		for _, riskLevel := range riskLevels {
			if !validRiskLevels[riskLevel] {
				return fmt.Errorf("invalid risk level '%s'. Valid levels: low, medium, high, critical", riskLevel)
			}
		}
		
		// Parse project ID if provided as argument
		if len(args) > 0 {
			if id, err := strconv.ParseUint(args[0], 10, 32); err == nil {
				updateProjectID = uint(id)
			} else {
				updateProject = args[0]
			}
		}
		
		// Validate conflicting options
		if autoApprove && interactive {
			return fmt.Errorf("cannot use both --auto-approve and --interactive")
		}
		
		if preview && (dryRun || force || autoApprove) {
			logger.Warn("Preview mode ignores --dry-run, --force, and --auto-approve flags")
		}
		
		return nil
	}
}
