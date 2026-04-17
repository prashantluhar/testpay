package paynamics_test

import (
	"encoding/json"
	"testing"

	"github.com/prashantluhar/testpay/internal/adapters/paynamics"
	"github.com/prashantluhar/testpay/internal/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdapter_Name(t *testing.T) {
	assert.Equal(t, "paynamics", paynamics.New().Name())
}

func TestBuildResponse_successShape(t *testing.T) {
	a := paynamics.New()
	body := []byte(`{"amount":9900,"currency":"PHP"}`)
	status, raw, headers := a.BuildResponse(&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}, body)

	assert.Equal(t, 200, status)
	assert.Equal(t, "application/json", headers["Content-Type"])

	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.Equal(t, "GR001", resp["response_code"])
	assert.Equal(t, "SUCCESS", resp["response_advise"])
	assert.Equal(t, "Transaction successful", resp["response_message"])
	assert.Equal(t, "mock-md5-signature", resp["signature"])
	assert.Equal(t, "MOCK_MERCHANT", resp["merchant_id"])
	assert.NotEmpty(t, resp["response_id"])
	assert.NotEmpty(t, resp["request_id"])
	assert.Contains(t, resp["redirect_url"], "mock.paynamics.net/pay/")
	assert.Equal(t, "PHP", resp["currency"])
	assert.Equal(t, "99.00", resp["total_amount"], "Paynamics encodes amount as decimal-string")
}

func TestBuildResponse_hardFailureShape(t *testing.T) {
	a := paynamics.New()
	status, raw, _ := a.BuildResponse(&engine.Result{
		Mode:       engine.ModeBankInvalidCVV,
		HTTPStatus: 402,
	}, nil)
	assert.Equal(t, 402, status)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.Equal(t, "GR052", resp["response_code"], "CVV fail maps to GR052")
	assert.Equal(t, "DECLINED", resp["response_advise"])
	assert.Equal(t, "Invalid CVV", resp["response_message"])
	assert.Equal(t, "mock-md5-signature", resp["signature"])
	// Failure response must not carry a redirect_url (the whole point of
	// omitempty in the DTO — merchant should not send the user anywhere).
	_, hasRedirect := resp["redirect_url"]
	assert.False(t, hasRedirect, "error envelope should omit redirect_url")
}

func TestBuildResponse_pendingMapsToGR002(t *testing.T) {
	a := paynamics.New()
	_, raw, _ := a.BuildResponse(&engine.Result{
		Mode:       engine.ModePendingThenSuccess,
		HTTPStatus: 200,
		IsPending:  true,
	}, nil)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.Equal(t, "GR002", resp["response_code"])
	assert.Equal(t, "PENDING", resp["response_advise"])
}

func TestBuildResponse_defaultAmountPHPWhenBodyEmpty(t *testing.T) {
	a := paynamics.New()
	_, raw, _ := a.BuildResponse(&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}, nil)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.Equal(t, "PHP", resp["currency"])
	assert.Equal(t, "50.00", resp["total_amount"])
}

func TestBuildResponse_invalidJSONBodyFallsBackToDefaults(t *testing.T) {
	a := paynamics.New()
	_, raw, _ := a.BuildResponse(&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}, []byte("{not json"))
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.Equal(t, "PHP", resp["currency"])
	assert.Equal(t, "50.00", resp["total_amount"])
}

func TestBuildWebhookPayload_successShape(t *testing.T) {
	a := paynamics.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200},
		"PNMC_42",
		12345,
		"PHP",
		map[string]any{"request_id": "REQ_xyz"},
	)

	assert.Equal(t, "GR001", out["response_code"])
	assert.Equal(t, "SUCCESS", out["response_advise"])
	assert.Equal(t, "PNMC_42", out["response_id"])
	assert.Equal(t, "MOCK_MERCHANT", out["merchant_id"])
	assert.Equal(t, "PHP", out["currency"])
	assert.Equal(t, "123.45", out["total_amount"])
	assert.Equal(t, "mock-md5-signature", out["signature"])
	assert.NotEmpty(t, out["timestamp"])

	echo, ok := out["request_echo"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "REQ_xyz", echo["request_id"])
}

func TestBuildWebhookPayload_failureMapsCodes(t *testing.T) {
	a := paynamics.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeBankDeclineSoft, HTTPStatus: 402},
		"PNMC_FAIL",
		5000,
		"PHP",
		nil,
	)
	assert.Equal(t, "GR051", out["response_code"])
	assert.Equal(t, "DECLINED", out["response_advise"])
	assert.Equal(t, "Insufficient funds", out["response_message"])
}

func TestBuildWebhookPayload_pendingMapsToGR002(t *testing.T) {
	a := paynamics.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModePendingThenSuccess, HTTPStatus: 200, IsPending: true},
		"PNMC_PEND",
		5000,
		"PHP",
		nil,
	)
	assert.Equal(t, "GR002", out["response_code"])
	assert.Equal(t, "PENDING", out["response_advise"])
}

func TestBuildResponse_nestedTransactionAmountStringParses(t *testing.T) {
	a := paynamics.New()
	body := []byte(`{"transaction":{"amount":"250.75","currency":"USD"}}`)
	_, raw, _ := a.BuildResponse(&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}, body)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	// 250.75 major units → 25075 minor units → "250.75" total_amount
	assert.Equal(t, "USD", resp["currency"])
	assert.Equal(t, "250.75", resp["total_amount"])
}
