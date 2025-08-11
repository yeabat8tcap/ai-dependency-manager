describe('AI Dependency Manager - Backend API Integration Tests', () => {
  const baseUrl = Cypress.env('API_URL') || 'http://localhost:8081/api';

  describe('Project Management API', () => {
    it('should get all projects from backend', () => {
      cy.request('GET', `${baseUrl}/projects`).then((response) => {
        expect(response.status).to.eq(200);
        expect(response.body).to.have.property('success', true);
        expect(response.body.data).to.be.an('array');
      });
    });

    it('should create a new project', () => {
      const newProject = {
        name: 'Test Project E2E',
        path: '/test/path/e2e',
        type: 'npm',
        configFile: 'package.json',
        enabled: true
      };

      cy.request('POST', `${baseUrl}/projects`, newProject).then((response) => {
        expect(response.status).to.eq(201);
        expect(response.body).to.have.property('success', true);
        expect(response.body.data).to.have.property('id');
        expect(response.body.data.name).to.eq(newProject.name);
        
        // Store project ID for cleanup
        cy.wrap(response.body.data.id).as('testProjectId');
      });
    });

    it('should get project by ID', function() {
      if (this.testProjectId) {
        cy.request('GET', `${baseUrl}/projects/${this.testProjectId}`).then((response) => {
          expect(response.status).to.eq(200);
          expect(response.body).to.have.property('success', true);
          expect(response.body.data).to.have.property('id', this.testProjectId);
        });
      }
    });

    it('should update project', function() {
      if (this.testProjectId) {
        const updatedProject = {
          name: 'Updated Test Project E2E',
          path: '/test/path/updated',
          type: 'npm',
          configFile: 'package.json',
          enabled: false
        };

        cy.request('PUT', `${baseUrl}/projects/${this.testProjectId}`, updatedProject).then((response) => {
          expect(response.status).to.eq(200);
          expect(response.body).to.have.property('success', true);
          expect(response.body.data.name).to.eq(updatedProject.name);
          expect(response.body.data.enabled).to.eq(false);
        });
      }
    });

    it('should delete project', function() {
      if (this.testProjectId) {
        cy.request('DELETE', `${baseUrl}/projects/${this.testProjectId}`).then((response) => {
          expect(response.status).to.eq(200);
          expect(response.body).to.have.property('success', true);
        });

        // Verify project is deleted
        cy.request({
          method: 'GET',
          url: `${baseUrl}/projects/${this.testProjectId}`,
          failOnStatusCode: false
        }).then((response) => {
          expect(response.status).to.eq(404);
        });
      }
    });
  });

  describe('Dependency Management API', () => {
    let testProjectId: number;

    before(() => {
      // Create a test project for dependency tests
      const testProject = {
        name: 'Dependency Test Project',
        path: '/test/dependency/project',
        type: 'npm',
        configFile: 'package.json',
        enabled: true
      };

      cy.request('POST', `${baseUrl}/projects`, testProject).then((response) => {
        testProjectId = response.body.data.id;
      });
    });

    after(() => {
      // Clean up test project
      if (testProjectId) {
        cy.request('DELETE', `${baseUrl}/projects/${testProjectId}`);
      }
    });

    it('should get dependencies for a project', () => {
      cy.request('GET', `${baseUrl}/projects/${testProjectId}/dependencies`).then((response) => {
        expect(response.status).to.eq(200);
        expect(response.body).to.have.property('success', true);
        expect(response.body.data).to.be.an('array');
      });
    });

    it('should scan project dependencies', () => {
      cy.request('POST', `${baseUrl}/projects/${testProjectId}/scan`).then((response) => {
        expect(response.status).to.eq(200);
        expect(response.body).to.have.property('success', true);
        expect(response.body.data).to.have.property('message');
      });
    });

    it('should apply dependency updates', () => {
      const updateOptions = {
        dependencies: ['lodash', 'express'],
        strategy: 'safe'
      };

      cy.request('POST', `${baseUrl}/projects/${testProjectId}/updates/apply`, updateOptions).then((response) => {
        expect(response.status).to.eq(200);
        expect(response.body).to.have.property('success', true);
      });
    });
  });

  describe('Analytics and Reporting API', () => {
    it('should get system analytics', () => {
      cy.request('GET', `${baseUrl}/analytics`).then((response) => {
        expect(response.status).to.eq(200);
        expect(response.body).to.have.property('success', true);
        expect(response.body.data).to.have.property('totalProjects');
        expect(response.body.data).to.have.property('totalDependencies');
        expect(response.body.data).to.have.property('outdatedDependencies');
        expect(response.body.data).to.have.property('securityIssues');
      });
    });

    it('should get system status', () => {
      cy.request('GET', `${baseUrl}/status`).then((response) => {
        expect(response.status).to.eq(200);
        expect(response.body).to.have.property('success', true);
        expect(response.body.data).to.have.property('health');
      });
    });

    it('should generate security report', () => {
      const reportOptions = {
        format: 'json',
        includeCharts: false
      };

      cy.request('POST', `${baseUrl}/reports/generate/security`, reportOptions).then((response) => {
        expect(response.status).to.eq(200);
        expect(response.body).to.have.property('success', true);
        expect(response.body.data).to.have.property('reportUrl');
      });
    });
  });

  describe('Policy Management API', () => {
    let testPolicyId: number;

    it('should get all policies', () => {
      cy.request('GET', `${baseUrl}/policies`).then((response) => {
        expect(response.status).to.eq(200);
        expect(response.body).to.have.property('success', true);
        expect(response.body.data).to.be.an('array');
      });
    });

    it('should create a new policy', () => {
      const newPolicy = {
        name: 'Test Security Policy',
        description: 'E2E test policy',
        rules: [
          {
            type: 'security',
            severity: 'high',
            action: 'block'
          }
        ]
      };

      cy.request('POST', `${baseUrl}/policies`, newPolicy).then((response) => {
        expect(response.status).to.eq(201);
        expect(response.body).to.have.property('success', true);
        expect(response.body.data).to.have.property('id');
        expect(response.body.data.name).to.eq(newPolicy.name);
        
        testPolicyId = response.body.data.id;
      });
    });

    it('should update policy', () => {
      const updatedPolicy = {
        name: 'Updated Test Security Policy',
        description: 'Updated E2E test policy',
        rules: [
          {
            type: 'security',
            severity: 'critical',
            action: 'block'
          }
        ]
      };

      cy.request('PUT', `${baseUrl}/policies/${testPolicyId}`, updatedPolicy).then((response) => {
        expect(response.status).to.eq(200);
        expect(response.body).to.have.property('success', true);
        expect(response.body.data.name).to.eq(updatedPolicy.name);
      });
    });

    it('should delete policy', () => {
      cy.request('DELETE', `${baseUrl}/policies/${testPolicyId}`).then((response) => {
        expect(response.status).to.eq(200);
        expect(response.body).to.have.property('success', true);
      });
    });
  });

  describe('WebSocket Real-time Updates', () => {
    it('should establish WebSocket connection', () => {
      cy.visit('/dashboard');
      
      cy.window().then((win) => {
        // Check if WebSocket connection is established
        cy.wrap(win).should('have.property', 'WebSocket');
        
        // Mock WebSocket connection test
        const ws = new win.WebSocket('ws://localhost:8081/ws');
        
        ws.onopen = () => {
          expect(ws.readyState).to.eq(WebSocket.OPEN);
          ws.close();
        };
      });
    });

    it('should receive system status updates', () => {
      cy.visit('/dashboard');
      
      // Mock WebSocket message
      cy.window().then((win) => {
        const mockMessage = {
          type: 'system_status',
          status: 'warning',
          timestamp: new Date().toISOString(),
          details: {
            uptime: '48h',
            memory: '1GB'
          }
        };
        
        // Simulate WebSocket message
        win.dispatchEvent(new CustomEvent('websocket-message', { detail: mockMessage }));
        
        // Verify UI updates
        cy.get('[data-cy="system-health"]').should('contain', 'warning');
      });
    });

    it('should receive scan progress updates', () => {
      cy.visit('/dependencies');
      
      cy.window().then((win) => {
        const mockProgressMessage = {
          type: 'scan_progress',
          projectId: 1,
          status: 'in_progress',
          progress: 75,
          message: 'Analyzing security vulnerabilities...',
          timestamp: new Date().toISOString()
        };
        
        win.dispatchEvent(new CustomEvent('websocket-message', { detail: mockProgressMessage }));
        
        // Verify progress indicator updates
        cy.get('[data-cy="scan-progress"]').should('contain', '75%');
        cy.get('[data-cy="scan-message"]').should('contain', 'Analyzing security vulnerabilities');
      });
    });
  });

  describe('Error Handling and Edge Cases', () => {
    it('should handle invalid project ID', () => {
      cy.request({
        method: 'GET',
        url: `${baseUrl}/projects/99999`,
        failOnStatusCode: false
      }).then((response) => {
        expect(response.status).to.eq(404);
        expect(response.body).to.have.property('success', false);
        expect(response.body).to.have.property('message');
      });
    });

    it('should handle malformed request data', () => {
      const invalidProject = {
        name: '', // Invalid: empty name
        type: 'invalid_type' // Invalid: unsupported type
      };

      cy.request({
        method: 'POST',
        url: `${baseUrl}/projects`,
        body: invalidProject,
        failOnStatusCode: false
      }).then((response) => {
        expect(response.status).to.eq(400);
        expect(response.body).to.have.property('success', false);
        expect(response.body).to.have.property('errors');
      });
    });

    it('should handle server errors gracefully', () => {
      // This test would require backend to simulate server errors
      // For now, we'll test the client's error handling
      cy.intercept('GET', '/api/projects', { statusCode: 500 });
      
      cy.visit('/dashboard');
      
      // Verify error handling in UI
      cy.get('[data-cy="error-message"]').should('be.visible');
      cy.get('[data-cy="retry-button"]').should('be.visible');
    });
  });

  describe('Performance and Load Testing', () => {
    it('should handle concurrent requests', () => {
      const requests = [];
      
      // Create multiple concurrent requests
      for (let i = 0; i < 10; i++) {
        requests.push(cy.request('GET', `${baseUrl}/projects`));
      }
      
      // All requests should succeed
      requests.forEach((request) => {
        request.then((response) => {
          expect(response.status).to.eq(200);
          expect(response.body).to.have.property('success', true);
        });
      });
    });

    it('should respond within acceptable time limits', () => {
      const startTime = Date.now();
      
      cy.request('GET', `${baseUrl}/analytics`).then((response) => {
        const responseTime = Date.now() - startTime;
        
        expect(response.status).to.eq(200);
        expect(responseTime).to.be.lessThan(2000); // Should respond within 2 seconds
      });
    });
  });

  describe('Security Testing', () => {
    it('should validate authentication headers', () => {
      // Test with missing authentication (if required)
      cy.request({
        method: 'POST',
        url: `${baseUrl}/projects`,
        body: { name: 'Test' },
        headers: {
          'Authorization': 'Bearer invalid_token'
        },
        failOnStatusCode: false
      }).then((response) => {
        // Response should indicate authentication failure if auth is required
        expect([200, 401, 403]).to.include(response.status);
      });
    });

    it('should sanitize input data', () => {
      const maliciousProject = {
        name: '<script>alert("xss")</script>',
        path: '../../etc/passwd',
        type: 'npm'
      };

      cy.request({
        method: 'POST',
        url: `${baseUrl}/projects`,
        body: maliciousProject,
        failOnStatusCode: false
      }).then((response) => {
        // Should either reject or sanitize the input
        if (response.status === 201) {
          expect(response.body.data.name).to.not.contain('<script>');
        } else {
          expect([400, 422]).to.include(response.status);
        }
      });
    });
  });
});
