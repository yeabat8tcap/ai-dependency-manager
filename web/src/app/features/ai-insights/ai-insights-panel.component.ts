import { Component, OnInit, OnDestroy, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatChipsModule } from '@angular/material/chips';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatDividerModule } from '@angular/material/divider';
import { Subject, takeUntil } from 'rxjs';

import { ApiService } from '../../core/services/api.service';
import { WebSocketService } from '../../core/services/websocket.service';

export interface AIInsight {
  id: string;
  type: 'breaking_change' | 'security_risk' | 'performance_impact' | 'compatibility' | 'recommendation';
  title: string;
  description: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  confidence: number;
  affectedDependencies: string[];
  recommendation: string;
  impact: string;
  timestamp: Date;
  metadata?: any;
}

export interface BreakingChangeAnalysis {
  dependencyName: string;
  fromVersion: string;
  toVersion: string;
  breakingChanges: {
    type: 'api_removal' | 'api_change' | 'behavior_change' | 'dependency_change';
    description: string;
    impact: 'low' | 'medium' | 'high';
    mitigation?: string;
  }[];
  migrationGuide?: string;
  estimatedEffort: 'low' | 'medium' | 'high';
}

export interface SecurityRiskAnalysis {
  cve: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  description: string;
  affectedVersions: string[];
  fixedInVersion?: string;
  workaround?: string;
  exploitability: number;
  impact: number;
}

