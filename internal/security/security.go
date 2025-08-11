package security

import (
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/config"
	"github.com/8tcapital/ai-dep-manager/internal/database"
	"github.com/8tcapital/ai-dep-manager/internal/logger"
	"github.com/8tcapital/ai-dep-manager/internal/models"
	"gorm.io/gorm"
)

// SecurityService handles security-related operations
type SecurityService struct {
	config *config.Config
	db     *gorm.DB
	client *http.Client
}

// SecurityCheck represents a security check result
type SecurityCheck struct {
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Package     string                 `json:"package"`
	Version     string                 `json:"version"`
	CVE         string                 `json:"cve,omitempty"`
	CVSS        float64                `json:"cvss,omitempty"`
	References  []string               `json:"references,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// IntegrityCheck represents package integrity verification
type IntegrityCheck struct {
	Package         string            `json:"package"`
	Version         string            `json:"version"`
	ExpectedHashes  map[string]string `json:"expected_hashes"`
	ActualHashes    map[string]string `json:"actual_hashes"`
	Verified        bool              `json:"verified"`
	TrustedSource   bool              `json:"trusted_source"`
	SignatureValid  bool              `json:"signature_valid"`
	Timestamp       time.Time         `json:"timestamp"`
}

// VulnerabilityDatabase represents vulnerability data
type VulnerabilityDatabase struct {
	LastUpdated   time.Time                        `json:"last_updated"`
	Vulnerabilities map[string][]VulnerabilityEntry `json:"vulnerabilities"`
}

// VulnerabilityEntry represents a single vulnerability
type VulnerabilityEntry struct {
	ID          string   `json:"id"`
	CVE         string   `json:"cve"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Severity    string   `json:"severity"`
	CVSS        float64  `json:"cvss"`
	Versions    []string `json:"versions"`
	References  []string `json:"references"`
	PublishedAt string   `json:"published_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// NewSecurityService creates a new security service
func NewSecurityService(cfg *config.Config) *SecurityService {
	return &SecurityService{
		config: cfg,
		db:     database.GetDB(),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// VerifyPackageIntegrity verifies the integrity of a package
func (ss *SecurityService) VerifyPackageIntegrity(ctx context.Context, packageName, version, packageType string) (*IntegrityCheck, error) {
	logger.Debug("Verifying integrity for package: %s@%s (%s)", packageName, version, packageType)
	
	check := &IntegrityCheck{
		Package:        packageName,
		Version:        version,
		ExpectedHashes: make(map[string]string),
		ActualHashes:   make(map[string]string),
		Timestamp:      time.Now(),
	}
	
	// Get expected hashes from registry
	expectedHashes, err := ss.getExpectedHashes(ctx, packageName, version, packageType)
	if err != nil {
		logger.Warn("Failed to get expected hashes for %s@%s: %v", packageName, version, err)
	} else {
		check.ExpectedHashes = expectedHashes
		check.TrustedSource = true
	}
	
	// Download and verify package
	actualHashes, err := ss.downloadAndHash(ctx, packageName, version, packageType)
	if err != nil {
		return check, fmt.Errorf("failed to download and hash package: %w", err)
	}
	
	check.ActualHashes = actualHashes
	
	// Compare hashes
	check.Verified = ss.compareHashes(check.ExpectedHashes, check.ActualHashes)
	
	// Store integrity check result
	if err := ss.storeIntegrityCheck(check); err != nil {
		logger.Warn("Failed to store integrity check: %v", err)
	}
	
	logger.Info("Integrity check for %s@%s: verified=%t", packageName, version, check.Verified)
	return check, nil
}

// ScanForVulnerabilities scans a package for known vulnerabilities
func (ss *SecurityService) ScanForVulnerabilities(ctx context.Context, packageName, version, packageType string) ([]SecurityCheck, error) {
	logger.Debug("Scanning for vulnerabilities: %s@%s (%s)", packageName, version, packageType)
	
	var checks []SecurityCheck
	
	// Check against vulnerability databases
	vulnChecks, err := ss.checkVulnerabilityDatabases(ctx, packageName, version, packageType)
	if err != nil {
		logger.Warn("Failed to check vulnerability databases: %v", err)
	} else {
		checks = append(checks, vulnChecks...)
	}
	
	// Check for malicious patterns
	maliciousChecks, err := ss.checkMaliciousPatterns(ctx, packageName, version, packageType)
	if err != nil {
		logger.Warn("Failed to check malicious patterns: %v", err)
	} else {
		checks = append(checks, maliciousChecks...)
	}
	
	// Check package reputation
	reputationChecks, err := ss.checkPackageReputation(ctx, packageName, packageType)
	if err != nil {
		logger.Warn("Failed to check package reputation: %v", err)
	} else {
		checks = append(checks, reputationChecks...)
	}
	
	// Store security checks
	for _, check := range checks {
		if err := ss.storeSecurityCheck(&check); err != nil {
			logger.Warn("Failed to store security check: %v", err)
		}
	}
	
	logger.Info("Found %d security issues for %s@%s", len(checks), packageName, version)
	return checks, nil
}

// IsPackageAllowed checks if a package is allowed based on whitelist/blacklist
func (ss *SecurityService) IsPackageAllowed(packageName, packageType string) (bool, string) {
	// Check blacklist first
	if ss.isBlacklisted(packageName, packageType) {
		return false, "Package is blacklisted"
	}
	
	// Check whitelist if enabled
	if ss.config.Security.WhitelistEnabled {
		if !ss.isWhitelisted(packageName, packageType) {
			return false, "Package is not whitelisted"
		}
	}
	
	return true, ""
}

// UpdateVulnerabilityDatabase updates the vulnerability database
func (ss *SecurityService) UpdateVulnerabilityDatabase(ctx context.Context) error {
	logger.Info("Updating vulnerability database")
	
	// Update from multiple sources
	sources := []string{
		"https://api.osv.dev/v1/query",
		"https://api.github.com/advisories",
		"https://services.nvd.nist.gov/rest/json/cves/2.0",
	}
	
	for _, source := range sources {
		if err := ss.updateFromSource(ctx, source); err != nil {
			logger.Warn("Failed to update from source %s: %v", source, err)
		}
	}
	
	logger.Info("Vulnerability database update completed")
	return nil
}

// Private methods

func (ss *SecurityService) getExpectedHashes(ctx context.Context, packageName, version, packageType string) (map[string]string, error) {
	hashes := make(map[string]string)
	
	switch packageType {
	case "npm":
		return ss.getNpmHashes(ctx, packageName, version)
	case "pip":
		return ss.getPipHashes(ctx, packageName, version)
	case "maven":
		return ss.getMavenHashes(ctx, packageName, version)
	default:
		return hashes, fmt.Errorf("unsupported package type: %s", packageType)
	}
}

func (ss *SecurityService) getNpmHashes(ctx context.Context, packageName, version string) (map[string]string, error) {
	url := fmt.Sprintf("https://registry.npmjs.org/%s/%s", packageName, version)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	resp, err := ss.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned status %d", resp.StatusCode)
	}
	
	var data struct {
		Dist struct {
			Shasum   string `json:"shasum"`
			Integrity string `json:"integrity"`
		} `json:"dist"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	
	hashes := make(map[string]string)
	if data.Dist.Shasum != "" {
		hashes["sha1"] = data.Dist.Shasum
	}
	if data.Dist.Integrity != "" {
		hashes["integrity"] = data.Dist.Integrity
	}
	
	return hashes, nil
}

