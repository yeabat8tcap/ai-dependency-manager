package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/8tcapital/superint-dep-manager/internal/logger"
	"github.com/8tcapital/superint-dep-manager/internal/scanner"
	"github.com/8tcapital/superint-dep-manager/internal/services"
	"github.com/spf13/cobra"
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan projects for dependency updates",
	Long: `Scan configured projects for dependency updates. This command:
- Analyzes current dependencies
- Checks for available updates
- Identifies potential security vulnerabilities
- Provides update recommendations

Examples:
  superint-dep-manager scan                    # Scan all projects
  superint-dep-manager scan --project my-app   # Scan specific project
  superint-dep-manager scan --type security    # Security-focused scan`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runScan(cmd, args)
	},
}

var (
	scanProject     string
	scanProjectID   uint
	scanType        string
	forceRefresh    bool
	maxConcurrency  int
	scanTimeout     time.Duration
	aiAnalysis      bool
)

func runScan(cmd *cobra.Command, args []string) error {
	projectService := services.NewProjectService()
	dependencyScanner := scanner.NewScanner(maxConcurrency)
	
	// Prepare scan options
	options := &scanner.ScanOptions{
		ScanType:     scanType,
		ForceRefresh: forceRefresh,
		Timeout:      scanTimeout,
	}
	
	if scanProject != "" {
		return runScanProject(cmd, projectService, dependencyScanner, options)
	}
	
	if scanProjectID != 0 {
		return runScanProjectByID(cmd, dependencyScanner, options)
	}
	
	return runScanAllProjects(cmd, projectService, dependencyScanner, options)
}

func runScanProject(cmd *cobra.Command, projectService *services.ProjectService, dependencyScanner *scanner.Scanner, options *scanner.ScanOptions) error {
	fmt.Printf("🔍 Scanning project: %s\n", scanProject)
	
	// Get project by name
	project, err := projectService.GetProjectByName(cmd.Context(), scanProject)
	if err != nil {
		return fmt.Errorf("failed to find project '%s': %w", scanProject, err)
	}
	
	if !project.Enabled {
		fmt.Printf("⚠️  Project '%s' is disabled. Enable it first with: superint-dep-manager configure\n", scanProject)
		return nil
	}
	
	options.ProjectID = project.ID
	
	// Run scan
	result, err := dependencyScanner.ScanProject(cmd.Context(), project.ID, options)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}
	
	// Display results
	displayScanResult(project.Name, result)
	
	return nil
}

func runScanProjectByID(cmd *cobra.Command, dependencyScanner *scanner.Scanner, options *scanner.ScanOptions) error {
	fmt.Printf("🔍 Scanning project ID: %d\n", scanProjectID)
	
	options.ProjectID = scanProjectID
	
	// Run scan
	result, err := dependencyScanner.ScanProject(cmd.Context(), scanProjectID, options)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}
	
	// Display results
	displayScanResult(fmt.Sprintf("Project %d", scanProjectID), result)
	
	return nil
}

func runScanAllProjects(cmd *cobra.Command, projectService *services.ProjectService, dependencyScanner *scanner.Scanner, options *scanner.ScanOptions) error {
	fmt.Println("🔍 Scanning all enabled projects...")
	
	// Get all enabled projects
	enabled := true
	projects, err := projectService.ListProjects(cmd.Context(), &enabled)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}
	
	if len(projects) == 0 {
		fmt.Println("❌ No enabled projects found. Configure projects first with: superint-dep-manager configure")
		return nil
	}
	
	fmt.Printf("Found %d enabled project(s)\n\n", len(projects))
	
	// Scan all projects
	results, err := dependencyScanner.ScanAllProjects(cmd.Context(), options)
	if err != nil {
		logger.Error("Scan completed with errors: %v", err)
	}
	
	// Display summary
	totalDependencies := 0
	totalUpdates := 0
	successfulScans := 0
	
	for i, result := range results {
		if i < len(projects) {
			fmt.Printf("📦 %s:\n", projects[i].Name)
		} else {
			fmt.Printf("📦 Project %d:\n", result.ProjectID)
		}
		
		if len(result.Errors) > 0 {
			fmt.Printf("   ❌ %d error(s) occurred\n", len(result.Errors))
			for _, scanErr := range result.Errors {
				fmt.Printf("      - %v\n", scanErr)
			}
		} else {
			fmt.Printf("   ✅ %d dependencies, %d updates available\n", result.DependenciesFound, result.UpdatesFound)
			totalDependencies += result.DependenciesFound
			totalUpdates += result.UpdatesFound
			successfulScans++
		}
		fmt.Println()
	}
	
	// Summary
	fmt.Println("📊 Scan Summary:")
	fmt.Printf("   Projects scanned: %d/%d successful\n", successfulScans, len(projects))
	fmt.Printf("   Total dependencies: %d\n", totalDependencies)
	fmt.Printf("   Total updates available: %d\n", totalUpdates)
	
	if totalUpdates > 0 {
		fmt.Printf("\n💡 Run 'superint-dep-manager check' to see detailed update information\n")
		fmt.Printf("💡 Run 'superint-dep-manager update --preview' to see what would be updated\n")
	}
	
	return nil
}

