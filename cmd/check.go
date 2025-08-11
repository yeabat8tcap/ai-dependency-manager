package cmd

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/8tcapital/ai-dep-manager/internal/database"
	"github.com/8tcapital/ai-dep-manager/internal/models"
	"github.com/8tcapital/ai-dep-manager/internal/services"
	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check dependency status and available updates",
	Long: `Check the current status of dependencies and available updates. This command:
- Shows current dependency versions
- Lists available updates with details
- Highlights security vulnerabilities
- Provides update recommendations

Examples:
  ai-dep-manager check                    # Check all projects
  ai-dep-manager check --project my-app   # Check specific project
  ai-dep-manager check --outdated-only    # Show only outdated dependencies`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCheck(cmd, args)
	},
}

var (
	checkProject     string
	checkProjectID   uint
	outdatedOnly     bool
	securityOnly     bool
	showDetails      bool
	sortBy           string
	aiInsights       bool
)

func runCheck(cmd *cobra.Command, args []string) error {
	projectService := services.NewProjectService()
	
	if checkProject != "" {
		return runCheckProject(cmd, projectService)
	}
	
	if checkProjectID != 0 {
		return runCheckProjectByID(cmd, projectService)
	}
	
	return runCheckAllProjects(cmd, projectService)
}

func runCheckProject(cmd *cobra.Command, projectService *services.ProjectService) error {
	fmt.Printf("📋 Checking project: %s\n", checkProject)
	
	// Get project by name
	project, err := projectService.GetProjectByName(cmd.Context(), checkProject)
	if err != nil {
		return fmt.Errorf("failed to find project '%s': %w", checkProject, err)
	}
	
	return displayProjectStatus(project)
}

func runCheckProjectByID(cmd *cobra.Command, projectService *services.ProjectService) error {
	fmt.Printf("📋 Checking project ID: %d\n", checkProjectID)
	
	// Get project by ID
	project, err := projectService.GetProject(cmd.Context(), checkProjectID)
	if err != nil {
		return fmt.Errorf("failed to find project with ID %d: %w", checkProjectID, err)
	}
	
	return displayProjectStatus(project)
}

func runCheckAllProjects(cmd *cobra.Command, projectService *services.ProjectService) error {
	fmt.Println("📋 Checking all projects...")
	
	// Get all projects
	projects, err := projectService.ListProjects(cmd.Context(), nil)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}
	
	if len(projects) == 0 {
		fmt.Println("❌ No projects configured. Run 'ai-dep-manager configure' to add projects.")
		return nil
	}
	
	// Display summary for all projects
	totalProjects := 0
	totalDependencies := 0
	totalOutdated := 0
	totalSecurity := 0
	
	for _, project := range projects {
		if !project.Enabled {
			continue
		}
		
		totalProjects++
		
		// Get dependencies for this project
		var dependencies []models.Dependency
		db := database.GetDB()
		query := db.Where("project_id = ?", project.ID)
		
		if outdatedOnly {
			query = query.Where("status = ?", "outdated")
		}
		
		if err := query.Find(&dependencies).Error; err != nil {
			fmt.Printf("❌ Error loading dependencies for %s: %v\n", project.Name, err)
			continue
		}
		
		// Get updates for this project
		var updates []models.Update
		if len(dependencies) > 0 {
			depIDs := make([]uint, len(dependencies))
			for i, dep := range dependencies {
				depIDs[i] = dep.ID
			}
			
			updateQuery := db.Where("dependency_id IN ?", depIDs).Where("status = ?", "pending")
			if securityOnly {
				updateQuery = updateQuery.Where("security_fix = ?", true)
			}
			
			updateQuery.Find(&updates)
		}
		
		outdatedCount := 0
		securityCount := 0
		
		for _, dep := range dependencies {
			if dep.Status == "outdated" {
				outdatedCount++
			}
		}
		
		for _, update := range updates {
			if update.SecurityFix {
				securityCount++
			}
		}
		
		totalDependencies += len(dependencies)
		totalOutdated += outdatedCount
		totalSecurity += securityCount
		
		// Display project summary
		status := "✅"
		if securityCount > 0 {
			status = "🔒"
		} else if outdatedCount > 0 {
			status = "🔄"
		}
		
		fmt.Printf("%s %s (%s)\n", status, project.Name, project.Type)
		fmt.Printf("   Dependencies: %d total, %d outdated", len(dependencies), outdatedCount)
		if securityCount > 0 {
			fmt.Printf(", %d security updates", securityCount)
		}
		fmt.Println()
		
		if project.LastScan != nil {
			fmt.Printf("   Last scan: %s\n", project.LastScan.Format("2006-01-02 15:04:05"))
		} else {
			fmt.Printf("   Last scan: Never (run 'ai-dep-manager scan --project %s')\n", project.Name)
		}
		fmt.Println()
	}
	
	// Overall summary
	fmt.Println("📊 Overall Summary:")
	fmt.Printf("   Projects: %d enabled\n", totalProjects)
	fmt.Printf("   Dependencies: %d total, %d outdated\n", totalDependencies, totalOutdated)
	if totalSecurity > 0 {
		fmt.Printf("   🔒 Security updates: %d\n", totalSecurity)
	}
	
	if totalOutdated > 0 {
		fmt.Printf("\n💡 Run 'ai-dep-manager check --project <name>' for detailed information\n")
		fmt.Printf("💡 Run 'ai-dep-manager update --preview' to see what would be updated\n")
	}
	
	return nil
}