func (ss *SecurityService) getPipHashes(ctx context.Context, packageName, version string) (map[string]string, error) {
	url := fmt.Sprintf("https://pypi.org/pypi/%s/%s/json", packageName, version)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	resp, err := ss.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned status %d", resp.StatusCode)
	}
	
	var data struct {
		URLs []struct {
			Digests map[string]string `json:"digests"`
		} `json:"urls"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	
	hashes := make(map[string]string)
	for _, url := range data.URLs {
		for algo, hash := range url.Digests {
			hashes[algo] = hash
		}
		break // Use first URL's hashes
	}
	
	return hashes, nil
}

func (ss *SecurityService) getMavenHashes(ctx context.Context, packageName, version string) (map[string]string, error) {
	// Maven Central provides checksums as separate files
	parts := strings.Split(packageName, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid Maven package name format")
	}
	
	groupId := strings.ReplaceAll(parts[0], ".", "/")
	artifactId := parts[1]
	
	baseUrl := fmt.Sprintf("https://repo1.maven.org/maven2/%s/%s/%s", groupId, artifactId, version)
	
	hashes := make(map[string]string)
	
	// Try to get SHA1 and MD5 checksums
	for _, algo := range []string{"sha1", "md5"} {
		url := fmt.Sprintf("%s/%s-%s.jar.%s", baseUrl, artifactId, version, algo)
		
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			continue
		}
		
		resp, err := ss.client.Do(req)
		if err != nil {
			continue
		}
		
		if resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err == nil {
				hashes[algo] = strings.TrimSpace(string(body))
			}
		}
		resp.Body.Close()
	}
	
	return hashes, nil
}

func (ss *SecurityService) downloadAndHash(ctx context.Context, packageName, version, packageType string) (map[string]string, error) {
	// This is a simplified implementation
	// In practice, you'd download the actual package and compute hashes
	
	hashes := make(map[string]string)
	
	// For demonstration, we'll create placeholder hashes
	// In a real implementation, you would:
	// 1. Download the package tarball/wheel/jar
	// 2. Compute SHA256, SHA512, etc.
	// 3. Return the actual hashes
	
	data := fmt.Sprintf("%s@%s", packageName, version)
	
	sha256Hash := sha256.Sum256([]byte(data))
	hashes["sha256"] = hex.EncodeToString(sha256Hash[:])
	
	sha512Hash := sha512.Sum512([]byte(data))
	hashes["sha512"] = hex.EncodeToString(sha512Hash[:])
	
	return hashes, nil
}

func (ss *SecurityService) compareHashes(expected, actual map[string]string) bool {
	if len(expected) == 0 {
		// No expected hashes to compare against
		return false
	}
	
	for algo, expectedHash := range expected {
		if actualHash, exists := actual[algo]; exists {
			if expectedHash == actualHash {
				return true // At least one hash matches
			}
		}
	}
	
	return false
}

func (ss *SecurityService) checkVulnerabilityDatabases(ctx context.Context, packageName, version, packageType string) ([]SecurityCheck, error) {
	var checks []SecurityCheck
	
	// Check OSV database
	osvChecks, err := ss.queryOSV(ctx, packageName, version, packageType)
	if err == nil {
		checks = append(checks, osvChecks...)
	}
	
	// Check GitHub Security Advisories
	ghChecks, err := ss.queryGitHubAdvisories(ctx, packageName, version, packageType)
	if err == nil {
		checks = append(checks, ghChecks...)
	}
	
	return checks, nil
}

func (ss *SecurityService) queryOSV(ctx context.Context, packageName, version, packageType string) ([]SecurityCheck, error) {
	// Query OSV (Open Source Vulnerabilities) database
	ecosystem := ss.getOSVEcosystem(packageType)
	
	query := map[string]interface{}{
		"package": map[string]string{
			"name":      packageName,
			"ecosystem": ecosystem,
		},
		"version": version,
	}
	
	queryData, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.osv.dev/v1/query", strings.NewReader(string(queryData)))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := ss.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OSV API returned status %d", resp.StatusCode)
	}
	
	var result struct {
		Vulns []struct {
			ID      string `json:"id"`
			Summary string `json:"summary"`
			Details string `json:"details"`
			Aliases []string `json:"aliases"`
			Severity []struct {
				Type  string  `json:"type"`
				Score string  `json:"score"`
			} `json:"severity"`
			References []struct {
				URL string `json:"url"`
			} `json:"references"`
		} `json:"vulns"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	var checks []SecurityCheck
	for _, vuln := range result.Vulns {
		severity := "medium"
		cvss := 5.0
		
		if len(vuln.Severity) > 0 {
			severity = ss.mapSeverity(vuln.Severity[0].Score)
			if score, err := ss.parseCVSS(vuln.Severity[0].Score); err == nil {
				cvss = score
			}
		}
		
		var references []string
		for _, ref := range vuln.References {
			references = append(references, ref.URL)
		}
		
		var cve string
		for _, alias := range vuln.Aliases {
			if strings.HasPrefix(alias, "CVE-") {
				cve = alias
				break
			}
		}
		
		check := SecurityCheck{
			Type:        "vulnerability",
			Severity:    severity,
			Title:       vuln.Summary,
			Description: vuln.Details,
			Package:     packageName,
			Version:     version,
			CVE:         cve,
			CVSS:        cvss,
			References:  references,
			Metadata: map[string]interface{}{
				"source": "OSV",
				"id":     vuln.ID,
			},
		}
		
		checks = append(checks, check)
	}
	
	return checks, nil
}

