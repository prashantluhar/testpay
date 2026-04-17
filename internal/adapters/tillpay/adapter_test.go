package tillpay_test

import (
	"encoding/json"
	"testing"

	"github.com/prashantluhar/testpay/internal/adapters/tillpay"
	"github.com/prashantluhar/testpay/internal/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdapter_Name(t *testing.T) {
	assert.Equal(t, "tillpay", tillpay.New().Name())
}

func TestBuildResponse_successShape(t *testing.T) {
	a := tillpay.New()
	status, raw, headers := a.BuildResponse(&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}, nil)

	assert.Equal(t, 200, status)
	assert.Equal(t, "application/json", headers["Content-Type"])
	assert.Equal(t, "mock-sha512-signature", headers["X-Signature"], "TillPayment must sign every response")

	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.Equal(t, true, resp["success"])
	assert.NotEmpty(t, resp["uuid"])
	assert.NotEmpty(t, resp["purchaseId"])
	assert.Equal(t, "FINISHED", resp["returnType"])
	assert.Equal(t, "CREDITCARD", resp["paymentMethod"])
	// Non-redirect success must not carry a redirectUrl.
	_, hasRedirect := resp["redirectUrl"]
	assert.False(t, hasRedirect)
	// Success responses have no errors array.
	_, hasErrors := resp["errors"]
	assert.False(t, hasErrors)
}

func TestBuildResponse_hardFailureShape(t *testing.T) {
	a := tillpay.New()
	status, raw, _ := a.BuildResponse(&engine.Result{
		Mode:       engine.ModeBankInvalidCVV,
		HTTPStatus: 402,
		ErrorCode:  string(engine.ModeBankInvalidCVV),
	}, nil)
	assert.Equal(t, 402, status)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.Equal(t, false, resp["success"])
	assert.Equal(t, "Invalid CVV", resp["errorMessage"])
	assert.EqualValues(t, 2003, resp["errorCode"], "ModeBankInvalidCVV must map to TillPayment code 2003")
	assert.Equal(t, "ERROR", resp["returnType"])

	errs, ok := resp["errors"].([]any)
	require.True(t, ok, "errors array must be populated on failure")
	require.Len(t, errs, 1)
	first := errs[0].(map[string]any)
	assert.Equal(t, "Invalid CVV", first["errorMessage"])
	assert.Equal(t, "PAYMENT_SYSTEM_ERROR", first["adapterCode"])
}

func TestBuildResponse_redirectSuccessIncludesRedirectURL(t *testing.T) {
	a := tillpay.New()
	_, raw, _ := a.BuildResponse(&engine.Result{Mode: engine.ModeRedirectSuccess, HTTPStatus: 200}, nil)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.Equal(t, "REDIRECT", resp["returnType"])
	url, ok := resp["redirectUrl"].(string)
	require.True(t, ok)
	assert.Contains(t, url, "tillpayments.com/redirect/")
}

func TestBuildResponse_pendingReturnsRedirectFlow(t *testing.T) {
	// Pending modes in TillPayment's production flow come back with a
	// redirect URL and returnType=REDIRECT so the merchant awaits webhook.
	a := tillpay.New()
	_, raw, _ := a.BuildResponse(&engine.Result{
		Mode:       engine.ModePendingThenSuccess,
		HTTPStatus: 200,
		IsPending:  true,
	}, nil)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.Equal(t, "REDIRECT", resp["returnType"])
	assert.NotEmpty(t, resp["redirectUrl"])
}

func TestBuildResponse_defaultErrorCodeForUnmappedMode(t *testing.T) {
	// The reference DTO defines DefaultErrorCode = "5000" for unmapped
	// failures. Use a redirect-abandoned to verify the catch-all path — it
	// maps to HTTP 402 via engine but isn't in the explicit code map.
	a := tillpay.New()
	_, raw, _ := a.BuildResponse(&engine.Result{
		Mode:       engine.ModeRedirectAbandoned,
		HTTPStatus: 402,
	}, nil)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.EqualValues(t, 5000, resp["errorCode"], "unmapped modes must fall back to DefaultErrorCode 5000")
	assert.Equal(t, "Customer abandoned redirect", resp["errorMessage"])
}

