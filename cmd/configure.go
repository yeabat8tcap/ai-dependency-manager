package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"


	"github.com/8tcapital/ai-dep-manager/internal/packagemanager"
	pmtypes "github.com/8tcapital/ai-dep-manager/internal/packagemanager/types"
	"github.com/8tcapital/ai-dep-manager/internal/services"
	"github.com/spf13/cobra"
)

// configureCmd represents the configure command
var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure projects for monitoring",
	Long: `Configure projects for dependency monitoring. This command helps you:
- Add new projects to monitor
- Auto-discover projects in a directory
- Configure project-specific settings
- Set up package manager preferences`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigure(cmd, args)
	},
}

var (
	projectPath     string
	projectName     string
	packageManager  string
	autoDiscover    bool
	discoveryPath   string
	interactive     bool
)

func runConfigure(cmd *cobra.Command, args []string) error {
	projectService := services.NewProjectService()
	
	if autoDiscover {
		return runAutoDiscover(cmd, projectService)
	}
	
	if interactive {
		return runInteractiveConfigure(cmd, projectService)
	}
	
	if projectPath == "" {
		return fmt.Errorf("project path is required. Use --project-path or --interactive")
	}
	
	return runDirectConfigure(cmd, projectService)
}

func runAutoDiscover(cmd *cobra.Command, projectService *services.ProjectService) error {
	path := discoveryPath
	if path == "" {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}
	
	fmt.Printf("🔍 Auto-discovering projects in: %s\n", path)
	
	projects, err := projectService.AutoDiscoverProjects(cmd.Context(), path)
	if err != nil {
		return fmt.Errorf("auto-discovery failed: %w", err)
	}
	
	if len(projects) == 0 {
		fmt.Println("❌ No projects found")
		return nil
	}
	
	fmt.Printf("✅ Discovered %d projects:\n\n", len(projects))
	for _, project := range projects {
		fmt.Printf("  📦 %s (%s)\n", project.Name, project.Type)
		fmt.Printf("      Path: %s\n", project.Path)
		fmt.Printf("      Config: %s\n\n", project.ConfigFile)
	}
	
	return nil
}

func runInteractiveConfigure(cmd *cobra.Command, projectService *services.ProjectService) error {
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Println("🔧 Interactive Project Configuration")
	fmt.Println("===================================")
	
	// Get project path
	fmt.Print("Enter project path (or press Enter for current directory): ")
	path, _ := reader.ReadString('\n')
	path = strings.TrimSpace(path)
	
	if path == "" {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}
	
	// Expand path
	if strings.HasPrefix(path, "~") {
		homeDir, _ := os.UserHomeDir()
		path = filepath.Join(homeDir, path[1:])
	}
	
	// Auto-detect package managers
	fmt.Printf("\n🔍 Detecting package managers in: %s\n", path)
	
	availablePMs := packagemanager.GetAvailablePackageManagers(cmd.Context())
	if len(availablePMs) == 0 {
		fmt.Println("❌ No package managers available on this system")
		return fmt.Errorf("no package managers available")
	}
	
	detectedProjects, err := packagemanager.DetectAllProjects(cmd.Context(), path)
	if err != nil {
		return fmt.Errorf("failed to detect projects: %w", err)
	}
	
	if len(detectedProjects) == 0 {
		fmt.Println("❌ No projects detected in the specified path")
		return fmt.Errorf("no projects found")
	}
	
	fmt.Printf("✅ Found %d project(s):\n\n", len(detectedProjects))
	
	for i, detected := range detectedProjects {
		fmt.Printf("%d. %s (%s)\n", i+1, detected.Name, detected.PackageManager)
		fmt.Printf("   Path: %s\n", detected.Path)
		fmt.Printf("   Config: %s\n\n", detected.ConfigFile)
	}
	
	// Let user choose which projects to configure
	fmt.Print("Enter project numbers to configure (comma-separated, or 'all'): ")
	selection, _ := reader.ReadString('\n')
	selection = strings.TrimSpace(selection)
	
	var selectedProjects []pmtypes.Project
	
	if selection == "all" {
		selectedProjects = detectedProjects
	} else {
		// Parse selection
		parts := strings.Split(selection, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			var idx int
			if _, err := fmt.Sscanf(part, "%d", &idx); err == nil && idx > 0 && idx <= len(detectedProjects) {
				selectedProjects = append(selectedProjects, detectedProjects[idx-1])
			}
		}
	}
	
	if len(selectedProjects) == 0 {
		fmt.Println("❌ No projects selected")
		return nil
	}
	
	// Configure selected projects
	fmt.Printf("\n🔧 Configuring %d project(s)...\n\n", len(selectedProjects))
	
	for _, detected := range selectedProjects {
		name := detected.Name
		if name == "" {
			name = filepath.Base(detected.Path)
		}
		
		fmt.Printf("Configuring: %s\n", name)
		
		project, err := projectService.CreateProject(cmd.Context(), name, detected.Path, detected.PackageManager)
		if err != nil {
			fmt.Printf("❌ Failed to configure %s: %v\n\n", name, err)
			continue
		}
		
		fmt.Printf("✅ Successfully configured: %s (ID: %d)\n\n", project.Name, project.ID)
	}
	
	return nil
}

func runDirectConfigure(cmd *cobra.Command, projectService *services.ProjectService) error {
	// Expand path
	path := projectPath
	if strings.HasPrefix(path, "~") {
		homeDir, _ := os.UserHomeDir()
		path = filepath.Join(homeDir, path[1:])
	}
	
	// Auto-detect package manager if not specified
	if packageManager == "" {
		detectedProjects, err := packagemanager.DetectAllProjects(cmd.Context(), path)
		if err != nil {
			return fmt.Errorf("failed to detect project type: %w", err)
		}
		
		if len(detectedProjects) == 0 {
			return fmt.Errorf("no supported projects found in %s", path)
		}
		
		// Use the first detected project type
		packageManager = detectedProjects[0].PackageManager
		fmt.Printf("Auto-detected package manager: %s\n", packageManager)
	}
	
	// Generate project name if not provided
	name := projectName
	if name == "" {
		name = filepath.Base(path)
	}
	
	fmt.Printf("Configuring project: %s\n", name)
	fmt.Printf("Path: %s\n", path)
	fmt.Printf("Package Manager: %s\n", packageManager)
	
	project, err := projectService.CreateProject(cmd.Context(), name, path, packageManager)
	if err != nil {
		return fmt.Errorf("failed to configure project: %w", err)
	}
	
	fmt.Printf("✅ Successfully configured project: %s (ID: %d)\n", project.Name, project.ID)
	
	return nil
}

func init() {
	rootCmd.AddCommand(configureCmd)
	
	configureCmd.Flags().StringVar(&projectPath, "project-path", "", "Path to the project directory")
	configureCmd.Flags().StringVar(&projectName, "project-name", "", "Name for the project (defaults to directory name)")
	configureCmd.Flags().StringVar(&packageManager, "package-manager", "", "Package manager type (npm, pip, maven, gradle)")
	configureCmd.Flags().BoolVar(&autoDiscover, "auto-discover", false, "Auto-discover projects in directory")
	configureCmd.Flags().StringVar(&discoveryPath, "discovery-path", "", "Path for auto-discovery (defaults to current directory)")
	configureCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive configuration mode")
}
