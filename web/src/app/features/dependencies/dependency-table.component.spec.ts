import { ComponentFixture, TestBed } from '@angular/core/testing';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { ReactiveFormsModule } from '@angular/forms';
import { of } from 'rxjs';

import { DependencyTableComponent, DependencyTableItem } from './dependency-table.component';
import { ApiService, Dependency } from '../../core/services/api.service';

describe('DependencyTableComponent', () => {
  let component: DependencyTableComponent;
  let fixture: ComponentFixture<DependencyTableComponent>;
  let mockApiService: jasmine.SpyObj<ApiService>;

  const mockDependencies: Dependency[] = [
    {
      id: 1,
      projectId: 1,
      name: 'lodash',
      currentVersion: '4.17.20',
      latestVersion: '4.17.21',
      type: 'direct',
      status: 'outdated',
      securityIssues: [
        {
          cve: 'CVE-2021-23337',
          severity: 'high',
          description: 'Command injection vulnerability',
          fixedVersion: '4.17.21',
          publishedAt: new Date('2021-02-15')
        }
      ]
    },
    {
      id: 2,
      projectId: 1,
      name: 'express',
      currentVersion: '4.18.0',
      latestVersion: '4.18.2',
      type: 'direct',
      status: 'outdated'
    },
    {
      id: 3,
      projectId: 1,
      name: 'react',
      currentVersion: '18.2.0',
      latestVersion: '18.2.0',
      type: 'direct',
      status: 'up-to-date'
    }
  ];

  beforeEach(async () => {
    const apiServiceSpy = jasmine.createSpyObj('ApiService', [
      'getDependencies',
      'scanProject',
      'applyUpdates'
    ]);

    await TestBed.configureTestingModule({
      imports: [
        DependencyTableComponent,
        BrowserAnimationsModule,
        ReactiveFormsModule
      ],
      providers: [
        { provide: ApiService, useValue: apiServiceSpy }
      ]
    }).compileComponents();

    mockApiService = TestBed.inject(ApiService) as jasmine.SpyObj<ApiService>;
    mockApiService.getDependencies.and.returnValue(of({ success: true, data: mockDependencies }));

    fixture = TestBed.createComponent(DependencyTableComponent);
    component = fixture.componentInstance;
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should initialize with default values', () => {
    // Test component initialization
    expect(component.dataSource.data.length).toBe(0);
    expect(component.loading).toBe(true);
    expect(component.dataSource.data).toEqual([]);
    expect(component.totalDependencies).toBe(0);
    expect(component.availableProjects).toEqual([]);
  });

  it('should load dependencies on init', () => {
    component.ngOnInit();
    fixture.detectChanges();

    expect(mockApiService.getDependencies).toHaveBeenCalledWith(1);
    expect(component.loading).toBe(false);
    expect(component.dataSource.data.length).toBe(3);
    expect(component.totalDependencies).toBe(3);
  });

  it('should process dependency data correctly', () => {
    component.ngOnInit();
    fixture.detectChanges();

    const tableItems = component.dataSource.data;
    expect(tableItems[0].name).toBe('lodash');
    expect(tableItems[0].currentVersion).toBe('4.17.20');
    expect(tableItems[0].latestVersion).toBe('4.17.21');
    expect(tableItems[0].riskLevel).toBe('high'); // Based on security issue
    expect(tableItems[0].updateAvailable).toBe(true);
    expect(tableItems[0].securityIssues).toBe(1);
  });

  it('should calculate risk levels correctly', () => {
    const depWithSecurity: Dependency = {
      ...mockDependencies[0],
      securityIssues: [{ severity: 'critical' } as any]
    };
    const riskLevel = component['calculateRiskLevel'](depWithSecurity);
    expect(riskLevel).toBe('critical');

    const depVulnerable: Dependency = {
      ...mockDependencies[0],
      status: 'vulnerable',
      securityIssues: []
    };
    const riskLevel2 = component['calculateRiskLevel'](depVulnerable);
    expect(riskLevel2).toBe('high');

    const depUpToDate: Dependency = {
      ...mockDependencies[2],
      status: 'up-to-date'
    };
    const riskLevel3 = component['calculateRiskLevel'](depUpToDate);
    expect(riskLevel3).toBe('low');
  });

  it('should check update availability correctly', () => {
    const depWithUpdate = mockDependencies[0];
    const updateAvailable = component['checkUpdateAvailable'](depWithUpdate);
    expect(updateAvailable).toBe(true);

    const depUpToDate = mockDependencies[2];
    const updateAvailable2 = component['checkUpdateAvailable'](depUpToDate);
    expect(updateAvailable2).toBe(false);

    const depWithoutLatest: Dependency = {
      ...mockDependencies[0],
      latestVersion: undefined
    };
    const updateAvailable3 = component['checkUpdateAvailable'](depWithoutLatest);
    expect(updateAvailable3).toBe(false);
  });

  it('should calculate update recommendations correctly', () => {
    const depSafe: Dependency = {
      ...mockDependencies[2],
      status: 'up-to-date'
    };
    const recommendation = component['calculateUpdateRecommendation'](depSafe);
    expect(recommendation).toBe('safe');

    const depWithSecurity: Dependency = {
      ...mockDependencies[0],
      securityIssues: [{ severity: 'critical' } as any]
    };
    const recommendation2 = component['calculateUpdateRecommendation'](depWithSecurity);
    expect(recommendation2).toBe('critical');
  });

  it('should handle search filtering', () => {
    component.ngOnInit();
    fixture.detectChanges();

    // Test search functionality
    component.searchControl.setValue('lodash');
    fixture.detectChanges();

    // The filter should be applied (implementation depends on MatTableDataSource)
    expect(component.dataSource.filter).toBeTruthy();
  });

  it('should handle project filtering', () => {
    component.ngOnInit();
    fixture.detectChanges();

    component.projectFilter.setValue(['Project Name']);
    fixture.detectChanges();

    expect(component.dataSource.filter).toBeTruthy();
  });

  it('should handle status filtering', () => {
    component.ngOnInit();
    fixture.detectChanges();

    component.statusFilter.setValue(['outdated']);
    fixture.detectChanges();

    expect(component.dataSource.filter).toBeTruthy();
  });

  it('should handle risk filtering', () => {
    component.ngOnInit();
    fixture.detectChanges();

    component.riskFilter.setValue(['high']);
    fixture.detectChanges();

    expect(component.dataSource.filter).toBeTruthy();
  });

  it('should clear all filters', () => {
    component.searchControl.setValue('test');
    component.projectFilter.setValue(['test']);
    component.statusFilter.setValue(['outdated']);
    component.riskFilter.setValue(['high']);

    component.clearFilters();

    expect(component.searchControl.value).toBe('');
    expect(component.projectFilter.value).toEqual([]);
    expect(component.statusFilter.value).toEqual([]);
    expect(component.riskFilter.value).toEqual([]);
  });

  it('should handle row selection', () => {
    component.ngOnInit();
    fixture.detectChanges();

    const firstItem = component.dataSource.data[0];
    component.selection.select(firstItem);

    expect(component.selection.selected.length).toBe(1);
    expect(component.selection.isSelected(firstItem)).toBe(true);
  });

  it('should handle select all functionality', () => {
    component.ngOnInit();
    fixture.detectChanges();

    expect(component.isAllSelected()).toBe(false);

    component.toggleAllRows();
    expect(component.selection.selected.length).toBe(component.dataSource.filteredData.length);
    expect(component.isAllSelected()).toBe(true);

    component.toggleAllRows();
    expect(component.selection.selected.length).toBe(0);
    expect(component.isAllSelected()).toBe(false);
  });

  it('should get correct package icons', () => {
    expect(component.getPackageIcon('npm')).toBe('javascript');
    expect(component.getPackageIcon('pip')).toBe('code');
    expect(component.getPackageIcon('maven')).toBe('coffee');
    expect(component.getPackageIcon('gradle')).toBe('build');
    expect(component.getPackageIcon('unknown')).toBe('package');
  });

  it('should get correct status classes', () => {
    const itemWithSecurity: DependencyTableItem = {
      ...component.dataSource.data[0],
      securityIssues: 1
    } as any;
    expect(component.getStatusClass(itemWithSecurity)).toBe('security-issue');

    const itemOutdated: DependencyTableItem = {
      currentVersion: '1.0.0',
      latestVersion: '1.0.1',
      securityIssues: 0
    } as any;
    expect(component.getStatusClass(itemOutdated)).toBe('outdated');

    const itemUpToDate: DependencyTableItem = {
      currentVersion: '1.0.0',
      latestVersion: '1.0.0',
      securityIssues: 0
    } as any;
    expect(component.getStatusClass(itemUpToDate)).toBe('up-to-date');
  });

  it('should get correct status icons', () => {
    const itemWithSecurity: DependencyTableItem = {
      securityIssues: 1
    } as any;
    expect(component.getStatusIcon(itemWithSecurity)).toBe('security');

    const itemOutdated: DependencyTableItem = {
      currentVersion: '1.0.0',
      latestVersion: '1.0.1',
      securityIssues: 0
    } as any;
    expect(component.getStatusIcon(itemOutdated)).toBe('update');

    const itemUpToDate: DependencyTableItem = {
      currentVersion: '1.0.0',
      latestVersion: '1.0.0',
      securityIssues: 0
    } as any;
    expect(component.getStatusIcon(itemUpToDate)).toBe('check_circle');
  });

  it('should get correct risk icons', () => {
    expect(component.getRiskIcon('safe')).toBe('check_circle');
    expect(component.getRiskIcon('caution')).toBe('warning');
    expect(component.getRiskIcon('risky')).toBe('error');
    expect(component.getRiskIcon('critical')).toBe('dangerous');
    expect(component.getRiskIcon('unknown')).toBe('help');
  });

  it('should handle action methods', () => {
    spyOn(console, 'log');
    
    const mockItem = component.dataSource.data[0];
    
    component.updateSelected();
    expect(console.log).toHaveBeenCalledWith('Update selected dependencies:', jasmine.any(Array));

    component.scanAll();
    expect(console.log).toHaveBeenCalledWith('Scan all dependencies');

    component.viewDetails(mockItem);
    expect(console.log).toHaveBeenCalledWith('View dependency details:', mockItem);

    component.updateDependency(mockItem);
    expect(console.log).toHaveBeenCalledWith('Update dependency:', mockItem);

    component.viewChangelog(mockItem);
    expect(console.log).toHaveBeenCalledWith('View changelog:', mockItem);

    component.analyzeImpact(mockItem);
    expect(console.log).toHaveBeenCalledWith('Analyze impact:', mockItem);
  });

  it('should handle API errors gracefully', () => {
    mockApiService.getDependencies.and.returnValue(of({ success: false, data: [], message: 'Error loading dependencies' }));

    spyOn(console, 'error');
    component.ngOnInit();
    fixture.detectChanges();

    expect(component.loading).toBe(false);
    expect(component.dataSource.data).toEqual([]);
  });

  it('should clean up subscriptions on destroy', () => {
    spyOn(component['destroy$'], 'next');
    spyOn(component['destroy$'], 'complete');

    component.ngOnDestroy();

    expect(component['destroy$'].next).toHaveBeenCalled();
    expect(component['destroy$'].complete).toHaveBeenCalled();
  });
});
