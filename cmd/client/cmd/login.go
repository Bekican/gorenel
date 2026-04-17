package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "API anahtari ile oturum acin",
	Long: `Gorenel hesabinizdan aldiginiz API anahtarini (gk_...) kullanarak 
cihazinizin yetkilendirmesini saglar. Bu anahtar guvenli bir sekilde 
yerel konfigurasyona kaydedilir.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		fmt.Println("Gorenel'e hos geldiniz!")
		fmt.Println("Lutfen dashboard'dan aldiginiz API anahtarini girin.")
		fmt.Print("API Key: ")

		// Secure masked input
		byteKey, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return fmt.Errorf("\ngiris okunurken hata: %w", err)
		}
		fmt.Println() // New line after hidden input

		apiKey := strings.TrimSpace(string(byteKey))
		if apiKey == "" {
			return fmt.Errorf("API anahtari bos olamaz")
		}

		if err := validateAPIKeyFormat(apiKey); err != nil {
			return err
		}

		viper.Set("api_key", apiKey)
		if err := saveConfig(); err != nil {
			return fmt.Errorf("ayarlar kaydedilemedi: %w", err)
		}

		fmt.Println("\n-------------------------------------------")
		fmt.Println("Basarili! API anahtariniz kaydedildi.")
		fmt.Printf("Konum: %s\n", activeConfigPath())
		fmt.Println("Simdi tünel baslatabilirsiniz:")
		fmt.Println("  gorenel connect --port 3000")
		fmt.Println("-------------------------------------------")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
