import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatChipsModule } from '@angular/material/chips';
import { MatDividerModule } from '@angular/material/divider';

import { LoggingService, LogLevel } from '../../core/services/logging.service';
import { ApiService } from '../../core/services/api.service';
import { environment } from '../../../environments/environment';

@Component({
  selector: 'app-logging-test',
  standalone: true,
  imports: [
    CommonModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatProgressSpinnerModule,
    MatChipsModule,
    MatDividerModule
  ],
  template: `
    <div class="logging-test-container">
      <mat-card class="test-header-card">
        <mat-card-header>
          <mat-card-title>
            <mat-icon>bug_report</mat-icon>
            Comprehensive Logging System Test
          </mat-card-title>
          <mat-card-subtitle>
            Testing all log levels and functionality across frontend and backend
          </mat-card-subtitle>
        </mat-card-header>
        <mat-card-content>
          <div class="environment-info">
            <mat-chip-set>
              <mat-chip [color]="environment.production ? 'accent' : 'primary'">
                {{ environment.production ? 'Production' : 'Development' }}
              </mat-chip>
              <mat-chip>{{ environment.dataSource === 'mock' ? 'Mock Data' : 'Real API' }}</mat-chip>
              <mat-chip>Log Level: {{ environment.logging.level }}</mat-chip>
            </mat-chip-set>
          </div>
        </mat-card-content>
      </mat-card>

      <div class="test-grid">
        <!-- Frontend Logging Tests -->
        <mat-card class="test-card">
          <mat-card-header>
            <mat-card-title>Frontend Logging Tests</mat-card-title>
          </mat-card-header>
          <mat-card-content>
            <div class="test-buttons">
              <button mat-raised-button color="primary" (click)="testDebugLogging()">
                <mat-icon>code</mat-icon>
                Test DEBUG Logging
              </button>
              <button mat-raised-button color="primary" (click)="testInfoLogging()">
                <mat-icon>info</mat-icon>
                Test INFO Logging
              </button>
              <button mat-raised-button color="warn" (click)="testWarnLogging()">
                <mat-icon>warning</mat-icon>
                Test WARN Logging
              </button>
              <button mat-raised-button color="warn" (click)="testErrorLogging()">
                <mat-icon>error</mat-icon>
                Test ERROR Logging
              </button>
              <button mat-raised-button color="warn" (click)="testFatalLogging()">
                <mat-icon>dangerous</mat-icon>
                Test FATAL Logging
              </button>
            </div>
          </mat-card-content>
        </mat-card>

        <!-- Performance Logging Tests -->
        <mat-card class="test-card">
          <mat-card-header>
            <mat-card-title>Performance Logging Tests</mat-card-title>
          </mat-card-header>
          <mat-card-content>
            <div class="test-buttons">
              <button mat-raised-button color="accent" (click)="testPerformanceLogging()">
                <mat-icon>timer</mat-icon>
                Test Performance Timing
              </button>
              <button mat-raised-button color="accent" (click)="testApiCallLogging()">
                <mat-icon>api</mat-icon>
                Test API Call Logging
              </button>
              <button mat-raised-button color="accent" (click)="testUserActionLogging()">
                <mat-icon>touch_app</mat-icon>
                Test User Action Logging
              </button>
            </div>
          </mat-card-content>
        </mat-card>

        <!-- Error Handling Tests -->
        <mat-card class="test-card">
          <mat-card-header>
            <mat-card-title>Error Handling Tests</mat-card-title>
          </mat-card-header>
          <mat-card-content>
            <div class="test-buttons">
              <button mat-raised-button color="warn" (click)="testComponentError()">
                <mat-icon>bug_report</mat-icon>
                Test Component Error
              </button>
              <button mat-raised-button color="warn" (click)="testUnhandledError()">
                <mat-icon>error_outline</mat-icon>
                Test Unhandled Error
              </button>
              <button mat-raised-button color="warn" (click)="testPromiseRejection()">
                <mat-icon>cancel</mat-icon>
                Test Promise Rejection
              </button>
            </div>
          </mat-card-content>
        </mat-card>

        <!-- Backend Integration Tests -->
        <mat-card class="test-card">
          <mat-card-header>
            <mat-card-title>Backend Integration Tests</mat-card-title>
          </mat-card-header>
          <mat-card-content>
            <div class="test-buttons">
              <button mat-raised-button color="primary" (click)="testBackendApiCall()">
                <mat-icon>cloud</mat-icon>
                Test Backend API Call
              </button>
              <button mat-raised-button color="primary" (click)="testRemoteLogging()">
                <mat-icon>send</mat-icon>
                Test Remote Logging
              </button>
              <button mat-raised-button color="primary" (click)="testFullStackLogging()">
                <mat-icon>layers</mat-icon>
                Test Full-Stack Logging
              </button>
            </div>
          </mat-card-content>
        </mat-card>

        <!-- Log Management Tests -->
        <mat-card class="test-card">
          <mat-card-header>
            <mat-card-title>Log Management Tests</mat-card-title>
          </mat-card-header>
          <mat-card-content>
            <div class="test-buttons">
              <button mat-raised-button (click)="viewLogs()">
                <mat-icon>list</mat-icon>
                View Recent Logs
              </button>
              <button mat-raised-button (click)="exportLogs()">
                <mat-icon>download</mat-icon>
                Export Logs
              </button>
              <button mat-raised-button (click)="clearLogs()">
                <mat-icon>clear</mat-icon>
                Clear Logs
              </button>
              <button mat-raised-button (click)="changeLogLevel()">
                <mat-icon>tune</mat-icon>
                Change Log Level
              </button>
            </div>
          </mat-card-content>
        </mat-card>

        <!-- Test Results -->
        <mat-card class="test-card full-width">
          <mat-card-header>
            <mat-card-title>Test Results</mat-card-title>
          </mat-card-header>
          <mat-card-content>
            <div class="test-results">
              <p><strong>Check browser console and network tab for logging output</strong></p>
              <p>Recent test results:</p>
              <div class="results-list">
                <div *ngFor="let result of testResults" class="result-item">
                  <mat-icon [color]="result.success ? 'primary' : 'warn'">
                    {{ result.success ? 'check_circle' : 'error' }}
                  </mat-icon>
                  <span>{{ result.message }}</span>
                  <small>{{ result.timestamp | date:'medium' }}</small>
                </div>
              </div>
            </div>
          </mat-card-content>
        </mat-card>
      </div>
    </div>
  `,
  styles: [`
    .logging-test-container {
      padding: 20px;
      max-width: 1200px;
      margin: 0 auto;
    }

    .test-header-card {
      margin-bottom: 20px;
    }

    .environment-info {
      margin-top: 16px;
    }

    .test-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
      gap: 20px;
    }

    .test-card {
      height: fit-content;
    }

    .full-width {
      grid-column: 1 / -1;
    }

    .test-buttons {
      display: flex;
      flex-direction: column;
      gap: 12px;
    }

    .test-buttons button {
      justify-content: flex-start;
      text-align: left;
    }

    .test-buttons mat-icon {
      margin-right: 8px;
    }

    .test-results {
      max-height: 400px;
      overflow-y: auto;
    }

    .results-list {
      display: flex;
      flex-direction: column;
      gap: 8px;
    }

    .result-item {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 8px;
      border-radius: 4px;
      background-color: #f5f5f5;
    }

    .result-item small {
      margin-left: auto;
      color: #666;
    }
  `]
})
export class LoggingTestComponent implements OnInit {
  private logger = inject(LoggingService);
  private apiService = inject(ApiService);

