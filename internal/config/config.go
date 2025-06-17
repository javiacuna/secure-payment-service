package config

import (
	"os"
)

type Config struct {
	DatabaseURL string
	Address     string
}

func Load() (Config, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgresql://user:password@localhost:5432/securepayment?sslmode=disable"
	}

	address := os.Getenv("ADDRESS")
	if address == "" {
		address = ":8080"
	}

	cfg := Config{
		DatabaseURL: databaseURL,
		Address:     address,
	}

	return cfg, nil
}
