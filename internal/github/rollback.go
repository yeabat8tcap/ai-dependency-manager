package github

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// RollbackManager handles rollback operations for failed patch applications
type RollbackManager struct {
	rollbackPoints map[string]*RollbackPoint
}

// NewRollbackManager creates a new rollback manager
func NewRollbackManager() *RollbackManager {
	return &RollbackManager{
		rollbackPoints: make(map[string]*RollbackPoint),
	}
}

// RollbackPoint represents a point in time that can be rolled back to
type RollbackPoint struct {
	ID          string                 `json:"id"`
	Repository  string                 `json:"repository"`
	Branch      string                 `json:"branch"`
	CommitHash  string                 `json:"commit_hash"`
	CreatedAt   time.Time              `json:"created_at"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
	BackupPath  string                 `json:"backup_path,omitempty"`
	FileStates  []*FileState          `json:"file_states"`
}

// FileState represents the state of a file at a rollback point
type FileState struct {
	Path         string    `json:"path"`
	Hash         string    `json:"hash"`
	Size         int64     `json:"size"`
	ModTime      time.Time `json:"mod_time"`
	Permissions  os.FileMode `json:"permissions"`
	Content      []byte    `json:"content,omitempty"`
	Exists       bool      `json:"exists"`
}

// RollbackResult represents the result of a rollback operation
type RollbackResult struct {
	Success       bool          `json:"success"`
	RollbackPoint *RollbackPoint `json:"rollback_point"`
	FilesRestored int           `json:"files_restored"`
	FilesSkipped  int           `json:"files_skipped"`
	Errors        []string      `json:"errors"`
	Duration      time.Duration `json:"duration"`
	Summary       string        `json:"summary"`
}

// CreateRollbackPoint creates a new rollback point
func (rm *RollbackManager) CreateRollbackPoint(repository, branch string) (*RollbackPoint, error) {
	// Generate unique ID
	id := fmt.Sprintf("rollback_%d", time.Now().Unix())
	
	// Get current commit hash
	commitHash, err := rm.getCurrentCommitHash(repository)
	if err != nil {
		return nil, fmt.Errorf("failed to get current commit hash: %w", err)
	}
	
	// Create backup directory
	backupPath := filepath.Join(repository, ".aidep_backups", id)
	err = os.MkdirAll(backupPath, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}
	
	// Capture file states
	fileStates, err := rm.captureFileStates(repository)
	if err != nil {
		return nil, fmt.Errorf("failed to capture file states: %w", err)
	}
	
	// Create rollback point
	rollbackPoint := &RollbackPoint{
		ID:          id,
		Repository:  repository,
		Branch:      branch,
		CommitHash:  commitHash,
		CreatedAt:   time.Now(),
		Description: fmt.Sprintf("Rollback point for %s at %s", branch, commitHash[:8]),
		Metadata:    make(map[string]interface{}),
		BackupPath:  backupPath,
		FileStates:  fileStates,
	}
	
	// Store rollback point
	rm.rollbackPoints[id] = rollbackPoint
	
	// Create physical backup
	err = rm.createPhysicalBackup(repository, backupPath, fileStates)
	if err != nil {
		return nil, fmt.Errorf("failed to create physical backup: %w", err)
	}
	
	return rollbackPoint, nil
}

// Rollback performs a rollback to the specified rollback point
func (rm *RollbackManager) Rollback(rollbackPoint *RollbackPoint) error {
	startTime := time.Now()
	
	result := &RollbackResult{
		RollbackPoint: rollbackPoint,
		Errors:        []string{},
	}
	
	// Validate rollback point
	if rollbackPoint == nil {
		return fmt.Errorf("rollback point is nil")
	}
	
	// Check if repository exists
	if _, err := os.Stat(rollbackPoint.Repository); os.IsNotExist(err) {
		return fmt.Errorf("repository does not exist: %s", rollbackPoint.Repository)
	}
	
	// Perform Git-based rollback if commit hash is available
	if rollbackPoint.CommitHash != "" {
		err := rm.performGitRollback(rollbackPoint)
		if err != nil {
			// Fallback to file-based rollback
			return rm.performFileRollback(rollbackPoint, result)
		}
		result.Success = true
		result.Summary = "Git rollback completed successfully"
	} else {
		// Perform file-based rollback
		return rm.performFileRollback(rollbackPoint, result)
	}
	
	result.Duration = time.Since(startTime)
	return nil
}

// performGitRollback performs a Git-based rollback
func (rm *RollbackManager) performGitRollback(rollbackPoint *RollbackPoint) error {
	// Reset to the specific commit
	cmd := exec.Command("git", "reset", "--hard", rollbackPoint.CommitHash)
	cmd.Dir = rollbackPoint.Repository
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git reset failed: %w, output: %s", err, string(output))
	}
	
	// Clean untracked files
	cmd = exec.Command("git", "clean", "-fd")
	cmd.Dir = rollbackPoint.Repository
	
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clean failed: %w, output: %s", err, string(output))
	}
	
	return nil
}

// performFileRollback performs a file-based rollback
func (rm *RollbackManager) performFileRollback(rollbackPoint *RollbackPoint, result *RollbackResult) error {
	for _, fileState := range rollbackPoint.FileStates {
		err := rm.restoreFile(rollbackPoint, fileState)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to restore %s: %v", fileState.Path, err))
			result.FilesSkipped++
			continue
		}
		result.FilesRestored++
	}
	
	result.Success = result.FilesRestored > 0
	result.Summary = fmt.Sprintf("File rollback completed: %d restored, %d skipped", result.FilesRestored, result.FilesSkipped)
	
	if len(result.Errors) > 0 {
		return fmt.Errorf("rollback completed with errors: %v", result.Errors)
	}
	
	return nil
}

// restoreFile restores a single file from backup
func (rm *RollbackManager) restoreFile(rollbackPoint *RollbackPoint, fileState *FileState) error {
	targetPath := filepath.Join(rollbackPoint.Repository, fileState.Path)
	
	if !fileState.Exists {
		// File should not exist, remove it if it does
		if _, err := os.Stat(targetPath); err == nil {
			return os.Remove(targetPath)
		}
		return nil
	}
	
	// Restore file from backup
	backupFilePath := filepath.Join(rollbackPoint.BackupPath, fileState.Path)
	
	// Read backup content
	content, err := os.ReadFile(backupFilePath)
	if err != nil {
		// Fallback to stored content
		if len(fileState.Content) > 0 {
			content = fileState.Content
		} else {
			return fmt.Errorf("no backup content available for %s", fileState.Path)
		}
	}
	
	// Ensure target directory exists
	targetDir := filepath.Dir(targetPath)
	err = os.MkdirAll(targetDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}
	
	// Write file content
	err = os.WriteFile(targetPath, content, fileState.Permissions)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	// Restore modification time
	err = os.Chtimes(targetPath, fileState.ModTime, fileState.ModTime)
	if err != nil {
		// Non-critical error, log but continue
		return nil
	}
	
	return nil
}

// getCurrentCommitHash gets the current Git commit hash
func (rm *RollbackManager) getCurrentCommitHash(repository string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repository
	
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	
	return strings.TrimSpace(string(output)), nil
}

// captureFileStates captures the current state of all files in the repository
func (rm *RollbackManager) captureFileStates(repository string) ([]*FileState, error) {
	var fileStates []*FileState
	
	err := filepath.Walk(repository, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip backup directories and Git directories
		if strings.Contains(path, ".aidep_backups") || strings.Contains(path, ".git") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		
		// Skip directories
		if info.IsDir() {
			return nil
		}
		
		// Get relative path
		relPath, err := filepath.Rel(repository, path)
		if err != nil {
			return err
		}
		
		// Create file state
		fileState := &FileState{
			Path:        relPath,
			Size:        info.Size(),
			ModTime:     info.ModTime(),
			Permissions: info.Mode(),
			Exists:      true,
		}
		
		// Calculate hash
		hash, err := rm.calculateFileHash(path)
		if err != nil {
			return err
		}
		fileState.Hash = hash
		
		// Store content for small files
		if info.Size() < 1024*1024 { // 1MB threshold
			content, err := os.ReadFile(path)
			if err == nil {
				fileState.Content = content
			}
		}
		
		fileStates = append(fileStates, fileState)
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	return fileStates, nil
}

// createPhysicalBackup creates a physical backup of files
func (rm *RollbackManager) createPhysicalBackup(repository, backupPath string, fileStates []*FileState) error {
	for _, fileState := range fileStates {
		if !fileState.Exists {
			continue
		}
		
		sourcePath := filepath.Join(repository, fileState.Path)
		targetPath := filepath.Join(backupPath, fileState.Path)
		
		// Create target directory
		targetDir := filepath.Dir(targetPath)
		err := os.MkdirAll(targetDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create backup directory for %s: %w", fileState.Path, err)
		}
		
		// Copy file
		err = rm.copyFile(sourcePath, targetPath)
		if err != nil {
			return fmt.Errorf("failed to backup file %s: %w", fileState.Path, err)
		}
	}
	
	return nil
}

// copyFile copies a file from source to destination
func (rm *RollbackManager) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	_, err = destFile.ReadFrom(sourceFile)
	if err != nil {
		return err
	}
	
	// Copy permissions
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}
	
	return os.Chmod(dst, sourceInfo.Mode())
}

// calculateFileHash calculates a simple hash for a file
func (rm *RollbackManager) calculateFileHash(filePath string) (string, error) {
	// Simplified hash calculation - in real implementation would use crypto/sha256
	info, err := os.Stat(filePath)
	if err != nil {
		return "", err
	}
	
	// Use size and modification time as simple hash
	return fmt.Sprintf("%d_%d", info.Size(), info.ModTime().Unix()), nil
}

// ListRollbackPoints lists all available rollback points
func (rm *RollbackManager) ListRollbackPoints() []*RollbackPoint {
	var points []*RollbackPoint
	
	for _, point := range rm.rollbackPoints {
		points = append(points, point)
	}
	
	return points
}

// GetRollbackPoint retrieves a rollback point by ID
func (rm *RollbackManager) GetRollbackPoint(id string) *RollbackPoint {
	return rm.rollbackPoints[id]
}

// DeleteRollbackPoint deletes a rollback point and its backup
func (rm *RollbackManager) DeleteRollbackPoint(id string) error {
	point := rm.rollbackPoints[id]
	if point == nil {
		return fmt.Errorf("rollback point not found: %s", id)
	}
	
	// Remove backup directory
	if point.BackupPath != "" {
		err := os.RemoveAll(point.BackupPath)
		if err != nil {
			return fmt.Errorf("failed to remove backup directory: %w", err)
		}
	}
	
	// Remove from memory
	delete(rm.rollbackPoints, id)
	
	return nil
}

// CleanupOldRollbackPoints removes rollback points older than the specified duration
func (rm *RollbackManager) CleanupOldRollbackPoints(maxAge time.Duration) error {
	cutoff := time.Now().Add(-maxAge)
	
	var toDelete []string
	for id, point := range rm.rollbackPoints {
		if point.CreatedAt.Before(cutoff) {
			toDelete = append(toDelete, id)
		}
	}
	
	for _, id := range toDelete {
		err := rm.DeleteRollbackPoint(id)
		if err != nil {
			return fmt.Errorf("failed to delete rollback point %s: %w", id, err)
		}
	}
	
	return nil
}

// ValidateRollbackPoint validates that a rollback point is still valid
func (rm *RollbackManager) ValidateRollbackPoint(rollbackPoint *RollbackPoint) error {
	// Check if repository exists
	if _, err := os.Stat(rollbackPoint.Repository); os.IsNotExist(err) {
		return fmt.Errorf("repository no longer exists: %s", rollbackPoint.Repository)
	}
	
	// Check if backup exists
	if rollbackPoint.BackupPath != "" {
		if _, err := os.Stat(rollbackPoint.BackupPath); os.IsNotExist(err) {
			return fmt.Errorf("backup directory no longer exists: %s", rollbackPoint.BackupPath)
		}
	}
	
	// Validate Git commit if available
	if rollbackPoint.CommitHash != "" {
		cmd := exec.Command("git", "cat-file", "-e", rollbackPoint.CommitHash)
		cmd.Dir = rollbackPoint.Repository
		
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("commit no longer exists: %s", rollbackPoint.CommitHash)
		}
	}
	
	return nil
}

// GetRollbackStatus gets the status of rollback operations
func (rm *RollbackManager) GetRollbackStatus() *RollbackStatus {
	status := &RollbackStatus{
		TotalRollbackPoints: len(rm.rollbackPoints),
		RollbackPoints:      []*RollbackPointSummary{},
	}
	
	var totalSize int64
	for _, point := range rm.rollbackPoints {
		summary := &RollbackPointSummary{
			ID:          point.ID,
			Repository:  point.Repository,
			Branch:      point.Branch,
			CreatedAt:   point.CreatedAt,
			Description: point.Description,
			FileCount:   len(point.FileStates),
		}
		
		// Calculate backup size
		for _, fileState := range point.FileStates {
			summary.BackupSize += fileState.Size
		}
		totalSize += summary.BackupSize
		
		status.RollbackPoints = append(status.RollbackPoints, summary)
	}
	
	status.TotalBackupSize = totalSize
	return status
}

// RollbackStatus represents the status of rollback operations
type RollbackStatus struct {
	TotalRollbackPoints int                      `json:"total_rollback_points"`
	TotalBackupSize     int64                    `json:"total_backup_size"`
	RollbackPoints      []*RollbackPointSummary `json:"rollback_points"`
}

// RollbackPointSummary provides a summary of a rollback point
type RollbackPointSummary struct {
	ID          string    `json:"id"`
	Repository  string    `json:"repository"`
	Branch      string    `json:"branch"`
	CreatedAt   time.Time `json:"created_at"`
	Description string    `json:"description"`
	FileCount   int       `json:"file_count"`
	BackupSize  int64     `json:"backup_size"`
}


