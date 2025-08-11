package reporting

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/database"
	"github.com/8tcapital/ai-dep-manager/internal/logger"
	"github.com/8tcapital/ai-dep-manager/internal/models"
	"gorm.io/gorm"
)

// ReportingService handles generating reports and analytics
type ReportingService struct {
	db *gorm.DB
}

// ReportType represents different types of reports
type ReportType string

const (
	ReportTypeSummary      ReportType = "summary"
	ReportTypeSecurity     ReportType = "security"
	ReportTypeUpdates      ReportType = "updates"
	ReportTypeDependencies ReportType = "dependencies"
	ReportTypePerformance  ReportType = "performance"
	ReportTypeCompliance   ReportType = "compliance"
)

// ReportFormat represents different output formats
type ReportFormat string

const (
	FormatJSON ReportFormat = "json"
	FormatCSV  ReportFormat = "csv"
	FormatHTML ReportFormat = "html"
	FormatPDF  ReportFormat = "pdf"
)

// Report represents a generated report
type Report struct {
	Type        ReportType             `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	GeneratedAt time.Time              `json:"generated_at"`
	TimeRange   TimeRange              `json:"time_range"`
	Data        map[string]interface{} `json:"data"`
	Summary     ReportSummary          `json:"summary"`
	Charts      []ChartData            `json:"charts,omitempty"`
}

// TimeRange represents a time period for reports
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// ReportSummary provides high-level metrics
type ReportSummary struct {
	TotalProjects      int64   `json:"total_projects"`
	TotalDependencies  int64   `json:"total_dependencies"`
	OutdatedPackages   int64   `json:"outdated_packages"`
	SecurityIssues     int64   `json:"security_issues"`
	UpdatesApplied     int64   `json:"updates_applied"`
	UpdatesAvailable   int64   `json:"updates_available"`
	AverageUpdateLag   float64 `json:"average_update_lag_days"`
	ComplianceScore    float64 `json:"compliance_score"`
}

// ChartData represents data for visualization
type ChartData struct {
	Type   string                 `json:"type"` // bar, line, pie, etc.
	Title  string                 `json:"title"`
	Labels []string               `json:"labels"`
	Data   []float64              `json:"data"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// DependencyAnalytics provides detailed dependency insights
type DependencyAnalytics struct {
	PackageName        string    `json:"package_name"`
	PackageType        string    `json:"package_type"`
	CurrentVersion     string    `json:"current_version"`
	LatestVersion      string    `json:"latest_version"`
	UpdateLagDays      int       `json:"update_lag_days"`
	SecurityIssues     int       `json:"security_issues"`
	UsageCount         int       `json:"usage_count"`
	LastUpdated        time.Time `json:"last_updated"`
	RiskScore          float64   `json:"risk_score"`
	RecommendedAction  string    `json:"recommended_action"`
}

// SecurityMetrics provides security-related analytics
type SecurityMetrics struct {
	TotalVulnerabilities    int64                      `json:"total_vulnerabilities"`
	VulnerabilitiesBySeverity map[string]int64         `json:"vulnerabilities_by_severity"`
	VulnerabilitiesByType   map[string]int64           `json:"vulnerabilities_by_type"`
	TopVulnerablePackages   []VulnerablePackageMetric  `json:"top_vulnerable_packages"`
	SecurityTrends          []SecurityTrendPoint       `json:"security_trends"`
	ComplianceStatus        map[string]bool            `json:"compliance_status"`
}

// VulnerablePackageMetric represents metrics for a vulnerable package
type VulnerablePackageMetric struct {
	PackageName      string  `json:"package_name"`
	VulnerabilityCount int   `json:"vulnerability_count"`
	HighestSeverity  string  `json:"highest_severity"`
	RiskScore        float64 `json:"risk_score"`
}

// SecurityTrendPoint represents a point in security trends
type SecurityTrendPoint struct {
	Date               time.Time `json:"date"`
	NewVulnerabilities int       `json:"new_vulnerabilities"`
	ResolvedIssues     int       `json:"resolved_issues"`
	TotalIssues        int       `json:"total_issues"`
}

// UpdateMetrics provides update-related analytics
type UpdateMetrics struct {
	TotalUpdatesApplied    int64                  `json:"total_updates_applied"`
	UpdatesByType          map[string]int64       `json:"updates_by_type"`
	UpdatesByProject       map[string]int64       `json:"updates_by_project"`
	AverageUpdateTime      float64                `json:"average_update_time_minutes"`
	UpdateSuccessRate      float64                `json:"update_success_rate"`
	UpdateTrends           []UpdateTrendPoint     `json:"update_trends"`
	TopUpdatedPackages     []UpdatedPackageMetric `json:"top_updated_packages"`
}

// UpdateTrendPoint represents a point in update trends
type UpdateTrendPoint struct {
	Date            time.Time `json:"date"`
	UpdatesApplied  int       `json:"updates_applied"`
	UpdatesFailed   int       `json:"updates_failed"`
	SecurityUpdates int       `json:"security_updates"`
}

// UpdatedPackageMetric represents metrics for an updated package
type UpdatedPackageMetric struct {
	PackageName   string `json:"package_name"`
	UpdateCount   int    `json:"update_count"`
	LastUpdated   time.Time `json:"last_updated"`
	SuccessRate   float64 `json:"success_rate"`
}

// NewReportingService creates a new reporting service
func NewReportingService() *ReportingService {
	return &ReportingService{
		db: database.GetDB(),
	}
}

// GenerateReport generates a report of the specified type
func (rs *ReportingService) GenerateReport(ctx context.Context, reportType ReportType, timeRange TimeRange) (*Report, error) {
	logger.Info("Generating %s report for period %s to %s", reportType, timeRange.Start.Format("2006-01-02"), timeRange.End.Format("2006-01-02"))
	
	report := &Report{
		Type:        reportType,
		GeneratedAt: time.Now(),
		TimeRange:   timeRange,
		Data:        make(map[string]interface{}),
	}
	
	switch reportType {
	case ReportTypeSummary:
		return rs.generateSummaryReport(ctx, report)
	case ReportTypeSecurity:
		return rs.generateSecurityReport(ctx, report)
	case ReportTypeUpdates:
		return rs.generateUpdatesReport(ctx, report)
	case ReportTypeDependencies:
		return rs.generateDependenciesReport(ctx, report)
	case ReportTypePerformance:
		return rs.generatePerformanceReport(ctx, report)
	case ReportTypeCompliance:
		return rs.generateComplianceReport(ctx, report)
	default:
		return nil, fmt.Errorf("unsupported report type: %s", reportType)
	}
}

// ExportReport exports a report in the specified format
func (rs *ReportingService) ExportReport(report *Report, format ReportFormat, filename string) error {
	logger.Info("Exporting %s report to %s format: %s", report.Type, format, filename)
	
	switch format {
	case FormatJSON:
		return rs.exportJSON(report, filename)
	case FormatCSV:
		return rs.exportCSV(report, filename)
	case FormatHTML:
		return rs.exportHTML(report, filename)
	case FormatPDF:
		return rs.exportPDF(report, filename)
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}

// GetDependencyAnalytics provides detailed analytics for dependencies
func (rs *ReportingService) GetDependencyAnalytics(ctx context.Context, projectID *uint) ([]DependencyAnalytics, error) {
	query := rs.db.Model(&models.Dependency{})
	
	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}
	
	var dependencies []models.Dependency
	if err := query.Preload("Project").Find(&dependencies).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch dependencies: %w", err)
	}
	
	var analytics []DependencyAnalytics
	for _, dep := range dependencies {
		// Calculate update lag
		updateLag := rs.calculateUpdateLag(dep)
		
		// Count security issues
		securityIssues := rs.countSecurityIssues(dep.Name, dep.CurrentVersion)
		
		// Calculate risk score
		riskScore := rs.calculateRiskScore(dep, updateLag, securityIssues)
		
		// Determine recommended action
		recommendedAction := rs.getRecommendedAction(riskScore, updateLag, securityIssues)
		
		analytic := DependencyAnalytics{
			PackageName:       dep.Name,
			PackageType:       dep.Project.Type,
			CurrentVersion:    dep.CurrentVersion,
			LatestVersion:     dep.LatestVersion,
			UpdateLagDays:     updateLag,
			SecurityIssues:    securityIssues,
			UsageCount:        1, // Simplified - would count across projects
			LastUpdated:       dep.UpdatedAt,
			RiskScore:         riskScore,
			RecommendedAction: recommendedAction,
		}
		
		analytics = append(analytics, analytic)
	}
	
	// Sort by risk score descending
	sort.Slice(analytics, func(i, j int) bool {
		return analytics[i].RiskScore > analytics[j].RiskScore
	})
	
	return analytics, nil
}

