// Package epay mocks Epay payment API.
package epay

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/prashantluhar/testpay/internal/engine"
)

type Adapter struct{}

func New() *Adapter             { return &Adapter{} }
func (a *Adapter) Name() string { return "epay" }

func (a *Adapter) BuildResponse(result *engine.Result, body []byte) (int, []byte, map[string]string) {
	h := map[string]string{"Content-Type": "application/json"}
	if result.HTTPStatus >= 400 {
		b, _ := json.Marshal(map[string]any{
			"merchant_no":    "TEST_MERCHANT",
			"order_no":       fmt.Sprintf("EP_%d", time.Now().UnixNano()),
			"status_code":    "FAILED",
			"status_message": string(result.Mode),
		})
		return result.HTTPStatus, b, h
	}
	amt, cur := jsonAmt(body, 5000, "VND")
	b, _ := json.Marshal(map[string]any{
		"merchant_no":    "TEST_MERCHANT",
		"order_no":       fmt.Sprintf("EP_%d", time.Now().UnixNano()),
		"status_code":    "SUCCESS",
		"status_message": "OK",
		"amount":         amt,
		"currency":       cur,
		"signature":      "mock-sha256-signature",
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	})
	return 200, b, h
}

func (a *Adapter) BuildWebhookPayload(result *engine.Result, chargeID string, amount int64, currency string, req map[string]any) map[string]any {
	status := "SUCCESS"
	if result.HTTPStatus >= 400 {
		status = "FAILED"
	}
	out := map[string]any{
		"merchant_no":    "TEST_MERCHANT",
		"order_no":       chargeID,
		"status_code":    status,
		"status_message": "OK",
		"amount":         amount,
		"currency":       currency,
		"signature":      "mock-sha256-signature",
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
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
