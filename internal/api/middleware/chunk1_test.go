package middleware_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prashantluhar/testpay/internal/api/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCORS_allowlistedOrigin(t *testing.T) {
	mw := middleware.CORS([]string{"https://example.com"})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) }))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, "https://example.com", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "Origin", rec.Header().Get("Vary"))
}

func TestCORS_disallowedOriginLeavesHeaderEmpty(t *testing.T) {
	mw := middleware.CORS([]string{"https://example.com"})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) }))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://evil.tld")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_wildcardEchoesOrigin(t *testing.T) {
	mw := middleware.CORS([]string{"*"})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) }))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://anywhere.tld")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, "https://anywhere.tld", rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_emptyListAllowsAll(t *testing.T) {
	mw := middleware.CORS(nil)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) }))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://local.tld")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, "https://local.tld", rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_preflightShortCircuits(t *testing.T) {
	mw := middleware.CORS([]string{"https://example.com"})
	nextCalled := false
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { nextCalled = true }))
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.False(t, nextCalled, "preflight should not hit the next handler")
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Equal(t, "3600", rec.Header().Get("Access-Control-Max-Age"))
}

func TestGatewayResolver_setsGatewayFromPath(t *testing.T) {
	h := middleware.GatewayResolver(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "stripe", middleware.GetGateway(r))
		w.WriteHeader(200)
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/stripe/v1/charges", nil))
}

func TestGatewayResolver_skipsApiPrefix(t *testing.T) {
	h := middleware.GatewayResolver(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "agnostic", middleware.GetGateway(r))
		w.WriteHeader(200)
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/api/scenarios", nil))
}

func TestGetGateway_defaultsToAgnostic(t *testing.T) {
	// No middleware in the chain — context has no gateway key.
	r := httptest.NewRequest("GET", "/", nil)
	assert.Equal(t, "agnostic", middleware.GetGateway(r))
}