@Component({
  selector: 'app-ai-insights-panel',
  standalone: true,
  imports: [
    CommonModule,
    MatCardModule,
    MatExpansionModule,
    MatIconModule,
    MatButtonModule,
    MatChipsModule,
    MatProgressBarModule,
    MatTooltipModule,
    MatDividerModule
  ],
  template: `
    <div class="ai-insights-container">
      <mat-card class="insights-header-card">
        <mat-card-header>
          <mat-card-title>
            <mat-icon class="ai-icon">psychology</mat-icon>
            AI-Powered Insights
          </mat-card-title>
          <mat-card-subtitle>
            Intelligent analysis and recommendations for your dependencies
          </mat-card-subtitle>
        </mat-card-header>
        <mat-card-content>
          <div class="insights-summary">
            <div class="summary-item">
              <div class="summary-number">{{ insights.length }}</div>
              <div class="summary-label">Total Insights</div>
            </div>
            <div class="summary-item">
              <div class="summary-number">{{ getCriticalInsightsCount() }}</div>
              <div class="summary-label">Critical Issues</div>
            </div>
            <div class="summary-item">
              <div class="summary-number">{{ getAverageConfidence() }}%</div>
              <div class="summary-label">Avg Confidence</div>
            </div>
          </div>
        </mat-card-content>
      </mat-card>

      <!-- Insights List -->
      <mat-accordion class="insights-accordion" *ngIf="insights.length > 0">
        <mat-expansion-panel 
          *ngFor="let insight of insights; trackBy: trackByInsightId"
          [class]="'insight-panel severity-' + insight.severity">
          
          <mat-expansion-panel-header>
            <mat-panel-title>
              <div class="insight-header">
                <mat-icon class="insight-type-icon" [class]="'type-' + insight.type">
                  {{ getInsightIcon(insight.type) }}
                </mat-icon>
                <div class="insight-title-info">
                  <div class="insight-title">{{ insight.title }}</div>
                  <div class="insight-meta">
                    <mat-chip class="severity-chip" [class]="'severity-' + insight.severity">
                      {{ insight.severity | titlecase }}
                    </mat-chip>
                    <span class="confidence-score">
                      Confidence: {{ insight.confidence }}%
                    </span>
                  </div>
                </div>
              </div>
            </mat-panel-title>
            <mat-panel-description>
              <div class="affected-count">
                {{ insight.affectedDependencies.length }} dependencies affected
              </div>
            </mat-panel-description>
          </mat-expansion-panel-header>

          <div class="insight-content">
            <!-- Description -->
            <div class="insight-section">
              <h4>
                <mat-icon>description</mat-icon>
                Analysis
              </h4>
              <p>{{ insight.description }}</p>
            </div>

            <!-- Affected Dependencies -->
            <div class="insight-section" *ngIf="insight.affectedDependencies.length > 0">
              <h4>
                <mat-icon>account_tree</mat-icon>
                Affected Dependencies
              </h4>
              <mat-chip-set>
                <mat-chip *ngFor="let dep of insight.affectedDependencies">
                  {{ dep }}
                </mat-chip>
              </mat-chip-set>
            </div>

            <!-- Impact Analysis -->
            <div class="insight-section">
              <h4>
                <mat-icon>assessment</mat-icon>
                Impact Assessment
              </h4>
              <p>{{ insight.impact }}</p>
            </div>

            <!-- Recommendation -->
            <div class="insight-section recommendation-section">
              <h4>
                <mat-icon>lightbulb</mat-icon>
                AI Recommendation
              </h4>
              <p>{{ insight.recommendation }}</p>
            </div>

            <!-- Breaking Change Details -->
            <div class="insight-section" *ngIf="insight.type === 'breaking_change' && insight.metadata?.breakingChanges">
              <h4>
                <mat-icon>warning</mat-icon>
                Breaking Changes Details
              </h4>
              <div class="breaking-changes-list">
                <div class="breaking-change-item" 
                     *ngFor="let change of insight.metadata.breakingChanges">
                  <div class="change-header">
                    <mat-chip class="change-type-chip">{{ change.type | titlecase }}</mat-chip>
                    <mat-chip class="change-impact-chip" [class]="'impact-' + change.impact">
                      {{ change.impact | titlecase }} Impact
                    </mat-chip>
                  </div>
                  <p class="change-description">{{ change.description }}</p>
                  <div class="change-mitigation" *ngIf="change.mitigation">
                    <strong>Mitigation:</strong> {{ change.mitigation }}
                  </div>
                </div>
              </div>
            </div>

            <!-- Security Risk Details -->
            <div class="insight-section" *ngIf="insight.type === 'security_risk' && insight.metadata?.cve">
              <h4>
                <mat-icon>security</mat-icon>
                Security Vulnerability Details
              </h4>
              <div class="security-details">
                <div class="security-item">
                  <strong>CVE ID:</strong> {{ insight.metadata.cve }}
                </div>
                <div class="security-item">
                  <strong>Affected Versions:</strong> {{ insight.metadata.affectedVersions?.join(', ') }}
                </div>
                <div class="security-item" *ngIf="insight.metadata.fixedInVersion">
                  <strong>Fixed in Version:</strong> {{ insight.metadata.fixedInVersion }}
                </div>
                <div class="security-scores">
                  <div class="score-item">
                    <span class="score-label">Exploitability:</span>
                    <mat-progress-bar mode="determinate" 
                                    [value]="insight.metadata.exploitability * 10"
                                    class="score-bar exploitability-bar">
                    </mat-progress-bar>
                    <span class="score-value">{{ insight.metadata.exploitability }}/10</span>
                  </div>
                  <div class="score-item">
                    <span class="score-label">Impact:</span>
                    <mat-progress-bar mode="determinate" 
                                    [value]="insight.metadata.impact * 10"
                                    class="score-bar impact-bar">
                    </mat-progress-bar>
                    <span class="score-value">{{ insight.metadata.impact }}/10</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- Actions -->
            <div class="insight-actions">
              <button mat-raised-button color="primary" (click)="applyRecommendation(insight)">
                <mat-icon>auto_fix_high</mat-icon>
                Apply Recommendation
              </button>
              <button mat-button (click)="viewDetailedAnalysis(insight)">
                <mat-icon>analytics</mat-icon>
                Detailed Analysis
              </button>
              <button mat-button (click)="dismissInsight(insight)">
                <mat-icon>close</mat-icon>
                Dismiss
              </button>
            </div>

            <mat-divider></mat-divider>
            <div class="insight-timestamp">
              <mat-icon>schedule</mat-icon>
              Generated {{ insight.timestamp | date:'medium' }}
            </div>
          </div>
        </mat-expansion-panel>
      </mat-accordion>

      <!-- No Insights Message -->
      <mat-card class="no-insights-card" *ngIf="insights.length === 0 && !loading">
        <mat-card-content>
          <div class="no-insights">
            <mat-icon class="no-insights-icon">psychology</mat-icon>
            <h3>No AI Insights Available</h3>
            <p>Run a dependency scan to generate AI-powered insights and recommendations.</p>
            <button mat-raised-button color="primary" (click)="triggerAnalysis()">
              <mat-icon>refresh</mat-icon>
              Generate Insights
            </button>
          </div>
        </mat-card-content>
      </mat-card>

      <!-- Loading State -->
      <mat-card class="loading-card" *ngIf="loading">
        <mat-card-content>
          <div class="loading-state">
            <mat-icon class="loading-icon">psychology</mat-icon>
            <h3>AI Analysis in Progress</h3>
            <p>Analyzing dependencies and generating insights...</p>
            <mat-progress-bar mode="indeterminate"></mat-progress-bar>
          </div>
        </mat-card-content>
      </mat-card>
    </div>
  `,
  styles: [`
    .ai-insights-container {
      padding: 24px;
      max-width: 1000px;
      margin: 0 auto;
    }

    .insights-header-card {
      margin-bottom: 24px;
    }

    .ai-icon {
      color: #9c27b0;
    }

    .insights-summary {
      display: flex;
      gap: 32px;
      margin-top: 16px;
    }

    .summary-item {
      text-align: center;
    }

    .summary-number {
      font-size: 2rem;
      font-weight: 300;
      color: #9c27b0;
    }

    .summary-label {
      color: rgba(0, 0, 0, 0.6);
      font-size: 0.9rem;
    }

    .insights-accordion {
      margin-bottom: 24px;
    }

    .insight-panel {
      margin-bottom: 16px;
      border-radius: 8px !important;
    }

    .severity-low {
      border-left: 4px solid #4caf50;
    }

    .severity-medium {
      border-left: 4px solid #ff9800;
    }

    .severity-high {
      border-left: 4px solid #f44336;
    }

    .severity-critical {
      border-left: 4px solid #b71c1c;
    }

    .insight-header {
      display: flex;
      align-items: center;
      gap: 12px;
      width: 100%;
    }

    .insight-type-icon {
      font-size: 1.5rem;
      width: 1.5rem;
      height: 1.5rem;
    }

    .type-breaking_change { color: #f44336; }
    .type-security_risk { color: #d32f2f; }
    .type-performance_impact { color: #ff9800; }
    .type-compatibility { color: #2196f3; }
    .type-recommendation { color: #4caf50; }

    .insight-title-info {
      flex: 1;
    }

    .insight-title {
      font-weight: 500;
      margin-bottom: 4px;
    }

    .insight-meta {
      display: flex;
      align-items: center;
      gap: 12px;
    }

    .severity-chip {
      font-size: 0.7rem;
      height: 20px;
    }

    .severity-low { background-color: #e8f5e8; color: #2e7d32; }
    .severity-medium { background-color: #fff3e0; color: #f57c00; }
    .severity-high { background-color: #ffebee; color: #d32f2f; }
    .severity-critical { background-color: #ffebee; color: #b71c1c; }

    .confidence-score {
      font-size: 0.8rem;
      color: rgba(0, 0, 0, 0.6);
    }

    .affected-count {
      font-size: 0.8rem;
      color: rgba(0, 0, 0, 0.6);
    }

    .insight-content {
      padding: 16px 0;
    }

    .insight-section {
      margin-bottom: 24px;
    }

    .insight-section h4 {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-bottom: 12px;
      color: rgba(0, 0, 0, 0.8);
    }

    .recommendation-section {
      background-color: #f3e5f5;
      padding: 16px;
      border-radius: 8px;
      border-left: 4px solid #9c27b0;
    }

    .breaking-changes-list {
      display: flex;
      flex-direction: column;
      gap: 16px;
    }

    .breaking-change-item {
      padding: 16px;
      border: 1px solid rgba(0, 0, 0, 0.12);
      border-radius: 8px;
      background-color: #fafafa;
    }

    .change-header {
      display: flex;
      gap: 8px;
      margin-bottom: 8px;
    }

    .change-type-chip {
      background-color: #e3f2fd;
      color: #1976d2;
    }

    .change-impact-chip {
      font-size: 0.7rem;
      height: 20px;
    }

    .impact-low { background-color: #e8f5e8; color: #2e7d32; }
    .impact-medium { background-color: #fff3e0; color: #f57c00; }
    .impact-high { background-color: #ffebee; color: #d32f2f; }

    .change-description {
      margin-bottom: 8px;
    }

    .change-mitigation {
      font-size: 0.9rem;
      color: rgba(0, 0, 0, 0.7);
    }

    .security-details {
      display: flex;
      flex-direction: column;
      gap: 12px;
    }

    .security-item {
      display: flex;
      gap: 8px;
    }

    .security-scores {
      display: flex;
      flex-direction: column;
      gap: 12px;
      margin-top: 16px;
    }

    .score-item {
      display: flex;
      align-items: center;
      gap: 12px;
    }

    .score-label {
      min-width: 100px;
      font-weight: 500;
    }

    .score-bar {
      flex: 1;
      max-width: 200px;
    }

    .exploitability-bar {
      --mdc-linear-progress-active-indicator-color: #f44336;
    }

    .impact-bar {
      --mdc-linear-progress-active-indicator-color: #ff9800;
    }

    .score-value {
      min-width: 40px;
      text-align: right;
      font-weight: 500;
    }

    .insight-actions {
      display: flex;
      gap: 12px;
      margin: 24px 0 16px 0;
      flex-wrap: wrap;
    }

    .insight-actions button {
      display: flex;
      align-items: center;
      gap: 8px;
    }

    .insight-timestamp {
      display: flex;
      align-items: center;
      gap: 8px;
      font-size: 0.8rem;
      color: rgba(0, 0, 0, 0.6);
      margin-top: 16px;
    }

    .no-insights-card,
    .loading-card {
      margin-top: 24px;
    }

    .no-insights,
    .loading-state {
      text-align: center;
      padding: 48px 24px;
    }

    .no-insights-icon,
    .loading-icon {
      font-size: 4rem;
      width: 4rem;
      height: 4rem;
      color: #9c27b0;
      margin-bottom: 16px;
    }

    @media (max-width: 768px) {
      .ai-insights-container {
        padding: 16px;
      }

      .insights-summary {
        flex-direction: column;
        gap: 16px;
        text-align: center;
      }

      .insight-header {
        flex-direction: column;
        align-items: flex-start;
        gap: 8px;
      }

      .insight-meta {
        flex-direction: column;
        align-items: flex-start;
        gap: 8px;
      }

      .insight-actions {
        flex-direction: column;
      }

      .insight-actions button {
        width: 100%;
        justify-content: center;
      }

      .security-scores {
        gap: 8px;
      }

      .score-item {
        flex-direction: column;
        align-items: flex-start;
        gap: 4px;
      }

      .score-bar {
        width: 100%;
        max-width: none;
      }
    }
  `]
})
export class AIInsightsPanelComponent implements OnInit, OnDestroy {
  private destroy$ = new Subject<void>();
  private apiService = inject(ApiService);
  private wsService = inject(WebSocketService);

