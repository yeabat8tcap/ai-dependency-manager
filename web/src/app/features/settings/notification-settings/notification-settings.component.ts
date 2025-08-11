import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';

@Component({
  selector: 'app-notification-settings',
  standalone: true,
  imports: [CommonModule, MatCardModule, MatButtonModule, MatIconModule],
  template: `
    <div class="notification-settings">
      <mat-card>
        <mat-card-header>
          <mat-card-title>Notification Settings</mat-card-title>
          <mat-card-subtitle>Configure alerts and notifications</mat-card-subtitle>
        </mat-card-header>
        <mat-card-content>
          <p>Notification settings component - Coming soon!</p>
        </mat-card-content>
      </mat-card>
    </div>
  `,
  styles: [`
    .notification-settings {
      padding: 24px;
    }
  `]
})
export class NotificationSettingsComponent {
}
