package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to Gorenel tunnel (alias for 'start')",
	Long: `Belirtilen local port'u internete ac¸ar ve size public bir URL veya Port verir.
'start' komutunun kisayoludur.

Ornekler:
  gorenel connect
  gorenel connect --port 3000
  gorenel connect --key GK_... --port 8080 --type tcp`,
	Run: runStart,
}

func init() {
	rootCmd.AddCommand(connectCmd)

	connectCmd.Flags().StringVarP(&serverAddr, "server", "s", "", "Server adresi (default: gorenel.site)")
	connectCmd.Flags().IntVarP(&localPort, "port", "p", 3000, "Local port numarasi")
	connectCmd.Flags().StringVar(&customSubdomain, "subdomain", "", "Ozel subdomain")
	connectCmd.Flags().StringVarP(&apiKey, "key", "k", "", "API key (authentication icin)")
	connectCmd.Flags().StringVar(&apiKey, "api-key", "", "API key (authentication icin)")
	connectCmd.Flags().StringVarP(&customDomain, "domain", "d", "", "Ozel alan adi")
	connectCmd.Flags().StringVarP(&tunnelType, "type", "t", "http", "Tunnel tipi (http, tcp, udp)")
	connectCmd.Flags().IntVarP(&remotePort, "remote-port", "r", 0, "Istenen uzak port")

	viper.BindPFlag("server", connectCmd.Flags().Lookup("server"))
	viper.BindPFlag("port", connectCmd.Flags().Lookup("port"))
	viper.BindPFlag("api_key", connectCmd.Flags().Lookup("key"))
	viper.BindPFlag("domain", connectCmd.Flags().Lookup("domain"))
	viper.BindPFlag("type", connectCmd.Flags().Lookup("type"))
}
