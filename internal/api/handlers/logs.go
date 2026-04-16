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
)

func ListLogs(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit == 0 { limit = 50 }
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		logs, err := s.ListRequestLogs(r.Context(), store.LocalWorkspaceID, limit, offset)
		if err != nil {
			http.Error(w, `{"error":"failed to list logs"}`, 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(logs)
	}
}

func GetLog(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l, err := s.GetRequestLog(r.Context(), chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, `{"error":"not found"}`, 404)
			return
		}
		wl, _ := s.GetWebhookLogByRequestID(r.Context(), l.ID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"request": l, "webhook": wl})
	}
}

func ReplayLog(s store.Store, eng *engine.Engine, reg *adapters.Registry, d *webhook.Dispatcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l, err := s.GetRequestLog(r.Context(), chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, `{"error":"log not found"}`, 404)
			return
		}
		adapter, err := reg.Resolve(l.Path)
		if err != nil {
			http.Error(w, `{"error":"unknown gateway"}`, 400)
			return
		}
		sc := &store.Scenario{Steps: []store.Step{{Event: "charge", Outcome: "success"}}}
		result, _ := eng.Execute(sc, 0)
		status, body, headers := adapter.BuildResponse(result, nil)
		for k, v := range headers { w.Header().Set(k, v) }
		w.WriteHeader(status)
		w.Write(body)
	}
}
