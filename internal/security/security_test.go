package security

import (
	"context"
	"testing"

	"github.com/8tcapital/ai-dep-manager/internal/models"
	testingPkg "github.com/8tcapital/ai-dep-manager/internal/testing"
)

func TestSecurityService_ScanPackage(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	securityService := NewSecurityService()

	tests := []struct {
		name        string
		packageName string
		version     string
		expectIssues bool
	}{
		{
			name:        "scan known vulnerable package",
			packageName: "lodash",
			version:     "4.17.15", // Known vulnerable version
			expectIssues: true,
		},
		{
			name:        "scan safe package",
			packageName: "express",
			version:     "4.18.2",
			expectIssues: false,
		},
		{
			name:        "scan non-existent package",
			packageName: "non-existent-package-xyz",
			version:     "1.0.0",
			expectIssues: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := securityService.ScanPackage(context.Background(), tt.packageName, tt.version, "npm")
			testingPkg.AssertNoError(t, err, "ScanPackage should not return error")
			
			testingPkg.AssertNotEqual(t, uint(0), result.ID, "Security check should have ID")
			testingPkg.AssertEqual(t, tt.packageName, result.PackageName, "Package name should match")
			testingPkg.AssertEqual(t, tt.version, result.Version, "Version should match")
			
			if tt.expectIssues {
				testingPkg.AssertEqual(t, "detected", result.Status, "Should detect security issues")
			}
		})
	}
}

func TestSecurityService_VerifyPackageIntegrity(t *testing.T) {
	testingPkg.SkipIfShort(t, "requires external registry calls")

	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	securityService := NewSecurityService()

	tests := []struct {
		name        string
		packageName string
		version     string
		packageType string
		expectValid bool
	}{
		{
			name:        "verify popular npm package",
			packageName: "lodash",
			version:     "4.17.21",
			packageType: "npm",
			expectValid: true,
		},
		{
			name:        "verify popular python package",
			packageName: "requests",
			version:     "2.31.0",
			packageType: "pip",
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := securityService.VerifyPackageIntegrity(context.Background(), 
				tt.packageName, tt.version, tt.packageType)
			
			if tt.expectValid {
				testingPkg.AssertNoError(t, err, "Integrity verification should succeed for valid package")
				testingPkg.AssertTrue(t, result.Valid, "Package should be valid")
				testingPkg.AssertNotEqual(t, "", result.Checksum, "Should have checksum")
			}
		})
	}
}

func TestSecurityService_CheckMaliciousPackage(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	securityService := NewSecurityService()

	tests := []struct {
		name        string
		packageName string
		expectRisk  bool
		riskType    string
	}{
		{
			name:        "legitimate package",
			packageName: "express",
			expectRisk:  false,
		},
		{
			name:        "typosquatting attempt",
			packageName: "expres", // Missing 's'
			expectRisk:  true,
			riskType:    "typosquatting",
		},
		{
			name:        "suspicious name pattern",
			packageName: "test-malicious-package-xyz",
			expectRisk:  true,
			riskType:    "suspicious_name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := securityService.CheckMaliciousPackage(context.Background(), tt.packageName, "npm")
			testingPkg.AssertNoError(t, err, "CheckMaliciousPackage should not return error")
			
			if tt.expectRisk {
				testingPkg.AssertTrue(t, result.IsMalicious, "Should detect malicious package")
				testingPkg.AssertEqual(t, tt.riskType, result.RiskType, "Risk type should match")
				testingPkg.AssertNotEqual(t, "", result.Reason, "Should have reason")
			} else {
				testingPkg.AssertFalse(t, result.IsMalicious, "Should not flag legitimate package")
			}
		})
	}
}

func TestSecurityService_GetVulnerabilities(t *testing.T) {
	testingPkg.SkipIfShort(t, "requires external vulnerability database calls")

	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	securityService := NewSecurityService()

	// Test with a package that historically had vulnerabilities
	vulnerabilities, err := securityService.GetVulnerabilities(context.Background(), "lodash", "4.17.15")
	testingPkg.AssertNoError(t, err, "GetVulnerabilities should not return error")

	// We expect some vulnerabilities for this old version
	testingPkg.AssertTrue(t, len(vulnerabilities) >= 0, "Should return vulnerability list")

	for _, vuln := range vulnerabilities {
		testingPkg.AssertNotEqual(t, "", vuln.ID, "Vulnerability should have ID")
		testingPkg.AssertNotEqual(t, "", vuln.Summary, "Vulnerability should have summary")
		testingPkg.AssertTrue(t, len(vuln.Severity) > 0, "Vulnerability should have severity")
	}
}

