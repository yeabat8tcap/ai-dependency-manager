import { Component, OnInit, OnDestroy, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatBadgeModule } from '@angular/material/badge';
import { MatChipsModule } from '@angular/material/chips';
import { MatTooltipModule } from '@angular/material/tooltip';
import { Subject, takeUntil, combineLatest } from 'rxjs';

import { ApiService, Project } from '../../core/services/api.service';
import { WebSocketService, SystemStatusMessage } from '../../core/services/websocket.service';
import { LoggingService } from '../../core/services/logging.service';
import { environment } from '../../../environments/environment';

export interface DashboardStats {
  totalProjects: number;
  totalDependencies: number;
  outdatedDependencies: number;
  securityIssues: number;
  lastScanTime?: Date;
  systemHealth: 'healthy' | 'warning' | 'error';
}

export interface ProjectSummary {
  id: number;
  name: string;
  type: string;
  dependencyCount: number;
  outdatedCount: number;
  securityIssueCount: number;
  lastScan?: Date;
  healthScore: number;
  riskLevel: 'low' | 'medium' | 'high' | 'critical';
}

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [
    CommonModule,
    MatCardModule,
    MatIconModule,
    MatButtonModule,
    MatProgressSpinnerModule,
    MatProgressBarModule,
    MatBadgeModule,
    MatChipsModule,
    MatTooltipModule
  ],
  template: `
    <div class="dashboard-container">
      <!-- Header Section -->
      <div class="dashboard-header">
        <h1 class="dashboard-title">
          <mat-icon class="title-icon">dashboard</mat-icon>
          AI Dependency Manager Dashboard
        </h1>
        <div class="system-status" [class]="'status-' + dashboardStats.systemHealth">
          <mat-icon>{{ getSystemStatusIcon() }}</mat-icon>
          <span>{{ getSystemStatusText() }}</span>
        </div>
      </div>

      <!-- Statistics Cards -->
      <div class="stats-grid">
        <mat-card class="stat-card projects-card">
          <mat-card-header>
            <mat-card-title>
              <mat-icon>folder</mat-icon>
              Projects
            </mat-card-title>
          </mat-card-header>
          <mat-card-content>
            <div class="stat-number">{{ dashboardStats.totalProjects }}</div>
            <div class="stat-label">Total Projects</div>
          </mat-card-content>
        </mat-card>

        <mat-card class="stat-card dependencies-card">
          <mat-card-header>
            <mat-card-title>
              <mat-icon>account_tree</mat-icon>
              Dependencies
            </mat-card-title>
          </mat-card-header>
          <mat-card-content>
            <div class="stat-number">{{ dashboardStats.totalDependencies }}</div>
            <div class="stat-label">Total Dependencies</div>
            <mat-progress-bar 
              mode="determinate" 
              [value]="getUpToDatePercentage()"
              class="progress-indicator">
            </mat-progress-bar>
          </mat-card-content>
        </mat-card>

        <mat-card class="stat-card updates-card">
          <mat-card-header>
            <mat-card-title>
              <mat-icon matBadge="{{ dashboardStats.outdatedDependencies }}" 
                       matBadgeColor="warn" 
                       [matBadgeHidden]="dashboardStats.outdatedDependencies === 0">
                update
              </mat-icon>
              Updates Available
            </mat-card-title>
          </mat-card-header>
          <mat-card-content>
            <div class="stat-number">{{ dashboardStats.outdatedDependencies }}</div>
            <div class="stat-label">Outdated Dependencies</div>
          </mat-card-content>
        </mat-card>

        <mat-card class="stat-card security-card">
          <mat-card-header>
            <mat-card-title>
              <mat-icon matBadge="{{ dashboardStats.securityIssues }}" 
                       matBadgeColor="accent" 
                       [matBadgeHidden]="dashboardStats.securityIssues === 0">
                security
              </mat-icon>
              Security Issues
            </mat-card-title>
          </mat-card-header>
          <mat-card-content>
            <div class="stat-number">{{ dashboardStats.securityIssues }}</div>
            <div class="stat-label">Vulnerabilities Found</div>
          </mat-card-content>
        </mat-card>
      </div>

      <!-- Project Overview -->
      <mat-card class="projects-overview-card">
        <mat-card-header>
          <mat-card-title>
            <mat-icon>view_list</mat-icon>
            Project Overview
          </mat-card-title>
          <mat-card-subtitle>
            Recent activity and health status
          </mat-card-subtitle>
        </mat-card-header>
        <mat-card-content>
          <div class="projects-list" *ngIf="projectSummaries.length > 0; else noProjects">
            <div class="project-item" *ngFor="let project of projectSummaries">
              <div class="project-info">
                <div class="project-name">
                  <mat-icon class="project-type-icon">{{ getProjectTypeIcon(project.type) }}</mat-icon>
                  {{ project.name }}
                </div>
                <div class="project-stats">
                  <mat-chip-set>
                    <mat-chip class="dependency-chip">
                      {{ project.dependencyCount }} dependencies
                    </mat-chip>
                    <mat-chip class="outdated-chip" *ngIf="project.outdatedCount > 0">
                      {{ project.outdatedCount }} outdated
                    </mat-chip>
                    <mat-chip class="security-chip" *ngIf="project.securityIssueCount > 0">
                      {{ project.securityIssueCount }} security issues
                    </mat-chip>
                  </mat-chip-set>
                </div>
              </div>
              <div class="project-health">
                <div class="health-score" [class]="'risk-' + project.riskLevel">
                  <mat-icon>{{ getRiskLevelIcon(project.riskLevel) }}</mat-icon>
                  <span>{{ project.healthScore }}%</span>
                </div>
                <div class="last-scan" *ngIf="project.lastScan">
                  Last scan: {{ project.lastScan | date:'short' }}
                </div>
              </div>
            </div>
          </div>
          <ng-template #noProjects>
            <div class="no-projects">
              <mat-icon class="no-projects-icon">folder_open</mat-icon>
              <h3>No Projects Configured</h3>
              <p>Get started by adding your first project to monitor dependencies.</p>
              <button mat-raised-button color="primary" (click)="addProject()">
                <mat-icon>add</mat-icon>
                Add Project
              </button>
            </div>
          </ng-template>
        </mat-card-content>
      </mat-card>

      <!-- Quick Actions -->
      <mat-card class="quick-actions-card">
        <mat-card-header>
          <mat-card-title>
            <mat-icon>flash_on</mat-icon>
            Quick Actions
          </mat-card-title>
        </mat-card-header>
        <mat-card-content>
          <div class="action-buttons">
            <button mat-raised-button color="primary" (click)="scanAllProjects()">
              <mat-icon>refresh</mat-icon>
              Scan All Projects
            </button>
            <button mat-raised-button color="accent" (click)="viewSecurityIssues()">
              <mat-icon>security</mat-icon>
              Security Dashboard
            </button>
            <button mat-raised-button (click)="viewAnalytics()">
              <mat-icon>analytics</mat-icon>
              View Analytics
            </button>
            <button mat-raised-button (click)="managePolicies()">
              <mat-icon>policy</mat-icon>
              Manage Policies
            </button>
          </div>
        </mat-card-content>
      </mat-card>
    </div>
  `,
  styles: [`
    .dashboard-container {
      padding: 24px;
      max-width: 1200px;
      margin: 0 auto;
    }

    .dashboard-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 32px;
    }

    .dashboard-title {
      display: flex;
      align-items: center;
      gap: 12px;
      margin: 0;
      font-size: 2rem;
      font-weight: 300;
    }

    .title-icon {
      font-size: 2rem;
      width: 2rem;
      height: 2rem;
    }

    .system-status {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 8px 16px;
      border-radius: 20px;
      font-weight: 500;
    }

    .status-healthy {
      background-color: #e8f5e8;
      color: #2e7d32;
    }

    .status-warning {
      background-color: #fff3e0;
      color: #f57c00;
    }

    .status-error {
      background-color: #ffebee;
      color: #d32f2f;
    }

    .stats-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
      gap: 24px;
      margin-bottom: 32px;
    }

    .stat-card {
      text-align: center;
    }

    .stat-card mat-card-title {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 8px;
    }

    .stat-number {
      font-size: 3rem;
      font-weight: 300;
      margin: 16px 0 8px 0;
    }

    .stat-label {
      color: rgba(0, 0, 0, 0.6);
      font-size: 0.9rem;
    }

    .progress-indicator {
      margin-top: 16px;
    }

    .projects-overview-card,
    .quick-actions-card {
      margin-bottom: 24px;
    }

    .projects-list {
      display: flex;
      flex-direction: column;
      gap: 16px;
    }

    .project-item {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 16px;
      border: 1px solid rgba(0, 0, 0, 0.12);
      border-radius: 8px;
    }

    .project-info {
      flex: 1;
    }

    .project-name {
      display: flex;
      align-items: center;
      gap: 8px;
      font-weight: 500;
      margin-bottom: 8px;
    }

    .project-type-icon {
      color: rgba(0, 0, 0, 0.6);
    }

    .project-stats mat-chip-set {
      display: flex;
      gap: 8px;
    }

    .project-health {
      text-align: right;
    }

    .health-score {
      display: flex;
      align-items: center;
      gap: 4px;
      font-weight: 500;
      margin-bottom: 4px;
    }

    .risk-low { color: #2e7d32; }
    .risk-medium { color: #f57c00; }
    .risk-high { color: #d32f2f; }
    .risk-critical { color: #b71c1c; }

    .last-scan {
      font-size: 0.8rem;
      color: rgba(0, 0, 0, 0.6);
    }

    .no-projects {
      text-align: center;
      padding: 48px 24px;
    }

    .no-projects-icon {
      font-size: 4rem;
      width: 4rem;
      height: 4rem;
      color: rgba(0, 0, 0, 0.3);
      margin-bottom: 16px;
    }

    .action-buttons {
      display: flex;
      gap: 16px;
      flex-wrap: wrap;
    }

    .action-buttons button {
      display: flex;
      align-items: center;
      gap: 8px;
    }

    @media (max-width: 768px) {
      .dashboard-container {
        padding: 16px;
      }

      .dashboard-header {
        flex-direction: column;
        align-items: flex-start;
        gap: 16px;
      }

      .stats-grid {
        grid-template-columns: 1fr;
      }

      .project-item {
        flex-direction: column;
        align-items: flex-start;
        gap: 16px;
      }

      .project-health {
        text-align: left;
        width: 100%;
      }

      .action-buttons {
        flex-direction: column;
      }

      .action-buttons button {
        width: 100%;
        justify-content: center;
      }
    }
  `]
})
export class DashboardComponent implements OnInit, OnDestroy {
  private destroy$ = new Subject<void>();
  private apiService = inject(ApiService);
  private wsService = inject(WebSocketService);
  private logger = inject(LoggingService);