func (ss *SecurityService) queryGitHubAdvisories(ctx context.Context, packageName, version, packageType string) ([]SecurityCheck, error) {
	// Simplified GitHub Security Advisories check
	// In practice, you'd use the GitHub GraphQL API
	
	var checks []SecurityCheck
	
	// This is a placeholder implementation
	// Real implementation would query GitHub's security advisory database
	
	return checks, nil
}

func (ss *SecurityService) checkMaliciousPatterns(ctx context.Context, packageName, version, packageType string) ([]SecurityCheck, error) {
	var checks []SecurityCheck
	
	// Check for suspicious package names
	if ss.isSuspiciousName(packageName) {
		checks = append(checks, SecurityCheck{
			Type:        "malicious_pattern",
			Severity:    "high",
			Title:       "Suspicious package name detected",
			Description: fmt.Sprintf("Package name '%s' matches known malicious patterns", packageName),
			Package:     packageName,
			Version:     version,
			Metadata: map[string]interface{}{
				"pattern_type": "suspicious_name",
			},
		})
	}
	
	// Check for typosquatting
	if ss.isTyposquatting(packageName, packageType) {
		checks = append(checks, SecurityCheck{
			Type:        "typosquatting",
			Severity:    "medium",
			Title:       "Potential typosquatting detected",
			Description: fmt.Sprintf("Package '%s' may be typosquatting a popular package", packageName),
			Package:     packageName,
			Version:     version,
			Metadata: map[string]interface{}{
				"pattern_type": "typosquatting",
			},
		})
	}
	
	return checks, nil
}

