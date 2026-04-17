package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config is the fully-resolved runtime configuration.
type Config struct {
	Environment  string
	Server       ServerConfig
	Database     DatabaseConfig
	Logging      LoggingConfig
	Webhook      WebhookConfig
	Auth         AuthConfig
	CORS         CORSConfig
	Cloud        CloudConfig
	Integrations IntegrationsConfig
	RateLimit    RateLimitConfig
	Retention    RetentionConfig
}

type ServerConfig struct {
	Host                   string
	Port                   int
	Mode                   string // local | hosted
	ReadTimeoutSeconds     int
	WriteTimeoutSeconds    int
	ShutdownTimeoutSeconds int
}

type DatabaseConfig struct {
	URL                      string // resolved from url_env
	MaxConnections           int
	MaxIdleConnections       int
	ConnectionTimeoutSeconds int
}

type LoggingConfig struct {
	Level  string
	Format string
}

type WebhookConfig struct {
	MaxAttempts    int
	BaseDelayMs    int
	TimeoutSeconds int
}

type AuthConfig struct {
	APIKey    string // resolved from api_key_env
	JWTSecret string // resolved from jwt_secret_env
}

type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
}

type CloudConfig struct {
	Provider string // aws | gcp | none
	AWS      AWSConfig
	GCP      GCPConfig
}

type AWSConfig struct {
	Region           string
	S3Bucket         string
	LogRetentionDays int
}

type GCPConfig struct {
	ProjectID string
	Region    string
}

type IntegrationsConfig struct {
	Sentry   SentryConfig
	Datadog  DatadogConfig
	Slack    SlackConfig
}

type SentryConfig struct {
	DSN        string // resolved from dsn_env
	SampleRate float64
}

type DatadogConfig struct {
	Enabled bool
	APIKey  string
	Service string
	Env     string
}

type SlackConfig struct {
	Enabled    bool
	WebhookURL string // resolved from webhook_url_env
}

type RateLimitConfig struct {
	RequestsPerMinute int
	Burst             int
}

// RetentionConfig controls the background log-trimmer + anonymous-workspace
// quota. Hosted deployments tune these to stay within storage limits and
// prevent a runaway SDK integration from draining the demo's free-tier
// resources.
type RetentionConfig struct {
	// LogRetentionDays — request_logs older than this are swept by the
	// trimmer. Defaults to 10.
	LogRetentionDays int
	// TrimmerIntervalMinutes — how often the trimmer wakes. Defaults to 60.
	TrimmerIntervalMinutes int
	// DemoWorkspaceDailyCap — requests/day cap for the seeded "local"
	// workspace (anonymous demo traffic). 0 = unlimited. Defaults to 200.
	// Signed-up workspaces get no cap by default; change via the
	// workspaces.max_daily_requests column per-workspace if needed.
	DemoWorkspaceDailyCap int
}

