package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to Gorenel tunnel (alias for 'start')",
	Long: `Opens the selected local port to the internet and returns a public URL or port.
This is a shortcut for the 'start' command.

Examples:
  gorenel connect
  gorenel connect --port 3000
  gorenel connect --key GK_... --port 8080 --type tcp`,
	Run: runStart,
}

func init() {
	rootCmd.AddCommand(connectCmd)

	connectCmd.Flags().StringVarP(&serverAddr, "server", "s", "", "Server address (default: gorenel.site)")
	connectCmd.Flags().IntVarP(&localPort, "port", "p", 3000, "Local port")
	connectCmd.Flags().StringVar(&customSubdomain, "subdomain", "", "Custom subdomain")
	connectCmd.Flags().StringVarP(&apiKey, "key", "k", "", "API key (authentication)")
	connectCmd.Flags().StringVar(&apiKey, "api-key", "", "API key (authentication)")
	connectCmd.Flags().StringVarP(&customDomain, "domain", "d", "", "Custom domain")
	connectCmd.Flags().StringVarP(&tunnelType, "type", "t", "http", "Tunnel type (http, tcp, udp)")
	connectCmd.Flags().IntVarP(&remotePort, "remote-port", "r", 0, "Requested remote port")

	viper.BindPFlag("server", connectCmd.Flags().Lookup("server"))
	viper.BindPFlag("port", connectCmd.Flags().Lookup("port"))
	viper.BindPFlag("api_key", connectCmd.Flags().Lookup("key"))
	viper.BindPFlag("domain", connectCmd.Flags().Lookup("domain"))
	viper.BindPFlag("type", connectCmd.Flags().Lookup("type"))
}
