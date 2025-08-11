import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';

@Component({
  selector: 'app-report-detail',
  standalone: true,
  imports: [CommonModule, MatCardModule],
  template: `
    <div class="report-detail">
      <mat-card>
        <mat-card-header>
          <mat-card-title>Report Detail</mat-card-title>
        </mat-card-header>
        <mat-card-content>
          <p>Report detail component - Coming soon!</p>
        </mat-card-content>
      </mat-card>
    </div>
  `,
  styles: [`
    .report-detail {
      padding: 24px;
    }
  `]
})
export class ReportDetailComponent {
}
