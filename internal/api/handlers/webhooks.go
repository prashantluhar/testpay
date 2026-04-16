package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prashantluhar/testpay/internal/store"
	"github.com/prashantluhar/testpay/internal/webhook"
)

func TestWebhook(d *webhook.Dispatcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			TargetURL string         `json:"target_url"`
			Payload   map[string]any `json:"payload"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TargetURL == "" {
			http.Error(w, `{"error":"target_url required"}`, 400)
			return
		}
		result, err := d.Dispatch(req.TargetURL, req.Payload)
		if err != nil {
			http.Error(w, `{"error":"delivery failed","attempts":`+string(rune(result.Attempts))+`}`, 502)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

func GetWebhookStatus(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wl, err := s.GetWebhookLogByRequestID(r.Context(), chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, `{"error":"not found"}`, 404)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(wl)
	}
}
