import { Routes } from '@angular/router';

export const policyRoutes: Routes = [
  {
    path: '',
    loadComponent: () => import('./policy-list/policy-list.component').then(m => m.PolicyListComponent),
    title: 'Policies - AI Dependency Manager'
  },
  {
    path: 'create',
    loadComponent: () => import('./policy-create/policy-create.component').then(m => m.PolicyCreateComponent),
    title: 'Create Policy - AI Dependency Manager'
  },
  {
    path: 'templates',
    loadComponent: () => import('./policy-templates/policy-templates.component').then(m => m.PolicyTemplatesComponent),
    title: 'Policy Templates - AI Dependency Manager'
  },
  {
    path: ':id',
    loadComponent: () => import('./policy-detail/policy-detail.component').then(m => m.PolicyDetailComponent),
    title: 'Policy Details - AI Dependency Manager'
  },
  {
    path: ':id/edit',
    loadComponent: () => import('./policy-edit/policy-edit.component').then(m => m.PolicyEditComponent),
    title: 'Edit Policy - AI Dependency Manager'
  },
  {
    path: ':id/test',
    loadComponent: () => import('./policy-test/policy-test.component').then(m => m.PolicyTestComponent),
    title: 'Test Policy - AI Dependency Manager'
  },
  {
    path: 'compliance',
    loadComponent: () => import('./policy-compliance/policy-compliance.component').then(m => m.PolicyComplianceComponent),
    title: 'Policy Compliance - AI Dependency Manager'
  }
];
