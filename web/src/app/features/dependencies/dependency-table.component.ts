import { Component, OnInit, OnDestroy, ViewChild, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatTableModule, MatTableDataSource } from '@angular/material/table';
import { MatPaginatorModule, MatPaginator } from '@angular/material/paginator';
import { MatSortModule, MatSort } from '@angular/material/sort';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatChipsModule } from '@angular/material/chips';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatMenuModule } from '@angular/material/menu';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { SelectionModel } from '@angular/cdk/collections';
import { Subject, takeUntil, debounceTime, distinctUntilChanged } from 'rxjs';
import { FormControl } from '@angular/forms';
import { ReactiveFormsModule } from '@angular/forms';

import { ApiService, Dependency } from '../../core/services/api.service';
import { WebSocketService } from '../../core/services/websocket.service';

export interface DependencyTableItem {
  id: number;
  name: string;
  currentVersion: string;
  latestVersion: string;
  type: string;
  projectName: string;
  riskScore: number;
  updateRecommendation: 'safe' | 'caution' | 'risky' | 'critical';
  aiConfidence: number;
  hasBreakingChanges: boolean;
  securityIssues: number;
  riskLevel: 'low' | 'medium' | 'high' | 'critical';
  updateAvailable: boolean;
  aiRecommendation?: string;
}

