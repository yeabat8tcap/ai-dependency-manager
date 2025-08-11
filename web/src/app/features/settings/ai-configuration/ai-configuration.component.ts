import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';

@Component({
  selector: 'app-ai-configuration',
  standalone: true,
  imports: [CommonModule, MatCardModule],
  template: `
    <div class="ai-configuration">
      <mat-card>
        <mat-card-header>
          <mat-card-title>AI Configuration</mat-card-title>
        </mat-card-header>
        <mat-card-content>
          <p>AI configuration component - Coming soon!</p>
        </mat-card-content>
      </mat-card>
    </div>
  `,
  styles: [`
    .ai-configuration {
      padding: 24px;
    }
  `]
})
export class AiConfigurationComponent {
}