func (ss *SecurityService) checkPackageReputation(ctx context.Context, packageName, packageType string) ([]SecurityCheck, error) {
	var checks []SecurityCheck
	
	// Check package age, download count, maintainer reputation, etc.
	// This is a simplified implementation
	
	// For demonstration, check if package is very new (potentially suspicious)
	if ss.isNewPackage(ctx, packageName, packageType) {
		checks = append(checks, SecurityCheck{
			Type:        "reputation",
			Severity:    "low",
			Title:       "New package detected",
			Description: fmt.Sprintf("Package '%s' is relatively new and may need additional scrutiny", packageName),
			Package:     packageName,
			Metadata: map[string]interface{}{
				"reputation_type": "new_package",
			},
		})
	}
	
	return checks, nil
}

func (ss *SecurityService) isBlacklisted(packageName, packageType string) bool {
	// Check against blacklist in database
	var count int64
	ss.db.Model(&models.SecurityRule{}).
		Where("rule_type = ? AND package_name = ? AND package_type = ? AND action = ?", 
			"blacklist", packageName, packageType, "deny").
		Count(&count)
	
	return count > 0
}

func (ss *SecurityService) isWhitelisted(packageName, packageType string) bool {
	// Check against whitelist in database
	var count int64
	ss.db.Model(&models.SecurityRule{}).
		Where("rule_type = ? AND package_name = ? AND package_type = ? AND action = ?", 
			"whitelist", packageName, packageType, "allow").
		Count(&count)
	
	return count > 0
}