// rawYAML mirrors the on-disk YAML structure. All _env fields stay as env var
// NAMES here; the actual secret values are resolved into Config later.
type rawYAML struct {
	Environment string `yaml:"environment"`
	Server      struct {
		Host                   string `yaml:"host"`
		Port                   int    `yaml:"port"`
		Mode                   string `yaml:"mode"`
		ReadTimeoutSeconds     int    `yaml:"read_timeout_seconds"`
		WriteTimeoutSeconds    int    `yaml:"write_timeout_seconds"`
		ShutdownTimeoutSeconds int    `yaml:"shutdown_timeout_seconds"`
	} `yaml:"server"`
	Database struct {
		URLEnv                   string `yaml:"url_env"`
		MaxConnections           int    `yaml:"max_connections"`
		MaxIdleConnections       int    `yaml:"max_idle_connections"`
		ConnectionTimeoutSeconds int    `yaml:"connection_timeout_seconds"`
	} `yaml:"database"`
	Logging struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
	} `yaml:"logging"`
	Webhook struct {
		MaxAttempts    int `yaml:"max_attempts"`
		BaseDelayMs    int `yaml:"base_delay_ms"`
		TimeoutSeconds int `yaml:"timeout_seconds"`
	} `yaml:"webhook"`
	Auth struct {
		APIKeyEnv    string `yaml:"api_key_env"`
		JWTSecretEnv string `yaml:"jwt_secret_env"`
	} `yaml:"auth"`
	CORS struct {
		AllowedOrigins []string `yaml:"allowed_origins"`
		AllowedMethods []string `yaml:"allowed_methods"`
	} `yaml:"cors"`
	Cloud struct {
		Provider string `yaml:"provider"`
		AWS      struct {
			Region           string `yaml:"region"`
			S3Bucket         string `yaml:"s3_bucket"`
			LogRetentionDays int    `yaml:"log_retention_days"`
		} `yaml:"aws"`
		GCP struct {
			ProjectID string `yaml:"project_id"`
			Region    string `yaml:"region"`
		} `yaml:"gcp"`
	} `yaml:"cloud"`
	Integrations struct {
		Sentry struct {
			DSNEnv     string  `yaml:"dsn_env"`
			SampleRate float64 `yaml:"sample_rate"`
		} `yaml:"sentry"`
		Datadog struct {
			Enabled  bool   `yaml:"enabled"`
			APIKeyEnv string `yaml:"api_key_env"`
			Service  string `yaml:"service"`
			Env      string `yaml:"env"`
		} `yaml:"datadog"`
		Slack struct {
			Enabled       bool   `yaml:"enabled"`
			WebhookURLEnv string `yaml:"webhook_url_env"`
		} `yaml:"slack"`
	} `yaml:"integrations"`
	RateLimit struct {
		RequestsPerMinute int `yaml:"requests_per_minute"`
		Burst             int `yaml:"burst"`
	} `yaml:"rate_limit"`
	Retention struct {
		LogRetentionDays       int `yaml:"log_retention_days"`
		TrimmerIntervalMinutes int `yaml:"trimmer_interval_minutes"`
		DemoWorkspaceDailyCap  int `yaml:"demo_workspace_daily_cap"`
	} `yaml:"retention"`
}

// Load is the default entry (no config file).
func Load() (*Config, error) { return LoadFromFile("") }

// LoadFromFile loads config with precedence: defaults < YAML < env.
// Secrets referenced via *_env fields are resolved from env vars.
func LoadFromFile(path string) (*Config, error) {
	cfg := defaults()

	if path == "" {
		path = "./testpay.yaml"
	}
	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading config %s: %w", path, err)
		}
		var y rawYAML
		if err := yaml.Unmarshal(data, &y); err != nil {
			return nil, fmt.Errorf("parsing config %s: %w", path, err)
		}
		if err := applyYAML(cfg, &y); err != nil {
			return nil, err
		}
	}

	applyEnvOverrides(cfg)

	if cfg.Database.URL == "" {
		return nil, fmt.Errorf("database URL is required (set DATABASE_URL or database.url_env in config)")
	}
	if cfg.Server.Mode == "hosted" && cfg.Auth.APIKey == "" {
		return nil, fmt.Errorf("API_KEY required when mode=hosted")
	}
	return cfg, nil
}

func defaults() *Config {
	return &Config{
		Environment: "local",
		Server: ServerConfig{
			Host: "0.0.0.0", Port: 7700, Mode: "local",
			ReadTimeoutSeconds: 30, WriteTimeoutSeconds: 30, ShutdownTimeoutSeconds: 15,
		},
		Database: DatabaseConfig{
			MaxConnections: 10, MaxIdleConnections: 5, ConnectionTimeoutSeconds: 10,
		},
		Logging: LoggingConfig{Level: "info", Format: "console"},
		Webhook: WebhookConfig{MaxAttempts: 3, BaseDelayMs: 500, TimeoutSeconds: 10},
		CORS: CORSConfig{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		},
		Cloud:        CloudConfig{Provider: "none"},
		Integrations: IntegrationsConfig{Sentry: SentryConfig{SampleRate: 1.0}},
		RateLimit: RateLimitConfig{RequestsPerMinute: 600, Burst: 100},
		Retention: RetentionConfig{
			LogRetentionDays:       10,
			TrimmerIntervalMinutes: 60,
			DemoWorkspaceDailyCap:  200,
		},
	}
}