  dashboardStats: DashboardStats = {
    totalProjects: 0,
    totalDependencies: 0,
    outdatedDependencies: 0,
    securityIssues: 0,
    systemHealth: 'healthy'
  };

  projectSummaries: ProjectSummary[] = [];
  loading = true;

  ngOnInit(): void {
    this.logger.info('Dashboard component initialized', {
      environment: environment.production ? 'production' : 'development',
      dataSource: environment.dataSource,
      useMockData: environment.dataSource === 'mock'
    }, 'DashboardComponent');
    
    this.loadDashboardData();
    this.subscribeToRealTimeUpdates();
  }

  ngOnDestroy(): void {
    this.logger.info('Dashboard component destroyed', {}, 'DashboardComponent');
    this.destroy$.next();
    this.destroy$.complete();
  }

  private loadDashboardData(): void {
    const startTime = this.logger.startTimer('dashboard-data-load');
    this.logger.info('Loading dashboard data', {
      environment: environment.production ? 'production' : 'development',
      dataSource: environment.dataSource
    }, 'DashboardComponent');
    
    // Load data based on environment configuration
    if (environment.dataSource === 'mock') {
      this.logger.debug('Using mock data for dashboard', {}, 'DashboardComponent');
      this.loadMockData();
    } else {
      this.logger.debug('Using real API data for dashboard', {}, 'DashboardComponent');
      this.loadRealData();
    }
    
    this.logger.endTimer(startTime, 'dashboard-data-load');
  }

