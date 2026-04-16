package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
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
	workspaceID := store.LocalWorkspaceID

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

	// Persist log + dispatch webhook async
	if h.store != nil {
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
		}(ctx, reqLog)

		if !result.SkipWebhook && h.dispatcher != nil {
			targetURL := r.Header.Get("X-Webhook-URL")
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
				}(ctx, wl)
				webhook.DispatchAsync(ctx, h.dispatcher, h.store, wl, result.WebhookDelayMs)

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
					}(ctx, wl2)
					webhook.DispatchAsync(ctx, h.dispatcher, h.store, wl2, result.WebhookDelayMs+500)
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
