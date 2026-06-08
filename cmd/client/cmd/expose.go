package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var exposeCmd = &cobra.Command{
	Use:   "expose [port]",
	Short: "Expose a local port to the internet (shortcut for 'start')",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			// If port is provided as first argument, override flag
			viper.Set("port", args[0])
		}
		runStart(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(exposeCmd)

	exposeCmd.Flags().StringVarP(&serverAddr, "server", "s", "", "Server address")
	exposeCmd.Flags().IntVarP(&localPort, "port", "p", 3000, "Local port")
	exposeCmd.Flags().StringVar(&customSubdomain, "subdomain", "", "Custom subdomain")
	exposeCmd.Flags().StringVarP(&apiKey, "key", "k", "", "API key")
	exposeCmd.Flags().StringVarP(&customDomain, "domain", "d", "", "Custom domain")
	exposeCmd.Flags().StringVarP(&tunnelType, "type", "t", "http", "Tunnel type")
	exposeCmd.Flags().StringVar(&authCredentials, "auth", "", "Password protect your tunnel (form: 'user:pass')")

	viper.BindPFlag("server", exposeCmd.Flags().Lookup("server"))
	viper.BindPFlag("port", exposeCmd.Flags().Lookup("port"))
	viper.BindPFlag("api_key", exposeCmd.Flags().Lookup("key"))
}
