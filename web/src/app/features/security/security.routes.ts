import { Routes } from '@angular/router';

export const securityRoutes: Routes = [
  {
    path: '',
    loadComponent: () => import('./security-dashboard/security-dashboard.component').then(m => m.SecurityDashboardComponent),
    title: 'Security Dashboard - AI Dependency Manager'
  },
  {
    path: 'vulnerabilities',
    loadComponent: () => import('./vulnerability-list/vulnerability-list.component').then(m => m.VulnerabilityListComponent),
    title: 'Vulnerabilities - AI Dependency Manager'
  },
  {
    path: 'vulnerabilities/:id',
    loadComponent: () => import('./vulnerability-detail/vulnerability-detail.component').then(m => m.VulnerabilityDetailComponent),
    title: 'Vulnerability Details - AI Dependency Manager'
  },
  {
    path: 'policies',
    loadComponent: () => import('./security-policies/security-policies.component').then(m => m.SecurityPoliciesComponent),
    title: 'Security Policies - AI Dependency Manager'
  },
  {
    path: 'compliance',
    loadComponent: () => import('./compliance-report/compliance-report.component').then(m => m.ComplianceReportComponent),
    title: 'Compliance Report - AI Dependency Manager'
  },
  {
    path: 'alerts',
    loadComponent: () => import('./security-alerts/security-alerts.component').then(m => m.SecurityAlertsComponent),
    title: 'Security Alerts - AI Dependency Manager'
  }
];
