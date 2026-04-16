package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/prashantluhar/testpay/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeYAML(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "testpay.yaml")
	require.NoError(t, os.WriteFile(path, []byte(contents), 0644))
	return path
}

func TestLoad_fromYAMLFile(t *testing.T) {
	path := writeYAML(t, `
environment: prod
server:
  host: 0.0.0.0
  port: 9999
  mode: hosted
database:
  url_env: DB_URL_CUSTOM
  max_connections: 25
webhook:
  max_attempts: 5
  base_delay_ms: 1000
  timeout_seconds: 15
logging:
  level: debug
  format: json
auth:
  api_key_env: MY_API_KEY
cors:
  allowed_origins:
    - https://app.testpay.dev
cloud:
  provider: aws
  aws:
    region: eu-west-1
    s3_bucket: logs-bucket
integrations:
  sentry:
    dsn_env: SENTRY_DSN
    sample_rate: 0.5
rate_limit:
  requests_per_minute: 300
  burst: 50
`)

	t.Setenv("DB_URL_CUSTOM", "postgres://env-provided-dsn")
	t.Setenv("MY_API_KEY", "secret-key")
	t.Setenv("SENTRY_DSN", "https://sentry.io/abc")

	cfg, err := config.LoadFromFile(path)
	require.NoError(t, err)
	assert.Equal(t, "prod", cfg.Environment)
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 9999, cfg.Server.Port)
	assert.Equal(t, "hosted", cfg.Server.Mode)
	assert.Equal(t, "postgres://env-provided-dsn", cfg.Database.URL)
	assert.Equal(t, 25, cfg.Database.MaxConnections)
	assert.Equal(t, 5, cfg.Webhook.MaxAttempts)
	assert.Equal(t, "secret-key", cfg.Auth.APIKey)
	assert.Equal(t, "aws", cfg.Cloud.Provider)
	assert.Equal(t, "eu-west-1", cfg.Cloud.AWS.Region)
	assert.Equal(t, "https://sentry.io/abc", cfg.Integrations.Sentry.DSN)
	assert.Equal(t, 300, cfg.RateLimit.RequestsPerMinute)
	assert.Len(t, cfg.CORS.AllowedOrigins, 1)
}

func TestLoad_defaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://default")
	cfg, err := config.LoadFromFile("")
	require.NoError(t, err)
	assert.Equal(t, "local", cfg.Environment)
	assert.Equal(t, 7700, cfg.Server.Port)
	assert.Equal(t, "local", cfg.Server.Mode)
	assert.Equal(t, 3, cfg.Webhook.MaxAttempts)
}

func TestLoad_envOverridesYAML(t *testing.T) {
	path := writeYAML(t, `
server:
  port: 9999
database:
  url_env: DATABASE_URL
`)
	t.Setenv("PORT", "7777")
	t.Setenv("DATABASE_URL", "postgres://from-env")
	cfg, err := config.LoadFromFile(path)
	require.NoError(t, err)
	assert.Equal(t, 7777, cfg.Server.Port)
	assert.Equal(t, "postgres://from-env", cfg.Database.URL)
}

func TestLoad_failsIfDatabaseURLMissing(t *testing.T) {
	os.Unsetenv("DATABASE_URL")
	_, err := config.LoadFromFile("")
	assert.Error(t, err)
}

func TestLoad_failsIfSecretEnvMissing(t *testing.T) {
	path := writeYAML(t, `
database:
  url_env: NONEXISTENT_VAR
`)
	os.Unsetenv("NONEXISTENT_VAR")
	_, err := config.LoadFromFile(path)
	assert.Error(t, err)
}
