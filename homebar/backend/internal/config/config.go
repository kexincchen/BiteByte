package config

import (
	"github.com/caarlos0/env/v10"
	"log"
)

type Config struct {
	DBHost     string `env:"POSTGRES_HOST"`
	DBPort     int    `env:"POSTGRES_PORT"`
	DBUser     string `env:"POSTGRES_USER"`
	DBPassword string `env:"POSTGRES_PASSWORD"`
	DBName     string `env:"POSTGRES_DB"`
}

func Load() *Config {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("config: %v", err)
	}
	return &cfg
}
