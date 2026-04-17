package handlers

import (
	"encoding/json"
	"net/http"
)

// Healthz returns a minimal liveness probe suitable for uptime pingers.
//
// Keep this handler side-effect-free: no DB work, no store calls, no logs
// on the happy path. It runs on every 5-minute uptime ping and we don't
// want the logs table polluted with thousands of probe entries.
//
// Returns 200 OK with a tiny JSON body so observability tools that parse
// the response can confirm they spoke to the Go binary (not an upstream
// proxy or CDN that happens to return 200 on anything).
func Healthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "ok",
			"service": "testpay",
		})
	}
}
