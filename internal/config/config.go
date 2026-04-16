package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port        int
	DatabaseURL string
	Mode        string // "local" | "hosted"
	APIKey      string // required in hosted mode
}

func Load() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	port := 7700
	if p := os.Getenv("PORT"); p != "" {
		var err error
		port, err = strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid PORT: %w", err)
		}
	}

	mode := os.Getenv("TESTPAY_MODE")
	if mode == "" {
		mode = "local"
	}

	return &Config{
		Port:        port,
		DatabaseURL: dbURL,
		Mode:        mode,
		APIKey:      os.Getenv("API_KEY"),
	}, nil
}