  environment = environment;
  testResults: Array<{message: string, success: boolean, timestamp: Date}> = [];

  ngOnInit(): void {
    this.logger.info('Logging Test Component initialized', {
      environment: environment.production ? 'production' : 'development',
      dataSource: environment.dataSource
    }, 'LoggingTestComponent');
    
    this.addTestResult('Logging Test Component initialized', true);
  }

  // Frontend Logging Tests
  testDebugLogging(): void {
    this.logger.debug('Debug logging test', {
      testType: 'debug',
      timestamp: new Date(),
      data: { key: 'value', number: 42, boolean: true }
    }, 'LoggingTestComponent');
    
    this.addTestResult('DEBUG logging test completed', true);
  }

  testInfoLogging(): void {
    this.logger.info('Info logging test', {
      testType: 'info',
      message: 'This is an informational log entry',
      metadata: { version: '1.0.0', feature: 'logging-test' }
    }, 'LoggingTestComponent');
    
    this.addTestResult('INFO logging test completed', true);
  }

  testWarnLogging(): void {
    this.logger.warn('Warning logging test', {
      testType: 'warn',
      warning: 'This is a test warning message',
      severity: 'medium'
    }, 'LoggingTestComponent');
    
    this.addTestResult('WARN logging test completed', true);
  }

  testErrorLogging(): void {
    this.logger.error('Error logging test', {
      testType: 'error',
      error: 'This is a test error message',
      stack: 'Test stack trace information',
      severity: 'high'
    }, 'LoggingTestComponent');
    
    this.addTestResult('ERROR logging test completed', true);
  }

  testFatalLogging(): void {
    this.logger.fatal('Fatal logging test', {
      testType: 'fatal',
      error: 'This is a test fatal error message',
      critical: true,
      action_required: 'immediate'
    }, 'LoggingTestComponent');
    
    this.addTestResult('FATAL logging test completed', true);
  }

  // Performance Logging Tests
  testPerformanceLogging(): void {
    const timerId = this.logger.startTimer('performance-test');
    
    // Simulate some work
    setTimeout(() => {
      this.logger.endTimer(timerId, 'performance-test');
      this.addTestResult('Performance timing test completed', true);
    }, Math.random() * 1000 + 500);
  }