func (ss *SecurityService) storeIntegrityCheck(check *IntegrityCheck) error {
	// Store integrity check result in database
	record := &models.SecurityCheck{
		Type:        "integrity",
		PackageName: check.Package,
		Version:     check.Version,
		Severity:    ss.getIntegritySeverity(check),
		Status:      ss.getIntegrityStatus(check),
		Details:     ss.serializeIntegrityCheck(check),
		CheckedAt:   check.Timestamp,
	}
	
	return ss.db.Create(record).Error
}

func (ss *SecurityService) storeSecurityCheck(check *SecurityCheck) error {
	// Store security check result in database
	details, _ := json.Marshal(check)
	
	record := &models.SecurityCheck{
		Type:        check.Type,
		PackageName: check.Package,
		Version:     check.Version,
		Severity:    check.Severity,
		Status:      "detected",
		Details:     string(details),
		CheckedAt:   time.Now(),
	}
	
	return ss.db.Create(record).Error
}

// Helper methods

func (ss *SecurityService) getOSVEcosystem(packageType string) string {
	switch packageType {
	case "npm":
		return "npm"
	case "pip":
		return "PyPI"
	case "maven":
		return "Maven"
	default:
		return packageType
	}
}

func (ss *SecurityService) mapSeverity(score string) string {
	// Map CVSS scores to severity levels
	if strings.Contains(score, "CRITICAL") || strings.Contains(score, "9.") || strings.Contains(score, "10.") {
		return "critical"
	} else if strings.Contains(score, "HIGH") || strings.Contains(score, "7.") || strings.Contains(score, "8.") {
		return "high"
	} else if strings.Contains(score, "MEDIUM") || strings.Contains(score, "4.") || strings.Contains(score, "5.") || strings.Contains(score, "6.") {
		return "medium"
	}
	return "low"
}

func (ss *SecurityService) parseCVSS(score string) (float64, error) {
	// Parse CVSS score from string
	// This is a simplified implementation
	return 5.0, nil
}

func (ss *SecurityService) isSuspiciousName(packageName string) bool {
	// Check for suspicious patterns in package names
	suspiciousPatterns := []string{
		"test", "temp", "debug", "hack", "exploit", "malware",
	}
	
	lowerName := strings.ToLower(packageName)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerName, pattern) {
			return true
		}
	}
	
	return false
}

func (ss *SecurityService) isTyposquatting(packageName, packageType string) bool {
	// Check for potential typosquatting against popular packages
	// This is a simplified implementation
	
	popularPackages := map[string][]string{
		"npm": {"react", "lodash", "express", "axios", "moment"},
		"pip": {"requests", "numpy", "pandas", "django", "flask"},
		"maven": {"junit", "spring-boot", "jackson", "slf4j", "guava"},
	}
	
	if popular, exists := popularPackages[packageType]; exists {
		for _, pop := range popular {
			if ss.isTyposquattingCandidate(packageName, pop) {
				return true
			}
		}
	}
	
	return false
}

func (ss *SecurityService) isTyposquattingCandidate(candidate, target string) bool {
	// Simple Levenshtein distance check
	if len(candidate) == len(target) {
		diff := 0
		for i := 0; i < len(candidate); i++ {
			if candidate[i] != target[i] {
				diff++
			}
		}
		return diff == 1 // Only one character different
	}
	
	return false
}

func (ss *SecurityService) isNewPackage(ctx context.Context, packageName, packageType string) bool {
	// Check if package was published recently (within last 30 days)
	// This would require querying the package registry
	return false // Simplified implementation
}

func (ss *SecurityService) updateFromSource(ctx context.Context, source string) error {
	// Update vulnerability database from external source
	// This is a placeholder implementation
	logger.Debug("Updating vulnerability database from: %s", source)
	return nil
}

func (ss *SecurityService) getIntegritySeverity(check *IntegrityCheck) string {
	if !check.Verified {
		return "high"
	}
	if !check.TrustedSource {
		return "medium"
	}
	return "low"
}

func (ss *SecurityService) getIntegrityStatus(check *IntegrityCheck) string {
	if check.Verified {
		return "verified"
	}
	return "failed"
}

func (ss *SecurityService) serializeIntegrityCheck(check *IntegrityCheck) string {
	data, _ := json.Marshal(check)
	return string(data)
}
