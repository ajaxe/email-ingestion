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
}

func (c *AppConfig) String() string {
	return fmt.Sprintf("Server Port: %d, Database DSN: %s, Redis: %s", c.Server.Port, c.Database.DSN, c.Database.Redis)
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
