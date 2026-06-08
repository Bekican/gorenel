package cmd

import (
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var tcpCmd = &cobra.Command{
	Use:   "tcp [port]",
	Short: "Expose a local raw TCP connection to the internet",
	Long: `Expose a local raw TCP connection (like SSH, Databases, or VNC) to the internet. This is a shortcut for 'start --type tcp'.

Examples:
  gorenel tcp 22
  gorenel tcp 5432 --remote-port 15432
  gorenel tcp 3306 --key "YOUR_API_KEY"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		portVal, err := strconv.Atoi(args[0])
		if err == nil {
			localPort = portVal
			viper.Set("port", portVal)
		}
		tunnelType = "tcp"
		viper.Set("type", "tcp")

		runStart(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(tcpCmd)

	tcpCmd.Flags().StringVarP(&serverAddr, "server", "s", "", "Server address")
	tcpCmd.Flags().IntVarP(&remotePort, "remote-port", "r", 0, "Requested remote port on the relay server")
	tcpCmd.Flags().StringVarP(&apiKey, "key", "k", "", "API key")
	tcpCmd.Flags().StringArrayVar(&ipWhitelist, "ip-whitelist", []string{}, "Allowed client IP/CIDR (repeatable)")
	tcpCmd.Flags().StringVar(&preferRegion, "region", "", "Prefer a Fly.io region for tunnel control-plane")

	viper.BindPFlag("server", tcpCmd.Flags().Lookup("server"))
	viper.BindPFlag("api_key", tcpCmd.Flags().Lookup("key"))
	viper.BindPFlag("region", tcpCmd.Flags().Lookup("region"))
}
