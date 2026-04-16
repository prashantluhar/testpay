package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/prashantluhar/testpay/internal/store"
	"github.com/rs/zerolog"
)

func GetWorkspace(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		zerolog.Ctx(ctx).Info().Str("handler", "GetWorkspace").Msg("handler entry")
		ws, err := s.GetWorkspaceBySlug(ctx, "local")
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Str("handler", "GetWorkspace").Msg("workspace not found")
			http.Error(w, `{"error":"workspace not found"}`, 404)
			return
		}
		zerolog.Ctx(ctx).Info().Str("handler", "GetWorkspace").Str("workspace_id", ws.ID).Msg("handler exit")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ws)
	}
}
