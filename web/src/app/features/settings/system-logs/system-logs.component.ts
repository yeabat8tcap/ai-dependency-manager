import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';

@Component({
  selector: 'app-system-logs',
  standalone: true,
  imports: [CommonModule, MatCardModule],
  template: `
    <div class="system-logs">
      <mat-card>
        <mat-card-header>
          <mat-card-title>System Logs</mat-card-title>
        </mat-card-header>
        <mat-card-content>
          <p>System logs component - Coming soon!</p>
        </mat-card-content>
      </mat-card>
    </div>
  `,
  styles: [`
    .system-logs {
      padding: 24px;
    }
  `]
})
export class SystemLogsComponent {
}