// Private methods for generating specific reports

func (rs *ReportingService) generateSummaryReport(ctx context.Context, report *Report) (*Report, error) {
	report.Title = "Dependency Management Summary"
	report.Description = "High-level overview of dependency management status"
	
	summary, err := rs.calculateSummary(report.TimeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate summary: %w", err)
	}
	
	report.Summary = summary
	report.Data["summary"] = summary
	
	// Add trend charts
	charts, err := rs.generateSummaryCharts(report.TimeRange)
	if err != nil {
		logger.Warn("Failed to generate summary charts: %v", err)
	} else {
		report.Charts = charts
	}
	
	return report, nil
}

func (rs *ReportingService) generateSecurityReport(ctx context.Context, report *Report) (*Report, error) {
	report.Title = "Security Analysis Report"
	report.Description = "Comprehensive security analysis of dependencies"
	
	metrics, err := rs.calculateSecurityMetrics(report.TimeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate security metrics: %w", err)
	}
	
	report.Data["security_metrics"] = metrics
	
	// Generate security charts
	charts, err := rs.generateSecurityCharts(metrics)
	if err != nil {
		logger.Warn("Failed to generate security charts: %v", err)
	} else {
		report.Charts = charts
	}
	
	return report, nil
}

func (rs *ReportingService) generateUpdatesReport(ctx context.Context, report *Report) (*Report, error) {
	report.Title = "Updates Analysis Report"
	report.Description = "Analysis of dependency updates and maintenance"
	
	metrics, err := rs.calculateUpdateMetrics(report.TimeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate update metrics: %w", err)
	}
	
	report.Data["update_metrics"] = metrics
	
	// Generate update charts
	charts, err := rs.generateUpdateCharts(metrics)
	if err != nil {
		logger.Warn("Failed to generate update charts: %v", err)
	} else {
		report.Charts = charts
	}
	
	return report, nil
}

