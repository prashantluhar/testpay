// Package adyen mocks Adyen's payment API surface.
// Real-world shape: resultCode + pspReference + additionalData map.
package adyen

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/prashantluhar/testpay/internal/engine"
)

type Adapter struct{}

func New() *Adapter             { return &Adapter{} }
func (a *Adapter) Name() string { return "adyen" }

func (a *Adapter) BuildResponse(result *engine.Result, body []byte) (int, []byte, map[string]string) {
	headers := map[string]string{"Content-Type": "application/json"}
	psp := fmt.Sprintf("PSP%d", time.Now().UnixNano())

	if result.HTTPStatus >= 400 {
		errBody, _ := json.Marshal(map[string]any{
			"status":       result.HTTPStatus,
			"errorCode":    result.ErrorCode,
			"message":      adyenMessage(result.Mode),
			"errorType":    "validation",
			"pspReference": psp,
		})
		return result.HTTPStatus, errBody, headers
	}

	amount, currency := extractAmountCurrency(body, 5000, "USD")
	resp, _ := json.Marshal(map[string]any{
		"pspReference": psp,
		"resultCode":   adyenResultCode(result.Mode),
		"amount":       map[string]any{"value": amount, "currency": currency},
		"merchantReference": fmt.Sprintf("ref_%d", time.Now().UnixNano()),
	})
	return 200, resp, headers
}

func (a *Adapter) BuildWebhookPayload(result *engine.Result, chargeID string, amount int64, currency string, requestBody map[string]any) map[string]any {
	eventCode := "AUTHORISATION"
	success := result.HTTPStatus < 400

	addl := map[string]any{}
	if requestBody != nil {
		if md, ok := requestBody["additionalData"].(map[string]any); ok {
			addl = md
		}
	}

	item := map[string]any{
		"eventCode":         eventCode,
		"pspReference":      chargeID,
		"success":           success,
		"amount":            map[string]any{"value": amount, "currency": currency},
		"merchantReference": fmt.Sprintf("ref_%d", time.Now().UnixNano()),
		"additionalData":    addl,
	}
	if requestBody != nil {
		item["request_echo"] = requestBody
	}

	return map[string]any{
		"live":              "false",
		"notificationItems": []map[string]any{{"NotificationRequestItem": item}},
	}
}

func adyenResultCode(mode engine.FailureMode) string {
	switch mode {
	case engine.ModeSuccess:
		return "Authorised"
	case engine.ModePendingThenFailed, engine.ModePendingThenSuccess, engine.ModeFailedThenSuccess:
		return "Pending"
	default:
		return "Refused"
	}
}

func adyenMessage(mode engine.FailureMode) string {
	m := map[engine.FailureMode]string{
		engine.ModeBankDeclineHard:  "Refused",
		engine.ModeBankDeclineSoft:  "Not enough balance",
		engine.ModeBankInvalidCVV:   "CVC Declined",
		engine.ModeBankTimeout:      "Issuer Unavailable",
		engine.ModePGServerError:    "Internal Error",
		engine.ModePGRateLimited:    "Too many requests",
	}
	if v, ok := m[mode]; ok {
		return v
	}
	return "Refused"
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
	// Adyen nests amount: { "amount": { "value": N, "currency": "USD" } }
	if a, ok := m["amount"].(map[string]any); ok {
		if v, ok := a["value"].(float64); ok {
			amount = int64(v)
		}
		if c, ok := a["currency"].(string); ok && c != "" {
			currency = c
		}
	}
	return amount, currency
}
