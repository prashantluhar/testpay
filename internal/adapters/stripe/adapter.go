package stripe

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/prashantluhar/testpay/internal/engine"
)

type Adapter struct{}

func New() *Adapter              { return &Adapter{} }
func (a *Adapter) Name() string  { return "stripe" }

func (a *Adapter) BuildResponse(result *engine.Result, body []byte) (int, []byte, map[string]string) {
	headers := map[string]string{"Content-Type": "application/json"}

	if result.HTTPStatus >= 400 {
		errBody, _ := json.Marshal(map[string]any{
			"error": map[string]any{
				"type":    "card_error",
				"code":    result.ErrorCode,
				"message": errorMessage(result.Mode),
			},
		})
		return result.HTTPStatus, errBody, headers
	}

	amount, currency := extractAmountCurrency(body)

	if result.IsPending {
		respBody, _ := json.Marshal(map[string]any{
			"id":       fmt.Sprintf("pi_%d", time.Now().UnixNano()),
			"object":   "payment_intent",
			"status":   "processing",
			"amount":   amount,
			"currency": currency,
		})
		return 200, respBody, headers
	}

	respBody, _ := json.Marshal(map[string]any{
		"id":       fmt.Sprintf("pi_%d", time.Now().UnixNano()),
		"object":   "payment_intent",
		"status":   "succeeded",
		"amount":   amount,
		"currency": currency,
	})
	return 200, respBody, headers
}

func (a *Adapter) BuildWebhookPayload(result *engine.Result, chargeID string, amount int64, currency string, requestBody map[string]any) map[string]any {
	eventType := "payment_intent.succeeded"
	if result.IsPending {
		eventType = "payment_intent.processing"
	}
	if result.HTTPStatus >= 400 {
		eventType = "payment_intent.payment_failed"
	}

	// Echo caller-supplied metadata (matches real Stripe behavior).
	metadata := map[string]any{}
	if requestBody != nil {
		if md, ok := requestBody["metadata"].(map[string]any); ok {
			metadata = md
		}
	}

	object := map[string]any{
		"id":       chargeID,
		"object":   "payment_intent",
		"amount":   amount,
		"currency": currency,
		"status":   statusFromMode(result.Mode),
		"metadata": metadata,
	}
	if requestBody != nil {
		object["request_echo"] = requestBody
	}

	return map[string]any{
		"id":      fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		"type":    eventType,
		"created": time.Now().Unix(),
		"data": map[string]any{
			"object": object,
		},
	}
}

func errorMessage(mode engine.FailureMode) string {
	messages := map[engine.FailureMode]string{
		engine.ModeBankDeclineHard:   "Your card was declined.",
		engine.ModeBankDeclineSoft:   "Insufficient funds.",
		engine.ModeBankInvalidCVV:    "Your card's security code is incorrect.",
		engine.ModeBankDoNotHonour:   "Your card was declined.",
		engine.ModeBankServerDown:    "An error occurred communicating with your bank.",
		engine.ModeBankTimeout:       "The bank did not respond in time.",
		engine.ModePGServerError:     "An unexpected error occurred.",
		engine.ModePGTimeout:         "The request timed out.",
		engine.ModePGRateLimited:     "Too many requests.",
		engine.ModePGMaintenance:     "The service is under maintenance.",
		engine.ModeRedirectFailed:    "Authentication failed.",
		engine.ModeRedirectAbandoned: "Authentication was not completed.",
		engine.ModeRedirectTimeout:   "Authentication timed out.",
	}
	if msg, ok := messages[mode]; ok {
		return msg
	}
	return "An error occurred."
}

func statusFromMode(mode engine.FailureMode) string {
	switch mode {
	case engine.ModeSuccess,
		engine.ModeWebhookMissing, engine.ModeWebhookDelayed, engine.ModeWebhookDuplicate,
		engine.ModeWebhookOutOfOrder, engine.ModeWebhookMalformed,
		engine.ModeDoubleCharge, engine.ModeAmountMismatch, engine.ModePartialSuccess,
		engine.ModeRedirectSuccess:
		return "succeeded"
	case engine.ModePendingThenFailed, engine.ModePendingThenSuccess, engine.ModeFailedThenSuccess:
		return "processing"
	case engine.ModeSuccessThenReversed:
		// Reversed after a success — Stripe flips to canceled.
		return "canceled"
	default:
		// All declines, timeouts, redirect failures, network errors, etc.
		// Stripe's payment_intent state after a failed charge is requires_payment_method.
		return "requires_payment_method"
	}
}

// extractAmountCurrency pulls amount + currency from the request body
// (with sensible defaults) so the mock response reflects the caller's intent.
func extractAmountCurrency(body []byte) (int64, string) {
	var amount int64 = 5000
	currency := "usd"
	if len(body) == 0 {
		return amount, currency
	}
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		return amount, currency
	}
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