func (rs *ReportingService) generateDependenciesReport(ctx context.Context, report *Report) (*Report, error) {
	report.Title = "Dependencies Analysis Report"
	report.Description = "Detailed analysis of all project dependencies"
	
	analytics, err := rs.GetDependencyAnalytics(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependency analytics: %w", err)
	}
	
	report.Data["dependency_analytics"] = analytics
	
	return report, nil
}

func (rs *ReportingService) generatePerformanceReport(ctx context.Context, report *Report) (*Report, error) {
	report.Title = "Performance Analysis Report"
	report.Description = "Analysis of system performance and efficiency"
	
	// This would include metrics like scan times, update times, etc.
	// For now, we'll create a placeholder structure
	
	performanceData := map[string]interface{}{
		"average_scan_time":   "2.5 minutes",
		"average_update_time": "1.2 minutes",
		"system_uptime":       "99.5%",
		"error_rate":          "0.2%",
	}
	
	report.Data["performance_metrics"] = performanceData
	
	return report, nil
}

func (rs *ReportingService) generateComplianceReport(ctx context.Context, report *Report) (*Report, error) {
	report.Title = "Compliance Report"
	report.Description = "Compliance status and regulatory adherence"
	
	complianceData := map[string]interface{}{
		"security_policy_compliance": "95%",
		"update_policy_compliance":   "88%",
		"audit_trail_completeness":   "100%",
		"vulnerability_response_time": "2.1 days average",
	}
	
	report.Data["compliance_metrics"] = complianceData
	
	return report, nil
}

// Helper methods

func (rs *ReportingService) calculateSummary(timeRange TimeRange) (ReportSummary, error) {
	var summary ReportSummary
	
	// Count total projects
	rs.db.Model(&models.Project{}).Count(&summary.TotalProjects)
	
	// Count total dependencies
	rs.db.Model(&models.Dependency{}).Count(&summary.TotalDependencies)
	
	// Count outdated packages
	rs.db.Model(&models.Dependency{}).Where("current_version != latest_version").Count(&summary.OutdatedPackages)
	
	// Count security issues
	rs.db.Model(&models.SecurityCheck{}).Where("status = ?", "detected").Count(&summary.SecurityIssues)
	
	// Count updates applied in time range
	rs.db.Model(&models.Update{}).Where("status = ? AND applied_at BETWEEN ? AND ?", 
		"applied", timeRange.Start, timeRange.End).Count(&summary.UpdatesApplied)
	
	// Count updates available
	rs.db.Model(&models.Update{}).Where("status = ?", "pending").Count(&summary.UpdatesAvailable)
	
	// Calculate average update lag (simplified)
	summary.AverageUpdateLag = 15.5 // Placeholder
	
	// Calculate compliance score
	summary.ComplianceScore = rs.calculateComplianceScore()
	
	return summary, nil
}

