import { Component, OnInit, OnDestroy, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterOutlet, RouterModule, Router } from '@angular/router';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatSidenavModule } from '@angular/material/sidenav';
import { MatListModule } from '@angular/material/list';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatBadgeModule } from '@angular/material/badge';
import { BreakpointObserver, Breakpoints } from '@angular/cdk/layout';
import { Subject, takeUntil } from 'rxjs';

import { WebSocketService } from './core/services/websocket.service';

interface NavigationItem {
  label: string;
  icon: string;
  route: string;
  badge?: number;
  tooltip?: string;
}

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [
    CommonModule,
    RouterOutlet,
    RouterModule,
    MatToolbarModule,
    MatSidenavModule,
    MatListModule,
    MatIconModule,
    MatButtonModule,
    MatTooltipModule,
    MatBadgeModule
  ],
  template: `
    <mat-sidenav-container class="sidenav-container">
      <!-- Side Navigation -->
      <mat-sidenav 
        #drawer 
        class="sidenav" 
        fixedInViewport
        [attr.role]="isHandset ? 'dialog' : 'navigation'"
        [mode]="isHandset ? 'over' : 'side'"
        [opened]="!isHandset">
        
        <!-- App Logo and Title -->
        <div class="sidenav-header">
          <div class="app-logo">
            <mat-icon class="logo-icon">psychology</mat-icon>
            <div class="app-title">
              <h1 class="serif-title">AI Dependency</h1>
              <h2 class="serif-title">Manager</h2>
            </div>
          </div>
        </div>

        <!-- Navigation Menu -->
        <mat-nav-list class="nav-list">
          <a 
            mat-list-item 
            *ngFor="let item of navigationItems" 
            [routerLink]="item.route"
            routerLinkActive="active-nav-item"
            [matTooltip]="item.tooltip || item.label"
            matTooltipPosition="right"
            class="nav-item">
            <mat-icon 
              matListItemIcon 
              [matBadge]="item.badge" 
              [matBadgeHidden]="!item.badge"
              matBadgeColor="accent"
              matBadgeSize="small">
              {{ item.icon }}
            </mat-icon>
            <span matListItemTitle class="nav-label sans-medium">{{ item.label }}</span>
          </a>
        </mat-nav-list>

        <!-- System Status -->
        <div class="sidenav-footer">
          <div class="system-status" [class]="'status-' + systemStatus">
            <mat-icon class="status-icon">{{ getStatusIcon() }}</mat-icon>
            <span class="status-text sans-body">{{ getStatusText() }}</span>
          </div>
        </div>
      </mat-sidenav>

      <!-- Main Content Area -->
      <mat-sidenav-content class="main-content">
        <!-- Top Toolbar -->
        <mat-toolbar class="toolbar" color="primary">
          <button
            type="button"
            aria-label="Toggle sidenav"
            mat-icon-button
            (click)="drawer.toggle()"
            *ngIf="isHandset">
            <mat-icon aria-label="Side nav toggle icon">menu</mat-icon>
          </button>
          
          <span class="toolbar-spacer"></span>
          
          <!-- Real-time Connection Status -->
          <div class="connection-status" [class]="'connection-' + connectionStatus">
            <mat-icon class="connection-icon">{{ getConnectionIcon() }}</mat-icon>
            <span class="connection-text sans-body" *ngIf="!isHandset">
              {{ getConnectionText() }}
            </span>
          </div>

          <!-- Notification Bell -->
          <button 
            mat-icon-button 
            class="notification-button"
            [matBadge]="notificationCount" 
            [matBadgeHidden]="notificationCount === 0"
            matBadgeColor="warn"
            matTooltip="Notifications">
            <mat-icon>notifications</mat-icon>
          </button>

          <!-- User Menu -->
          <button mat-icon-button matTooltip="User Settings">
            <mat-icon>account_circle</mat-icon>
          </button>
        </mat-toolbar>

        <!-- Router Outlet for Page Content -->
        <div class="page-content">
          <router-outlet></router-outlet>
        </div>
      </mat-sidenav-content>
    </mat-sidenav-container>
  `,
  styles: [`
    .sidenav-container {
      height: 100vh;
    }

    .sidenav {
      width: 280px;
      background-color: #ffffff;
      border-right: 1px solid rgba(0, 0, 0, 0.12);
    }

    .sidenav-header {
      padding: 24px 16px;
      border-bottom: 1px solid rgba(0, 0, 0, 0.12);
      background: linear-gradient(135deg, #4a148c 0%, #7b1fa2 100%);
      color: white;
    }

    .app-logo {
      display: flex;
      align-items: center;
      gap: 16px;
    }

    .logo-icon {
      font-size: 2.5rem;
      width: 2.5rem;
      height: 2.5rem;
      color: #ffc107;
    }

    .app-title h1,
    .app-title h2 {
      margin: 0;
      line-height: 1.2;
      color: white;
    }

    .app-title h1 {
      font-size: 1.25rem;
      font-weight: 600;
    }

    .app-title h2 {
      font-size: 1rem;
      font-weight: 400;
      opacity: 0.9;
    }

    .nav-list {
      padding-top: 16px;
    }

    .nav-item {
      margin: 4px 12px;
      border-radius: 8px;
      transition: all 0.2s ease;
    }

    .nav-item:hover {
      background-color: rgba(74, 20, 140, 0.08);
    }

    .active-nav-item {
      background-color: rgba(74, 20, 140, 0.12) !important;
      color: #4a148c !important;
    }

    .active-nav-item .nav-label {
      font-weight: 600 !important;
    }

    .nav-label {
      font-size: 0.9rem;
    }

    .sidenav-footer {
      position: absolute;
      bottom: 0;
      left: 0;
      right: 0;
      padding: 16px;
      border-top: 1px solid rgba(0, 0, 0, 0.12);
      background-color: #fafafa;
    }

    .system-status {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 8px 12px;
      border-radius: 6px;
      font-size: 0.8rem;
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

    .status-icon {
      font-size: 1rem;
      width: 1rem;
      height: 1rem;
    }

    .toolbar {
      position: sticky;
      top: 0;
      z-index: 1000;
      background: linear-gradient(90deg, #4a148c 0%, #7b1fa2 100%);
      color: white;
      box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    }

    .toolbar-spacer {
      flex: 1 1 auto;
    }

    .connection-status {
      display: flex;
      align-items: center;
      gap: 6px;
      padding: 4px 8px;
      border-radius: 4px;
      margin-right: 16px;
      font-size: 0.8rem;
    }

    .connection-connected {
      background-color: rgba(76, 175, 80, 0.2);
      color: #4caf50;
    }

    .connection-connecting {
      background-color: rgba(255, 193, 7, 0.2);
      color: #ffc107;
    }

    .connection-disconnected {
      background-color: rgba(244, 67, 54, 0.2);
      color: #f44336;
    }

    .connection-icon {
      font-size: 1rem;
      width: 1rem;
      height: 1rem;
    }

    .notification-button {
      margin-right: 8px;
    }

    .page-content {
      height: calc(100vh - 64px);
      overflow-y: auto;
      background-color: #fafafa;
    }

    /* Mobile Styles */
    @media (max-width: 768px) {
      .sidenav {
        width: 100%;
        max-width: 320px;
      }

      .sidenav-header {
        padding: 16px;
      }

      .app-title h1 {
        font-size: 1.1rem;
      }

      .app-title h2 {
        font-size: 0.9rem;
      }

      .logo-icon {
        font-size: 2rem;
        width: 2rem;
        height: 2rem;
      }

      .connection-text {
        display: none;
      }

      .page-content {
        height: calc(100vh - 56px);
      }
    }

    /* Dark mode support (future enhancement) */
    @media (prefers-color-scheme: dark) {
      .sidenav {
        background-color: #303030;
        color: white;
      }

      .sidenav-footer {
        background-color: #424242;
        border-top-color: rgba(255, 255, 255, 0.12);
      }

      .page-content {
        background-color: #121212;
      }
    }
  `]
})
export class AppComponent implements OnInit, OnDestroy {
  private destroy$ = new Subject<void>();
  private breakpointObserver = inject(BreakpointObserver);
  private wsService = inject(WebSocketService);
  private router = inject(Router);

