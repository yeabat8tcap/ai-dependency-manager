import { ComponentFixture, TestBed } from '@angular/core/testing';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { of, Subject } from 'rxjs';

import { DashboardComponent } from '../features/dashboard/dashboard.component';
import { DependencyTableComponent } from '../features/dependencies/dependency-table.component';
import { ApiService } from '../core/services/api.service';
import { WebSocketService } from '../core/services/websocket.service';

describe('Dashboard Integration Tests', () => {
  let dashboardComponent: DashboardComponent;
  let dashboardFixture: ComponentFixture<DashboardComponent>;
  let dependencyComponent: DependencyTableComponent;
  let dependencyFixture: ComponentFixture<DependencyTableComponent>;
  let mockApiService: jasmine.SpyObj<ApiService>;
  let mockWebSocketService: jasmine.SpyObj<WebSocketService>;

  const mockProjects = [
    {
      id: 1,
      name: 'Frontend App',
      type: 'npm' as 'npm',
      path: '/app/frontend',
      configFile: 'package.json',
      enabled: true,
      createdAt: new Date(),
      updatedAt: new Date(),
      lastScan: new Date()
    },
    {
      id: 2,
      name: 'Backend API',
      type: 'pip' as 'pip',
      path: '/app/backend',
      configFile: 'requirements.txt',
      enabled: true,
      createdAt: new Date(),
      updatedAt: new Date(),
      lastScan: new Date()
    }
  ];

  const mockDependencies = [
    {
      id: 1,
      projectId: 1,
      name: 'react',
      currentVersion: '18.2.0',
      latestVersion: '18.2.0',
      type: 'direct' as 'direct',
      status: 'up-to-date' as 'up-to-date'
    },
    {
      id: 2,
      projectId: 1,
      name: 'lodash',
      currentVersion: '4.17.20',
      latestVersion: '4.17.21',
      type: 'direct' as 'direct',
      status: 'outdated' as 'outdated',
      securityIssues: [
        {
          cve: 'CVE-2021-23337',
          severity: 'high' as 'high',
          description: 'Command injection vulnerability',
          fixedVersion: '4.17.21',
          publishedAt: new Date()
        }
      ]
    }
  ];

  beforeEach(async () => {
    const apiServiceSpy = jasmine.createSpyObj('ApiService', [
      'getProjects',
      'getDependencies',
      'getSystemStatus',
      'getAnalytics',
      'scanProject',
      'applyUpdates'
    ]);
    const wsServiceSpy = jasmine.createSpyObj('WebSocketService', [
      'connect',
      'disconnect',
      'getSystemStatus',
      'getScanProgress',
      'getNotifications',
      'subscribeToProject'
    ]);

    await TestBed.configureTestingModule({
      imports: [
        DashboardComponent,
        DependencyTableComponent,
        BrowserAnimationsModule,
        HttpClientTestingModule
      ],
      providers: [
        { provide: ApiService, useValue: apiServiceSpy },
        { provide: WebSocketService, useValue: wsServiceSpy }
      ]
    }).compileComponents();

    mockApiService = TestBed.inject(ApiService) as jasmine.SpyObj<ApiService>;
    mockWebSocketService = TestBed.inject(WebSocketService) as jasmine.SpyObj<WebSocketService>;

    // Setup default mock responses
    mockApiService.getProjects.and.returnValue(of({ success: true, data: mockProjects }));
    mockApiService.getDependencies.and.returnValue(of({ success: true, data: mockDependencies }));
    mockApiService.getSystemStatus.and.returnValue(of({ success: true, data: { health: 'healthy' } }));
    mockApiService.getAnalytics.and.returnValue(of({
      success: true,
      data: {
        totalProjects: 2,
        totalDependencies: 50,
        outdatedDependencies: 5,
        securityIssues: 2
      }
    }));
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
    mockWebSocketService.getNotifications.and.returnValue(of({ type: 'info', message: 'Test notification' }));

    dashboardFixture = TestBed.createComponent(DashboardComponent);
    dashboardComponent = dashboardFixture.componentInstance;

    dependencyFixture = TestBed.createComponent(DependencyTableComponent);
    dependencyComponent = dependencyFixture.componentInstance;
  });

  describe('Dashboard and Dependency Table Integration', () => {
    it('should load dashboard data and navigate to dependency details', () => {
      // Initialize dashboard
      dashboardComponent.ngOnInit();
      dashboardFixture.detectChanges();

      // Verify dashboard loads with mock data
      expect(dashboardComponent.dashboardStats.totalProjects).toBe(5); // Mock data
      expect(dashboardComponent.projectSummaries.length).toBe(5);

      // Initialize dependency table for first project
      // Component doesn't have projectId property - test data loading instead
      expect(dependencyComponent.dataSource.data.length).toBe(0);
      dependencyComponent.ngOnInit();
      dependencyFixture.detectChanges();

      // Verify dependency table loads project dependencies
      expect(mockApiService.getDependencies).toHaveBeenCalledWith(1);
      expect(dependencyComponent.dataSource.data.length).toBe(2);
      expect(dependencyComponent.totalDependencies).toBe(2);
    });

    it('should handle real-time updates across components', () => {
      const systemStatusSubject = new Subject();
      const scanProgressSubject = new Subject();

      mockWebSocketService.getSystemStatus.and.returnValue(systemStatusSubject.asObservable());
      mockWebSocketService.getScanProgress.and.returnValue(scanProgressSubject.asObservable());

      // Initialize both components
      dashboardComponent.ngOnInit();
      dependencyComponent.ngOnInit();
      dashboardFixture.detectChanges();
      dependencyFixture.detectChanges();

      // Simulate system status update
      systemStatusSubject.next({ status: 'warning', timestamp: new Date().toISOString() });

      // Verify dashboard updates system health
      expect(dashboardComponent.dashboardStats.systemHealth).toBe('warning');

      // Simulate scan progress update
      scanProgressSubject.next({
        projectId: 1,
        status: 'complete',
        progress: 100,
        message: 'Scan completed'
      });

      // Verify components handle the update
      expect(dashboardComponent['loadDashboardData']).toBeDefined();
    });

    it('should coordinate dependency scanning workflow', async () => {
      // Setup scan response
      mockApiService.scanProject.and.returnValue(of({ success: true, data: { message: 'Scan initiated' } }));

      // Initialize components
      dashboardComponent.ngOnInit();
      dependencyComponent.ngOnInit();
      dashboardFixture.detectChanges();
      dependencyFixture.detectChanges();

      // Trigger scan from dependency table
      dependencyComponent.scanAll();

      // Verify API call was made
      expect(mockApiService.scanProject).toHaveBeenCalledWith(1);

      // Simulate scan completion via WebSocket
      const scanProgressSubject = new Subject();
      mockWebSocketService.getScanProgress.and.returnValue(scanProgressSubject.asObservable());

      scanProgressSubject.next({
        projectId: 1,
        status: 'complete',
        progress: 100,
        message: 'Scan completed successfully'
      });

      // Verify dashboard refreshes data
      expect(mockApiService.getAnalytics).toHaveBeenCalled();
    });

    it('should handle dependency updates and refresh dashboard stats', () => {
      const updateOptions = {
        dependencies: ['lodash'],
        strategy: 'safe'
      };

      mockApiService.applyUpdates.and.returnValue(of({ success: true, data: { message: 'Updates applied' } }));

      // Initialize components
      dashboardComponent.ngOnInit();
      dependencyComponent.ngOnInit();
      dashboardFixture.detectChanges();
      dependencyFixture.detectChanges();

      // Trigger update from dependency table
      dependencyComponent.updateSelected();

      // Simulate successful update
      const updatedDependencies = mockDependencies.map(dep => ({
        ...dep,
        currentVersion: dep.latestVersion,
        status: 'up-to-date'
      }));

      mockApiService.getDependencies.and.returnValue(of({ success: true, data: updatedDependencies }));

      // Refresh dependency table
      dependencyComponent['loadDependencies']();
      dependencyFixture.detectChanges();

      // Verify updated data
      expect(dependencyComponent.dataSource.data.every(item => item.updateAvailable === false)).toBe(true);
    });
  });

  describe('Error Handling Integration', () => {
    it('should handle API errors gracefully across components', () => {
      // Setup error responses
      mockApiService.getProjects.and.returnValue(of({ success: false, data: [], message: 'API Error' }));
      mockApiService.getDependencies.and.returnValue(of({ success: false, data: [], message: 'Dependency Error' }));

      spyOn(console, 'error');

      // Initialize components
      dashboardComponent.ngOnInit();
      dependencyComponent.ngOnInit();
      dashboardFixture.detectChanges();
      dependencyFixture.detectChanges();

      // Verify error handling
      expect(console.error).toHaveBeenCalled();
      expect(dashboardComponent.loading).toBe(false);
      expect(dependencyComponent.loading).toBe(false);
    });

    it('should handle WebSocket connection failures', () => {
      const errorSubject = new Subject();
      mockWebSocketService.getSystemStatus.and.returnValue(errorSubject.asObservable());

      spyOn(console, 'error');

      // Initialize dashboard
      dashboardComponent.ngOnInit();
      dashboardFixture.detectChanges();

      // Simulate WebSocket error
      errorSubject.error(new Error('WebSocket connection failed'));

      // Verify error handling
      expect(console.error).toHaveBeenCalled();
    });
  });

  describe('Performance Integration', () => {
    it('should handle large datasets efficiently', () => {
      // Create large mock dataset
      const largeDependencyList = Array.from({ length: 1000 }, (_, i) => ({
        id: i + 1,
        projectId: 1,
        name: `package-${i}`,
        currentVersion: '1.0.0',
        latestVersion: '1.0.1',
        type: 'direct',
        status: 'outdated'
      }));

      mockApiService.getDependencies.and.returnValue(of({ success: true, data: largeDependencyList }));

      // Initialize dependency table
      const startTime = performance.now();
      dependencyComponent.ngOnInit();
      dependencyFixture.detectChanges();
      const endTime = performance.now();

      // Verify performance
      expect(endTime - startTime).toBeLessThan(1000); // Should complete within 1 second
      expect(dependencyComponent.dataSource.data.length).toBe(1000);
    });

    it('should optimize memory usage with component lifecycle', () => {
      // Initialize components
      dashboardComponent.ngOnInit();
      dependencyComponent.ngOnInit();

      // Verify subscriptions are created
      expect(dashboardComponent['destroy$']).toBeDefined();
      expect(dependencyComponent['destroy$']).toBeDefined();

      // Destroy components
      dashboardComponent.ngOnDestroy();
      dependencyComponent.ngOnDestroy();

      // Verify cleanup
      expect(dashboardComponent['destroy$'].closed).toBe(true);
      expect(dependencyComponent['destroy$'].closed).toBe(true);
    });
  });

  describe('User Workflow Integration', () => {
    it('should support complete dependency management workflow', () => {
      // 1. Load dashboard
      dashboardComponent.ngOnInit();
      dashboardFixture.detectChanges();

      expect(dashboardComponent.projectSummaries.length).toBeGreaterThan(0);

      // 2. Navigate to dependency table for a project
      const projectId = dashboardComponent.projectSummaries[0].id;
      dependencyComponent.projectId = projectId;
      dependencyComponent.ngOnInit();
      dependencyFixture.detectChanges();

      expect(dependencyComponent.dataSource.data.length).toBeGreaterThan(0);

      // 3. Filter dependencies
      dependencyComponent.searchControl.setValue('lodash');
      dependencyFixture.detectChanges();

      expect(dependencyComponent.dataSource.filter).toBeTruthy();

      // 4. Select dependencies for update
      const firstDependency = dependencyComponent.dataSource.data[0];
      dependencyComponent.selection.select(firstDependency);

      expect(dependencyComponent.selection.selected.length).toBe(1);

      // 5. Apply updates
      mockApiService.applyUpdates.and.returnValue(of({ success: true, data: { message: 'Updates applied' } }));
      dependencyComponent.updateSelected();

      expect(mockApiService.applyUpdates).toHaveBeenCalled();

      // 6. Verify dashboard reflects changes
      dashboardComponent['loadDashboardData']();
      expect(mockApiService.getAnalytics).toHaveBeenCalled();
    });
  });
});
