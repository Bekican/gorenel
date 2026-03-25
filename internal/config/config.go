package config

import (
	"errors"
	"fmt"
	"strconv"
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

	GithubClientID     string `mapstructure:"GITHUB_CLIENT_ID"`
	GithubClientSecret string `mapstructure:"GITHUB_CLIENT_SECRET"`
	GithubRedirectURL  string `mapstructure:"GITHUB_REDIRECT_URL"`

	// Iyzico (Payments)
	IyzicoAPIKey    string `mapstructure:"IYZICO_API_KEY"`
	IyzicoSecretKey string `mapstructure:"IYZICO_SECRET_KEY"`
	IyzicoBaseURL   string `mapstructure:"IYZICO_BASE_URL"`

	// Rate Limiting
	RateLimitRequests int           `mapstructure:"RATE_LIMIT_REQUESTS"`
	RateLimitWindow   time.Duration `mapstructure:"RATE_LIMIT_WINDOW"`

	// Traffic Inspector
	InspectorHistorySize int `mapstructure:"INSPECTOR_HISTORY_SIZE"`
	// Hard cap for captured request/response bodies (for inspector + ML).
	InspectorMaxBodyBytes int64 `mapstructure:"INSPECTOR_MAX_BODY_BYTES"`
	// Sampling rate for inspector + ML analysis. 1.0 = always, 0.0 = never.
	InspectorSamplingRate float64 `mapstructure:"INSPECTOR_SAMPLING_RATE"`
}

func Load() (*Config, error) {
	viper.SetDefault("GO_ENV", "development")
	viper.SetDefault("JWT_SECRET", "SUPER_SECRET_KEY_CHANGE_THIS_IN_PROD")
	viper.SetDefault("CONTROL_PORT", ":7000")
	viper.SetDefault("PROXY_PORT", ":8085")
	viper.SetDefault("MONITOR_PORT", ":9091")
	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("DB_URL", "postgres://postgres:postgres@localhost:5432/gorenel?sslmode=disable")
	viper.SetDefault("ML_URL", "http://localhost:5000")

	viper.SetDefault("CLICKHOUSE_ADDR", "localhost:9001")
	viper.SetDefault("CLICKHOUSE_DB", "gorenel")
	viper.SetDefault("BASE_DOMAIN", "gorenel.site")
	viper.SetDefault("ACME_EMAIL", "admin@gorenel.site")

	viper.SetDefault("RATE_LIMIT_REQUESTS", 1000)
	viper.SetDefault("RATE_LIMIT_WINDOW", 1*time.Minute)
	viper.SetDefault("INSPECTOR_HISTORY_SIZE", 100)
	viper.SetDefault("INSPECTOR_MAX_BODY_BYTES", int64(5*1024*1024))
	viper.SetDefault("INSPECTOR_SAMPLING_RATE", 1.0)

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Explicitly bind environment variables
	viper.BindEnv("GO_ENV")
	viper.BindEnv("JWT_SECRET")
	viper.BindEnv("GOOGLE_CLIENT_ID")
	viper.BindEnv("GOOGLE_CLIENT_SECRET")
	viper.BindEnv("GOOGLE_REDIRECT_URL")
	viper.BindEnv("GITHUB_CLIENT_ID")
	viper.BindEnv("GITHUB_CLIENT_SECRET")
	viper.BindEnv("GITHUB_REDIRECT_URL")
	viper.BindEnv("CLICKHOUSE_ADDR")
	viper.BindEnv("CLICKHOUSE_USER")
	viper.BindEnv("BASE_DOMAIN")
	viper.BindEnv("ACME_EMAIL")

	// Optionally load from .env file
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	_ = viper.ReadInConfig() // Ignore error if .env file is missing

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	// In production, prefer faster defaults unless user explicitly set sampling rate.
	if config.Env == "production" && !viper.IsSet("INSPECTOR_SAMPLING_RATE") {
		config.InspectorSamplingRate = 0.05
	}

	// Security Check: Fail if using default JWT secret in production
	if config.Env == "production" {
		if config.JWTSecret == "SUPER_SECRET_KEY_CHANGE_THIS_IN_PROD" {
			return nil, errors.New("JWT_SECRET must be set to a secure random string in production environment")
		}
		if config.GoogleClientID == "" || config.GoogleClientSecret == "" {
			return nil, errors.New("GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET must be set in production environment")
		}
		if config.ClickHouseAddr == "" || config.ClickHouseUser == "" {
			return nil, errors.New("CLICKHOUSE_ADDR and CLICKHOUSE_USER must be set in production environment")
		}
	}

	if config.RateLimitRequests <= 0 {
		return nil, errors.New("RATE_LIMIT_REQUESTS must be greater than 0")
	}
	if config.RateLimitWindow <= 0 {
		return nil, errors.New("RATE_LIMIT_WINDOW must be greater than 0")
	}
	if config.InspectorMaxBodyBytes <= 0 {
		return nil, errors.New("INSPECTOR_MAX_BODY_BYTES must be greater than 0")
	}
	if config.InspectorSamplingRate < 0 || config.InspectorSamplingRate > 1 {
		return nil, errors.New("INSPECTOR_SAMPLING_RATE must be between 0 and 1")
	}
	if config.RedisAddr == "" {
		return nil, errors.New("REDIS_ADDR must be set")
	}
	if config.DBURL == "" {
		return nil, errors.New("DB_URL must be set")
	}
	if config.ControlPort == config.ProxyPort || config.ControlPort == config.MonitorPort || config.ProxyPort == config.MonitorPort {
		return nil, errors.New("CONTROL_PORT, PROXY_PORT and MONITOR_PORT must be different")
	}

	for _, p := range []struct {
		name  string
		value string
	}{
		{name: "CONTROL_PORT", value: config.ControlPort},
		{name: "PROXY_PORT", value: config.ProxyPort},
		{name: "MONITOR_PORT", value: config.MonitorPort},
	} {
		if err := validatePort(p.value); err != nil {
			return nil, fmt.Errorf("%s is invalid: %w", p.name, err)
		}
	}

	return &config, nil
}

func validatePort(port string) error {
	if !strings.HasPrefix(port, ":") {
		return errors.New("must start with ':' (example :7000)")
	}
	numeric := strings.TrimPrefix(port, ":")
	if numeric == "" {
		return errors.New("empty port value")
	}
	p, err := strconv.Atoi(numeric)
	if err != nil {
		return err
	}
	if p < 1 || p > 65535 {
		return errors.New("port must be between 1 and 65535")
	}
	return nil
}
