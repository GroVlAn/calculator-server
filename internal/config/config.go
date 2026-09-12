package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type HTTP struct {
	Port              string        `yaml:"port"`
	MaxHeaderBytes    int           `yaml:"max_header_bytes" env-default:"4096"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout" env-default:"10s"`
	WriteTimeout      time.Duration `yaml:"write_timeout" env-default:"10s"`
	BaseHTTPPath      string        `yaml:"base_http_path" env-default:"/api"`
	DefaultTimeout    time.Duration `yaml:"default_timeout"`
}

type Settings struct {
	LoggingInterval time.Duration `yaml:"logging_interval"`
}

type Config struct {
	HTTP     HTTP     `yaml:"http"`
	Settings Settings `yaml:"settings"`
}

func New(path string) (*Config, error) {
	cfg := &Config{}

	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		return nil, fmt.Errorf("reading config file %s: %w", path, err)
	}

	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
