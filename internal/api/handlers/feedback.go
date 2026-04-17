package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/prashantluhar/testpay/internal/api/middleware"
	"github.com/prashantluhar/testpay/internal/store"
	"github.com/rs/zerolog"
)

type feedbackRequest struct {
	WhatTried string `json:"what_tried"`
	Worked    string `json:"worked"`
	Missing   string `json:"missing"`
	Email     string `json:"email"`
	PageURL   string `json:"page_url"`
}

// Feedback accepts submissions from the in-app widget. Public endpoint —
// anonymous docs visitors can submit too. Writes to feedback_submissions
// in Postgres; identity fields (workspace_id, user_id) populate when a
// session cookie is present, otherwise nil.
//
// Size-cap each free-text field so a noisy caller can't balloon the DB.
func Feedback(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := zerolog.Ctx(ctx)

		var req feedbackRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid body"}`, 400)
			return
		}
		// At least one of the text fields must be non-empty — otherwise
		// we're just logging empty rows.
		if strings.TrimSpace(req.WhatTried+req.Worked+req.Missing) == "" {
			http.Error(w, `{"error":"at least one field required"}`, 400)
			return
		}

		fb := &store.Feedback{
			ID:        uuid.NewString(),
			WhatTried: cap(req.WhatTried, 2000),
			Worked:    cap(req.Worked, 2000),
			Missing:   cap(req.Missing, 2000),
			Email:     cap(req.Email, 256),
			PageURL:   cap(req.PageURL, 512),
			UserAgent: cap(r.UserAgent(), 512),
		}
		if wsID, ok := middleware.WorkspaceIDFromContext(ctx); ok && wsID != "" {
			fb.WorkspaceID = &wsID
		}
		if userID, ok := middleware.UserIDFromContext(ctx); ok && userID != "" {
			fb.UserID = &userID
		}

		if err := s.CreateFeedback(ctx, fb); err != nil {
			log.Error().Err(err).Str("handler", "Feedback").Msg("store error")
			http.Error(w, `{"error":"failed to record feedback"}`, 500)
			return
		}
		log.Info().Str("handler", "Feedback").Str("id", fb.ID).Bool("email_present", fb.Email != "").Msg("feedback submitted")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		_ = json.NewEncoder(w).Encode(map[string]string{"id": fb.ID})
	}
}

func cap(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
