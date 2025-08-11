package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/logger"
	"github.com/8tcapital/ai-dep-manager/internal/reporting"
	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate reports and analytics",
	Long:  `Generate comprehensive reports and analytics about dependency management, security, and system performance.`,
}

var generateReportCmd = &cobra.Command{
	Use:   "generate [type]",
	Short: "Generate a specific type of report",
	Long: `Generate a report of the specified type. Available types:
- summary: High-level overview of dependency management status
- security: Comprehensive security analysis of dependencies
- updates: Analysis of dependency updates and maintenance
- dependencies: Detailed analysis of all project dependencies
- performance: Analysis of system performance and efficiency
- compliance: Compliance status and regulatory adherence`,
	Args: cobra.ExactArgs(1),
	Run:  runGenerateReport,
}

var exportReportCmd = &cobra.Command{
	Use:   "export [file]",
	Short: "Export the last generated report",
	Long:  `Export the last generated report to a file in the specified format.`,
	Args:  cobra.ExactArgs(1),
	Run:   runExportReport,
}

var analyticsCmd = &cobra.Command{
	Use:   "analytics",
	Short: "View detailed analytics",
	Long:  `View detailed analytics about dependencies, security, and system performance.`,
	Run:   runAnalytics,
}

var dependencyAnalyticsCmd = &cobra.Command{
	Use:   "dependencies",
	Short: "View dependency analytics",
	Long:  `View detailed analytics for all dependencies including risk scores and recommendations.`,
	Run:   runDependencyAnalytics,
}

// Global variables for report command flags
var (
	reportFormat    string
	reportTimeRange string
	reportOutput    string
	reportProject   string
	reportLimit     int
	reportSort      string
)

// Global variable to store the last generated report
var lastGeneratedReport *reporting.Report

func init() {
	rootCmd.AddCommand(reportCmd)
	reportCmd.AddCommand(generateReportCmd)
	reportCmd.AddCommand(exportReportCmd)
	reportCmd.AddCommand(analyticsCmd)
	analyticsCmd.AddCommand(dependencyAnalyticsCmd)

	// Generate report flags
	generateReportCmd.Flags().StringVar(&reportFormat, "format", "json", "Output format (json, csv, html, pdf)")
	generateReportCmd.Flags().StringVar(&reportTimeRange, "time-range", "30d", "Time range for the report (7d, 30d, 90d, 1y, or custom YYYY-MM-DD:YYYY-MM-DD)")
	generateReportCmd.Flags().StringVar(&reportOutput, "output", "", "Output file path (if not specified, prints to stdout)")
	generateReportCmd.Flags().StringVar(&reportProject, "project", "", "Filter by specific project")

	// Export report flags
	exportReportCmd.Flags().StringVar(&reportFormat, "format", "json", "Export format (json, csv, html, pdf)")

	// Analytics flags
	dependencyAnalyticsCmd.Flags().StringVar(&reportProject, "project", "", "Filter by specific project")
	dependencyAnalyticsCmd.Flags().IntVar(&reportLimit, "limit", 50, "Limit number of results")
	dependencyAnalyticsCmd.Flags().StringVar(&reportSort, "sort", "risk", "Sort by (risk, name, lag, security)")
}

func runGenerateReport(cmd *cobra.Command, args []string) {
	reportType := args[0]
	
	// Validate report type
	validTypes := []string{"summary", "security", "updates", "dependencies", "performance", "compliance"}
	if !contains(validTypes, reportType) {
		logger.Error("Invalid report type: %s. Valid types: %s", reportType, strings.Join(validTypes, ", "))
		os.Exit(1)
	}
	
	// Parse time range
	timeRange, err := parseTimeRange(reportTimeRange)
	if err != nil {
		logger.Error("Invalid time range: %v", err)
		os.Exit(1)
	}
	
	// Create reporting service
	reportingService := reporting.NewReportingService()
	
	// Generate report
	logger.Info("Generating %s report for time range: %s to %s", 
		reportType, timeRange.Start.Format("2006-01-02"), timeRange.End.Format("2006-01-02"))
	
	report, err := reportingService.GenerateReport(context.Background(), 
		reporting.ReportType(reportType), timeRange)
	if err != nil {
		logger.Error("Failed to generate report: %v", err)
		os.Exit(1)
	}
	
	// Store for potential export
	lastGeneratedReport = report
	
	// Output or export report
	if reportOutput != "" {
		// Determine format from file extension if not specified
		if reportFormat == "json" && reportOutput != "" {
			ext := strings.ToLower(filepath.Ext(reportOutput))
			switch ext {
			case ".csv":
				reportFormat = "csv"
			case ".html":
				reportFormat = "html"
			case ".pdf":
				reportFormat = "pdf"
			}
		}
		
		if err := reportingService.ExportReport(report, reporting.ReportFormat(reportFormat), reportOutput); err != nil {
			logger.Error("Failed to export report: %v", err)
			os.Exit(1)
		}
		
		logger.Info("Report exported to: %s", reportOutput)
	} else {
		// Print to stdout
		printReport(report)
	}
}

