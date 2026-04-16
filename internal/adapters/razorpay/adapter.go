package razorpay

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/prashantluhar/testpay/internal/engine"
)

type Adapter struct{}

func New() *Adapter         { return &Adapter{} }
func (a *Adapter) Name() string { return "razorpay" }

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
	resp, _ := json.Marshal(map[string]any{
		"id":       fmt.Sprintf("pay_%d", time.Now().UnixNano()),
		"entity":   "payment",
		"status":   "captured",
		"amount":   5000,
		"currency": "INR",
	})
	return 200, resp, headers
}

func (a *Adapter) BuildWebhookPayload(result *engine.Result, chargeID string, amount int64, currency string) map[string]any {
	event := "payment.captured"
	if result.HTTPStatus >= 400 {
		event = "payment.failed"
	}
	if result.IsPending {
		event = "payment.authorized"
	}
	return map[string]any{
		"entity":     "event",
		"event":      event,
		"contains":   []string{"payment"},
		"created_at": time.Now().Unix(),
		"payload": map[string]any{
			"payment": map[string]any{
				"entity": map[string]any{
					"id":       chargeID,
					"amount":   amount,
					"currency": currency,
					"status":   "captured",
				},
			},
		},
	}
}
