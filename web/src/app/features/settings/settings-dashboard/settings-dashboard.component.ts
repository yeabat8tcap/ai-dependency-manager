import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';

@Component({
  selector: 'app-settings-dashboard',
  standalone: true,
  imports: [CommonModule, MatCardModule, MatButtonModule, MatIconModule],
  template: `
    <div class="settings-dashboard">
      <mat-card>
        <mat-card-header>
          <mat-card-title>Settings Dashboard</mat-card-title>
          <mat-card-subtitle>System configuration overview</mat-card-subtitle>
        </mat-card-header>
        <mat-card-content>
          <p>Settings dashboard component - Coming soon!</p>
        </mat-card-content>
      </mat-card>
    </div>
  `,
  styles: [`
    .settings-dashboard {
      padding: 24px;
    }
  `]
})
export class SettingsDashboardComponent {
}
