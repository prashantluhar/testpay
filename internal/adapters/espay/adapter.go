// Package espay mocks ESPay (Indonesia) payment API.
// Real-world: RSA-2048 signed payloads. We mock with fake signature fields.
package espay

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/prashantluhar/testpay/internal/engine"
)

type Adapter struct{}

func New() *Adapter             { return &Adapter{} }
func (a *Adapter) Name() string { return "espay" }

func (a *Adapter) BuildResponse(result *engine.Result, body []byte) (int, []byte, map[string]string) {
	h := map[string]string{"Content-Type": "application/json"}
	if result.HTTPStatus >= 400 {
		b, _ := json.Marshal(map[string]any{
			"error_code":    "ESP99",
			"error_message": string(result.Mode),
			"order_id":      fmt.Sprintf("ESP_%d", time.Now().UnixNano()),
			"signature":     "mock-rsa-signature",
		})
		return result.HTTPStatus, b, h
	}
	amt, cur := jsonAmt(body, 5000, "IDR")
	b, _ := json.Marshal(map[string]any{
		"error_code":    "0000",
		"error_message": "Success",
		"order_id":      fmt.Sprintf("ESP_%d", time.Now().UnixNano()),
		"payment_code":  fmt.Sprintf("VA_%d", time.Now().UnixNano()),
		"amount":        amt,
		"currency":      cur,
		"signature":     "mock-rsa-signature",
		"expired_time":  time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339),
	})
	return 200, b, h
}

func (a *Adapter) BuildWebhookPayload(result *engine.Result, chargeID string, amount int64, currency string, req map[string]any) map[string]any {
	code := "0000"
	msg := "Success"
	if result.HTTPStatus >= 400 {
		code, msg = "ESP99", "Transaction failed"
	}
	out := map[string]any{
		"error_code":    code,
		"error_message": msg,
		"order_id":      chargeID,
		"amount":        amount,
		"currency":      currency,
		"signature":     "mock-rsa-signature",
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
	}
	if req != nil {
		out["request_echo"] = req
	}
	return out
}

func jsonAmt(body []byte, defAmt int64, defCur string) (int64, string) {
	amt, cur := defAmt, defCur
	var m map[string]any
	if len(body) == 0 || json.Unmarshal(body, &m) != nil {
		return amt, cur
	}
	if v, ok := m["amount"].(float64); ok {
		amt = int64(v)
	}
	if v, ok := m["currency"].(string); ok && v != "" {
		cur = v
	}
	return amt, cur
}
