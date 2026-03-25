//go:build !windows

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

import "github.com/spf13/cobra"

var serviceInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Gorenel as a background service",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := persistServiceConfigFromFlags(cmd); err != nil {
			return err
		}
		exe, err := exePath()
		if err != nil {
			return err
		}
		switch runtime.GOOS {
		case "linux":
			return installSystemdUserUnit(exe, serviceName)
		case "darwin":
			return installLaunchdAgent(exe, serviceName)
		default:
			return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
		}
	},
}

var serviceStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the service",
	RunE: func(cmd *cobra.Command, args []string) error {
		switch runtime.GOOS {
		case "linux":
			return systemctlUser("start", unitName(serviceName))
		case "darwin":
			return launchctl("kickstart", "gui/"+strconv.Itoa(os.Getuid())+"/"+plistLabel(serviceName))
		default:
			return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
		}
	},
}

var serviceStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the service",
	RunE: func(cmd *cobra.Command, args []string) error {
		switch runtime.GOOS {
		case "linux":
			return systemctlUser("stop", unitName(serviceName))
		case "darwin":
			return launchctl("bootout", "gui/"+strconv.Itoa(os.Getuid()), plistLabel(serviceName))
		default:
			return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
		}
	},
}

var serviceStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show service status",
	RunE: func(cmd *cobra.Command, args []string) error {
		switch runtime.GOOS {
		case "linux":
			return systemctlUser("status", unitName(serviceName))
		case "darwin":
			return launchctl("print", "gui/"+strconv.Itoa(os.Getuid())+"/"+plistLabel(serviceName))
		default:
			return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
		}
	},
}

var serviceUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove the background service",
	RunE: func(cmd *cobra.Command, args []string) error {
		switch runtime.GOOS {
		case "linux":
			return systemctlUser("disable", "--now", unitName(serviceName))
		case "darwin":
			return launchctl("bootout", "gui/"+strconv.Itoa(os.Getuid()), plistLabel(serviceName))
		default:
			return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
		}
	},
}

func installSystemdUserUnit(exe, name string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".config", "systemd", "user")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	unitPath := filepath.Join(dir, unitName(name))
	unit := fmt.Sprintf(`[Unit]
Description=Gorenel Tunnel Service
After=network-online.target

[Service]
ExecStart=%s service run
Restart=always
RestartSec=3

[Install]
WantedBy=default.target
`, exe)
	if err := os.WriteFile(unitPath, []byte(unit), 0o644); err != nil {
		return err
	}
	_ = systemctlUser("daemon-reload")
	_ = systemctlUser("enable", "--now", unitName(name))
	return nil
}

func installLaunchdAgent(exe, name string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, "Library", "LaunchAgents")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	label := plistLabel(name)
	plistPath := filepath.Join(dir, label+".plist")
	plist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>Label</key><string>%s</string>
    <key>ProgramArguments</key>
    <array>
      <string>%s</string>
      <string>service</string>
      <string>run</string>
    </array>
    <key>RunAtLoad</key><true/>
    <key>KeepAlive</key><true/>
  </dict>
</plist>
`, label, exe)
	if err := os.WriteFile(plistPath, []byte(plist), 0o644); err != nil {
		return err
	}
	_ = launchctl("bootstrap", "gui/"+strconv.Itoa(os.Getuid()), plistPath)
	return nil
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

