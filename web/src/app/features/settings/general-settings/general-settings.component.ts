import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';

@Component({
  selector: 'app-general-settings',
  standalone: true,
  imports: [CommonModule, MatCardModule, MatButtonModule, MatIconModule],
  template: `
    <div class="general-settings">
      <mat-card>
        <mat-card-header>
          <mat-card-title>General Settings</mat-card-title>
          <mat-card-subtitle>System configuration</mat-card-subtitle>
        </mat-card-header>
        <mat-card-content>
          <p>General settings component - Coming soon!</p>
        </mat-card-content>
      </mat-card>
    </div>
  `,
  styles: [`
    .general-settings {
      padding: 24px;
    }
  `]
})
export class GeneralSettingsComponent {
}
