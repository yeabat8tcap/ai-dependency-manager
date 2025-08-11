import { Routes } from '@angular/router';

export const analyticsRoutes: Routes = [
  {
    path: '',
    loadComponent: () => import('./analytics-dashboard/analytics-dashboard.component').then(m => m.AnalyticsDashboardComponent),
    title: 'Analytics Dashboard - AI Dependency Manager'
  },
  {
    path: 'dependency-lag',
    loadComponent: () => import('./dependency-lag/dependency-lag.component').then(m => m.DependencyLagComponent),
    title: 'Dependency Lag Analysis - AI Dependency Manager'
  },
  {
    path: 'update-trends',
    loadComponent: () => import('./update-trends/update-trends.component').then(m => m.UpdateTrendsComponent),
    title: 'Update Trends - AI Dependency Manager'
  },
  {
    path: 'performance-metrics',
    loadComponent: () => import('./performance-metrics/performance-metrics.component').then(m => m.PerformanceMetricsComponent),
    title: 'Performance Metrics - AI Dependency Manager'
  },
  {
    path: 'reports',
    loadComponent: () => import('./reports/reports.component').then(m => m.ReportsComponent),
    title: 'Reports - AI Dependency Manager'
  },
  {
    path: 'reports/create',
    loadComponent: () => import('./report-builder/report-builder.component').then(m => m.ReportBuilderComponent),
    title: 'Create Report - AI Dependency Manager'
  },
  {
    path: 'reports/:id',
    loadComponent: () => import('./report-detail/report-detail.component').then(m => m.ReportDetailComponent),
    title: 'Report Details - AI Dependency Manager'
  }
];
