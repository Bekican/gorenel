package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string //configfile
	Verbose bool   //debug
)

// ana komut
var rootCmd = &cobra.Command{
	Use:   "gorenel",
	Short: "Localhost'u internete açan modern tunnel tool",
	Long: `Gorenel - Go ile yazılmış,Ngrok alternatifi bir tunnel sistemi
	Tek komutla localhost'unuzu global bir URL'e dönüştürün.
Network engineering prensipleriyle yazılmış, production-ready bir araç.

Örnekler:
  gorenel start --port 3000
  gorenel start --port 8080 --subdomain my-app
  gorenel start --config tunnel.yaml`,
	Version: "1.0.0",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config dosyası (default : $HOME/.gorenel.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Detaylı log çıktısı")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Home directory bulunamadı : ", err)
			os.Exit(1)
		}
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".gorenel")
	}
	viper.SetEnvPrefix("GORENEL")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil && Verbose {
		fmt.Println("Config dosyası :", viper.ConfigFileUsed())
	}
}
