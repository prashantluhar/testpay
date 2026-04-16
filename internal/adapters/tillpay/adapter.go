// Package tillpay mocks TillPay payment API.
// Real-world: HMAC-SHA512 signing, X-Signature header.
package tillpay

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/prashantluhar/testpay/internal/engine"
)

type Adapter struct{}

func New() *Adapter             { return &Adapter{} }
func (a *Adapter) Name() string { return "tillpay" }

func (a *Adapter) BuildResponse(result *engine.Result, body []byte) (int, []byte, map[string]string) {
	h := map[string]string{
		"Content-Type": "application/json",
		"X-Signature":  "mock-sha512-signature",
	}
	txnID := fmt.Sprintf("TP%d", time.Now().UnixNano())
	orderID := fmt.Sprintf("ORD_%d", time.Now().UnixNano())

	if result.HTTPStatus >= 400 {
		b, _ := json.Marshal(map[string]any{
			"OrderId":           orderID,
			"Status":            "FAILED",
			"StatusCode":        result.HTTPStatus,
			"StatusDescription": string(result.Mode),
			"TransactionId":     txnID,
		})
		return result.HTTPStatus, b, h
	}
	amt, cur := jsonAmt(body, 5000, "USD")
	b, _ := json.Marshal(map[string]any{
		"OrderId":           orderID,
		"Status":            pickStatus(result, "SUCCESS", "PENDING"),
		"StatusCode":        200,
		"StatusDescription": "OK",
		"TransactionId":     txnID,
		"Amount":            amt,
		"Currency":          cur,
		"Timestamp":         time.Now().UTC().Format(time.RFC3339),
	})
	return 200, b, h
}

func (a *Adapter) BuildWebhookPayload(result *engine.Result, chargeID string, amount int64, currency string, req map[string]any) map[string]any {
	out := map[string]any{
		"EventType":      pickEvent(result, "payment.success", "payment.failed"),
		"OrderId":        chargeID,
		"TransactionId":  fmt.Sprintf("TP%d", time.Now().UnixNano()),
		"Status":         pickStatus(result, "SUCCESS", "PENDING"),
		"Amount":         amount,
		"Currency":       currency,
		"Timestamp":      time.Now().UTC().Format(time.RFC3339),
		"Signature":      "mock-sha512-signature",
	}
	if req != nil {
		out["RequestEcho"] = req
	}
	return out
}

func pickStatus(r *engine.Result, success, pending string) string {
	if r.HTTPStatus >= 400 {
		return "FAILED"
	}
	if r.IsPending {
		return pending
	}
	return success
}
func pickEvent(r *engine.Result, success, failed string) string {
	if r.HTTPStatus >= 400 {
		return failed
	}
	return success
}
func jsonAmt(body []byte, defAmt int64, defCur string) (int64, string) {
	amt, cur := defAmt, defCur
	var m map[string]any
	if len(body) == 0 || json.Unmarshal(body, &m) != nil {
		return amt, cur
	}
	if v, ok := m["Amount"].(float64); ok {
		amt = int64(v)
	} else if v, ok := m["amount"].(float64); ok {
		amt = int64(v)
	}
	if v, ok := m["Currency"].(string); ok && v != "" {
		cur = v
	} else if v, ok := m["currency"].(string); ok && v != "" {
		cur = v
	}
	return amt, cur
}
