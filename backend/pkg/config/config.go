package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type AppConfig struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Smtp     SmtpConfig     `mapstructure:"smtp"`
	Storage  StorageConfig  `mapstructure:"storage"`
}

type ServerConfig struct {
	Port        int    `mapstructure:"port"`
	LogLevel    string `mapstructure:"log_level"`
	Environment string `mapstructure:"environment"`
}

type DatabaseConfig struct {
	DSN   string `mapstructure:"dsn"`
	Redis string `mapstructure:"redis"`
}

// SmtpConfig holds configuration settings for the SMTP server.
type SmtpConfig struct {
	// ListenAddress is the address and port the SMTP server listens on (e.g., "0.0.0.0:2525")
	ListenAddress string `mapstructure:"listen_address"`
	// Domain is the domain name used for SMTP helo/ehlo greeting
	Domain string `mapstructure:"domain"`
	// ReadTimeoutSec is the maximum duration for reading SMTP commands (in seconds)
	ReadTimeoutSec int `mapstructure:"read_timeout_seconds"`
	// WriteTimeoutSec is the maximum duration for writing SMTP responses (in seconds)
	WriteTimeoutSec int `mapstructure:"write_timeout_seconds"`
	// EmailSizeMaxMB is the maximum allowed email size in megabytes
	EmailSizeMaxMB int `mapstructure:"email_size_max_mb"`
	// MaxLineLength is the maximum allowed length for a single SMTP line
	MaxLineLength int `mapstructure:"max_line_length"`
	// EmailDomain is the domain used for email addresses (e.g., for bounce handling)
	EmailDomain string `mapstructure:"email_domain"`
	// S3Bucket is the AWS S3 bucket name for spooling and storage
	S3Bucket string `mapstructure:"s3_bucket"`
	// AwsRegion is the AWS region for the S3 bucket
	AwsRegion string `mapstructure:"aws_region"`
}

type StorageConfig struct {
	// AwsRegion is the AWS region for the S3 bucket
	AwsRegion string `mapstructure:"aws_region"`
	S3Bucket  string `mapstructure:"s3_bucket"`
	// IngestionPrefix is the prefix for the S3 bucket used for storing raw email files before they are processed
	IngestionPrefix string `mapstructure:"ingestion_prefix"`
	// StoragePrefix is the prefix for the S3 bucket used for storing processed email files after conversion
	StoragePrefix string `mapstructure:"storage_prefix"`
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
