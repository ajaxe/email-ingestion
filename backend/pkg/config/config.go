package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type AppConfig struct {
	Server struct {
		Port        int    `mapstructure:"port"`
		LogLevel    string `mapstructure:"log_level"`
		Environment string `mapstructure:"environment"`
	} `mapstructure:"server"`
	Database struct {
		DSN   string `mapstructure:"dsn"`
		Redis string `mapstructure:"redis"`
	} `mapstructure:"database"`
	Smtp SmtpConfig `mapstructure:"smtp"`
}

type SmtpConfig struct {
	ListenAddress   string `mapstructure:"listen_address"`
	Domain          string `mapstructure:"domain"`
	ReadTimeoutSec  int    `mapstructure:"read_timeout_seconds"`
	WriteTimeoutSec int    `mapstructure:"write_timeout_seconds"`
	EmailSizeMaxMB  int    `mapstructure:"email_size_max_mb"`
	MaxLineLength   int    `mapstructure:"max_line_length"`
	EmailDomain     string `mapstructure:"email_domain"`
}

func (s *SmtpConfig) EmailMaxSizeBytes() int64 {
	return int64(s.EmailSizeMaxMB) * 1024 * 1024
}

// LoadConfig loads the application configuration from the specified path.
func LoadConfig(path string) (*AppConfig, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Environment variable overrides (EM_...)
	// e.g., EM_DATABASE_DSN overrides database.dsn
	viper.SetEnvPrefix("EM")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// It's acceptable if the config file doesn't exist *if* environment variables are provided,
		// but typically we want a base config, so we log that it wasn't found but don't strictly crash here yet.
		// For our case, let's gracefully continue because the defaults or envs might suffice.
	}

	var cfg AppConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
