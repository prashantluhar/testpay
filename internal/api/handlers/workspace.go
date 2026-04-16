package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/prashantluhar/testpay/internal/api/middleware"
	"github.com/prashantluhar/testpay/internal/store"
	"github.com/rs/zerolog"
)

// workspaceFromCtx resolves the current workspace. Uses the session's
// workspace_id if available, otherwise falls back to the local workspace.
func workspaceFromCtx(r *http.Request, s store.Store) (*store.Workspace, error) {
	ctx := r.Context()
	if id, ok := middleware.WorkspaceIDFromContext(ctx); ok && id != "" {
		return s.GetWorkspaceByID(ctx, id)
	}
	return s.GetWorkspaceBySlug(ctx, "local")
}

func GetWorkspace(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		zerolog.Ctx(ctx).Info().Str("handler", "GetWorkspace").Msg("handler entry")
		ws, err := workspaceFromCtx(r, s)
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

// UpdateWorkspace lets the authenticated user edit mutable workspace fields.
// Currently only WebhookURL is editable.
func UpdateWorkspace(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		zerolog.Ctx(ctx).Info().Str("handler", "UpdateWorkspace").Msg("handler entry")

		var body struct {
			WebhookURL string `json:"webhook_url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("invalid body")
			http.Error(w, `{"error":"invalid body"}`, 400)
			return
		}

		ws, err := workspaceFromCtx(r, s)
		if err != nil {
			http.Error(w, `{"error":"workspace not found"}`, 404)
			return
		}
		ws.WebhookURL = body.WebhookURL

		if err := s.UpdateWorkspace(ctx, ws); err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("update workspace failed")
			http.Error(w, `{"error":"update failed"}`, 500)
			return
		}

		zerolog.Ctx(ctx).Info().Str("workspace_id", ws.ID).Msg("handler exit")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ws)
	}
}