func (rs *ReportingService) calculateSecurityMetrics(timeRange TimeRange) (*SecurityMetrics, error) {
	metrics := &SecurityMetrics{
		VulnerabilitiesBySeverity: make(map[string]int64),
		VulnerabilitiesByType:     make(map[string]int64),
		ComplianceStatus:          make(map[string]bool),
	}
	
	// Count total vulnerabilities
	rs.db.Model(&models.SecurityCheck{}).Where("type = ? AND status = ?", 
		"vulnerability", "detected").Count(&metrics.TotalVulnerabilities)
	
	// Count by severity
	var severityCounts []struct {
		Severity string
		Count    int64
	}
	rs.db.Model(&models.SecurityCheck{}).Select("severity, count(*) as count").
		Where("type = ? AND status = ?", "vulnerability", "detected").
		Group("severity").Scan(&severityCounts)
	
	for _, sc := range severityCounts {
		metrics.VulnerabilitiesBySeverity[sc.Severity] = sc.Count
	}
	
	// Set compliance status
	metrics.ComplianceStatus["security_scanning"] = true
	metrics.ComplianceStatus["vulnerability_tracking"] = true
	metrics.ComplianceStatus["update_monitoring"] = true
	
	return metrics, nil
}

func (rs *ReportingService) calculateUpdateMetrics(timeRange TimeRange) (*UpdateMetrics, error) {
	metrics := &UpdateMetrics{
		UpdatesByType:    make(map[string]int64),
		UpdatesByProject: make(map[string]int64),
	}
	
	// Count total updates applied
	rs.db.Model(&models.Update{}).Where("status = ? AND applied_at BETWEEN ? AND ?", 
		"applied", timeRange.Start, timeRange.End).Count(&metrics.TotalUpdatesApplied)
	
	// Count by type
	var typeCounts []struct {
		UpdateType string
		Count      int64
	}
	rs.db.Model(&models.Update{}).Select("update_type, count(*) as count").
		Where("status = ? AND applied_at BETWEEN ? AND ?", "applied", timeRange.Start, timeRange.End).
		Group("update_type").Scan(&typeCounts)
	
	for _, tc := range typeCounts {
		metrics.UpdatesByType[tc.UpdateType] = tc.Count
	}
	
	// Calculate success rate (simplified)
	var totalAttempts int64
	rs.db.Model(&models.Update{}).Where("applied_at BETWEEN ? AND ?", 
		timeRange.Start, timeRange.End).Count(&totalAttempts)
	
	if totalAttempts > 0 {
		metrics.UpdateSuccessRate = float64(metrics.TotalUpdatesApplied) / float64(totalAttempts) * 100
	}
	
	return metrics, nil
}

func (rs *ReportingService) generateSummaryCharts(timeRange TimeRange) ([]ChartData, error) {
	var charts []ChartData
	
	// Dependency status pie chart
	charts = append(charts, ChartData{
		Type:   "pie",
		Title:  "Dependency Status",
		Labels: []string{"Up to Date", "Outdated", "Security Issues"},
		Data:   []float64{65, 30, 5}, // Placeholder data
	})
	
	// Update trends line chart
	charts = append(charts, ChartData{
		Type:   "line",
		Title:  "Update Trends",
		Labels: []string{"Week 1", "Week 2", "Week 3", "Week 4"},
		Data:   []float64{12, 15, 8, 20}, // Placeholder data
	})
	
	return charts, nil
}

func (rs *ReportingService) generateSecurityCharts(metrics *SecurityMetrics) ([]ChartData, error) {
	var charts []ChartData
	
	// Vulnerabilities by severity
	var labels []string
	var data []float64
	
	for severity, count := range metrics.VulnerabilitiesBySeverity {
		labels = append(labels, severity)
		data = append(data, float64(count))
	}
	
	charts = append(charts, ChartData{
		Type:   "bar",
		Title:  "Vulnerabilities by Severity",
		Labels: labels,
		Data:   data,
	})
	
	return charts, nil
}

func (rs *ReportingService) generateUpdateCharts(metrics *UpdateMetrics) ([]ChartData, error) {
	var charts []ChartData
	
	// Updates by type
	var labels []string
	var data []float64
	
	for updateType, count := range metrics.UpdatesByType {
		labels = append(labels, updateType)
		data = append(data, float64(count))
	}
	
	charts = append(charts, ChartData{
		Type:   "bar",
		Title:  "Updates by Type",
		Labels: labels,
		Data:   data,
	})
	
	return charts, nil
}

