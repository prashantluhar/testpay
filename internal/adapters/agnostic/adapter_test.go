package agnostic_test

import (
	"testing"

	"github.com/prashantluhar/testpay/internal/adapters/agnostic"
	"github.com/prashantluhar/testpay/internal/engine"
	"github.com/stretchr/testify/assert"
)

func TestBuildResponse_success(t *testing.T) {
	a := agnostic.New()
	result := &engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}
	status, body, _ := a.BuildResponse(result, nil)
	assert.Equal(t, 200, status)
	assert.Contains(t, string(body), `"status":"success"`)
}

func TestBuildWebhookPayload_success(t *testing.T) {
	a := agnostic.New()
	result := &engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}
	payload := a.BuildWebhookPayload(result, "txn_test", 5000, "usd")
	assert.Equal(t, "transaction.success", payload["event"])
}
