import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpHeaders, HttpErrorResponse } from '@angular/common/http';
import { Observable, throwError, of } from 'rxjs';
import { retry, catchError, tap } from 'rxjs/operators';
import { environment } from '../../../environments/environment';
import { LoggingService } from './logging.service';

// API Response interface
export interface ApiResponse<T> {
  data: T;
  success: boolean;
  message?: string;
  error?: string;
}

// Project interface
export interface Project {
  id: number;
  name: string;
  type: 'npm' | 'pip' | 'maven' | 'gradle';
  path: string;
  configFile: string;
  enabled: boolean;
  createdAt: Date;
  updatedAt: Date;
  lastScan?: Date;
}

// Dependency interface
export interface Dependency {
  id: number;
  projectId: number;
  name: string;
  currentVersion: string;
  latestVersion?: string;
  type: 'direct' | 'dev' | 'peer' | 'optional';
  status: 'up-to-date' | 'outdated' | 'vulnerable' | 'unknown';
  securityIssues?: SecurityIssue[];
  aiAnalysis?: AIAnalysis;
}

// Security Issue interface
export interface SecurityIssue {
  cve: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  description: string;
  fixedVersion?: string;
  publishedAt: Date;
}

// AI Analysis interface
export interface AIAnalysis {
  riskLevel: 'low' | 'medium' | 'high' | 'critical';
  confidence: number;
  recommendation: string;
  breakingChanges: boolean;
  updateSafety: 'safe' | 'caution' | 'risky' | 'critical';
}

@Injectable({
  providedIn: 'root'
})
export class ApiService {
  private readonly baseUrl = environment.apiUrl;
  private readonly useMockData = environment.dataSource === 'mock';
  private readonly defaultHeaders = new HttpHeaders({
    'Content-Type': 'application/json',
    'Accept': 'application/json'
  });

  constructor(
    private http: HttpClient,
    private logger: LoggingService
  ) {
    this.logger.info('API Service initialized', {
      dataSource: environment.dataSource,
      baseUrl: this.baseUrl,
      useMockData: this.useMockData
    }, 'ApiService');
  }

  // Project Management
  getProjects(): Observable<ApiResponse<Project[]>> {
    const startTime = performance.now();
    this.logger.debug('Getting projects', { useMockData: this.useMockData }, 'ApiService');
    
    if (this.useMockData) {
      return this.getMockProjects().pipe(
        tap(() => {
          const duration = performance.now() - startTime;
          this.logger.info('Projects retrieved (mock)', { duration }, 'ApiService');
        })
      );
    }
    
    return this.http.get<ApiResponse<Project[]>>(`${this.baseUrl}/projects`, { headers: this.defaultHeaders })
      .pipe(
        tap((response) => {
          const duration = performance.now() - startTime;
          this.logger.logApiCall('GET', `${this.baseUrl}/projects`, 200, duration, response);
        }),
        retry(environment.api.retryAttempts),
        catchError(this.handleError)
      );
  }

  getProject(id: number): Observable<ApiResponse<Project>> {
    if (this.useMockData) {
      return this.getMockProject(id);
    }
    
    return this.http.get<ApiResponse<Project>>(`${this.baseUrl}/projects/${id}`, { headers: this.defaultHeaders })
      .pipe(
        retry(environment.api.retryAttempts),
        catchError(this.handleError)
      );
  }

  createProject(project: Partial<Project>): Observable<ApiResponse<Project>> {
    if (this.useMockData) {
      return this.createMockProject(project);
    }
    
    return this.http.post<ApiResponse<Project>>(`${this.baseUrl}/projects`, project, { headers: this.defaultHeaders })
      .pipe(catchError(this.handleError));
  }

  updateProject(id: number, project: Partial<Project>): Observable<ApiResponse<Project>> {
    if (this.useMockData) {
      return this.updateMockProject(id, project);
    }
    
    return this.http.put<ApiResponse<Project>>(`${this.baseUrl}/projects/${id}`, project, { headers: this.defaultHeaders })
      .pipe(catchError(this.handleError));
  }

  deleteProject(id: number): Observable<ApiResponse<void>> {
    if (this.useMockData) {
      return this.deleteMockProject(id);
    }
    
    return this.http.delete<ApiResponse<void>>(`${this.baseUrl}/projects/${id}`, { headers: this.defaultHeaders })
      .pipe(catchError(this.handleError));
  }

  // Dependency Management
  getDependencies(projectId?: number): Observable<ApiResponse<Dependency[]>> {
    if (this.useMockData) {
      return this.getMockDependencies(projectId);
    }
    
    const url = projectId ? 
      `${this.baseUrl}/projects/${projectId}/dependencies` : 
      `${this.baseUrl}/dependencies`;
    return this.http.get<ApiResponse<Dependency[]>>(url, { headers: this.defaultHeaders })
      .pipe(
        retry(environment.api.retryAttempts),
        catchError(this.handleError)
      );
  }

