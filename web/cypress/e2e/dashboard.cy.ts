describe('AI Dependency Manager - Dashboard E2E Tests', () => {
  beforeEach(() => {
    // Visit the dashboard page
    cy.visit('/dashboard');
    
    // Wait for the page to load
    cy.get('[data-cy="dashboard-container"]', { timeout: 10000 }).should('be.visible');
  });

  describe('Dashboard Loading and Display', () => {
    it('should load dashboard with mock data', () => {
      // Check that loading spinner disappears
      cy.get('[data-cy="loading-spinner"]').should('not.exist');
      
      // Verify dashboard stats are displayed
      cy.get('[data-cy="total-projects"]').should('contain', '5');
      cy.get('[data-cy="total-dependencies"]').should('contain', '247');
      cy.get('[data-cy="outdated-dependencies"]').should('contain', '23');
      cy.get('[data-cy="security-issues"]').should('contain', '7');
      
      // Verify system health status
      cy.get('[data-cy="system-health"]').should('contain', 'warning');
    });

    it('should display project summaries', () => {
      // Check that project cards are displayed
      cy.get('[data-cy="project-card"]').should('have.length', 5);
      
      // Verify specific project details
      cy.get('[data-cy="project-card"]').first().within(() => {
        cy.get('[data-cy="project-name"]').should('contain', 'E-Commerce Frontend');
        cy.get('[data-cy="project-type"]').should('contain', 'npm');
        cy.get('[data-cy="dependency-count"]').should('contain', '89');
        cy.get('[data-cy="outdated-count"]').should('contain', '12');
        cy.get('[data-cy="security-count"]').should('contain', '3');
        cy.get('[data-cy="health-score"]').should('contain', '78');
        cy.get('[data-cy="risk-level"]').should('contain', 'medium');
      });
    });

    it('should display last scan times', () => {
      cy.get('[data-cy="last-scan-time"]').should('exist');
      cy.get('[data-cy="project-card"]').each(($card) => {
        cy.wrap($card).find('[data-cy="project-last-scan"]').should('exist');
      });
    });
  });

  describe('Dashboard Interactions', () => {
    it('should navigate to project details when clicked', () => {
      cy.get('[data-cy="project-card"]').first().click();
      cy.url().should('include', '/projects/1');
    });

    it('should trigger scan all action', () => {
      cy.get('[data-cy="scan-all-btn"]').click();
      
      // Should show confirmation or progress indicator
      cy.get('[data-cy="scan-progress"]').should('be.visible');
    });

    it('should refresh dashboard data', () => {
      cy.get('[data-cy="refresh-btn"]').click();
      
      // Loading state should appear briefly
      cy.get('[data-cy="loading-spinner"]').should('be.visible');
      cy.get('[data-cy="loading-spinner"]').should('not.exist');
    });
  });

  describe('Responsive Design', () => {
    it('should display correctly on mobile devices', () => {
      cy.viewport('iphone-x');
      
      // Check that dashboard adapts to mobile layout
      cy.get('[data-cy="dashboard-container"]').should('be.visible');
      cy.get('[data-cy="stats-grid"]').should('have.class', 'mobile-layout');
      
      // Project cards should stack vertically
      cy.get('[data-cy="project-card"]').should('be.visible');
    });

    it('should display correctly on tablet devices', () => {
      cy.viewport('ipad-2');
      
      cy.get('[data-cy="dashboard-container"]').should('be.visible');
      cy.get('[data-cy="project-card"]').should('have.length', 5);
    });
  });

  describe('Real-time Updates', () => {
    it('should handle WebSocket connection', () => {
      // Mock WebSocket connection
      cy.window().then((win) => {
        // Simulate WebSocket message
        const mockMessage = {
          type: 'system_status',
          status: 'healthy',
          timestamp: new Date().toISOString()
        };
        
        // Trigger WebSocket message handler
        win.dispatchEvent(new CustomEvent('websocket-message', { detail: mockMessage }));
      });
      
      // Verify that system status updates
      cy.get('[data-cy="system-health"]').should('contain', 'healthy');
    });

    it('should update scan progress in real-time', () => {
      // Start a scan
      cy.get('[data-cy="scan-all-btn"]').click();
      
      // Mock progress updates
      cy.window().then((win) => {
        const progressMessage = {
          type: 'scan_progress',
          projectId: 1,
          status: 'in_progress',
          progress: 50,
          message: 'Scanning dependencies...'
        };
        
        win.dispatchEvent(new CustomEvent('websocket-message', { detail: progressMessage }));
      });
      
      // Verify progress indicator updates
      cy.get('[data-cy="scan-progress"]').should('contain', '50%');
      cy.get('[data-cy="scan-message"]').should('contain', 'Scanning dependencies...');
    });
  });

  describe('Error Handling', () => {
    it('should handle API errors gracefully', () => {
      // Intercept API calls and return errors
      cy.intercept('GET', '/api/analytics', { statusCode: 500, body: { error: 'Server error' } });
      cy.intercept('GET', '/api/projects', { statusCode: 500, body: { error: 'Server error' } });
      
      cy.visit('/dashboard');
      
      // Should display error message
      cy.get('[data-cy="error-message"]').should('be.visible');
      cy.get('[data-cy="error-message"]').should('contain', 'Failed to load dashboard data');
    });

    it('should handle network connectivity issues', () => {
      // Simulate network failure
      cy.intercept('GET', '/api/**', { forceNetworkError: true });
      
      cy.visit('/dashboard');
      
      // Should display offline indicator
      cy.get('[data-cy="offline-indicator"]').should('be.visible');
    });
  });

  describe('Performance', () => {
    it('should load dashboard within acceptable time', () => {
      const startTime = Date.now();
      
      cy.visit('/dashboard');
      cy.get('[data-cy="dashboard-container"]').should('be.visible');
      
      cy.then(() => {
        const loadTime = Date.now() - startTime;
        expect(loadTime).to.be.lessThan(3000); // Should load within 3 seconds
      });
    });

    it('should handle large datasets efficiently', () => {
      // Mock large dataset
      cy.intercept('GET', '/api/projects', { fixture: 'large-project-dataset.json' });
      
      cy.visit('/dashboard');
      cy.get('[data-cy="project-card"]').should('have.length.at.least', 50);
      
      // Should still be responsive
      cy.get('[data-cy="project-card"]').first().click();
      cy.get('[data-cy="project-details"]').should('be.visible');
    });
  });
});
