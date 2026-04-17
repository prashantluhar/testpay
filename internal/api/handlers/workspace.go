package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/prashantluhar/testpay/internal/api/middleware"
	"github.com/prashantluhar/testpay/internal/store"
	"github.com/rs/zerolog"
)

// WorkspaceFromRequest resolves the workspace the current request should act on.
// Priority:
//  1. Authorization: Bearer <api_key>        → lookup by api_key
//  2. Session cookie (Session middleware injects workspace_id into ctx) → by id
//  3. Local fallback: the "local" workspace
//
// Exposed so other handlers can scope queries by caller's workspace.
func WorkspaceFromRequest(r *http.Request, s store.Store) (*store.Workspace, error) {
	ctx := r.Context()

	if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		apiKey := strings.TrimPrefix(auth, "Bearer ")
		if apiKey != "" {
			if ws, err := s.GetWorkspaceByAPIKey(ctx, apiKey); err == nil && ws != nil {
				return ws, nil
			}
		}
	}
	if id, ok := middleware.WorkspaceIDFromContext(ctx); ok && id != "" {
		return s.GetWorkspaceByID(ctx, id)
	}
	return s.GetWorkspaceBySlug(ctx, "local")
}

// WorkspaceIDFromRequest is a thin convenience wrapper returning just the ID.
// Falls back to LocalWorkspaceID if lookup fails.
func WorkspaceIDFromRequest(r *http.Request, s store.Store) string {
	if ws, err := WorkspaceFromRequest(r, s); err == nil && ws != nil {
		return ws.ID
	}
	return store.LocalWorkspaceID
}

// workspaceFromCtx is kept as a package-local shim for older callers; prefer
// WorkspaceFromRequest in new code.
func workspaceFromCtx(r *http.Request, s store.Store) (*store.Workspace, error) {
	return WorkspaceFromRequest(r, s)
}

// ensureLocalWorkspaceContext is used by handlers that historically hardcoded
// LocalWorkspaceID; left as a tiny helper so we can audit callers.
func ensureLocalWorkspaceContext(ctx context.Context) context.Context { return ctx }

func GetWorkspace(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		zerolog.Ctx(ctx).Info().Str("handler", "GetWorkspace").Msg("handler entry")
		ws, err := WorkspaceFromRequest(r, s)
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

// WorkspaceUsage returns the current workspace's 24h request count and its
// configured daily cap. Frontend polls this to render the quota pill. Cheap —
// one indexed COUNT on request_logs.
//
// Response shape:
//
//	{ "used_today": 47, "cap": 200 }
//
// `cap` is nil when the workspace is unlimited (standard signed-up case).
// Frontend hides the pill entirely in that case.
func WorkspaceUsage(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := zerolog.Ctx(ctx)

		ws, err := WorkspaceFromRequest(r, s)
		if err != nil {
			http.Error(w, `{"error":"workspace not found"}`, 404)
			return
		}
		dayAgo := time.Now().Add(-24 * time.Hour)
		used, cerr := s.CountRequestsSince(ctx, ws.ID, dayAgo)
		if cerr != nil {
			log.Error().Err(cerr).Str("handler", "WorkspaceUsage").Msg("count failed")
			http.Error(w, `{"error":"usage lookup failed"}`, 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"used_today": used,
			"cap":        ws.MaxDailyRequests, // *int — marshals as null if nil
		})
	}
}

// UpdateWorkspace lets the authenticated caller edit mutable workspace fields.
// Currently only WebhookURLs (a gateway → URL map) is editable.
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

		ws, err := WorkspaceFromRequest(r, s)
		if err != nil {
			log.Error().Err(err).Str("step", "workspace_lookup").Msg("workspace not found")
			http.Error(w, `{"error":"workspace not found"}`, 404)
			return
		}

		if body.WebhookURLs != nil {
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
