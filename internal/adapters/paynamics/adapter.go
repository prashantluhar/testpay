// Package paynamics mocks Paynamics's (PH / SEA) payment API surface.
// Real-world shape: MD5-signed JSON envelopes with a response_code /
// response_advise / response_message triple. "GR001" is the canonical
// success code; "GR051" and the "GR0xx" family cover failures. The mock
// emits a fixed signature placeholder — callers verifying MD5 must detect
// the prefix and skip verification.
package paynamics

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/prashantluhar/testpay/internal/engine"
)

// mockSignature is the placeholder we emit in every signed field. Real
// Paynamics signs with MD5(mkey + concatenated-fields); our mock skips the
// cryptography and returns this marker so consumers can short-circuit
// verification when running in the mock harness.
const mockSignature = "mock-md5-signature"

// merchantID is the deterministic merchant id the mock reports. Tests assert
// on this, so keep it stable.
const merchantID = "MOCK_MERCHANT"

// successCode / successMessage — Paynamics's happy-path triple. Consumers
// MUST check response_code against "GR001" (not response_message) before
// trusting amount / redirect_url.
const (
	successCode    = "GR001"
	successMessage = "Transaction successful"
	successAdvise  = "SUCCESS"
)

type Adapter struct{}

func New() *Adapter             { return &Adapter{} }
func (a *Adapter) Name() string { return "paynamics" }

func (a *Adapter) BuildResponse(result *engine.Result, body []byte) (int, []byte, map[string]string) {
	headers := map[string]string{"Content-Type": "application/json"}
	respID := fmt.Sprintf("PNMC_%d", time.Now().UnixNano())
	reqID := fmt.Sprintf("REQ_%d", time.Now().UnixNano())
	now := time.Now().UTC().Format(time.RFC3339)

	if result.HTTPStatus >= 400 {
		errResp := errorResponse{
			ResponseCode:    paynamicsErrorCode(result.Mode),
			ResponseAdvise:  paynamicsAdvise(result.Mode),
			ResponseMessage: paynamicsMessage(result.Mode),
			Signature:       mockSignature,
			ResponseID:      respID,
			MerchantID:      merchantID,
			RequestID:       reqID,
			Timestamp:       now,
		}
		out, _ := json.Marshal(errResp)
		return result.HTTPStatus, out, headers
	}

	amount, currency := extractAmountCurrency(body, 5000, "PHP")
	resp := initRespPayload{
		ResponseCode:    successCode,
		ResponseAdvise:  successAdvise,
		ResponseMessage: successMessage,
		Signature:       mockSignature,
		ResponseID:      respID,
		MerchantID:      merchantID,
		RequestID:       reqID,
		RedirectURL:     fmt.Sprintf("https://mock.paynamics.net/pay/%s", respID),
		Timestamp:       now,
		Currency:        currency,
		TotalAmount:     fmt.Sprintf("%.2f", float64(amount)/100),
	}
	// Paynamics returns a soft "GR002 / Pending" on async-pending init flows
	// (e.g. bank-initiated OTP pending). Preserve HTTP 200 for those.
	if result.IsPending {
		resp.ResponseCode = "GR002"
		resp.ResponseAdvise = "PENDING"
		resp.ResponseMessage = "Transaction pending"
	}

	out, _ := json.Marshal(resp)
	return 200, out, headers
}

func (a *Adapter) BuildWebhookPayload(result *engine.Result, chargeID string, amount int64, currency string, requestBody map[string]any) map[string]any {
	now := time.Now().UTC().Format(time.RFC3339)

	code, advise, msg := successCode, successAdvise, successMessage
	if result.HTTPStatus >= 400 {
		code = paynamicsErrorCode(result.Mode)
		advise = paynamicsAdvise(result.Mode)
		msg = paynamicsMessage(result.Mode)
	} else if result.IsPending {
		code, advise, msg = "GR002", "PENDING", "Transaction pending"
	}

	notif := webhookPayload{
		ResponseCode:    code,
		ResponseAdvise:  advise,
		ResponseMessage: msg,
		Signature:       mockSignature,
		ResponseID:      chargeID,
		MerchantID:      merchantID,
		RequestID:       fmt.Sprintf("REQ_%d", time.Now().UnixNano()),
		Timestamp:       now,
		TotalAmount:     fmt.Sprintf("%.2f", float64(amount)/100),
		Currency:        currency,
	}

	// Round-trip through JSON so the caller sees map[string]any — matches
	// dispatcher expectations while we keep the strongly typed DTO at the
	// adapter boundary.
	raw, _ := json.Marshal(notif)
	var out map[string]any
	_ = json.Unmarshal(raw, &out)

	if requestBody != nil {
		out["request_echo"] = requestBody
	}
	return out
}

