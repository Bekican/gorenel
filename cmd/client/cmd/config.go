package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Kalici istemci ayarlari",
}

var configGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Aktif config degerlerini yazdir",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("config_file: %s\n", activeConfigPath())
		fmt.Printf("server: %s\n", viper.GetString("server"))
		fmt.Printf("port: %d\n", viper.GetInt("port"))
		fmt.Printf("type: %s\n", viper.GetString("type"))
		if viper.GetString("api_key") != "" {
			fmt.Println("api_key: ****")
		} else {
			fmt.Println("api_key: <empty>")
		}
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Config anahtarini kalici olarak set et",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := strings.TrimSpace(args[0])
		val := strings.TrimSpace(args[1])
		switch key {
		case "api_key", "server", "type":
			viper.Set(key, val)
		case "port":
			p, err := strconv.Atoi(val)
			if err != nil || p <= 0 || p > 65535 {
				return fmt.Errorf("port gecersiz: %s", val)
			}
			viper.Set("port", p)
		default:
			return fmt.Errorf("desteklenmeyen key: %s (api_key|server|port|type)", key)
		}
		return saveConfig()
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Interaktif olarak config dosyasi olustur",
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)
		server := promptWithDefault(reader, "Server URL", viper.GetString("server"))
		portStr := promptWithDefault(reader, "Varsayilan local port", fmt.Sprintf("%d", viper.GetInt("port")))
		tType := promptWithDefault(reader, "Tunnel tipi (http/tcp/udp)", viper.GetString("type"))
		api := promptWithDefault(reader, "API key", viper.GetString("api_key"))

		p, err := strconv.Atoi(portStr)
		if err != nil || p <= 0 || p > 65535 {
			return fmt.Errorf("port gecersiz: %s", portStr)
		}

		viper.Set("server", server)
		viper.Set("port", p)
		viper.Set("type", tType)
		viper.Set("api_key", api)
		if err := saveConfig(); err != nil {
			return err
		}
		fmt.Printf("Config kaydedildi: %s\n", activeConfigPath())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configGetCmd, configSetCmd, configInitCmd)
}

func promptWithDefault(reader *bufio.Reader, label, def string) string {
	if def != "" {
		fmt.Printf("%s [%s]: ", label, def)
	} else {
		fmt.Printf("%s: ", label)
	}
	input, _ := reader.ReadString('\n')
	val := strings.TrimSpace(input)
	if val == "" {
		return def
	}
	return val
}

func activeConfigPath() string {
	if used := viper.ConfigFileUsed(); used != "" {
		return used
	}
	if cfgFile != "" {
		return cfgFile
	}
	return defaultConfigPath()
}

func saveConfig() error {
	p := activeConfigPath()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(p); os.IsNotExist(err) {
		if err := viper.SafeWriteConfigAs(p); err != nil {
			return err
		}
		return nil
	}
	return viper.WriteConfigAs(p)
}
