package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/prashantluhar/testpay/internal/adapters"
	"github.com/prashantluhar/testpay/internal/engine"
	"github.com/prashantluhar/testpay/internal/store"
	"github.com/prashantluhar/testpay/internal/webhook"
	"github.com/rs/zerolog"
)

type MockHandler struct {
	engine     *engine.Engine
	registry   *adapters.Registry
	store      store.Store
	dispatcher *webhook.Dispatcher
}

func NewMock(eng *engine.Engine, reg *adapters.Registry, s store.Store, d *webhook.Dispatcher) http.Handler {
	return &MockHandler{engine: eng, registry: reg, store: s, dispatcher: d}
}

func (h *MockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Resolve the calling workspace. Priority:
	// 1. Authorization: Bearer <api_key>  → lookup workspace by api_key
	// 2. Fallback to LocalWorkspaceID (local mode)
	workspaceID := store.LocalWorkspaceID
	var workspace *store.Workspace
	if h.store != nil {
		if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
			apiKey := strings.TrimPrefix(auth, "Bearer ")
			if ws, werr := h.store.GetWorkspaceByAPIKey(ctx, apiKey); werr == nil && ws != nil {
				workspace = ws
				workspaceID = ws.ID
			}
		}
		if workspace == nil {
			if ws, werr := h.store.GetWorkspaceByID(ctx, workspaceID); werr == nil {
				workspace = ws
			}
		}
	}

	// Resolve adapter
	adapter, err := h.registry.Resolve(r.URL.Path)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Str("path", r.URL.Path).Msg("unknown gateway")
		http.Error(w, `{"error":"unknown gateway"}`, http.StatusNotFound)
		return
	}

	// Read body
	rawBody, _ := io.ReadAll(r.Body)
	var bodyMap map[string]any
	json.Unmarshal(rawBody, &bodyMap)

	// Determine active scenario
	var sc *store.Scenario
	var stepIndex int
	if h.store != nil {
		sess, _ := h.store.GetActiveSession(ctx, workspaceID)
		if sess != nil {
			sc, _ = h.store.GetScenario(ctx, sess.ScenarioID)
		}
		if sc == nil {
			sc, _ = h.store.GetDefaultScenario(ctx, workspaceID)
		}
	}
	if sc == nil {
		sc = &store.Scenario{Steps: []store.Step{{Event: "charge", Outcome: "success"}}}
	}

	// Execute engine
	result, _ := h.engine.Execute(sc, stepIndex)

	// Build response
	status, body, headers := adapter.BuildResponse(result, rawBody)
	for k, v := range headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(status)
	w.Write(body)

	durationMs := int(time.Since(start).Milliseconds())

	zerolog.Ctx(ctx).Info().
		Str("gateway", adapter.Name()).
		Str("scenario_id", sc.ID).
		Str("scenario_name", sc.Name).
		Int("step_index", stepIndex).
		Str("outcome", string(result.Mode)).
		Int("response_status", status).
		Int("duration_ms", durationMs).
		Bool("skip_webhook", result.SkipWebhook).
		Bool("duplicate_webhook", result.DuplicateWebhook).
		Msg("simulation step executed")

	// Persist log + dispatch webhook async.
	// context.WithoutCancel keeps the trace_id logger in context but prevents
	// the goroutine's DB work from being canceled when the HTTP response is
	// written (r.Context() is canceled after ServeHTTP returns).
	if h.store != nil {
		persistCtx := context.WithoutCancel(ctx)
		chargeID := uuid.NewString()
		reqLog := &store.RequestLog{
			ID:             uuid.NewString(),
			WorkspaceID:    workspaceID,
			Gateway:        adapter.Name(),
			Method:         r.Method,
			Path:           r.URL.Path,
			RequestHeaders: headerToMap(r.Header),
			RequestBody:    bodyMap,
			ResponseStatus: status,
			DurationMs:     durationMs,
			ClientIP:       r.RemoteAddr,
		}
		go func(ctx context.Context, l *store.RequestLog) {
			if err := h.store.CreateRequestLog(ctx, l); err != nil {
				zerolog.Ctx(ctx).Error().Err(err).Str("request_log_id", l.ID).Msg("failed to persist request log")
			}
		}(persistCtx, reqLog)

		if !result.SkipWebhook && h.dispatcher != nil {
			// Webhook target priority: X-Webhook-URL header > workspace.webhook_url
			targetURL := r.Header.Get("X-Webhook-URL")
			if targetURL == "" && workspace != nil {
				targetURL = workspace.WebhookURL
			}
			if targetURL != "" {
				payload := adapter.BuildWebhookPayload(result, chargeID, 5000, "usd")
				wl := &store.WebhookLog{
					ID:             uuid.NewString(),
					RequestLogID:   reqLog.ID,
					Payload:        payload,
					TargetURL:      targetURL,
					DeliveryStatus: "pending",
				}
				go func(ctx context.Context, l *store.WebhookLog) {
					if err := h.store.CreateWebhookLog(ctx, l); err != nil {
						zerolog.Ctx(ctx).Error().Err(err).Str("webhook_log_id", l.ID).Msg("failed to persist webhook log")
					}
				}(persistCtx, wl)
				webhook.DispatchAsync(persistCtx, h.dispatcher, h.store, wl, result.WebhookDelayMs)

				if result.DuplicateWebhook {
					wl2 := &store.WebhookLog{
						ID:             uuid.NewString(),
						RequestLogID:   reqLog.ID,
						Payload:        payload,
						TargetURL:      targetURL,
						DeliveryStatus: "duplicate",
					}
					go func(ctx context.Context, l *store.WebhookLog) {
						if err := h.store.CreateWebhookLog(ctx, l); err != nil {
							zerolog.Ctx(ctx).Error().Err(err).Str("webhook_log_id", l.ID).Msg("failed to persist duplicate webhook log")
						}
					}(persistCtx, wl2)
					webhook.DispatchAsync(persistCtx, h.dispatcher, h.store, wl2, result.WebhookDelayMs+500)
				}
			}
		}
	}
}

func headerToMap(h http.Header) map[string]string {
	m := make(map[string]string)
	for k, v := range h {
		if len(v) > 0 {
			m[k] = v[0]
		}
	}
	return m
}
