package mastercard_test

import (
	"encoding/json"
	"testing"

	"github.com/prashantluhar/testpay/internal/adapters/mastercard"
	"github.com/prashantluhar/testpay/internal/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdapter_Name(t *testing.T) {
	assert.Equal(t, "mastercard", mastercard.New().Name())
}

func TestBuildResponse_successShape(t *testing.T) {
	a := mastercard.New()
	body := []byte(`{"order":{"amount":7777,"currency":"EUR"}}`)
	status, raw, headers := a.BuildResponse(&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}, body)

	assert.Equal(t, 200, status)
	assert.Equal(t, "application/json", headers["Content-Type"])

	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.Equal(t, "SUCCESS", resp["result"], "MPGS result=SUCCESS on approval")
	assert.Equal(t, "TESTMERCHANT", resp["merchant"])
	assert.NotEmpty(t, resp["timeOfRecord"])

	order, ok := resp["order"].(map[string]any)
	require.True(t, ok, "order must be an object")
	assert.NotEmpty(t, order["id"])
	assert.EqualValues(t, 7777, order["amount"])
	assert.Equal(t, "EUR", order["currency"])
	assert.Equal(t, "CAPTURED", order["status"])
	assert.EqualValues(t, 7777, order["totalAuthorizedAmount"])
	assert.EqualValues(t, 7777, order["totalCapturedAmount"])

	responseBlock, ok := resp["response"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "APPROVED", responseBlock["gatewayCode"])
	assert.Equal(t, "PROCEED", responseBlock["gatewayRecommendation"])
	assert.Equal(t, "00", responseBlock["acquirerCode"])

	csc, ok := responseBlock["cardSecurityCode"].(map[string]any)
	require.True(t, ok, "cardSecurityCode block must be present on card auth")
	assert.Equal(t, "MATCH", csc["gatewayCode"])

	txn, ok := resp["transaction"].(map[string]any)
	require.True(t, ok)
	assert.NotEmpty(t, txn["id"])
	assert.Equal(t, "PAYMENT", txn["type"])
	assert.EqualValues(t, 7777, txn["amount"])
	assert.NotEmpty(t, txn["authorizationCode"])

	// Success path has no error block.
	_, hasErr := resp["error"]
	assert.False(t, hasErr, "success response must not include an error block")
}

func TestBuildResponse_failureShapeMPGSCodes(t *testing.T) {
	tests := []struct {
		name             string
		mode             engine.FailureMode
		httpStatus       int
		wantGatewayCode  string
		wantResult       string
		wantAcquirerCode string
	}{
		{"hard decline -> DECLINED", engine.ModeBankDeclineHard, 402, "DECLINED", "FAILURE", "05"},
		{"invalid cvv -> INVALID_CSC", engine.ModeBankInvalidCVV, 402, "INVALID_CSC", "FAILURE", "N7"},
		{"bank timeout -> TIMED_OUT", engine.ModeBankTimeout, 504, "TIMED_OUT", "ERROR", "91"},
		{"rate limited -> BLOCKED", engine.ModePGRateLimited, 429, "BLOCKED", "FAILURE", "62"},
		{"redirect timeout -> EXPIRED_CARD", engine.ModeRedirectTimeout, 408, "EXPIRED_CARD", "FAILURE", "54"},
		{"pg server error -> SYSTEM_ERROR", engine.ModePGServerError, 500, "SYSTEM_ERROR", "ERROR", "96"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := mastercard.New()
			status, raw, _ := a.BuildResponse(&engine.Result{
				Mode:       tc.mode,
				HTTPStatus: tc.httpStatus,
				ErrorCode:  string(tc.mode),
			}, nil)
			assert.Equal(t, tc.httpStatus, status)

			var resp map[string]any
			require.NoError(t, json.Unmarshal(raw, &resp))
			assert.Equal(t, tc.wantResult, resp["result"])

			rb := resp["response"].(map[string]any)
			assert.Equal(t, tc.wantGatewayCode, rb["gatewayCode"])
			assert.Equal(t, tc.wantAcquirerCode, rb["acquirerCode"])
			if tc.wantResult == "FAILURE" {
				assert.Equal(t, "DO_NOT_PROCEED", rb["gatewayRecommendation"])
			} else {
				assert.Equal(t, "RESUBMIT", rb["gatewayRecommendation"])
			}

			// Failure/error path always includes the error block.
			errBlock, ok := resp["error"].(map[string]any)
			require.True(t, ok, "failure/error must include MPGS error block")
			assert.NotEmpty(t, errBlock["explanation"])
			assert.NotEmpty(t, errBlock["cause"])

			order := resp["order"].(map[string]any)
			assert.Equal(t, "FAILED", order["status"])
		})
	}
}

func TestBuildResponse_invalidCVVPopulatesValidationFields(t *testing.T) {
	a := mastercard.New()
	_, raw, _ := a.BuildResponse(&engine.Result{
		Mode:       engine.ModeBankInvalidCVV,
		HTTPStatus: 402,
	}, nil)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	errBlock := resp["error"].(map[string]any)
	assert.Equal(t, "INVALID", errBlock["validationType"])
	assert.Equal(t, "sourceOfFunds.provided.card.securityCode", errBlock["field"])
	assert.Equal(t, "Invalid Card Security Code", errBlock["explanation"])
}

