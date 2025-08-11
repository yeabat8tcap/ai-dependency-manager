package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/8tcapital/ai-dep-manager/internal/logger"
	"github.com/8tcapital/ai-dep-manager/internal/services"
	"github.com/spf13/cobra"
)

var lagCmd = &cobra.Command{
	Use:   "lag",
	Short: "Analyze and resolve dependency lag",
	Long:  `Analyze dependency lag across projects and create resolution plans to bring dependencies up to date.`,
}

var analyzeLagCmd = &cobra.Command{
	Use:   "analyze [project-id]",
	Short: "Analyze dependency lag",
	Long:  `Analyze dependency lag for a specific project or all projects. Shows outdated packages, lag distribution, and recommendations.`,
	Args:  cobra.MaximumNArgs(1),
	Run:   runAnalyzeLag,
}

var planLagCmd = &cobra.Command{
	Use:   "plan [project-id]",
	Short: "Create lag resolution plan",
	Long:  `Create a comprehensive plan to resolve dependency lag for a project.`,
	Args:  cobra.ExactArgs(1),
	Run:   runPlanLag,
}

var executeLagCmd = &cobra.Command{
	Use:   "execute [project-id]",
	Short: "Execute lag resolution plan",
	Long:  `Execute a previously created lag resolution plan for a project.`,
	Args:  cobra.ExactArgs(1),
	Run:   runExecuteLag,
}

// Global variables for lag command flags
var (
	lagStrategy string
	lagDryRun   bool
	lagFormat   string
	lagLimit    int
)

func init() {
	rootCmd.AddCommand(lagCmd)
	lagCmd.AddCommand(analyzeLagCmd)
	lagCmd.AddCommand(planLagCmd)
	lagCmd.AddCommand(executeLagCmd)

	// Analyze lag flags
	analyzeLagCmd.Flags().StringVar(&lagFormat, "format", "table", "Output format (table, json)")
	analyzeLagCmd.Flags().IntVar(&lagLimit, "limit", 20, "Limit number of lagged packages to show")

	// Plan lag flags
	planLagCmd.Flags().StringVar(&lagStrategy, "strategy", "balanced", "Resolution strategy (conservative, balanced, aggressive)")
	planLagCmd.Flags().StringVar(&lagFormat, "format", "table", "Output format (table, json)")

	// Execute lag flags
	executeLagCmd.Flags().BoolVar(&lagDryRun, "dry-run", false, "Show what would be executed without actually running")
	executeLagCmd.Flags().StringVar(&lagStrategy, "strategy", "balanced", "Resolution strategy (conservative, balanced, aggressive)")
}

func runAnalyzeLag(cmd *cobra.Command, args []string) {
	var projectID *uint
	
	if len(args) > 0 {
		id, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			logger.Error("Invalid project ID: %v", err)
			os.Exit(1)
		}
		projectIDVal := uint(id)
		projectID = &projectIDVal
	}
	
	lagService := services.NewLagService()
	
	logger.Info("Analyzing dependency lag...")
	analysis, err := lagService.AnalyzeLag(context.Background(), projectID)
	if err != nil {
		logger.Error("Failed to analyze lag: %v", err)
		os.Exit(1)
	}
	
	if lagFormat == "json" {
		printJSON(analysis)
		return
	}
	
	// Display analysis in table format
	displayLagAnalysis(analysis)
}

func runPlanLag(cmd *cobra.Command, args []string) {
	projectID, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		logger.Error("Invalid project ID: %v", err)
		os.Exit(1)
	}
	
	// Validate strategy
	validStrategies := []string{"conservative", "balanced", "aggressive"}
	if !contains(validStrategies, lagStrategy) {
		logger.Error("Invalid strategy: %s. Valid strategies: %s", lagStrategy, strings.Join(validStrategies, ", "))
		os.Exit(1)
	}
	
	lagService := services.NewLagService()
	
	logger.Info("Creating lag resolution plan for project %d with %s strategy...", projectID, lagStrategy)
	plan, err := lagService.CreateResolutionPlan(context.Background(), uint(projectID), lagStrategy)
	if err != nil {
		logger.Error("Failed to create resolution plan: %v", err)
		os.Exit(1)
	}
	
	if lagFormat == "json" {
		printJSON(plan)
		return
	}
	
	// Display plan in table format
	displayResolutionPlan(plan)
}