func runExportReport(cmd *cobra.Command, args []string) {
	if lastGeneratedReport == nil {
		logger.Error("No report available to export. Generate a report first.")
		os.Exit(1)
	}
	
	filename := args[0]
	
	// Determine format from file extension if not specified
	if reportFormat == "json" {
		ext := strings.ToLower(filepath.Ext(filename))
		switch ext {
		case ".csv":
			reportFormat = "csv"
		case ".html":
			reportFormat = "html"
		case ".pdf":
			reportFormat = "pdf"
		}
	}
	
	reportingService := reporting.NewReportingService()
	
	if err := reportingService.ExportReport(lastGeneratedReport, reporting.ReportFormat(reportFormat), filename); err != nil {
		logger.Error("Failed to export report: %v", err)
		os.Exit(1)
	}
	
	logger.Info("Report exported to: %s", filename)
}

func runAnalytics(cmd *cobra.Command, args []string) {
	fmt.Println("📊 AI Dependency Manager - Analytics Dashboard")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()
	
	// Generate a quick summary report
	reportingService := reporting.NewReportingService()
	timeRange := reporting.TimeRange{
		Start: time.Now().AddDate(0, 0, -30),
		End:   time.Now(),
	}
	
	report, err := reportingService.GenerateReport(context.Background(), 
		reporting.ReportTypeSummary, timeRange)
	if err != nil {
		logger.Error("Failed to generate analytics: %v", err)
		os.Exit(1)
	}
	
	// Display summary metrics
	fmt.Println("📈 Summary Metrics (Last 30 days)")
	fmt.Println(strings.Repeat("-", 30))
	fmt.Printf("Total Projects:      %d\n", report.Summary.TotalProjects)
	fmt.Printf("Total Dependencies:  %d\n", report.Summary.TotalDependencies)
	fmt.Printf("Outdated Packages:   %d\n", report.Summary.OutdatedPackages)
	fmt.Printf("Security Issues:     %d\n", report.Summary.SecurityIssues)
	fmt.Printf("Updates Applied:     %d\n", report.Summary.UpdatesApplied)
	fmt.Printf("Updates Available:   %d\n", report.Summary.UpdatesAvailable)
	fmt.Printf("Avg Update Lag:      %.1f days\n", report.Summary.AverageUpdateLag)
	fmt.Printf("Compliance Score:    %.1f%%\n", report.Summary.ComplianceScore)
	fmt.Println()
	
	// Show available commands
	fmt.Println("🔍 Available Analytics Commands:")
	fmt.Println("  ai-dep-manager report analytics dependencies  - View dependency analytics")
	fmt.Println("  ai-dep-manager report generate summary        - Generate summary report")
	fmt.Println("  ai-dep-manager report generate security       - Generate security report")
	fmt.Println("  ai-dep-manager report generate updates        - Generate updates report")
}

func runDependencyAnalytics(cmd *cobra.Command, args []string) {
	fmt.Println("📦 Dependency Analytics")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()
	
	reportingService := reporting.NewReportingService()
	
	// Parse project filter if provided
	var projectID *uint
	if reportProject != "" {
		// In a real implementation, we'd look up the project ID by name
		// For now, we'll pass nil to get all projects
		projectID = nil
	}
	
	analytics, err := reportingService.GetDependencyAnalytics(context.Background(), projectID)
	if err != nil {
		logger.Error("Failed to get dependency analytics: %v", err)
		os.Exit(1)
	}
	
	// Sort analytics based on the sort flag
	sortAnalytics(analytics, reportSort)
	
	// Limit results
	if reportLimit > 0 && len(analytics) > reportLimit {
		analytics = analytics[:reportLimit]
	}
	
	if len(analytics) == 0 {
		fmt.Println("No dependencies found.")
		return
	}
	
	// Display analytics in a table format
	fmt.Printf("%-30s %-15s %-15s %-8s %-8s %-8s %s\n", 
		"Package", "Current", "Latest", "Lag", "Risk", "Issues", "Recommendation")
	fmt.Println(strings.Repeat("-", 120))
	
	for _, analytic := range analytics {
		riskColor := getRiskColor(analytic.RiskScore)
		lagDays := fmt.Sprintf("%dd", analytic.UpdateLagDays)
		if analytic.UpdateLagDays == 0 {
			lagDays = "0d"
		}
		
		issues := fmt.Sprintf("%d", analytic.SecurityIssues)
		if analytic.SecurityIssues > 0 {
			issues = fmt.Sprintf("🚨%d", analytic.SecurityIssues)
		}
		
		fmt.Printf("%-30s %-15s %-15s %-8s %s%-6.1f%s %-8s %s\n",
			truncateString(analytic.PackageName, 30),
			truncateString(analytic.CurrentVersion, 15),
			truncateString(analytic.LatestVersion, 15),
			lagDays,
			riskColor,
			analytic.RiskScore,
			"\033[0m", // Reset color
			issues,
			truncateString(analytic.RecommendedAction, 40),
		)
	}
	
	fmt.Println()
	fmt.Printf("Showing %d of %d dependencies\n", len(analytics), len(analytics))
	
	// Show summary statistics
	highRisk := 0
	mediumRisk := 0
	lowRisk := 0
	totalSecurityIssues := 0
	
	for _, analytic := range analytics {
		if analytic.RiskScore >= 7.0 {
			highRisk++
		} else if analytic.RiskScore >= 4.0 {
			mediumRisk++
		} else {
			lowRisk++
		}
		totalSecurityIssues += analytic.SecurityIssues
	}
	
	fmt.Println()
	fmt.Println("📊 Risk Summary:")
	fmt.Printf("  High Risk (≥7.0):    %d\n", highRisk)
	fmt.Printf("  Medium Risk (4-7):   %d\n", mediumRisk)
	fmt.Printf("  Low Risk (<4.0):     %d\n", lowRisk)
	fmt.Printf("  Security Issues:     %d\n", totalSecurityIssues)
}