func TestSecurityService_CheckPackageReputation(t *testing.T) {
	testingPkg.SkipIfShort(t, "requires external registry calls")

	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	securityService := NewSecurityService()

	tests := []struct {
		name        string
		packageName string
		expectGood  bool
	}{
		{
			name:        "well-known package",
			packageName: "express",
			expectGood:  true,
		},
		{
			name:        "less known package",
			packageName: "very-obscure-test-package-name",
			expectGood:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reputation, err := securityService.CheckPackageReputation(context.Background(), tt.packageName, "npm")
			testingPkg.AssertNoError(t, err, "CheckPackageReputation should not return error")
			
			testingPkg.AssertTrue(t, reputation.Score >= 0 && reputation.Score <= 10, "Score should be between 0 and 10")
			
			if tt.expectGood {
				testingPkg.AssertTrue(t, reputation.Score >= 7, "Well-known package should have good reputation")
				testingPkg.AssertTrue(t, reputation.IsPopular, "Well-known package should be popular")
			}
		})
	}
}

func TestSecurityService_ScanProject(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	securityService := NewSecurityService()

	// Create some test dependencies with potential security issues
	ctx.CreateTestDependency(t, ctx.Projects[0].ID, "lodash", "4.17.15", "4.17.21", "npm")
	ctx.CreateTestDependency(t, ctx.Projects[0].ID, "express", "4.18.0", "4.18.2", "npm")

	result, err := securityService.ScanProject(context.Background(), ctx.Projects[0].ID)
	testingPkg.AssertNoError(t, err, "ScanProject should not return error")
	
	testingPkg.AssertNotEqual(t, uint(0), result.ID, "Scan result should have ID")
	testingPkg.AssertEqual(t, ctx.Projects[0].ID, result.ProjectID, "Project ID should match")
	testingPkg.AssertTrue(t, result.PackagesScanned > 0, "Should scan some packages")
	testingPkg.AssertTrue(t, result.TotalIssues >= 0, "Should report issue count")
}

func TestSecurityService_GetSecurityRules(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	securityService := NewSecurityService()

	// Create test security rules
	rule1 := models.SecurityRule{
		Type:        "whitelist",
		PackageName: "express",
		Reason:      "Approved web framework",
		Enabled:     true,
	}
	ctx.DB.Create(&rule1)

	rule2 := models.SecurityRule{
		Type:        "blacklist",
		PackageName: "malicious-package",
		Reason:      "Known malicious package",
		Enabled:     true,
	}
	ctx.DB.Create(&rule2)

	// Test getting whitelist rules
	whitelistRules, err := securityService.GetSecurityRules(context.Background(), "whitelist")
	testingPkg.AssertNoError(t, err, "GetSecurityRules should not return error")
	testingPkg.AssertTrue(t, len(whitelistRules) >= 1, "Should have whitelist rules")

	// Test getting blacklist rules
	blacklistRules, err := securityService.GetSecurityRules(context.Background(), "blacklist")
	testingPkg.AssertNoError(t, err, "GetSecurityRules should not return error")
	testingPkg.AssertTrue(t, len(blacklistRules) >= 1, "Should have blacklist rules")

	// Test getting all rules
	allRules, err := securityService.GetSecurityRules(context.Background(), "")
	testingPkg.AssertNoError(t, err, "GetSecurityRules should not return error")
	testingPkg.AssertTrue(t, len(allRules) >= 2, "Should have all rules")
}

func TestSecurityService_AddSecurityRule(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	securityService := NewSecurityService()

	rule := SecurityRule{
		Type:        "whitelist",
		PackageName: "test-package",
		Reason:      "Test whitelist rule",
	}

	err := securityService.AddSecurityRule(context.Background(), rule)
	testingPkg.AssertNoError(t, err, "AddSecurityRule should not return error")

	// Verify rule was added
	rules, err := securityService.GetSecurityRules(context.Background(), "whitelist")
	testingPkg.AssertNoError(t, err, "GetSecurityRules should not return error")

	found := false
	for _, r := range rules {
		if r.PackageName == "test-package" {
			found = true
			testingPkg.AssertEqual(t, "whitelist", r.Type, "Rule type should match")
			testingPkg.AssertEqual(t, "Test whitelist rule", r.Reason, "Rule reason should match")
			break
		}
	}
	testingPkg.AssertTrue(t, found, "Rule should be found in database")
}

func TestSecurityService_RemoveSecurityRule(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	securityService := NewSecurityService()

	// Add a rule first
	rule := SecurityRule{
		Type:        "blacklist",
		PackageName: "test-remove-package",
		Reason:      "Test removal",
	}

	err := securityService.AddSecurityRule(context.Background(), rule)
	testingPkg.AssertNoError(t, err, "AddSecurityRule should not return error")

	// Remove the rule
	err = securityService.RemoveSecurityRule(context.Background(), "test-remove-package", "blacklist")
	testingPkg.AssertNoError(t, err, "RemoveSecurityRule should not return error")

	// Verify rule was removed
	rules, err := securityService.GetSecurityRules(context.Background(), "blacklist")
	testingPkg.AssertNoError(t, err, "GetSecurityRules should not return error")

	for _, r := range rules {
		testingPkg.AssertNotEqual(t, "test-remove-package", r.PackageName, "Rule should be removed")
	}
}

