package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/prashantluhar/testpay/internal/store"
	"github.com/prashantluhar/testpay/internal/webhook"
	"github.com/rs/zerolog"
)

func TestWebhook(d *webhook.Dispatcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		zerolog.Ctx(ctx).Info().Str("handler", "TestWebhook").Msg("handler entry")
		var req struct {
			TargetURL string         `json:"target_url"`
			Payload   map[string]any `json:"payload"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TargetURL == "" {
			zerolog.Ctx(ctx).Error().Err(err).Str("handler", "TestWebhook").Msg("target_url required")
			http.Error(w, `{"error":"target_url required"}`, 400)
			return
		}
		result, err := d.Dispatch(req.TargetURL, req.Payload)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Str("handler", "TestWebhook").Str("target_url", req.TargetURL).Int("attempts", result.Attempts).Msg("delivery failed")
			http.Error(w, `{"error":"delivery failed","attempts":`+string(rune(result.Attempts))+`}`, 502)
			return
		}
		zerolog.Ctx(ctx).Info().Str("handler", "TestWebhook").Str("target_url", req.TargetURL).Int("attempts", result.Attempts).Int("status_code", result.StatusCode).Msg("handler exit")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

// ListWebhooks returns all webhook dispatches for the caller's workspace,
// newest first. Each entry includes target URL, delivery status, attempts,
// attempt_logs, and the payload sent.
func ListWebhooks(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit == 0 {
			limit = 50
		}
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		zerolog.Ctx(ctx).Info().Str("handler", "ListWebhooks").Int("limit", limit).Int("offset", offset).Msg("handler entry")
		wsID := WorkspaceIDFromRequest(r, s)
		list, err := s.ListWebhookLogs(ctx, wsID, limit, offset)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Str("handler", "ListWebhooks").Msg("store error")
			http.Error(w, `{"error":"failed to list webhooks"}`, 500)
			return
		}
		zerolog.Ctx(ctx).Info().Str("handler", "ListWebhooks").Int("count", len(list)).Msg("handler exit")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
	}
}

// GetWebhook returns one webhook_log by its id (not by request_log_id).
func GetWebhook(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := chi.URLParam(r, "id")
		zerolog.Ctx(ctx).Info().Str("handler", "GetWebhook").Str("webhook_id", id).Msg("handler entry")
		wl, err := s.GetWebhookLog(ctx, id)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Str("handler", "GetWebhook").Str("webhook_id", id).Msg("not found")
			http.Error(w, `{"error":"not found"}`, 404)
			return
		}
		zerolog.Ctx(ctx).Info().Str("handler", "GetWebhook").Str("webhook_id", id).Msg("handler exit")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(wl)
	}
}

func GetWebhookStatus(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := chi.URLParam(r, "id")
		zerolog.Ctx(ctx).Info().Str("handler", "GetWebhookStatus").Str("request_id", id).Msg("handler entry")
		wl, err := s.GetWebhookLogByRequestID(ctx, id)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Str("handler", "GetWebhookStatus").Str("request_id", id).Msg("webhook log not found")
			http.Error(w, `{"error":"not found"}`, 404)
			return
		}
		zerolog.Ctx(ctx).Info().Str("handler", "GetWebhookStatus").Str("request_id", id).Str("delivery_status", wl.DeliveryStatus).Msg("handler exit")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(wl)
	}
}
