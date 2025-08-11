import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';

@Component({
  selector: 'app-policy-edit',
  standalone: true,
  imports: [CommonModule, MatCardModule],
  template: `
    <div class="component-container">
      <mat-card>
        <mat-card-header>
          <mat-card-title>Edit Policy</mat-card-title>
        </mat-card-header>
        <mat-card-content>
          <p>Edit Policy component - Coming soon!</p>
        </mat-card-content>
      </mat-card>
    </div>
  `,
  styles: [`
    .component-container {
      padding: 24px;
    }
  `]
})
export class PolicyEditComponent {
}