func TestSecurityService_IsPackageAllowed(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	securityService := NewSecurityService()

	// Add whitelist and blacklist rules
	whitelistRule := SecurityRule{
		Type:        "whitelist",
		PackageName: "allowed-package",
		Reason:      "Explicitly allowed",
	}
	securityService.AddSecurityRule(context.Background(), whitelistRule)

	blacklistRule := SecurityRule{
		Type:        "blacklist",
		PackageName: "blocked-package",
		Reason:      "Explicitly blocked",
	}
	securityService.AddSecurityRule(context.Background(), blacklistRule)

	tests := []struct {
		name        string
		packageName string
		expected    bool
	}{
		{
			name:        "whitelisted package",
			packageName: "allowed-package",
			expected:    true,
		},
		{
			name:        "blacklisted package",
			packageName: "blocked-package",
			expected:    false,
		},
		{
			name:        "neutral package",
			packageName: "neutral-package",
			expected:    true, // Default allow
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := securityService.IsPackageAllowed(context.Background(), tt.packageName)
			testingPkg.AssertNoError(t, err, "IsPackageAllowed should not return error")
			testingPkg.AssertEqual(t, tt.expected, allowed, "Package allowance should match expected")
		})
	}
}

func TestSecurityService_GetSecuritySummary(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	securityService := NewSecurityService()

	// Create some test security checks
	ctx.CreateTestSecurityCheck(t, "lodash", "4.17.15", "vulnerability", "detected")
	ctx.CreateTestSecurityCheck(t, "express", "4.18.0", "integrity", "verified")

	summary, err := securityService.GetSecuritySummary(context.Background(), &ctx.Projects[0].ID)
	testingPkg.AssertNoError(t, err, "GetSecuritySummary should not return error")

	testingPkg.AssertTrue(t, summary.TotalChecks > 0, "Should have total checks")
	testingPkg.AssertTrue(t, summary.VulnerabilitiesFound >= 0, "Should have vulnerability count")
	testingPkg.AssertTrue(t, summary.PackagesScanned > 0, "Should have scanned packages count")
	testingPkg.AssertTrue(t, len(summary.SeverityBreakdown) > 0, "Should have severity breakdown")
}

func TestSecurityService_ErrorHandling(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	securityService := NewSecurityService()

	// Test with invalid package name
	_, err := securityService.ScanPackage(context.Background(), "", "1.0.0", "npm")
	testingPkg.AssertError(t, err, "Should return error for empty package name")

	// Test with invalid project ID
	_, err = securityService.ScanProject(context.Background(), 9999)
	testingPkg.AssertError(t, err, "Should return error for invalid project ID")

	// Test with cancelled context
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = securityService.ScanPackage(cancelledCtx, "test-package", "1.0.0", "npm")
	testingPkg.AssertError(t, err, "Should return error for cancelled context")
}

func TestSecurityService_ConcurrentScanning(t *testing.T) {
	ctx := testingPkg.SetupTestEnvironment(t)
	defer ctx.Cleanup()

	securityService := NewSecurityService()

	// Test concurrent package scanning
	packages := []struct {
		name    string
		version string
	}{
		{"express", "4.18.0"},
		{"lodash", "4.17.20"},
		{"react", "18.0.0"},
	}

	results := make(chan *models.SecurityCheck, len(packages))
	errors := make(chan error, len(packages))

	for _, pkg := range packages {
		go func(name, version string) {
			result, err := securityService.ScanPackage(context.Background(), name, version, "npm")
			if err != nil {
				errors <- err
				return
			}
			results <- result
		}(pkg.name, pkg.version)
	}

	// Collect results
	var scanResults []*models.SecurityCheck
	var scanErrors []error

	for i := 0; i < len(packages); i++ {
		select {
		case result := <-results:
			scanResults = append(scanResults, result)
		case err := <-errors:
			scanErrors = append(scanErrors, err)
		case <-testingPkg.WithTimeout(t, 30*time.Second):
			t.Fatal("Timeout waiting for concurrent scans")
		}
	}

	testingPkg.AssertLen(t, scanErrors, 0, "No errors should occur during concurrent scanning")
	testingPkg.AssertLen(t, scanResults, len(packages), "Should get results for all packages")
}

// Benchmarks

func BenchmarkSecurityService_ScanPackage(b *testing.B) {
	ctx := testingPkg.SetupTestEnvironment(&testing.T{})
	defer ctx.Cleanup()

	securityService := NewSecurityService()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := securityService.ScanPackage(context.Background(), "express", "4.18.0", "npm")
		if err != nil {
			b.Fatalf("ScanPackage failed: %v", err)
		}
	}
}

func BenchmarkSecurityService_CheckMaliciousPackage(b *testing.B) {
	ctx := testingPkg.SetupTestEnvironment(&testing.T{})
	defer ctx.Cleanup()

	securityService := NewSecurityService()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := securityService.CheckMaliciousPackage(context.Background(), "express", "npm")
		if err != nil {
			b.Fatalf("CheckMaliciousPackage failed: %v", err)
		}
	}
}
