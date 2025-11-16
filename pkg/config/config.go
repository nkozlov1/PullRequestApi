package config

import (
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	MaxCountReviewers int `envconfig:"MaxCountReviewers" default:"2"`
	Server            struct {
		Port uint16 `envconfig:"HTTP_PORT" default:"8080"`
	}
	DB struct {
		User     string `envconfig:"POSTGRES_USER"`
		Password string `envconfig:"POSTGRES_PASSWORD"`
		Host     string `envconfig:"POSTGRES_HOST"`
		Port     uint16 `envconfig:"POSTGRES_PORT"`
		Database string `envconfig:"POSTGRES_DB"`
	}
}

func Load(envFile string) (*Config, error) {
	err := godotenv.Load(envFile)
	if err != nil {
		slog.Info("no .env file, parsed exported variables")
	}
	c := &Config{}
	err = envconfig.Process("", c)
	if err != nil {
		return nil, fmt.Errorf("fail to load config: %e", err)
	}
	return c, nil
}

func (c *Config) DBUrl() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.DB.User,
		c.DB.Password,
		c.DB.Host,
		c.DB.Port,
		c.DB.Database,
	)
}

func (c *Config) PGXConfig() (*pgxpool.Config, error) {
	pgxConfig, err := pgxpool.ParseConfig(c.DBUrl())
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}
	return pgxConfig, nil
}