// Helper functions

func parseTimeRange(timeRange string) (reporting.TimeRange, error) {
	now := time.Now()
	
	switch timeRange {
	case "7d":
		return reporting.TimeRange{
			Start: now.AddDate(0, 0, -7),
			End:   now,
		}, nil
	case "30d":
		return reporting.TimeRange{
			Start: now.AddDate(0, 0, -30),
			End:   now,
		}, nil
	case "90d":
		return reporting.TimeRange{
			Start: now.AddDate(0, 0, -90),
			End:   now,
		}, nil
	case "1y":
		return reporting.TimeRange{
			Start: now.AddDate(-1, 0, 0),
			End:   now,
		}, nil
	default:
		// Try to parse custom range YYYY-MM-DD:YYYY-MM-DD
		if strings.Contains(timeRange, ":") {
			parts := strings.Split(timeRange, ":")
			if len(parts) != 2 {
				return reporting.TimeRange{}, fmt.Errorf("invalid custom time range format")
			}
			
			start, err := time.Parse("2006-01-02", parts[0])
			if err != nil {
				return reporting.TimeRange{}, fmt.Errorf("invalid start date: %w", err)
			}
			
			end, err := time.Parse("2006-01-02", parts[1])
			if err != nil {
				return reporting.TimeRange{}, fmt.Errorf("invalid end date: %w", err)
			}
			
			return reporting.TimeRange{Start: start, End: end}, nil
		}
		
		return reporting.TimeRange{}, fmt.Errorf("invalid time range format")
	}
}

func printReport(report *reporting.Report) {
	fmt.Printf("📊 %s\n", report.Title)
	fmt.Printf("📅 Generated: %s\n", report.GeneratedAt.Format("2006-01-02 15:04:05 UTC"))
	fmt.Printf("📆 Period: %s to %s\n", 
		report.TimeRange.Start.Format("2006-01-02"), 
		report.TimeRange.End.Format("2006-01-02"))
	fmt.Printf("📝 %s\n", report.Description)
	fmt.Println()
	
	// Print summary
	fmt.Println("📈 Summary:")
	fmt.Printf("  Total Projects:      %d\n", report.Summary.TotalProjects)
	fmt.Printf("  Total Dependencies:  %d\n", report.Summary.TotalDependencies)
	fmt.Printf("  Outdated Packages:   %d\n", report.Summary.OutdatedPackages)
	fmt.Printf("  Security Issues:     %d\n", report.Summary.SecurityIssues)
	fmt.Printf("  Updates Applied:     %d\n", report.Summary.UpdatesApplied)
	fmt.Printf("  Updates Available:   %d\n", report.Summary.UpdatesAvailable)
	fmt.Printf("  Avg Update Lag:      %.1f days\n", report.Summary.AverageUpdateLag)
	fmt.Printf("  Compliance Score:    %.1f%%\n", report.Summary.ComplianceScore)
	fmt.Println()
	
	// Print charts if available
	if len(report.Charts) > 0 {
		fmt.Println("📊 Charts:")
		for _, chart := range report.Charts {
			fmt.Printf("  %s (%s)\n", chart.Title, chart.Type)
			for i, label := range chart.Labels {
				if i < len(chart.Data) {
					fmt.Printf("    %s: %.1f\n", label, chart.Data[i])
				}
			}
			fmt.Println()
		}
	}
}

func sortAnalytics(analytics []reporting.DependencyAnalytics, sortBy string) {
	// The analytics are already sorted by risk score in the service
	// Additional sorting could be implemented here if needed
}

func getRiskColor(riskScore float64) string {
	if riskScore >= 7.0 {
		return "\033[31m" // Red
	} else if riskScore >= 4.0 {
		return "\033[33m" // Yellow
	}
	return "\033[32m" // Green
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
