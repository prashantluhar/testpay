package postgres

import (
	"errors"
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
// "no rows" results are demoted to debug — they're an expected outcome for
// existence/lookup checks (e.g. uniqueness validation, "is there a session?").
func logSlow(ctx context.Context, queryName string, start time.Time, err error) {
	elapsed := time.Since(start)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			zerolog.Ctx(ctx).Debug().
				Str("query", queryName).
				Int64("duration_ms", elapsed.Milliseconds()).
				Msg("store query: no rows")
			return
		}
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
	urls, err := json.Marshal(normaliseWebhookURLs(w.WebhookURLs))
	if err != nil {
		logSlow(ctx, "CreateWorkspace", start, err)
		return err
	}
	_, err = s.pool.Exec(ctx,
		`INSERT INTO workspaces (id, slug, api_key, webhook_urls, max_daily_requests)
		 VALUES ($1, $2, $3, $4, $5)`,
		w.ID, w.Slug, w.APIKey, urls, w.MaxDailyRequests,
	)
	logSlow(ctx, "CreateWorkspace", start, err)
	return err
}

func (s *Store) GetWorkspaceByAPIKey(ctx context.Context, apiKey string) (*store.Workspace, error) {
	start := time.Now()
	row := s.pool.QueryRow(ctx,
		`SELECT id, slug, api_key, webhook_urls, max_daily_requests, created_at FROM workspaces WHERE api_key = $1`, apiKey)
	ws, err := scanWorkspace(row)
	logSlow(ctx, "GetWorkspaceByAPIKey", start, err)
	return ws, err
}

func (s *Store) GetWorkspaceBySlug(ctx context.Context, slug string) (*store.Workspace, error) {
	start := time.Now()
	row := s.pool.QueryRow(ctx,
		`SELECT id, slug, api_key, webhook_urls, max_daily_requests, created_at FROM workspaces WHERE slug = $1`, slug)
	ws, err := scanWorkspace(row)
	logSlow(ctx, "GetWorkspaceBySlug", start, err)
	return ws, err
}

func (s *Store) GetWorkspaceByID(ctx context.Context, id string) (*store.Workspace, error) {
	start := time.Now()
	row := s.pool.QueryRow(ctx,
		`SELECT id, slug, api_key, webhook_urls, max_daily_requests, created_at FROM workspaces WHERE id = $1`, id)
	ws, err := scanWorkspace(row)
	logSlow(ctx, "GetWorkspaceByID", start, err)
	return ws, err
}

// UpdateWorkspace persists changes to mutable workspace fields.
// Currently only WebhookURLs is editable; slug and api_key are immutable after create.
func (s *Store) UpdateWorkspace(ctx context.Context, w *store.Workspace) error {
	start := time.Now()
	urls, err := json.Marshal(normaliseWebhookURLs(w.WebhookURLs))
	if err != nil {
		logSlow(ctx, "UpdateWorkspace", start, err)
		return err
	}
	_, err = s.pool.Exec(ctx,
		`UPDATE workspaces SET webhook_urls = $1 WHERE id = $2`,
		urls, w.ID,
	)
	logSlow(ctx, "UpdateWorkspace", start, err)
	return err
}

func scanWorkspace(row pgx.Row) (*store.Workspace, error) {
	var w store.Workspace
	var urlsRaw []byte
	if err := row.Scan(&w.ID, &w.Slug, &w.APIKey, &urlsRaw, &w.MaxDailyRequests, &w.CreatedAt); err != nil {
		return nil, fmt.Errorf("workspace not found: %w", err)
	}
	w.WebhookURLs = map[string]string{}
	if len(urlsRaw) > 0 {
		_ = json.Unmarshal(urlsRaw, &w.WebhookURLs)
	}
	return &w, nil
}

