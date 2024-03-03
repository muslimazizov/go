package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load()
}

type Config struct {
	Host,
	Port,
	DatabaseDSN,
	BrokerDSN,
	RedisURL,
	PostgresUser,
	PostgresPassword,
	PostgresHost,
	PostgresPort,
	PostgresDb string
}

func New() *Config {
	var cfg Config

	if cfg.Host = os.Getenv("HTTP_HOST"); cfg.Host == "" {
		cfg.Host = "0.0.0.0"
	}

	if cfg.Port = strings.Replace(os.Getenv("HTTP_PORT"), ":", "", 1); cfg.Port == "" {
		cfg.Port = "8080"
	}

	if cfg.PostgresPort = os.Getenv("POSTGRES_PORT"); cfg.PostgresPort == "" {
		cfg.PostgresPort = "5432"
	}

	if cfg.PostgresHost = os.Getenv("POSTGRES_HOST"); cfg.PostgresHost == "" {
		cfg.PostgresHost = "localhost"
	}

	if cfg.PostgresUser = os.Getenv("POSTGRES_USER"); cfg.PostgresUser == "" {
		cfg.PostgresUser = "user"
	}

	if cfg.PostgresPassword = os.Getenv("POSTGRES_PASSWORD"); cfg.PostgresPassword == "" {
		cfg.PostgresPassword = "password"
	}

	if cfg.PostgresDb = os.Getenv("POSTGRES_DB"); cfg.PostgresDb == "" {
		cfg.PostgresDb = "postgres"
	}

	cfg.DatabaseDSN = fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresDb,
	)

	cfg.BrokerDSN = fmt.Sprintf(
		"nats://%s",
		os.Getenv("NATS_URL"),
	)

	if cfg.RedisURL = os.Getenv("REDIS_URL"); cfg.RedisURL == "" {
		cfg.RedisURL = "0.0.0.0:6379"
	}

	return &cfg
}