func displayScanResult(projectName string, result *scanner.ScanResult) {
	fmt.Printf("\n📊 Scan Results for %s:\n", projectName)
	fmt.Println("=" + fmt.Sprintf("%*s", len(projectName)+18, "="))
	
	if len(result.Errors) > 0 {
		fmt.Printf("❌ Errors: %d\n", len(result.Errors))
		for _, err := range result.Errors {
			fmt.Printf("   - %v\n", err)
		}
		fmt.Println()
	}
	
	fmt.Printf("📦 Dependencies found: %d\n", result.DependenciesFound)
	fmt.Printf("🔄 Updates available: %d\n", result.UpdatesFound)
	
	if len(result.NewDependencies) > 0 {
		fmt.Printf("🆕 New dependencies: %d\n", len(result.NewDependencies))
	}
	
	if len(result.UpdatedDependencies) > 0 {
		fmt.Printf("📝 Updated dependencies: %d\n", len(result.UpdatedDependencies))
	}
	
	if result.UpdatesFound > 0 {
		fmt.Println("\n🔄 Available Updates:")
		for _, update := range result.AvailableUpdates {
			status := "📦"
			if update.SecurityFix {
				status = "🔒"
			}
			if update.BreakingChange {
				status = "⚠️"
			}
			
			// Find the dependency name
			var depName string
			for _, dep := range result.NewDependencies {
				if dep.ID == update.DependencyID {
					depName = dep.Name
					break
				}
			}
			for _, dep := range result.UpdatedDependencies {
				if dep.ID == update.DependencyID {
					depName = dep.Name
					break
				}
			}
			
			fmt.Printf("   %s %s: %s → %s (%s)\n", 
				status, depName, update.FromVersion, update.ToVersion, update.UpdateType)
		}
		
		fmt.Printf("\n💡 Run 'superint-dep-manager check --project %s' for detailed information\n", projectName)
		fmt.Printf("💡 Run 'superint-dep-manager update --project %s --preview' to preview updates\n", projectName)
	} else {
		fmt.Println("\n✅ All dependencies are up to date!")
	}
}

func init() {
	rootCmd.AddCommand(scanCmd)
	
	scanCmd.Flags().StringVar(&scanProject, "project", "", "Scan specific project by name")
	scanCmd.Flags().UintVar(&scanProjectID, "project-id", 0, "Scan specific project by ID")
	scanCmd.Flags().StringVar(&scanType, "type", "full", "Scan type: full, incremental, security")
	scanCmd.Flags().BoolVar(&forceRefresh, "force-refresh", false, "Force refresh of dependency information")
	scanCmd.Flags().IntVar(&maxConcurrency, "concurrency", 5, "Maximum concurrent operations")
	scanCmd.Flags().DurationVar(&scanTimeout, "timeout", 5*time.Minute, "Scan timeout duration")
	scanCmd.Flags().BoolVar(&aiAnalysis, "ai-analysis", false, "Enable Superintelligence-powered changelog analysis")
	
	// Add validation for scan type
	scanCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		validTypes := map[string]bool{
			"full":        true,
			"incremental": true,
			"security":    true,
		}
		
		if !validTypes[scanType] {
			return fmt.Errorf("invalid scan type '%s'. Valid types: full, incremental, security", scanType)
		}
		
		// Parse project ID if provided as string argument
		if len(args) > 0 {
			if id, err := strconv.ParseUint(args[0], 10, 32); err == nil {
				scanProjectID = uint(id)
			} else {
				scanProject = args[0]
			}
		}
		
		return nil
	}
}