// paynamicsErrorCode maps engine failure modes onto the Paynamics
// response_code vocabulary. GR001 is success; GR0xx (2-digit trailing) are
// the documented failure codes. The mapping is deliberately lossy — only
// codes that drive consumer behaviour matter here.
func paynamicsErrorCode(mode engine.FailureMode) string {
	switch mode {
	case engine.ModeBankDeclineHard, engine.ModeBankDeclineSoft,
		engine.ModeBankDoNotHonour:
		return "GR051"
	case engine.ModeBankInvalidCVV:
		return "GR052"
	case engine.ModeBankTimeout, engine.ModeBankServerDown:
		return "GR053"
	case engine.ModePGTimeout:
		return "GR054"
	case engine.ModePGServerError, engine.ModePGMaintenance:
		return "GR099"
	case engine.ModePGRateLimited:
		return "GR055"
	case engine.ModeNetworkError:
		return "GR098"
	default:
		return "GR051"
	}
}

// paynamicsAdvise is the short machine-friendly advise label. Consumers
// typically switch on response_code; response_advise is an additional hint.
func paynamicsAdvise(mode engine.FailureMode) string {
	switch mode {
	case engine.ModeBankDeclineHard, engine.ModeBankDeclineSoft,
		engine.ModeBankDoNotHonour, engine.ModeBankInvalidCVV:
		return "DECLINED"
	case engine.ModeBankTimeout, engine.ModeBankServerDown,
		engine.ModePGTimeout, engine.ModeNetworkError:
		return "TIMEOUT"
	case engine.ModePGServerError, engine.ModePGMaintenance:
		return "SYSTEM_ERROR"
	case engine.ModePGRateLimited:
		return "RATE_LIMITED"
	default:
		return "FAILED"
	}
}

// paynamicsMessage is the human-readable counterpart. Merchants log it;
// machines parse response_code.
func paynamicsMessage(mode engine.FailureMode) string {
	m := map[engine.FailureMode]string{
		engine.ModeBankDeclineHard: "Transaction declined by issuer",
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
// the charge request body. Paynamics nests these under `transaction` in
// production (PaynamicsInitReqPayload.Transaction.Amount) but accepts flat
// top-level keys in many merchant integrations — we try both and fall back
// to defaults on any parse failure.
func extractAmountCurrency(body []byte, defAmount int64, defCurrency string) (int64, string) {
	amount, currency := defAmount, defCurrency
	if len(body) == 0 {
		return amount, currency
	}
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		return amount, currency
	}
	// Try top-level first.
	if v, ok := m["amount"].(float64); ok {
		amount = int64(v)
	}
	if v, ok := m["currency"].(string); ok && v != "" {
		currency = v
	}
	// Then the nested Paynamics shape.
	if tx, ok := m["transaction"].(map[string]any); ok {
		if v, ok := tx["amount"].(float64); ok {
			amount = int64(v)
		}
		if v, ok := tx["amount"].(string); ok {
			if n, err := parseAmountString(v); err == nil {
				amount = n
			}
		}
		if v, ok := tx["currency"].(string); ok && v != "" {
			currency = v
		}
	}
	return amount, currency
}

// parseAmountString accepts a decimal string ("123.45") and returns minor
// units (12345). Paynamics emits amount as a string in request bodies.
func parseAmountString(s string) (int64, error) {
	var major int64
	var minor int64
	var minorDigits int
	seenDot := false
	for _, r := range s {
		switch {
		case r == '.':
			seenDot = true
		case r >= '0' && r <= '9':
			if seenDot {
				if minorDigits < 2 {
					minor = minor*10 + int64(r-'0')
					minorDigits++
				}
			} else {
				major = major*10 + int64(r-'0')
			}
		default:
			return 0, fmt.Errorf("invalid amount char %q", r)
		}
	}
	for minorDigits < 2 {
		minor *= 10
		minorDigits++
	}
	return major*100 + minor, nil
}
