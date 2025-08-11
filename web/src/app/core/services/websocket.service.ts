import { Injectable, OnDestroy } from '@angular/core';
import { Observable, Subject, BehaviorSubject, interval, of } from 'rxjs';
import { webSocket, WebSocketSubject } from 'rxjs/webSocket';
import { retry, catchError, takeWhile, tap } from 'rxjs/operators';
import { environment } from '../../../environments/environment';

export interface WebSocketMessage {
  type: 'scan_progress' | 'update_complete' | 'security_alert' | 'system_status' | 'notification';
  payload: any;
  timestamp: Date;
}

export interface ScanProgressMessage {
  projectId: number;
  projectName: string;
  progress: number;
  currentDependency?: string;
  totalDependencies: number;
  scannedDependencies: number;
  status: 'scanning' | 'complete' | 'error';
}

export interface UpdateCompleteMessage {
  projectId: number;
  projectName: string;
  updatesApplied: number;
  updatesSuccessful: number;
  updatesFailed: number;
  duration: number;
}

export interface SecurityAlertMessage {
  projectId: number;
  dependencyName: string;
  cve: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  description: string;
}

export interface SystemStatusMessage {
  status: 'healthy' | 'warning' | 'error';
  uptime: number;
  projectsMonitored: number;
  backgroundScansActive: number;
  lastScan?: Date;
}

@Injectable({
  providedIn: 'root'
})
export class WebSocketService implements OnDestroy {
  private socket$: WebSocketSubject<any> | null = null;
  private messagesSubject$ = new Subject<WebSocketMessage>();
  private reconnectInterval = environment.websocket.reconnectInterval;
  private maxReconnectAttempts = environment.websocket.maxReconnectAttempts;
  private reconnectAttempts = 0;
  private isConnected$ = new BehaviorSubject<boolean>(false);
  private readonly wsUrl = environment.wsUrl;
  private readonly useMockData = environment.dataSource === 'mock';
  private connectionStatusSubject$ = new BehaviorSubject<'connecting' | 'connected' | 'disconnected' | 'error'>('disconnected');

  public messages$ = this.messagesSubject$.asObservable();
  public connectionStatus$ = this.connectionStatusSubject$.asObservable();

  constructor() {
    this.connect();
  }

  private connect(): void {
    if (this.socket$) {
      return;
    }

    this.connectionStatusSubject$.next('connecting');

    this.socket$ = webSocket({
      url: environment.wsUrl,
      openObserver: {
        next: () => {
          console.log('WebSocket connection opened');
          this.connectionStatusSubject$.next('connected');
        }
      },
      closeObserver: {
        next: () => {
          console.log('WebSocket connection closed');
          this.connectionStatusSubject$.next('disconnected');
          this.socket$ = null;
          // Attempt to reconnect after 5 seconds
          setTimeout(() => this.connect(), 5000);
        }
      }
    });

    this.socket$.subscribe({
      next: (message) => this.handleMessage(message),
      error: (error) => {
        console.error('WebSocket error:', error);
        this.connectionStatusSubject$.next('error');
        this.socket$ = null;
        // Attempt to reconnect after 5 seconds
        setTimeout(() => this.connect(), 5000);
      }
    });
  }

  private handleMessage(message: any): void {
    try {
      const parsedMessage: WebSocketMessage = {
        type: message.type,
        payload: message.payload,
        timestamp: new Date(message.timestamp || Date.now())
      };
      this.messagesSubject$.next(parsedMessage);
    } catch (error) {
      console.error('Error parsing WebSocket message:', error);
    }
  }

  public sendMessage(message: any): void {
    if (this.socket$ && this.connectionStatusSubject$.value === 'connected') {
      this.socket$.next(message);
    } else {
      console.warn('WebSocket is not connected. Message not sent:', message);
    }
  }

  public disconnect(): void {
    if (this.socket$) {
      this.socket$.complete();
      this.socket$ = null;
      this.isConnected$.next(false);
      this.connectionStatusSubject$.next('disconnected');
    }
  }

  ngOnDestroy(): void {
    this.disconnect();
    this.messagesSubject$.complete();
    this.isConnected$.complete();
    this.connectionStatusSubject$.complete();
  }

  // Typed message observers for specific message types
  public getScanProgress(): Observable<ScanProgressMessage> {
    return new Observable(observer => {
      const subscription = this.messages$.subscribe(message => {
        if (message.type === 'scan_progress') {
          observer.next(message.payload as ScanProgressMessage);
        }
      });
      return () => subscription.unsubscribe();
    });
  }

  public getUpdateComplete(): Observable<UpdateCompleteMessage> {
    return new Observable(observer => {
      const subscription = this.messages$.subscribe(message => {
        if (message.type === 'update_complete') {
          observer.next(message.payload as UpdateCompleteMessage);
        }
      });
      return () => subscription.unsubscribe();
    });
  }

  public getSecurityAlerts(): Observable<SecurityAlertMessage> {
    return new Observable(observer => {
      const subscription = this.messages$.subscribe(message => {
        if (message.type === 'security_alert') {
          observer.next(message.payload as SecurityAlertMessage);
        }
      });
      return () => subscription.unsubscribe();
    });
  }

  public getSystemStatus(): Observable<SystemStatusMessage> {
    return new Observable(observer => {
      const subscription = this.messages$.subscribe(message => {
        if (message.type === 'system_status') {
          observer.next(message.payload as SystemStatusMessage);
        }
      });
      return () => subscription.unsubscribe();
    });
  }

  public getNotifications(): Observable<any> {
    return new Observable(observer => {
      const subscription = this.messages$.subscribe(message => {
        if (message.type === 'notification') {
          observer.next(message.payload);
        }
      });
      return () => subscription.unsubscribe();
    });
  }
}
