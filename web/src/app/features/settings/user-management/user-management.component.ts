import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';

@Component({
  selector: 'app-user-management',
  standalone: true,
  imports: [CommonModule, MatCardModule],
  template: `
    <div class="user-management">
      <mat-card>
        <mat-card-header>
          <mat-card-title>User Management</mat-card-title>
        </mat-card-header>
        <mat-card-content>
          <p>User management component - Coming soon!</p>
        </mat-card-content>
      </mat-card>
    </div>
  `,
  styles: [`
    .user-management {
      padding: 24px;
    }
  `]
})
export class UserManagementComponent {
}