func runExecuteLag(cmd *cobra.Command, args []string) {
	projectID, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		logger.Error("Invalid project ID: %v", err)
		os.Exit(1)
	}
	
	lagService := services.NewLagService()
	
	// First create the plan
	logger.Info("Creating resolution plan...")
	plan, err := lagService.CreateResolutionPlan(context.Background(), uint(projectID), lagStrategy)
	if err != nil {
		logger.Error("Failed to create resolution plan: %v", err)
		os.Exit(1)
	}
	
	if len(plan.Phases) == 0 {
		fmt.Println("✅ No lag resolution needed - all dependencies are up to date!")
		return
	}
	
	// Show plan summary
	fmt.Printf("📋 Lag Resolution Plan for Project %d\n", projectID)
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Strategy: %s\n", lagStrategy)
	fmt.Printf("Total Packages: %d\n", plan.TotalPackages)
	fmt.Printf("Estimated Time: %s\n", plan.EstimatedTime)
	fmt.Printf("Risk Level: %s\n", plan.RiskLevel)
	fmt.Printf("Phases: %d\n", len(plan.Phases))
	fmt.Println()
	
	// Show prerequisites
	if len(plan.Prerequisites) > 0 {
		fmt.Println("📋 Prerequisites:")
		for _, prereq := range plan.Prerequisites {
			fmt.Printf("  - %s\n", prereq)
		}
		fmt.Println()
	}
	
	// Show phases summary
	fmt.Println("📊 Execution Phases:")
	for i, phase := range plan.Phases {
		fmt.Printf("  Phase %d: %s\n", i+1, phase.Name)
		fmt.Printf("    Packages: %d\n", len(phase.Packages))
		fmt.Printf("    Risk: %s\n", phase.RiskLevel)
		fmt.Printf("    Duration: %s\n", phase.Duration)
	}
	fmt.Println()
	
	if lagDryRun {
		fmt.Println("🔍 Dry run mode - showing execution plan without making changes")
		fmt.Println()
	} else {
		// Confirm execution
		fmt.Print("Do you want to proceed with the execution? (y/N): ")
		var response string
		fmt.Scanln(&response)
		
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Execution cancelled.")
			return
		}
		fmt.Println()
	}
	
	// Execute the plan
	logger.Info("Executing lag resolution plan...")
	if err := lagService.ExecuteResolutionPlan(context.Background(), plan, lagDryRun); err != nil {
		logger.Error("Failed to execute resolution plan: %v", err)
		os.Exit(1)
	}
	
	if lagDryRun {
		fmt.Println("✅ Dry run completed successfully")
	} else {
		fmt.Println("✅ Lag resolution plan executed successfully")
		fmt.Println("💡 Run 'ai-dep-manager scan' to verify the updates")
	}
}

// Display functions

