package webhook_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prashantluhar/testpay/internal/store"
	"github.com/prashantluhar/testpay/internal/webhook"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeWebhookStore is a minimal store.Store that only implements
// UpdateWebhookLog (the one method DispatchAsync uses). Everything else
// is the embedded interface — calls panic if touched.
type fakeWebhookStore struct {
	store.Store
	mu      sync.Mutex
	updated *store.WebhookLog
	calls   int32
}

func (f *fakeWebhookStore) UpdateWebhookLog(_ context.Context, l *store.WebhookLog) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	atomic.AddInt32(&f.calls, 1)
	f.updated = l
	return nil
}

func TestDispatch_exhaustsAllAttemptsAndReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(500)
	}))
	defer srv.Close()

	d := webhook.NewDispatcher(3, 1*time.Millisecond)
	result, err := d.Dispatch(srv.URL, map[string]any{"type": "fail"})
	assert.Error(t, err)
	assert.Equal(t, 3, result.Attempts)
	assert.Equal(t, 500, result.StatusCode)
	assert.Len(t, result.AttemptLogs, 3)
}

func TestDispatch_networkErrorRecordsZeroStatus(t *testing.T) {
	d := webhook.NewDispatcher(1, 1*time.Millisecond)
	// Refused-connection URL — reserved port
	result, err := d.Dispatch("http://127.0.0.1:1/nope", map[string]any{"type": "x"})
	assert.Error(t, err)
	require.Len(t, result.AttemptLogs, 1)
	assert.Equal(t, 0, result.AttemptLogs[0].Status)
}

func TestDispatch_unmarshalablePayloadErrors(t *testing.T) {
	d := webhook.NewDispatcher(1, 1*time.Millisecond)
	// channels can't be JSON-encoded
	_, err := d.Dispatch("http://localhost/", map[string]any{"ch": make(chan int)})
	assert.Error(t, err)
}

func TestDispatchAsync_delivered(t *testing.T) {
	delivered := make(chan struct{}, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		delivered <- struct{}{}
	}))
	defer srv.Close()

	fake := &fakeWebhookStore{}
	d := webhook.NewDispatcher(1, 1*time.Millisecond)
	wl := &store.WebhookLog{
		ID:        "w1",
		TargetURL: srv.URL,
		Payload:   map[string]any{"event": "ok"},
	}

	webhook.DispatchAsync(context.Background(), d, fake, wl, 0)

	select {
	case <-delivered:
	case <-time.After(2 * time.Second):
		t.Fatal("webhook was never delivered")
	}

	// DispatchAsync updates the log asynchronously after the HTTP round-trip.
	require.Eventually(t, func() bool {
		fake.mu.Lock()
		defer fake.mu.Unlock()
		return fake.updated != nil && fake.updated.DeliveryStatus == "delivered"
	}, 2*time.Second, 10*time.Millisecond)
	assert.Equal(t, 1, fake.updated.Attempts)
	assert.NotNil(t, fake.updated.DeliveredAt)
}

func TestDispatchAsync_failedMarksFailedStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(500)
	}))
	defer srv.Close()

	fake := &fakeWebhookStore{}
	d := webhook.NewDispatcher(2, 1*time.Millisecond)
	wl := &store.WebhookLog{ID: "w2", TargetURL: srv.URL, Payload: map[string]any{"event": "bad"}}

	webhook.DispatchAsync(context.Background(), d, fake, wl, 0)

	require.Eventually(t, func() bool {
		fake.mu.Lock()
		defer fake.mu.Unlock()
		return fake.updated != nil && fake.updated.DeliveryStatus == "failed"
	}, 2*time.Second, 10*time.Millisecond)
	assert.Equal(t, 2, fake.updated.Attempts)
	assert.Len(t, fake.updated.AttemptLogs, 2)
}

func TestDispatchAsync_delayRespected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()

	fake := &fakeWebhookStore{}
	d := webhook.NewDispatcher(1, 1*time.Millisecond)
	wl := &store.WebhookLog{ID: "w3", TargetURL: srv.URL, Payload: map[string]any{"delay": true}}

	start := time.Now()
	webhook.DispatchAsync(context.Background(), d, fake, wl, 100)
	require.Eventually(t, func() bool {
		fake.mu.Lock()
		defer fake.mu.Unlock()
		return fake.updated != nil && fake.updated.DeliveryStatus == "delivered"
	}, 2*time.Second, 10*time.Millisecond)
	elapsed := time.Since(start)
	assert.GreaterOrEqual(t, elapsed, 100*time.Millisecond, "delay should have been honored")
}
