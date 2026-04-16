package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/prashantluhar/testpay/internal/store"
	"github.com/rs/zerolog"
)

func CreateSession(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		zerolog.Ctx(ctx).Info().Str("handler", "CreateSession").Msg("handler entry")
		var req struct {
			ScenarioID string `json:"scenario_id"`
			TTLSeconds int    `json:"ttl_seconds"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ScenarioID == "" {
			zerolog.Ctx(ctx).Error().Err(err).Str("handler", "CreateSession").Msg("scenario_id required")
			http.Error(w, `{"error":"scenario_id required"}`, 400)
			return
		}
		if req.TTLSeconds == 0 {
			req.TTLSeconds = 3600
		}
		sess := &store.Session{
			ID:          uuid.NewString(),
			WorkspaceID: WorkspaceIDFromRequest(r, s),
			ScenarioID:  req.ScenarioID,
			TTLSeconds:  req.TTLSeconds,
			ExpiresAt:   time.Now().Add(time.Duration(req.TTLSeconds) * time.Second),
		}
		if err := s.CreateSession(ctx, sess); err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Str("handler", "CreateSession").Str("session_id", sess.ID).Msg("store error")
			http.Error(w, `{"error":"failed to create session"}`, 500)
			return
		}
		zerolog.Ctx(ctx).Info().Str("handler", "CreateSession").Str("session_id", sess.ID).Str("scenario_id", sess.ScenarioID).Int("ttl_seconds", sess.TTLSeconds).Msg("handler exit")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(sess)
	}
}

func DeleteSession(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := chi.URLParam(r, "id")
		zerolog.Ctx(ctx).Info().Str("handler", "DeleteSession").Str("session_id", id).Msg("handler entry")
		if err := s.DeleteSession(ctx, id); err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Str("handler", "DeleteSession").Str("session_id", id).Msg("store error")
		}
		zerolog.Ctx(ctx).Info().Str("handler", "DeleteSession").Str("session_id", id).Msg("handler exit")
		w.WriteHeader(204)
	}
}