// normaliseWebhookURLs guarantees a non-nil map so JSONB stores '{}' rather than 'null'.
func normaliseWebhookURLs(m map[string]string) map[string]string {
	if m == nil {
		return map[string]string{}
	}
	return m
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
		`SELECT id, workspace_id, scenario_id, ttl_seconds, expires_at, call_index, created_at
		 FROM sessions WHERE workspace_id = $1 AND expires_at > NOW() ORDER BY created_at DESC LIMIT 1`,
		workspaceID,
	).Scan(&sess.ID, &sess.WorkspaceID, &sess.ScenarioID, &sess.TTLSeconds, &sess.ExpiresAt, &sess.CallIndex, &sess.CreatedAt)
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

// BumpSessionCallIndex increments call_index by 1 atomically and returns the
// PRE-bump value so callers can use it as the step index for the current
// request without a round-trip + race.
func (s *Store) BumpSessionCallIndex(ctx context.Context, sessionID string) (int, error) {
	start := time.Now()
	var pre int
	err := s.pool.QueryRow(ctx,
		`UPDATE sessions
		 SET call_index = call_index + 1
		 WHERE id = $1
		 RETURNING call_index - 1`,
		sessionID,
	).Scan(&pre)
	logSlow(ctx, "BumpSessionCallIndex", start, err)
	if err != nil {
		return 0, fmt.Errorf("bump session call_index: %w", err)
	}
	return pre, nil
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
		 (id, workspace_id, scenario_run_id, scenario_id, merchant_order_id,
		  gateway, method, path,
		  request_headers, request_body, response_headers, response_body,
		  response_status, duration_ms, client_ip)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		l.ID, l.WorkspaceID, l.ScenarioRunID, l.ScenarioID, l.MerchantOrderID,
		l.Gateway, l.Method, l.Path,
		reqHeaders, reqBody, respHeaders, respBody,
		l.ResponseStatus, l.DurationMs, l.ClientIP,
	)
	logSlow(ctx, "CreateRequestLog", start, err)
	return err
}

func (s *Store) ListRequestLogs(ctx context.Context, workspaceID string, limit, offset int) ([]*store.RequestLog, error) {
	start := time.Now()
	rows, err := s.pool.Query(ctx,
		`SELECT id, workspace_id, scenario_run_id, scenario_id, merchant_order_id,
		        gateway, method, path,
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

func (s *Store) CountRequestsSince(ctx context.Context, workspaceID string, since time.Time) (int, error) {
	start := time.Now()
	var n int
	err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM request_logs WHERE workspace_id = $1 AND created_at >= $2`,
		workspaceID, since,
	).Scan(&n)
	logSlow(ctx, "CountRequestsSince", start, err)
	return n, err
}

