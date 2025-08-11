import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';

@Component({
  selector: 'app-backup-restore',
  standalone: true,
  imports: [CommonModule, MatCardModule],
  template: `
    <div class="backup-restore">
      <mat-card>
        <mat-card-header>
          <mat-card-title>Backup & Restore</mat-card-title>
        </mat-card-header>
        <mat-card-content>
          <p>Backup & restore component - Coming soon!</p>
        </mat-card-content>
      </mat-card>
    </div>
  `,
  styles: [`
    .backup-restore {
      padding: 24px;
    }
  `]
})
export class BackupRestoreComponent {
}
