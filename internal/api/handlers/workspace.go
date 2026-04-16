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
// Currently only WebhookURLs (a gateway → URL map) is editable.
//
// Accepted body shapes:
//
//	{"webhook_urls": {"stripe": "https://...", "razorpay": "...", "agnostic": "..."}}
//
// Unknown gateway keys are stored as-is (for future adapter additions).
// Empty-string values clear the entry for that gateway.
func UpdateWorkspace(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := zerolog.Ctx(ctx)
		log.Info().Str("handler", "UpdateWorkspace").Str("step", "entry").Msg("handler entry")

		var body struct {
			WebhookURLs map[string]string `json:"webhook_urls"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			log.Error().Err(err).Str("step", "decode").Msg("invalid body")
			http.Error(w, `{"error":"invalid body"}`, 400)
			return
		}

		ws, err := workspaceFromCtx(r, s)
		if err != nil {
			log.Error().Err(err).Str("step", "workspace_lookup").Msg("workspace not found")
			http.Error(w, `{"error":"workspace not found"}`, 404)
			return
		}

		if body.WebhookURLs != nil {
			// Normalise: drop empty-string entries so they disappear from the map.
			clean := make(map[string]string, len(body.WebhookURLs))
			for k, v := range body.WebhookURLs {
				if v != "" {
					clean[k] = v
				}
			}
			ws.WebhookURLs = clean
		}

		if err := s.UpdateWorkspace(ctx, ws); err != nil {
			log.Error().Err(err).Str("step", "persist").Msg("update workspace failed")
			http.Error(w, `{"error":"update failed"}`, 500)
			return
		}

		log.Info().
			Str("workspace_id", ws.ID).
			Str("step", "exit").
			Interface("webhook_urls", ws.WebhookURLs).
			Msg("handler exit")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ws)
	}
}
