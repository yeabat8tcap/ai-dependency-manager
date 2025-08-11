package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/agent"
	"github.com/8tcapital/ai-dep-manager/internal/config"
	"github.com/8tcapital/ai-dep-manager/internal/logger"
)

// Daemon manages the background service lifecycle
type Daemon struct {
	config    *config.Config
	agent     *agent.Agent
	pidFile   string
	logFile   string
	running   bool
}

// DaemonConfig holds daemon-specific configuration
type DaemonConfig struct {
	PidFile     string `json:"pid_file"`
	LogFile     string `json:"log_file"`
	WorkingDir  string `json:"working_dir"`
	User        string `json:"user,omitempty"`
	Group       string `json:"group,omitempty"`
}

// NewDaemon creates a new daemon instance
func NewDaemon(cfg *config.Config) *Daemon {
	daemonConfig := getDaemonConfig(cfg)
	
	return &Daemon{
		config:  cfg,
		pidFile: daemonConfig.PidFile,
		logFile: daemonConfig.LogFile,
	}
}

// Start starts the daemon
func (d *Daemon) Start() error {
	// Check if already running
	if d.IsRunning() {
		return fmt.Errorf("daemon is already running (PID: %d)", d.GetPID())
	}
	
	logger.Info("Starting AI Dependency Manager daemon")
	
	// Create PID file directory if it doesn't exist
	pidDir := filepath.Dir(d.pidFile)
	if err := os.MkdirAll(pidDir, 0755); err != nil {
		return fmt.Errorf("failed to create PID directory: %w", err)
	}
	
	// Write PID file
	pid := os.Getpid()
	if err := d.writePIDFile(pid); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}
	
	// Ensure PID file is cleaned up on exit
	defer d.removePIDFile()
	
	// Create and start agent
	d.agent = agent.NewAgent(d.config)
	if err := d.agent.Start(); err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}
	
	d.running = true
	logger.Info("Daemon started successfully with PID: %d", pid)
	
	// Wait for agent to finish (this blocks)
	d.waitForShutdown()
	
	return nil
}

// Stop stops the daemon
func (d *Daemon) Stop() error {
	pid := d.GetPID()
	if pid == 0 {
		return fmt.Errorf("daemon is not running")
	}
	
	logger.Info("Stopping daemon with PID: %d", pid)
	
	// Send SIGTERM to the process
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}
	
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM: %w", err)
	}
	
	// Wait for process to stop (with timeout)
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-timeout:
			// Force kill if graceful shutdown failed
			logger.Warn("Graceful shutdown timeout, force killing process")
			if err := process.Kill(); err != nil {
				return fmt.Errorf("failed to force kill process: %w", err)
			}
			return nil
			
		case <-ticker.C:
			if !d.IsRunning() {
				logger.Info("Daemon stopped successfully")
				return nil
			}
		}
	}
}

// Restart restarts the daemon
func (d *Daemon) Restart() error {
	logger.Info("Restarting daemon")
	
	if d.IsRunning() {
		if err := d.Stop(); err != nil {
			return fmt.Errorf("failed to stop daemon: %w", err)
		}
	}
	
	// Wait a moment before restarting
	time.Sleep(2 * time.Second)
	
	return d.Start()
}

// IsRunning checks if the daemon is currently running
func (d *Daemon) IsRunning() bool {
	pid := d.GetPID()
	if pid == 0 {
		return false
	}
	
	// Check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	
	// Send signal 0 to check if process is alive
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// GetPID returns the daemon PID from the PID file
func (d *Daemon) GetPID() int {
	data, err := os.ReadFile(d.pidFile)
	if err != nil {
		return 0
	}
	
	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return 0
	}
	
	return pid
}

// GetStatus returns the daemon status
func (d *Daemon) GetStatus() *DaemonStatus {
	status := &DaemonStatus{
		Running: d.IsRunning(),
		PidFile: d.pidFile,
		LogFile: d.logFile,
	}
	
	if status.Running {
		status.PID = d.GetPID()
		if d.agent != nil {
			status.AgentStats = d.agent.GetStats()
		}
	}
	
	return status
}

// DaemonStatus represents the current daemon status
type DaemonStatus struct {
	Running    bool               `json:"running"`
	PID        int                `json:"pid,omitempty"`
	PidFile    string             `json:"pid_file"`
	LogFile    string             `json:"log_file"`
	AgentStats *agent.AgentStats  `json:"agent_stats,omitempty"`
}

// Private methods

func (d *Daemon) writePIDFile(pid int) error {
	return os.WriteFile(d.pidFile, []byte(strconv.Itoa(pid)), 0644)
}

func (d *Daemon) removePIDFile() {
	if err := os.Remove(d.pidFile); err != nil && !os.IsNotExist(err) {
		logger.Warn("Failed to remove PID file: %v", err)
	}
}

func (d *Daemon) waitForShutdown() {
	// This would typically set up signal handlers and wait
	// For now, we'll simulate waiting for the agent
	select {}
}

func getDaemonConfig(cfg *config.Config) *DaemonConfig {
	dataDir := cfg.DataDir
	if dataDir == "" {
		dataDir = filepath.Join(os.Getenv("HOME"), ".ai-dep-manager")
	}
	
	return &DaemonConfig{
		PidFile:    filepath.Join(dataDir, "ai-dep-manager.pid"),
		LogFile:    filepath.Join(dataDir, "ai-dep-manager.log"),
		WorkingDir: dataDir,
	}
}

// Service management functions for different platforms

// InstallService installs the daemon as a system service
func (d *Daemon) InstallService() error {
	// This would install the service based on the platform
	// Linux: systemd service file
	// macOS: launchd plist
	// Windows: Windows service
	
	logger.Info("Installing AI Dependency Manager as system service")
	
	// For now, create a basic systemd service file on Linux
	return d.createSystemdService()
}

// UninstallService removes the daemon from system services
func (d *Daemon) UninstallService() error {
	logger.Info("Uninstalling AI Dependency Manager system service")
	
	// This would remove the service based on the platform
	return d.removeSystemdService()
}

func (d *Daemon) createSystemdService() error {
	serviceContent := `[Unit]
Description=AI Dependency Manager
After=network.target

[Service]
Type=simple
User=%s
WorkingDirectory=%s
ExecStart=%s agent start --daemon
ExecStop=%s agent stop
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
`
	
	// Get current user and executable path
	user := os.Getenv("USER")
	if user == "" {
		user = "root"
	}
	
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	
	workingDir := filepath.Dir(d.pidFile)
	
	serviceFile := fmt.Sprintf(serviceContent, user, workingDir, execPath, execPath)
	
	// Write service file (requires sudo)
	servicePath := "/etc/systemd/system/ai-dep-manager.service"
	logger.Info("Creating systemd service file: %s", servicePath)
	logger.Warn("Note: This requires sudo privileges")
	
	// In a real implementation, this would use sudo or prompt for privileges
	return os.WriteFile(servicePath, []byte(serviceFile), 0644)
}

func (d *Daemon) removeSystemdService() error {
	servicePath := "/etc/systemd/system/ai-dep-manager.service"
	
	if err := os.Remove(servicePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove service file: %w", err)
	}
	
	logger.Info("Systemd service file removed")
	return nil
}
