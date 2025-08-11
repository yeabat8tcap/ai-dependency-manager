#!/bin/bash

# AI Dependency Manager - Component Generation Script
# Creates all missing component stubs for successful Angular build

echo "🚀 Creating missing Angular component stubs..."

# Create component directories
mkdir -p src/app/features/projects/{project-list,project-create,project-detail,project-edit,project-dependencies,project-settings}
mkdir -p src/app/features/security/{security-dashboard,vulnerability-list,vulnerability-detail,security-policies,compliance-report,security-alerts}
mkdir -p src/app/features/analytics/{analytics-dashboard,dependency-lag,update-trends,performance-metrics,analytics-reports}
mkdir -p src/app/features/policies/{policy-list,policy-create,policy-templates,policy-detail,policy-edit,policy-test,policy-compliance}

# Function to create a basic component stub
create_component() {
    local path=$1
    local name=$2
    local selector=$3
    local title=$4
    
    cat > "$path" << EOF
import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';

@Component({
  selector: '$selector',
  standalone: true,
  imports: [CommonModule, MatCardModule],
  template: \`
    <div class="component-container">
      <mat-card>
        <mat-card-header>
          <mat-card-title>$title</mat-card-title>
        </mat-card-header>
        <mat-card-content>
          <p>$title component - Coming soon!</p>
        </mat-card-content>
      </mat-card>
    </div>
  \`,
  styles: [\`
    .component-container {
      padding: 24px;
    }
  \`]
})
export class $name {
}
EOF
}

# Projects components
create_component "src/app/features/projects/project-list/project-list.component.ts" "ProjectListComponent" "app-project-list" "Project List"
create_component "src/app/features/projects/project-create/project-create.component.ts" "ProjectCreateComponent" "app-project-create" "Create Project"
create_component "src/app/features/projects/project-detail/project-detail.component.ts" "ProjectDetailComponent" "app-project-detail" "Project Details"
create_component "src/app/features/projects/project-edit/project-edit.component.ts" "ProjectEditComponent" "app-project-edit" "Edit Project"
create_component "src/app/features/projects/project-dependencies/project-dependencies.component.ts" "ProjectDependenciesComponent" "app-project-dependencies" "Project Dependencies"
create_component "src/app/features/projects/project-settings/project-settings.component.ts" "ProjectSettingsComponent" "app-project-settings" "Project Settings"

# Security components
create_component "src/app/features/security/security-dashboard/security-dashboard.component.ts" "SecurityDashboardComponent" "app-security-dashboard" "Security Dashboard"
create_component "src/app/features/security/vulnerability-list/vulnerability-list.component.ts" "VulnerabilityListComponent" "app-vulnerability-list" "Vulnerabilities"
create_component "src/app/features/security/vulnerability-detail/vulnerability-detail.component.ts" "VulnerabilityDetailComponent" "app-vulnerability-detail" "Vulnerability Details"
create_component "src/app/features/security/security-policies/security-policies.component.ts" "SecurityPoliciesComponent" "app-security-policies" "Security Policies"
create_component "src/app/features/security/compliance-report/compliance-report.component.ts" "ComplianceReportComponent" "app-compliance-report" "Compliance Report"
create_component "src/app/features/security/security-alerts/security-alerts.component.ts" "SecurityAlertsComponent" "app-security-alerts" "Security Alerts"

# Analytics components
create_component "src/app/features/analytics/analytics-dashboard/analytics-dashboard.component.ts" "AnalyticsDashboardComponent" "app-analytics-dashboard" "Analytics Dashboard"
create_component "src/app/features/analytics/dependency-lag/dependency-lag.component.ts" "DependencyLagComponent" "app-dependency-lag" "Dependency Lag Analysis"
create_component "src/app/features/analytics/update-trends/update-trends.component.ts" "UpdateTrendsComponent" "app-update-trends" "Update Trends"
create_component "src/app/features/analytics/performance-metrics/performance-metrics.component.ts" "PerformanceMetricsComponent" "app-performance-metrics" "Performance Metrics"
create_component "src/app/features/analytics/analytics-reports/analytics-reports.component.ts" "AnalyticsReportsComponent" "app-analytics-reports" "Analytics Reports"

# Policies components
create_component "src/app/features/policies/policy-list/policy-list.component.ts" "PolicyListComponent" "app-policy-list" "Policy List"
create_component "src/app/features/policies/policy-create/policy-create.component.ts" "PolicyCreateComponent" "app-policy-create" "Create Policy"
create_component "src/app/features/policies/policy-templates/policy-templates.component.ts" "PolicyTemplatesComponent" "app-policy-templates" "Policy Templates"
create_component "src/app/features/policies/policy-detail/policy-detail.component.ts" "PolicyDetailComponent" "app-policy-detail" "Policy Details"
create_component "src/app/features/policies/policy-edit/policy-edit.component.ts" "PolicyEditComponent" "app-policy-edit" "Edit Policy"
create_component "src/app/features/policies/policy-test/policy-test.component.ts" "PolicyTestComponent" "app-policy-test" "Test Policy"
create_component "src/app/features/policies/policy-compliance/policy-compliance.component.ts" "PolicyComplianceComponent" "app-policy-compliance" "Policy Compliance"

echo "✅ All component stubs created successfully!"
echo "🎯 Total components created: 25"
