package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var shareCmd = &cobra.Command{
	Use:   "share [port]",
	Short: "Share a local port with the world (shortcut for 'start')",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			viper.Set("port", args[0])
		}
		runStart(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(shareCmd)

	shareCmd.Flags().StringVarP(&serverAddr, "server", "s", "", "Server address")
	shareCmd.Flags().IntVarP(&localPort, "port", "p", 3000, "Local port")
	shareCmd.Flags().StringVar(&customSubdomain, "subdomain", "", "Custom subdomain")
	shareCmd.Flags().StringVarP(&apiKey, "key", "k", "", "API key")
	shareCmd.Flags().StringVarP(&customDomain, "domain", "d", "", "Custom domain")
	shareCmd.Flags().StringVarP(&tunnelType, "type", "t", "http", "Tunnel type")
	shareCmd.Flags().StringVar(&authCredentials, "auth", "", "Password protect your tunnel (form: 'user:pass')")

	viper.BindPFlag("server", shareCmd.Flags().Lookup("server"))
	viper.BindPFlag("port", shareCmd.Flags().Lookup("port"))
	viper.BindPFlag("api_key", shareCmd.Flags().Lookup("key"))
}
