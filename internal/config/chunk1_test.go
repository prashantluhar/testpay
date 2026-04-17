package config_test

import (
	"os"
	"testing"

	"github.com/prashantluhar/testpay/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_noArgsDelegatesToLoadFromFile(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://load-test")
	// Load() is the default entry — no config file.
	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, "postgres://load-test", cfg.Database.URL)
}

func TestLoad_envOverridesAll(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://base")
	t.Setenv("ENVIRONMENT", "stage")
	t.Setenv("HOST", "127.0.0.1")
	t.Setenv("PORT", "8888")
	t.Setenv("TESTPAY_MODE", "hosted")
	t.Setenv("API_KEY", "env-api-key")
	t.Setenv("JWT_SECRET", "env-jwt")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("LOG_FORMAT", "json")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://a.example , https://b.example ,,  ")

	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, "stage", cfg.Environment)
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, 8888, cfg.Server.Port)
	assert.Equal(t, "hosted", cfg.Server.Mode)
	assert.Equal(t, "env-api-key", cfg.Auth.APIKey)
	assert.Equal(t, "env-jwt", cfg.Auth.JWTSecret)
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
	assert.Equal(t, []string{"https://a.example", "https://b.example"}, cfg.CORS.AllowedOrigins)
}

func TestLoad_hostedModeRequiresAPIKey(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("TESTPAY_MODE", "hosted")
	os.Unsetenv("API_KEY")
	_, err := config.Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API_KEY")
}

func TestLoad_invalidPortEnvIsIgnored(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("PORT", "not-a-number")
	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, 7700, cfg.Server.Port, "bad PORT env shouldn't clobber the default")
}

func TestLoad_corsEnvIgnoredWhenEmpty(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("CORS_ALLOWED_ORIGINS", "")
	cfg, err := config.Load()
	require.NoError(t, err)
	// Default is ["*"] — env should not have cleared it.
	assert.NotEmpty(t, cfg.CORS.AllowedOrigins)
}

func TestLoad_corsEnvOnlyWhitespaceKeepsDefault(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("CORS_ALLOWED_ORIGINS", " , , ")
	cfg, err := config.Load()
	require.NoError(t, err)
	// All entries empty after trim — override should not replace default.
	assert.NotEmpty(t, cfg.CORS.AllowedOrigins)
}

func TestLoad_applyYAMLMissingFileIsNotAnError(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	_, err := config.LoadFromFile("/nonexistent/path/testpay.yaml")
	assert.NoError(t, err, "missing YAML file falls through to defaults + env")
}

func TestLoad_malformedYAMLErrors(t *testing.T) {
	path := writeYAML(t, "this is not: yaml: indentation\n  - broken\n: :")
	_, err := config.LoadFromFile(path)
	assert.Error(t, err)
}

func TestLoad_yamlWithJWTSecretEnv(t *testing.T) {
	path := writeYAML(t, `
database:
  url_env: DATABASE_URL
auth:
  jwt_secret_env: MY_JWT
`)
	t.Setenv("DATABASE_URL", "postgres://yaml")
	t.Setenv("MY_JWT", "yaml-jwt")
	cfg, err := config.LoadFromFile(path)
	require.NoError(t, err)
	assert.Equal(t, "yaml-jwt", cfg.Auth.JWTSecret)
}
