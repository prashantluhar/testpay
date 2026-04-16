package store

import (
	"context"
	"time"
)

// LocalWorkspaceID is the fixed workspace ID used in local (single-tenant) mode.
const LocalWorkspaceID = "00000000-0000-0000-0000-000000000001"

type Store interface {
	// Workspace
	CreateWorkspace(ctx context.Context, w *Workspace) error
	GetWorkspaceByAPIKey(ctx context.Context, apiKey string) (*Workspace, error)
	GetWorkspaceBySlug(ctx context.Context, slug string) (*Workspace, error)
	GetWorkspaceByID(ctx context.Context, id string) (*Workspace, error)
	UpdateWorkspace(ctx context.Context, w *Workspace) error

	// Scenarios
	ListScenarios(ctx context.Context, workspaceID string) ([]*Scenario, error)
	CreateScenario(ctx context.Context, s *Scenario) error
	GetScenario(ctx context.Context, id string) (*Scenario, error)
	UpdateScenario(ctx context.Context, s *Scenario) error
	DeleteScenario(ctx context.Context, id string) error
	GetDefaultScenario(ctx context.Context, workspaceID string) (*Scenario, error)

	// ScenarioRuns
	CreateScenarioRun(ctx context.Context, r *ScenarioRun) error
	UpdateScenarioRun(ctx context.Context, r *ScenarioRun) error

	// Sessions
	CreateSession(ctx context.Context, s *Session) error
	GetActiveSession(ctx context.Context, workspaceID string) (*Session, error)
	DeleteSession(ctx context.Context, id string) error

	// RequestLogs
	CreateRequestLog(ctx context.Context, l *RequestLog) error
	ListRequestLogs(ctx context.Context, workspaceID string, limit, offset int) ([]*RequestLog, error)
	GetRequestLog(ctx context.Context, id string) (*RequestLog, error)

	// WebhookLogs
	CreateWebhookLog(ctx context.Context, l *WebhookLog) error
	UpdateWebhookLog(ctx context.Context, l *WebhookLog) error
	GetWebhookLogByRequestID(ctx context.Context, requestLogID string) (*WebhookLog, error)
	GetWebhookLog(ctx context.Context, id string) (*WebhookLog, error)
	ListWebhookLogs(ctx context.Context, workspaceID string, limit, offset int) ([]*WebhookLog, error)

	// Users
	CreateUser(ctx context.Context, u *User, passwordHash string) error
	GetUserByEmail(ctx context.Context, email string) (*User, string, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	UpdateUserLastLogin(ctx context.Context, id string, at time.Time) error
}