  title = 'AI Dependency Manager';
  isHandset = false;
  systemStatus: 'healthy' | 'warning' | 'error' = 'healthy';
  connectionStatus: 'connecting' | 'connected' | 'disconnected' | 'error' = 'connecting';
  notificationCount = 0;

  navigationItems: NavigationItem[] = [
    {
      label: 'Dashboard',
      icon: 'dashboard',
      route: '/dashboard',
      tooltip: 'Main dashboard with project overview'
    },
    {
      label: 'Dependencies',
      icon: 'account_tree',
      route: '/dependencies',
      tooltip: 'View and manage all dependencies'
    },
    {
      label: 'AI Insights',
      icon: 'psychology',
      route: '/ai-insights',
      tooltip: 'AI-powered analysis and recommendations'
    },
    {
      label: 'Projects',
      icon: 'folder',
      route: '/projects',
      tooltip: 'Manage your projects'
    },
    {
      label: 'Security',
      icon: 'security',
      route: '/security',
      badge: 0,
      tooltip: 'Security vulnerabilities and compliance'
    },
    {
      label: 'Analytics',
      icon: 'analytics',
      route: '/analytics',
      tooltip: 'Detailed analytics and reporting'
    },
    {
      label: 'Policies',
      icon: 'policy',
      route: '/policies',
      tooltip: 'Update policies and rules'
    },
    {
      label: 'Settings',
      icon: 'settings',
      route: '/settings',
      tooltip: 'Application settings and configuration'
    }
  ];

