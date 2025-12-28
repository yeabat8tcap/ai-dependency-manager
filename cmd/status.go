package cmd

import (
	"fmt"

	"github.com/8tcapital/superint-dep-manager/internal/database"
	"github.com/8tcapital/superint-dep-manager/internal/logger"
	"github.com/8tcapital/superint-dep-manager/internal/models"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show system status and project overview",
	Long: `Display the current status of Superintelligence Dependency Manager including:
- Database connectivity
- Monitored projects
- Recent scan results
- Background agent status`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStatus()
	},
}

func runStatus() error {
	fmt.Println("Superintelligence Dependency Manager Status")
	fmt.Println("============================")
	
	// Check database health
	fmt.Print("Database: ")
	if err := database.Health(); err != nil {
		fmt.Printf("❌ Error - %v\n", err)
		return err
	}
	fmt.Println("✅ Connected")
	
	// Get project count
	db := database.GetDB()
	var projectCount int64
	if err := db.Model(&models.Project{}).Count(&projectCount); err != nil {
		logger.Error("Failed to count projects: %v", err)
		fmt.Printf("Projects: ❌ Error counting projects\n")
	} else {
		fmt.Printf("Projects: %d monitored\n", projectCount)
	}
	
	// Get recent scan results
	var recentScans []models.ScanResult
	if err := db.Order("created_at DESC").Limit(5).Find(&recentScans).Error; err != nil {
		logger.Error("Failed to fetch recent scans: %v", err)
		fmt.Printf("Recent Scans: ❌ Error fetching scan history\n")
	} else {
		fmt.Printf("Recent Scans: %d in history\n", len(recentScans))
		if len(recentScans) > 0 {
			fmt.Println("\nLast 5 scans:")
			for _, scan := range recentScans {
				status := "✅"
				if scan.Status == "failed" {
					status = "❌"
				} else if scan.Status == "running" {
					status = "🔄"
				}
				fmt.Printf("  %s %s - %s (%d deps, %d updates)\n", 
					status, scan.StartedAt.Format("2006-01-02 15:04"), 
					scan.ScanType, scan.DependenciesFound, scan.UpdatesFound)
			}
		}
	}
	
	// Agent status (placeholder for now)
	fmt.Printf("Background Agent: %s\n", getAgentStatus())
	
	return nil
}

func getAgentStatus() string {
	// TODO: Implement actual agent status checking
	if cfg != nil && cfg.Agent.Enabled {
		return "🔄 Enabled (not implemented yet)"
	}
	return "⏸️  Disabled"
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
