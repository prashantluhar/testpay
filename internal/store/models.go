package store

import "time"

type Workspace struct {
	ID        string
	Slug      string
	APIKey    string
	CreatedAt time.Time
}

type User struct {
	ID          string
	WorkspaceID string
	Email       string
	Role        string
	CreatedAt   time.Time
}

type Step struct {
	Event   string `json:"event"`
	Outcome string `json:"outcome"`
	Code    string `json:"code,omitempty"`
}

type Scenario struct {
	ID             string
	WorkspaceID    string
	Name           string
	Description    string
	Gateway        string
	Steps          []Step
	WebhookDelayMs int
	IsDefault      bool
	CreatedAt      time.Time
}

type ScenarioRun struct {
	ID          string
	ScenarioID  string
	Status      string // running | completed | failed
	StartedAt   time.Time
	CompletedAt *time.Time
}

type Session struct {
	ID          string
	WorkspaceID string
	ScenarioID  string
	TTLSeconds  int
	ExpiresAt   time.Time
	CreatedAt   time.Time
}

type AttemptLog struct {
	Status     int    `json:"status"`
	DurationMs int    `json:"duration_ms"`
	Response   string `json:"response"`
	AttemptedAt string `json:"attempted_at"`
}

type RequestLog struct {
	ID              string
	WorkspaceID     string
	ScenarioRunID   *string
	Gateway         string
	Method          string
	Path            string
	RequestHeaders  map[string]string
	RequestBody     map[string]any
	ResponseHeaders map[string]string
	ResponseBody    map[string]any
	ResponseStatus  int
	DurationMs      int
	ClientIP        string
	CreatedAt       time.Time
}

type WebhookLog struct {
	ID             string
	RequestLogID   string
	Payload        map[string]any
	TargetURL      string
	DeliveryStatus string // pending | delivered | failed | duplicate
	Attempts       int
	AttemptLogs    []AttemptLog
	DeliveredAt    *time.Time
	CreatedAt      time.Time
}
