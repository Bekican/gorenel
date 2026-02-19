package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Env         string `mapstructure:"GO_ENV"`
	JWTSecret   string `mapstructure:"JWT_SECRET"`
	ControlPort string `mapstructure:"CONTROL_PORT"`
	ProxyPort   string `mapstructure:"PROXY_PORT"`
	MonitorPort string `mapstructure:"MONITOR_PORT"`
	DBURL       string `mapstructure:"DB_URL"`
	RedisAddr   string `mapstructure:"REDIS_ADDR"`
	MLURL       string `mapstructure:"ML_URL"`

	// Analytics (ClickHouse)
	ClickHouseAddr     string `mapstructure:"CLICKHOUSE_ADDR"`
	ClickHouseDB       string `mapstructure:"CLICKHOUSE_DB"`
	ClickHouseUser     string `mapstructure:"CLICKHOUSE_USER"`
	ClickHousePassword string `mapstructure:"CLICKHOUSE_PASSWORD"`

	// SSL/TLS (Certmagic)
	BaseDomain string `mapstructure:"BASE_DOMAIN"`
	AcmeEmail  string `mapstructure:"ACME_EMAIL"`

	// OAuth
	GoogleClientID     string `mapstructure:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `mapstructure:"GOOGLE_CLIENT_SECRET"`
	GoogleRedirectURL  string `mapstructure:"GOOGLE_REDIRECT_URL"`

	// Iyzico (Payments)
	IyzicoAPIKey    string `mapstructure:"IYZICO_API_KEY"`
	IyzicoSecretKey string `mapstructure:"IYZICO_SECRET_KEY"`
	IyzicoBaseURL   string `mapstructure:"IYZICO_BASE_URL"`

	// Rate Limiting
	RateLimitRequests int           `mapstructure:"RATE_LIMIT_REQUESTS"`
	RateLimitWindow   time.Duration `mapstructure:"RATE_LIMIT_WINDOW"`

	// Traffic Inspector
	InspectorHistorySize int `mapstructure:"INSPECTOR_HISTORY_SIZE"`
}

func Load() (*Config, error) {
	viper.SetDefault("GO_ENV", "development")
	viper.SetDefault("JWT_SECRET", "SUPER_SECRET_KEY_CHANGE_THIS_IN_PROD")
	viper.SetDefault("CONTROL_PORT", ":7000")
	viper.SetDefault("PROXY_PORT", ":8080")
	viper.SetDefault("MONITOR_PORT", ":9091")
	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("DB_URL", "postgres://postgres:postgres@localhost:5432/gorenel?sslmode=disable")
	viper.SetDefault("ML_URL", "http://localhost:5000")

	viper.SetDefault("CLICKHOUSE_ADDR", "localhost:9001")
	viper.SetDefault("CLICKHOUSE_DB", "gorenel")
	viper.SetDefault("BASE_DOMAIN", ".gorenel.io")
	viper.SetDefault("ACME_EMAIL", "admin@gorenel.io")

	viper.SetDefault("RATE_LIMIT_REQUESTS", 1000)
	viper.SetDefault("RATE_LIMIT_WINDOW", 1*time.Minute)
	viper.SetDefault("INSPECTOR_HISTORY_SIZE", 100)

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Optionally load from .env file
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	_ = viper.ReadInConfig() // Ignore error if .env file is missing

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
