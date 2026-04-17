package razorpay_test

import (
	"encoding/json"
	"testing"

	"github.com/prashantluhar/testpay/internal/adapters/razorpay"
	"github.com/prashantluhar/testpay/internal/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdapter_Name(t *testing.T) {
	assert.Equal(t, "razorpay", razorpay.New().Name())
}

func TestBuildResponse_successShape(t *testing.T) {
	a := razorpay.New()
	body := []byte(`{"amount":7777,"currency":"INR","notes":{"order_id":"ord_42"}}`)
	status, raw, headers := a.BuildResponse(&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}, body)

	assert.Equal(t, 200, status)
	assert.Equal(t, "application/json", headers["Content-Type"])

	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.Equal(t, "payment", resp["entity"])
	assert.Equal(t, "captured", resp["status"])
	assert.EqualValues(t, 7777, resp["amount"])
	assert.Equal(t, "INR", resp["currency"])
	assert.Equal(t, "card", resp["method"])
	require.NotEmpty(t, resp["id"])
	assert.Contains(t, resp["id"], "pay_")

	// notes must be echoed onto the payment entity — Razorpay convention.
	notes, ok := resp["notes"].(map[string]any)
	require.True(t, ok, "notes must be an object")
	assert.Equal(t, "ord_42", notes["order_id"])

	// Success path has no error_* fields populated.
	_, hasErrCode := resp["error_code"]
	assert.False(t, hasErrCode, "success response must not carry error_code")
}

func TestBuildResponse_hardFailureShape(t *testing.T) {
	a := razorpay.New()
	status, raw, _ := a.BuildResponse(&engine.Result{
		Mode:       engine.ModeBankInvalidCVV,
		HTTPStatus: 402,
		ErrorCode:  string(engine.ModeBankInvalidCVV),
	}, nil)
	assert.Equal(t, 402, status)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))

	// 4xx responses use Razorpay's { "error": {...} } envelope, not a
	// payment entity.
	_, hasEntity := resp["entity"]
	assert.False(t, hasEntity, "4xx error response must not include a payment entity")

	errObj, ok := resp["error"].(map[string]any)
	require.True(t, ok, "4xx body must wrap fields under `error`")
	assert.Equal(t, "GATEWAY_ERROR", errObj["code"], "bank-side failures map to GATEWAY_ERROR")
	assert.Equal(t, "bank", errObj["source"])
	assert.Equal(t, "payment_authentication", errObj["step"])
	assert.Equal(t, "bank_invalid_cvv", errObj["reason"])
	assert.NotEmpty(t, errObj["description"])
}

func TestBuildResponse_pendingMapsToAuthorized(t *testing.T) {
	a := razorpay.New()
	_, raw, _ := a.BuildResponse(&engine.Result{
		Mode:       engine.ModePendingThenSuccess,
		HTTPStatus: 200,
		IsPending:  true,
	}, nil)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.Equal(t, "authorized", resp["status"],
		"pending modes map to Razorpay 'authorized' — held at issuer, awaiting capture")
}

func TestBuildResponse_inBandAmountMismatchFailsAt200(t *testing.T) {
	// ModeAmountMismatch is HTTPStatus 200 but semantically a failed payment —
	// Razorpay surfaces this as status="failed" on the payment entity with
	// error_* fields populated (not the 4xx error envelope).
	a := razorpay.New()
	status, raw, _ := a.BuildResponse(&engine.Result{
		Mode:       engine.ModeAmountMismatch,
		HTTPStatus: 200,
	}, nil)
	assert.Equal(t, 200, status)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.Equal(t, "payment", resp["entity"])
	assert.Equal(t, "failed", resp["status"])
	assert.Equal(t, "BAD_REQUEST_ERROR", resp["error_code"])
	assert.Equal(t, "business", resp["error_source"])
	assert.Equal(t, "payment_capture", resp["error_step"])
	assert.Equal(t, "amount_mismatch", resp["error_reason"])
}

func TestBuildResponse_defaultAmountINRWhenBodyEmpty(t *testing.T) {
	a := razorpay.New()
	_, raw, _ := a.BuildResponse(&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}, nil)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.EqualValues(t, 5000, resp["amount"])
	assert.Equal(t, "INR", resp["currency"])
}

func TestBuildResponse_invalidJSONBodyFallsBackToDefaults(t *testing.T) {
	a := razorpay.New()
	_, raw, _ := a.BuildResponse(&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}, []byte("not json"))
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.EqualValues(t, 5000, resp["amount"])
	assert.Equal(t, "INR", resp["currency"])
	// notes should be present but empty (Razorpay always emits the key).
	notes, ok := resp["notes"].(map[string]any)
	require.True(t, ok)
	assert.Empty(t, notes)
}

