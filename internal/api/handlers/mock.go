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
	mode       string
}

func NewMock(eng *engine.Engine, reg *adapters.Registry, s store.Store, d *webhook.Dispatcher) http.Handler {
	return &MockHandler{engine: eng, registry: reg, store: s, dispatcher: d, mode: "local"}
}

func NewMockWithMode(eng *engine.Engine, reg *adapters.Registry, s store.Store, d *webhook.Dispatcher, mode string) http.Handler {
	return &MockHandler{engine: eng, registry: reg, store: s, dispatcher: d, mode: mode}
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
	// In hosted mode the fallback is disabled — an invalid or missing key 401s.
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
			if h.mode == "hosted" {
				log.Warn().Str("step", "workspace_resolve").Msg("hosted mode: missing or invalid api_key")
				http.Error(w, `{"error":"invalid api_key"}`, http.StatusUnauthorized)
				return
			}
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
	merchantOrderID := extractMerchantOrderID(bodyMap)
	operation := extractOperation(r.URL.Path)
	log.Info().
		Str("step", "body_parsed").
		Int("body_bytes", len(rawBody)).
		Str("operation", operation).
		Str("merchant_order_id", merchantOrderID).
		Interface("request_body", bodyMap).
		Msg("request body parsed")

	// ── 4. Resolve scenario ─────────────────────────────────────────────────
	// Priority:
	//   1. X-TestPay-Outcome header — inline single-step override, no scenario
	//      needed. Builds an ephemeral scenario on the fly from the header value.
	//   2. X-TestPay-Scenario-ID header — per-request override, no session needed
	//   3. Active session for this workspace — the session's scenario drives
	//      multi-step behavior (see step-index block below)
	//   4. Workspace default scenario
	//   5. Built-in "always succeed" fallback
	//
	// For paths 1/2/4/5 step 0 fires every time. For path 3 we bump the
	// session's call_index and use (pre-bump % len(steps)) so scenarios
	// with multiple steps advance across successive SDK calls.
	var sc *store.Scenario
	var activeSession *store.Session
	stepIndex := 0

	if outcome := r.Header.Get("X-TestPay-Outcome"); outcome != "" {
		if engine.IsValidMode(outcome) {
			sc = &store.Scenario{Steps: []store.Step{{Event: "charge", Outcome: outcome}}}
			log.Info().Str("step", "scenario_from_outcome").Str("outcome", outcome).Msg("scenario synthesized from X-TestPay-Outcome")
		} else {
			log.Warn().Str("step", "outcome_header_invalid").Str("outcome", outcome).Msg("X-TestPay-Outcome not a recognized mode; falling back")
		}
	}

	if h.store != nil {
		if sc == nil {
			if headerID := r.Header.Get("X-TestPay-Scenario-ID"); headerID != "" {
				if s2, err := h.store.GetScenario(ctx, headerID); err == nil && s2 != nil {
					sc = s2
					log.Info().Str("step", "scenario_from_header").Str("scenario_id", headerID).Msg("scenario resolved via X-TestPay-Scenario-ID")
				} else {
					log.Warn().Str("step", "scenario_header_invalid").Str("scenario_id", headerID).Msg("X-TestPay-Scenario-ID did not resolve; falling back")
				}
			}
		}
		if sc == nil {
			if sess, err := h.store.GetActiveSession(ctx, workspaceID); err == nil && sess != nil {
				if s2, serr := h.store.GetScenario(ctx, sess.ScenarioID); serr == nil && s2 != nil {
					sc = s2
					activeSession = sess
				}
			}
		}
		if sc == nil {
			sc, _ = h.store.GetDefaultScenario(ctx, workspaceID)
		}
	}
	if sc == nil {
		sc = &store.Scenario{Steps: []store.Step{{Event: "charge", Outcome: "success"}}}
	}

	// If the scenario came from an active session, advance the per-session
	// call counter atomically so each call gets the next step.
	if activeSession != nil && h.store != nil && len(sc.Steps) > 0 {
		pre, berr := h.store.BumpSessionCallIndex(ctx, activeSession.ID)
		if berr == nil {
			stepIndex = pre % len(sc.Steps)
		} else {
			log.Warn().Err(berr).Str("step", "bump_session_call_index").Msg("failed to bump session call_index; using step 0")
		}
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

	// ── 7. Persist + dispatch ───────────────────────────────────────────────
	// Single goroutine that persists request_log first, THEN webhook_log(s),
	// THEN triggers dispatch. Sequencing avoids FK violations on webhook_logs
	// (request_log_id → request_logs.id must exist first).
	if h.store != nil {
		persistCtx := context.WithoutCancel(ctx)
		chargeID := uuid.NewString()

		// Capture the response body + headers for the persisted log so the
		// dashboard Log Detail drawer can show what was actually returned.
		var responseBodyMap map[string]any
		if len(respBody) > 0 {
			if uerr := json.Unmarshal(respBody, &responseBodyMap); uerr != nil {
				// Non-JSON response (shouldn't happen for these adapters) —
				// stash the raw text so it's still inspectable.
				responseBodyMap = map[string]any{"_raw": string(respBody)}
			}
		}

		// Only record scenario_id when a real scenario drove this request.
		// The built-in "always succeed" fallback has no ID so sc.ID is empty.
		var scenarioIDPtr *string
		if sc != nil && sc.ID != "" {
			sid := sc.ID
			scenarioIDPtr = &sid
		}

		reqLog := &store.RequestLog{
			ID:              uuid.NewString(),
			WorkspaceID:     workspaceID,
			ScenarioID:      scenarioIDPtr,
			MerchantOrderID: merchantOrderID,
			Gateway:         adapter.Name(),
			Method:          r.Method,
			Path:            r.URL.Path,
			RequestHeaders:  headerToMap(r.Header),
			RequestBody:     bodyMap,
			ResponseHeaders: headers,
			ResponseBody:    responseBodyMap,
			ResponseStatus:  status,
			DurationMs:      durationMs,
			ClientIP:        r.RemoteAddr,
		}

		// Resolve webhook target before entering goroutine so log ordering
		// is clear in the foreground trace.
		var (
			targetURL       string
			willSendWebhook bool
		)
		if !result.SkipWebhook && h.dispatcher != nil {
			// Priority: X-Webhook-URL header > workspace.webhook_urls[gateway]
			//                                 > workspace.webhook_urls[_default]
			targetURL = r.Header.Get("X-Webhook-URL")
			if targetURL == "" && workspace != nil {
				if u, ok := workspace.WebhookURLs[adapter.Name()]; ok && u != "" {
					targetURL = u
				} else if u, ok := workspace.WebhookURLs["_default"]; ok {
					targetURL = u
				}
			}
			willSendWebhook = targetURL != ""
			if !willSendWebhook {
				log.Info().
					Str("step", "webhook_skipped").
					Str("reason", "no target URL configured").
					Str("gateway", adapter.Name()).
					Msg("webhook skipped")
			} else {
				log.Info().
					Str("step", "webhook_scheduled").
					Str("target_url", targetURL).
					Int("delay_ms", result.WebhookDelayMs).
					Msg("webhook dispatch scheduled")
			}
		}

		go func() {
			// 1. Persist the request log first.
			if err := h.store.CreateRequestLog(persistCtx, reqLog); err != nil {
				zerolog.Ctx(persistCtx).Error().
					Err(err).
					Str("step", "persist_request_log").
					Str("request_log_id", reqLog.ID).
					Msg("failed to persist request log — skipping webhook persistence")
				return
			}
			zerolog.Ctx(persistCtx).Debug().
				Str("step", "persist_request_log").
				Str("request_log_id", reqLog.ID).
				Msg("request log persisted")

			if !willSendWebhook {
				return
			}

			// 2. Build payload + persist webhook log(s).
			amount, currency := extractAmountCurrency(bodyMap, 5000, "usd")
			payload := adapter.BuildWebhookPayload(result, chargeID, amount, currency, bodyMap)

			wl := &store.WebhookLog{
				ID:             uuid.NewString(),
				RequestLogID:   reqLog.ID,
				Payload:        payload,
				TargetURL:      targetURL,
				DeliveryStatus: "pending",
			}
			if err := h.store.CreateWebhookLog(persistCtx, wl); err != nil {
				zerolog.Ctx(persistCtx).Error().
					Err(err).
					Str("step", "persist_webhook_log").
					Str("webhook_log_id", wl.ID).
					Msg("failed to persist webhook log — skipping dispatch")
				return
			}

			// 3. Fire-and-forget dispatch (DispatchAsync starts its own goroutine).
			webhook.DispatchAsync(persistCtx, h.dispatcher, h.store, wl, result.WebhookDelayMs)

			if result.DuplicateWebhook {
				wl2 := &store.WebhookLog{
					ID:             uuid.NewString(),
					RequestLogID:   reqLog.ID,
					Payload:        payload,
					TargetURL:      targetURL,
					DeliveryStatus: "duplicate",
				}
				if err := h.store.CreateWebhookLog(persistCtx, wl2); err != nil {
					zerolog.Ctx(persistCtx).Error().
						Err(err).
						Str("step", "persist_webhook_log_duplicate").
						Str("webhook_log_id", wl2.ID).
						Msg("failed to persist duplicate webhook log")
					return
				}
				webhook.DispatchAsync(persistCtx, h.dispatcher, h.store, wl2, result.WebhookDelayMs+500)
			}
		}()
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

// extractMerchantOrderID looks for the common places a caller might stash their
// own order reference: top-level order_id / merchant_order_id, or nested under
// metadata / notes (Stripe / Razorpay conventions).
func extractMerchantOrderID(m map[string]any) string {
	if m == nil {
		return ""
	}
	for _, k := range []string{"order_id", "merchant_order_id", "merchantOrderId", "reference"} {
		if v, ok := m[k].(string); ok && v != "" {
			return v
		}
	}
	for _, nestKey := range []string{"metadata", "notes"} {
		if nest, ok := m[nestKey].(map[string]any); ok {
			for _, k := range []string{"order_id", "merchant_order_id", "reference"} {
				if v, ok := nest[k].(string); ok && v != "" {
					return v
				}
			}
		}
	}
	return ""
}

// extractOperation derives a semantic operation name from the URL path:
//   /stripe/v1/charges            -> charge
//   /stripe/v1/refunds            -> refund
//   /stripe/v1/charges/:id/capture -> capture
func extractOperation(path string) string {
	p := strings.ToLower(path)
	switch {
	case strings.Contains(p, "/refund"):
		return "refund"
	case strings.Contains(p, "/capture"):
		return "capture"
	case strings.Contains(p, "/authoriz"):
		return "authorize"
	case strings.Contains(p, "/charge"), strings.Contains(p, "/payment"):
		return "charge"
	default:
		return "unknown"
	}
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
