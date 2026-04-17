// Package espay mocks ESPay (Indonesia) payment API.
// Real-world shape: RSA-2048 signed payloads with a uniform
// error_code / error_message envelope. "0000" is the success code; anything
// else is an ESPxx-style failure. We mock with a fixed signature placeholder.
package espay

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/prashantluhar/testpay/internal/engine"
)

// mockSignature is the fixed placeholder ESPay's mock emits in every signed
// field. Real ESPay signs with an RSA-2048 private key per merchant. Consumers
// validating signatures should detect this prefix and skip verification.
const mockSignature = "mock-rsa-signature"

// successCode / successMsg — ESPay's "zero-is-success" protocol. Callers
// MUST check error_code == "0000" before trusting the rest of the payload.
const (
	successCode = "0000"
	successMsg  = "Success"
)

type Adapter struct{}

func New() *Adapter             { return &Adapter{} }
func (a *Adapter) Name() string { return "espay" }

func (a *Adapter) BuildResponse(result *engine.Result, body []byte) (int, []byte, map[string]string) {
	headers := map[string]string{"Content-Type": "application/json"}
	uuid := fmt.Sprintf("rq_%d", time.Now().UnixNano())
	orderID := fmt.Sprintf("ESP_%d", time.Now().UnixNano())
	now := time.Now().UTC().Format("20060102150405")

	if result.HTTPStatus >= 400 {
		errResp := errorResponse{
			RqUuid:       uuid,
			RsDateTime:   now,
			ErrorCode:    espayErrorCode(result.Mode),
			ErrorMessage: espayMessage(result.Mode),
			Signature:    mockSignature,
			OrderID:      orderID,
		}
		out, _ := json.Marshal(errResp)
		return result.HTTPStatus, out, headers
	}

	amount, currency := extractAmountCurrency(body, 5000, "IDR")
	resp := inquireResponse{
		RqUuid:       uuid,
		RsDateTime:   now,
		ErrorCode:    successCode,
		ErrorMessage: successMsg,
		Signature:    mockSignature,
		OrderID:      orderID,
		Amount:       fmt.Sprintf("%d", amount),
		CCY:          currency,
		Description:  "Mock ESPay charge",
		TrxDate:      now,
		CustomerDetails: customerDetails{
			FirstName:   "Mock",
			LastName:    "Customer",
			PhoneNumber: "081200000000",
			Email:       "mock@example.com",
		},
	}
	// Pending-like modes degrade to a soft "Processing" code while still
	// returning a 200 — ESPay does use distinct codes for the async states.
	if result.IsPending {
		resp.ErrorCode = "0001"
		resp.ErrorMessage = "Processing"
	}
	out, _ := json.Marshal(resp)
	return 200, out, headers
}

func (a *Adapter) BuildWebhookPayload(result *engine.Result, chargeID string, amount int64, currency string, requestBody map[string]any) map[string]any {
	now := time.Now().UTC().Format("20060102150405")
	reconcileID := fmt.Sprintf("RCN_%d", time.Now().UnixNano())

	notif := webhookResponse{
		RqUuid:            fmt.Sprintf("rq_%d", time.Now().UnixNano()),
		RsDateTime:        now,
		ErrorCode:         successCode,
		ErrorMessage:      successMsg,
		Signature:         mockSignature,
		OrderID:           chargeID,
		ReconcileID:       reconcileID,
		ReconcileDateTime: now,
	}
	if result.HTTPStatus >= 400 {
		notif.ErrorCode = espayErrorCode(result.Mode)
		notif.ErrorMessage = espayMessage(result.Mode)
	}

	// Round-trip through JSON so the caller receives map[string]any — matches
	// the dispatcher's expectations while keeping the typed DTO at the boundary.
	raw, _ := json.Marshal(notif)
	var out map[string]any
	_ = json.Unmarshal(raw, &out)

	// ESPay webhooks don't carry amount/currency by default (merchants correlate
	// by order_id) — expose them as top-level fields for downstream convenience.
	out["amount"] = amount
	out["currency"] = currency

	if requestBody != nil {
		out["request_echo"] = requestBody
	}
	return out
}

// espayErrorCode maps the engine's failure-mode taxonomy onto ESPay's error
// code vocabulary. ESPay uses numeric strings; real production has dozens of
// codes but the mock sticks to a representative subset.
func espayErrorCode(mode engine.FailureMode) string {
	switch mode {
	case engine.ModeBankDeclineHard, engine.ModeBankDeclineSoft,
		engine.ModeBankInvalidCVV, engine.ModeBankDoNotHonour:
		return "ESP02"
	case engine.ModeBankTimeout, engine.ModeBankServerDown:
		return "ESP04"
	case engine.ModePGTimeout, engine.ModePGServerError,
		engine.ModePGMaintenance:
		return "ESP99"
	case engine.ModePGRateLimited:
		return "ESP05"
	case engine.ModeNetworkError:
		return "ESP98"
	default:
		return "ESP99"
	}
}

// espayMessage is a human-readable counterpart to the error code. Merchants
// log this; machines should parse error_code.
func espayMessage(mode engine.FailureMode) string {
	m := map[engine.FailureMode]string{
		engine.ModeBankDeclineHard: "Transaction declined",
		engine.ModeBankDeclineSoft: "Insufficient funds",
		engine.ModeBankInvalidCVV:  "Invalid CVV",
		engine.ModeBankDoNotHonour: "Do Not Honour",
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
// the charge request body. ESPay accepts both numeric and string amounts;
// we coerce either to int64 minor units and fall back to defaults on any
// parse failure.
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
	if v, ok := m["ccy"].(string); ok && v != "" {
		currency = v
	} else if v, ok := m["currency"].(string); ok && v != "" {
		currency = v
	}
	return amount, currency
}
