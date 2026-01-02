package cmd

import (
	"fmt"

	"github.com/FmTod/ghost-backup/internal/config"
	svc "github.com/FmTod/ghost-backup/internal/service"
	"github.com/kardianos/service"
	"github.com/spf13/cobra"
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage the ghost-backup service",
	Long:  `Manage the ghost-backup background service (install, uninstall, start, stop, status).`,
}

var serviceInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the user service",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Installing ghost-backup user service...")
		if err := svc.InstallService(); err != nil {
			return err
		}
		fmt.Println("✓ Service installed successfully")
		fmt.Println("\nTo start the service: ghost-backup service start")
		return nil
	},
}

var serviceUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the user service",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Uninstalling ghost-backup user service...")
		if err := svc.UninstallService(); err != nil {
			return err
		}
		fmt.Println("✓ Service uninstalled successfully")
		return nil
	},
}

var serviceStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the user service",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Starting ghost-backup user service...")
		if err := svc.StartService(); err != nil {
			return err
		}
		fmt.Println("✓ Service started successfully")
		return nil
	},
}

var serviceStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the user service",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Stopping ghost-backup user service...")
		if err := svc.StopService(); err != nil {
			return err
		}
		fmt.Println("✓ Service stopped successfully")
		return nil
	},
}

var serviceRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the user service",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Restarting ghost-backup user service...")
		if err := svc.RestartService(); err != nil {
			return err
		}
		fmt.Println("✓ Service restarted successfully")
		return nil
	},
}

var serviceStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check service status",
	RunE: func(cmd *cobra.Command, args []string) error {
		status, err := svc.GetServiceStatus()
		if err != nil {
			return err
		}

		fmt.Printf("Service Status: %s\n", getStatusString(status))

		// Show registry info
		registry, err := config.LoadRegistry()
		if err != nil {
			return fmt.Errorf("failed to load registry: %w", err)
		}

		repos := registry.GetRepositories()
		fmt.Printf("\nMonitored Repositories: %d\n", len(repos))
		for _, repo := range repos {
			fmt.Printf("  - %s\n", repo)
		}

		// Show log file location
		logPath, err := svc.GetLogFilePath()
		if err == nil {
			fmt.Printf("\nLog file: %s\n", logPath)
		}

		return nil
	},
}

var serviceRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the service in foreground (for debugging)",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Running ghost-backup service in foreground...")
		fmt.Println("Press Ctrl+C to stop")
		return svc.RunService()
	},
}

func init() {
	rootCmd.AddCommand(serviceCmd)

	serviceCmd.AddCommand(serviceInstallCmd)
	serviceCmd.AddCommand(serviceUninstallCmd)
	serviceCmd.AddCommand(serviceStartCmd)
	serviceCmd.AddCommand(serviceStopCmd)
	serviceCmd.AddCommand(serviceRestartCmd)
	serviceCmd.AddCommand(serviceStatusCmd)
	serviceCmd.AddCommand(serviceRunCmd)
}

func getStatusString(status service.Status) string {
	switch status {
	case service.StatusRunning:
		return "Running"
	case service.StatusStopped:
		return "Stopped"
	case service.StatusUnknown:
		return "Unknown"
	default:
		return fmt.Sprintf("Status(%d)", status)
	}
}
