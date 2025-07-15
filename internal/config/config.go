package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Database          DatabaseConfig   `mapstructure:"database"`
	NATS              NATSConfig       `mapstructure:"nats"`
	Log               LogConfig        `mapstructure:"log"`
	Credinform        CredinformConfig `mapstructure:"credinform"`
	WorkerConcurrency int              `mapstructure:"worker_concurrency" env-default:"5"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type NATSConfig struct {
	URL string `mapstructure:"url"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
	JSON  bool   `mapstructure:"json"`
}

type CredinformConfig struct {
	BaseURL       string `mapstructure:"base_url"`
	Username      string `mapstructure:"username"`
	Password      string `mapstructure:"password"`
	Timeout       int    `mapstructure:"timeout"`
	RetryAttempts int    `mapstructure:"retry_attempts"`
	RetryDelay    int    `mapstructure:"retry_delay"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

func (c *Config) DatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host, c.Database.Port, c.Database.User, c.Database.Password, c.Database.DBName, c.Database.SSLMode)
}
