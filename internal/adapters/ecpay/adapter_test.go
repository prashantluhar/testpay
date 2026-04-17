package ecpay_test

import (
	"encoding/json"
	"testing"

	"github.com/prashantluhar/testpay/internal/adapters/ecpay"
	"github.com/prashantluhar/testpay/internal/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdapter_Name(t *testing.T) {
	assert.Equal(t, "ecpay", ecpay.New().Name())
}

func TestBuildResponse_successShape(t *testing.T) {
	a := ecpay.New()
	body := []byte(`{"MerchantTradeNo":"ORDER123","TradeAmt":500,"PaymentMethod":"ECPAY_CREDIT_CARD"}`)
	status, raw, headers := a.BuildResponse(&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}, body)

	assert.Equal(t, 200, status)
	assert.Equal(t, "application/json", headers["Content-Type"])

	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	// ECPay's primary success signal is RtnCode == 1.
	assert.EqualValues(t, 1, resp["RtnCode"])
	assert.Equal(t, "交易成功", resp["RtnMsg"])
	assert.Equal(t, "3002607", resp["MerchantID"], "ECPay stage MerchantID")
	assert.Equal(t, "ORDER123", resp["MerchantTradeNo"], "must echo merchant's correlation ID")
	assert.NotEmpty(t, resp["TradeNo"])
	assert.EqualValues(t, 500, resp["TradeAmt"])
	assert.Equal(t, "Credit_CreditCard", resp["PaymentType"])
	assert.NotEmpty(t, resp["CheckMacValue"])
	assert.NotEmpty(t, resp["TradeDate"])
}

func TestBuildResponse_failureShape(t *testing.T) {
	a := ecpay.New()
	status, raw, _ := a.BuildResponse(&engine.Result{
		Mode:       engine.ModeBankInvalidCVV,
		HTTPStatus: 402,
		ErrorCode:  string(engine.ModeBankInvalidCVV),
	}, nil)
	assert.Equal(t, 402, status)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	// Failure path: RtnCode must NOT be 1, and ErrorType signals the
	// validation-layer failure envelope.
	assert.NotEqualValues(t, 1, resp["RtnCode"])
	assert.Equal(t, "Invalid security code", resp["RtnMsg"])
	assert.Equal(t, "validation", resp["ErrorType"])
	assert.NotEmpty(t, resp["MerchantTradeNo"])
	assert.NotEmpty(t, resp["CheckMacValue"])
}

func TestBuildResponse_paymentMethodMapping(t *testing.T) {
	// The production DTO's PaymentMethodToECPay map translates our internal
	// ECPAY_* names to ECPay's wire `PaymentType` strings. Verify all four.
	cases := []struct {
		internal string
		wire     string
	}{
		{"ECPAY_CREDIT_CARD", "Credit_CreditCard"},
		{"ECPAY_ATM_CARD", "ATM_TAISHIN"},
		{"ECPAY_CVS", "CVS_CVS"},
		{"ECPAY_BARCODE", "BARCODE_BARCODE"},
	}
	a := ecpay.New()
	for _, tc := range cases {
		t.Run(tc.internal, func(t *testing.T) {
			body := []byte(`{"PaymentMethod":"` + tc.internal + `"}`)
			_, raw, _ := a.BuildResponse(&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}, body)
			var resp map[string]any
			require.NoError(t, json.Unmarshal(raw, &resp))
			assert.Equal(t, tc.wire, resp["PaymentType"])
		})
	}
}

func TestBuildResponse_defaultsWhenBodyEmpty(t *testing.T) {
	a := ecpay.New()
	_, raw, _ := a.BuildResponse(&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}, nil)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	// Default test-tier amount is 100 TWD; PaymentType defaults to credit.
	assert.EqualValues(t, 100, resp["TradeAmt"])
	assert.Equal(t, "Credit_CreditCard", resp["PaymentType"])
	// Fallback MerchantTradeNo must be present (generated) when merchant
	// didn't send one.
	mtn, ok := resp["MerchantTradeNo"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, mtn)
}

func TestBuildResponse_invalidJSONFallsBack(t *testing.T) {
	a := ecpay.New()
	_, raw, _ := a.BuildResponse(&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}, []byte("not json"))
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.EqualValues(t, 100, resp["TradeAmt"])
	assert.EqualValues(t, 1, resp["RtnCode"])
}

