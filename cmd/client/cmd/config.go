package cmd

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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
	Args:  validateConfigSetArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		key := normalizeConfigKey(strings.TrimSpace(args[0]))
		val := strings.TrimSpace(args[1])

		switch key {
		case "api_key", "server", "type":
			if key == "api_key" {
				if err := validateAPIKeyFormat(val); err != nil {
					return err
				}
			}
			viper.Set(key, val)
		case "port":
			p, err := strconv.Atoi(val)
			if err != nil || p <= 0 || p > 65535 {
				return fmt.Errorf("gecersiz port: %s (ornek: gorenel config set port 3000)", val)
			}
			viper.Set("port", p)
		default:
			return fmt.Errorf("desteklenmeyen key: %s\nDesteklenen key'ler: api_key, server, port, type", args[0])
		}

		if err := saveConfig(); err != nil {
			return err
		}

		fmt.Printf("Kaydedildi: %s = %s\n", key, renderConfigValue(key, val))
		fmt.Println("Hazir. Tunnel baslatmak icin: gorenel connect --port 3000")
		return nil
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Interaktif olarak config dosyasi olustur",
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Hizli kurulum: sadece API key zorunlu. Diger ayarlar varsayilanlarla gelir.")

		api := promptWithDefault(reader, "API key", viper.GetString("api_key"))
		if err := validateAPIKeyFormat(api); err != nil {
			return err
		}

		portStr := promptWithDefault(reader, "Local port (bos=3000)", "3000")
		p, err := strconv.Atoi(portStr)
		if err != nil || p <= 0 || p > 65535 {
			return fmt.Errorf("gecersiz port: %s", portStr)
		}

		viper.Set("api_key", api)
		viper.Set("port", p)
		if viper.GetString("server") == "" {
			viper.Set("server", "wss://gorenel.site/tunnel/connect")
		}
		if viper.GetString("type") == "" {
			viper.Set("type", "http")
		}

		if err := saveConfig(); err != nil {
			return err
		}
		fmt.Printf("Config kaydedildi: %s\n", activeConfigPath())
		fmt.Println("Hazir. Artik sadece `gorenel connect` yazman yeterli.")
		return nil
	},
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Config degerlerini ve server erisimini test et",
	RunE: func(cmd *cobra.Command, args []string) error {
		api := viper.GetString("api_key")
		if err := validateAPIKeyFormat(api); err != nil {
			return err
		}
		server := viper.GetString("server")
		if server == "" {
			return fmt.Errorf("server bos")
		}

		httpURL := strings.Replace(server, "wss://", "https://", 1)
		httpURL = strings.Replace(httpURL, "ws://", "http://", 1)
		if strings.Contains(httpURL, "/tunnel/connect") {
			httpURL = strings.Replace(httpURL, "/tunnel/connect", "/health", 1)
		} else {
			httpURL = strings.TrimRight(httpURL, "/") + "/health"
		}

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(httpURL)
		if err != nil {
			return fmt.Errorf("server health ulasilamadi: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			return fmt.Errorf("server health basarisiz: %s", resp.Status)
		}

		fmt.Println("Config OK: key formati ve server health basarili")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configGetCmd, configSetCmd, configInitCmd, configValidateCmd)
}

func validateConfigSetArgs(cmd *cobra.Command, args []string) error {
	if len(args) == 2 {
		return nil
	}

	help := "Kullanim: gorenel config set <key> <value>\n" +
		"Ornekler:\n" +
		"  gorenel config set api_key gk_xxx\n" +
		"  gorenel config set port 3000\n" +
		"Not: key adi 'api_key' olmalidir (api key degil)."

	if len(args) == 3 && strings.EqualFold(args[0], "api") && strings.EqualFold(args[1], "key") {
		return fmt.Errorf("'api key' yerine 'api_key' kullanmalisin.\n\n%s", help)
	}

	if len(args) == 0 {
		return fmt.Errorf("Eksik arguman.\n\n%s", help)
	}
	if len(args) == 1 {
		return fmt.Errorf("Value eksik.\n\n%s", help)
	}
	return fmt.Errorf("Fazla arguman girdin (%d).\n\n%s", len(args), help)
}

func normalizeConfigKey(key string) string {
	k := strings.ToLower(strings.TrimSpace(key))
	k = strings.ReplaceAll(k, "-", "_")
	if k == "apikey" {
		return "api_key"
	}
	return k
}

func renderConfigValue(key, val string) string {
	if key == "api_key" {
		return "****"
	}
	return val
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

func validateAPIKeyFormat(api string) error {
	if strings.TrimSpace(api) == "" {
		return fmt.Errorf("api_key bos olamaz")
	}
	if !strings.HasPrefix(api, "gk_") {
		return fmt.Errorf("api_key gecersiz: 'gk_' ile baslamali")
	}
	if len(api) < 12 {
		return fmt.Errorf("api_key cok kisa")
	}
	return nil
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