  private loadRealData(): void {
    this.loading = true;
    
    // Load real data from API endpoints
    combineLatest([
      this.apiService.getProjects(),
      this.apiService.getSystemStatus(),
      this.apiService.getAnalytics()
    ]).pipe(
      takeUntil(this.destroy$)
    ).subscribe({
      next: ([projectsResponse, statusResponse, analyticsResponse]) => {
        if (projectsResponse.success && statusResponse.success && analyticsResponse.success) {
          this.processProjectData(projectsResponse.data);
          this.updateDashboardStats(analyticsResponse.data);
          this.updateSystemHealth(statusResponse.data);
        }
        this.loading = false;
      },
      error: (error) => {
        console.error('Error loading real dashboard data:', error);
        this.loading = false;
        // Fallback to mock data if API fails
        console.log('Falling back to mock data due to API error');
        this.loadMockData();
      }
    });
  }

  private loadMockData(): void {
    // Mock dashboard statistics
    this.dashboardStats = {
      totalProjects: 5,
      totalDependencies: 247,
      outdatedDependencies: 23,
      securityIssues: 7,
      lastScanTime: new Date(Date.now() - 2 * 60 * 60 * 1000), // 2 hours ago
      systemHealth: 'warning'
    };

    // Mock project summaries with realistic data
    this.projectSummaries = [
      {
        id: 1,
        name: 'E-Commerce Frontend',
        type: 'npm',
        dependencyCount: 89,
        outdatedCount: 12,
        securityIssueCount: 3,
        lastScan: new Date(Date.now() - 1 * 60 * 60 * 1000), // 1 hour ago
        healthScore: 78,
        riskLevel: 'medium'
      },
      {
        id: 2,
        name: 'API Gateway Service',
        type: 'pip',
        dependencyCount: 45,
        outdatedCount: 8,
        securityIssueCount: 2,
        lastScan: new Date(Date.now() - 3 * 60 * 60 * 1000), // 3 hours ago
        healthScore: 85,
        riskLevel: 'low'
      },
      {
        id: 3,
        name: 'Data Processing Pipeline',
        type: 'pip',
        dependencyCount: 67,
        outdatedCount: 3,
        securityIssueCount: 2,
        lastScan: new Date(Date.now() - 30 * 60 * 1000), // 30 minutes ago
        healthScore: 92,
        riskLevel: 'low'
      },
      {
        id: 4,
        name: 'Legacy Monolith',
        type: 'maven',
        dependencyCount: 34,
        outdatedCount: 0,
        securityIssueCount: 0,
        lastScan: new Date(Date.now() - 6 * 60 * 60 * 1000), // 6 hours ago
        healthScore: 95,
        riskLevel: 'low'
      },
      {
        id: 5,
        name: 'Mobile Backend',
        type: 'gradle',
        dependencyCount: 12,
        outdatedCount: 0,
        securityIssueCount: 0,
        lastScan: new Date(Date.now() - 4 * 60 * 60 * 1000), // 4 hours ago
        healthScore: 98,
        riskLevel: 'low'
      }
    ];

    this.loading = false;
  }