func displayProjectStatus(project *models.Project) error {
	fmt.Printf("\n📦 Project: %s (%s)\n", project.Name, project.Type)
	fmt.Printf("Path: %s\n", project.Path)
	fmt.Printf("Config: %s\n", project.ConfigFile)
	fmt.Printf("Status: %s\n", getProjectStatusText(project))
	
	if project.LastScan != nil {
		fmt.Printf("Last scan: %s\n", project.LastScan.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Printf("Last scan: Never\n")
		fmt.Printf("💡 Run 'ai-dep-manager scan --project %s' to scan for updates\n", project.Name)
		return nil
	}
	
	// Get dependencies
	db := database.GetDB()
	var dependencies []models.Dependency
	query := db.Where("project_id = ?", project.ID)
	
	if outdatedOnly {
		query = query.Where("status = ?", "outdated")
	}
	
	if err := query.Find(&dependencies).Error; err != nil {
		return fmt.Errorf("failed to load dependencies: %w", err)
	}
	
	if len(dependencies) == 0 {
		if outdatedOnly {
			fmt.Println("\n✅ No outdated dependencies found!")
		} else {
			fmt.Println("\n❌ No dependencies found. This might indicate the project hasn't been scanned yet.")
		}
		return nil
	}
	
	// Sort dependencies
	sortDependencies(dependencies, sortBy)
	
	// Get updates for these dependencies
	depIDs := make([]uint, len(dependencies))
	for i, dep := range dependencies {
		depIDs[i] = dep.ID
	}
	
	var updates []models.Update
	updateQuery := db.Where("dependency_id IN ?", depIDs).Where("status = ?", "pending")
	if securityOnly {
		updateQuery = updateQuery.Where("security_fix = ?", true)
	}
	updateQuery.Find(&updates)
	
	// Create update map for quick lookup
	updateMap := make(map[uint][]models.Update)
	for _, update := range updates {
		updateMap[update.DependencyID] = append(updateMap[update.DependencyID], update)
	}
	
	// Display dependencies
	fmt.Printf("\n📋 Dependencies (%d):\n", len(dependencies))
	fmt.Println(strings.Repeat("=", 80))
	
	upToDateCount := 0
	outdatedCount := 0
	unknownCount := 0
	securityCount := 0
	
	for _, dep := range dependencies {
		depUpdates := updateMap[dep.ID]
		
		// Count by status
		switch dep.Status {
		case "up-to-date":
			upToDateCount++
		case "outdated":
			outdatedCount++
		default:
			unknownCount++
		}
		
		// Check for security updates
		hasSecurityUpdate := false
		for _, update := range depUpdates {
			if update.SecurityFix {
				hasSecurityUpdate = true
				securityCount++
				break
			}
		}
		
		// Skip if showing security only and no security updates
		if securityOnly && !hasSecurityUpdate {
			continue
		}
		
		// Display dependency
		status := getStatusIcon(dep.Status, hasSecurityUpdate)
		fmt.Printf("%s %s\n", status, dep.Name)
		fmt.Printf("   Current: %s", dep.CurrentVersion)
		if dep.RequiredVersion != "" && dep.RequiredVersion != dep.CurrentVersion {
			fmt.Printf(" (required: %s)", dep.RequiredVersion)
		}
		fmt.Println()
		
		if dep.LatestVersion != "" && dep.LatestVersion != dep.CurrentVersion {
			fmt.Printf("   Latest:  %s\n", dep.LatestVersion)
		}
		
		if len(depUpdates) > 0 {
			fmt.Printf("   Updates: ")
			for i, update := range depUpdates {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Printf("%s→%s", update.FromVersion, update.ToVersion)
				if update.SecurityFix {
					fmt.Print(" 🔒")
				}
				if update.BreakingChange {
					fmt.Print(" ⚠️")
				}
			}
			fmt.Println()
		}
		
		if showDetails && dep.LastChecked != nil {
			fmt.Printf("   Checked: %s\n", dep.LastChecked.Format("2006-01-02 15:04"))
		}
		
		fmt.Println()
	}
	
	// Summary
	fmt.Println("📊 Summary:")
	fmt.Printf("   ✅ Up to date: %d\n", upToDateCount)
	fmt.Printf("   🔄 Outdated: %d\n", outdatedCount)
	if securityCount > 0 {
		fmt.Printf("   🔒 Security updates: %d\n", securityCount)
	}
	if unknownCount > 0 {
		fmt.Printf("   ❓ Unknown: %d\n", unknownCount)
	}
	
	if outdatedCount > 0 || securityCount > 0 {
		fmt.Printf("\n💡 Run 'ai-dep-manager update --project %s --preview' to see update details\n", project.Name)
	}
	
	return nil
}

func getProjectStatusText(project *models.Project) string {
	if !project.Enabled {
		return "❌ Disabled"
	}
	if project.LastScan == nil {
		return "❓ Not scanned"
	}
	return "✅ Enabled"
}

func getStatusIcon(status string, hasSecurityUpdate bool) string {
	if hasSecurityUpdate {
		return "🔒"
	}
	
	switch status {
	case "up-to-date":
		return "✅"
	case "outdated":
		return "🔄"
	case "vulnerable":
		return "⚠️"
	default:
		return "❓"
	}
}

func sortDependencies(dependencies []models.Dependency, sortBy string) {
	switch sortBy {
	case "name":
		sort.Slice(dependencies, func(i, j int) bool {
			return dependencies[i].Name < dependencies[j].Name
		})
	case "status":
		sort.Slice(dependencies, func(i, j int) bool {
			statusOrder := map[string]int{
				"vulnerable": 0,
				"outdated":   1,
				"unknown":    2,
				"up-to-date": 3,
			}
			return statusOrder[dependencies[i].Status] < statusOrder[dependencies[j].Status]
		})
	case "updated":
		sort.Slice(dependencies, func(i, j int) bool {
			if dependencies[i].UpdatedAt.IsZero() && dependencies[j].UpdatedAt.IsZero() {
				return false
			}
			if dependencies[i].UpdatedAt.IsZero() {
				return false
			}
			if dependencies[j].UpdatedAt.IsZero() {
				return true
			}
			return dependencies[i].UpdatedAt.After(dependencies[j].UpdatedAt)
		})
	default:
		// Default sort by name
		sort.Slice(dependencies, func(i, j int) bool {
			return dependencies[i].Name < dependencies[j].Name
		})
	}
}

func init() {
	rootCmd.AddCommand(checkCmd)
	
	checkCmd.Flags().StringVar(&checkProject, "project", "", "Check specific project by name")
	checkCmd.Flags().UintVar(&checkProjectID, "project-id", 0, "Check specific project by ID")
	checkCmd.Flags().BoolVar(&outdatedOnly, "outdated-only", false, "Show only outdated dependencies")
	checkCmd.Flags().BoolVar(&securityOnly, "security-only", false, "Show only dependencies with security updates")
	checkCmd.Flags().BoolVar(&showDetails, "details", false, "Show detailed information")
	checkCmd.Flags().StringVar(&sortBy, "sort", "name", "Sort by: name, status, updated")
	checkCmd.Flags().BoolVar(&aiInsights, "ai-insights", false, "Enable AI-powered dependency insights and recommendations")
	
	// Add validation and argument parsing
	checkCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		validSorts := map[string]bool{
			"name":    true,
			"status":  true,
			"updated": true,
		}
		
		if !validSorts[sortBy] {
			return fmt.Errorf("invalid sort option '%s'. Valid options: name, status, updated", sortBy)
		}
		
		// Parse project ID if provided as string argument
		if len(args) > 0 {
			if id, err := strconv.ParseUint(args[0], 10, 32); err == nil {
				checkProjectID = uint(id)
			} else {
				checkProject = args[0]
			}
		}
		
		return nil
	}
}
