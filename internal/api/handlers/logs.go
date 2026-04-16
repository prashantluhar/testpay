package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/prashantluhar/testpay/internal/adapters"
	"github.com/prashantluhar/testpay/internal/engine"
	"github.com/prashantluhar/testpay/internal/store"
	"github.com/prashantluhar/testpay/internal/webhook"
	"github.com/rs/zerolog"
)

func ListLogs(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit == 0 {
			limit = 50
		}
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		zerolog.Ctx(ctx).Info().Str("handler", "ListLogs").Int("limit", limit).Int("offset", offset).Msg("handler entry")
		logs, err := s.ListRequestLogs(ctx, WorkspaceIDFromRequest(r, s), limit, offset)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Str("handler", "ListLogs").Msg("store error")
			http.Error(w, `{"error":"failed to list logs"}`, 500)
			return
		}
		zerolog.Ctx(ctx).Info().Str("handler", "ListLogs").Int("count", len(logs)).Msg("handler exit")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(logs)
	}
}

func GetLog(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := chi.URLParam(r, "id")
		zerolog.Ctx(ctx).Info().Str("handler", "GetLog").Str("log_id", id).Msg("handler entry")
		l, err := s.GetRequestLog(ctx, id)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Str("handler", "GetLog").Str("log_id", id).Msg("request log not found")
			http.Error(w, `{"error":"not found"}`, 404)
			return
		}
		wl, err := s.GetWebhookLogByRequestID(ctx, l.ID)
		if err != nil {
			zerolog.Ctx(ctx).Debug().Err(err).Str("handler", "GetLog").Str("log_id", id).Msg("no webhook log found for request")
		}
		zerolog.Ctx(ctx).Info().Str("handler", "GetLog").Str("log_id", id).Msg("handler exit")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"request": l, "webhook": wl})
	}
}

func ReplayLog(s store.Store, eng *engine.Engine, reg *adapters.Registry, d *webhook.Dispatcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := chi.URLParam(r, "id")
		zerolog.Ctx(ctx).Info().Str("handler", "ReplayLog").Str("log_id", id).Msg("handler entry")
		l, err := s.GetRequestLog(ctx, id)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Str("handler", "ReplayLog").Str("log_id", id).Msg("log not found")
			http.Error(w, `{"error":"log not found"}`, 404)
			return
		}
		adapter, err := reg.Resolve(l.Path)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Str("handler", "ReplayLog").Str("log_id", id).Str("path", l.Path).Msg("unknown gateway")
			http.Error(w, `{"error":"unknown gateway"}`, 400)
			return
		}
		sc := &store.Scenario{Steps: []store.Step{{Event: "charge", Outcome: "success"}}}
		result, _ := eng.Execute(sc, 0)
		status, body, headers := adapter.BuildResponse(result, nil)
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		w.WriteHeader(status)
		w.Write(body)
		zerolog.Ctx(ctx).Info().Str("handler", "ReplayLog").Str("log_id", id).Int("response_status", status).Msg("handler exit")
	}
}