// TrimOldLogs deletes request_logs older than cutoff. Webhook rows cascade
// via request_log_id FK. Returns rows deleted.
func (s *Store) TrimOldLogs(ctx context.Context, cutoff time.Time) (int64, error) {
	start := time.Now()
	tag, err := s.pool.Exec(ctx,
		`DELETE FROM request_logs WHERE created_at < $1`, cutoff,
	)
	logSlow(ctx, "TrimOldLogs", start, err)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (s *Store) GetRequestLog(ctx context.Context, id string) (*store.RequestLog, error) {
	start := time.Now()
	row := s.pool.QueryRow(ctx,
		`SELECT id, workspace_id, scenario_run_id, scenario_id, merchant_order_id,
		        gateway, method, path,
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
	var merchant *string
	if err := row.Scan(&l.ID, &l.WorkspaceID, &l.ScenarioRunID, &l.ScenarioID, &merchant,
		&l.Gateway, &l.Method, &l.Path,
		&reqH, &reqB, &respH, &respB, &l.ResponseStatus, &l.DurationMs, &l.ClientIP, &l.CreatedAt); err != nil {
		return nil, err
	}
	if merchant != nil {
		l.MerchantOrderID = *merchant
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

func (s *Store) GetWebhookLog(ctx context.Context, id string) (*store.WebhookLog, error) {
	start := time.Now()
	var l store.WebhookLog
	var payload, attemptLogs []byte
	err := s.pool.QueryRow(ctx,
		`SELECT id, request_log_id, payload, target_url, delivery_status, attempts, attempt_logs, delivered_at, created_at
		 FROM webhook_logs WHERE id = $1`, id,
	).Scan(&l.ID, &l.RequestLogID, &payload, &l.TargetURL, &l.DeliveryStatus,
		&l.Attempts, &attemptLogs, &l.DeliveredAt, &l.CreatedAt)
	if err != nil {
		logSlow(ctx, "GetWebhookLog", start, err)
		return nil, err
	}
	json.Unmarshal(payload, &l.Payload)
	json.Unmarshal(attemptLogs, &l.AttemptLogs)
	logSlow(ctx, "GetWebhookLog", start, nil)
	return &l, nil
}

// ListWebhookLogs returns webhook logs whose parent request_log belongs to the
// given workspace. Newest first.
func (s *Store) ListWebhookLogs(ctx context.Context, workspaceID string, limit, offset int) ([]*store.WebhookLog, error) {
	start := time.Now()
	rows, err := s.pool.Query(ctx,
		`SELECT wl.id, wl.request_log_id, wl.payload, wl.target_url, wl.delivery_status,
		        wl.attempts, wl.attempt_logs, wl.delivered_at, wl.created_at
		 FROM webhook_logs wl
		 JOIN request_logs rl ON rl.id = wl.request_log_id
		 WHERE rl.workspace_id = $1
		 ORDER BY wl.created_at DESC
		 LIMIT $2 OFFSET $3`,
		workspaceID, limit, offset,
	)
	if err != nil {
		logSlow(ctx, "ListWebhookLogs", start, err)
		return nil, err
	}
	defer rows.Close()

	out := make([]*store.WebhookLog, 0)
	for rows.Next() {
		var l store.WebhookLog
		var payload, attemptLogs []byte
		if scanErr := rows.Scan(&l.ID, &l.RequestLogID, &payload, &l.TargetURL, &l.DeliveryStatus,
			&l.Attempts, &attemptLogs, &l.DeliveredAt, &l.CreatedAt); scanErr != nil {
			logSlow(ctx, "ListWebhookLogs", start, scanErr)
			return nil, scanErr
		}
		json.Unmarshal(payload, &l.Payload)
		json.Unmarshal(attemptLogs, &l.AttemptLogs)
		out = append(out, &l)
	}
	logSlow(ctx, "ListWebhookLogs", start, rows.Err())
	return out, rows.Err()
}

// ── Feedback ─────────────────────────────────────────────────────────────────

func (s *Store) CreateFeedback(ctx context.Context, f *store.Feedback) error {
	start := time.Now()
	_, err := s.pool.Exec(ctx,
		`INSERT INTO feedback_submissions
		 (id, workspace_id, user_id, what_tried, worked, missing, email, user_agent, page_url)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		f.ID, f.WorkspaceID, f.UserID, f.WhatTried, f.Worked, f.Missing, f.Email, f.UserAgent, f.PageURL,
	)
	logSlow(ctx, "CreateFeedback", start, err)
	return err
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
// dailyCap is the per-24h request limit (0 = unlimited). Migrations already
// set this to 200 on existing databases, but fresh boots need the seed to
// carry the cap in the INSERT payload.
func SeedLocalWorkspace(ctx context.Context, s *Store, dailyCap int) error {
	_, err := s.GetWorkspaceBySlug(ctx, "local")
	if err == nil {
		return nil // already exists
	}
	ws := &store.Workspace{
		ID:     store.LocalWorkspaceID,
		Slug:   "local",
		APIKey: "local",
	}
	if dailyCap > 0 {
		ws.MaxDailyRequests = &dailyCap
	}
	return s.CreateWorkspace(ctx, ws)
}

// unused import guard
var _ = time.Now
