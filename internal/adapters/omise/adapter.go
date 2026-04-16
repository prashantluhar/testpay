// Package omise mocks Omise's charge API surface.
// Real-world shape: object="charge", status, amount, currency, authorized.
package omise

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/prashantluhar/testpay/internal/engine"
)

type Adapter struct{}

func New() *Adapter             { return &Adapter{} }
func (a *Adapter) Name() string { return "omise" }

func (a *Adapter) BuildResponse(result *engine.Result, body []byte) (int, []byte, map[string]string) {
	headers := map[string]string{"Content-Type": "application/json"}

	if result.HTTPStatus >= 400 {
		errBody, _ := json.Marshal(map[string]any{
			"object":   "error",
			"location": "https://docs.opn.ooo/api-errors",
			"code":     omiseCode(result.Mode),
			"message":  omiseMessage(result.Mode),
			"status":   result.HTTPStatus,
		})
		return result.HTTPStatus, errBody, headers
	}

	amount, currency := extractAmountCurrency(body, 5000, "USD")
	resp, _ := json.Marshal(map[string]any{
		"object":     "charge",
		"id":         fmt.Sprintf("chrg_test_%d", time.Now().UnixNano()),
		"livemode":   false,
		"location":   "/charges/mock",
		"amount":     amount,
		"currency":   currency,
		"status":     omiseStatus(result.Mode),
		"authorized": !result.IsPending,
		"paid":       !result.IsPending,
		"captured":   !result.IsPending,
		"created_at": time.Now().UTC().Format(time.RFC3339),
	})
	return 200, resp, headers
}

func (a *Adapter) BuildWebhookPayload(result *engine.Result, chargeID string, amount int64, currency string, requestBody map[string]any) map[string]any {
	key := "charge.complete"
	if result.HTTPStatus >= 400 {
		key = "charge.failed"
	}
	if result.IsPending {
		key = "charge.pending"
	}

	metadata := map[string]any{}
	if requestBody != nil {
		if md, ok := requestBody["metadata"].(map[string]any); ok {
			metadata = md
		}
	}

	charge := map[string]any{
		"object":     "charge",
		"id":         chargeID,
		"amount":     amount,
		"currency":   currency,
		"status":     omiseStatus(result.Mode),
		"metadata":   metadata,
		"created_at": time.Now().UTC().Format(time.RFC3339),
	}
	if requestBody != nil {
		charge["request_echo"] = requestBody
	}

	return map[string]any{
		"object":     "event",
		"id":         fmt.Sprintf("evnt_test_%d", time.Now().UnixNano()),
		"key":        key,
		"livemode":   false,
		"created_at": time.Now().UTC().Format(time.RFC3339),
		"data":       charge,
	}
}

func omiseStatus(mode engine.FailureMode) string {
	switch mode {
	case engine.ModeSuccess, engine.ModeDoubleCharge, engine.ModeWebhookMissing,
		engine.ModeWebhookDelayed, engine.ModeWebhookDuplicate, engine.ModeRedirectSuccess:
		return "successful"
	case engine.ModePendingThenFailed, engine.ModePendingThenSuccess, engine.ModeFailedThenSuccess:
		return "pending"
	default:
		return "failed"
	}
}

func omiseCode(mode engine.FailureMode) string {
	m := map[engine.FailureMode]string{
		engine.ModeBankDeclineHard: "insufficient_fund",
		engine.ModeBankDeclineSoft: "insufficient_fund",
		engine.ModeBankInvalidCVV:  "invalid_security_code",
		engine.ModeBankTimeout:     "payment_timeout",
		engine.ModePGRateLimited:   "rate_limit",
	}
	if v, ok := m[mode]; ok {
		return v
	}
	return "failed_fraud_check"
}

func omiseMessage(mode engine.FailureMode) string {
	m := map[engine.FailureMode]string{
		engine.ModeBankDeclineHard: "Insufficient funds on the card",
		engine.ModeBankInvalidCVV:  "The security code is invalid",
		engine.ModeBankTimeout:     "The payment timed out",
		engine.ModePGRateLimited:   "Too many requests",
	}
	if v, ok := m[mode]; ok {
		return v
	}
	return "The charge failed"
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
	switch v := m["amount"].(type) {
	case float64:
		amount = int64(v)
	}
	if c, ok := m["currency"].(string); ok && c != "" {
		currency = c
	}
	return amount, currency
}
