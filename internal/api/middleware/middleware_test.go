package middleware_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prashantluhar/testpay/internal/api/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestID_storesInContextAndHeader(t *testing.T) {
	var ctxID string
	handler := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxID = middleware.TraceID(r.Context())
		w.WriteHeader(200)
	}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	assert.NotEmpty(t, ctxID)
	assert.Equal(t, ctxID, rec.Header().Get("X-Request-ID"))
}

func TestRequestID_honoursIncomingHeader(t *testing.T) {
	var ctxID string
	handler := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxID = middleware.TraceID(r.Context())
	}))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", "incoming-trace-123")
	handler.ServeHTTP(httptest.NewRecorder(), req)
	assert.Equal(t, "incoming-trace-123", ctxID)
}

func TestLogger_injectsTraceIDIntoContextLogger(t *testing.T) {
	var buf bytes.Buffer
	prev := log.Logger
	log.Logger = zerolog.New(&buf)
	defer func() { log.Logger = prev }()

	chain := middleware.RequestID(middleware.Logger("test", "testpay")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Ctx(r.Context()).Info().Msg("inside handler")
		w.WriteHeader(200)
	})))
	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, httptest.NewRequest("GET", "/some/path", nil))

	// Find the "inside handler" log line
	var foundHandlerLog bool
	for _, line := range bytes.Split(buf.Bytes(), []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		var entry map[string]any
		require.NoError(t, json.Unmarshal(line, &entry))
		if entry["message"] == "inside handler" {
			foundHandlerLog = true
			assert.NotEmpty(t, entry["trace_id"], "handler log must include trace_id")
			assert.Equal(t, "test", entry["env"])
			assert.Equal(t, "testpay", entry["service"])
		}
	}
	assert.True(t, foundHandlerLog, "expected to find handler log line")
}

func TestLogger_emitsEdgeRequestLog(t *testing.T) {
	var buf bytes.Buffer
	prev := log.Logger
	log.Logger = zerolog.New(&buf)
	defer func() { log.Logger = prev }()

	chain := middleware.RequestID(middleware.Logger("prod", "testpay")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
	})))
	chain.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("POST", "/stripe/v1/charges", nil))

	var found bool
	for _, line := range bytes.Split(buf.Bytes(), []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		var e map[string]any
		json.Unmarshal(line, &e)
		if e["message"] == "request completed" {
			found = true
			assert.Equal(t, "POST", e["method"])
			assert.Equal(t, "/stripe/v1/charges", e["path"])
			assert.EqualValues(t, 201, e["status"])
			assert.NotEmpty(t, e["trace_id"])
			assert.NotEmpty(t, e["duration_ms"])
		}
	}
	assert.True(t, found, "expected edge 'request completed' log")
}

func TestAuth_localModePasses(t *testing.T) {
	handler := middleware.Auth("local", "any_key")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	assert.Equal(t, 200, rec.Code)
}

func TestAuth_hostedModeRejectsNoKey(t *testing.T) {
	handler := middleware.Auth("hosted", "secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	assert.Equal(t, 401, rec.Code)
}

func TestAuth_hostedModeAcceptsValidKey(t *testing.T) {
	handler := middleware.Auth("hosted", "secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer secret")
	handler.ServeHTTP(rec, req)
	assert.Equal(t, 200, rec.Code)
}
