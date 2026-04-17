package instamojo_test

import (
	"encoding/json"
	"testing"

	"github.com/prashantluhar/testpay/internal/adapters/instamojo"
	"github.com/prashantluhar/testpay/internal/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdapter_Name(t *testing.T) {
	assert.Equal(t, "instamojo", instamojo.New().Name())
}

func TestBuildResponse_successShape(t *testing.T) {
	a := instamojo.New()
	body := []byte(`{"amount":2500,"currency":"INR","purpose":"book"}`)
	status, raw, headers := a.BuildResponse(&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}, body)

	assert.Equal(t, 200, status)
	assert.Equal(t, "application/json", headers["Content-Type"])

	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.Equal(t, true, resp["success"])

	pr, ok := resp["payment_request"].(map[string]any)
	require.True(t, ok, "payment_request must be an object")
	assert.NotEmpty(t, pr["id"])
	assert.Equal(t, "2500", pr["amount"], "Instamojo encodes amount as string")
	assert.Equal(t, "Sent", pr["status"])
	assert.Contains(t, pr["longurl"], "instamojo.com")
}

func TestBuildResponse_hardFailureShape(t *testing.T) {
	a := instamojo.New()
	status, raw, _ := a.BuildResponse(&engine.Result{
		Mode:       engine.ModeBankInvalidCVV,
		HTTPStatus: 400,
	}, nil)
	assert.Equal(t, 400, status)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.Equal(t, false, resp["success"])
	assert.Equal(t, "Invalid CVV", resp["message"])

	errs, ok := resp["errors"].(map[string]any)
	require.True(t, ok, "errors must be a map")
	codeList, ok := errs["code"].([]any)
	require.True(t, ok)
	assert.Contains(t, codeList, string(engine.ModeBankInvalidCVV))
}

func TestBuildResponse_pendingMapsToPendingStatus(t *testing.T) {
	a := instamojo.New()
	_, raw, _ := a.BuildResponse(&engine.Result{
		Mode:       engine.ModePendingThenSuccess,
		HTTPStatus: 200,
		IsPending:  true,
	}, nil)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	pr := resp["payment_request"].(map[string]any)
	assert.Equal(t, "Pending", pr["status"])
}

func TestBuildResponse_defaultAmountINRWhenBodyEmpty(t *testing.T) {
	a := instamojo.New()
	_, raw, _ := a.BuildResponse(&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}, nil)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	pr := resp["payment_request"].(map[string]any)
	assert.Equal(t, "5000", pr["amount"])
}

func TestBuildResponse_invalidJSONBodyFallsBackToDefaults(t *testing.T) {
	a := instamojo.New()
	_, raw, _ := a.BuildResponse(&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}, []byte("not json"))
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	pr := resp["payment_request"].(map[string]any)
	assert.Equal(t, "5000", pr["amount"])
}

func TestBuildWebhookPayload_successShape(t *testing.T) {
	a := instamojo.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200},
		"MOJOPAY_42",
		2500,
		"INR",
		map[string]any{"purpose": "T-shirt"},
	)
	assert.Equal(t, "MOJOPAY_42", out["payment_id"])
	assert.Equal(t, "Credit", out["status"])
	assert.Equal(t, "2500", out["amount"], "webhook amount is string-encoded too")
	assert.Equal(t, "INR", out["currency"])
	assert.Equal(t, "T-shirt", out["purpose"], "purpose echoed as top-level field")
	assert.NotEmpty(t, out["payment_request_id"])
	assert.NotEmpty(t, out["created_at"])

	echo, ok := out["request_echo"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "T-shirt", echo["purpose"])
}

func TestBuildWebhookPayload_failureSetsFailedStatusAndReason(t *testing.T) {
	a := instamojo.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeBankDeclineSoft, HTTPStatus: 402},
		"MOJOPAY_FAIL",
		2500,
		"INR",
		nil,
	)
	assert.Equal(t, "Failed", out["status"])
	assert.Equal(t, "Insufficient funds", out["failure_message"])
	assert.Equal(t, string(engine.ModeBankDeclineSoft), out["failure_reason"])
}