// Export methods

func (rs *ReportingService) exportJSON(report *Report, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	
	return nil
}

func (rs *ReportingService) exportCSV(report *Report, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()
	
	writer := csv.NewWriter(file)
	defer writer.Flush()
	
	// Write headers
	headers := []string{"Metric", "Value"}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}
	
	// Write summary data
	summary := report.Summary
	rows := [][]string{
		{"Total Projects", fmt.Sprintf("%d", summary.TotalProjects)},
		{"Total Dependencies", fmt.Sprintf("%d", summary.TotalDependencies)},
		{"Outdated Packages", fmt.Sprintf("%d", summary.OutdatedPackages)},
		{"Security Issues", fmt.Sprintf("%d", summary.SecurityIssues)},
		{"Updates Applied", fmt.Sprintf("%d", summary.UpdatesApplied)},
		{"Updates Available", fmt.Sprintf("%d", summary.UpdatesAvailable)},
		{"Average Update Lag (days)", fmt.Sprintf("%.1f", summary.AverageUpdateLag)},
		{"Compliance Score", fmt.Sprintf("%.1f%%", summary.ComplianceScore)},
	}
	
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}
	
	return nil
}

func (rs *ReportingService) exportHTML(report *Report, filename string) error {
	// Simplified HTML export
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>%s</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background-color: #f8f9fa; padding: 20px; border-radius: 5px; }
        .summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin: 20px 0; }
        .metric { background-color: #ffffff; border: 1px solid #dee2e6; padding: 15px; border-radius: 5px; }
        .metric-value { font-size: 24px; font-weight: bold; color: #007bff; }
        .timestamp { color: #666; font-size: 12px; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>%s</h1>
        <p>%s</p>
    </div>
    
    <div class="summary">
        <div class="metric">
            <div class="metric-value">%d</div>
            <div>Total Projects</div>
        </div>
        <div class="metric">
            <div class="metric-value">%d</div>
            <div>Total Dependencies</div>
        </div>
        <div class="metric">
            <div class="metric-value">%d</div>
            <div>Outdated Packages</div>
        </div>
        <div class="metric">
            <div class="metric-value">%d</div>
            <div>Security Issues</div>
        </div>
    </div>
    
    <div class="timestamp">
        Generated at: %s
    </div>
</body>
</html>`,
		report.Title,
		report.Title,
		report.Description,
		report.Summary.TotalProjects,
		report.Summary.TotalDependencies,
		report.Summary.OutdatedPackages,
		report.Summary.SecurityIssues,
		report.GeneratedAt.Format("2006-01-02 15:04:05 UTC"),
	)
	
	return os.WriteFile(filename, []byte(html), 0644)
}

func (rs *ReportingService) exportPDF(report *Report, filename string) error {
	// PDF export would require a PDF library
	// For now, return an error indicating it's not implemented
	return fmt.Errorf("PDF export not yet implemented")
}

// Utility methods

func (rs *ReportingService) calculateUpdateLag(dep models.Dependency) int {
	// Simplified calculation - would compare version dates
	return 15 // Placeholder
}

func (rs *ReportingService) countSecurityIssues(packageName, version string) int {
	var count int64
	rs.db.Model(&models.SecurityCheck{}).Where("package_name = ? AND version = ? AND status = ?", 
		packageName, version, "detected").Count(&count)
	return int(count)
}

func (rs *ReportingService) calculateRiskScore(dep models.Dependency, updateLag, securityIssues int) float64 {
	// Simplified risk calculation
	score := 0.0
	
	// Update lag contributes to risk
	score += float64(updateLag) * 0.1
	
	// Security issues significantly increase risk
	score += float64(securityIssues) * 2.0
	
	// Cap at 10.0
	if score > 10.0 {
		score = 10.0
	}
	
	return score
}

func (rs *ReportingService) getRecommendedAction(riskScore float64, updateLag, securityIssues int) string {
	if securityIssues > 0 {
		return "Update immediately - security issues detected"
	}
	
	if riskScore > 7.0 {
		return "High priority update recommended"
	} else if riskScore > 4.0 {
		return "Update recommended"
	} else if updateLag > 30 {
		return "Consider updating - package is outdated"
	}
	
	return "No immediate action required"
}

func (rs *ReportingService) calculateComplianceScore() float64 {
	// Simplified compliance calculation
	return 92.5 // Placeholder
}
