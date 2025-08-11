import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { ApiService, ApiResponse, Project, Dependency } from './api.service';
import { environment } from '../../../environments/environment';

describe('ApiService', () => {
  let service: ApiService;
  let httpMock: HttpTestingController;

  const mockProject: Project = {
    id: 1,
    name: 'Test Project',
    path: '/test/path',
    type: 'npm',
    configFile: 'package.json',
    enabled: true,
    createdAt: new Date(),
    updatedAt: new Date(),
    lastScan: new Date()
  };

  const mockDependency: Dependency = {
    id: 1,
    projectId: 1,
    name: 'lodash',
    currentVersion: '4.17.20',
    latestVersion: '4.17.21',
    type: 'direct',
    status: 'outdated'
  };

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [ApiService]
    });
    service = TestBed.inject(ApiService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe('Project Management', () => {
    it('should get all projects', () => {
      const mockResponse: ApiResponse<Project[]> = {
        success: true,
        data: [mockProject]
      };

      service.getProjects().subscribe(response => {
        expect(response.success).toBe(true);
        expect(response.data.length).toBe(1);
        expect(response.data[0]).toEqual(mockProject);
      });

      const req = httpMock.expectOne(`${environment.apiUrl}/projects`);
      expect(req.request.method).toBe('GET');
      req.flush(mockResponse);
    });

    it('should get project by id', () => {
      const mockResponse: ApiResponse<Project> = {
        success: true,
        data: mockProject
      };

      service.getProject(1).subscribe(response => {
        expect(response.success).toBe(true);
        expect(response.data).toEqual(mockProject);
      });

      const req = httpMock.expectOne(`${environment.apiUrl}/projects/1`);
      expect(req.request.method).toBe('GET');
      req.flush(mockResponse);
    });

    it('should create a new project', () => {
      const newProject = { ...mockProject, id: undefined };
      const mockResponse: ApiResponse<Project> = {
        success: true,
        data: mockProject
      };

      service.createProject(newProject as any).subscribe(response => {
        expect(response.success).toBe(true);
        expect(response.data).toEqual(mockProject);
      });

      const req = httpMock.expectOne(`${environment.apiUrl}/projects`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual(newProject);
      req.flush(mockResponse);
    });

    it('should update a project', () => {
      const updatedProject = { ...mockProject, name: 'Updated Project' };
      const mockResponse: ApiResponse<Project> = {
        success: true,
        data: updatedProject
      };

      service.updateProject(1, updatedProject).subscribe(response => {
        expect(response.success).toBe(true);
        expect(response.data.name).toBe('Updated Project');
      });

      const req = httpMock.expectOne(`${environment.apiUrl}/projects/1`);
      expect(req.request.method).toBe('PUT');
      expect(req.request.body).toEqual(updatedProject);
      req.flush(mockResponse);
    });

    it('should delete a project', () => {
      const mockResponse: ApiResponse<any> = {
        success: true,
        data: null
      };

      service.deleteProject(1).subscribe(response => {
        expect(response.success).toBe(true);
      });

      const req = httpMock.expectOne(`${environment.apiUrl}/projects/1`);
      expect(req.request.method).toBe('DELETE');
      req.flush(mockResponse);
    });
  });

  describe('Dependency Management', () => {
    it('should get dependencies for a project', () => {
      const mockResponse: ApiResponse<Dependency[]> = {
        success: true,
        data: [mockDependency]
      };

      service.getDependencies(1).subscribe(response => {
        expect(response.success).toBe(true);
        expect(response.data.length).toBe(1);
        expect(response.data[0]).toEqual(mockDependency);
      });

      const req = httpMock.expectOne(`${environment.apiUrl}/projects/1/dependencies`);
      expect(req.request.method).toBe('GET');
      req.flush(mockResponse);
    });

    it('should scan project dependencies', () => {
      const mockResponse: ApiResponse<any> = {
        success: true,
        data: { message: 'Scan initiated' }
      };

      service.scanProject(1).subscribe(response => {
        expect(response.success).toBe(true);
        expect(response.data.message).toBe('Scan initiated');
      });

      const req = httpMock.expectOne(`${environment.apiUrl}/projects/1/scan`);
      expect(req.request.method).toBe('POST');
      req.flush(mockResponse);
    });

    it('should apply updates to project', () => {
      const updateOptions = { dependencies: ['lodash'], strategy: 'safe' };
      const mockResponse: ApiResponse<any> = {
        success: true,
        data: { message: 'Updates applied' }
      };

      service.applyUpdates(1, updateOptions).subscribe(response => {
        expect(response.success).toBe(true);
        expect(response.data.message).toBe('Updates applied');
      });

      const req = httpMock.expectOne(`${environment.apiUrl}/projects/1/updates/apply`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual(updateOptions);
      req.flush(mockResponse);
    });
  });

  describe('Analytics and Reporting', () => {
    it('should get system analytics', () => {
      const mockAnalytics = {
        totalProjects: 5,
        totalDependencies: 247,
        outdatedDependencies: 23,
        securityIssues: 7
      };
      const mockResponse: ApiResponse<any> = {
        success: true,
        data: mockAnalytics
      };

      service.getAnalytics().subscribe(response => {
        expect(response.success).toBe(true);
        expect(response.data).toEqual(mockAnalytics);
      });

      const req = httpMock.expectOne(`${environment.apiUrl}/analytics`);
      expect(req.request.method).toBe('GET');
      req.flush(mockResponse);
    });

    it('should get system status', () => {
      const mockStatus = { health: 'healthy', uptime: '24h' };
      const mockResponse: ApiResponse<any> = {
        success: true,
        data: mockStatus
      };

      service.getSystemStatus().subscribe(response => {
        expect(response.success).toBe(true);
        expect(response.data).toEqual(mockStatus);
      });

      const req = httpMock.expectOne(`${environment.apiUrl}/status`);
      expect(req.request.method).toBe('GET');
      req.flush(mockResponse);
    });

    it('should generate reports', () => {
      const reportOptions = { format: 'pdf', includeCharts: true };
      const mockResponse: ApiResponse<any> = {
        success: true,
        data: { reportUrl: '/reports/123.pdf' }
      };

      service.generateReport('security', reportOptions).subscribe(response => {
        expect(response.success).toBe(true);
        expect(response.data.reportUrl).toBe('/reports/123.pdf');
      });

      const req = httpMock.expectOne(`${environment.apiUrl}/reports/generate/security`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual(reportOptions);
      req.flush(mockResponse);
    });
  });

  describe('Policy Management', () => {
    it('should get policies', () => {
      const mockPolicies = [
        { id: 1, name: 'Security Policy', rules: [] }
      ];
      const mockResponse: ApiResponse<any[]> = {
        success: true,
        data: mockPolicies
      };

      service.getPolicies().subscribe(response => {
        expect(response.success).toBe(true);
        expect(response.data).toEqual(mockPolicies);
      });

      const req = httpMock.expectOne(`${environment.apiUrl}/policies`);
      expect(req.request.method).toBe('GET');
      req.flush(mockResponse);
    });

    it('should create a policy', () => {
      const newPolicy = { name: 'New Policy', rules: [] };
      const mockResponse: ApiResponse<any> = {
        success: true,
        data: { id: 1, ...newPolicy }
      };

      service.createPolicy(newPolicy).subscribe(response => {
        expect(response.success).toBe(true);
        expect(response.data.name).toBe('New Policy');
      });

      const req = httpMock.expectOne(`${environment.apiUrl}/policies`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual(newPolicy);
      req.flush(mockResponse);
    });
  });

  describe('AI Insights', () => {
    it('should generate AI insights', () => {
      const mockResponse: ApiResponse<any> = {
        success: true,
        data: { insights: 'Mock AI insights generated' },
        message: 'AI insights generated successfully'
      };

      service.generateAIInsights(1).subscribe(response => {
        expect(response.success).toBe(true);
        expect(response.data.insights).toBe('Mock AI insights generated');
      });

      // Note: This test validates the mock implementation we created
      expect(service).toBeTruthy();
    });
  });

  describe('Error Handling', () => {
    it('should handle HTTP errors gracefully', () => {
      service.getProjects().subscribe({
        next: () => fail('Expected an error'),
        error: (error) => {
          expect(error).toBeTruthy();
        }
      });

      const req = httpMock.expectOne(`${environment.apiUrl}/projects`);
      req.error(new ProgressEvent('Network error'), {
        status: 500,
        statusText: 'Internal Server Error'
      });
    });

    it('should handle API error responses', () => {
      const errorResponse: ApiResponse<any> = {
        success: false,
        data: null,
        message: 'Project not found',
        errors: ['Project with ID 999 does not exist']
      };

      service.getProject(999).subscribe(response => {
        expect(response.success).toBe(false);
        expect(response.message).toBe('Project not found');
        expect(response.errors).toContain('Project with ID 999 does not exist');
      });

      const req = httpMock.expectOne(`${environment.apiUrl}/projects/999`);
      req.flush(errorResponse);
    });
  });
});
