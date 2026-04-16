package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prashantluhar/testpay/internal/store"
	"github.com/rs/zerolog"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

const slowQueryThreshold = 100 * time.Millisecond

// logSlow emits an error log if err != nil, or a warning if the query exceeded
// the slow-query threshold. Called at the bottom of every store method.
func logSlow(ctx context.Context, queryName string, start time.Time, err error) {
	elapsed := time.Since(start)
	if err != nil {
		zerolog.Ctx(ctx).Error().
			Err(err).
			Str("query", queryName).
			Int64("duration_ms", elapsed.Milliseconds()).
			Msg("store query failed")
		return
	}
	if elapsed > slowQueryThreshold {
		zerolog.Ctx(ctx).Warn().
			Str("query", queryName).
			Int64("duration_ms", elapsed.Milliseconds()).
			Msg("slow store query")
	}
}

// ── Workspace ────────────────────────────────────────────────────────────────

func (s *Store) CreateWorkspace(ctx context.Context, w *store.Workspace) error {
	start := time.Now()
	_, err := s.pool.Exec(ctx,
		`INSERT INTO workspaces (id, slug, api_key, webhook_url) VALUES ($1, $2, $3, $4)`,
		w.ID, w.Slug, w.APIKey, w.WebhookURL,
	)
	logSlow(ctx, "CreateWorkspace", start, err)
	return err
}

func (s *Store) GetWorkspaceByAPIKey(ctx context.Context, apiKey string) (*store.Workspace, error) {
	start := time.Now()
	row := s.pool.QueryRow(ctx,
		`SELECT id, slug, api_key, webhook_url, created_at FROM workspaces WHERE api_key = $1`, apiKey)
	ws, err := scanWorkspace(row)
	logSlow(ctx, "GetWorkspaceByAPIKey", start, err)
	return ws, err
}

func (s *Store) GetWorkspaceBySlug(ctx context.Context, slug string) (*store.Workspace, error) {
	start := time.Now()
	row := s.pool.QueryRow(ctx,
		`SELECT id, slug, api_key, webhook_url, created_at FROM workspaces WHERE slug = $1`, slug)
	ws, err := scanWorkspace(row)
	logSlow(ctx, "GetWorkspaceBySlug", start, err)
	return ws, err
}

func (s *Store) GetWorkspaceByID(ctx context.Context, id string) (*store.Workspace, error) {
	start := time.Now()
	row := s.pool.QueryRow(ctx,
		`SELECT id, slug, api_key, webhook_url, created_at FROM workspaces WHERE id = $1`, id)
	ws, err := scanWorkspace(row)
	logSlow(ctx, "GetWorkspaceByID", start, err)
	return ws, err
}

// UpdateWorkspace persists changes to the mutable workspace fields.
// Currently only WebhookURL is editable; slug and api_key are immutable after create.
func (s *Store) UpdateWorkspace(ctx context.Context, w *store.Workspace) error {
	start := time.Now()
	_, err := s.pool.Exec(ctx,
		`UPDATE workspaces SET webhook_url = $1 WHERE id = $2`,
		w.WebhookURL, w.ID,
	)
	logSlow(ctx, "UpdateWorkspace", start, err)
	return err
}

func scanWorkspace(row pgx.Row) (*store.Workspace, error) {
	var w store.Workspace
	if err := row.Scan(&w.ID, &w.Slug, &w.APIKey, &w.WebhookURL, &w.CreatedAt); err != nil {
		return nil, fmt.Errorf("workspace not found: %w", err)
	}
	return &w, nil
}

// ── Scenarios ────────────────────────────────────────────────────────────────

