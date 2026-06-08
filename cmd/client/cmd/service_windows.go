//go:build windows

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

var serviceInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Gorenel as a Windows Service",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := persistServiceConfigFromFlags(cmd); err != nil {
			return err
		}
		exe, err := exePath()
		if err != nil {
			return err
		}
		m, err := mgr.Connect()
		if err != nil {
			return err
		}
		defer m.Disconnect()
		if s, _ := m.OpenService(serviceName); s != nil {
			_ = s.Close()
			return fmt.Errorf("service already exists: %s", serviceName)
		}
		s, err := m.CreateService(serviceName, exe, mgr.Config{
			DisplayName: "Gorenel Tunnel",
			StartType:   mgr.StartAutomatic,
		}, "service", "run")
		if err != nil {
			return err
		}
		_ = s.Close()
		fmt.Println("Installed service:", serviceName)
		return nil
	},
}

var serviceUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove the Windows Service",
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := mgr.Connect()
		if err != nil {
			return err
		}
		defer m.Disconnect()
		s, err := m.OpenService(serviceName)
		if err != nil {
			return err
		}
		defer s.Close()
		_ = s.Delete()
		fmt.Println("Uninstalled service:", serviceName)
		return nil
	},
}

var serviceStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Windows Service",
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := mgr.Connect()
		if err != nil {
			return err
		}
		defer m.Disconnect()
		s, err := m.OpenService(serviceName)
		if err != nil {
			return err
		}
		defer s.Close()
		return s.Start()
	},
}

var serviceStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Windows Service",
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := mgr.Connect()
		if err != nil {
			return err
		}
		defer m.Disconnect()
		s, err := m.OpenService(serviceName)
		if err != nil {
			return err
		}
		defer s.Close()
		_, err = s.Control(svc.Stop)
		return err
	},
}

var serviceStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show Windows Service status",
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := mgr.Connect()
		if err != nil {
			return err
		}
		defer m.Disconnect()
		s, err := m.OpenService(serviceName)
		if err != nil {
			return err
		}
		defer s.Close()
		st, err := s.Query()
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, "state:", st.State)
		return nil
	},
}

func init() {
	serviceCmd.AddCommand(serviceInstallCmd, serviceUninstallCmd, serviceStartCmd, serviceStopCmd, serviceStatusCmd)
	serviceInstallCmd.Flags().String("server", "", "Override server URL for service")
	serviceInstallCmd.Flags().Int("port", 0, "Override local port for service")
	serviceInstallCmd.Flags().String("type", "", "Override tunnel type for service (http,tcp,udp)")
	serviceInstallCmd.Flags().String("domain", "", "Custom domain")
	serviceInstallCmd.Flags().String("subdomain", "", "Reserved subdomain")
	serviceInstallCmd.Flags().String("key-auth", "", "X-TOKEN key auth")
	serviceInstallCmd.Flags().StringArray("ip-whitelist", []string{}, "Allowed IP/CIDR")
}
