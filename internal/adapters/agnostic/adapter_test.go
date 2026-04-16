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
	reqBody := map[string]any{"order_id": "agn_9", "amount": 5000}
	payload := a.BuildWebhookPayload(result, "txn_test", 5000, "usd", reqBody)
	assert.Equal(t, "transaction.success", payload["event"])
	echo := payload["request_echo"].(map[string]any)
	assert.Equal(t, "agn_9", echo["order_id"])
}