func (s *Store) CreateScenario(ctx context.Context, sc *store.Scenario) error {
	start := time.Now()
	stepsJSON, err := json.Marshal(sc.Steps)
	if err != nil {
		logSlow(ctx, "CreateScenario", start, err)
		return err
	}
	_, err = s.pool.Exec(ctx,
		`INSERT INTO scenarios (id, workspace_id, name, description, gateway, steps, webhook_delay_ms, is_default)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		sc.ID, sc.WorkspaceID, sc.Name, sc.Description, sc.Gateway,
		stepsJSON, sc.WebhookDelayMs, sc.IsDefault,
	)
	logSlow(ctx, "CreateScenario", start, err)
	return err
}

func (s *Store) ListScenarios(ctx context.Context, workspaceID string) ([]*store.Scenario, error) {
	start := time.Now()
	rows, err := s.pool.Query(ctx,
		`SELECT id, workspace_id, name, description, gateway, steps, webhook_delay_ms, is_default, created_at
		 FROM scenarios WHERE workspace_id = $1 ORDER BY created_at DESC`, workspaceID)
	if err != nil {
		logSlow(ctx, "ListScenarios", start, err)
		return nil, err
	}
	defer rows.Close()

	out := make([]*store.Scenario, 0)
	for rows.Next() {
		sc, scanErr := scanScenario(rows)
		if scanErr != nil {
			logSlow(ctx, "ListScenarios", start, scanErr)
			return nil, scanErr
		}
		out = append(out, sc)
	}
	err = rows.Err()
	logSlow(ctx, "ListScenarios", start, err)
	return out, err
}

func (s *Store) GetScenario(ctx context.Context, id string) (*store.Scenario, error) {
	start := time.Now()
	row := s.pool.QueryRow(ctx,
		`SELECT id, workspace_id, name, description, gateway, steps, webhook_delay_ms, is_default, created_at
		 FROM scenarios WHERE id = $1`, id)
	sc, err := scanScenario(row)
	logSlow(ctx, "GetScenario", start, err)
	return sc, err
}

func (s *Store) UpdateScenario(ctx context.Context, sc *store.Scenario) error {
	start := time.Now()
	stepsJSON, err := json.Marshal(sc.Steps)
	if err != nil {
		logSlow(ctx, "UpdateScenario", start, err)
		return err
	}
	_, err = s.pool.Exec(ctx,
		`UPDATE scenarios SET name=$1, description=$2, gateway=$3, steps=$4, webhook_delay_ms=$5, is_default=$6
		 WHERE id=$7`,
		sc.Name, sc.Description, sc.Gateway, stepsJSON, sc.WebhookDelayMs, sc.IsDefault, sc.ID,
	)
	logSlow(ctx, "UpdateScenario", start, err)
	return err
}

func (s *Store) DeleteScenario(ctx context.Context, id string) error {
	start := time.Now()
	_, err := s.pool.Exec(ctx, `DELETE FROM scenarios WHERE id = $1`, id)
	logSlow(ctx, "DeleteScenario", start, err)
	return err
}

func (s *Store) GetDefaultScenario(ctx context.Context, workspaceID string) (*store.Scenario, error) {
	start := time.Now()
	row := s.pool.QueryRow(ctx,
		`SELECT id, workspace_id, name, description, gateway, steps, webhook_delay_ms, is_default, created_at
		 FROM scenarios WHERE workspace_id = $1 AND is_default = TRUE LIMIT 1`, workspaceID)
	sc, err := scanScenario(row)
	logSlow(ctx, "GetDefaultScenario", start, err)
	return sc, err
}

func scanScenario(row interface {
	Scan(...any) error
}) (*store.Scenario, error) {
	var sc store.Scenario
	var stepsRaw []byte
	if err := row.Scan(&sc.ID, &sc.WorkspaceID, &sc.Name, &sc.Description, &sc.Gateway,
		&stepsRaw, &sc.WebhookDelayMs, &sc.IsDefault, &sc.CreatedAt); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(stepsRaw, &sc.Steps); err != nil {
		return nil, err
	}
	return &sc, nil
}

// ── ScenarioRuns ─────────────────────────────────────────────────────────────

func (s *Store) CreateScenarioRun(ctx context.Context, r *store.ScenarioRun) error {
	start := time.Now()
	_, err := s.pool.Exec(ctx,
		`INSERT INTO scenario_runs (id, scenario_id, status) VALUES ($1, $2, $3)`,
		r.ID, r.ScenarioID, r.Status,
	)
	logSlow(ctx, "CreateScenarioRun", start, err)
	return err
}

func (s *Store) UpdateScenarioRun(ctx context.Context, r *store.ScenarioRun) error {
	start := time.Now()
	_, err := s.pool.Exec(ctx,
		`UPDATE scenario_runs SET status=$1, completed_at=$2 WHERE id=$3`,
		r.Status, r.CompletedAt, r.ID,
	)
	logSlow(ctx, "UpdateScenarioRun", start, err)
	return err
}

// ── Sessions ─────────────────────────────────────────────────────────────────

func (s *Store) CreateSession(ctx context.Context, sess *store.Session) error {
	start := time.Now()
	_, err := s.pool.Exec(ctx,
		`INSERT INTO sessions (id, workspace_id, scenario_id, ttl_seconds, expires_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		sess.ID, sess.WorkspaceID, sess.ScenarioID, sess.TTLSeconds, sess.ExpiresAt,
	)
	logSlow(ctx, "CreateSession", start, err)
	return err
}

func (s *Store) GetActiveSession(ctx context.Context, workspaceID string) (*store.Session, error) {
	start := time.Now()
	var sess store.Session
	err := s.pool.QueryRow(ctx,
		`SELECT id, workspace_id, scenario_id, ttl_seconds, expires_at, created_at
		 FROM sessions WHERE workspace_id = $1 AND expires_at > NOW() ORDER BY created_at DESC LIMIT 1`,
		workspaceID,
	).Scan(&sess.ID, &sess.WorkspaceID, &sess.ScenarioID, &sess.TTLSeconds, &sess.ExpiresAt, &sess.CreatedAt)
	if err != nil {
		logSlow(ctx, "GetActiveSession", start, err)
		return nil, fmt.Errorf("no active session: %w", err)
	}
	logSlow(ctx, "GetActiveSession", start, nil)
	return &sess, nil
}

func (s *Store) DeleteSession(ctx context.Context, id string) error {
	start := time.Now()
	_, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE id = $1`, id)
	logSlow(ctx, "DeleteSession", start, err)
	return err
}

// ── RequestLogs ──────────────────────────────────────────────────────────────

func (s *Store) CreateRequestLog(ctx context.Context, l *store.RequestLog) error {
	start := time.Now()
	reqHeaders, _ := json.Marshal(l.RequestHeaders)
	reqBody, _ := json.Marshal(l.RequestBody)
	respHeaders, _ := json.Marshal(l.ResponseHeaders)
	respBody, _ := json.Marshal(l.ResponseBody)

	_, err := s.pool.Exec(ctx,
		`INSERT INTO request_logs
		 (id, workspace_id, scenario_run_id, gateway, method, path,
		  request_headers, request_body, response_headers, response_body,
		  response_status, duration_ms, client_ip)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		l.ID, l.WorkspaceID, l.ScenarioRunID, l.Gateway, l.Method, l.Path,
		reqHeaders, reqBody, respHeaders, respBody,
		l.ResponseStatus, l.DurationMs, l.ClientIP,
	)
	logSlow(ctx, "CreateRequestLog", start, err)
	return err
}

