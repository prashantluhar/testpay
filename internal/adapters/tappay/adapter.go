// Package tappay mocks TapPay's payment API surface.
// The underlying wire format matches the Appotapay DTOs from the reference
// library (errorCode float + message + signature envelope). Real Appotapay
// signs payloads with HMAC-SHA256 over concatenated request fields; the
// mock emits a fixed signature placeholder.
package tappay

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/prashantluhar/testpay/internal/engine"
)

// Package-wide constants. partnerCode / apiKey are the deterministic
// credentials the mock reports in webhook payloads — tests assert on these,
// so keep them stable.
const (
	mockSignature = "mock-hmac-signature"
	partnerCode   = "MOCK_PARTNER"
	mockAPIKey    = "MOCK_API_KEY"
)

// errorCodeSuccess — Appotapay's "zero is success" protocol. Any non-zero
// value is a failure; callers MUST switch on this (not message) before
// trusting paymentUrl / amount.
const errorCodeSuccess = 0

type Adapter struct{}

func New() *Adapter             { return &Adapter{} }
func (a *Adapter) Name() string { return "tappay" }

func (a *Adapter) BuildResponse(result *engine.Result, body []byte) (int, []byte, map[string]string) {
	headers := map[string]string{"Content-Type": "application/json"}
	orderID := fmt.Sprintf("TAP_%d", time.Now().UnixNano())

	if result.HTTPStatus >= 400 {
		errResp := errorResponse{
			ErrorCode: tappayErrorCode(result.Mode),
			Message:   tappayMessage(result.Mode),
			Signature: mockSignature,
			OrderID:   orderID,
		}
		out, _ := json.Marshal(errResp)
		return result.HTTPStatus, out, headers
	}

	amount, _ := extractAmountCurrency(body, 5000, "TWD")
	resp := paymentResponse{
		ErrorCode:  errorCodeSuccess,
		Message:    "Success",
		OrderID:    orderID,
		Amount:     float64(amount),
		PaymentURL: fmt.Sprintf("https://mock.tappay.io/pay/%s", orderID),
		Signature:  mockSignature,
	}
	// Appotapay returns errorCode 1 ("processing") on async-pending init
	// while still emitting HTTP 200 — preserve that.
	if result.IsPending {
		resp.ErrorCode = 1
		resp.Message = "Processing"
	}

	out, _ := json.Marshal(resp)
	return 200, out, headers
}

func (a *Adapter) BuildWebhookPayload(result *engine.Result, chargeID string, amount int64, currency string, requestBody map[string]any) map[string]any {
	ec := errorCodeSuccess
	msg := "Success"
	switch {
	case result.HTTPStatus >= 400:
		ec = tappayErrorCode(result.Mode)
		msg = tappayMessage(result.Mode)
	case result.IsPending:
		ec = 1
		msg = "Processing"
	}

	notif := webhookOrderStatusResponse{
		ErrorCode:        ec,
		Message:          msg,
		PartnerCode:      partnerCode,
		APIKey:           mockAPIKey,
		Amount:           int(amount),
		Currency:         currency,
		OrderID:          chargeID,
		BankCode:         "MOCK_BANK",
		PaymentMethod:    "ATM",
		PaymentType:      "BANK_TRANSFER",
		AppotapayTransID: fmt.Sprintf("APT_%d", time.Now().UnixNano()),
		TransactionTs:    time.Now().Unix(),
		Signature:        mockSignature,
	}

	// Round-trip through JSON so the caller receives a map[string]any — the
	// dispatcher expects that shape, while we keep the strongly typed DTO at
	// the adapter boundary.
	raw, _ := json.Marshal(notif)
	var out map[string]any
	_ = json.Unmarshal(raw, &out)

	if requestBody != nil {
		out["request_echo"] = requestBody
	}
	return out
}

// tappayErrorCode maps engine failure modes onto Appotapay's integer
// errorCode vocabulary. 0 is success; 1 is "processing"; everything else
// is a failure. The mapping is deliberately lossy — only codes that drive
// consumer behaviour matter here.
func tappayErrorCode(mode engine.FailureMode) int {
	switch mode {
	case engine.ModeBankDeclineHard, engine.ModeBankDeclineSoft,
		engine.ModeBankDoNotHonour:
		return 2
	case engine.ModeBankInvalidCVV:
		return 3
	case engine.ModeBankTimeout, engine.ModeBankServerDown:
		return 4
	case engine.ModePGTimeout:
		return 5
	case engine.ModePGServerError, engine.ModePGMaintenance:
		return 99
	case engine.ModePGRateLimited:
		return 6
	case engine.ModeNetworkError:
		return 98
	default:
		return 2
	}
}

// tappayMessage is the human-readable counterpart to errorCode. Merchants
// log it; machines parse errorCode.
func tappayMessage(mode engine.FailureMode) string {
	m := map[engine.FailureMode]string{
		engine.ModeBankDeclineHard: "Transaction declined",
		engine.ModeBankDeclineSoft: "Insufficient funds",
		engine.ModeBankInvalidCVV:  "Invalid CVV",
		engine.ModeBankDoNotHonour: "Do not honour",
		engine.ModeBankTimeout:     "Bank timeout",
		engine.ModeBankServerDown:  "Bank unavailable",
		engine.ModePGTimeout:       "Gateway timeout",
		engine.ModePGServerError:   "Internal server error",
		engine.ModePGRateLimited:   "Too many requests",
		engine.ModePGMaintenance:   "Gateway under maintenance",
		engine.ModeNetworkError:    "Network error",
	}
	if v, ok := m[mode]; ok {
		return v
	}
	return "Transaction failed"
}

// extractAmountCurrency pulls the customer-supplied amount+currency out of
// the charge request body. Appotapay accepts top-level `amount` (int minor
// units) and an optional `currency`; we fall back to defaults on any parse
// failure so the mock always emits a well-formed response.
func extractAmountCurrency(body []byte, defAmount int64, defCurrency string) (int64, string) {
	amount, currency := defAmount, defCurrency
	if len(body) == 0 {
		return amount, currency
	}
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		return amount, currency
	}
	if v, ok := m["amount"].(float64); ok {
		amount = int64(v)
	}
	if v, ok := m["currency"].(string); ok && v != "" {
		currency = v
	}
	return amount, currency
}
