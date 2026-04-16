package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/prashantluhar/testpay/internal/adapters"
	"github.com/rs/zerolog"
)

// ListGateways returns the set of canonical gateway names known to the server.
// The dashboard uses this to render per-gateway webhook configuration dynamically
// so the UI stays in sync with whatever adapters are registered.
func ListGateways(reg *adapters.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		zerolog.Ctx(r.Context()).Info().Str("handler", "ListGateways").Msg("handler entry")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(reg.KnownGateways())
	}
}