func (s *Store) ListRequestLogs(ctx context.Context, workspaceID string, limit, offset int) ([]*store.RequestLog, error) {
	start := time.Now()
	rows, err := s.pool.Query(ctx,
		`SELECT id, workspace_id, scenario_run_id, gateway, method, path,
		        request_headers, request_body, response_headers, response_body,
		        response_status, duration_ms, client_ip, created_at
		 FROM request_logs WHERE workspace_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		workspaceID, limit, offset,
	)
	if err != nil {
		logSlow(ctx, "ListRequestLogs", start, err)
		return nil, err
	}
	defer rows.Close()

	out := make([]*store.RequestLog, 0)
	for rows.Next() {
		l, scanErr := scanRequestLog(rows)
		if scanErr != nil {
			logSlow(ctx, "ListRequestLogs", start, scanErr)
			return nil, scanErr
		}
		out = append(out, l)
	}
	err = rows.Err()
	logSlow(ctx, "ListRequestLogs", start, err)
	return out, err
}

func (s *Store) GetRequestLog(ctx context.Context, id string) (*store.RequestLog, error) {
	start := time.Now()
	row := s.pool.QueryRow(ctx,
		`SELECT id, workspace_id, scenario_run_id, gateway, method, path,
		        request_headers, request_body, response_headers, response_body,
		        response_status, duration_ms, client_ip, created_at
		 FROM request_logs WHERE id = $1`, id)
	l, err := scanRequestLog(row)
	logSlow(ctx, "GetRequestLog", start, err)
	return l, err
}

func scanRequestLog(row interface{ Scan(...any) error }) (*store.RequestLog, error) {
	var l store.RequestLog
	var reqH, reqB, respH, respB []byte
	if err := row.Scan(&l.ID, &l.WorkspaceID, &l.ScenarioRunID, &l.Gateway, &l.Method, &l.Path,
		&reqH, &reqB, &respH, &respB, &l.ResponseStatus, &l.DurationMs, &l.ClientIP, &l.CreatedAt); err != nil {
		return nil, err
	}
	json.Unmarshal(reqH, &l.RequestHeaders)
	json.Unmarshal(reqB, &l.RequestBody)
	json.Unmarshal(respH, &l.ResponseHeaders)
	json.Unmarshal(respB, &l.ResponseBody)
	return &l, nil
}

// ── WebhookLogs ──────────────────────────────────────────────────────────────

func (s *Store) CreateWebhookLog(ctx context.Context, l *store.WebhookLog) error {
	start := time.Now()
	payload, _ := json.Marshal(l.Payload)
	attemptLogs, _ := json.Marshal(l.AttemptLogs)
	_, err := s.pool.Exec(ctx,
		`INSERT INTO webhook_logs (id, request_log_id, payload, target_url, delivery_status, attempts, attempt_logs)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		l.ID, l.RequestLogID, payload, l.TargetURL, l.DeliveryStatus, l.Attempts, attemptLogs,
	)
	logSlow(ctx, "CreateWebhookLog", start, err)
	return err
}

