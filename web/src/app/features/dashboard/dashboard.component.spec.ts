import { ComponentFixture, TestBed } from '@angular/core/testing';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { of } from 'rxjs';

import { DashboardComponent, DashboardStats, ProjectSummary } from './dashboard.component';
import { ApiService } from '../../core/services/api.service';
import { WebSocketService } from '../../core/services/websocket.service';

describe('DashboardComponent', () => {
  let component: DashboardComponent;
  let fixture: ComponentFixture<DashboardComponent>;
  let mockApiService: jasmine.SpyObj<ApiService>;
  let mockWebSocketService: jasmine.SpyObj<WebSocketService>;

  const mockDashboardStats: DashboardStats = {
    totalProjects: 5,
    totalDependencies: 247,
    outdatedDependencies: 23,
    securityIssues: 7,
    lastScanTime: new Date('2024-01-15T10:00:00Z'),
    systemHealth: 'warning'
  };

  const mockProjectSummaries: ProjectSummary[] = [
    {
      id: 1,
      name: 'E-Commerce Frontend',
      type: 'npm',
      dependencyCount: 89,
      outdatedCount: 12,
      securityIssueCount: 3,
      lastScan: new Date('2024-01-15T09:00:00Z'),
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
      lastScan: new Date('2024-01-15T08:00:00Z'),
      healthScore: 85,
      riskLevel: 'low'
    }
  ];

  beforeEach(async () => {
    const apiServiceSpy = jasmine.createSpyObj('ApiService', [
      'getProjects',
      'getSystemStatus',
      'getAnalytics'
    ]);
    const wsServiceSpy = jasmine.createSpyObj('WebSocketService', [
      'getSystemStatus',
      'getScanProgress'
    ]);

    await TestBed.configureTestingModule({
      imports: [
        DashboardComponent,
        BrowserAnimationsModule
      ],
      providers: [
        { provide: ApiService, useValue: apiServiceSpy },
        { provide: WebSocketService, useValue: wsServiceSpy }
      ]
    }).compileComponents();

    mockApiService = TestBed.inject(ApiService) as jasmine.SpyObj<ApiService>;
    mockWebSocketService = TestBed.inject(WebSocketService) as jasmine.SpyObj<WebSocketService>;

    // Setup default mock responses
    mockApiService.getProjects.and.returnValue(of({ success: true, data: [] }));
    mockApiService.getSystemStatus.and.returnValue(of({ success: true, data: { health: 'healthy' } }));
    mockApiService.getAnalytics.and.returnValue(of({ success: true, data: mockDashboardStats }));
    mockWebSocketService.getSystemStatus.and.returnValue(of({ 
      status: 'healthy',
      uptime: 3600,
      projectsMonitored: 5,
      backgroundScansActive: 2
    }));
    mockWebSocketService.getScanProgress.and.returnValue(of({ 
      projectId: 1, 
      projectName: 'Test Project',
      status: 'complete',
      progress: 100,
      totalDependencies: 50,
      scannedDependencies: 50
    }));

    fixture = TestBed.createComponent(DashboardComponent);
    component = fixture.componentInstance;
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should initialize with default dashboard stats', () => {
    expect(component.dashboardStats).toEqual({
      totalProjects: 0,
      totalDependencies: 0,
      outdatedDependencies: 0,
      securityIssues: 0,
      systemHealth: 'healthy'
    });
    expect(component.projectSummaries).toEqual([]);
    expect(component.loading).toBe(true);
  });

  it('should load mock data on init', () => {
    component.ngOnInit();
    fixture.detectChanges();

    expect(component.loading).toBe(false);
    expect(component.dashboardStats.totalProjects).toBe(5);
    expect(component.dashboardStats.totalDependencies).toBe(247);
    expect(component.dashboardStats.outdatedDependencies).toBe(23);
    expect(component.dashboardStats.securityIssues).toBe(7);
    expect(component.projectSummaries.length).toBe(5);
  });

  it('should display correct project summaries', () => {
    component.ngOnInit();
    fixture.detectChanges();

    const projectSummaries = component.projectSummaries;
    expect(projectSummaries[0].name).toBe('E-Commerce Frontend');
    expect(projectSummaries[0].type).toBe('npm');
    expect(projectSummaries[0].riskLevel).toBe('medium');
    expect(projectSummaries[1].name).toBe('API Gateway Service');
    expect(projectSummaries[1].type).toBe('pip');
    expect(projectSummaries[1].riskLevel).toBe('low');
  });

  it('should calculate health scores correctly', () => {
    const mockProject = { id: 1, name: 'Test', type: 'npm' } as any;
    const healthScore = component['calculateHealthScore'](mockProject);
    expect(healthScore).toBeGreaterThanOrEqual(60);
    expect(healthScore).toBeLessThanOrEqual(100);
  });

  it('should calculate risk levels correctly', () => {
    const mockProject = { id: 1, name: 'Test', type: 'npm' } as any;
    const riskLevel = component['calculateRiskLevel'](mockProject);
    expect(['low', 'medium', 'high', 'critical']).toContain(riskLevel);
  });

  it('should display dashboard stats in template', () => {
    component.ngOnInit();
    fixture.detectChanges();

    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('5'); // Total projects
    expect(compiled.textContent).toContain('247'); // Total dependencies
    expect(compiled.textContent).toContain('23'); // Outdated dependencies
    expect(compiled.textContent).toContain('7'); // Security issues
  });

  it('should handle system health status correctly', () => {
    component.ngOnInit();
    fixture.detectChanges();

    expect(component.dashboardStats.systemHealth).toBe('warning');
  });

  it('should clean up subscriptions on destroy', () => {
    spyOn(component['destroy$'], 'next');
    spyOn(component['destroy$'], 'complete');

    component.ngOnDestroy();

    expect(component['destroy$'].next).toHaveBeenCalled();
    expect(component['destroy$'].complete).toHaveBeenCalled();
  });

  it('should format last scan time correctly', () => {
    component.ngOnInit();
    fixture.detectChanges();

    expect(component.dashboardStats.lastScanTime).toBeInstanceOf(Date);
  });

  it('should handle different project types', () => {
    component.ngOnInit();
    fixture.detectChanges();

    const projectTypes = component.projectSummaries.map(p => p.type);
    expect(projectTypes).toContain('npm');
    expect(projectTypes).toContain('pip');
    expect(projectTypes).toContain('maven');
    expect(projectTypes).toContain('gradle');
  });
});
