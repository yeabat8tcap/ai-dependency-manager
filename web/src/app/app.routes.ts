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
    title: 'Dashboard - AI Dependency Manager'
  },
  {
    path: 'dependencies',
    loadComponent: () => import('./features/dependencies/dependency-table.component').then(m => m.DependencyTableComponent),
    title: 'Dependencies - AI Dependency Manager'
  },
  {
    path: 'ai-insights',
    loadComponent: () => import('./features/ai-insights/ai-insights-panel.component').then(m => m.AIInsightsPanelComponent),
    title: 'AI Insights - AI Dependency Manager'
  },
  {
    path: 'logging-test',
    loadComponent: () => import('./features/logging-test/logging-test.component').then(m => m.LoggingTestComponent),
    title: 'Logging Test - AI Dependency Manager'
  },
  {
    path: 'projects',
    loadChildren: () => import('./features/projects/projects.routes').then(m => m.projectRoutes),
    title: 'Projects - AI Dependency Manager'
  },
  {
    path: 'security',
    loadChildren: () => import('./features/security/security.routes').then(m => m.securityRoutes),
    title: 'Security - AI Dependency Manager'
  },
  {
    path: 'analytics',
    loadChildren: () => import('./features/analytics/analytics.routes').then(m => m.analyticsRoutes),
    title: 'Analytics - AI Dependency Manager'
  },
  {
    path: 'policies',
    loadChildren: () => import('./features/policies/policies.routes').then(m => m.policyRoutes),
    title: 'Policies - AI Dependency Manager'
  },
  {
    path: 'settings',
    loadChildren: () => import('./features/settings/settings.routes').then(m => m.settingsRoutes),
    title: 'Settings - AI Dependency Manager'
  },
  {
    path: '**',
    loadComponent: () => import('./shared/components/not-found/not-found.component').then(m => m.NotFoundComponent),
    title: 'Page Not Found - AI Dependency Manager'
  }
];