func (s *Store) UpdateWebhookLog(ctx context.Context, l *store.WebhookLog) error {
	start := time.Now()
	attemptLogs, _ := json.Marshal(l.AttemptLogs)
	_, err := s.pool.Exec(ctx,
		`UPDATE webhook_logs SET delivery_status=$1, attempts=$2, attempt_logs=$3, delivered_at=$4 WHERE id=$5`,
		l.DeliveryStatus, l.Attempts, attemptLogs, l.DeliveredAt, l.ID,
	)
	logSlow(ctx, "UpdateWebhookLog", start, err)
	return err
}

func (s *Store) GetWebhookLogByRequestID(ctx context.Context, requestLogID string) (*store.WebhookLog, error) {
	start := time.Now()
	var l store.WebhookLog
	var payload, attemptLogs []byte
	err := s.pool.QueryRow(ctx,
		`SELECT id, request_log_id, payload, target_url, delivery_status, attempts, attempt_logs, delivered_at, created_at
		 FROM webhook_logs WHERE request_log_id = $1`, requestLogID,
	).Scan(&l.ID, &l.RequestLogID, &payload, &l.TargetURL, &l.DeliveryStatus,
		&l.Attempts, &attemptLogs, &l.DeliveredAt, &l.CreatedAt)
	if err != nil {
		logSlow(ctx, "GetWebhookLogByRequestID", start, err)
		return nil, err
	}
	json.Unmarshal(payload, &l.Payload)
	json.Unmarshal(attemptLogs, &l.AttemptLogs)
	logSlow(ctx, "GetWebhookLogByRequestID", start, nil)
	return &l, nil
}

// ── Users ────────────────────────────────────────────────────────────────────

func (s *Store) CreateUser(ctx context.Context, u *store.User, passwordHash string) error {
	start := time.Now()
	_, err := s.pool.Exec(ctx,
		`INSERT INTO users (id, workspace_id, email, role, password_hash)
		 VALUES ($1, $2, $3, $4, $5)`,
		u.ID, u.WorkspaceID, u.Email, u.Role, passwordHash,
	)
	logSlow(ctx, "CreateUser", start, err)
	return err
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*store.User, string, error) {
	start := time.Now()
	var u store.User
	var hash string
	err := s.pool.QueryRow(ctx,
		`SELECT id, workspace_id, email, role, created_at, password_hash
		 FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.WorkspaceID, &u.Email, &u.Role, &u.CreatedAt, &hash)
	logSlow(ctx, "GetUserByEmail", start, err)
	if err != nil {
		return nil, "", err
	}
	return &u, hash, nil
}

func (s *Store) GetUserByID(ctx context.Context, id string) (*store.User, error) {
	start := time.Now()
	var u store.User
	err := s.pool.QueryRow(ctx,
		`SELECT id, workspace_id, email, role, created_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.WorkspaceID, &u.Email, &u.Role, &u.CreatedAt)
	logSlow(ctx, "GetUserByID", start, err)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Store) UpdateUserLastLogin(ctx context.Context, id string, at time.Time) error {
	start := time.Now()
	_, err := s.pool.Exec(ctx,
		`UPDATE users SET last_login_at = $1 WHERE id = $2`, at, id,
	)
	logSlow(ctx, "UpdateUserLastLogin", start, err)
	return err
}

// ConnectPool opens a pgxpool connection and pings it.
func ConnectPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connecting to postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("pinging postgres: %w", err)
	}
	return pool, nil
}

// ensure Store implements store.Store at compile time
var _ store.Store = (*Store)(nil)

// LocalWorkspaceID is retained for backward compatibility; the canonical
// constant now lives in the store package.
const LocalWorkspaceID = store.LocalWorkspaceID

// SeedLocalWorkspace creates the default local workspace if it doesn't exist.
func SeedLocalWorkspace(ctx context.Context, s *Store) error {
	_, err := s.GetWorkspaceBySlug(ctx, "local")
	if err == nil {
		return nil // already exists
	}
	return s.CreateWorkspace(ctx, &store.Workspace{
		ID:     store.LocalWorkspaceID,
		Slug:   "local",
		APIKey: "local",
	})
}

// unused import guard
var _ = time.Now