func TestReadBody_restoresForDownstreamReads(t *testing.T) {
	body := `{"amount":5000}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	got := middleware.ReadBody(r)
	assert.Equal(t, body, string(got))

	// Second read should work — body is restored.
	second, err := io.ReadAll(r.Body)
	require.NoError(t, err)
	assert.Equal(t, body, string(second))
}

func TestReadBody_nilBodyReturnsNil(t *testing.T) {
	r := &http.Request{}
	assert.Nil(t, middleware.ReadBody(r))
}

func TestCapture_recordsStatusBodyHeaders(t *testing.T) {
	captured := &middleware.CapturedResponse{}
	mw := middleware.Capture(captured)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("POST", "/", nil))
	assert.Equal(t, 201, captured.Status)
	assert.Equal(t, `{"ok":true}`, string(captured.Body))
	assert.Equal(t, "application/json", captured.Headers["Content-Type"])
}

func TestClearSessionCookie_hostedModeSetsSecureAndNone(t *testing.T) {
	rec := httptest.NewRecorder()
	middleware.ClearSessionCookie(rec, "hosted")
	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	c := cookies[0]
	assert.Equal(t, "testpay_session", c.Name)
	assert.True(t, c.Secure)
	assert.Equal(t, http.SameSiteNoneMode, c.SameSite)
	assert.Equal(t, -1, c.MaxAge)
}

func TestClearSessionCookie_localModeDefaults(t *testing.T) {
	rec := httptest.NewRecorder()
	middleware.ClearSessionCookie(rec, "local")
	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.False(t, cookies[0].Secure)
	assert.Equal(t, http.SameSiteLaxMode, cookies[0].SameSite)
}

func TestSetSessionCookie_hostedModeSetsSecureAndNone(t *testing.T) {
	rec := httptest.NewRecorder()
	middleware.SetSessionCookie(rec, "tok", "hosted")
	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.True(t, cookies[0].Secure)
	assert.Equal(t, http.SameSiteNoneMode, cookies[0].SameSite)
	assert.Equal(t, "tok", cookies[0].Value)
	assert.True(t, cookies[0].HttpOnly)
}

func TestParseToken_invalidSignatureErrors(t *testing.T) {
	good, err := middleware.IssueToken("secret-a", "user-1", "ws-1", time.Hour)
	require.NoError(t, err)
	_, err = middleware.ParseToken("secret-b", good)
	assert.Error(t, err)
}

func TestRequestID_echoesProvidedHeader(t *testing.T) {
	h := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "req-xyz", middleware.TraceID(r.Context()))
		w.WriteHeader(200)
	}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set(middleware.RequestIDHeader, "req-xyz")
	h.ServeHTTP(rec, req)
	assert.Equal(t, "req-xyz", rec.Header().Get(middleware.RequestIDHeader))
}

func TestRequestID_generatesWhenMissing(t *testing.T) {
	var generated string
	h := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		generated = middleware.TraceID(r.Context())
		w.WriteHeader(200)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	assert.NotEmpty(t, generated)
	assert.Equal(t, generated, rec.Header().Get(middleware.RequestIDHeader))
}

func TestTraceID_missingKeyReturnsEmpty(t *testing.T) {
	assert.Equal(t, "", middleware.TraceID(context.Background()))
}

func TestRateLimiter_allowsUntilExhaustedThen429s(t *testing.T) {
	rl := middleware.NewRateLimiter(60, 2, 0) // 60/min, burst 2
	h := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) }))

	send := func() int {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "203.0.113.1:9999"
		h.ServeHTTP(rec, req)
		return rec.Code
	}
	assert.Equal(t, 200, send())
	assert.Equal(t, 200, send())
	assert.Equal(t, 429, send(), "burst exhausted")
}

func TestRateLimiter_perIPBucketIsolated(t *testing.T) {
	rl := middleware.NewRateLimiter(60, 1, 0)
	h := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) }))

	send := func(ip string) int {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = ip + ":1234"
		h.ServeHTTP(rec, req)
		return rec.Code
	}
	assert.Equal(t, 200, send("10.0.0.1"))
	assert.Equal(t, 429, send("10.0.0.1"), "same IP burst 1 exhausted")
	assert.Equal(t, 200, send("10.0.0.2"), "different IP has its own bucket")
}

func TestRateLimiter_globalBucketTrumpsPerIP(t *testing.T) {
	// Per-IP generous (1000/min, burst 10), global tight (60/min = 1/s, burst 60).
	rl := middleware.NewRateLimiter(1000, 10, 60)
	h := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) }))

	accept, reject := 0, 0
	for i := 0; i < 100; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "10.0.0.9:1234"
		h.ServeHTTP(rec, req)
		if rec.Code == 200 {
			accept++
		} else {
			reject++
		}
	}
	assert.LessOrEqual(t, accept, 70, "global 60/min burst 60 caps well under 100")
	assert.Greater(t, reject, 0, "some requests should be rejected by the global bucket")
}

func TestRateLimiter_xForwardedForIsPreferred(t *testing.T) {
	rl := middleware.NewRateLimiter(60, 1, 0)
	h := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) }))

	send := func(xff string) int {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "10.0.0.1:1" // same proxy for every caller
		req.Header.Set("X-Forwarded-For", xff)
		h.ServeHTTP(rec, req)
		return rec.Code
	}
	assert.Equal(t, 200, send("198.51.100.1"))
	assert.Equal(t, 429, send("198.51.100.1"), "same XFF → same bucket")
	assert.Equal(t, 200, send("198.51.100.2, 10.0.0.1"), "multi-hop XFF, first hop is a different client")
}

func TestRateLimiter_disabledBothTiers(t *testing.T) {
	rl := middleware.NewRateLimiter(0, 0, 0)
	h := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) }))
	for i := 0; i < 50; i++ {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
		assert.Equal(t, 200, rec.Code)
	}
}
