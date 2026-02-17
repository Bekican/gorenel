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

	// Redis
	RedisAddr string `mapstructure:"REDIS_ADDR"`

	// ML
	MLURL string `mapstructure:"ML_URL"`

	// OAuth
	GoogleClientID     string `mapstructure:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `mapstructure:"GOOGLE_CLIENT_SECRET"`
	GoogleRedirectURL  string `mapstructure:"GOOGLE_REDIRECT_URL"`

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
	viper.SetDefault("MONITOR_PORT", ":9090")
	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("ML_URL", "http://localhost:5000")
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