  scanProject(projectId: number): Observable<ApiResponse<any>> {
    if (this.useMockData) {
      return this.scanMockProject(projectId);
    }
    
    return this.http.post<ApiResponse<any>>(`${this.baseUrl}/projects/${projectId}/scan`, {}, { headers: this.defaultHeaders })
      .pipe(catchError(this.handleError));
  }

  updateDependency(projectId: number, dependencyId: number): Observable<ApiResponse<any>> {
    if (this.useMockData) {
      return this.updateMockDependency(projectId, dependencyId);
    }
    
    return this.http.post<ApiResponse<any>>(`${this.baseUrl}/projects/${projectId}/dependencies/${dependencyId}/update`, {}, { headers: this.defaultHeaders })
      .pipe(catchError(this.handleError));
  }

  // Analytics
  getAnalytics(projectId?: number): Observable<ApiResponse<any>> {
    if (this.useMockData) {
      return this.getMockAnalytics(projectId);
    }
    
    const url = projectId ? 
      `${this.baseUrl}/projects/${projectId}/analytics` : 
      `${this.baseUrl}/analytics`;
    return this.http.get<ApiResponse<any>>(url, { headers: this.defaultHeaders })
      .pipe(
        retry(environment.api.retryAttempts),
        catchError(this.handleError)
      );
  }

  // AI Insights
  getAIInsights(projectId?: number): Observable<ApiResponse<any[]>> {
    if (this.useMockData) {
      return this.getMockAIInsightsList(projectId);
    }
    
    const url = projectId ? 
      `${this.baseUrl}/projects/${projectId}/ai/insights` : 
      `${this.baseUrl}/ai/insights`;
    
    return this.http.get<ApiResponse<any[]>>(url, { headers: this.defaultHeaders })
      .pipe(
        retry(environment.api.retryAttempts),
        catchError(this.handleError)
      );
  }

  generateAIInsights(projectId?: number): Observable<ApiResponse<any>> {
    if (this.useMockData) {
      return this.getMockAIInsights(projectId);
    }
    
    const url = projectId ? 
      `${this.baseUrl}/projects/${projectId}/ai/generate` : 
      `${this.baseUrl}/ai/generate`;
    
    return this.http.post<ApiResponse<any>>(url, {}, { headers: this.defaultHeaders })
      .pipe(
        retry(environment.api.retryAttempts),
        catchError(this.handleError)
      );
  }

  // System Status
  getSystemStatus(): Observable<ApiResponse<any>> {
    if (this.useMockData) {
      return this.getMockSystemStatus();
    }
    
    return this.http.get<ApiResponse<any>>(`${this.baseUrl}/status`, { headers: this.defaultHeaders })
      .pipe(
        retry(environment.api.retryAttempts),
        catchError(this.handleError)
      );
  }

  // Mock Data Methods
  private getMockProjects(): Observable<ApiResponse<Project[]>> {
    const mockProjects: Project[] = [
      {
        id: 1,
        name: 'Frontend Dashboard',
        type: 'npm',
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
        type: 'pip',
        path: '/app/backend',
        configFile: 'requirements.txt',
        enabled: true,
        createdAt: new Date(),
        updatedAt: new Date(),
        lastScan: new Date()
      },
      {
        id: 3,
        name: 'Data Processing Service',
        type: 'maven',
        path: '/app/data-service',
        configFile: 'pom.xml',
        enabled: true,
        createdAt: new Date(),
        updatedAt: new Date(),
        lastScan: new Date()
      }
    ];

    return of({
      success: true,
      data: mockProjects,
      message: 'Mock projects loaded successfully'
    });
  }

  private getMockProject(id: number): Observable<ApiResponse<Project>> {
    const mockProject: Project = {
      id,
      name: `Project ${id}`,
      type: 'npm',
      path: `/app/project-${id}`,
      configFile: 'package.json',
      enabled: true,
      createdAt: new Date(),
      updatedAt: new Date(),
      lastScan: new Date()
    };

    return of({
      success: true,
      data: mockProject,
      message: 'Mock project loaded successfully'
    });
  }

  private createMockProject(project: Partial<Project>): Observable<ApiResponse<Project>> {
    const newProject: Project = {
      id: Math.floor(Math.random() * 1000),
      name: project.name || 'New Project',
      type: project.type || 'npm',
      path: project.path || '/app/new-project',
      configFile: project.configFile || 'package.json',
      enabled: project.enabled !== undefined ? project.enabled : true,
      createdAt: new Date(),
      updatedAt: new Date()
    };

    return of({
      success: true,
      data: newProject,
      message: 'Mock project created successfully'
    });
  }

  private updateMockProject(id: number, project: Partial<Project>): Observable<ApiResponse<Project>> {
    const updatedProject: Project = {
      id,
      name: project.name || `Updated Project ${id}`,
      type: project.type || 'npm',
      path: project.path || `/app/project-${id}`,
      configFile: project.configFile || 'package.json',
      enabled: project.enabled !== undefined ? project.enabled : true,
      createdAt: new Date(Date.now() - 86400000), // Yesterday
      updatedAt: new Date()
    };

    return of({
      success: true,
      data: updatedProject,
      message: 'Mock project updated successfully'
    });
  }

