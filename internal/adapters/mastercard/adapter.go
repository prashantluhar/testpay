// Package mastercard mocks the Mastercard Gateway API (MPGS).
// Real-world shape: result + gatewayCode + transactionId/orderId.
package mastercard

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/prashantluhar/testpay/internal/engine"
)

type Adapter struct{}

func New() *Adapter             { return &Adapter{} }
func (a *Adapter) Name() string { return "mastercard" }

func (a *Adapter) BuildResponse(result *engine.Result, body []byte) (int, []byte, map[string]string) {
	headers := map[string]string{"Content-Type": "application/json"}

	resultField := "SUCCESS"
	gwCode := "APPROVED"
	if result.HTTPStatus >= 400 {
		resultField = "FAILURE"
		gwCode = mastercardGatewayCode(result.Mode)
	} else if result.IsPending {
		resultField = "PENDING"
		gwCode = "PENDING"
	}

	amount, currency := extractAmountCurrency(body, 5000, "USD")
	orderId := fmt.Sprintf("ORD%d", time.Now().UnixNano())
	txnId := fmt.Sprintf("TXN%d", time.Now().UnixNano())

	resp, _ := json.Marshal(map[string]any{
		"result":      resultField,
		"order":       map[string]any{"id": orderId, "amount": amount, "currency": currency},
		"transaction": map[string]any{"id": txnId, "amount": amount, "currency": currency, "type": "PAYMENT"},
		"response":    map[string]any{"gatewayCode": gwCode, "acquirerCode": "00"},
		"timeOfRecord": time.Now().UTC().Format(time.RFC3339),
	})
	statusCode := result.HTTPStatus
	if statusCode < 400 {
		statusCode = 200
	}
	return statusCode, resp, headers
}

func (a *Adapter) BuildWebhookPayload(result *engine.Result, chargeID string, amount int64, currency string, requestBody map[string]any) map[string]any {
	event := "payment.completed"
	if result.HTTPStatus >= 400 {
		event = "payment.failed"
	}
	if result.IsPending {
		event = "payment.pending"
	}

	metadata := map[string]any{}
	if requestBody != nil {
		if md, ok := requestBody["metadata"].(map[string]any); ok {
			metadata = md
		}
	}

	out := map[string]any{
		"eventType":    event,
		"orderId":      chargeID,
		"transactionId": fmt.Sprintf("TXN%d", time.Now().UnixNano()),
		"amount":       amount,
		"currency":     currency,
		"gatewayCode":  mastercardGatewayCode(result.Mode),
		"metadata":     metadata,
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
	}
	if requestBody != nil {
		out["request_echo"] = requestBody
	}
	return out
}

func mastercardGatewayCode(mode engine.FailureMode) string {
	m := map[engine.FailureMode]string{
		engine.ModeSuccess:          "APPROVED",
		engine.ModeBankDeclineHard:  "DECLINED",
		engine.ModeBankDeclineSoft:  "DECLINED_DO_NOT_CONTACT",
		engine.ModeBankInvalidCVV:   "INVALID_CSC",
		engine.ModeBankDoNotHonour:  "DECLINED",
		engine.ModeBankTimeout:      "TIMED_OUT",
		engine.ModePGServerError:    "SYSTEM_ERROR",
		engine.ModePGRateLimited:    "BLOCKED",
	}
	if v, ok := m[mode]; ok {
		return v
	}
	return "DECLINED"
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
	// Mastercard nests: { "order": { "amount": N, "currency": "USD" } }
	if o, ok := m["order"].(map[string]any); ok {
		if v, ok := o["amount"].(float64); ok {
			amount = int64(v)
		}
		if c, ok := o["currency"].(string); ok && c != "" {
			currency = c
		}
	}
	return amount, currency
}
