package stripe_test

import (
	"testing"

	"github.com/prashantluhar/testpay/internal/adapters/stripe"
	"github.com/prashantluhar/testpay/internal/engine"
	"github.com/stretchr/testify/assert"
)

func TestBuildResponse_success(t *testing.T) {
	a := stripe.New()
	result := &engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}
	status, body, _ := a.BuildResponse(result, nil)
	assert.Equal(t, 200, status)
	assert.Contains(t, string(body), `"status":"succeeded"`)
}

func TestBuildResponse_decline(t *testing.T) {
	a := stripe.New()
	result := &engine.Result{Mode: engine.ModeBankDeclineHard, HTTPStatus: 402, ErrorCode: "card_declined"}
	status, body, _ := a.BuildResponse(result, nil)
	assert.Equal(t, 402, status)
	assert.Contains(t, string(body), `"code":"card_declined"`)
}

func TestBuildWebhookPayload_success(t *testing.T) {
	a := stripe.New()
	result := &engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}
	reqBody := map[string]any{"metadata": map[string]any{"order_id": "ord_42"}}
	payload := a.BuildWebhookPayload(result, "ch_test_123", 5000, "usd", reqBody)
	assert.Equal(t, "payment_intent.succeeded", payload["type"])
	obj := payload["data"].(map[string]any)["object"].(map[string]any)
	md := obj["metadata"].(map[string]any)
	assert.Equal(t, "ord_42", md["order_id"])
}
