package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config failed: %w", err)
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config failed: %w", err)
	}

	setDefaults(cfg)
	return cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.App.Mode == "" {
		cfg.App.Mode = "debug"
	}
	if cfg.App.Port == 0 {
		cfg.App.Port = 8080
	}
	if cfg.App.JWTExpireHours == 0 {
		cfg.App.JWTExpireHours = 24
	}
	if cfg.App.JWTSecret == "" {
		cfg.App.JWTSecret = "admin-demo-dev-secret"
	}
	if cfg.MySQL.Database == "" {
		cfg.MySQL.Database = "admin_demo"
	}
	if cfg.Storage.Mode == "" {
		cfg.Storage.Mode = "local"
	}
	if cfg.Storage.Local.BaseDir == "" {
		cfg.Storage.Local.BaseDir = "storage"
	}
}

func ConfigPath() string {
	if p := os.Getenv("ADMIN_DEMO_CONFIG"); p != "" {
		return p
	}
	return "configs/config.example.yaml"
}
