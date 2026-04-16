// Package instamojo mocks Instamojo (IN) payment API.
package instamojo

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/prashantluhar/testpay/internal/engine"
)

type Adapter struct{}

func New() *Adapter             { return &Adapter{} }
func (a *Adapter) Name() string { return "instamojo" }

func (a *Adapter) BuildResponse(result *engine.Result, body []byte) (int, []byte, map[string]string) {
	h := map[string]string{"Content-Type": "application/json"}
	if result.HTTPStatus >= 400 {
		b, _ := json.Marshal(map[string]any{
			"success": false,
			"message": "Payment request creation failed",
			"errors":  map[string]any{"code": []string{string(result.Mode)}},
		})
		return result.HTTPStatus, b, h
	}
	id := fmt.Sprintf("MOJO_%d", time.Now().UnixNano())
	amt, cur := extractAmount(body)
	b, _ := json.Marshal(map[string]any{
		"success": true,
		"payment_request": map[string]any{
			"id":         id,
			"status":     pickStatus(result),
			"amount":     fmt.Sprintf("%d", amt),
			"currency":   cur,
			"longurl":    fmt.Sprintf("https://instamojo.com/@mock/%s", id),
			"created_at": time.Now().UTC().Format(time.RFC3339),
			"modified_at": time.Now().UTC().Format(time.RFC3339),
		},
	})
	return 200, b, h
}

func (a *Adapter) BuildWebhookPayload(result *engine.Result, chargeID string, amount int64, currency string, req map[string]any) map[string]any {
	out := map[string]any{
		"payment_id":     chargeID,
		"status":         pickStatus(result),
		"amount":         fmt.Sprintf("%d", amount),
		"currency":       currency,
		"buyer":          "mock@example.com",
		"created_at":     time.Now().UTC().Format(time.RFC3339),
		"payment_request_id": fmt.Sprintf("MOJO_%d", time.Now().UnixNano()),
	}
	if req != nil {
		if ref, ok := req["purpose"].(string); ok {
			out["purpose"] = ref
		}
		out["request_echo"] = req
	}
	return out
}

func pickStatus(r *engine.Result) string {
	if r.HTTPStatus >= 400 {
		return "Failed"
	}
	if r.IsPending {
		return "Pending"
	}
	return "Credit"
}

func extractAmount(body []byte) (int64, string) {
	amt, cur := int64(5000), "INR"
	var m map[string]any
	if len(body) == 0 || json.Unmarshal(body, &m) != nil {
		return amt, cur
	}
	switch v := m["amount"].(type) {
	case float64:
		amt = int64(v)
	case string:
		// instamojo sometimes accepts string-encoded amounts
	}
	if c, ok := m["currency"].(string); ok && c != "" {
		cur = c
	}
	return amt, cur
}
