package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/agent"
	"github.com/8tcapital/ai-dep-manager/internal/config"
	"github.com/8tcapital/ai-dep-manager/internal/logger"
	"github.com/spf13/cobra"
)

// agentCmd represents the agent command
var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage the background dependency agent",
	Long: `Manage the AI Dependency Manager background agent that automatically
monitors and updates dependencies. The agent runs continuously in the background,
performing scheduled scans and applying safe updates based on your configuration.

Examples:
  ai-dep-manager agent start                    # Start the background agent
  ai-dep-manager agent stop                     # Stop the background agent
  ai-dep-manager agent status                   # Show agent status
  ai-dep-manager agent restart                  # Restart the agent`,
}

// agentStartCmd starts the background agent
var agentStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the background agent",
	Long: `Start the AI Dependency Manager background agent. The agent will run
continuously, performing scheduled dependency scans and applying updates
based on your configuration.

The agent will:
- Monitor all enabled projects
- Perform regular dependency scans
- Apply safe updates automatically (if configured)
- Log all activities for audit purposes`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAgentStart(cmd, args)
	},
}

// agentStopCmd stops the background agent
var agentStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the background agent",
	Long: `Stop the AI Dependency Manager background agent gracefully.
This will stop all scheduled operations and shut down the agent cleanly.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAgentStop(cmd, args)
	},
}

// agentStatusCmd shows agent status
var agentStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show agent status and statistics",
	Long: `Display the current status of the AI Dependency Manager background agent,
including runtime statistics, recent activity, and configuration details.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAgentStatus(cmd, args)
	},
}