func displayLagAnalysis(analysis *services.LagAnalysis) {
	fmt.Println("📊 Dependency Lag Analysis")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()
	
	// Project info
	if analysis.ProjectName != "" {
		fmt.Printf("Project: %s (ID: %d)\n", analysis.ProjectName, analysis.ProjectID)
		fmt.Println()
	}
	
	// Summary metrics
	fmt.Println("📈 Summary:")
	fmt.Printf("  Total Dependencies:  %d\n", analysis.TotalDependencies)
	fmt.Printf("  Lagged Dependencies: %d\n", analysis.LaggedDependencies)
	fmt.Printf("  Average Lag:         %.1f days\n", analysis.AverageLagDays)
	fmt.Printf("  Maximum Lag:         %d days\n", analysis.MaxLagDays)
	fmt.Println()
	
	// Lag distribution
	if len(analysis.LagDistribution) > 0 {
		fmt.Println("📊 Lag Distribution:")
		for category, count := range analysis.LagDistribution {
			percentage := float64(count) / float64(analysis.TotalDependencies) * 100
			fmt.Printf("  %-12s: %3d (%.1f%%)\n", strings.Title(category), count, percentage)
		}
		fmt.Println()
	}
	
	// Top lagged packages
	if len(analysis.TopLaggedPackages) > 0 {
		fmt.Printf("🔍 Top Lagged Packages (showing %d):\n", len(analysis.TopLaggedPackages))
		fmt.Printf("%-30s %-15s %-15s %-8s %-8s %s\n", 
			"Package", "Current", "Latest", "Lag", "Risk", "Reason")
		fmt.Println(strings.Repeat("-", 100))
		
		for _, pkg := range analysis.TopLaggedPackages {
			riskIcons := ""
			if pkg.SecurityRisk {
				riskIcons += "🚨"
			}
			if pkg.BreakingRisk {
				riskIcons += "⚠️"
			}
			if riskIcons == "" {
				riskIcons = "✅"
			}
			
			fmt.Printf("%-30s %-15s %-15s %-8s %-8s %s\n",
				truncateString(pkg.Name, 30),
				truncateString(pkg.CurrentVersion, 15),
				truncateString(pkg.LatestVersion, 15),
				fmt.Sprintf("%dd", pkg.LagDays),
				riskIcons,
				truncateString(pkg.Reason, 40),
			)
		}
		fmt.Println()
	}
	
	// Recommendations
	if len(analysis.RecommendedActions) > 0 {
		fmt.Println("💡 Recommendations:")
		for i, rec := range analysis.RecommendedActions {
			priorityIcon := getPriorityIcon(rec.Priority)
			fmt.Printf("%d. %s %s\n", i+1, priorityIcon, rec.Description)
			fmt.Printf("   Action: %s\n", rec.Action)
			fmt.Printf("   Impact: %s | Effort: %s\n", rec.Impact, rec.Effort)
			if len(rec.Packages) > 0 && rec.Packages[0] != "multiple" {
				fmt.Printf("   Packages: %s\n", strings.Join(rec.Packages, ", "))
			}
			fmt.Println()
		}
	}
	
	// Lag trend
	if len(analysis.LagTrend) > 0 {
		fmt.Println("📈 Lag Trend (Last 30 days):")
		fmt.Printf("%-12s %-10s %-8s %s\n", "Date", "Avg Lag", "Max Lag", "Count")
		fmt.Println(strings.Repeat("-", 40))
		
		for _, point := range analysis.LagTrend {
			fmt.Printf("%-12s %-10.1f %-8d %d\n",
				point.Date.Format("2006-01-02"),
				point.AverageLag,
				point.MaxLag,
				point.Count,
			)
		}
		fmt.Println()
	}
	
	fmt.Println("💡 Use 'ai-dep-manager lag plan [project-id]' to create a resolution plan")
}

func displayResolutionPlan(plan *services.LagResolutionPlan) {
	fmt.Printf("📋 Lag Resolution Plan for Project %d\n", plan.ProjectID)
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Generated: %s\n", plan.GeneratedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Total Packages: %d\n", plan.TotalPackages)
	fmt.Printf("Estimated Time: %s\n", plan.EstimatedTime)
	fmt.Printf("Risk Level: %s\n", plan.RiskLevel)
	fmt.Println()
	
	// Prerequisites
	if len(plan.Prerequisites) > 0 {
		fmt.Println("📋 Prerequisites:")
		for _, prereq := range plan.Prerequisites {
			fmt.Printf("  - %s\n", prereq)
		}
		fmt.Println()
	}
	
	// Phases
	if len(plan.Phases) > 0 {
		fmt.Println("📊 Execution Phases:")
		for _, phase := range plan.Phases {
			fmt.Printf("\nPhase %d: %s\n", phase.Phase, phase.Name)
			fmt.Printf("Description: %s\n", phase.Description)
			fmt.Printf("Packages: %d | Order: %s | Risk: %s | Duration: %s\n", 
				len(phase.Packages), phase.Order, phase.RiskLevel, phase.Duration)
			
			if len(phase.Packages) > 0 {
				fmt.Println("\nPackages:")
				fmt.Printf("  %-25s %-12s %-12s %-10s %s\n", 
					"Name", "Current", "Target", "Type", "Reason")
				fmt.Println("  " + strings.Repeat("-", 80))
				
				for _, pkg := range phase.Packages {
					fmt.Printf("  %-25s %-12s %-12s %-10s %s\n",
						truncateString(pkg.Name, 25),
						truncateString(pkg.CurrentVersion, 12),
						truncateString(pkg.TargetVersion, 12),
						pkg.UpdateType,
						truncateString(pkg.Reason, 30),
					)
				}
			}
		}
		fmt.Println()
	}
	
	if plan.TotalPackages == 0 {
		fmt.Println("✅ No lag resolution needed - all dependencies are up to date!")
	} else {
		fmt.Printf("💡 Use 'ai-dep-manager lag execute %d' to execute this plan\n", plan.ProjectID)
	}
}

// Helper functions

func getPriorityIcon(priority string) string {
	switch priority {
	case "critical":
		return "🚨"
	case "high":
		return "⚠️"
	case "medium":
		return "📋"
	case "low":
		return "💡"
	default:
		return "📋"
	}
}

// printJSON prints the given data as formatted JSON
func printJSON(data interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
}
