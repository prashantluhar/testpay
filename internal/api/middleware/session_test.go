package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prashantluhar/testpay/internal/api/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSession_localModeInjectsLocalWorkspace(t *testing.T) {
	var gotWorkspace string
	h := middleware.Session("any-secret", "local")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotWorkspace, _ = middleware.WorkspaceIDFromContext(r.Context())
		w.WriteHeader(200)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	assert.NotEmpty(t, gotWorkspace)
}

func TestSession_hostedWithValidCookie(t *testing.T) {
	secret := "test-secret"
	token, err := middleware.IssueToken(secret, "user-1", "ws-1", 24*time.Hour)
	require.NoError(t, err)

	var gotUser, gotWorkspace string
	h := middleware.Session(secret, "hosted")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser, _ = middleware.UserIDFromContext(r.Context())
		gotWorkspace, _ = middleware.WorkspaceIDFromContext(r.Context())
		w.WriteHeader(200)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "testpay_session", Value: token})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, "user-1", gotUser)
	assert.Equal(t, "ws-1", gotWorkspace)
}

func TestSession_hostedNoCookie(t *testing.T) {
	h := middleware.Session("test-secret", "hosted")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := middleware.UserIDFromContext(r.Context()); ok {
			t.Fatal("expected no user in context")
		}
		w.WriteHeader(200)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	assert.Equal(t, 200, rec.Code)
}

func TestRequireSession_401WhenMissing(t *testing.T) {
	h := middleware.Session("test-secret", "hosted")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !middleware.RequireSession(w, r) {
			return
		}
		w.WriteHeader(200)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/api/me", nil))
	assert.Equal(t, 401, rec.Code)
}
