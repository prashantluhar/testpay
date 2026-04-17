package adyen_test

import (
	"encoding/json"
	"testing"

	"github.com/prashantluhar/testpay/internal/adapters/adyen"
	"github.com/prashantluhar/testpay/internal/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdapter_Name(t *testing.T) {
	assert.Equal(t, "adyen", adyen.New().Name())
}

func TestBuildResponse_successShape(t *testing.T) {
	a := adyen.New()
	body := []byte(`{"amount":{"value":7777,"currency":"EUR"}}`)
	status, raw, headers := a.BuildResponse(&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}, body)

	assert.Equal(t, 200, status)
	assert.Equal(t, "application/json", headers["Content-Type"])

	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.Equal(t, "Authorised", resp["resultCode"])
	assert.NotEmpty(t, resp["pspReference"])
	assert.NotEmpty(t, resp["merchantReference"])

	amount, ok := resp["amount"].(map[string]any)
	require.True(t, ok, "amount must be an object")
	assert.EqualValues(t, 7777, amount["value"])
	assert.Equal(t, "EUR", amount["currency"])

	pm, ok := resp["paymentMethod"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "scheme", pm["type"])
	assert.Equal(t, "visa", pm["brand"])

	// Success path has no refusal fields.
	_, hasReason := resp["refusalReason"]
	assert.False(t, hasReason)
}

func TestBuildResponse_hardFailureShape(t *testing.T) {
	a := adyen.New()
	status, raw, _ := a.BuildResponse(&engine.Result{
		Mode:       engine.ModeBankInvalidCVV,
		HTTPStatus: 402,
		ErrorCode:  string(engine.ModeBankInvalidCVV),
	}, nil)
	assert.Equal(t, 402, status)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.EqualValues(t, 402, resp["status"])
	assert.Equal(t, "7", resp["errorCode"], "CVC declined → Adyen refusalReasonCode 7")
	assert.Equal(t, "CVC Declined", resp["message"])
	assert.Equal(t, "validation", resp["errorType"])
	assert.NotEmpty(t, resp["pspReference"])
}

func TestBuildResponse_softRefusalHas200StatusAndRefusalFields(t *testing.T) {
	// Adyen convention: acquirer declines can come back as HTTP 200 with
	// resultCode=Refused. Our engine happens to use 402 for BankDecline,
	// but Refused also applies to other modes. Verify the shape when
	// HTTPStatus is 200 but the mode implies a decline.
	a := adyen.New()
	// NetworkError → HTTPStatus 503, so use a different mode that maps to
	// "Refused" but where we force HTTPStatus=200 for the test.
	status, raw, _ := a.BuildResponse(&engine.Result{
		Mode:       engine.ModeBankDoNotHonour,
		HTTPStatus: 200,
	}, nil)
	assert.Equal(t, 200, status)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.Equal(t, "Refused", resp["resultCode"])
	assert.Equal(t, "Do Not Honour", resp["refusalReason"])
	assert.Equal(t, "4", resp["refusalReasonCode"])
}

func TestBuildResponse_pendingMapsToReceived(t *testing.T) {
	a := adyen.New()
	_, raw, _ := a.BuildResponse(&engine.Result{
		Mode:       engine.ModePendingThenSuccess,
		HTTPStatus: 200,
		IsPending:  true,
	}, nil)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	assert.Equal(t, "Received", resp["resultCode"])
}

func TestBuildResponse_defaultAmountUSDWhenBodyEmpty(t *testing.T) {
	a := adyen.New()
	_, raw, _ := a.BuildResponse(&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}, nil)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	amount := resp["amount"].(map[string]any)
	assert.EqualValues(t, 5000, amount["value"])
	assert.Equal(t, "USD", amount["currency"])
}

func TestBuildResponse_invalidJSONBodyFallsBackToDefaults(t *testing.T) {
	a := adyen.New()
	_, raw, _ := a.BuildResponse(&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200}, []byte("not json"))
	var resp map[string]any
	require.NoError(t, json.Unmarshal(raw, &resp))
	amount := resp["amount"].(map[string]any)
	assert.EqualValues(t, 5000, amount["value"])
}

func TestBuildWebhookPayload_successShape(t *testing.T) {
	a := adyen.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200},
		"PSP12345",
		10000,
		"USD",
		map[string]any{"additionalData": map[string]any{"cardSummary": "1111"}},
	)

	assert.Equal(t, "false", out["live"])
	items, ok := out["notificationItems"].([]any)
	require.True(t, ok)
	require.Len(t, items, 1)

	wrapper := items[0].(map[string]any)
	nri := wrapper["NotificationRequestItem"].(map[string]any)

	assert.Equal(t, "AUTHORISATION", nri["eventCode"])
	assert.Equal(t, "true", nri["success"], "success MUST be string 'true' — Adyen convention")
	assert.Equal(t, "PSP12345", nri["pspReference"])
	assert.Equal(t, "scheme", nri["paymentMethod"])
	assert.Equal(t, "TestMerchantAccount", nri["merchantAccountCode"])

	amt := nri["amount"].(map[string]any)
	assert.EqualValues(t, 10000, amt["value"])
	assert.Equal(t, "USD", amt["currency"])

	addl := nri["additionalData"].(map[string]any)
	assert.Equal(t, "1111", addl["cardSummary"])
	assert.Equal(t, "123456", addl["authCode"])

	// Operations should list CANCEL/CAPTURE/REFUND — what the merchant can do
	// with this pspReference.
	ops := nri["operations"].([]any)
	assert.Contains(t, ops, "REFUND")
	assert.Contains(t, ops, "CAPTURE")
	assert.Contains(t, ops, "CANCEL")
}

func TestBuildWebhookPayload_failureSetsSuccessFalseAndReason(t *testing.T) {
	a := adyen.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeBankDeclineSoft, HTTPStatus: 402},
		"PSP_FAIL",
		5000,
		"USD",
		nil,
	)
	items := out["notificationItems"].([]any)
	nri := items[0].(map[string]any)["NotificationRequestItem"].(map[string]any)
	assert.Equal(t, "false", nri["success"])
	assert.Equal(t, "Not enough balance", nri["reason"])
}

func TestBuildWebhookPayload_reversalSetsCancellationAndOriginalReference(t *testing.T) {
	a := adyen.New()
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeSuccessThenReversed, HTTPStatus: 200},
		"PSP_ORIG",
		5000,
		"USD",
		nil,
	)
	items := out["notificationItems"].([]any)
	nri := items[0].(map[string]any)["NotificationRequestItem"].(map[string]any)
	assert.Equal(t, "CANCELLATION", nri["eventCode"])
	assert.Equal(t, "PSP_ORIG", nri["originalReference"])
}

func TestBuildWebhookPayload_requestEchoIncluded(t *testing.T) {
	a := adyen.New()
	reqBody := map[string]any{"merchantReference": "my-order-42"}
	out := a.BuildWebhookPayload(
		&engine.Result{Mode: engine.ModeSuccess, HTTPStatus: 200},
		"PSP_X", 5000, "USD", reqBody,
	)
	assert.Equal(t, reqBody, out["request_echo"])
}
