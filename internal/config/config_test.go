package config_test

import (
	"os"
	"testing"

	"github.com/prashantluhar/testpay/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_defaults(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://localhost/testpay")
	defer os.Unsetenv("DATABASE_URL")

	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, 7700, cfg.Port)
	assert.Equal(t, "local", cfg.Mode)
}

func TestLoad_missingDatabaseURL(t *testing.T) {
	os.Unsetenv("DATABASE_URL")
	_, err := config.Load()
	assert.Error(t, err)
}
