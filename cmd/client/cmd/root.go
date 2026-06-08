package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	Verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "gorenel",
	Short: "Localhost'u internete acan modern tunnel tool",
	Long: `Gorenel - Go ile yazilmis, Ngrok alternatifi bir tunnel sistemi
Tek komutla localhost'unuzu global bir URL'e donusturun.

Ornekler:
  gorenel login
  gorenel connect
  gorenel connect --port 4000
  gorenel config init`,
	Version: "1.2.5",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config dosyasi (default: OS user config dir)")
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Detayli log cikisi")
}

func initConfig() {
	viper.SetEnvPrefix("GORENEL")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("server", "wss://gorenel.site/tunnel/connect")
	viper.SetDefault("port", 3000)
	viper.SetDefault("type", "http")

	configPath := cfgFile
	if configPath == "" {
		configPath = defaultConfigPath()
	}
	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err == nil && Verbose {
		fmt.Println("Config dosyasi:", viper.ConfigFileUsed())
	} else if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok && Verbose {
			fmt.Println("Config dosyasi okunamadi:", err)
		}
	}
}

func defaultConfigPath() string {
	if userCfgDir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(userCfgDir, "gorenel", "config.yaml")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".gorenel.yaml"
	}
	return filepath.Join(home, ".gorenel.yaml")
}
