package webhook_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prashantluhar/testpay/internal/webhook"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDispatch_success(t *testing.T) {
	received := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = true
		w.WriteHeader(200)
	}))
	defer srv.Close()

	d := webhook.NewDispatcher(3, 10*time.Millisecond)
	result, err := d.Dispatch(srv.URL, map[string]any{"type": "payment.success"})
	require.NoError(t, err)
	assert.True(t, received)
	assert.Equal(t, 200, result.StatusCode)
	assert.Equal(t, 1, result.Attempts)
}

func TestDispatch_retryOnFailure(t *testing.T) {
	attempt := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt < 3 {
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	d := webhook.NewDispatcher(3, 1*time.Millisecond)
	result, err := d.Dispatch(srv.URL, map[string]any{"event": "test"})
	require.NoError(t, err)
	assert.Equal(t, 3, result.Attempts)
	assert.Equal(t, 200, result.StatusCode)
}
