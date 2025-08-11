describe('AI Dependency Manager - Dependency Table E2E Tests', () => {
  beforeEach(() => {
    // Visit the dependencies page
    cy.visit('/dependencies');
    
    // Wait for the dependency table to load
    cy.get('[data-cy="dependency-table"]', { timeout: 10000 }).should('be.visible');
  });

  describe('Dependency Table Loading and Display', () => {
    it('should load dependency table with mock data', () => {
      // Check that loading spinner disappears
      cy.get('[data-cy="loading-spinner"]').should('not.exist');
      
      // Verify table headers are present
      cy.get('[data-cy="table-header-name"]').should('contain', 'Name');
      cy.get('[data-cy="table-header-current-version"]').should('contain', 'Current');
      cy.get('[data-cy="table-header-latest-version"]').should('contain', 'Latest');
      cy.get('[data-cy="table-header-risk-level"]').should('contain', 'Risk');
      cy.get('[data-cy="table-header-status"]').should('contain', 'Status');
      
      // Verify dependency rows are displayed
      cy.get('[data-cy="dependency-row"]').should('have.length.at.least', 1);
    });

    it('should display dependency information correctly', () => {
      cy.get('[data-cy="dependency-row"]').first().within(() => {
        // Check dependency name and package manager icon
        cy.get('[data-cy="dependency-name"]').should('be.visible');
        cy.get('[data-cy="package-icon"]').should('be.visible');
        
        // Check version information
        cy.get('[data-cy="current-version"]').should('be.visible');
        cy.get('[data-cy="latest-version"]').should('be.visible');
        
        // Check risk level and status indicators
        cy.get('[data-cy="risk-level"]').should('be.visible');
        cy.get('[data-cy="status-indicator"]').should('be.visible');
        
        // Check security issues badge if present
        cy.get('[data-cy="security-badge"]').should('exist');
      });
    });

    it('should display correct package manager icons', () => {
      // Test different package manager types
      const packageTypes = ['npm', 'pip', 'maven', 'gradle'];
      
      packageTypes.forEach(type => {
        cy.get(`[data-cy="package-icon-${type}"]`).should('exist');
      });
    });
  });

  describe('Filtering and Search', () => {
    it('should filter dependencies by search term', () => {
      // Type in search box
      cy.get('[data-cy="search-input"]').type('lodash');
      
      // Verify filtered results
      cy.get('[data-cy="dependency-row"]').should('have.length.at.most', 5);
      cy.get('[data-cy="dependency-name"]').should('contain', 'lodash');
    });

    it('should filter by project', () => {
      // Open project filter
      cy.get('[data-cy="project-filter"]').click();
      cy.get('[data-cy="project-option-1"]').click();
      
      // Verify filtered results
      cy.get('[data-cy="dependency-row"]').should('be.visible');
      cy.get('[data-cy="filter-chip-project"]').should('contain', 'Project Name');
    });

    it('should filter by status', () => {
      // Open status filter
      cy.get('[data-cy="status-filter"]').click();
      cy.get('[data-cy="status-option-outdated"]').click();
      
      // Verify filtered results show only outdated dependencies
      cy.get('[data-cy="dependency-row"]').each(($row) => {
        cy.wrap($row).find('[data-cy="status-indicator"]').should('have.class', 'outdated');
      });
    });

    it('should filter by risk level', () => {
      // Open risk filter
      cy.get('[data-cy="risk-filter"]').click();
      cy.get('[data-cy="risk-option-high"]').click();
      
      // Verify filtered results show only high-risk dependencies
      cy.get('[data-cy="dependency-row"]').each(($row) => {
        cy.wrap($row).find('[data-cy="risk-level"]').should('contain', 'high');
      });
    });

    it('should clear all filters', () => {
      // Apply multiple filters
      cy.get('[data-cy="search-input"]').type('test');
      cy.get('[data-cy="status-filter"]').click();
      cy.get('[data-cy="status-option-outdated"]').click();
      
      // Clear all filters
      cy.get('[data-cy="clear-filters-btn"]').click();
      
      // Verify all filters are cleared
      cy.get('[data-cy="search-input"]').should('have.value', '');
      cy.get('[data-cy="filter-chip"]').should('not.exist');
      cy.get('[data-cy="dependency-row"]').should('have.length.at.least', 10);
    });
  });

  describe('Row Selection and Bulk Actions', () => {
    it('should select individual dependencies', () => {
      // Select first dependency
      cy.get('[data-cy="dependency-row"]').first().find('[data-cy="row-checkbox"]').click();
      
      // Verify selection
      cy.get('[data-cy="dependency-row"]').first().should('have.class', 'selected');
      cy.get('[data-cy="selected-count"]').should('contain', '1');
    });

    it('should select all dependencies', () => {
      // Click select all checkbox
      cy.get('[data-cy="select-all-checkbox"]').click();
      
      // Verify all rows are selected
      cy.get('[data-cy="dependency-row"]').should('have.class', 'selected');
      cy.get('[data-cy="selected-count"]').should('contain.text', 'All');
    });

    it('should perform bulk update action', () => {
      // Select multiple dependencies
      cy.get('[data-cy="dependency-row"]').first().find('[data-cy="row-checkbox"]').click();
      cy.get('[data-cy="dependency-row"]').eq(1).find('[data-cy="row-checkbox"]').click();
      
      // Click bulk update button
      cy.get('[data-cy="bulk-update-btn"]').click();
      
      // Verify update confirmation dialog
      cy.get('[data-cy="update-confirmation-dialog"]').should('be.visible');
      cy.get('[data-cy="confirm-update-btn"]').click();
      
      // Verify update progress
      cy.get('[data-cy="update-progress"]').should('be.visible');
    });
  });

  describe('Individual Dependency Actions', () => {
    it('should view dependency details', () => {
      cy.get('[data-cy="dependency-row"]').first().within(() => {
        cy.get('[data-cy="view-details-btn"]').click();
      });
      
      // Verify details dialog opens
      cy.get('[data-cy="dependency-details-dialog"]').should('be.visible');
      cy.get('[data-cy="dependency-details-name"]').should('be.visible');
      cy.get('[data-cy="dependency-details-versions"]').should('be.visible');
      cy.get('[data-cy="dependency-details-security"]').should('be.visible');
    });

    it('should update individual dependency', () => {
      cy.get('[data-cy="dependency-row"]').first().within(() => {
        cy.get('[data-cy="update-dependency-btn"]').click();
      });
      
      // Verify update confirmation
      cy.get('[data-cy="update-confirmation-dialog"]').should('be.visible');
      cy.get('[data-cy="confirm-single-update-btn"]').click();
      
      // Verify update progress
      cy.get('[data-cy="single-update-progress"]').should('be.visible');
    });

    it('should view changelog', () => {
      cy.get('[data-cy="dependency-row"]').first().within(() => {
        cy.get('[data-cy="view-changelog-btn"]').click();
      });
      
      // Verify changelog dialog opens
      cy.get('[data-cy="changelog-dialog"]').should('be.visible');
      cy.get('[data-cy="changelog-content"]').should('be.visible');
    });

    it('should analyze impact', () => {
      cy.get('[data-cy="dependency-row"]').first().within(() => {
        cy.get('[data-cy="analyze-impact-btn"]').click();
      });
      
      // Verify impact analysis dialog opens
      cy.get('[data-cy="impact-analysis-dialog"]').should('be.visible');
      cy.get('[data-cy="impact-analysis-results"]').should('be.visible');
    });
  });

  describe('Sorting and Pagination', () => {
    it('should sort by dependency name', () => {
      cy.get('[data-cy="table-header-name"]').click();
      
      // Verify sorting indicator
      cy.get('[data-cy="sort-indicator-name"]').should('have.class', 'asc');
      
      // Click again to reverse sort
      cy.get('[data-cy="table-header-name"]').click();
      cy.get('[data-cy="sort-indicator-name"]').should('have.class', 'desc');
    });

    it('should sort by risk level', () => {
      cy.get('[data-cy="table-header-risk-level"]').click();
      
      // Verify high-risk dependencies appear first
      cy.get('[data-cy="dependency-row"]').first().within(() => {
        cy.get('[data-cy="risk-level"]').should('contain', 'critical');
      });
    });

    it('should handle pagination', () => {
      // Assuming we have more than 10 dependencies
      cy.get('[data-cy="pagination-next"]').click();
      
      // Verify page change
      cy.get('[data-cy="current-page"]').should('contain', '2');
      
      // Go back to first page
      cy.get('[data-cy="pagination-prev"]').click();
      cy.get('[data-cy="current-page"]').should('contain', '1');
    });

    it('should change page size', () => {
      cy.get('[data-cy="page-size-selector"]').click();
      cy.get('[data-cy="page-size-25"]').click();
      
      // Verify more rows are displayed
      cy.get('[data-cy="dependency-row"]').should('have.length', 25);
    });
  });

  describe('Security Features', () => {
    it('should highlight security vulnerabilities', () => {
      // Find dependencies with security issues
      cy.get('[data-cy="dependency-row"]').each(($row) => {
        cy.wrap($row).find('[data-cy="security-badge"]').then(($badge) => {
          if ($badge.length > 0) {
            // Verify security styling
            cy.wrap($row).should('have.class', 'has-security-issues');
            cy.wrap($badge).should('be.visible');
            cy.wrap($badge).should('contain.text', 'CVE');
          }
        });
      });
    });

    it('should display security issue details', () => {
      // Click on a dependency with security issues
      cy.get('[data-cy="dependency-row"]').contains('[data-cy="security-badge"]', 'CVE').first().click();
      
      // Verify security details in dialog
      cy.get('[data-cy="dependency-details-dialog"]').should('be.visible');
      cy.get('[data-cy="security-issues-section"]').should('be.visible');
      cy.get('[data-cy="cve-details"]').should('be.visible');
    });
  });

  describe('Real-time Updates', () => {
    it('should update dependency status in real-time', () => {
      // Mock WebSocket message for dependency update
      cy.window().then((win) => {
        const updateMessage = {
          type: 'dependency_update',
          dependencyId: 1,
          status: 'updated',
          newVersion: '4.17.21'
        };
        
        win.dispatchEvent(new CustomEvent('websocket-message', { detail: updateMessage }));
      });
      
      // Verify dependency status updates
      cy.get('[data-cy="dependency-row"]').first().within(() => {
        cy.get('[data-cy="status-indicator"]').should('have.class', 'up-to-date');
        cy.get('[data-cy="current-version"]').should('contain', '4.17.21');
      });
    });
  });

  describe('Error Handling', () => {
    it('should handle dependency loading errors', () => {
      // Intercept API call and return error
      cy.intercept('GET', '/api/projects/*/dependencies', { statusCode: 500 });
      
      cy.visit('/dependencies');
      
      // Verify error message
      cy.get('[data-cy="error-message"]').should('be.visible');
      cy.get('[data-cy="error-message"]').should('contain', 'Failed to load dependencies');
    });

    it('should handle update failures gracefully', () => {
      // Intercept update API call and return error
      cy.intercept('POST', '/api/projects/*/updates/apply', { statusCode: 500 });
      
      // Try to update a dependency
      cy.get('[data-cy="dependency-row"]').first().within(() => {
        cy.get('[data-cy="update-dependency-btn"]').click();
      });
      
      cy.get('[data-cy="confirm-single-update-btn"]').click();
      
      // Verify error handling
      cy.get('[data-cy="update-error-message"]').should('be.visible');
      cy.get('[data-cy="update-error-message"]').should('contain', 'Update failed');
    });
  });

  describe('Accessibility', () => {
    it('should be keyboard navigable', () => {
      // Tab through table elements
      cy.get('[data-cy="search-input"]').focus().tab();
      cy.get('[data-cy="project-filter"]').should('be.focused').tab();
      cy.get('[data-cy="status-filter"]').should('be.focused').tab();
      cy.get('[data-cy="risk-filter"]').should('be.focused');
      
      // Navigate to table rows
      cy.get('[data-cy="dependency-row"]').first().focus();
      cy.focused().should('have.attr', 'data-cy', 'dependency-row');
    });

    it('should have proper ARIA labels', () => {
      cy.get('[data-cy="dependency-table"]').should('have.attr', 'role', 'table');
      cy.get('[data-cy="table-header-name"]').should('have.attr', 'role', 'columnheader');
      cy.get('[data-cy="dependency-row"]').should('have.attr', 'role', 'row');
    });
  });
});
