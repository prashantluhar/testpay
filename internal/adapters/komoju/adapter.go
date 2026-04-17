// Package komoju mocks Komoju (Japan) payment API.
// Real-world shape: resource-oriented JSON — `{ id, resource: "payment",
// status, amount, ... }` on success; `{ error: { code, message, ... } }`
// on failure. Webhook is wrapped in an event envelope:
//   { id, type: "payment.<state>", resource: "event", data: {...payment...} }
package komoju

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/prashantluhar/testpay/internal/engine"
)

type Adapter struct{}

func New() *Adapter             { return &Adapter{} }
func (a *Adapter) Name() string { return "komoju" }

func (a *Adapter) BuildResponse(result *engine.Result, body []byte) (int, []byte, map[string]string) {
	headers := map[string]string{"Content-Type": "application/json"}
	id := fmt.Sprintf("komoju_%d", time.Now().UnixNano())

	if result.HTTPStatus >= 400 {
		errResp := errorResponse{
			Error: errorBody{
				Code:    komojuErrorCode(result.Mode),
				Message: komojuMessage(result.Mode),
				Param:   string(result.Mode),
			},
		}
		out, _ := json.Marshal(errResp)
		return result.HTTPStatus, out, headers
	}

	amount, currency := extractAmountCurrency(body, 5000, "JPY")
	now := time.Now().UTC()
	var captured *time.Time
	status := komojuStatus(result)
	if status == "captured" {
		cap := now
		captured = &cap
	}

	resp := paymentResource{
		ID:               id,
		Resource:         "payment",
		Status:           status,
		Amount:           int(amount),
		Tax:              0,
		Total:            int(amount),
		Currency:         currency,
		Description:      "Mock Komoju charge",
		PaymentMethodFee: 0,
		PaymentDetails: paymentDetails{
			Type:        "credit_card",
			Email:       "mock@example.com",
			RedirectURL: fmt.Sprintf("https://komoju.com/mock/%s", id),
		},
		CapturedAt:     captured,
		CreatedAt:      now,
		AmountRefunded: 0,
		Locale:         "ja",
		Metadata:       map[string]any{},
		Refunds:        []refundSummary{},
	}
	out, _ := json.Marshal(resp)
	return 200, out, headers
}

func (a *Adapter) BuildWebhookPayload(result *engine.Result, chargeID string, amount int64, currency string, requestBody map[string]any) map[string]any {
	now := time.Now().UTC()
	eventType := komojuEventType(result)

	var captured *time.Time
	status := komojuStatus(result)
	if status == "captured" {
		cap := now
		captured = &cap
	}

	metadata := map[string]any{}
	if requestBody != nil {
		if m, ok := requestBody["metadata"].(map[string]any); ok {
			metadata = m
		}
	}

	data := paymentResource{
		ID:               chargeID,
		Resource:         "payment",
		Status:           status,
		Amount:           int(amount),
		Total:            int(amount),
		Currency:         currency,
		PaymentMethodFee: 0,
		PaymentDetails: paymentDetails{
			Type:  "credit_card",
			Email: "mock@example.com",
		},
		CapturedAt: captured,
		CreatedAt:  now,
		Metadata:   metadata,
		Refunds:    []refundSummary{},
	}

	envelope := webhookEnvelope{
		ID:        fmt.Sprintf("event_%d", time.Now().UnixNano()),
		Type:      eventType,
		Resource:  "event",
		Data:      data,
		CreatedAt: now,
	}
	if result.HTTPStatus >= 400 {
		envelope.Reason = komojuMessage(result.Mode)
		envelope.Details = &webhookDetails{ErrorCode: komojuErrorCode(result.Mode)}
	}

	raw, _ := json.Marshal(envelope)
	var out map[string]any
	_ = json.Unmarshal(raw, &out)

	if requestBody != nil {
		out["request_echo"] = requestBody
	}
	return out
}

// komojuStatus maps engine modes onto Komoju's status vocabulary:
//   captured  — terminal success
//   authorized — awaiting capture (pending modes land here)
//   failed     — terminal failure
//   cancelled  — reversed
func komojuStatus(r *engine.Result) string {
	switch {
	case r.HTTPStatus >= 400:
		return "failed"
	case r.Mode == engine.ModeSuccessThenReversed:
		return "cancelled"
	case r.IsPending:
		return "authorized"
	default:
		return "captured"
	}
}

// komojuEventType maps modes onto Komoju's webhook event namespace.
func komojuEventType(r *engine.Result) string {
	switch {
	case r.HTTPStatus >= 400:
		return "payment.failed"
	case r.Mode == engine.ModeSuccessThenReversed:
		return "payment.refunded"
	case r.IsPending:
		return "payment.authorized"
	default:
		return "payment.captured"
	}
}

// komojuErrorCode maps engine modes onto Komoju's string error-code set
// (see https://docs.komoju.com/en/api/overview/#error-codes).
func komojuErrorCode(mode engine.FailureMode) string {
	m := map[engine.FailureMode]string{
		engine.ModeBankDeclineHard: "card_declined",
		engine.ModeBankDeclineSoft: "insufficient_funds",
		engine.ModeBankInvalidCVV:  "invalid_cvv",
		engine.ModeBankDoNotHonour: "card_declined",
		engine.ModeBankTimeout:     "processing_error",
		engine.ModeBankServerDown:  "processing_error",
		engine.ModePGTimeout:       "request_timeout",
		engine.ModePGServerError:   "internal_server_error",
		engine.ModePGRateLimited:   "rate_limit_exceeded",
		engine.ModePGMaintenance:   "service_unavailable",
		engine.ModeNetworkError:    "network_error",
	}
	if v, ok := m[mode]; ok {
		return v
	}
	return "bad_request"
}

func komojuMessage(mode engine.FailureMode) string {
	m := map[engine.FailureMode]string{
		engine.ModeBankDeclineHard: "The card was declined.",
		engine.ModeBankDeclineSoft: "Insufficient funds on card.",
		engine.ModeBankInvalidCVV:  "Invalid security code.",
		engine.ModeBankDoNotHonour: "Do not honour.",
		engine.ModeBankTimeout:     "Bank timed out.",
		engine.ModeBankServerDown:  "Bank is unavailable.",
		engine.ModePGTimeout:       "The request timed out.",
		engine.ModePGServerError:   "An internal error occurred.",
		engine.ModePGRateLimited:   "Too many requests.",
		engine.ModePGMaintenance:   "Komoju is temporarily unavailable.",
		engine.ModeNetworkError:    "A network error occurred.",
	}
	if v, ok := m[mode]; ok {
		return v
	}
	return "Payment failed."
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
	if v, ok := m["amount"].(float64); ok {
		amount = int64(v)
	}
	if v, ok := m["currency"].(string); ok && v != "" {
		currency = v
	}
	return amount, currency
}
