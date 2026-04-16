package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
)

const (
	SessionCookieName = "testpay_session"
	// LocalWorkspaceID mirrors store.LocalWorkspaceID; duplicated to avoid circular import.
	LocalWorkspaceID = "00000000-0000-0000-0000-000000000001"
)

type sessionCtxKey string

const (
	userIDKey      sessionCtxKey = "user_id"
	workspaceIDKey sessionCtxKey = "workspace_id"
)

type Claims struct {
	WorkspaceID string `json:"workspace_id"`
	jwt.RegisteredClaims
}

// IssueToken returns a signed JWT for the given user + workspace.
func IssueToken(secret, userID, workspaceID string, ttl time.Duration) (string, error) {
	claims := Claims{
		WorkspaceID: workspaceID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(secret))
}

// ParseToken validates and returns the claims or error.
func ParseToken(secret, raw string) (*Claims, error) {
	var c Claims
	_, err := jwt.ParseWithClaims(raw, &c, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// Session populates context with workspace_id (and user_id when available).
// - local mode: injects LocalWorkspaceID, no user
// - hosted mode: parses the cookie if present; invalid/missing is silent (RequireSession enforces)
func Session(secret, mode string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			if mode == "local" {
				ctx = context.WithValue(ctx, workspaceIDKey, LocalWorkspaceID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			cookie, err := r.Cookie(SessionCookieName)
			if err == nil {
				claims, perr := ParseToken(secret, cookie.Value)
				if perr == nil {
					ctx = context.WithValue(ctx, userIDKey, claims.Subject)
					ctx = context.WithValue(ctx, workspaceIDKey, claims.WorkspaceID)
				} else {
					zerolog.Ctx(ctx).Warn().Err(perr).Msg("invalid session cookie")
				}
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireSession writes a 401 if no user is in context and returns false.
// Call at the top of any handler that requires an authenticated user.
func RequireSession(w http.ResponseWriter, r *http.Request) bool {
	if _, ok := UserIDFromContext(r.Context()); ok {
		return true
	}
	// In local mode we still allow access (workspace is set), so we check for workspace too.
	if _, ok := WorkspaceIDFromContext(r.Context()); ok {
		// local mode — no user id but workspace is implied
		cookie, _ := r.Cookie(SessionCookieName)
		if cookie == nil {
			// no cookie set means this is probably local mode; allow
			return true
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
	return false
}

// UserIDFromContext returns the authenticated user id if present.
func UserIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(userIDKey).(string)
	return v, ok && v != ""
}

// WorkspaceIDFromContext returns the workspace id (local or authenticated).
func WorkspaceIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(workspaceIDKey).(string)
	return v, ok && v != ""
}

// SetSessionCookie writes the session cookie.
func SetSessionCookie(w http.ResponseWriter, token string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   30 * 24 * 60 * 60, // 30 days
	})
}

// ClearSessionCookie writes an expired cookie.
func ClearSessionCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
