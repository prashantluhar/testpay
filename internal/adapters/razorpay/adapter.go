// Package razorpay mocks Razorpay's payment API surface.
// Real-world shape: payment entity with id/entity/amount/currency/status,
// and webhook envelope { entity: "event", event, contains, payload, created_at }.
package razorpay

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/prashantluhar/testpay/internal/engine"
)

type Adapter struct{}

func New() *Adapter              { return &Adapter{} }
func (a *Adapter) Name() string  { return "razorpay" }

func (a *Adapter) BuildResponse(result *engine.Result, body []byte) (int, []byte, map[string]string) {
	headers := map[string]string{"Content-Type": "application/json"}
	payID := fmt.Sprintf("pay_%d", time.Now().UnixNano())

	// HTTP 4xx — Razorpay returns the `{ "error": {...} }` envelope before
	// the request ever produces a payment entity.
	if result.HTTPStatus >= 400 {
		errResp := errorResponse{
			Error: errorBody{
				Code:        razorpayErrorCode(result.Mode),
				Description: razorpayErrorDescription(result.Mode),
				Source:      razorpayErrorSource(result.Mode),
				Step:        razorpayErrorStep(result.Mode),
				Reason:      string(result.Mode),
			},
		}
		out, _ := json.Marshal(errResp)
		return result.HTTPStatus, out, headers
	}

	amount, currency := extractAmountCurrency(body, 5000, "INR")
	notes := extractNotes(body)

	resp := paymentEntity{
		ID:       payID,
		Entity:   "payment",
		Amount:   amount,
		Currency: currency,
		Status:   razorpayStatus(result.Mode),
		Method:   "card",
		Notes:    notes,
	}
	// Failed-in-band (status == "failed" at HTTP 200) — Razorpay populates
	// error_* on the payment entity itself. Our engine sends >=400 for most
	// failures, but modes like ModeAmountMismatch land here.
	if resp.Status == statusFailed {
		resp.ErrorCode = razorpayErrorCode(result.Mode)
		resp.ErrorDescription = razorpayErrorDescription(result.Mode)
		resp.ErrorSource = razorpayErrorSource(result.Mode)
		resp.ErrorStep = razorpayErrorStep(result.Mode)
		resp.ErrorReason = string(result.Mode)
	}
	out, _ := json.Marshal(resp)
	return 200, out, headers
}

func (a *Adapter) BuildWebhookPayload(result *engine.Result, chargeID string, amount int64, currency string, requestBody map[string]any) map[string]any {
	status := razorpayStatus(result.Mode)
	event := razorpayEvent(result.Mode, result.HTTPStatus)
	// A webhook firing on a 4xx-rejected payment is still a "payment.failed"
	// event, with the entity status reflecting failure.
	if result.HTTPStatus >= 400 {
		status = statusFailed
	}

	// Razorpay convention: echo `notes` from the charge request into every
	// downstream artefact so the merchant can correlate on their own keys.
	notes := map[string]any{}
	if requestBody != nil {
		if n, ok := requestBody["notes"].(map[string]any); ok {
			notes = n
		}
	}

	entity := paymentEntity{
		ID:       chargeID,
		Entity:   "payment",
		Amount:   amount,
		Currency: currency,
		Status:   status,
		Method:   "card",
		Notes:    notes,
	}
	if status == statusFailed {
		entity.ErrorCode = razorpayErrorCode(result.Mode)
		entity.ErrorDescription = razorpayErrorDescription(result.Mode)
		entity.ErrorSource = razorpayErrorSource(result.Mode)
		entity.ErrorStep = razorpayErrorStep(result.Mode)
		entity.ErrorReason = string(result.Mode)
	}

	env := webhookEnvelope{
		Entity:    "event",
		Event:     event,
		Contains:  []string{"payment"},
		CreatedAt: time.Now().Unix(),
		Payload: webhookPayload{
			Payment: webhookPaymentWrapper{Entity: entity},
		},
	}

	// Round-trip through JSON so the dispatcher receives a map[string]any
	// consistent with its webhook contract, while we keep the typed DTO at
	// the adapter boundary.
	raw, _ := json.Marshal(env)
	var out map[string]any
	_ = json.Unmarshal(raw, &out)
	if requestBody != nil {
		out["request_echo"] = requestBody
	}
	return out
}

