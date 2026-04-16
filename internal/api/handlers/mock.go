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
	log := zerolog.Ctx(ctx)

	log.Info().
		Str("step", "mock_entry").
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Msg("mock request received")

	// ── 1. Resolve workspace ────────────────────────────────────────────────
	// Priority: Authorization: Bearer <api_key> → LocalWorkspaceID fallback.
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
	log.Info().
		Str("step", "workspace_resolved").
		Str("workspace_id", workspaceID).
		Msg("workspace resolved")

	// ── 2. Resolve adapter ──────────────────────────────────────────────────
	adapter, err := h.registry.Resolve(r.URL.Path)
	if err != nil {
		log.Error().
			Err(err).
			Str("step", "adapter_resolve").
			Str("path", r.URL.Path).
			Msg("unknown gateway")
		http.Error(w, `{"error":"unknown gateway"}`, http.StatusNotFound)
		return
	}
	log.Info().
		Str("step", "adapter_resolved").
		Str("gateway", adapter.Name()).
		Msg("adapter resolved")

	// ── 3. Read + parse body ────────────────────────────────────────────────
	rawBody, _ := io.ReadAll(r.Body)
	var bodyMap map[string]any
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &bodyMap); err != nil {
			log.Warn().
				Err(err).
				Str("step", "body_parse").
				Int("body_bytes", len(rawBody)).
				Msg("request body is not valid JSON — continuing with nil body")
		}
	}
	log.Info().
		Str("step", "body_parsed").
		Int("body_bytes", len(rawBody)).
		Interface("request_body", bodyMap).
		Msg("request body parsed")

	// ── 4. Resolve scenario ─────────────────────────────────────────────────
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
	log.Info().
		Str("step", "scenario_resolved").
		Str("scenario_id", sc.ID).
		Str("scenario_name", sc.Name).
		Int("step_index", stepIndex).
		Msg("scenario resolved")

	// ── 5. Execute engine ───────────────────────────────────────────────────
	result, _ := h.engine.Execute(sc, stepIndex)
	log.Info().
		Str("step", "engine_executed").
		Str("outcome", string(result.Mode)).
		Int("http_status", result.HTTPStatus).
		Bool("skip_webhook", result.SkipWebhook).
		Bool("duplicate_webhook", result.DuplicateWebhook).
		Msg("engine step executed")

	// ── 6. Build + write response ───────────────────────────────────────────
	status, respBody, headers := adapter.BuildResponse(result, rawBody)
	for k, v := range headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(status)
	_, _ = w.Write(respBody)

	durationMs := int(time.Since(start).Milliseconds())
	log.Info().
		Str("step", "response_sent").
		Int("response_status", status).
		Int("response_bytes", len(respBody)).
		Int("duration_ms", durationMs).
		Msg("response written")

	// ── 7. Persist + dispatch (fire-and-forget) ─────────────────────────────
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
				zerolog.Ctx(ctx).Error().
					Err(err).
					Str("step", "persist_request_log").
					Str("request_log_id", l.ID).
					Msg("failed to persist request log")
				return
			}
			zerolog.Ctx(ctx).Debug().
				Str("step", "persist_request_log").
				Str("request_log_id", l.ID).
				Msg("request log persisted")
		}(persistCtx, reqLog)

		if !result.SkipWebhook && h.dispatcher != nil {
			// Webhook target priority:
			//   1. X-Webhook-URL header (per-request override)
			//   2. workspace.webhook_urls[adapter.Name()] (gateway-specific)
			targetURL := r.Header.Get("X-Webhook-URL")
			if targetURL == "" && workspace != nil {
				targetURL = workspace.WebhookURLs[adapter.Name()]
			}

			if targetURL == "" {
				log.Info().
					Str("step", "webhook_skipped").
					Str("reason", "no target URL configured").
					Str("gateway", adapter.Name()).
					Msg("webhook skipped")
			} else {
				amount, currency := extractAmountCurrency(bodyMap, 5000, "usd")
				payload := adapter.BuildWebhookPayload(result, chargeID, amount, currency, bodyMap)

				wl := &store.WebhookLog{
					ID:             uuid.NewString(),
					RequestLogID:   reqLog.ID,
					Payload:        payload,
					TargetURL:      targetURL,
					DeliveryStatus: "pending",
				}
				log.Info().
					Str("step", "webhook_scheduled").
					Str("webhook_log_id", wl.ID).
					Str("target_url", targetURL).
					Int("delay_ms", result.WebhookDelayMs).
					Msg("webhook dispatch scheduled")

				go func(ctx context.Context, l *store.WebhookLog) {
					if err := h.store.CreateWebhookLog(ctx, l); err != nil {
						zerolog.Ctx(ctx).Error().
							Err(err).
							Str("step", "persist_webhook_log").
							Str("webhook_log_id", l.ID).
							Msg("failed to persist webhook log")
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
							zerolog.Ctx(ctx).Error().
								Err(err).
								Str("step", "persist_webhook_log_duplicate").
								Str("webhook_log_id", l.ID).
								Msg("failed to persist duplicate webhook log")
						}
					}(persistCtx, wl2)
					webhook.DispatchAsync(persistCtx, h.dispatcher, h.store, wl2, result.WebhookDelayMs+500)
				}
			}
		}
	}

	log.Info().
		Str("step", "mock_exit").
		Str("gateway", adapter.Name()).
		Int("duration_ms", durationMs).
		Msg("mock request completed")
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

// extractAmountCurrency pulls amount + currency from a parsed request body
// with sensible defaults. Used when building the webhook payload.
func extractAmountCurrency(m map[string]any, defAmount int64, defCurrency string) (int64, string) {
	amount := defAmount
	currency := defCurrency
	if m == nil {
		return amount, currency
	}
	switch v := m["amount"].(type) {
	case float64:
		amount = int64(v)
	case int64:
		amount = v
	case int:
		amount = int64(v)
	}
	if c, ok := m["currency"].(string); ok && c != "" {
		currency = c
	}
	return amount, currency
}
