// Package paynamics mocks Paynamics (PH/SEA) payment API.
// Real-world: MD5 signature, form-encoded responses. We mock as JSON.
package paynamics

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/prashantluhar/testpay/internal/engine"
)

type Adapter struct{}

func New() *Adapter             { return &Adapter{} }
func (a *Adapter) Name() string { return "paynamics" }

func (a *Adapter) BuildResponse(result *engine.Result, body []byte) (int, []byte, map[string]string) {
	h := map[string]string{"Content-Type": "application/json"}
	respID := fmt.Sprintf("PNMC_%d", time.Now().UnixNano())
	ref := fmt.Sprintf("MER_%d", time.Now().UnixNano())

	if result.HTTPStatus >= 400 {
		b, _ := json.Marshal(map[string]any{
			"response_id":      respID,
			"mer_reference_no": ref,
			"response_code":    "GR051",
			"response_message": "Transaction failed",
			"response_advice":  string(result.Mode),
		})
		return result.HTTPStatus, b, h
	}
	amt, cur := jsonAmt(body, 5000, "PHP")
	b, _ := json.Marshal(map[string]any{
		"response_id":      respID,
		"mer_reference_no": ref,
		"response_code":    "GR001",
		"response_message": "Transaction successful",
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
		"currency":         cur,
		"total_amount":     fmt.Sprintf("%.2f", float64(amt)/100),
		"signature":        "mock-md5-signature",
	})
	return 200, b, h
}

func (a *Adapter) BuildWebhookPayload(result *engine.Result, chargeID string, amount int64, currency string, req map[string]any) map[string]any {
	code := "GR001"
	msg := "Transaction successful"
	if result.HTTPStatus >= 400 {
		code = "GR051"
		msg = "Transaction failed"
	}
	out := map[string]any{
		"response_id":      chargeID,
		"mer_reference_no": fmt.Sprintf("MER_%d", time.Now().UnixNano()),
		"response_code":    code,
		"response_message": msg,
		"total_amount":     fmt.Sprintf("%.2f", float64(amount)/100),
		"currency":         currency,
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
		"signature":        "mock-md5-signature",
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