// razorpayStatus maps the engine's failure-mode taxonomy onto Razorpay's
// payment status vocabulary. Pending modes resolve to "authorized" — money
// is held at the issuer but not yet captured; the webhook will later
// capture or fail it.
func razorpayStatus(mode engine.FailureMode) string {
	switch mode {
	case engine.ModeSuccess, engine.ModeWebhookMissing, engine.ModeWebhookDelayed,
		engine.ModeWebhookDuplicate, engine.ModeWebhookOutOfOrder,
		engine.ModeWebhookMalformed, engine.ModePartialSuccess,
		engine.ModeDoubleCharge, engine.ModeRedirectSuccess:
		return statusCaptured
	case engine.ModePendingThenFailed, engine.ModePendingThenSuccess,
		engine.ModeFailedThenSuccess, engine.ModeSuccessThenReversed:
		return statusAuthorized
	case engine.ModeAmountMismatch:
		// Razorpay treats amount mismatches as in-band failures at HTTP 200.
		return statusFailed
	default:
		return statusFailed
	}
}

// razorpayEvent maps the engine mode onto the webhook event vocabulary.
// Pending/authorized modes fire payment.authorized; everything else is
// either payment.captured (success) or payment.failed.
func razorpayEvent(mode engine.FailureMode, httpStatus int) string {
	if httpStatus >= 400 {
		return eventPaymentFailed
	}
	switch mode {
	case engine.ModePendingThenFailed, engine.ModePendingThenSuccess,
		engine.ModeFailedThenSuccess, engine.ModeSuccessThenReversed:
		return eventPaymentAuthorized
	case engine.ModeAmountMismatch:
		return eventPaymentFailed
	case engine.ModeSuccess, engine.ModeWebhookMissing, engine.ModeWebhookDelayed,
		engine.ModeWebhookDuplicate, engine.ModeWebhookOutOfOrder,
		engine.ModeWebhookMalformed, engine.ModePartialSuccess,
		engine.ModeDoubleCharge, engine.ModeRedirectSuccess:
		return eventPaymentCaptured
	default:
		return eventPaymentFailed
	}
}

// razorpayErrorCode maps engine modes to Razorpay's error.code vocabulary.
// The public code set is a small enum — BAD_REQUEST_ERROR, GATEWAY_ERROR,
// SERVER_ERROR — with the detailed reason in error.reason.
// See https://razorpay.com/docs/api/errors/
func razorpayErrorCode(mode engine.FailureMode) string {
	switch mode {
	case engine.ModeBankDeclineHard, engine.ModeBankDeclineSoft,
		engine.ModeBankInvalidCVV, engine.ModeBankDoNotHonour,
		engine.ModeBankServerDown, engine.ModeBankTimeout:
		return "GATEWAY_ERROR"
	case engine.ModePGServerError, engine.ModePGMaintenance, engine.ModePGTimeout,
		engine.ModeNetworkError:
		return "SERVER_ERROR"
	case engine.ModePGRateLimited:
		return "BAD_REQUEST_ERROR"
	case engine.ModeAmountMismatch:
		return "BAD_REQUEST_ERROR"
	case engine.ModeRedirectAbandoned, engine.ModeRedirectTimeout,
		engine.ModeRedirectFailed:
		return "BAD_REQUEST_ERROR"
	default:
		return "BAD_REQUEST_ERROR"
	}
}