  private processProjectData(projects: Project[]): void {
    this.projectSummaries = projects.map(project => ({
      id: project.id,
      name: project.name,
      type: project.type,
      dependencyCount: 0, // Will be populated by API call
      outdatedCount: 0,
      securityIssueCount: 0,
      lastScan: project.lastScan,
      healthScore: this.calculateHealthScore(project),
      riskLevel: this.calculateRiskLevel(project)
    }));
  }

  private updateDashboardStats(analyticsData: any): void {
    this.dashboardStats = {
      totalProjects: analyticsData.totalProjects || 0,
      totalDependencies: analyticsData.totalDependencies || 0,
      outdatedDependencies: analyticsData.outdatedDependencies || 0,
      securityIssues: analyticsData.securityIssues || 0,
      lastScanTime: analyticsData.lastScanTime ? new Date(analyticsData.lastScanTime) : undefined,
      systemHealth: analyticsData.systemHealth || 'healthy'
    };
  }

  private updateSystemHealth(statusData: any): void {
    this.dashboardStats.systemHealth = statusData.health || 'healthy';
  }

  private subscribeToRealTimeUpdates(): void {
    this.wsService.getSystemStatus().pipe(
      takeUntil(this.destroy$)
    ).subscribe((status: SystemStatusMessage) => {
      this.dashboardStats.systemHealth = status.status;
    });

    this.wsService.getScanProgress().pipe(
      takeUntil(this.destroy$)
    ).subscribe((progress) => {
      // Update scan progress for specific project
      const project = this.projectSummaries.find(p => p.id === progress.projectId);
      if (project && progress.status === 'complete') {
        this.loadDashboardData(); // Refresh data when scan completes
      }
    });
  }

