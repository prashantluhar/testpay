package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/prashantluhar/testpay/internal/api/middleware"
	"github.com/prashantluhar/testpay/internal/store"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	User      *store.User      `json:"user"`
	Workspace *store.Workspace `json:"workspace"`
}

const (
	bcryptCost     = 12
	sessionTTL     = 30 * 24 * time.Hour
	minPasswordLen = 8
)

func Signup(s store.Store, jwtSecret, mode string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		zerolog.Ctx(ctx).Info().Str("handler", "Signup").Msg("handler entry")

		var req authRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			authError(w, 400, "invalid body")
			return
		}
		email := strings.ToLower(strings.TrimSpace(req.Email))
		if email == "" || !strings.Contains(email, "@") {
			authError(w, 400, "valid email required")
			return
		}
		if len(req.Password) < minPasswordLen {
			authError(w, 400, fmt.Sprintf("password must be at least %d characters", minPasswordLen))
			return
		}

		if existing, _, _ := s.GetUserByEmail(ctx, email); existing != nil {
			zerolog.Ctx(ctx).Warn().Str("email", email).Msg("signup with existing email")
			authError(w, 409, "email already registered")
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("bcrypt hash failed")
			authError(w, 500, "internal error")
			return
		}

		apiKey, err := randomKey(32)
		if err != nil {
			authError(w, 500, "internal error")
			return
		}

		slug := uniqueSlug(ctx, s, email)

		ws := &store.Workspace{
			ID:     uuid.NewString(),
			Slug:   slug,
			APIKey: apiKey,
		}
		if err := s.CreateWorkspace(ctx, ws); err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("create workspace failed")
			authError(w, 500, "failed to create workspace")
			return
		}

		u := &store.User{
			ID:          uuid.NewString(),
			WorkspaceID: ws.ID,
			Email:       email,
			Role:        "owner",
		}
		if err := s.CreateUser(ctx, u, string(hash)); err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("create user failed")
			authError(w, 500, "failed to create user")
			return
		}

		token, err := middleware.IssueToken(jwtSecret, u.ID, ws.ID, sessionTTL)
		if err != nil {
			authError(w, 500, "failed to issue session")
			return
		}
		middleware.SetSessionCookie(w, token, mode)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(authResponse{User: u, Workspace: ws})
	}
}

func Login(s store.Store, jwtSecret, mode string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		zerolog.Ctx(ctx).Info().Str("handler", "Login").Msg("handler entry")

		var req authRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			authError(w, 400, "invalid body")
			return
		}
		email := strings.ToLower(strings.TrimSpace(req.Email))

		u, hash, err := s.GetUserByEmail(ctx, email)
		if err != nil || u == nil {
			zerolog.Ctx(ctx).Warn().Str("email", email).Msg("login: user not found")
			authError(w, 401, "invalid credentials")
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
			zerolog.Ctx(ctx).Warn().Str("email", email).Msg("login: bad password")
			authError(w, 401, "invalid credentials")
			return
		}

		_ = s.UpdateUserLastLogin(ctx, u.ID, time.Now())

		token, err := middleware.IssueToken(jwtSecret, u.ID, u.WorkspaceID, sessionTTL)
		if err != nil {
			authError(w, 500, "failed to issue session")
			return
		}
		middleware.SetSessionCookie(w, token, mode)

		ws, _ := getWorkspaceByID(ctx, s, u.WorkspaceID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(authResponse{User: u, Workspace: ws})
	}
}

func Logout(mode string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		zerolog.Ctx(r.Context()).Info().Str("handler", "Logout").Msg("handler entry")
		middleware.ClearSessionCookie(w, mode)
		w.WriteHeader(204)
	}
}

func Me(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		zerolog.Ctx(ctx).Info().Str("handler", "Me").Msg("handler entry")

		userID, hasUser := middleware.UserIDFromContext(ctx)
		wsID, hasWs := middleware.WorkspaceIDFromContext(ctx)

		// Local mode: no user, workspace is LocalWorkspaceID
		if !hasUser && hasWs && wsID == middleware.LocalWorkspaceID {
			ws, err := s.GetWorkspaceBySlug(ctx, "local")
			if err != nil {
				authError(w, 404, "workspace not found")
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(authResponse{User: nil, Workspace: ws})
			return
		}

		if !hasUser {
			authError(w, 401, "unauthorized")
			return
		}

		u, err := s.GetUserByID(ctx, userID)
		if err != nil {
			authError(w, 401, "unauthorized")
			return
		}
		ws, err := getWorkspaceByID(ctx, s, u.WorkspaceID)
		if err != nil {
			authError(w, 500, "workspace lookup failed")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(authResponse{User: u, Workspace: ws})
	}
}

// ---- helpers ---------------------------------------------------------------

func authError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func randomKey(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// uniqueSlug returns a slug derived from the email's local-part plus a 3-char suffix.
func uniqueSlug(ctx context.Context, s store.Store, email string) string {
	base := strings.Split(email, "@")[0]
	// keep it simple: lowercase + hyphenate underscores/dots
	base = strings.ReplaceAll(strings.ReplaceAll(base, "_", "-"), ".", "-")
	for i := 0; i < 5; i++ {
		suffix, _ := randomKey(2) // 4 hex chars; plenty of room
		slug := fmt.Sprintf("%s-%s", base, suffix[:4])
		if _, err := s.GetWorkspaceBySlug(ctx, slug); err != nil {
			return slug
		}
	}
	// fallback: uuid-based
	return uuid.NewString()
}

// getWorkspaceByID is a small helper since store.Store does not have a GetByID
// method for workspaces (GetWorkspaceBySlug / GetWorkspaceByAPIKey are the
// existing lookups). Reuse slug lookup via a cache-free tour — see note in plan.
// For simplicity: we look up by the slug we know. In MVP hosted mode, we
// instead use GetWorkspaceByAPIKey when available, but here we pull from a
// workspace lookup by slug if slug is known. Since Login has no slug handy,
// we fetch by scanning the api_key we stored. Simpler: assume callers pass
// us the workspace id via context and use a dedicated query below.
//
// To avoid a schema change right now, implement a small helper that runs a
// direct GetWorkspaceBySlug lookup using the user's email's local-part —
// BUT that fails on already-customized slugs. Therefore we add a thin
// postgres method in the impl; here we hide it behind a helper.
//
// Implementation below uses a type assertion to the concrete postgres store.
func getWorkspaceByID(ctx context.Context, s store.Store, id string) (*store.Workspace, error) {
	type byIDer interface {
		GetWorkspaceByID(ctx context.Context, id string) (*store.Workspace, error)
	}
	if b, ok := s.(byIDer); ok {
		return b.GetWorkspaceByID(ctx, id)
	}
	return nil, errors.New("store does not support GetWorkspaceByID")
}