func TestBuildResponse_pendingShape(t *testing.T) {
	a := mastercard.New()
	status, raw, _ := a.BuildResponse(&engine.Result{
		Mode:       engine.ModePendingThenSuccess,
		HTTPStatus: 200,
		IsPending:  true,
	}, nil)
	assert.Equal(t, 200, status)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.Equal(t, "PENDING", resp["result"])
	rb := resp["response"].(map[string]any)
	assert.Equal(t, "PENDING", rb["gatewayCode"])
	assert.Equal(t, "PROCEED", rb["gatewayRecommendation"])
	order := resp["order"].(map[string]any)
	assert.Equal(t, "PENDING", order["status"])
}

func TestBuildResponse_defaultAmountUSDWhenBodyEmpty(t *testing.T) {
	a := mastercard.New()
	_, raw, _ := a.BuildResponse(&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}, nil)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	order := resp["order"].(map[string]any)
	assert.EqualValues(t, 5000, order["amount"])
	assert.Equal(t, "USD", order["currency"])
	txn := resp["transaction"].(map[string]any)
	assert.EqualValues(t, 5000, txn["amount"])
	assert.Equal(t, "USD", txn["currency"])
}

func TestBuildResponse_invalidJSONBodyFallsBackToDefaults(t *testing.T) {
	a := mastercard.New()
	_, raw, _ := a.BuildResponse(&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}, []byte("not json"))
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	order := resp["order"].(map[string]any)
	assert.EqualValues(t, 5000, order["amount"], "invalid body must not crash and must fall back to 5000/USD default")
	assert.Equal(t, "USD", order["currency"])
}

func TestBuildResponse_customAmountEcho(t *testing.T) {
	a := mastercard.New()
	body := []byte(`{"order":{"amount":12345,"currency":"SGD"},"transaction":{"reference":"ref-1"}}`)
	_, raw, _ := a.BuildResponse(&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}, body)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	order := resp["order"].(map[string]any)
	assert.EqualValues(t, 12345, order["amount"])
	assert.Equal(t, "SGD", order["currency"])
	txn := resp["transaction"].(map[string]any)
	assert.EqualValues(t, 12345, txn["amount"])
	assert.Equal(t, "SGD", txn["currency"])
}

func TestBuildWebhookPayload_successShape(t *testing.T) {
	a := mastercard.New()
	req := map[string]any{"order": map[string]any{"reference": "my-order-42"}}
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200},
		"ORD12345",
		10000,
		"USD",
		req,
	)

	assert.Equal(t, "SUCCESS", out["result"])
	assert.Equal(t, "ORDER", out["notificationType"])
	assert.NotEmpty(t, out["notificationId"])
	assert.NotEmpty(t, out["timeOfNotification"])
	assert.Equal(t, "TESTMERCHANT", out["merchant"])

	order := out["order"].(map[string]any)
	assert.Equal(t, "ORD12345", order["id"])
	assert.EqualValues(t, 10000, order["amount"])
	assert.Equal(t, "USD", order["currency"])
	assert.Equal(t, "CAPTURED", order["status"])
	assert.EqualValues(t, 10000, order["totalAuthorizedAmount"])

	rb := out["response"].(map[string]any)
	assert.Equal(t, "APPROVED", rb["gatewayCode"])
	assert.Equal(t, "00", rb["acquirerCode"])

	txn := out["transaction"].(map[string]any)
	assert.Equal(t, "PAYMENT", txn["type"])
	assert.EqualValues(t, 10000, txn["amount"])

	// request_echo must be carried through for consumer correlation.
	assert.Equal(t, req, out["request_echo"])
}

func TestBuildWebhookPayload_failureShape(t *testing.T) {
	a := mastercard.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeBankDeclineSoft, HTTPStatus: 402},
		"ORD_FAIL",
		5000,
		"USD",
		nil,
	)
	assert.Equal(t, "FAILURE", out["result"])
	assert.Equal(t, "TRANSACTION", out["notificationType"], "failed webhooks use notificationType=TRANSACTION")

	rb := out["response"].(map[string]any)
	assert.Equal(t, "DECLINED_DO_NOT_CONTACT", rb["gatewayCode"])
	assert.Equal(t, "DO_NOT_PROCEED", rb["gatewayRecommendation"])
	assert.Equal(t, "Insufficient Funds", rb["acquirerMessage"])

	order := out["order"].(map[string]any)
	assert.Equal(t, "FAILED", order["status"])
	// No totals on failure.
	assert.EqualValues(t, 0, order["totalAuthorizedAmount"])
}

func TestBuildWebhookPayload_reversalSetsRefundTransactionType(t *testing.T) {
	a := mastercard.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeSuccessThenReversed, HTTPStatus: 200},
		"ORD_ORIG",
		5000,
		"USD",
		nil,
	)
	assert.Equal(t, "TRANSACTION", out["notificationType"])
	txn := out["transaction"].(map[string]any)
	assert.Equal(t, "REFUND", txn["type"], "reversal → transaction.type=REFUND")
}

func TestBuildWebhookPayload_pendingShape(t *testing.T) {
	a := mastercard.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModePendingThenSuccess, HTTPStatus: 200, IsPending: true},
		"ORD_PEND",
		5000,
		"USD",
		nil,
	)
	assert.Equal(t, "PENDING", out["result"])
	rb := out["response"].(map[string]any)
	assert.Equal(t, "PENDING", rb["gatewayCode"])
	order := out["order"].(map[string]any)
	assert.Equal(t, "PENDING", order["status"])
}

func TestBuildWebhookPayload_requestEchoIncluded(t *testing.T) {
	a := mastercard.New()
	reqBody := map[string]any{"order": map[string]any{"reference": "corr-xyz"}}
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200},
		"ORD_X", 5000, "USD", reqBody,
	)
	assert.Equal(t, reqBody, out["request_echo"])
}