@Component({
  selector: 'app-dependency-table',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatTableModule,
    MatPaginatorModule,
    MatSortModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
    MatChipsModule,
    MatIconModule,
    MatButtonModule,
    MatMenuModule,
    MatTooltipModule,
    MatProgressSpinnerModule,
    MatCheckboxModule
  ],
  template: `
    <div class="dependency-table-container">
      <!-- Header and Filters -->
      <div class="table-header">
        <div class="header-title">
          <h2>
            <mat-icon>account_tree</mat-icon>
            Dependencies
          </h2>
          <span class="dependency-count" *ngIf="!loading">
            {{ dataSource.filteredData.length }} of {{ totalDependencies }} dependencies
          </span>
        </div>
        
        <div class="table-actions">
          <button mat-raised-button color="primary" 
                  [disabled]="selection.isEmpty()" 
                  (click)="updateSelected()">
            <mat-icon>update</mat-icon>
            Update Selected ({{ selection.selected.length }})
          </button>
          <button mat-raised-button color="accent" (click)="scanAll()">
            <mat-icon>refresh</mat-icon>
            Scan All
          </button>
        </div>
      </div>

      <!-- Filters -->
      <div class="filters-section">
        <mat-form-field appearance="outline" class="search-field">
          <mat-label>Search dependencies</mat-label>
          <mat-icon matPrefix>search</mat-icon>
          <input matInput [formControl]="searchControl" placeholder="Search by name, version, or project">
        </mat-form-field>

        <mat-form-field appearance="outline">
          <mat-label>Project</mat-label>
          <mat-select [formControl]="projectFilter" multiple>
            <mat-option value="all">All Projects</mat-option>
            <mat-option *ngFor="let project of availableProjects" [value]="project">
              {{ project }}
            </mat-option>
          </mat-select>
        </mat-form-field>

        <mat-form-field appearance="outline">
          <mat-label>Status</mat-label>
          <mat-select [formControl]="statusFilter" multiple>
            <mat-option value="outdated">Outdated</mat-option>
            <mat-option value="up-to-date">Up to Date</mat-option>
            <mat-option value="security-issue">Security Issues</mat-option>
            <mat-option value="breaking-changes">Breaking Changes</mat-option>
          </mat-select>
        </mat-form-field>

        <mat-form-field appearance="outline">
          <mat-label>Risk Level</mat-label>
          <mat-select [formControl]="riskFilter" multiple>
            <mat-option value="safe">Safe</mat-option>
            <mat-option value="caution">Caution</mat-option>
            <mat-option value="risky">Risky</mat-option>
            <mat-option value="critical">Critical</mat-option>
          </mat-select>
        </mat-form-field>

        <button mat-icon-button (click)="clearFilters()" matTooltip="Clear all filters">
          <mat-icon>clear</mat-icon>
        </button>
      </div>

      <!-- Loading Spinner -->
      <div class="loading-container" *ngIf="loading">
        <mat-spinner></mat-spinner>
        <p>Loading dependencies...</p>
      </div>

      <!-- Data Table -->
      <div class="table-wrapper" *ngIf="!loading">
        <table mat-table [dataSource]="dataSource" matSort class="dependency-table">
          <!-- Selection Column -->
          <ng-container matColumnDef="select">
            <th mat-header-cell *matHeaderCellDef>
              <mat-checkbox (change)="$event ? toggleAllRows() : null"
                          [checked]="selection.hasValue() && isAllSelected()"
                          [indeterminate]="selection.hasValue() && !isAllSelected()">
              </mat-checkbox>
            </th>
            <td mat-cell *matCellDef="let row">
              <mat-checkbox (click)="$event.stopPropagation()"
                          (change)="$event ? selection.toggle(row) : null"
                          [checked]="selection.isSelected(row)">
              </mat-checkbox>
            </td>
          </ng-container>

          <!-- Name Column -->
          <ng-container matColumnDef="name">
            <th mat-header-cell *matHeaderCellDef mat-sort-header>Name</th>
            <td mat-cell *matCellDef="let dependency">
              <div class="dependency-name">
                <mat-icon class="package-icon">{{ getPackageIcon(dependency.type) }}</mat-icon>
                <div class="name-info">
                  <div class="package-name">{{ dependency.name }}</div>
                  <div class="project-name">{{ dependency.projectName }}</div>
                </div>
              </div>
            </td>
          </ng-container>

          <!-- Current Version Column -->
          <ng-container matColumnDef="currentVersion">
            <th mat-header-cell *matHeaderCellDef mat-sort-header>Current Version</th>
            <td mat-cell *matCellDef="let dependency">
              <mat-chip class="version-chip current-version">
                {{ dependency.currentVersion }}
              </mat-chip>
            </td>
          </ng-container>

          <!-- Latest Version Column -->
          <ng-container matColumnDef="latestVersion">
            <th mat-header-cell *matHeaderCellDef mat-sort-header>Latest Version</th>
            <td mat-cell *matCellDef="let dependency">
              <mat-chip class="version-chip latest-version" 
                       [class.outdated]="dependency.currentVersion !== dependency.latestVersion">
                {{ dependency.latestVersion }}
              </mat-chip>
            </td>
          </ng-container>

          <!-- Status Column -->
          <ng-container matColumnDef="status">
            <th mat-header-cell *matHeaderCellDef mat-sort-header>Status</th>
            <td mat-cell *matCellDef="let dependency">
              <div class="status-indicators">
                <mat-chip class="status-chip" [class]="'status-' + getStatusClass(dependency)">
                  <mat-icon>{{ getStatusIcon(dependency) }}</mat-icon>
                  {{ getStatusText(dependency) }}
                </mat-chip>
                <mat-icon class="security-warning" 
                         *ngIf="dependency.securityIssues > 0"
                         matTooltip="{{ dependency.securityIssues }} security issues"
                         color="warn">
                  security
                </mat-icon>
                <mat-icon class="breaking-changes-warning" 
                         *ngIf="dependency.hasBreakingChanges"
                         matTooltip="Contains breaking changes"
                         color="accent">
                  warning
                </mat-icon>
              </div>
            </td>
          </ng-container>

          <!-- Risk Assessment Column -->
          <ng-container matColumnDef="riskAssessment">
            <th mat-header-cell *matHeaderCellDef mat-sort-header>AI Risk Assessment</th>
            <td mat-cell *matCellDef="let dependency">
              <div class="risk-assessment">
                <mat-chip class="risk-chip" [class]="'risk-' + dependency.updateRecommendation">
                  <mat-icon>{{ getRiskIcon(dependency.updateRecommendation) }}</mat-icon>
                  {{ dependency.updateRecommendation | titlecase }}
                </mat-chip>
                <div class="confidence-score">
                  <span class="confidence-label">Confidence:</span>
                  <span class="confidence-value">{{ dependency.aiConfidence }}%</span>
                </div>
              </div>
            </td>
          </ng-container>

          <!-- Actions Column -->
          <ng-container matColumnDef="actions">
            <th mat-header-cell *matHeaderCellDef>Actions</th>
            <td mat-cell *matCellDef="let dependency">
              <button mat-icon-button [matMenuTriggerFor]="actionMenu">
                <mat-icon>more_vert</mat-icon>
              </button>
              <mat-menu #actionMenu="matMenu">
                <button mat-menu-item (click)="viewDetails(dependency)">
                  <mat-icon>info</mat-icon>
                  View Details
                </button>
                <button mat-menu-item (click)="updateDependency(dependency)" 
                        [disabled]="dependency.currentVersion === dependency.latestVersion">
                  <mat-icon>update</mat-icon>
                  Update
                </button>
                <button mat-menu-item (click)="viewChangelog(dependency)">
                  <mat-icon>history</mat-icon>
                  View Changelog
                </button>
                <button mat-menu-item (click)="analyzeImpact(dependency)">
                  <mat-icon>analytics</mat-icon>
                  Analyze Impact
                </button>
              </mat-menu>
            </td>
          </ng-container>

          <tr mat-header-row *matHeaderRowDef="displayedColumns"></tr>
          <tr mat-row *matRowDef="let row; columns: displayedColumns;" 
              (click)="viewDetails(row)" 
              class="dependency-row"></tr>
        </table>

        <!-- No Data Message -->
        <div class="no-data" *ngIf="dataSource.filteredData.length === 0">
          <mat-icon class="no-data-icon">search_off</mat-icon>
          <h3>No dependencies found</h3>
          <p>Try adjusting your filters or search criteria.</p>
        </div>
      </div>

      <!-- Paginator -->
      <mat-paginator [pageSizeOptions]="[25, 50, 100]" 
                     [pageSize]="50"
                     showFirstLastButtons
                     *ngIf="!loading">
      </mat-paginator>
    </div>
  `,
  styles: [`
    .dependency-table-container {
      padding: 24px;
    }

    .table-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 24px;
    }

    .header-title {
      display: flex;
      align-items: center;
      gap: 16px;
    }

    .header-title h2 {
      display: flex;
      align-items: center;
      gap: 8px;
      margin: 0;
    }

    .dependency-count {
      color: rgba(0, 0, 0, 0.6);
      font-size: 0.9rem;
    }

    .table-actions {
      display: flex;
      gap: 12px;
    }

    .table-actions button {
      display: flex;
      align-items: center;
      gap: 8px;
    }

    .filters-section {
      display: flex;
      gap: 16px;
      margin-bottom: 24px;
      flex-wrap: wrap;
      align-items: center;
    }

    .search-field {
      flex: 1;
      min-width: 300px;
    }

    .filters-section mat-form-field {
      min-width: 150px;
    }

    .loading-container {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      padding: 48px;
      gap: 16px;
    }

    .table-wrapper {
      background: white;
      border-radius: 8px;
      overflow: hidden;
      box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    }

    .dependency-table {
      width: 100%;
    }

    .dependency-row {
      cursor: pointer;
      transition: background-color 0.2s;
    }

    .dependency-row:hover {
      background-color: rgba(0, 0, 0, 0.04);
    }

    .dependency-name {
      display: flex;
      align-items: center;
      gap: 12px;
    }

    .package-icon {
      color: rgba(0, 0, 0, 0.6);
    }

    .name-info {
      display: flex;
      flex-direction: column;
    }

    .package-name {
      font-weight: 500;
    }

    .project-name {
      font-size: 0.8rem;
      color: rgba(0, 0, 0, 0.6);
    }

    .version-chip {
      font-family: 'Courier New', monospace;
      font-size: 0.8rem;
    }

    .current-version {
      background-color: #e3f2fd;
      color: #1976d2;
    }

    .latest-version {
      background-color: #e8f5e8;
      color: #2e7d32;
    }

    .latest-version.outdated {
      background-color: #fff3e0;
      color: #f57c00;
    }

    .status-indicators {
      display: flex;
      align-items: center;
      gap: 8px;
    }

    .status-chip {
      display: flex;
      align-items: center;
      gap: 4px;
    }

    .status-up-to-date {
      background-color: #e8f5e8;
      color: #2e7d32;
    }

    .status-outdated {
      background-color: #fff3e0;
      color: #f57c00;
    }

    .status-security-issue {
      background-color: #ffebee;
      color: #d32f2f;
    }

    .risk-assessment {
      display: flex;
      flex-direction: column;
      gap: 4px;
    }

    .risk-chip {
      display: flex;
      align-items: center;
      gap: 4px;
    }

    .risk-safe {
      background-color: #e8f5e8;
      color: #2e7d32;
    }

    .risk-caution {
      background-color: #fff3e0;
      color: #f57c00;
    }

    .risk-risky {
      background-color: #ffebee;
      color: #d32f2f;
    }

    .risk-critical {
      background-color: #ffebee;
      color: #b71c1c;
    }

    .confidence-score {
      font-size: 0.8rem;
      color: rgba(0, 0, 0, 0.6);
    }

    .confidence-label {
      margin-right: 4px;
    }

    .confidence-value {
      font-weight: 500;
    }

    .no-data {
      text-align: center;
      padding: 48px 24px;
    }

    .no-data-icon {
      font-size: 4rem;
      width: 4rem;
      height: 4rem;
      color: rgba(0, 0, 0, 0.3);
      margin-bottom: 16px;
    }

    @media (max-width: 768px) {
      .dependency-table-container {
        padding: 16px;
      }

      .table-header {
        flex-direction: column;
        align-items: flex-start;
        gap: 16px;
      }

      .table-actions {
        width: 100%;
        justify-content: stretch;
      }

      .table-actions button {
        flex: 1;
      }

      .filters-section {
        flex-direction: column;
        align-items: stretch;
      }

      .search-field {
        min-width: unset;
      }

      .filters-section mat-form-field {
        min-width: unset;
      }
    }
  `]
})
export class DependencyTableComponent implements OnInit, OnDestroy {
  @ViewChild(MatPaginator) paginator!: MatPaginator;
  @ViewChild(MatSort) sort!: MatSort;