func TestBuildWebhookPayload_successShape(t *testing.T) {
	a := tillpay.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200},
		"TP-12345",
		12345, // minor units → "123.45"
		"USD",
		map[string]any{
			"merchantTransactionId": "my-order-42",
			"customer": map[string]any{
				"email":     "buyer@example.com",
				"firstName": "Jane",
			},
		},
	)

	assert.Equal(t, "OK", out["result"])
	assert.Equal(t, true, out["success"])
	assert.Equal(t, "SUCCESS", out["transactionStatus"])
	assert.Equal(t, "TP-12345", out["uuid"])
	assert.Equal(t, "my-order-42", out["merchantTransactionId"], "must echo merchant's correlation ID")
	assert.Equal(t, "PUR_TP-12345", out["purchaseId"])
	assert.Equal(t, "DEBIT", out["transactionType"])
	assert.Equal(t, "CREDITCARD", out["paymentMethod"])
	// TillPayment quirk: amount is a string with 2 decimals, not a number.
	assert.Equal(t, "123.45", out["amount"])
	assert.Equal(t, "USD", out["currency"])

	cust := out["customer"].(map[string]any)
	assert.Equal(t, "buyer@example.com", cust["email"])
	assert.Equal(t, "Jane", cust["firstName"])

	ret := out["returnData"].(map[string]any)
	assert.Equal(t, "411111", ret["firstSixDigits"], "BIN snapshot must be present on card webhooks")
	assert.Equal(t, "1111", ret["lastFourDigits"])

	// Success webhooks must not carry error fields.
	_, hasErrors := out["errors"]
	assert.False(t, hasErrors)
}

func TestBuildWebhookPayload_failureShape(t *testing.T) {
	a := tillpay.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeBankDeclineSoft, HTTPStatus: 402, ErrorCode: "bank_decline_soft"},
		"TP-FAIL",
		5000,
		"USD",
		nil,
	)
	assert.Equal(t, "NOK", out["result"])
	assert.Equal(t, false, out["success"])
	assert.Equal(t, "FAILED", out["transactionStatus"])
	assert.Equal(t, "Insufficient funds", out["message"])
	assert.EqualValues(t, 2002, out["code"])
	assert.Equal(t, "PAYMENT_SYSTEM_ERROR", out["adapterCode"])

	errs, ok := out["errors"].([]any)
	require.True(t, ok)
	require.Len(t, errs, 1)
	first := errs[0].(map[string]any)
	assert.EqualValues(t, 2002, first["errorCode"])
	assert.Equal(t, "Insufficient funds", first["errorMessage"])
}

func TestBuildWebhookPayload_reversalMapsToChargebackReversed(t *testing.T) {
	a := tillpay.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeSuccessThenReversed, HTTPStatus: 200},
		"TP-REV",
		5000,
		"USD",
		nil,
	)
	assert.Equal(t, "CHARGEBACK-REVERSED", out["transactionStatus"])
}

func TestBuildWebhookPayload_requestEchoAndDefaults(t *testing.T) {
	a := tillpay.New()
	reqBody := map[string]any{"hello": "world"}
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200},
		"TP-X", 5000, "USD", reqBody,
	)
	assert.Equal(t, reqBody, out["request_echo"])
	// Absent merchantTransactionId must fall back to a generated MTX_ prefix.
	mtx, ok := out["merchantTransactionId"].(string)
	require.True(t, ok)
	assert.Contains(t, mtx, "MTX_")
	// Default synthetic customer when request provides none.
	cust := out["customer"].(map[string]any)
	assert.Equal(t, "test@example.com", cust["email"])
}

func TestBuildWebhookPayload_amountFormatsSubDollarCorrectly(t *testing.T) {
	// Guard against floating-point surprises — 1 cent must serialise as
	// "0.01" not "0.010000" or "1e-02".
	a := tillpay.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200},
		"TP-CENT", 1, "USD", nil,
	)
	assert.Equal(t, "0.01", out["amount"])
}