func razorpayErrorDescription(mode engine.FailureMode) string {
	m := map[engine.FailureMode]string{
		engine.ModeBankDeclineHard:   "Payment was declined by the bank",
		engine.ModeBankDeclineSoft:   "Your payment could not be completed due to insufficient balance",
		engine.ModeBankInvalidCVV:    "Invalid CVV provided",
		engine.ModeBankDoNotHonour:   "Payment could not be processed. Please contact your bank",
		engine.ModeBankTimeout:       "Bank is not responding. Please try again",
		engine.ModeBankServerDown:    "Bank servers are down. Please try again later",
		engine.ModePGTimeout:         "Gateway request timed out",
		engine.ModePGServerError:     "Internal server error at gateway",
		engine.ModePGRateLimited:     "Too many requests — rate limit exceeded",
		engine.ModePGMaintenance:     "Gateway under maintenance. Please try again later",
		engine.ModeNetworkError:      "Network error while reaching gateway",
		engine.ModeAmountMismatch:    "Amount does not match the order amount",
		engine.ModeRedirectAbandoned: "Payment cancelled by the customer",
		engine.ModeRedirectTimeout:   "Checkout session expired",
		engine.ModeRedirectFailed:    "Redirect-flow payment failed",
	}
	if v, ok := m[mode]; ok {
		return v
	}
	return "Payment failed"
}

// razorpayErrorSource tags which side of the pipeline owned the failure —
// Razorpay splits this between "customer", "business", "bank", and
// "gateway" so merchants can triage at a glance.
func razorpayErrorSource(mode engine.FailureMode) string {
	switch mode {
	case engine.ModeBankDeclineHard, engine.ModeBankDeclineSoft,
		engine.ModeBankInvalidCVV, engine.ModeBankDoNotHonour,
		engine.ModeBankServerDown, engine.ModeBankTimeout:
		return "bank"
	case engine.ModePGServerError, engine.ModePGMaintenance, engine.ModePGTimeout,
		engine.ModePGRateLimited, engine.ModeNetworkError:
		return "gateway"
	case engine.ModeAmountMismatch:
		return "business"
	case engine.ModeRedirectAbandoned, engine.ModeRedirectTimeout,
		engine.ModeRedirectFailed:
		return "customer"
	default:
		return "gateway"
	}
}

// razorpayErrorStep tags the lifecycle phase in which the failure surfaced.
// Razorpay uses "payment_authentication", "payment_authorization",
// "payment_initiation", "payment_capture".
func razorpayErrorStep(mode engine.FailureMode) string {
	switch mode {
	case engine.ModeBankInvalidCVV, engine.ModeRedirectAbandoned,
		engine.ModeRedirectTimeout, engine.ModeRedirectFailed:
		return "payment_authentication"
	case engine.ModeBankDeclineHard, engine.ModeBankDeclineSoft,
		engine.ModeBankDoNotHonour:
		return "payment_authorization"
	case engine.ModePGServerError, engine.ModePGTimeout, engine.ModePGRateLimited,
		engine.ModePGMaintenance, engine.ModeNetworkError,
		engine.ModeBankServerDown, engine.ModeBankTimeout:
		return "payment_initiation"
	case engine.ModeAmountMismatch:
		return "payment_capture"
	default:
		return "payment_initiation"
	}
}

func extractAmountCurrency(body []byte, defAmount int64, defCurrency string) (int64, string) {
	amount, currency := defAmount, defCurrency
	if len(body) == 0 {
		return amount, currency
	}
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		return amount, currency
	}
	// Razorpay flattens amount at the top level (unlike Adyen's nested shape).
	switch v := m["amount"].(type) {
	case float64:
		amount = int64(v)
	case int64:
		amount = v
	case int:
		amount = int64(v)
	}
	if c, ok := m["currency"].(string); ok && c != "" {
		currency = c
	}
	return amount, currency
}

func extractNotes(body []byte) map[string]any {
	notes := map[string]any{}
	if len(body) == 0 {
		return notes
	}
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		return notes
	}
	if n, ok := m["notes"].(map[string]any); ok {
		return n
	}
	return notes
}
