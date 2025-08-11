import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';

@Component({
  selector: 'app-report-builder',
  standalone: true,
  imports: [CommonModule, MatCardModule],
  template: `
    <div class="report-builder">
      <mat-card>
        <mat-card-header>
          <mat-card-title>Report Builder</mat-card-title>
        </mat-card-header>
        <mat-card-content>
          <p>Report builder component - Coming soon!</p>
        </mat-card-content>
      </mat-card>
    </div>
  `,
  styles: [`
    .report-builder {
      padding: 24px;
    }
  `]
})
export class ReportBuilderComponent {
}