func TestBuildResponse_statusByMode(t *testing.T) {
	// Table-driven: confirm the engine-mode → payment.status mapping matches
	// Razorpay's lifecycle vocabulary.
	cases := []struct {
		name   string
		mode   engine.FailureMode
		status int
		want   string
	}{
		{"success", engine.ModeSuccess, 200, "captured"},
		{"double_charge_still_captured", engine.ModeDoubleCharge, 200, "captured"},
		{"redirect_success", engine.ModeRedirectSuccess, 200, "captured"},
		{"partial_success", engine.ModePartialSuccess, 200, "captured"},
		{"webhook_delayed_still_captured_sync", engine.ModeWebhookDelayed, 200, "captured"},
		{"pending_then_success", engine.ModePendingThenSuccess, 200, "authorized"},
		{"success_then_reversed", engine.ModeSuccessThenReversed, 200, "authorized"},
		{"amount_mismatch", engine.ModeAmountMismatch, 200, "failed"},
	}
	a := razorpay.New()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, raw, _ := a.BuildResponse(&engine.Result{Mode: tc.mode, HTTPStatus: tc.status}, nil)
			var resp map[string]any
			require.NoError(t, json.Unmarshal(raw, &resp))
			assert.Equal(t, tc.want, resp["status"])
		})
	}
}

func TestBuildWebhookPayload_successShape(t *testing.T) {
	a := razorpay.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200},
		"pay_ABC123",
		10000,
		"INR",
		map[string]any{"notes": map[string]any{"order_id": "ord_7"}},
	)

	assert.Equal(t, "event", out["entity"])
	assert.Equal(t, "payment.captured", out["event"])

	contains, ok := out["contains"].([]any)
	require.True(t, ok)
	require.Len(t, contains, 1)
	assert.Equal(t, "payment", contains[0])

	require.NotNil(t, out["created_at"], "envelope must carry created_at unix timestamp")

	entity := out["payload"].(map[string]any)["payment"].(map[string]any)["entity"].(map[string]any)
	assert.Equal(t, "pay_ABC123", entity["id"])
	assert.Equal(t, "payment", entity["entity"])
	assert.Equal(t, "captured", entity["status"])
	assert.EqualValues(t, 10000, entity["amount"])
	assert.Equal(t, "INR", entity["currency"])
	assert.Equal(t, "card", entity["method"])

	notes := entity["notes"].(map[string]any)
	assert.Equal(t, "ord_7", notes["order_id"])
}

func TestBuildWebhookPayload_failureSetsStatusAndErrorFields(t *testing.T) {
	a := razorpay.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeBankDeclineSoft, HTTPStatus: 402},
		"pay_FAIL",
		5000,
		"INR",
		nil,
	)
	assert.Equal(t, "payment.failed", out["event"])

	entity := out["payload"].(map[string]any)["payment"].(map[string]any)["entity"].(map[string]any)
	assert.Equal(t, "failed", entity["status"])
	assert.Equal(t, "GATEWAY_ERROR", entity["error_code"])
	assert.Equal(t, "bank_decline_soft", entity["error_reason"])
	assert.Equal(t, "bank", entity["error_source"])
	assert.Equal(t, "payment_authorization", entity["error_step"])
	assert.NotEmpty(t, entity["error_description"])
}

func TestBuildWebhookPayload_authorizedEvent(t *testing.T) {
	a := razorpay.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModePendingThenSuccess, HTTPStatus: 200, IsPending: true},
		"pay_PEND",
		5000,
		"INR",
		nil,
	)
	assert.Equal(t, "payment.authorized", out["event"])
	entity := out["payload"].(map[string]any)["payment"].(map[string]any)["entity"].(map[string]any)
	assert.Equal(t, "authorized", entity["status"])
}

func TestBuildWebhookPayload_notesEchoedThrough(t *testing.T) {
	a := razorpay.New()
	reqBody := map[string]any{
		"notes": map[string]any{
			"order_id":    "rzp_merchant_42",
			"customer_id": "cust_99",
		},
	}
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200},
		"pay_X", 5000, "INR", reqBody,
	)
	entity := out["payload"].(map[string]any)["payment"].(map[string]any)["entity"].(map[string]any)
	notes := entity["notes"].(map[string]any)
	assert.Equal(t, "rzp_merchant_42", notes["order_id"])
	assert.Equal(t, "cust_99", notes["customer_id"])

	// request_echo mirrors Adyen's convention — merchants can see everything
	// they sent without re-walking their own request state.
	assert.Equal(t, reqBody, out["request_echo"])
}

func TestBuildWebhookPayload_nilRequestBodyDoesNotPanic(t *testing.T) {
	// Defensive: webhooks can fire on retries where the original request
	// body is no longer available. Adapter must emit valid shape with empty
	// notes rather than nil-panic or omitting the key.
	a := razorpay.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200},
		"pay_NONIL", 5000, "INR", nil,
	)
	entity := out["payload"].(map[string]any)["payment"].(map[string]any)["entity"].(map[string]any)
	notes, ok := entity["notes"].(map[string]any)
	require.True(t, ok, "notes key must exist even when request body is nil")
	assert.Empty(t, notes)
	_, hasEcho := out["request_echo"]
	assert.False(t, hasEcho, "request_echo omitted when request body is nil")
}
