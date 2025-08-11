package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/config"
	"github.com/8tcapital/ai-dep-manager/internal/security"
	"github.com/spf13/cobra"
)

// securityCmd represents the security command
var securityCmd = &cobra.Command{
	Use:   "security",
	Short: "Manage security settings and credentials",
	Long: `Manage security-related features including:
- Package integrity verification
- Vulnerability scanning
- Whitelist/blacklist management
- Credential storage for private registries

Examples:
  ai-dep-manager security scan                      # Scan for vulnerabilities
  ai-dep-manager security whitelist add react      # Add package to whitelist
  ai-dep-manager security credential add npm-token # Add registry credential`,
}

// Security scan command
var securityScanCmd = &cobra.Command{
	Use:   "scan [package@version]",
	Short: "Scan packages for security vulnerabilities",
	Long: `Scan packages for known security vulnerabilities and integrity issues.
Can scan specific packages or all dependencies in a project.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSecurityScan(cmd, args)
	},
}

// Whitelist management commands
var whitelistCmd = &cobra.Command{
	Use:   "whitelist",
	Short: "Manage package whitelist",
	Long:  "Add, remove, or list packages in the security whitelist",
}

var whitelistAddCmd = &cobra.Command{
	Use:   "add <package> [package-type]",
	Short: "Add package to whitelist",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		packageType := "npm"
		if len(args) > 1 {
			packageType = args[1]
		}
		return runWhitelistAdd(args[0], packageType)
	},
}

var whitelistRemoveCmd = &cobra.Command{
	Use:   "remove <package> [package-type]",
	Short: "Remove package from whitelist",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		packageType := "npm"
		if len(args) > 1 {
			packageType = args[1]
		}
		return runWhitelistRemove(args[0], packageType)
	},
}

var whitelistListCmd = &cobra.Command{
	Use:   "list",
	Short: "List whitelisted packages",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWhitelistList()
	},
}

// Blacklist management commands
var blacklistCmd = &cobra.Command{
	Use:   "blacklist",
	Short: "Manage package blacklist",
	Long:  "Add, remove, or list packages in the security blacklist",
}

var blacklistAddCmd = &cobra.Command{
	Use:   "add <package> [package-type] [reason]",
	Short: "Add package to blacklist",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		packageType := "npm"
		reason := "Security concern"
		
		if len(args) > 1 {
			packageType = args[1]
		}
		if len(args) > 2 {
			reason = strings.Join(args[2:], " ")
		}
		
		return runBlacklistAdd(args[0], packageType, reason)
	},
}

var blacklistRemoveCmd = &cobra.Command{
	Use:   "remove <package> [package-type]",
	Short: "Remove package from blacklist",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		packageType := "npm"
		if len(args) > 1 {
			packageType = args[1]
		}
		return runBlacklistRemove(args[0], packageType)
	},
}

var blacklistListCmd = &cobra.Command{
	Use:   "list",
	Short: "List blacklisted packages",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBlacklistList()
	},
}

// Credential management commands
var credentialCmd = &cobra.Command{
	Use:   "credential",
	Short: "Manage registry credentials",
	Long:  "Securely store and manage credentials for private package registries",
}

var credentialAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add new registry credential",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCredentialAdd(args[0])
	},
}

var credentialListCmd = &cobra.Command{
	Use:   "list",
	Short: "List stored credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCredentialList()
	},
}

var credentialRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove stored credential",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCredentialRemove(args[0])
	},
}

var (
	securityScanProject     string
	securityScanProjectID   uint
	securityScanPackageType string
	securityScanSeverity    string
)

func runSecurityScan(cmd *cobra.Command, args []string) error {
	cfg := config.GetConfig()
	securityService := security.NewSecurityService(cfg)
	
	fmt.Println("🔍 Security Vulnerability Scan")
	fmt.Println(strings.Repeat("=", 50))
	
	if len(args) > 0 {
		// Scan specific package
		packageSpec := args[0]
		parts := strings.Split(packageSpec, "@")
		if len(parts) != 2 {
			return fmt.Errorf("invalid package specification. Use format: package@version")
		}
		
		packageName := parts[0]
		version := parts[1]
		packageType := securityScanPackageType
		
		return scanPackage(cmd, securityService, packageName, version, packageType)
	}
	
	// Scan project dependencies
	if securityScanProject != "" || securityScanProjectID != 0 {
		return scanProjectDependencies(securityService)
	}
	
	fmt.Println("❌ Please specify a package or project to scan")
	return nil
}

func scanPackage(cmd *cobra.Command, securityService *security.SecurityService, packageName, version, packageType string) error {
	fmt.Printf("🔍 Scanning %s@%s (%s)\n\n", packageName, version, packageType)
	
	// Verify package integrity
	fmt.Println("🔐 Checking package integrity...")
	integrityCheck, err := securityService.VerifyPackageIntegrity(cmd.Context(), packageName, version, packageType)
	if err != nil {
		fmt.Printf("⚠️  Integrity check failed: %v\n", err)
	} else {
		if integrityCheck.Verified {
			fmt.Println("✅ Package integrity verified")
		} else {
			fmt.Println("❌ Package integrity verification failed")
		}
		
		fmt.Printf("   Trusted source: %t\n", integrityCheck.TrustedSource)
		fmt.Printf("   Expected hashes: %d\n", len(integrityCheck.ExpectedHashes))
		fmt.Printf("   Actual hashes: %d\n", len(integrityCheck.ActualHashes))
	}
	
	// Scan for vulnerabilities
	fmt.Println("\n🛡️  Scanning for vulnerabilities...")
	vulnerabilities, err := securityService.ScanForVulnerabilities(cmd.Context(), packageName, version, packageType)
	if err != nil {
		fmt.Printf("⚠️  Vulnerability scan failed: %v\n", err)
		return nil
	}
	
	if len(vulnerabilities) == 0 {
		fmt.Println("✅ No vulnerabilities found")
		return nil
	}
	
	fmt.Printf("🚨 Found %d security issue(s):\n\n", len(vulnerabilities))
	
	for i, vuln := range vulnerabilities {
		fmt.Printf("%d. %s (%s)\n", i+1, vuln.Title, strings.ToUpper(vuln.Severity))
		fmt.Printf("   Type: %s\n", vuln.Type)
		fmt.Printf("   Description: %s\n", vuln.Description)
		
		if vuln.CVE != "" {
			fmt.Printf("   CVE: %s\n", vuln.CVE)
		}
		
		if vuln.CVSS > 0 {
			fmt.Printf("   CVSS Score: %.1f\n", vuln.CVSS)
		}
		
		if len(vuln.References) > 0 {
			fmt.Printf("   References:\n")
			for _, ref := range vuln.References {
				fmt.Printf("     - %s\n", ref)
			}
		}
		
		fmt.Println()
	}
	
	return nil
}

func scanProjectDependencies(securityService *security.SecurityService) error {
	fmt.Println("🔍 Scanning project dependencies...")
	fmt.Println("⚠️  Project scanning not yet implemented")
	return nil
}

func runWhitelistAdd(packageName, packageType string) error {
	fmt.Printf("➕ Adding %s (%s) to whitelist\n", packageName, packageType)
	
	// TODO: Implement whitelist management
	fmt.Println("⚠️  Whitelist management not yet implemented")
	return nil
}

func runWhitelistRemove(packageName, packageType string) error {
	fmt.Printf("➖ Removing %s (%s) from whitelist\n", packageName, packageType)
	
	// TODO: Implement whitelist management
	fmt.Println("⚠️  Whitelist management not yet implemented")
	return nil
}

func runWhitelistList() error {
	fmt.Println("📋 Whitelisted Packages")
	fmt.Println(strings.Repeat("=", 30))
	
	// TODO: Implement whitelist listing
	fmt.Println("⚠️  Whitelist management not yet implemented")
	return nil
}

func runBlacklistAdd(packageName, packageType, reason string) error {
	fmt.Printf("🚫 Adding %s (%s) to blacklist\n", packageName, packageType)
	fmt.Printf("   Reason: %s\n", reason)
	
	// TODO: Implement blacklist management
	fmt.Println("⚠️  Blacklist management not yet implemented")
	return nil
}

func runBlacklistRemove(packageName, packageType string) error {
	fmt.Printf("✅ Removing %s (%s) from blacklist\n", packageName, packageType)
	
	// TODO: Implement blacklist management
	fmt.Println("⚠️  Blacklist management not yet implemented")
	return nil
}

func runBlacklistList() error {
	fmt.Println("🚫 Blacklisted Packages")
	fmt.Println(strings.Repeat("=", 30))
	
	// TODO: Implement blacklist listing
	fmt.Println("⚠️  Blacklist management not yet implemented")
	return nil
}

func runCredentialAdd(name string) error {
	cfg := config.GetConfig()
	credService, err := security.NewCredentialService(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize credential service: %w", err)
	}
	
	fmt.Printf("🔐 Adding credential: %s\n", name)
	
	// Interactive credential creation
	cred := &security.StoredCredential{
		Name: name,
	}
	
	// Get credential type
	fmt.Print("Credential type (token/basic_auth/ssh_key): ")
	var credType string
	fmt.Scanln(&credType)
	
	switch credType {
	case "token":
		cred.Type = security.CredentialTypeToken
		fmt.Print("Registry URL: ")
		fmt.Scanln(&cred.Registry)
		fmt.Print("Token: ")
		fmt.Scanln(&cred.Token)
		
	case "basic_auth":
		cred.Type = security.CredentialTypeBasicAuth
		fmt.Print("Registry URL: ")
		fmt.Scanln(&cred.Registry)
		fmt.Print("Username: ")
		fmt.Scanln(&cred.Username)
		fmt.Print("Password: ")
		fmt.Scanln(&cred.Password)
		
	case "ssh_key":
		cred.Type = security.CredentialTypeSSHKey
		fmt.Print("Registry URL: ")
		fmt.Scanln(&cred.Registry)
		fmt.Print("Private key path: ")
		var keyPath string
		fmt.Scanln(&keyPath)
		
		keyData, err := os.ReadFile(keyPath)
		if err != nil {
			return fmt.Errorf("failed to read private key: %w", err)
		}
		cred.PrivateKey = string(keyData)
		
	default:
		return fmt.Errorf("unsupported credential type: %s", credType)
	}
	
	// Store credential
	if err := credService.StoreCredential(cred); err != nil {
		return fmt.Errorf("failed to store credential: %w", err)
	}
	
	fmt.Println("✅ Credential stored successfully")
	return nil
}

func runCredentialList() error {
	cfg := config.GetConfig()
	credService, err := security.NewCredentialService(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize credential service: %w", err)
	}
	
	fmt.Println("🔐 Stored Credentials")
	fmt.Println(strings.Repeat("=", 40))
	
	creds, err := credService.ListCredentials()
	if err != nil {
		return fmt.Errorf("failed to list credentials: %w", err)
	}
	
	if len(creds) == 0 {
		fmt.Println("No credentials stored")
		return nil
	}
	
	for _, cred := range creds {
		fmt.Printf("\n📋 %s\n", cred.Name)
		fmt.Printf("   Type: %s\n", cred.Type)
		fmt.Printf("   Registry: %s\n", cred.Registry)
		fmt.Printf("   Username: %s\n", cred.Username)
		fmt.Printf("   Created: %s\n", cred.CreatedAt.Format("2006-01-02 15:04:05"))
		
		if cred.ExpiresAt != nil {
			fmt.Printf("   Expires: %s\n", cred.ExpiresAt.Format("2006-01-02 15:04:05"))
			
			if cred.ExpiresAt.Before(time.Now()) {
				fmt.Printf("   Status: ❌ EXPIRED\n")
			} else if cred.ExpiresAt.Before(time.Now().Add(7*24*time.Hour)) {
				fmt.Printf("   Status: ⚠️  EXPIRES SOON\n")
			} else {
				fmt.Printf("   Status: ✅ VALID\n")
			}
		} else {
			fmt.Printf("   Status: ✅ VALID (no expiration)\n")
		}
	}
	
	return nil
}

func runCredentialRemove(name string) error {
	cfg := config.GetConfig()
	credService, err := security.NewCredentialService(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize credential service: %w", err)
	}
	
	fmt.Printf("🗑️  Removing credential: %s\n", name)
	
	// Confirm deletion
	fmt.Print("Are you sure? (y/N): ")
	var confirm string
	fmt.Scanln(&confirm)
	
	if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
		fmt.Println("❌ Deletion cancelled")
		return nil
	}
	
	if err := credService.DeleteCredential(name); err != nil {
		return fmt.Errorf("failed to delete credential: %w", err)
	}
	
	fmt.Println("✅ Credential deleted successfully")
	return nil
}

func init() {
	rootCmd.AddCommand(securityCmd)
	
	// Add subcommands
	securityCmd.AddCommand(securityScanCmd)
	securityCmd.AddCommand(whitelistCmd)
	securityCmd.AddCommand(blacklistCmd)
	securityCmd.AddCommand(credentialCmd)
	
	// Whitelist subcommands
	whitelistCmd.AddCommand(whitelistAddCmd)
	whitelistCmd.AddCommand(whitelistRemoveCmd)
	whitelistCmd.AddCommand(whitelistListCmd)
	
	// Blacklist subcommands
	blacklistCmd.AddCommand(blacklistAddCmd)
	blacklistCmd.AddCommand(blacklistRemoveCmd)
	blacklistCmd.AddCommand(blacklistListCmd)
	
	// Credential subcommands
	credentialCmd.AddCommand(credentialAddCmd)
	credentialCmd.AddCommand(credentialListCmd)
	credentialCmd.AddCommand(credentialRemoveCmd)
	
	// Add flags
	securityScanCmd.Flags().StringVar(&securityScanProject, "project", "", "Project name to scan")
	securityScanCmd.Flags().UintVar(&securityScanProjectID, "project-id", 0, "Project ID to scan")
	securityScanCmd.Flags().StringVar(&securityScanPackageType, "type", "npm", "Package type (npm, pip, maven, gradle)")
	securityScanCmd.Flags().StringVar(&securityScanSeverity, "severity", "", "Filter by severity (low, medium, high, critical)")
}
