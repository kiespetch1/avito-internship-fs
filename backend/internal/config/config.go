package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	HTTPAddr    string        `yaml:"httpAddr"`
	DatabaseURL string        `yaml:"databaseUrl"`
	JWTSecret   string        `yaml:"jwtSecret"`
	JWTTTL      time.Duration `yaml:"jwtTtl"`
	LLM         LLMConfig     `yaml:"llm"`
}

type LLMConfig struct {
	Provider     string        `yaml:"provider"`
	Timeout      time.Duration `yaml:"timeout"`
	BaseURL      string        `yaml:"baseUrl"`
	APIKey       string        `yaml:"apiKey"`
	DefaultModel string        `yaml:"defaultModel"`
}

// Load загружает конфиг
// Приоритет получения данных полей конфига: env > yaml-файл
func Load(configPath string) (Config, error) {
	var cfg Config

	path := configPath
	if path == "" {
		path = os.Getenv("CONFIG_PATH")
	}
	if path == "" {
		path = "config.yaml"
	}

	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return Config{}, fmt.Errorf("parse %s: %w", path, err)
		}
	case errors.Is(err, fs.ErrNotExist):
	default:
		return Config{}, fmt.Errorf("read %s: %w", path, err)
	}

	applyEnv(&cfg)

	return cfg, nil
}

func applyEnv(cfg *Config) {
	setString(&cfg.HTTPAddr, "HTTP_ADDR")
	setString(&cfg.DatabaseURL, "DATABASE_URL")
	setString(&cfg.JWTSecret, "JWT_SECRET")
	setDuration(&cfg.JWTTTL, "JWT_TTL")

	setString(&cfg.LLM.Provider, "LLM_PROVIDER")
	setDuration(&cfg.LLM.Timeout, "LLM_TIMEOUT")
	setString(&cfg.LLM.BaseURL, "LLM_BASE_URL")
	setString(&cfg.LLM.APIKey, "LLM_API_KEY")
	setString(&cfg.LLM.DefaultModel, "LLM_DEFAULT_MODEL")
}

func setString(dst *string, key string) {
	if v, ok := os.LookupEnv(key); ok {
		*dst = v
	}
}

func setDuration(dst *time.Duration, key string) {
	v, ok := os.LookupEnv(key)
	if !ok {
		return
	}
	if d, err := time.ParseDuration(v); err == nil {
		*dst = d
		return
	}
	if n, err := strconv.Atoi(v); err == nil {
		*dst = time.Duration(n) * time.Second
	}
}
