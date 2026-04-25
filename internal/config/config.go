package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server         ServerConfig         `mapstructure:"server"`
	Database       DatabaseConfig       `mapstructure:"database"`
	Binance        BinanceConfig        `mapstructure:"binance"`
	StaticMapping  StaticMappingConfig  `mapstructure:"static_mapping"`
	Normalizer     NormalizerConfig     `mapstructure:"normalizer"`
	Arkham         ArkhamConfig         `mapstructure:"arkham"`
	Logging        LoggingConfig        `mapstructure:"logging"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type BinanceConfig struct {
	BaseURL string `mapstructure:"base_url"`
	WSURL   string `mapstructure:"ws_url"`
}

type StaticMappingConfig struct {
	FilePath       string `mapstructure:"file_path"`
	AutoReload     bool   `mapstructure:"auto_reload"`
	ReloadInterval string `mapstructure:"reload_interval"`
}

type NormalizerConfig struct {
	Aliases map[string]string `mapstructure:"aliases"`
}

type ArkhamConfig struct {
	BaseURL string `mapstructure:"base_url"`
	APIKey  string `mapstructure:"api_key"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
