import { Routes } from '@angular/router';

export const routes: Routes = [
  {
    path: '',
    redirectTo: '/dashboard',
    pathMatch: 'full'
  },
  {
    path: 'dashboard',
    loadComponent: () => import('./features/dashboard/dashboard.component').then(m => m.DashboardComponent),
    title: 'Dashboard - Superintelligence Dependency Manager'
  },
  {
    path: 'dependencies',
    loadComponent: () => import('./features/dependencies/dependency-table.component').then(m => m.DependencyTableComponent),
    title: 'Dependencies - Superintelligence Dependency Manager'
  },
  {
    path: 'superint-insights',
    loadComponent: () => import('./features/superint-insights/superint-insights-panel.component').then(m => m.AIInsightsPanelComponent),
    title: 'Superintelligence Insights - Superintelligence Dependency Manager'
  },
  {
    path: 'logging-test',
    loadComponent: () => import('./features/logging-test/logging-test.component').then(m => m.LoggingTestComponent),
    title: 'Logging Test - Superintelligence Dependency Manager'
  },
  {
    path: 'projects',
    loadChildren: () => import('./features/projects/projects.routes').then(m => m.projectRoutes),
    title: 'Projects - Superintelligence Dependency Manager'
  },
  {
    path: 'security',
    loadChildren: () => import('./features/security/security.routes').then(m => m.securityRoutes),
    title: 'Security - Superintelligence Dependency Manager'
  },
  {
    path: 'analytics',
    loadChildren: () => import('./features/analytics/analytics.routes').then(m => m.analyticsRoutes),
    title: 'Analytics - Superintelligence Dependency Manager'
  },
  {
    path: 'policies',
    loadChildren: () => import('./features/policies/policies.routes').then(m => m.policyRoutes),
    title: 'Policies - Superintelligence Dependency Manager'
  },
  {
    path: 'settings',
    loadChildren: () => import('./features/settings/settings.routes').then(m => m.settingsRoutes),
    title: 'Settings - Superintelligence Dependency Manager'
  },
  {
    path: '**',
    loadComponent: () => import('./shared/components/not-found/not-found.component').then(m => m.NotFoundComponent),
    title: 'Page Not Found - Superintelligence Dependency Manager'
  }
];