// resolveEnv reads the env var named by `envName` and returns its value.
// If envName is empty, returns "" with no error. If set but missing, returns an error.
func resolveEnv(envName, purpose string) (string, error) {
	if envName == "" {
		return "", nil
	}
	v := os.Getenv(envName)
	if v == "" {
		return "", fmt.Errorf("env var %s referenced by %s is not set", envName, purpose)
	}
	return v, nil
}

func applyYAML(cfg *Config, y *rawYAML) error {
	if y.Environment != "" {
		cfg.Environment = y.Environment
	}

	// server
	if y.Server.Host != "" {
		cfg.Server.Host = y.Server.Host
	}
	if y.Server.Port > 0 {
		cfg.Server.Port = y.Server.Port
	}
	if y.Server.Mode != "" {
		cfg.Server.Mode = y.Server.Mode
	}
	if y.Server.ReadTimeoutSeconds > 0 {
		cfg.Server.ReadTimeoutSeconds = y.Server.ReadTimeoutSeconds
	}
	if y.Server.WriteTimeoutSeconds > 0 {
		cfg.Server.WriteTimeoutSeconds = y.Server.WriteTimeoutSeconds
	}
	if y.Server.ShutdownTimeoutSeconds > 0 {
		cfg.Server.ShutdownTimeoutSeconds = y.Server.ShutdownTimeoutSeconds
	}

	// database
	if y.Database.URLEnv != "" {
		v, err := resolveEnv(y.Database.URLEnv, "database.url_env")
		if err != nil {
			return err
		}
		cfg.Database.URL = v
	}
	if y.Database.MaxConnections > 0 {
		cfg.Database.MaxConnections = y.Database.MaxConnections
	}
	if y.Database.MaxIdleConnections > 0 {
		cfg.Database.MaxIdleConnections = y.Database.MaxIdleConnections
	}
	if y.Database.ConnectionTimeoutSeconds > 0 {
		cfg.Database.ConnectionTimeoutSeconds = y.Database.ConnectionTimeoutSeconds
	}

	// logging
	if y.Logging.Level != "" {
		cfg.Logging.Level = y.Logging.Level
	}
	if y.Logging.Format != "" {
		cfg.Logging.Format = y.Logging.Format
	}

	// webhook
	if y.Webhook.MaxAttempts > 0 {
		cfg.Webhook.MaxAttempts = y.Webhook.MaxAttempts
	}
	if y.Webhook.BaseDelayMs > 0 {
		cfg.Webhook.BaseDelayMs = y.Webhook.BaseDelayMs
	}
	if y.Webhook.TimeoutSeconds > 0 {
		cfg.Webhook.TimeoutSeconds = y.Webhook.TimeoutSeconds
	}

	// auth
	if y.Auth.APIKeyEnv != "" {
		v, err := resolveEnv(y.Auth.APIKeyEnv, "auth.api_key_env")
		if err == nil {
			cfg.Auth.APIKey = v
		}
	}
	if y.Auth.JWTSecretEnv != "" {
		v, _ := resolveEnv(y.Auth.JWTSecretEnv, "auth.jwt_secret_env")
		cfg.Auth.JWTSecret = v
	}

	// cors
	if len(y.CORS.AllowedOrigins) > 0 {
		cfg.CORS.AllowedOrigins = y.CORS.AllowedOrigins
	}
	if len(y.CORS.AllowedMethods) > 0 {
		cfg.CORS.AllowedMethods = y.CORS.AllowedMethods
	}

	// cloud
	if y.Cloud.Provider != "" {
		cfg.Cloud.Provider = y.Cloud.Provider
	}
	cfg.Cloud.AWS.Region = y.Cloud.AWS.Region
	cfg.Cloud.AWS.S3Bucket = y.Cloud.AWS.S3Bucket
	cfg.Cloud.AWS.LogRetentionDays = y.Cloud.AWS.LogRetentionDays
	cfg.Cloud.GCP.ProjectID = y.Cloud.GCP.ProjectID
	cfg.Cloud.GCP.Region = y.Cloud.GCP.Region

	// integrations
	if y.Integrations.Sentry.DSNEnv != "" {
		v, _ := resolveEnv(y.Integrations.Sentry.DSNEnv, "integrations.sentry.dsn_env")
		cfg.Integrations.Sentry.DSN = v
	}
	if y.Integrations.Sentry.SampleRate > 0 {
		cfg.Integrations.Sentry.SampleRate = y.Integrations.Sentry.SampleRate
	}
	cfg.Integrations.Datadog.Enabled = y.Integrations.Datadog.Enabled
	if y.Integrations.Datadog.APIKeyEnv != "" {
		v, _ := resolveEnv(y.Integrations.Datadog.APIKeyEnv, "integrations.datadog.api_key_env")
		cfg.Integrations.Datadog.APIKey = v
	}
	cfg.Integrations.Datadog.Service = y.Integrations.Datadog.Service
	cfg.Integrations.Datadog.Env = y.Integrations.Datadog.Env

	cfg.Integrations.Slack.Enabled = y.Integrations.Slack.Enabled
	if y.Integrations.Slack.WebhookURLEnv != "" {
		v, _ := resolveEnv(y.Integrations.Slack.WebhookURLEnv, "integrations.slack.webhook_url_env")
		cfg.Integrations.Slack.WebhookURL = v
	}

	// rate_limit
	if y.RateLimit.RequestsPerMinute > 0 {
		cfg.RateLimit.RequestsPerMinute = y.RateLimit.RequestsPerMinute
	}
	if y.RateLimit.Burst > 0 {
		cfg.RateLimit.Burst = y.RateLimit.Burst
	}

	// retention — only override defaults when the YAML explicitly sets a
	// positive value; 0 keeps the coded defaults (10 days / 60 min / 200 cap).
	if y.Retention.LogRetentionDays > 0 {
		cfg.Retention.LogRetentionDays = y.Retention.LogRetentionDays
	}
	if y.Retention.TrimmerIntervalMinutes > 0 {
		cfg.Retention.TrimmerIntervalMinutes = y.Retention.TrimmerIntervalMinutes
	}
	if y.Retention.DemoWorkspaceDailyCap > 0 {
		cfg.Retention.DemoWorkspaceDailyCap = y.Retention.DemoWorkspaceDailyCap
	}

	return nil
}

// applyEnvOverrides lets common env vars override YAML values. This is separate
// from the *_env resolution above (which is YAML-pointing-at-an-env-var).
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("ENVIRONMENT"); v != "" {
		cfg.Environment = v
	}
	if v := os.Getenv("PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = n
		}
	}
	if v := os.Getenv("HOST"); v != "" {
		cfg.Server.Host = v
	}
	if v := os.Getenv("TESTPAY_MODE"); v != "" {
		cfg.Server.Mode = v
	}
	if v := os.Getenv("DATABASE_URL"); v != "" {
		cfg.Database.URL = v
	}
	if v := os.Getenv("API_KEY"); v != "" {
		cfg.Auth.APIKey = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.Auth.JWTSecret = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.Logging.Level = v
	}
	if v := os.Getenv("LOG_FORMAT"); v != "" {
		cfg.Logging.Format = v
	}
	if v := os.Getenv("CORS_ALLOWED_ORIGINS"); v != "" {
		parts := strings.Split(v, ",")
		origins := parts[:0]
		for _, p := range parts {
			if t := strings.TrimSpace(p); t != "" {
				origins = append(origins, t)
			}
		}
		if len(origins) > 0 {
			cfg.CORS.AllowedOrigins = origins
		}
	}
}
