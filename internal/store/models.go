package store

import "time"

type Workspace struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	APIKey    string    `json:"api_key"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
}

type Step struct {
	Event   string `json:"event"`
	Outcome string `json:"outcome"`
	Code    string `json:"code,omitempty"`
}

type Scenario struct {
	ID             string    `json:"id"`
	WorkspaceID    string    `json:"workspace_id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Gateway        string    `json:"gateway"`
	Steps          []Step    `json:"steps"`
	WebhookDelayMs int       `json:"webhook_delay_ms"`
	IsDefault      bool      `json:"is_default"`
	CreatedAt      time.Time `json:"created_at"`
}

type ScenarioRun struct {
	ID          string     `json:"id"`
	ScenarioID  string     `json:"scenario_id"`
	Status      string     `json:"status"` // running | completed | failed
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type Session struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	ScenarioID  string    `json:"scenario_id"`
	TTLSeconds  int       `json:"ttl_seconds"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type AttemptLog struct {
	Status      int    `json:"status"`
	DurationMs  int    `json:"duration_ms"`
	Response    string `json:"response"`
	AttemptedAt string `json:"attempted_at"`
}

type RequestLog struct {
	ID              string            `json:"id"`
	WorkspaceID     string            `json:"workspace_id"`
	ScenarioRunID   *string           `json:"scenario_run_id,omitempty"`
	Gateway         string            `json:"gateway"`
	Method          string            `json:"method"`
	Path            string            `json:"path"`
	RequestHeaders  map[string]string `json:"request_headers"`
	RequestBody     map[string]any    `json:"request_body"`
	ResponseHeaders map[string]string `json:"response_headers"`
	ResponseBody    map[string]any    `json:"response_body"`
	ResponseStatus  int               `json:"response_status"`
	DurationMs      int               `json:"duration_ms"`
	ClientIP        string            `json:"client_ip"`
	CreatedAt       time.Time         `json:"created_at"`
}

type WebhookLog struct {
	ID             string         `json:"id"`
	RequestLogID   string         `json:"request_log_id"`
	Payload        map[string]any `json:"payload"`
	TargetURL      string         `json:"target_url"`
	DeliveryStatus string         `json:"delivery_status"` // pending | delivered | failed | duplicate
	Attempts       int            `json:"attempts"`
	AttemptLogs    []AttemptLog   `json:"attempt_logs"`
	DeliveredAt    *time.Time     `json:"delivered_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
}
