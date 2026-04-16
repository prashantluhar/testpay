// Package komoju mocks Komoju (JP) payment API.
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
	h := map[string]string{"Content-Type": "application/json"}
	if result.HTTPStatus >= 400 {
		b, _ := json.Marshal(map[string]any{
			"error": map[string]any{
				"code":    "bad_request",
				"message": "Payment failed",
				"param":   string(result.Mode),
			},
		})
		return result.HTTPStatus, b, h
	}
	amount, currency := jsonNumAndStr(body, "amount", 5000, "currency", "JPY")
	status := "captured"
	if result.IsPending {
		status = "authorized"
	}
	b, _ := json.Marshal(map[string]any{
		"id":               fmt.Sprintf("komoju_%d", time.Now().UnixNano()),
		"resource":         "payment",
		"status":           status,
		"amount":           amount,
		"currency":         currency,
		"payment_deadline": time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339),
		"captured_at":      time.Now().UTC().Format(time.RFC3339),
	})
	return 200, b, h
}

func (a *Adapter) BuildWebhookPayload(result *engine.Result, chargeID string, amount int64, currency string, req map[string]any) map[string]any {
	event := "payment.captured"
	if result.HTTPStatus >= 400 {
		event = "payment.failed"
	}
	data := map[string]any{
		"id":       chargeID,
		"resource": "payment",
		"status":   pick(result, "captured", "authorized", "failed"),
		"amount":   amount,
		"currency": currency,
	}
	if req != nil {
		data["metadata"] = nestedOrEmpty(req, "metadata")
		data["request_echo"] = req
	}
	return map[string]any{
		"type":       event,
		"resource":   "event",
		"data":       data,
		"created_at": time.Now().UTC().Format(time.RFC3339),
	}
}

// helpers reused across adapters added in this batch
func pick(r *engine.Result, success, pending, failed string) string {
	if r.HTTPStatus >= 400 {
		return failed
	}
	if r.IsPending {
		return pending
	}
	return success
}
func nestedOrEmpty(m map[string]any, key string) map[string]any {
	if n, ok := m[key].(map[string]any); ok {
		return n
	}
	return map[string]any{}
}
func jsonNumAndStr(body []byte, amtKey string, defAmt int64, curKey, defCur string) (int64, string) {
	amt, cur := defAmt, defCur
	var m map[string]any
	if len(body) == 0 || json.Unmarshal(body, &m) != nil {
		return amt, cur
	}
	if v, ok := m[amtKey].(float64); ok {
		amt = int64(v)
	}
	if v, ok := m[curKey].(string); ok && v != "" {
		cur = v
	}
	return amt, cur
}