func TestBuildWebhookPayload_successShape(t *testing.T) {
	a := ecpay.New()
	reqBody := map[string]any{
		"MerchantTradeNo": "ORDER-XYZ",
		"PaymentMethod":   "ECPAY_CREDIT_CARD",
		"CustomField1":    "user-42",
		"CustomField2":    "cart-99",
	}
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200},
		"EC-TRADE-001",
		2500,
		"TWD",
		reqBody,
	)
	assert.EqualValues(t, 1, out["RtnCode"], "RtnCode=1 is ECPay's only success signal")
	assert.Equal(t, "交易成功", out["RtnMsg"])
	assert.Equal(t, "EC-TRADE-001", out["TradeNo"])
	assert.Equal(t, "ORDER-XYZ", out["MerchantTradeNo"])
	assert.EqualValues(t, 2500, out["TradeAmt"])
	assert.Equal(t, "Credit_CreditCard", out["PaymentType"])
	assert.Equal(t, "1", out["SimulatePaid"], "SimulatePaid=1 marks this as a test webhook — merchants must not fulfil")
	assert.Equal(t, "TESTSTORE", out["StoreID"])
	assert.Equal(t, "user-42", out["CustomField1"])
	assert.Equal(t, "cart-99", out["CustomField2"])
	assert.NotEmpty(t, out["CheckMacValue"])
	// ECPay doesn't carry a currency field — ensure we haven't accidentally
	// leaked one into the payload.
	_, hasCurrency := out["Currency"]
	assert.False(t, hasCurrency, "ECPay has no currency field — TWD is implicit")
}

func TestBuildWebhookPayload_failureShape(t *testing.T) {
	a := ecpay.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeBankDeclineHard, HTTPStatus: 402, ErrorCode: "bank_decline_hard"},
		"EC-FAIL",
		500,
		"TWD",
		nil,
	)
	// Failure: RtnCode must NOT be 1.
	assert.NotEqualValues(t, 1, out["RtnCode"])
	assert.Equal(t, "Issuer declined transaction", out["RtnMsg"])
	assert.Equal(t, "EC-FAIL", out["TradeNo"])
}

func TestBuildWebhookPayload_pendingMapsToDistinctCode(t *testing.T) {
	// ATM/CVS pending flows use RtnCode 10100073 ("get payment code
	// succeeded") — distinct from the 1=success and the generic decline
	// code. Merchants key off this to render "awaiting bank transfer" UI.
	a := ecpay.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModePendingThenSuccess, HTTPStatus: 200, IsPending: true},
		"EC-PEND",
		100,
		"TWD",
		nil,
	)
	assert.EqualValues(t, 10100073, out["RtnCode"])
	assert.Equal(t, "Get CVS Code Succeeded", out["RtnMsg"])
}

func TestBuildWebhookPayload_httpFailureOverridesSuccessMode(t *testing.T) {
	// Guard: even if the mode's classification is ambiguous, an HTTPStatus
	// >= 400 must always force a failure RtnCode. Use ModePGServerError
	// which maps to HTTP 500.
	a := ecpay.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModePGServerError, HTTPStatus: 500},
		"EC-SRV", 100, "TWD", nil,
	)
	assert.NotEqualValues(t, 1, out["RtnCode"])
}

func TestBuildWebhookPayload_requestEchoIncluded(t *testing.T) {
	a := ecpay.New()
	reqBody := map[string]any{"foo": "bar", "MerchantTradeNo": "X"}
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200},
		"EC-Z", 100, "TWD", reqBody,
	)
	assert.Equal(t, reqBody, out["request_echo"])
}

func TestBuildWebhookPayload_fallbackMerchantTradeNoWhenAbsent(t *testing.T) {
	a := ecpay.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200},
		"EC-Q", 100, "TWD", nil,
	)
	mtn, ok := out["MerchantTradeNo"].(string)
	require.True(t, ok)
	assert.Contains(t, mtn, "MTN_", "absent MerchantTradeNo must fall back to a generated MTN_-prefixed id")
}

func TestBuildWebhookPayload_dateFormatIsECPayConvention(t *testing.T) {
	// ECPay uses `YYYY/MM/DD HH:MM:SS` (slashes, 24h, TW local) for
	// PaymentDate and TradeDate — not ISO-8601. Merchants parse this
	// string format directly.
	a := ecpay.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200},
		"EC-D", 100, "TWD", nil,
	)
	pd, ok := out["PaymentDate"].(string)
	require.True(t, ok)
	assert.Regexp(t, `^\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}$`, pd, "PaymentDate must use ECPay's YYYY/MM/DD HH:MM:SS format")
}