  private deleteMockProject(id: number): Observable<ApiResponse<void>> {
    return of({
      success: true,
      data: undefined as any,
      message: `Mock project ${id} deleted successfully`
    });
  }

  private getMockDependencies(projectId?: number): Observable<ApiResponse<Dependency[]>> {
    const mockDependencies: Dependency[] = [
      {
        id: 1,
        projectId: projectId || 1,
        name: 'react',
        currentVersion: '18.2.0',
        latestVersion: '18.2.0',
        type: 'direct',
        status: 'up-to-date'
      },
      {
        id: 2,
        projectId: projectId || 1,
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
            publishedAt: new Date()
          }
        ]
      },
      {
        id: 3,
        projectId: projectId || 2,
        name: 'flask',
        currentVersion: '2.0.1',
        latestVersion: '2.3.3',
        type: 'direct',
        status: 'outdated'
      },
      {
        id: 4,
        projectId: projectId || 2,
        name: 'requests',
        currentVersion: '2.28.1',
        latestVersion: '2.31.0',
        type: 'direct',
        status: 'outdated'
      }
    ];

    const filteredDependencies = projectId ? 
      mockDependencies.filter(dep => dep.projectId === projectId) : 
      mockDependencies;

    return of({
      success: true,
      data: filteredDependencies,
      message: 'Mock dependencies loaded successfully'
    });
  }

  private scanMockProject(projectId: number): Observable<ApiResponse<any>> {
    return of({
      success: true,
      data: {
        projectId,
        status: 'completed',
        dependenciesScanned: 25,
        vulnerabilitiesFound: 3,
        outdatedDependencies: 8
      },
      message: 'Mock project scan completed successfully'
    });
  }

  private updateMockDependency(projectId: number, dependencyId: number): Observable<ApiResponse<any>> {
    return of({
      success: true,
      data: {
        projectId,
        dependencyId,
        status: 'updated',
        newVersion: '1.0.0'
      },
      message: 'Mock dependency updated successfully'
    });
  }

  private getMockAnalytics(projectId?: number): Observable<ApiResponse<any>> {
    return of({
      success: true,
      data: {
        totalProjects: 5,
        totalDependencies: 127,
        outdatedDependencies: 23,
        vulnerabilities: 8,
        healthScore: 78,
        trends: {
          dependencyGrowth: 12,
          securityImprovements: 5,
          updateCompliance: 85
        }
      },
      message: 'Mock analytics loaded successfully'
    });
  }

  private getMockAIInsights(projectId?: number): Observable<ApiResponse<any>> {
    return of({
      success: true,
      data: {
        insights: 'Mock AI insights generated',
        recommendations: [
          'Update lodash to fix security vulnerability',
          'Consider upgrading React to latest stable version',
          'Review outdated dependencies for compatibility issues'
        ],
        riskAssessment: 'medium',
        confidence: 0.85
      },
      message: 'AI insights generated successfully'
    });
  }

  private getMockAIInsightsList(projectId?: number): Observable<ApiResponse<any[]>> {
    const mockInsights = [
      {
        id: 1,
        projectId: projectId || 1,
        type: 'security',
        title: 'Security Vulnerability Detected',
        description: 'lodash package has a known security vulnerability',
        severity: 'high',
        recommendation: 'Update to version 4.17.21 or higher',
        createdAt: new Date()
      },
      {
        id: 2,
        projectId: projectId || 1,
        type: 'update',
        title: 'Outdated Dependencies',
        description: 'Several dependencies are outdated and should be updated',
        severity: 'medium',
        recommendation: 'Review and update outdated packages',
        createdAt: new Date()
      }
    ];

    return of({
      success: true,
      data: mockInsights,
      message: 'Mock AI insights list loaded successfully'
    });
  }

  private getMockSystemStatus(): Observable<ApiResponse<any>> {
    return of({
      success: true,
      data: {
        status: 'healthy',
        uptime: 3600,
        projectsMonitored: 5,
        backgroundScansActive: 2,
        lastScan: new Date(),
        systemHealth: {
          cpu: 45,
          memory: 62,
          disk: 78
        }
      },
      message: 'System status retrieved successfully'
    });
  }

  // Error handling with context (simplified for production build)
  private handleErrorWithContext(context: string) {
    return this.handleError;
  }

  // Error handling compatible with RxJS catchError operator
  private handleError = (error: HttpErrorResponse): Observable<never> => {
    let errorMessage = 'An unexpected error occurred';
    
    if (error.error instanceof ErrorEvent) {
      // Client-side error
      errorMessage = `Error: ${error.error.message}`;
      this.logger.error('Client-side API error', {
        message: error.error.message,
        filename: error.error.filename,
        lineno: error.error.lineno,
        colno: error.error.colno
      }, 'ApiService');
    } else {
      // Server-side error
      errorMessage = error.error?.message || `Error Code: ${error.status}\nMessage: ${error.message}`;
      this.logger.error('Server-side API error', {
        status: error.status,
        statusText: error.statusText,
        message: error.message,
        url: error.url,
        error: error.error
      }, 'ApiService');
    }
    
    return throwError(() => new Error(errorMessage));
  }
}
