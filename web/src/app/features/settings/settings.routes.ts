import { Routes } from '@angular/router';

export const settingsRoutes: Routes = [
  {
    path: '',
    loadComponent: () => import('./settings-dashboard/settings-dashboard.component').then(m => m.SettingsDashboardComponent),
    title: 'Settings - Superintelligence Dependency Manager'
  },
  {
    path: 'general',
    loadComponent: () => import('./general-settings/general-settings.component').then(m => m.GeneralSettingsComponent),
    title: 'General Settings - Superintelligence Dependency Manager'
  },
  {
    path: 'notifications',
    loadComponent: () => import('./notification-settings/notification-settings.component').then(m => m.NotificationSettingsComponent),
    title: 'Notification Settings - Superintelligence Dependency Manager'
  },
  {
    path: 'integrations',
    loadComponent: () => import('./integration-settings/integration-settings.component').then(m => m.IntegrationSettingsComponent),
    title: 'Integration Settings - Superintelligence Dependency Manager'
  },
  {
    path: 'superint-configuration',
    loadComponent: () => import('./superint-configuration/superint-configuration.component').then(m => m.SuperintConfigurationComponent),
    title: 'Superintelligence Configuration - Superintelligence Dependency Manager'
  },
  {
    path: 'user-management',
    loadComponent: () => import('./user-management/user-management.component').then(m => m.UserManagementComponent),
    title: 'User Management - Superintelligence Dependency Manager'
  },
  {
    path: 'backup-restore',
    loadComponent: () => import('./backup-restore/backup-restore.component').then(m => m.BackupRestoreComponent),
    title: 'Backup & Restore - Superintelligence Dependency Manager'
  },
  {
    path: 'system-logs',
    loadComponent: () => import('./system-logs/system-logs.component').then(m => m.SystemLogsComponent),
    title: 'System Logs - Superintelligence Dependency Manager'
  }
];