  ngOnInit(): void {
    this.setupResponsiveLayout();
    this.subscribeToWebSocketUpdates();
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  private setupResponsiveLayout(): void {
    this.breakpointObserver.observe(Breakpoints.Handset)
      .pipe(takeUntil(this.destroy$))
      .subscribe(result => {
        this.isHandset = result.matches;
      });
  }

  private subscribeToWebSocketUpdates(): void {
    // Connection status
    this.wsService.connectionStatus$
      .pipe(takeUntil(this.destroy$))
      .subscribe(status => {
        this.connectionStatus = status;
      });

    // System status updates
    this.wsService.getSystemStatus()
      .pipe(takeUntil(this.destroy$))
      .subscribe(status => {
        this.systemStatus = status.status;
      });

    // Security alerts for badge updates
    this.wsService.getSecurityAlerts()
      .pipe(takeUntil(this.destroy$))
      .subscribe(() => {
        const securityItem = this.navigationItems.find(item => item.route === '/security');
        if (securityItem) {
          securityItem.badge = (securityItem.badge || 0) + 1;
        }
      });

    // General notifications
    this.wsService.getNotifications()
      .pipe(takeUntil(this.destroy$))
      .subscribe(() => {
        this.notificationCount++;
      });
  }

  getStatusIcon(): string {
    switch (this.systemStatus) {
      case 'healthy': return 'check_circle';
      case 'warning': return 'warning';
      case 'error': return 'error';
      default: return 'help';
    }
  }

  getStatusText(): string {
    switch (this.systemStatus) {
      case 'healthy': return 'System Healthy';
      case 'warning': return 'System Warning';
      case 'error': return 'System Error';
      default: return 'Unknown Status';
    }
  }

  getConnectionIcon(): string {
    switch (this.connectionStatus) {
      case 'connected': return 'wifi';
      case 'connecting': return 'wifi_tethering';
      case 'disconnected': return 'wifi_off';
      default: return 'help';
    }
  }

  getConnectionText(): string {
    switch (this.connectionStatus) {
      case 'connected': return 'Connected';
      case 'connecting': return 'Connecting...';
      case 'disconnected': return 'Disconnected';
      default: return 'Unknown';
    }
  }
}
