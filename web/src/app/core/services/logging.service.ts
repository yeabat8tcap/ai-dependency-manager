import { Injectable } from '@angular/core';
import { environment } from '../../../environments/environment';

export enum LogLevel {
  DEBUG = 0,
  INFO = 1,
  WARN = 2,
  ERROR = 3,
  FATAL = 4
}

export interface LogEntry {
  timestamp: Date;
  level: LogLevel;
  message: string;
  data?: any;
  source?: string;
  userId?: string;
  sessionId?: string;
  requestId?: string;
}

@Injectable({
  providedIn: 'root'
})
export class LoggingService {
  private logs: LogEntry[] = [];
  private maxLogEntries = 1000;
  private sessionId: string;
  private logLevel: LogLevel;

  constructor() {
    this.sessionId = this.generateSessionId();
    this.logLevel = environment.production ? LogLevel.INFO : LogLevel.DEBUG;
    this.initializeLogging();
  }

  private generateSessionId(): string {
    return 'session_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
  }

  private initializeLogging(): void {
    // Capture unhandled errors
    window.addEventListener('error', (event) => {
      this.error('Unhandled JavaScript Error', {
        message: event.message,
        filename: event.filename,
        lineno: event.lineno,
        colno: event.colno,
        error: event.error
      });
    });

    // Capture unhandled promise rejections
    window.addEventListener('unhandledrejection', (event) => {
      this.error('Unhandled Promise Rejection', {
        reason: event.reason,
        promise: event.promise
      });
    });

    this.info('Logging service initialized', { sessionId: this.sessionId });
  }

  private shouldLog(level: LogLevel): boolean {
    return level >= this.logLevel;
  }

  private createLogEntry(level: LogLevel, message: string, data?: any, source?: string): LogEntry {
    return {
      timestamp: new Date(),
      level,
      message,
      data,
      source,
      sessionId: this.sessionId,
      requestId: this.generateRequestId()
    };
  }

  private generateRequestId(): string {
    return 'req_' + Date.now() + '_' + Math.random().toString(36).substr(2, 6);
  }

  private addLogEntry(entry: LogEntry): void {
    this.logs.push(entry);
    
    // Maintain max log entries
    if (this.logs.length > this.maxLogEntries) {
      this.logs = this.logs.slice(-this.maxLogEntries);
    }

    // Console output
    this.outputToConsole(entry);

    // Send to backend if in production
    if (environment.production && entry.level >= LogLevel.WARN) {
      this.sendToBackend(entry);
    }
  }

  private outputToConsole(entry: LogEntry): void {
    const timestamp = entry.timestamp.toISOString();
    const levelName = LogLevel[entry.level];
    const prefix = `[${timestamp}] [${levelName}] [${entry.sessionId}]`;
    
    const message = entry.source ? 
      `${prefix} [${entry.source}] ${entry.message}` : 
      `${prefix} ${entry.message}`;

    switch (entry.level) {
      case LogLevel.DEBUG:
        console.debug(message, entry.data || '');
        break;
      case LogLevel.INFO:
        console.info(message, entry.data || '');
        break;
      case LogLevel.WARN:
        console.warn(message, entry.data || '');
        break;
      case LogLevel.ERROR:
      case LogLevel.FATAL:
        console.error(message, entry.data || '');
        break;
    }
  }

  private async sendToBackend(entry: LogEntry): Promise<void> {
    try {
      const response = await fetch(`${environment.apiUrl}/logs`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(entry)
      });

      if (!response.ok) {
        console.warn('Failed to send log to backend:', response.statusText);
      }
    } catch (error) {
      console.warn('Error sending log to backend:', error);
    }
  }

  // Public logging methods
  debug(message: string, data?: any, source?: string): void {
    if (this.shouldLog(LogLevel.DEBUG)) {
      const entry = this.createLogEntry(LogLevel.DEBUG, message, data, source);
      this.addLogEntry(entry);
    }
  }

  info(message: string, data?: any, source?: string): void {
    if (this.shouldLog(LogLevel.INFO)) {
      const entry = this.createLogEntry(LogLevel.INFO, message, data, source);
      this.addLogEntry(entry);
    }
  }

  warn(message: string, data?: any, source?: string): void {
    if (this.shouldLog(LogLevel.WARN)) {
      const entry = this.createLogEntry(LogLevel.WARN, message, data, source);
      this.addLogEntry(entry);
    }
  }

  error(message: string, data?: any, source?: string): void {
    if (this.shouldLog(LogLevel.ERROR)) {
      const entry = this.createLogEntry(LogLevel.ERROR, message, data, source);
      this.addLogEntry(entry);
    }
  }

  fatal(message: string, data?: any, source?: string): void {
    const entry = this.createLogEntry(LogLevel.FATAL, message, data, source);
    this.addLogEntry(entry);
  }

  // Utility methods
  setLogLevel(level: LogLevel): void {
    this.logLevel = level;
    this.info('Log level changed', { newLevel: LogLevel[level] });
  }

  getLogs(level?: LogLevel, limit?: number): LogEntry[] {
    let filteredLogs = level !== undefined ? 
      this.logs.filter(log => log.level >= level) : 
      this.logs;

    return limit ? filteredLogs.slice(-limit) : filteredLogs;
  }

  clearLogs(): void {
    this.logs = [];
    this.info('Logs cleared');
  }

  exportLogs(): string {
    return JSON.stringify(this.logs, null, 2);
  }

  // Performance logging
  startTimer(label: string): string {
    const timerId = `timer_${Date.now()}_${Math.random().toString(36).substr(2, 6)}`;
    this.debug(`Timer started: ${label}`, { timerId, label });
    return timerId;
  }

  endTimer(timerId: string, label: string): void {
    const duration = performance.now();
    this.info(`Timer ended: ${label}`, { timerId, label, duration });
  }

  // API call logging
  logApiCall(method: string, url: string, status?: number, duration?: number, data?: any): void {
    const logData = {
      method,
      url,
      status,
      duration,
      data: data ? JSON.stringify(data).substring(0, 500) : undefined
    };

    if (status && status >= 400) {
      this.error(`API call failed: ${method} ${url}`, logData);
    } else {
      this.info(`API call: ${method} ${url}`, logData);
    }
  }

  // User action logging
  logUserAction(action: string, component: string, data?: any): void {
    this.info(`User action: ${action}`, {
      action,
      component,
      data,
      timestamp: new Date().toISOString()
    });
  }

  // Error boundary logging
  logComponentError(componentName: string, error: Error, errorInfo?: any): void {
    this.error(`Component error in ${componentName}`, {
      componentName,
      error: {
        name: error.name,
        message: error.message,
        stack: error.stack
      },
      errorInfo
    });
  }
}
