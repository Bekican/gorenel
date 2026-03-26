package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serviceName string

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Always-on tunnel service management",
}

var serviceRunCmd = &cobra.Command{
	Use:    "run",
	Short:  "Internal: run the service loop",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		server := strings.TrimSpace(viper.GetString("service.server"))
		if server == "" {
			server = strings.TrimSpace(viper.GetString("server"))
		}
		localPort := viper.GetInt("service.port")
		if localPort == 0 {
			localPort = viper.GetInt("port")
		}
		tType := strings.ToLower(strings.TrimSpace(viper.GetString("service.type")))
		if tType == "" {
			tType = strings.ToLower(strings.TrimSpace(viper.GetString("type")))
		}
		customDomain := strings.TrimSpace(viper.GetString("service.domain"))
		customSubdomain = strings.TrimSpace(viper.GetString("service.subdomain"))
		keyAuthToken = strings.TrimSpace(viper.GetString("service.key_auth"))
		ipWhitelist = viper.GetStringSlice("service.ip_whitelist")

		apiKey = strings.TrimSpace(viper.GetString("api_key"))
		if apiKey == "" {
			return fmt.Errorf("api_key is missing; run: gorenel config set api_key <KEY>")
		}
		if server == "" {
			server = "wss://gorenel.site/tunnel/connect"
		}
		if tType == "" {
			tType = "http"
		}

		delay := 1 * time.Second
		maxDelay := 30 * time.Second
		for {
			ctx := context.Background()
			err := startTunnel(ctx, server, localPort, customDomain, tType)
			if err != nil {
				fmt.Println("service tunnel error:", err)
				time.Sleep(delay)
				delay *= 2
				if delay > maxDelay {
					delay = maxDelay
				}
				continue
			}
			delay = 1 * time.Second
		}
	},
}

func init() {
	serviceCmd.PersistentFlags().StringVar(&serviceName, "name", "gorenel", "Service name")
	serviceCmd.AddCommand(serviceRunCmd)
	rootCmd.AddCommand(serviceCmd)
}

func systemctlUser(args ...string) error {
	c := exec.Command("systemctl", append([]string{"--user"}, args...)...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func launchctl(args ...string) error {
	c := exec.Command("launchctl", args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func unitName(name string) string   { return name + ".service" }
func plistLabel(name string) string { return "com.gorenel." + name }

func persistServiceConfigFromFlags(cmd *cobra.Command) error {
	reader := bufio.NewReader(os.Stdin)

	getStr := func(flag, label, def string) string {
		if v, _ := cmd.Flags().GetString(flag); strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
		return promptWithDefault(reader, label, def)
	}
	getInt := func(flag, label string, def int) int {
		if v, _ := cmd.Flags().GetInt(flag); v > 0 {
			return v
		}
		val := promptWithDefault(reader, label, fmt.Sprintf("%d", def))
		i, _ := strconv.Atoi(strings.TrimSpace(val))
		if i <= 0 {
			return def
		}
		return i
	}
	getStrArr := func(flag string) []string {
		if v, _ := cmd.Flags().GetStringArray(flag); len(v) > 0 {
			return v
		}
		return nil
	}

	viper.Set("service.enabled", true)
	viper.Set("service.name", serviceName)
	viper.Set("service.server", getStr("server", "Service server", viper.GetString("server")))
	viper.Set("service.port", getInt("port", "Service port", viper.GetInt("port")))
	viper.Set("service.type", getStr("type", "Service type", viper.GetString("type")))
	viper.Set("service.domain", getStr("domain", "Service domain", viper.GetString("domain")))
	viper.Set("service.subdomain", getStr("subdomain", "Service subdomain", ""))
	viper.Set("service.key_auth", getStr("key-auth", "Service key-auth", ""))
	if ips := getStrArr("ip-whitelist"); ips != nil {
		viper.Set("service.ip_whitelist", ips)
	}
	return saveConfig()
}

func exePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	exe, _ = filepath.Abs(exe)
	return exe, nil
}
