package cmd

import (
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var httpCmd = &cobra.Command{
	Use:   "http [port]",
	Short: "Expose a local HTTP server to the internet",
	Long: `Expose a local HTTP port to the internet. This is a shortcut for 'start --type http'.

Examples:
  gorenel http 3000
  gorenel http 8080 --subdomain my-site
  gorenel http 4000 --auth "admin:secret"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		portVal, err := strconv.Atoi(args[0])
		if err == nil {
			localPort = portVal
			viper.Set("port", portVal)
		}
		tunnelType = "http"
		viper.Set("type", "http")

		runStart(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(httpCmd)

	httpCmd.Flags().StringVarP(&serverAddr, "server", "s", "", "Server address")
	httpCmd.Flags().StringVar(&customSubdomain, "subdomain", "", "Custom subdomain")
	httpCmd.Flags().StringVarP(&apiKey, "key", "k", "", "API key")
	httpCmd.Flags().StringVarP(&customDomain, "domain", "d", "", "Custom domain")
	httpCmd.Flags().StringVar(&authCredentials, "auth", "", "Password protect your tunnel (form: 'user:pass')")
	httpCmd.Flags().StringArrayVar(&ipWhitelist, "ip-whitelist", []string{}, "Allowed client IP/CIDR (repeatable)")
	httpCmd.Flags().BoolVar(&corsEnabled, "cors", false, "Enable built-in Smart CORS handling")
	httpCmd.Flags().StringVar(&preferRegion, "region", "", "Prefer a Fly.io region for tunnel control-plane")
	httpCmd.Flags().BoolVar(&stableSubdomain, "stable", false, "Use stable subdomain based on project+port")

	viper.BindPFlag("server", httpCmd.Flags().Lookup("server"))
	viper.BindPFlag("api_key", httpCmd.Flags().Lookup("key"))
	viper.BindPFlag("domain", httpCmd.Flags().Lookup("domain"))
	viper.BindPFlag("region", httpCmd.Flags().Lookup("region"))
	viper.BindPFlag("stable", httpCmd.Flags().Lookup("stable"))
}
