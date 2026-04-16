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
	if result.HTTPStatus >= 400 {
		errBody, _ := json.Marshal(map[string]any{
			"error": map[string]any{
				"code":        "BAD_REQUEST_ERROR",
				"description": "Payment failed",
				"reason":      string(result.Mode),
			},
		})
		return result.HTTPStatus, errBody, headers
	}
	amount, currency := extractAmountCurrency(body, 5000, "INR")
	resp, _ := json.Marshal(map[string]any{
		"id":       fmt.Sprintf("pay_%d", time.Now().UnixNano()),
		"entity":   "payment",
		"status":   "captured",
		"amount":   amount,
		"currency": currency,
	})
	return 200, resp, headers
}

func (a *Adapter) BuildWebhookPayload(result *engine.Result, chargeID string, amount int64, currency string, requestBody map[string]any) map[string]any {
	event := "payment.captured"
	if result.HTTPStatus >= 400 {
		event = "payment.failed"
	}
	if result.IsPending {
		event = "payment.authorized"
	}

	// Razorpay convention: echo `notes` (or whatever the caller sent) back.
	notes := map[string]any{}
	if requestBody != nil {
		if n, ok := requestBody["notes"].(map[string]any); ok {
			notes = n
		}
	}

	entity := map[string]any{
		"id":       chargeID,
		"amount":   amount,
		"currency": currency,
		"status":   "captured",
		"notes":    notes,
	}
	if requestBody != nil {
		entity["request_echo"] = requestBody
	}

	return map[string]any{
		"entity":     "event",
		"event":      event,
		"contains":   []string{"payment"},
		"created_at": time.Now().Unix(),
		"payload": map[string]any{
			"payment": map[string]any{
				"entity": entity,
			},
		},
	}
}

func extractAmountCurrency(body []byte, defAmount int64, defCurrency string) (int64, string) {
	amount := defAmount
	currency := defCurrency
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
