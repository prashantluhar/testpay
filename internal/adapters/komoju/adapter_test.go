package komoju_test

import (
	"encoding/json"
	"testing"

	"github.com/prashantluhar/testpay/internal/adapters/komoju"
	"github.com/prashantluhar/testpay/internal/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdapter_Name(t *testing.T) {
	assert.Equal(t, "komoju", komoju.New().Name())
}

func TestBuildResponse_successShape(t *testing.T) {
	a := komoju.New()
	body := []byte(`{"amount":3000,"currency":"JPY"}`)
	status, raw, headers := a.BuildResponse(&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}, body)

	assert.Equal(t, 200, status)
	assert.Equal(t, "application/json", headers["Content-Type"])

	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.NotEmpty(t, resp["id"])
	assert.Equal(t, "payment", resp["resource"])
	assert.Equal(t, "captured", resp["status"])
	assert.EqualValues(t, 3000, resp["amount"])
	assert.EqualValues(t, 3000, resp["total"])
	assert.Equal(t, "JPY", resp["currency"])

	pd, ok := resp["payment_details"].(map[string]any)
	require.True(t, ok, "payment_details must be an object")
	assert.Equal(t, "credit_card", pd["type"])
}

func TestBuildResponse_hardFailureShape(t *testing.T) {
	a := komoju.New()
	status, raw, _ := a.BuildResponse(&engine.Result{
		Mode:       engine.ModeBankInvalidCVV,
		HTTPStatus: 402,
	}, nil)
	assert.Equal(t, 402, status)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	errObj, ok := resp["error"].(map[string]any)
	require.True(t, ok, "failure envelope must wrap the error object")
	assert.Equal(t, "invalid_cvv", errObj["code"])
	assert.Equal(t, "Invalid security code.", errObj["message"])
}

func TestBuildResponse_pendingMapsToAuthorized(t *testing.T) {
	a := komoju.New()
	_, raw, _ := a.BuildResponse(&engine.Result{
		Mode:       engine.ModePendingThenSuccess,
		HTTPStatus: 200,
		IsPending:  true,
	}, nil)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.Equal(t, "authorized", resp["status"])
	// captured_at is null when not yet captured.
	assert.Nil(t, resp["captured_at"])
}

func TestBuildResponse_defaultAmountJPYWhenBodyEmpty(t *testing.T) {
	a := komoju.New()
	_, raw, _ := a.BuildResponse(&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}, nil)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.EqualValues(t, 5000, resp["amount"])
	assert.Equal(t, "JPY", resp["currency"])
}

func TestBuildResponse_invalidJSONBodyFallsBackToDefaults(t *testing.T) {
	a := komoju.New()
	_, raw, _ := a.BuildResponse(&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}, []byte("not json"))
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.EqualValues(t, 5000, resp["amount"])
}

func TestBuildWebhookPayload_successShape(t *testing.T) {
	a := komoju.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200},
		"komoju_42",
		5000,
		"JPY",
		map[string]any{"metadata": map[string]any{"order_id": "o-1"}},
	)
	assert.Equal(t, "event", out["resource"])
	assert.Equal(t, "payment.captured", out["type"])
	assert.NotEmpty(t, out["id"])
	assert.NotEmpty(t, out["created_at"])

	data, ok := out["data"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "komoju_42", data["id"])
	assert.Equal(t, "captured", data["status"])
	assert.EqualValues(t, 5000, data["amount"])
	assert.Equal(t, "JPY", data["currency"])

	meta, ok := data["metadata"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "o-1", meta["order_id"])

	// request_echo is carried for the dispatcher, not part of Komoju's wire format.
	assert.NotNil(t, out["request_echo"])
}

func TestBuildWebhookPayload_failureSetsPaymentFailedAndReason(t *testing.T) {
	a := komoju.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeBankDeclineSoft, HTTPStatus: 402},
		"komoju_fail",
		5000,
		"JPY",
		nil,
	)
	assert.Equal(t, "payment.failed", out["type"])
	assert.Equal(t, "Insufficient funds on card.", out["reason"])

	details, ok := out["details"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "insufficient_funds", details["error_code"])

	data := out["data"].(map[string]any)
	assert.Equal(t, "failed", data["status"])
}

func TestBuildWebhookPayload_reversalMapsToRefunded(t *testing.T) {
	a := komoju.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeSuccessThenReversed, HTTPStatus: 200},
		"komoju_rev",
		5000,
		"JPY",
		nil,
	)
	assert.Equal(t, "payment.refunded", out["type"])
	data := out["data"].(map[string]any)
	assert.Equal(t, "cancelled", data["status"])
}