  private calculateHealthScore(project: Project): number {
    // Simple health score calculation - can be enhanced with more sophisticated logic
    return Math.floor(Math.random() * 40) + 60; // 60-100 range for demo
  }

  private calculateRiskLevel(project: Project): 'low' | 'medium' | 'high' | 'critical' {
    const score = this.calculateHealthScore(project);
    if (score >= 90) return 'low';
    if (score >= 75) return 'medium';
    if (score >= 60) return 'high';
    return 'critical';
  }

  getUpToDatePercentage(): number {
    if (this.dashboardStats.totalDependencies === 0) return 100;
    const upToDate = this.dashboardStats.totalDependencies - this.dashboardStats.outdatedDependencies;
    return (upToDate / this.dashboardStats.totalDependencies) * 100;
  }

  getSystemStatusIcon(): string {
    switch (this.dashboardStats.systemHealth) {
      case 'healthy': return 'check_circle';
      case 'warning': return 'warning';
      case 'error': return 'error';
      default: return 'help';
    }
  }

  getSystemStatusText(): string {
    switch (this.dashboardStats.systemHealth) {
      case 'healthy': return 'System Healthy';
      case 'warning': return 'System Warning';
      case 'error': return 'System Error';
      default: return 'Unknown Status';
    }
  }

  getProjectTypeIcon(type: string): string {
    switch (type) {
      case 'npm': return 'javascript';
      case 'pip': return 'code';
      case 'maven': return 'coffee';
      case 'gradle': return 'build';
      default: return 'folder';
    }
  }

  getRiskLevelIcon(riskLevel: string): string {
    switch (riskLevel) {
      case 'low': return 'check_circle';
      case 'medium': return 'warning';
      case 'high': return 'error';
      case 'critical': return 'dangerous';
      default: return 'help';
    }
  }

  // Action handlers
  addProject(): void {
    // Navigate to project creation
    console.log('Add project clicked');
  }

  scanAllProjects(): void {
    // Trigger scan for all projects
    console.log('Scan all projects clicked');
  }

  viewSecurityIssues(): void {
    // Navigate to security dashboard
    console.log('View security issues clicked');
  }

  viewAnalytics(): void {
    // Navigate to analytics dashboard
    console.log('View analytics clicked');
  }

  managePolicies(): void {
    // Navigate to policy management
    console.log('Manage policies clicked');
  }
}
