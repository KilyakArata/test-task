package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	StoragePath       string        `yaml:"storage_path"`
	Address           string        `yaml:"address"`
	Timeout           time.Duration `yaml:"timeout"`
	IdleTimeout       time.Duration `yaml:"idle_timeout"`
	DefaultExpiration time.Duration `yaml:"default_expiration"`
	CleanupInterval   time.Duration `yaml:"cleanup_interval"`
	ShutdownTimeout   time.Duration `yaml:"shutdown_timeout"`
}

func MustLoad() *Config {
	configPath := "config.yaml"

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
