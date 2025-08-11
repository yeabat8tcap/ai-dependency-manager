package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/config"
	"github.com/8tcapital/ai-dep-manager/internal/database"
	"github.com/8tcapital/ai-dep-manager/internal/logger"
	"github.com/8tcapital/ai-dep-manager/internal/models"
	"golang.org/x/crypto/pbkdf2"
	"gorm.io/gorm"
)

// CredentialService handles secure credential storage and retrieval
type CredentialService struct {
	config     *config.Config
	db         *gorm.DB
	masterKey  []byte
	gcm        cipher.AEAD
}

// CredentialType represents the type of credential
type CredentialType string

const (
	CredentialTypeToken     CredentialType = "token"
	CredentialTypeBasicAuth CredentialType = "basic_auth"
	CredentialTypeSSHKey    CredentialType = "ssh_key"
)

// StoredCredential represents a credential for external services
type StoredCredential struct {
	ID         uint           `json:"id"`
	Name       string         `json:"name"`
	Type       CredentialType `json:"type"`
	Registry   string         `json:"registry"`
	Username   string         `json:"username,omitempty"`
	Password   string         `json:"-"` // Never expose in JSON
	Token      string         `json:"-"` // Never expose in JSON
	PrivateKey string         `json:"-"` // Never expose in JSON
	PublicKey  string         `json:"public_key,omitempty"`
	ExpiresAt  *time.Time     `json:"expires_at,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// NewCredentialService creates a new credential service
func NewCredentialService(cfg *config.Config) (*CredentialService, error) {
	cs := &CredentialService{
		config: cfg,
		db:     database.GetDB(),
	}
	
	// Initialize encryption
	if err := cs.initializeEncryption(); err != nil {
		return nil, fmt.Errorf("failed to initialize encryption: %w", err)
	}
	
	return cs, nil
}

// StoreCredential securely stores a credential
func (cs *CredentialService) StoreCredential(cred *StoredCredential) error {
	logger.Info("Storing credential: %s", cred.Name)
	
	// Validate credential
	if err := cs.validateCredential(cred); err != nil {
		return fmt.Errorf("credential validation failed: %w", err)
	}
	
	// Check if credential already exists
	var existing models.Credential
	if err := cs.db.Where("name = ?", cred.Name).First(&existing).Error; err == nil {
		return fmt.Errorf("credential with name '%s' already exists", cred.Name)
	}
	
	// Encrypt sensitive fields
	dbCred := &models.Credential{
		Name:      cred.Name,
		Type:      string(cred.Type),
		Registry:  cred.Registry,
		Username:  cred.Username,
		PublicKey: cred.PublicKey,
		ExpiresAt: cred.ExpiresAt,
	}
	
	// Encrypt password if provided
	if cred.Password != "" {
		encryptedPassword, err := cs.encrypt(cred.Password)
		if err != nil {
			return fmt.Errorf("failed to encrypt password: %w", err)
		}
		dbCred.Password = encryptedPassword
	}
	
	// Encrypt token if provided
	if cred.Token != "" {
		encryptedToken, err := cs.encrypt(cred.Token)
		if err != nil {
			return fmt.Errorf("failed to encrypt token: %w", err)
		}
		dbCred.Token = encryptedToken
	}
	
	// Encrypt private key if provided
	if cred.PrivateKey != "" {
		encryptedPrivateKey, err := cs.encrypt(cred.PrivateKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt private key: %w", err)
		}
		dbCred.PrivateKey = encryptedPrivateKey
	}
	
	// Store in database
	if err := cs.db.Create(dbCred).Error; err != nil {
		return fmt.Errorf("failed to store credential: %w", err)
	}
	
	logger.Info("Credential stored successfully: %s", cred.Name)
	return nil
}

// GetCredential retrieves and decrypts a credential
func (cs *CredentialService) GetCredential(name string) (*StoredCredential, error) {
	logger.Debug("Retrieving credential: %s", name)
	
	var dbCred models.Credential
	if err := cs.db.Where("name = ?", name).First(&dbCred).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("credential '%s' not found", name)
		}
		return nil, fmt.Errorf("failed to retrieve credential: %w", err)
	}
	
	// Check if credential is expired
	if dbCred.ExpiresAt != nil && dbCred.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("credential '%s' has expired", name)
	}
	
	cred := &StoredCredential{
		ID:        dbCred.ID,
		Name:      dbCred.Name,
		Type:      CredentialType(dbCred.Type),
		Registry:  dbCred.Registry,
		Username:  dbCred.Username,
		PublicKey: dbCred.PublicKey,
		ExpiresAt: dbCred.ExpiresAt,
		CreatedAt: dbCred.CreatedAt,
		UpdatedAt: dbCred.UpdatedAt,
	}
	
	// Decrypt password if present
	if dbCred.Password != "" {
		password, err := cs.decrypt(dbCred.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt password: %w", err)
		}
		cred.Password = password
	}
	
	// Decrypt token if present
	if dbCred.Token != "" {
		token, err := cs.decrypt(dbCred.Token)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt token: %w", err)
		}
		cred.Token = token
	}
	
	// Decrypt private key if present
	if dbCred.PrivateKey != "" {
		privateKey, err := cs.decrypt(dbCred.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt private key: %w", err)
		}
		cred.PrivateKey = privateKey
	}
	
	logger.Debug("Credential retrieved successfully: %s", name)
	return cred, nil
}

// ListCredentials returns a list of stored credentials (without sensitive data)
func (cs *CredentialService) ListCredentials() ([]*StoredCredential, error) {
	var dbCreds []models.Credential
	if err := cs.db.Find(&dbCreds).Error; err != nil {
		return nil, fmt.Errorf("failed to list credentials: %w", err)
	}
	
	var creds []*StoredCredential
	for _, dbCred := range dbCreds {
		cred := &StoredCredential{
			ID:        dbCred.ID,
			Name:      dbCred.Name,
			Type:      CredentialType(dbCred.Type),
			Registry:  dbCred.Registry,
			Username:  dbCred.Username,
			PublicKey: dbCred.PublicKey,
			ExpiresAt: dbCred.ExpiresAt,
			CreatedAt: dbCred.CreatedAt,
			UpdatedAt: dbCred.UpdatedAt,
		}
		creds = append(creds, cred)
	}
	
	return creds, nil
}

// UpdateCredential updates an existing credential
func (cs *CredentialService) UpdateCredential(name string, updates *StoredCredential) error {
	logger.Info("Updating credential: %s", name)
	
	var dbCred models.Credential
	if err := cs.db.Where("name = ?", name).First(&dbCred).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("credential '%s' not found", name)
		}
		return fmt.Errorf("failed to find credential: %w", err)
	}
	
	// Update non-sensitive fields
	if updates.Registry != "" {
		dbCred.Registry = updates.Registry
	}
	if updates.Username != "" {
		dbCred.Username = updates.Username
	}
	if updates.PublicKey != "" {
		dbCred.PublicKey = updates.PublicKey
	}
	if updates.ExpiresAt != nil {
		dbCred.ExpiresAt = updates.ExpiresAt
	}
	
	// Update encrypted fields if provided
	if updates.Password != "" {
		encryptedPassword, err := cs.encrypt(updates.Password)
		if err != nil {
			return fmt.Errorf("failed to encrypt password: %w", err)
		}
		dbCred.Password = encryptedPassword
	}
	
	if updates.Token != "" {
		encryptedToken, err := cs.encrypt(updates.Token)
		if err != nil {
			return fmt.Errorf("failed to encrypt token: %w", err)
		}
		dbCred.Token = encryptedToken
	}
	
	if updates.PrivateKey != "" {
		encryptedPrivateKey, err := cs.encrypt(updates.PrivateKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt private key: %w", err)
		}
		dbCred.PrivateKey = encryptedPrivateKey
	}
	
	// Save updates
	if err := cs.db.Save(&dbCred).Error; err != nil {
		return fmt.Errorf("failed to update credential: %w", err)
	}
	
	logger.Info("Credential updated successfully: %s", name)
	return nil
}

// DeleteCredential removes a credential
func (cs *CredentialService) DeleteCredential(name string) error {
	logger.Info("Deleting credential: %s", name)
	
	result := cs.db.Where("name = ?", name).Delete(&models.Credential{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete credential: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("credential '%s' not found", name)
	}
	
	logger.Info("Credential deleted successfully: %s", name)
	return nil
}

// GetCredentialForRegistry retrieves credentials for a specific registry
func (cs *CredentialService) GetCredentialForRegistry(registry string) (*StoredCredential, error) {
	var dbCred models.Credential
	if err := cs.db.Where("registry = ?", registry).First(&dbCred).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no credentials found for registry '%s'", registry)
		}
		return nil, fmt.Errorf("failed to retrieve credential: %w", err)
	}
	
	return cs.GetCredential(dbCred.Name)
}

// RotateCredentials rotates expiring credentials
func (cs *CredentialService) RotateCredentials() error {
	logger.Info("Checking for expiring credentials")
	
	// Find credentials expiring within 7 days
	expiryThreshold := time.Now().Add(7 * 24 * time.Hour)
	
	var expiringCreds []models.Credential
	if err := cs.db.Where("expires_at IS NOT NULL AND expires_at < ?", expiryThreshold).Find(&expiringCreds).Error; err != nil {
		return fmt.Errorf("failed to find expiring credentials: %w", err)
	}
	
	if len(expiringCreds) == 0 {
		logger.Info("No expiring credentials found")
		return nil
	}
	
	logger.Warn("Found %d expiring credentials", len(expiringCreds))
	
	// For now, just log the expiring credentials
	// In a real implementation, you might:
	// 1. Send notifications
	// 2. Attempt automatic renewal
	// 3. Mark credentials as requiring attention
	
	for _, cred := range expiringCreds {
		logger.Warn("Credential '%s' expires at %s", cred.Name, cred.ExpiresAt.Format("2006-01-02 15:04:05"))
	}
	
	return nil
}

// Private methods

func (cs *CredentialService) initializeEncryption() error {
	// Get or create master key
	masterKey, err := cs.getMasterKey()
	if err != nil {
		return fmt.Errorf("failed to get master key: %w", err)
	}
	
	cs.masterKey = masterKey
	
	// Create AES cipher
	block, err := aes.NewCipher(cs.masterKey)
	if err != nil {
		return fmt.Errorf("failed to create AES cipher: %w", err)
	}
	
	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}
	
	cs.gcm = gcm
	return nil
}

func (cs *CredentialService) getMasterKey() ([]byte, error) {
	// Try to get key from environment variable
	if keyStr := os.Getenv("AI_DEP_MANAGER_MASTER_KEY"); keyStr != "" {
		key, err := base64.StdEncoding.DecodeString(keyStr)
		if err != nil {
			return nil, fmt.Errorf("failed to decode master key from environment: %w", err)
		}
		if len(key) == 32 {
			return key, nil
		}
	}
	
	// Try to get key from config
	if cs.config.Security.MasterKey != "" {
		key, err := base64.StdEncoding.DecodeString(cs.config.Security.MasterKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decode master key from config: %w", err)
		}
		if len(key) == 32 {
			return key, nil
		}
	}
	
	// Generate a new key if none exists
	logger.Warn("No master key found, generating new key")
	return cs.generateMasterKey()
}

func (cs *CredentialService) generateMasterKey() ([]byte, error) {
	// Generate a random 32-byte key
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}
	
	// Derive key using PBKDF2 for additional security
	salt := []byte("ai-dep-manager-salt") // In production, use a random salt
	derivedKey := pbkdf2.Key(key, salt, 10000, 32, sha256.New)
	
	// Log warning about key storage
	keyStr := base64.StdEncoding.EncodeToString(derivedKey)
	logger.Warn("Generated new master key. Store this securely:")
	logger.Warn("AI_DEP_MANAGER_MASTER_KEY=%s", keyStr)
	logger.Warn("Without this key, stored credentials cannot be decrypted!")
	
	return derivedKey, nil
}

func (cs *CredentialService) encrypt(plaintext string) (string, error) {
	// Generate a random nonce
	nonce := make([]byte, cs.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	
	// Encrypt the plaintext
	ciphertext := cs.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	
	// Encode to base64 for storage
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (cs *CredentialService) decrypt(ciphertext string) (string, error) {
	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}
	
	// Extract nonce and ciphertext
	nonceSize := cs.gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	
	nonce, ciphertext_bytes := data[:nonceSize], data[nonceSize:]
	
	// Decrypt
	plaintext, err := cs.gcm.Open(nil, nonce, ciphertext_bytes, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}
	
	return string(plaintext), nil
}

func (cs *CredentialService) validateCredential(cred *StoredCredential) error {
	if cred.Name == "" {
		return fmt.Errorf("credential name is required")
	}
	
	if cred.Registry == "" {
		return fmt.Errorf("registry is required")
	}
	
	switch cred.Type {
	case CredentialTypeToken:
		if cred.Token == "" {
			return fmt.Errorf("token is required for token credentials")
		}
	case CredentialTypeBasicAuth:
		if cred.Username == "" || cred.Password == "" {
			return fmt.Errorf("username and password are required for basic auth credentials")
		}
	case CredentialTypeSSHKey:
		if cred.PrivateKey == "" {
			return fmt.Errorf("private key is required for SSH key credentials")
		}
	default:
		return fmt.Errorf("unsupported credential type: %s", cred.Type)
	}
	
	return nil
}