// agentRestartCmd restarts the background agent
var agentRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the background agent",
	Long: `Restart the AI Dependency Manager background agent. This will stop
the current agent instance and start a new one with the current configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAgentRestart(cmd, args)
	},
}

var (
	agentDaemon     bool
	agentForeground bool
)

func runAgentStart(cmd *cobra.Command, args []string) error {
	cfg := config.GetConfig()
	
	// Create and start the agent
	agentInstance := agent.NewAgent(cfg)
	
	fmt.Println("🚀 Starting AI Dependency Manager Agent...")
	
	if err := agentInstance.Start(); err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}
	
	fmt.Println("✅ Agent started successfully")
	
	// If running in foreground mode, wait for signals
	if agentForeground || !agentDaemon {
		fmt.Println("📊 Agent running in foreground mode. Press Ctrl+C to stop.")
		
		// Display initial status
		displayAgentStatus(agentInstance)
		
		// Set up signal handling
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		
		// Start status update ticker
		statusTicker := time.NewTicker(30 * time.Second)
		defer statusTicker.Stop()
		
		// Wait for signal or status updates
		for {
			select {
			case sig := <-sigChan:
				fmt.Printf("\n🛑 Received signal: %s\n", sig)
				fmt.Println("🔄 Stopping agent gracefully...")
				
				if err := agentInstance.Stop(); err != nil {
					logger.Error("Error stopping agent: %v", err)
					return err
				}
				
				fmt.Println("✅ Agent stopped successfully")
				return nil
				
			case <-statusTicker.C:
				displayAgentStatus(agentInstance)
			}
		}
	}
	
	fmt.Println("🔄 Agent started in daemon mode")
	return nil
}

func runAgentStop(cmd *cobra.Command, args []string) error {
	// In a real implementation, this would connect to a running daemon
	// For now, we'll show what the stop command would do
	
	fmt.Println("🛑 Stopping AI Dependency Manager Agent...")
	
	// TODO: Implement daemon communication to stop running agent
	// This would typically involve:
	// 1. Finding the running agent process (PID file)
	// 2. Sending a graceful shutdown signal
	// 3. Waiting for confirmation
	
	fmt.Println("⚠️  Note: Daemon mode not fully implemented yet")
	fmt.Println("💡 If running in foreground mode, use Ctrl+C to stop")
	
	return nil
}

func runAgentStatus(cmd *cobra.Command, args []string) error {
	cfg := config.GetConfig()
	
	// Create agent instance to check status
	agentInstance := agent.NewAgent(cfg)
	
	fmt.Println("📊 AI Dependency Manager Agent Status")
	fmt.Println(strings.Repeat("=", 50))
	
	if agentInstance.IsRunning() {
		fmt.Println("🟢 Status: Running")
		displayAgentStatus(agentInstance)
	} else {
		fmt.Println("🔴 Status: Stopped")
		fmt.Println("💡 Use 'ai-dep-manager agent start' to start the agent")
	}
	
	// Show configuration
	fmt.Println("\n⚙️  Agent Configuration:")
	fmt.Printf("   Scan Interval: %s\n", cfg.Agent.ScanInterval)
	fmt.Printf("   Auto Update Level: %s\n", cfg.Agent.AutoUpdateLevel)
	fmt.Printf("   Notification Mode: %s\n", cfg.Agent.NotificationMode)
	fmt.Printf("   Max Concurrency: %d\n", cfg.Agent.MaxConcurrency)
	
	return nil
}

func runAgentRestart(cmd *cobra.Command, args []string) error {
	cfg := config.GetConfig()
	agentInstance := agent.NewAgent(cfg)
	
	fmt.Println("🔄 Restarting AI Dependency Manager Agent...")
	
	if err := agentInstance.Restart(); err != nil {
		return fmt.Errorf("failed to restart agent: %w", err)
	}
	
	fmt.Println("✅ Agent restarted successfully")
	
	// Show status after restart
	displayAgentStatus(agentInstance)
	
	return nil
}

func displayAgentStatus(agentInstance *agent.Agent) {
	stats := agentInstance.GetStats()
	
	fmt.Printf("\n📈 Runtime Statistics:\n")
	fmt.Printf("   Uptime: %s\n", stats.Uptime.Round(time.Second))
	fmt.Printf("   Projects Monitored: %d\n", stats.ProjectsMonitored)
	fmt.Printf("   Total Scans: %d\n", stats.TotalScans)
	fmt.Printf("   Total Updates: %d\n", stats.TotalUpdates)
	fmt.Printf("   Total Errors: %d\n", stats.TotalErrors)
	
	if stats.LastScanTime != nil {
		fmt.Printf("   Last Scan: %s\n", stats.LastScanTime.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Printf("   Last Scan: Never\n")
	}
	
	if stats.LastUpdateTime != nil {
		fmt.Printf("   Last Update: %s\n", stats.LastUpdateTime.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Printf("   Last Update: Never\n")
	}
	
	fmt.Printf("   Pending Updates: %d\n", stats.PendingUpdates)
	fmt.Printf("   Security Updates: %d\n", stats.SecurityUpdates)
	
	// Health indicator
	if stats.TotalErrors == 0 {
		fmt.Println("   Health: 🟢 Healthy")
	} else if float64(stats.TotalErrors)/float64(stats.TotalScans+stats.TotalUpdates) < 0.1 {
		fmt.Println("   Health: 🟡 Minor Issues")
	} else {
		fmt.Println("   Health: 🔴 Needs Attention")
	}
}

func init() {
	rootCmd.AddCommand(agentCmd)
	
	// Add subcommands
	agentCmd.AddCommand(agentStartCmd)
	agentCmd.AddCommand(agentStopCmd)
	agentCmd.AddCommand(agentStatusCmd)
	agentCmd.AddCommand(agentRestartCmd)
	
	// Add flags
	agentStartCmd.Flags().BoolVar(&agentDaemon, "daemon", false, "Run agent as daemon in background")
	agentStartCmd.Flags().BoolVar(&agentForeground, "foreground", false, "Run agent in foreground with status updates")
	
	// Set default to foreground if no daemon support
	agentStartCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if !agentDaemon && !agentForeground {
			agentForeground = true
		}
		return nil
	}
}
