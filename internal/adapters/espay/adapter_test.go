package espay_test

import (
	"encoding/json"
	"testing"

	"github.com/prashantluhar/testpay/internal/adapters/espay"
	"github.com/prashantluhar/testpay/internal/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdapter_Name(t *testing.T) {
	assert.Equal(t, "espay", espay.New().Name())
}

func TestBuildResponse_successShape(t *testing.T) {
	a := espay.New()
	body := []byte(`{"amount":12345,"ccy":"IDR"}`)
	status, raw, headers := a.BuildResponse(&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}, body)

	assert.Equal(t, 200, status)
	assert.Equal(t, "application/json", headers["Content-Type"])

	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.Equal(t, "0000", resp["error_code"], "ESPay success MUST be '0000'")
	assert.Equal(t, "Success", resp["error_message"])
	assert.Equal(t, "mock-rsa-signature", resp["signature"])
	assert.Equal(t, "12345", resp["amount"], "amount is string-encoded on ESPay wire")
	assert.Equal(t, "IDR", resp["ccy"])
	assert.NotEmpty(t, resp["order_id"])
	assert.NotEmpty(t, resp["rq_uuid"])

	cust, ok := resp["customer_details"].(map[string]any)
	require.True(t, ok, "customer_details must be an object")
	assert.NotEmpty(t, cust["email"])
}

func TestBuildResponse_hardFailureShape(t *testing.T) {
	a := espay.New()
	status, raw, _ := a.BuildResponse(&engine.Result{
		Mode:       engine.ModeBankInvalidCVV,
		HTTPStatus: 402,
	}, nil)
	assert.Equal(t, 402, status)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.Equal(t, "ESP02", resp["error_code"])
	assert.Equal(t, "Invalid CVV", resp["error_message"])
	assert.Equal(t, "mock-rsa-signature", resp["signature"])
	assert.NotEmpty(t, resp["order_id"])
}

func TestBuildResponse_pendingUsesSoftCode(t *testing.T) {
	a := espay.New()
	_, raw, _ := a.BuildResponse(&engine.Result{
		Mode:       engine.ModePendingThenSuccess,
		HTTPStatus: 200,
		IsPending:  true,
	}, nil)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.Equal(t, "0001", resp["error_code"])
	assert.Equal(t, "Processing", resp["error_message"])
}

func TestBuildResponse_defaultAmountIDRWhenBodyEmpty(t *testing.T) {
	a := espay.New()
	_, raw, _ := a.BuildResponse(&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}, nil)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.Equal(t, "5000", resp["amount"])
	assert.Equal(t, "IDR", resp["ccy"])
}

func TestBuildResponse_invalidJSONBodyFallsBackToDefaults(t *testing.T) {
	a := espay.New()
	_, raw, _ := a.BuildResponse(&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}, []byte("not json"))
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.Equal(t, "5000", resp["amount"])
	assert.Equal(t, "IDR", resp["ccy"])
}

func TestBuildWebhookPayload_successShape(t *testing.T) {
	a := espay.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200},
		"ESP_ORDER_42",
		7777,
		"IDR",
		map[string]any{"merchant_ref": "m-1"},
	)

	assert.Equal(t, "0000", out["error_code"])
	assert.Equal(t, "Success", out["error_message"])
	assert.Equal(t, "ESP_ORDER_42", out["order_id"])
	assert.Equal(t, "mock-rsa-signature", out["signature"])
	assert.NotEmpty(t, out["reconcile_id"])
	assert.NotEmpty(t, out["reconcile_datetime"])
	assert.EqualValues(t, 7777, out["amount"])
	assert.Equal(t, "IDR", out["currency"])
	assert.Equal(t, map[string]any{"merchant_ref": "m-1"}, out["request_echo"])
}

func TestBuildWebhookPayload_failureCarriesErrorCode(t *testing.T) {
	a := espay.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeBankDeclineSoft, HTTPStatus: 402},
		"ESP_FAIL",
		1000,
		"IDR",
		nil,
	)
	assert.Equal(t, "ESP02", out["error_code"])
	assert.Equal(t, "Insufficient funds", out["error_message"])
	assert.Equal(t, "ESP_FAIL", out["order_id"])
	_, echoed := out["request_echo"]
	assert.False(t, echoed, "no echo when requestBody is nil")
}

func TestBuildWebhookPayload_networkErrorMapsToESP98(t *testing.T) {
	a := espay.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeNetworkError, HTTPStatus: 503},
		"ESP_NET",
		1000,
		"IDR",
		nil,
	)
	assert.Equal(t, "ESP98", out["error_code"])
}
