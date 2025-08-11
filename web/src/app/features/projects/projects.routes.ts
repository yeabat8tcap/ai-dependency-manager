import { Routes } from '@angular/router';

export const projectRoutes: Routes = [
  {
    path: '',
    loadComponent: () => import('./project-list/project-list.component').then(m => m.ProjectListComponent),
    title: 'Projects - AI Dependency Manager'
  },
  {
    path: 'create',
    loadComponent: () => import('./project-create/project-create.component').then(m => m.ProjectCreateComponent),
    title: 'Create Project - AI Dependency Manager'
  },
  {
    path: ':id',
    loadComponent: () => import('./project-detail/project-detail.component').then(m => m.ProjectDetailComponent),
    title: 'Project Details - AI Dependency Manager'
  },
  {
    path: ':id/edit',
    loadComponent: () => import('./project-edit/project-edit.component').then(m => m.ProjectEditComponent),
    title: 'Edit Project - AI Dependency Manager'
  },
  {
    path: ':id/dependencies',
    loadComponent: () => import('./project-dependencies/project-dependencies.component').then(m => m.ProjectDependenciesComponent),
    title: 'Project Dependencies - AI Dependency Manager'
  },
  {
    path: ':id/settings',
    loadComponent: () => import('./project-settings/project-settings.component').then(m => m.ProjectSettingsComponent),
    title: 'Project Settings - AI Dependency Manager'
  }
];