  testApiCallLogging(): void {
    const startTime = performance.now();
    
    this.logger.logApiCall('GET', '/api/test', 200, 150, {
      testData: 'API call logging test',
      responseSize: 1024
    });
    
    this.addTestResult('API call logging test completed', true);
  }

  testUserActionLogging(): void {
    this.logger.logUserAction('button-click', 'LoggingTestComponent', {
      action: 'test-user-action-logging',
      buttonId: 'user-action-test',
      timestamp: new Date()
    });
    
    this.addTestResult('User action logging test completed', true);
  }

  // Error Handling Tests
  testComponentError(): void {
    try {
      throw new Error('Test component error');
    } catch (error) {
      this.logger.logComponentError('LoggingTestComponent', error as Error, {
        testType: 'component-error',
        context: 'error-handling-test'
      });
      
      this.addTestResult('Component error logging test completed', true);
    }
  }

  testUnhandledError(): void {
    // This will be caught by the global error handler
    setTimeout(() => {
      throw new Error('Test unhandled error');
    }, 100);
    
    this.addTestResult('Unhandled error test triggered (check console)', true);
  }

  testPromiseRejection(): void {
    // This will be caught by the unhandled rejection handler
    Promise.reject(new Error('Test promise rejection'));
    
    this.addTestResult('Promise rejection test triggered (check console)', true);
  }

  // Backend Integration Tests
  testBackendApiCall(): void {
    this.apiService.getProjects().subscribe({
      next: (response) => {
        this.logger.info('Backend API call successful', {
          testType: 'backend-integration',
          response: response.success,
          dataLength: response.data?.length || 0
        }, 'LoggingTestComponent');
        
        this.addTestResult('Backend API call test completed successfully', true);
      },
      error: (error) => {
        this.logger.error('Backend API call failed', {
          testType: 'backend-integration',
          error: error.message
        }, 'LoggingTestComponent');
        
        this.addTestResult('Backend API call test completed with error', false);
      }
    });
  }

  testRemoteLogging(): void {
    // Force a log entry to be sent to backend
    this.logger.error('Remote logging test', {
      testType: 'remote-logging',
      forceRemote: true,
      timestamp: new Date()
    }, 'LoggingTestComponent');
    
    this.addTestResult('Remote logging test completed (check network tab)', true);
  }

  testFullStackLogging(): void {
    const correlationId = 'test_' + Date.now();
    
    this.logger.info('Full-stack logging test started', {
      correlationId,
      testType: 'full-stack',
      phase: 'frontend-start'
    }, 'LoggingTestComponent');
    
    // Make API call that should generate backend logs
    this.apiService.getProjects().subscribe({
      next: (response) => {
        this.logger.info('Full-stack logging test completed', {
          correlationId,
          testType: 'full-stack',
          phase: 'frontend-complete',
          backendResponse: response.success
        }, 'LoggingTestComponent');
        
        this.addTestResult('Full-stack logging test completed', true);
      },
      error: (error) => {
        this.logger.error('Full-stack logging test failed', {
          correlationId,
          testType: 'full-stack',
          phase: 'frontend-error',
          error: error.message
        }, 'LoggingTestComponent');
        
        this.addTestResult('Full-stack logging test failed', false);
      }
    });
  }

  // Log Management Tests
  viewLogs(): void {
    const logs = this.logger.getLogs(LogLevel.DEBUG, 10);
    console.log('Recent logs:', logs);
    
    this.logger.info('Viewed recent logs', {
      logCount: logs.length,
      testType: 'log-management'
    }, 'LoggingTestComponent');
    
    this.addTestResult(`Viewed ${logs.length} recent logs (check console)`, true);
  }

  exportLogs(): void {
    const exportData = this.logger.exportLogs();
    const blob = new Blob([exportData], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    
    const a = document.createElement('a');
    a.href = url;
    a.download = `logs_${new Date().toISOString()}.json`;
    a.click();
    
    URL.revokeObjectURL(url);
    
    this.logger.info('Exported logs', {
      exportSize: exportData.length,
      testType: 'log-management'
    }, 'LoggingTestComponent');
    
    this.addTestResult('Logs exported successfully', true);
  }

  clearLogs(): void {
    this.logger.clearLogs();
    this.addTestResult('Logs cleared successfully', true);
  }

  changeLogLevel(): void {
    const currentLevel = LogLevel.INFO;
    const newLevel = LogLevel.DEBUG;
    
    this.logger.setLogLevel(newLevel);
    this.addTestResult(`Log level changed from ${LogLevel[currentLevel]} to ${LogLevel[newLevel]}`, true);
  }

  private addTestResult(message: string, success: boolean): void {
    this.testResults.unshift({
      message,
      success,
      timestamp: new Date()
    });
    
    // Keep only last 20 results
    if (this.testResults.length > 20) {
      this.testResults = this.testResults.slice(0, 20);
    }
  }
}
