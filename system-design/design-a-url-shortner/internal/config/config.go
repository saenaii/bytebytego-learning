package config

import (
	"encoding/json"
	"errors"
	"os"
	"time"
)

type Config struct {
	Server         ServerConfig   `json:"server"`
	Postgres       PostgresConfig `json:"postgres"`
	Redis          RedisConfig    `json:"redis"`
	CacheTTL       string         `json:"cache_ttl"`
	DefaultMode    string         `json:"default_mode"`
	QueryBlacklist []string       `json:"query_blacklist"`
}

type ServerConfig struct {
	Port int `json:"port"`
}

type PostgresConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type RedisConfig struct {
	Address  string `json:"address"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

func Load(path string) (Config, error) {
	cfg := defaults()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return Config{}, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return errors.New("server port must be between 1 and 65535")
	}
	if c.DefaultMode != "full" && c.DefaultMode != "path_only" {
		return errors.New("default_mode must be full or path_only")
	}
	if _, err := time.ParseDuration(c.CacheTTL); err != nil {
		return errors.New("cache_ttl must be a valid duration")
	}
	if c.Postgres.Host == "" || c.Postgres.Database == "" || c.Postgres.User == "" {
		return errors.New("postgres host, database, and user are required")
	}
	if c.Redis.Address == "" {
		return errors.New("redis address is required")
	}
	return nil
}

func (c Config) TTL() time.Duration {
	d, err := time.ParseDuration(c.CacheTTL)
	if err != nil {
		return time.Hour
	}
	return d
}

func (c PostgresConfig) DSN() string {
	return "postgres://" + c.User + ":" + c.Password + "@" + c.Host + ":" + itoa(c.Port) + "/" + c.Database
}

func defaults() Config {
	return Config{
		Server:      ServerConfig{Port: 8080},
		Postgres:    PostgresConfig{Host: "localhost", Port: 5432, Database: "urlshortener", User: "urlshortener", Password: "urlshortener"},
		Redis:       RedisConfig{Address: "localhost:6379"},
		CacheTTL:    "1h",
		DefaultMode: "full",
	}
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	result := ""
	for value > 0 {
		result = string(rune('0'+value%10)) + result
		value /= 10
	}
	return result
}