  insights: AIInsight[] = [];
  loading = false;

  ngOnInit(): void {
    this.loadInsights();
    this.subscribeToRealTimeUpdates();
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  private loadInsights(): void {
    this.loading = true;
    this.apiService.getAIInsights().pipe(
      takeUntil(this.destroy$)
    ).subscribe({
      next: (response: any) => {
        if (response.success) {
          this.insights = this.processInsightsData(response.data);
        }
        this.loading = false;
      },
      error: (error: any) => {
        console.error('Error loading AI insights:', error);
        this.loading = false;
      }
    });
  }

  private processInsightsData(data: any[]): AIInsight[] {
    // Process and transform API data into AIInsight format
    return data.map(item => ({
      id: item.id || Math.random().toString(36).substr(2, 9),
      type: item.type || 'recommendation',
      title: item.title || 'AI Insight',
      description: item.description || '',
      severity: item.severity || 'medium',
      confidence: item.confidence || 85,
      affectedDependencies: item.affectedDependencies || [],
      recommendation: item.recommendation || '',
      impact: item.impact || '',
      timestamp: new Date(item.timestamp || Date.now()),
      metadata: item.metadata || {}
    }));
  }

  private subscribeToRealTimeUpdates(): void {
    this.wsService.getNotifications().pipe(
      takeUntil(this.destroy$)
    ).subscribe((notification) => {
      if (notification.type === 'ai_insight_generated') {
        this.loadInsights(); // Refresh insights when new ones are generated
      }
    });
  }

  getCriticalInsightsCount(): number {
    return this.insights.filter(insight => insight.severity === 'critical').length;
  }

  getAverageConfidence(): number {
    if (this.insights.length === 0) return 0;
    const total = this.insights.reduce((sum, insight) => sum + insight.confidence, 0);
    return Math.round(total / this.insights.length);
  }

  getInsightIcon(type: string): string {
    switch (type) {
      case 'breaking_change': return 'warning';
      case 'security_risk': return 'security';
      case 'performance_impact': return 'speed';
      case 'compatibility': return 'check_circle';
      case 'recommendation': return 'lightbulb';
      default: return 'info';
    }
  }

  trackByInsightId(index: number, insight: AIInsight): string {
    return insight.id;
  }

  // Action handlers
  applyRecommendation(insight: AIInsight): void {
    console.log('Apply recommendation for insight:', insight);
    // Implement recommendation application logic
  }

  viewDetailedAnalysis(insight: AIInsight): void {
    console.log('View detailed analysis for insight:', insight);
    // Open detailed analysis dialog or navigate to analysis page
  }

  dismissInsight(insight: AIInsight): void {
    console.log('Dismiss insight:', insight);
    // Remove insight from list and mark as dismissed in backend
    this.insights = this.insights.filter(i => i.id !== insight.id);
  }

  triggerAnalysis(): void {
    console.log('Trigger AI analysis');
    this.loading = true;
    // Trigger new AI analysis via API
    setTimeout(() => {
      this.loadInsights();
    }, 2000); // Simulate analysis time
  }
}
