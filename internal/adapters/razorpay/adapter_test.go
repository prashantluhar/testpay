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
	payload := a.BuildWebhookPayload(result, "pay_test", 5000, "INR")
	assert.Equal(t, "payment.captured", payload["event"])
}
