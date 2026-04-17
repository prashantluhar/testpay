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

// Session populates context with user_id and workspace_id if the session
// cookie is valid. No anonymous fallback — dashboard access is always
// login-gated. Mock endpoints resolve their workspace via api_key Bearer auth.
func Session(secret, mode string) func(http.Handler) http.Handler {
	_ = mode // retained for API compat; not used anymore
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
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

// RequireAuth is middleware that 401s if no authenticated session is present.
// Apply to dashboard /api routes (not /api/auth/*). Mock endpoints do their own
// api_key-based auth and should not use this.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := UserIDFromContext(r.Context()); !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"login required"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
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

// cookieAttrsFor returns the Secure and SameSite attributes appropriate for the
// given server mode. Hosted deploys sit behind a TLS-terminating proxy and are
// accessed from a different subdomain than the dashboard, so the browser
// requires SameSite=None with Secure=true for the cookie to be sent on XHR.
func cookieAttrsFor(mode string) (secure bool, sameSite http.SameSite) {
	if mode == "hosted" {
		return true, http.SameSiteNoneMode
	}
	return false, http.SameSiteLaxMode
}

// SetSessionCookie writes the session cookie.
func SetSessionCookie(w http.ResponseWriter, token, mode string) {
	secure, sameSite := cookieAttrsFor(mode)
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		MaxAge:   30 * 24 * 60 * 60, // 30 days
	})
}

// ClearSessionCookie writes an expired cookie.
func ClearSessionCookie(w http.ResponseWriter, mode string) {
	secure, sameSite := cookieAttrsFor(mode)
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		MaxAge:   -1,
	})
}
