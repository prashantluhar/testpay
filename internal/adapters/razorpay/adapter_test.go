package razorpay_test

import (
	"testing"

	"github.com/prashantluhar/testpay/internal/adapters/razorpay"
	"github.com/prashantluhar/testpay/internal/engine"
	"github.com/stretchr/testify/assert"
)

func TestBuildResponse_success(t *testing.T) {
	a := razorpay.New()
	result := &engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}
	status, body, _ := a.BuildResponse(result, nil)
	assert.Equal(t, 200, status)
	assert.Contains(t, string(body), `"status":"captured"`)
}

func TestBuildWebhookPayload_success(t *testing.T) {
	a := razorpay.New()
	result := &engine.Result{Mode: engine.ModeSuccess}
	reqBody := map[string]any{"notes": map[string]any{"order_id": "rzp_1"}}
	payload := a.BuildWebhookPayload(result, "pay_test", 5000, "INR", reqBody)
	assert.Equal(t, "payment.captured", payload["event"])
	entity := payload["payload"].(map[string]any)["payment"].(map[string]any)["entity"].(map[string]any)
	notes := entity["notes"].(map[string]any)
	assert.Equal(t, "rzp_1", notes["order_id"])
}
