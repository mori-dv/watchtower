package config

import (
	"errors"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Workers int      `yaml:"workers"`
	Targets []Target `yaml:"targets"`
	Alerts  Alerts   `yaml:"alerts"`
}

type Target struct {
	Name     string        `yaml:"name"`
	Type     string        `yaml:"type"`
	Address  string        `yaml:"address"`
	Interval time.Duration `yaml:"interval"`
	Timeout  time.Duration `yaml:"timeout"`
	Retries  int           `yaml:"retries"`
}

type Alerts struct {
	TelegramBotToken string        `yaml:"telegram_bot_token"`
	TelegramChatId   string        `yaml:"telegram_chat_id"`
	Cooldown         time.Duration `yaml:"cooldown"`
}

func Load(path string) (*Config, error) {
		data, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}
		var cfg Config

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
		if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
		if c.Workers <= 0 {
		return errors.New("workers must be greater than zero")
	}
	for _, t := range c.Targets {
		if t.Type != "http" && t.Type != "tcp" {
			return errors.New("invalid target type")
		}
		if t.Timeout <= 0 {
			return errors.New("timeout must be positive")
		}
		if t.Interval <= 0 {
			return errors.New("interval must be positive")
		}
	}
	return nil
}