  private destroy$ = new Subject<void>();
  private apiService = inject(ApiService);
  private wsService = inject(WebSocketService);

  displayedColumns: string[] = ['select', 'name', 'currentVersion', 'latestVersion', 'status', 'riskAssessment', 'actions'];
  dataSource = new MatTableDataSource<DependencyTableItem>([]);
  selection = new SelectionModel<DependencyTableItem>(true, []);

  // Form controls for filters
  searchControl = new FormControl('');
  projectFilter = new FormControl<string[]>([]);
  statusFilter = new FormControl<string[]>([]);
  riskFilter = new FormControl<string[]>([]);

  availableProjects: string[] = [];
  totalDependencies = 0;
  loading = true;

  ngOnInit(): void {
    this.setupFilters();
    this.loadDependencies();
    this.subscribeToRealTimeUpdates();
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  ngAfterViewInit(): void {
    this.dataSource.paginator = this.paginator;
    this.dataSource.sort = this.sort;
  }

  private setupFilters(): void {
    // Search filter
    this.searchControl.valueChanges.pipe(
      debounceTime(300),
      distinctUntilChanged(),
      takeUntil(this.destroy$)
    ).subscribe(() => {
      this.applyFilters();
    });

    // Other filters
    this.projectFilter.valueChanges.pipe(takeUntil(this.destroy$)).subscribe(() => this.applyFilters());
    this.statusFilter.valueChanges.pipe(takeUntil(this.destroy$)).subscribe(() => this.applyFilters());
    this.riskFilter.valueChanges.pipe(takeUntil(this.destroy$)).subscribe(() => this.applyFilters());
  }

  private loadDependencies(): void {
    this.apiService.getDependencies(1).pipe(
      takeUntil(this.destroy$)
    ).subscribe({
      next: (response) => {
        if (response.success) {
          this.processDependencyData(response.data);
        }
        this.loading = false;
      },
      error: (error) => {
        console.error('Error loading dependencies:', error);
        this.loading = false;
      }
    });
  }

  private processDependencyData(dependencies: Dependency[]): void {
    const tableItems: DependencyTableItem[] = dependencies.map(dep => ({
      ...dep,
      latestVersion: dep.latestVersion || dep.currentVersion, // Ensure latestVersion is always a string
      projectName: 'Project Name', // Will be populated from API
      riskScore: Math.floor(Math.random() * 100),
      updateRecommendation: this.calculateUpdateRecommendation(dep),
      aiConfidence: Math.floor(Math.random() * 40) + 60,
      hasBreakingChanges: Math.random() > 0.7,
      securityIssues: Math.floor(Math.random() * 3),
      riskLevel: this.calculateRiskLevel(dep),
      updateAvailable: this.checkUpdateAvailable(dep)
    }));

    this.dataSource.data = tableItems;
    this.totalDependencies = tableItems.length;
    this.availableProjects = [...new Set(tableItems.map(item => item.projectName))];
  }

  private calculateUpdateRecommendation(dependency: Dependency): 'safe' | 'caution' | 'risky' | 'critical' {
    // Simple logic for demo - can be enhanced with AI analysis
    const random = Math.random();
    if (random > 0.8) return 'critical';
    if (random > 0.6) return 'risky';
    if (random > 0.3) return 'caution';
    return 'safe';
  }

  private subscribeToRealTimeUpdates(): void {
    this.wsService.getScanProgress().pipe(
      takeUntil(this.destroy$)
    ).subscribe((progress) => {
      if (progress.status === 'complete') {
        this.loadDependencies();
      }
    });
  }

  private applyFilters(): void {
    this.dataSource.filterPredicate = (data: DependencyTableItem, filter: string) => {
      const searchTerm = this.searchControl.value?.toLowerCase() || '';
      const projectFilters = this.projectFilter.value || [];
      const statusFilters = this.statusFilter.value || [];
      const riskFilters = this.riskFilter.value || [];

      // Search filter
      const matchesSearch = !searchTerm || 
        data.name.toLowerCase().includes(searchTerm) ||
        data.currentVersion.toLowerCase().includes(searchTerm) ||
        data.projectName.toLowerCase().includes(searchTerm);

      // Project filter
      const matchesProject = projectFilters.length === 0 || 
        projectFilters.includes('all') ||
        projectFilters.includes(data.projectName);

      // Status filter
      const matchesStatus = statusFilters.length === 0 || this.matchesStatusFilter(data, statusFilters);

      // Risk filter
      const matchesRisk = riskFilters.length === 0 || riskFilters.includes(data.updateRecommendation);

      return matchesSearch && matchesProject && matchesStatus && matchesRisk;
    };

    this.dataSource.filter = 'trigger'; // Trigger filter
  }

  private matchesStatusFilter(data: DependencyTableItem, statusFilters: string[]): boolean {
    return statusFilters.some(filter => {
      switch (filter) {
        case 'outdated':
          return data.currentVersion !== data.latestVersion;
        case 'up-to-date':
          return data.currentVersion === data.latestVersion;
        case 'security-issue':
          return data.securityIssues > 0;
        case 'breaking-changes':
          return data.hasBreakingChanges;
        default:
          return false;
      }
    });
  }

  clearFilters(): void {
    this.searchControl.setValue('');
    this.projectFilter.setValue([]);
    this.statusFilter.setValue([]);
    this.riskFilter.setValue([]);
  }

  // Selection methods
  isAllSelected(): boolean {
    const numSelected = this.selection.selected.length;
    const numRows = this.dataSource.filteredData.length;
    return numSelected === numRows;
  }

  toggleAllRows(): void {
    if (this.isAllSelected()) {
      this.selection.clear();
    } else {
      this.dataSource.filteredData.forEach(row => this.selection.select(row));
    }
  }

  // Helper methods
  getPackageIcon(type: string): string {
    switch (type) {
      case 'npm': return 'javascript';
      case 'pip': return 'code';
      case 'maven': return 'coffee';
      case 'gradle': return 'build';
      default: return 'package';
    }
  }

  getStatusClass(dependency: DependencyTableItem): string {
    if (dependency.securityIssues > 0) return 'security-issue';
    if (dependency.currentVersion !== dependency.latestVersion) return 'outdated';
    return 'up-to-date';
  }

  getStatusIcon(dependency: DependencyTableItem): string {
    if (dependency.securityIssues > 0) return 'security';
    if (dependency.currentVersion !== dependency.latestVersion) return 'update';
    return 'check_circle';
  }

  getStatusText(dependency: DependencyTableItem): string {
    if (dependency.securityIssues > 0) return 'Security Issues';
    if (dependency.currentVersion !== dependency.latestVersion) return 'Outdated';
    return 'Up to Date';
  }

  getRiskIcon(risk: string): string {
    switch (risk) {
      case 'safe': return 'check_circle';
      case 'caution': return 'warning';
      case 'risky': return 'error';
      case 'critical': return 'dangerous';
      default: return 'help';
    }
  }

  // Action handlers
  updateSelected(): void {
    const selectedDependencies = this.selection.selected;
    console.log('Update selected dependencies:', selectedDependencies);
    // Implement update logic
  }

  scanAll(): void {
    console.log('Scan all dependencies');
    // Implement scan logic
  }

  viewDetails(dependency: DependencyTableItem): void {
    console.log('View dependency details:', dependency);
    // Navigate to dependency details
  }

  updateDependency(dependency: DependencyTableItem): void {
    console.log('Update dependency:', dependency);
    // Implement single dependency update
  }

  viewChangelog(dependency: DependencyTableItem): void {
    console.log('View changelog:', dependency);
    // Open changelog dialog
  }

  analyzeImpact(dependency: DependencyTableItem): void {
    console.log('Analyze impact:', dependency);
    // Open impact analysis dialog
  }

  private calculateRiskLevel(dep: Dependency): 'low' | 'medium' | 'high' | 'critical' {
    // Use AI analysis if available, otherwise calculate based on other factors
    if (dep.aiAnalysis?.riskLevel) {
      return dep.aiAnalysis.riskLevel;
    }

    // Calculate risk level based on security issues and status
    if (dep.securityIssues && dep.securityIssues.length > 0) {
      const highestSeverity = dep.securityIssues.reduce((max, issue) => {
        const severityOrder = { low: 1, medium: 2, high: 3, critical: 4 };
        return severityOrder[issue.severity] > severityOrder[max] ? issue.severity : max;
      }, 'low' as 'low' | 'medium' | 'high' | 'critical');
      return highestSeverity;
    }

    // Default risk assessment based on status
    switch (dep.status) {
      case 'vulnerable': return 'high';
      case 'outdated': return 'medium';
      case 'up-to-date': return 'low';
      default: return 'medium';
    }
  }

  private checkUpdateAvailable(dep: Dependency): boolean {
    // Check if there's a newer version available
    if (!dep.latestVersion || !dep.currentVersion) {
      return false;
    }
    
    // Simple version comparison - in a real app, you'd use a proper semver library
    return dep.currentVersion !== dep.latestVersion;
  }
}
