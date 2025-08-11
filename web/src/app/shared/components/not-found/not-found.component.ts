import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { RouterModule } from '@angular/router';

@Component({
  selector: 'app-not-found',
  standalone: true,
  imports: [
    CommonModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    RouterModule
  ],
  template: `
    <div class="not-found-container">
      <mat-card class="not-found-card">
        <mat-card-content>
          <div class="not-found-content">
            <mat-icon class="not-found-icon">search_off</mat-icon>
            <h1 class="serif-display">404</h1>
            <h2 class="serif-title">Page Not Found</h2>
            <p class="sans-body">
              The page you're looking for doesn't exist or has been moved.
            </p>
            <div class="not-found-actions">
              <button mat-raised-button color="primary" routerLink="/dashboard">
                <mat-icon>dashboard</mat-icon>
                Go to Dashboard
              </button>
              <button mat-button routerLink="/dependencies">
                <mat-icon>account_tree</mat-icon>
                View Dependencies
              </button>
            </div>
          </div>
        </mat-card-content>
      </mat-card>
    </div>
  `,
  styles: [`
    .not-found-container {
      display: flex;
      justify-content: center;
      align-items: center;
      min-height: 100vh;
      padding: 24px;
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    }

    .not-found-card {
      max-width: 500px;
      width: 100%;
      text-align: center;
    }

    .not-found-content {
      padding: 48px 24px;
    }

    .not-found-icon {
      font-size: 6rem;
      width: 6rem;
      height: 6rem;
      color: #9e9e9e;
      margin-bottom: 24px;
    }

    h1 {
      font-size: 4rem;
      margin: 0 0 16px 0;
      color: #4a148c;
    }

    h2 {
      margin: 0 0 16px 0;
      color: #666;
    }

    p {
      margin: 0 0 32px 0;
      color: #757575;
    }

    .not-found-actions {
      display: flex;
      gap: 16px;
      justify-content: center;
      flex-wrap: wrap;
    }

    .not-found-actions button {
      display: flex;
      align-items: center;
      gap: 8px;
    }

    @media (max-width: 768px) {
      .not-found-container {
        padding: 16px;
      }

      .not-found-content {
        padding: 32px 16px;
      }

      h1 {
        font-size: 3rem;
      }

      .not-found-actions {
        flex-direction: column;
      }

      .not-found-actions button {
        width: 100%;
        justify-content: center;
      }
    }
  `]
})
export class NotFoundComponent {}
