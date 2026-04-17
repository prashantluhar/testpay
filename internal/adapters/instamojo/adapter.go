// Package instamojo mocks Instamojo's (India) payment-request API.
// Real-world shape: a `{ success: bool, payment_request: {...} }` envelope on
// success and `{ success: false, message: "...", errors: {...} }` on
// validation failures. Amounts are string-encoded throughout.
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
	headers := map[string]string{"Content-Type": "application/json"}
	id := fmt.Sprintf("MOJO_%d", time.Now().UnixNano())

	if result.HTTPStatus >= 400 {
		errResp := errorResponse{
			Success: false,
			Message: instamojoMessage(result.Mode),
			Errors:  instamojoErrors(result.Mode),
		}
		out, _ := json.Marshal(errResp)
		return result.HTTPStatus, out, headers
	}

	amount, _ := extractAmountCurrency(body, 5000, "INR")
	now := time.Now().UTC()
	resp := initResponse{
		Success: true,
		PaymentRequest: paymentRequest{
			ID:          id,
			Phone:       "9999999999",
			Email:       "mock@example.com",
			BuyerName:   "Mock Buyer",
			Amount:      fmt.Sprintf("%d", amount),
			Purpose:     "Mock charge",
			Status:      instamojoStatus(result),
			SendSMS:     false,
			SendEmail:   false,
			Longurl:     fmt.Sprintf("https://instamojo.com/@mock/%s", id),
			CreatedAt:   now,
			ModifiedAt:  now,
			AllowRepeatedPayments: false,
		},
	}
	out, _ := json.Marshal(resp)
	return 200, out, headers
}

func (a *Adapter) BuildWebhookPayload(result *engine.Result, chargeID string, amount int64, currency string, requestBody map[string]any) map[string]any {
	notif := webhookPayload{
		PaymentID:        chargeID,
		PaymentRequestID: fmt.Sprintf("MOJOPR_%d", time.Now().UnixNano()),
		Status:           instamojoWebhookStatus(result),
		Amount:           fmt.Sprintf("%d", amount),
		Currency:         currency,
		Buyer:            "mock@example.com",
		BuyerName:        "Mock Buyer",
		BuyerPhone:       "9999999999",
		Fees:             "0.00",
		InstrumentType:   "Card",
		CreatedAt:        time.Now().UTC(),
	}
	if result.HTTPStatus >= 400 {
		notif.FailureReason = string(result.Mode)
		notif.FailureMessage = instamojoMessage(result.Mode)
	}

	raw, _ := json.Marshal(notif)
	var out map[string]any
	_ = json.Unmarshal(raw, &out)

	if requestBody != nil {
		// Echo purpose as a top-level field — it's how Instamojo merchants
		// commonly correlate webhooks to orders.
		if p, ok := requestBody["purpose"].(string); ok {
			out["purpose"] = p
		}
		out["request_echo"] = requestBody
	}
	return out
}

// instamojoStatus maps engine modes onto the Instamojo init-response status
// vocabulary. Instamojo uses "Pending" (awaiting payment), "Sent" (link
// delivered), "Completed" (paid).
func instamojoStatus(r *engine.Result) string {
	if r.IsPending {
		return "Pending"
	}
	return "Sent"
}

// instamojoWebhookStatus maps engine modes onto the webhook-side status set.
// Instamojo uses "Credit" for a successful capture and "Failed" for a decline.
func instamojoWebhookStatus(r *engine.Result) string {
	if r.HTTPStatus >= 400 {
		return "Failed"
	}
	if r.IsPending {
		return "Pending"
	}
	return "Credit"
}

func instamojoMessage(mode engine.FailureMode) string {
	m := map[engine.FailureMode]string{
		engine.ModeBankDeclineHard: "Payment declined by bank",
		engine.ModeBankDeclineSoft: "Insufficient funds",
		engine.ModeBankInvalidCVV:  "Invalid CVV",
		engine.ModeBankDoNotHonour: "Do not honour",
		engine.ModeBankTimeout:     "Bank timeout",
		engine.ModeBankServerDown:  "Bank unavailable",
		engine.ModePGTimeout:       "Gateway timeout",
		engine.ModePGServerError:   "Internal server error",
		engine.ModePGRateLimited:   "Too many requests",
		engine.ModePGMaintenance:   "Service unavailable",
		engine.ModeNetworkError:    "Network error",
	}
	if v, ok := m[mode]; ok {
		return v
	}
	return "Payment request creation failed"
}

// instamojoErrors shapes Instamojo's field-level errors map — keyed on the
// field that failed validation, each value a slice of string messages.
func instamojoErrors(mode engine.FailureMode) map[string][]string {
	return map[string][]string{
		"code": {string(mode)},
	}
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
	// Instamojo accepts numeric or string amounts.
	switch v := m["amount"].(type) {
	case float64:
		amount = int64(v)
	case string:
		var n int64
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil && n > 0 {
			amount = n
		}
	}
	if v, ok := m["currency"].(string); ok && v != "" {
		currency = v
	}
	return amount, currency
}
