package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/prashantluhar/testpay/internal/adapters"
	"github.com/prashantluhar/testpay/internal/engine"
	"github.com/prashantluhar/testpay/internal/store"
	"github.com/prashantluhar/testpay/internal/webhook"
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
	workspaceID := store.LocalWorkspaceID

	// Resolve adapter
	adapter, err := h.registry.Resolve(r.URL.Path)
	if err != nil {
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
		sess, _ := h.store.GetActiveSession(r.Context(), workspaceID)
		if sess != nil {
			sc, _ = h.store.GetScenario(r.Context(), sess.ScenarioID)
		}
		if sc == nil {
			sc, _ = h.store.GetDefaultScenario(r.Context(), workspaceID)
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
		go h.store.CreateRequestLog(r.Context(), reqLog)

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
				go h.store.CreateWebhookLog(r.Context(), wl)
				webhook.DispatchAsync(r.Context(), h.dispatcher, h.store, wl, result.WebhookDelayMs)

				if result.DuplicateWebhook {
					wl2 := &store.WebhookLog{
						ID:             uuid.NewString(),
						RequestLogID:   reqLog.ID,
						Payload:        payload,
						TargetURL:      targetURL,
						DeliveryStatus: "duplicate",
					}
					go h.store.CreateWebhookLog(r.Context(), wl2)
					webhook.DispatchAsync(r.Context(), h.dispatcher, h.store, wl2, result.WebhookDelayMs+500)
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
