package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/prashantluhar/testpay/internal/store"
)

func CreateSession(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ScenarioID string `json:"scenario_id"`
			TTLSeconds int    `json:"ttl_seconds"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ScenarioID == "" {
			http.Error(w, `{"error":"scenario_id required"}`, 400)
			return
		}
		if req.TTLSeconds == 0 {
			req.TTLSeconds = 3600
		}
		sess := &store.Session{
			ID:          uuid.NewString(),
			WorkspaceID: store.LocalWorkspaceID,
			ScenarioID:  req.ScenarioID,
			TTLSeconds:  req.TTLSeconds,
			ExpiresAt:   time.Now().Add(time.Duration(req.TTLSeconds) * time.Second),
		}
		if err := s.CreateSession(r.Context(), sess); err != nil {
			http.Error(w, `{"error":"failed to create session"}`, 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(sess)
	}
}

func DeleteSession(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.DeleteSession(r.Context(), chi.URLParam(r, "id"))
		w.WriteHeader(204)
	}
}
