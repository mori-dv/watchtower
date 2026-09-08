package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the complete service configuration.
type Config struct {
	Workers   int           `yaml:"workers"`
	QueueSize int           `yaml:"queue_size"`
	LogLevel  string        `yaml:"log_level"`
	LogFormat string        `yaml:"log_format"`
	Server    ServerConfig  `yaml:"server"`
	Storage   StorageConfig `yaml:"storage"`
	Targets   []Target      `yaml:"targets"`
	Alerts    Alerts        `yaml:"alerts"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

type StorageConfig struct {
	Type          string `yaml:"type"` // "memory" or "redis"
	RedisAddr     string `yaml:"redis_addr"`
	RedisPassword string `yaml:"redis_password"`
}

// Target defines an individual endpoint to probe.
type Target struct {
	Name               string            `yaml:"name" json:"name"`
	Type               string            `yaml:"type" json:"type"` // http, tcp, icmp
	Address            string            `yaml:"address" json:"address"`
	Interval           time.Duration     `yaml:"interval" json:"interval"`
	Timeout            time.Duration     `yaml:"timeout" json:"timeout"`
	Retries            int               `yaml:"retries" json:"retries"`
	Method             string            `yaml:"method,omitempty" json:"method,omitempty"`
	Headers            map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
	ExpectedStatus     int               `yaml:"expected_status,omitempty" json:"expected_status,omitempty"`
	CheckSSL           *bool             `yaml:"check_ssl,omitempty" json:"check_ssl,omitempty"`
	InsecureSkipVerify bool              `yaml:"insecure_skip_verify,omitempty" json:"insecure_skip_verify,omitempty"`
}

// Alerts defines multi-channel alert configurations.
type Alerts struct {
	Cooldown time.Duration `yaml:"cooldown"`

	// Legacy direct fields for backward compatibility
	TelegramBotToken string `yaml:"telegram_bot_token,omitempty"`
	TelegramChatId   string `yaml:"telegram_chat_id,omitempty"`

	Telegram TelegramConfig `yaml:"telegram"`
	Slack    SlackConfig    `yaml:"slack"`
	Webhook  WebhookConfig  `yaml:"webhook"`
}

type TelegramConfig struct {
	Enabled  bool   `yaml:"enabled"`
	BotToken string `yaml:"bot_token"`
	ChatID   string `yaml:"chat_id"`
}

type SlackConfig struct {
	Enabled    bool   `yaml:"enabled"`
	WebhookURL string `yaml:"webhook_url"`
	Channel    string `yaml:"channel"`
}

type WebhookConfig struct {
	Enabled bool              `yaml:"enabled"`
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers"`
}

// Load parses and validates configuration from the specified file path,
// expanding environment variables formatted as ${VAR_NAME}.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %q: %w", path, err)
	}

	// Expand environment variables
	expanded := os.ExpandEnv(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse yaml config: %w", err)
	}

	// Migrate legacy Telegram fields if configured
	if cfg.Alerts.TelegramBotToken != "" && cfg.Alerts.Telegram.BotToken == "" {
		cfg.Alerts.Telegram.BotToken = cfg.Alerts.TelegramBotToken
		cfg.Alerts.Telegram.Enabled = true
	}
	if cfg.Alerts.TelegramChatId != "" && cfg.Alerts.Telegram.ChatID == "" {
		cfg.Alerts.Telegram.ChatID = cfg.Alerts.TelegramChatId
		cfg.Alerts.Telegram.Enabled = true
	}

	// Apply sensible defaults
	cfg.applyDefaults()

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if c.Workers <= 0 {
		c.Workers = 10
	}
	if c.QueueSize <= 0 {
		c.QueueSize = 500
	}
	if c.Server.Port == "" {
		c.Server.Port = ":8080"
	} else if !strings.HasPrefix(c.Server.Port, ":") {
		c.Server.Port = ":" + c.Server.Port
	}
	if c.LogLevel == "" {
		c.LogLevel = "info"
	}
	if c.LogFormat == "" {
		c.LogFormat = "json"
	}
	if c.Storage.Type == "" {
		if c.Storage.RedisAddr != "" {
			c.Storage.Type = "redis"
		} else {
			c.Storage.Type = "memory"
		}
	}
	if c.Alerts.Cooldown <= 0 {
		c.Alerts.Cooldown = 15 * time.Minute
	}

	for i := range c.Targets {
		if c.Targets[i].Interval <= 0 {
			c.Targets[i].Interval = 15 * time.Second
		}
		if c.Targets[i].Timeout <= 0 {
			c.Targets[i].Timeout = 5 * time.Second
		}
		if c.Targets[i].Type == "http" && c.Targets[i].Method == "" {
			c.Targets[i].Method = "GET"
		}
		if c.Targets[i].Type == "" {
			c.Targets[i].Type = "http"
		}
		c.Targets[i].Type = strings.ToLower(c.Targets[i].Type)
	}
}

// Validate checks configuration sanity.
func (c *Config) Validate() error {
	if c.Workers <= 0 {
		return errors.New("workers must be greater than zero")
	}

	seen := make(map[string]struct{}, len(c.Targets))
	for idx, t := range c.Targets {
		if t.Name == "" {
			return fmt.Errorf("target at index %d has empty name", idx)
		}
		if _, exists := seen[t.Name]; exists {
			return fmt.Errorf("duplicate target name: %q", t.Name)
		}
		seen[t.Name] = struct{}{}

		switch t.Type {
		case "http", "tcp", "icmp":
			// valid
		default:
			return fmt.Errorf("target %q: unsupported type %q (must be http, tcp, or icmp)", t.Name, t.Type)
		}

		if t.Address == "" {
			return fmt.Errorf("target %q: address cannot be empty", t.Name)
		}
		if t.Timeout <= 0 {
			return fmt.Errorf("target %q: timeout must be positive", t.Name)
		}
		if t.Interval <= 0 {
			return fmt.Errorf("target %q: interval must be positive", t.Name)
		}
	}
	return nil
}