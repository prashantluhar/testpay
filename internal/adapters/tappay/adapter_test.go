package tappay_test

import (
	"encoding/json"
	"testing"

	"github.com/prashantluhar/testpay/internal/adapters/tappay"
	"github.com/prashantluhar/testpay/internal/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdapter_Name(t *testing.T) {
	assert.Equal(t, "tappay", tappay.New().Name())
}

func TestBuildResponse_successShape(t *testing.T) {
	a := tappay.New()
	body := []byte(`{"amount":12345,"currency":"TWD"}`)
	status, raw, headers := a.BuildResponse(&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}, body)

	assert.Equal(t, 200, status)
	assert.Equal(t, "application/json", headers["Content-Type"])

	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.EqualValues(t, 0, resp["errorCode"])
	assert.Equal(t, "Success", resp["message"])
	assert.EqualValues(t, 12345, resp["amount"])
	assert.NotEmpty(t, resp["orderId"])
	assert.Contains(t, resp["paymentUrl"], "mock.tappay.io/pay/")
	assert.Equal(t, "mock-hmac-signature", resp["signature"])
}

func TestBuildResponse_hardFailureShape(t *testing.T) {
	a := tappay.New()
	status, raw, _ := a.BuildResponse(&engine.Result{
		Mode:       engine.ModeBankInvalidCVV,
		HTTPStatus: 402,
	}, nil)
	assert.Equal(t, 402, status)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.EqualValues(t, 3, resp["errorCode"], "CVV fail maps to errorCode 3")
	assert.Equal(t, "Invalid CVV", resp["message"])
	assert.Equal(t, "mock-hmac-signature", resp["signature"])
	assert.NotEmpty(t, resp["orderId"])
	// Error envelope must not carry paymentUrl (omitempty in the DTO).
	_, hasURL := resp["paymentUrl"]
	assert.False(t, hasURL, "error envelope should omit paymentUrl")
}

func TestBuildResponse_pendingMapsToProcessing(t *testing.T) {
	a := tappay.New()
	_, raw, _ := a.BuildResponse(&engine.Result{
		Mode:       engine.ModePendingThenSuccess,
		HTTPStatus: 200,
		IsPending:  true,
	}, nil)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.EqualValues(t, 1, resp["errorCode"])
	assert.Equal(t, "Processing", resp["message"])
}

func TestBuildResponse_defaultAmountWhenBodyEmpty(t *testing.T) {
	a := tappay.New()
	_, raw, _ := a.BuildResponse(&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}, nil)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.EqualValues(t, 5000, resp["amount"])
}

func TestBuildResponse_invalidJSONBodyFallsBackToDefaults(t *testing.T) {
	a := tappay.New()
	_, raw, _ := a.BuildResponse(&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}, []byte("{not json"))
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.EqualValues(t, 5000, resp["amount"])
}

func TestBuildWebhookPayload_successShape(t *testing.T) {
	a := tappay.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200},
		"TAP_42",
		9999,
		"TWD",
		map[string]any{"orderId": "my-order-42"},
	)

	assert.EqualValues(t, 0, out["errorCode"])
	assert.Equal(t, "Success", out["message"])
	assert.Equal(t, "TAP_42", out["orderId"])
	assert.Equal(t, "MOCK_PARTNER", out["partnerCode"])
	assert.Equal(t, "MOCK_API_KEY", out["apiKey"])
	assert.EqualValues(t, 9999, out["amount"])
	assert.Equal(t, "TWD", out["currency"])
	assert.Equal(t, "mock-hmac-signature", out["signature"])
	assert.NotEmpty(t, out["appotapayTransId"])
	assert.NotEmpty(t, out["transactionTs"])

	echo, ok := out["request_echo"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "my-order-42", echo["orderId"])
}

func TestBuildWebhookPayload_failureMapsCodes(t *testing.T) {
	a := tappay.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeBankDeclineSoft, HTTPStatus: 402},
		"TAP_FAIL",
		5000,
		"TWD",
		nil,
	)
	assert.EqualValues(t, 2, out["errorCode"])
	assert.Equal(t, "Insufficient funds", out["message"])
}

func TestBuildWebhookPayload_pendingMapsToProcessing(t *testing.T) {
	a := tappay.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModePendingThenSuccess, HTTPStatus: 200, IsPending: true},
		"TAP_PEND",
		5000,
		"TWD",
		nil,
	)
	assert.EqualValues(t, 1, out["errorCode"])
	assert.Equal(t, "Processing", out["message"])
}

func TestBuildWebhookPayload_networkErrorMapsTo98(t *testing.T) {
	a := tappay.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeNetworkError, HTTPStatus: 503},
		"TAP_NET",
		5000,
		"TWD",
		nil,
	)
	assert.EqualValues(t, 98, out["errorCode"])
	assert.Equal(t, "Network error", out["message"])
}
