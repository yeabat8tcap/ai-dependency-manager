package cmd

import (
	"fmt"
	"net/http"

	"github.com/8tcapital/ai-dep-manager/internal/web"
	"github.com/spf13/cobra"
)

var serveSimpleCmd = &cobra.Command{
	Use:   "serve-simple",
	Short: "Start the unified web server (simplified version)",
	Long: `Start the unified web server that serves both the REST API and Angular frontend.
This is a simplified version for testing the unified full-stack application.`,
	RunE: runServeSimple,
}

func init() {
	rootCmd.AddCommand(serveSimpleCmd)
	serveSimpleCmd.Flags().StringP("port", "p", "8080", "Port to run the web server on")
	serveSimpleCmd.Flags().StringP("host", "H", "localhost", "Host to bind the web server to")
}

func runServeSimple(cmd *cobra.Command, args []string) error {
	port, _ := cmd.Flags().GetString("port")
	host, _ := cmd.Flags().GetString("host")
	
	fmt.Printf("🚀 Starting AI Dependency Manager Web Server (Simplified)\n")
	fmt.Printf("🌐 Server: http://%s:%s\n", host, port)
	fmt.Printf("📊 Frontend: http://%s:%s\n", host, port)
	fmt.Printf("🔗 API: http://%s:%s/api\n", host, port)
	fmt.Printf("✨ Unified Full-Stack Application Running!\n\n")

	// Create web server
	server := web.NewServer()

	// Start server
	addr := fmt.Sprintf("%s:%s", host, port)
	fmt.Printf("🎯 Server listening on %s\n", addr)
	
	if err := http.ListenAndServe(addr, server.Handler()); err != nil {
		return fmt.Errorf("failed to start web server: %w", err)
	}

	return nil
}
