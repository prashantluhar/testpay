// Package adyen mocks Adyen's payment API surface.
// Real-world shape: resultCode + pspReference + additionalData map.
package adyen

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/prashantluhar/testpay/internal/engine"
)

const merchantAccountCode = "TestMerchantAccount"

type Adapter struct{}

func New() *Adapter             { return &Adapter{} }
func (a *Adapter) Name() string { return "adyen" }

func (a *Adapter) BuildResponse(result *engine.Result, body []byte) (int, []byte, map[string]string) {
	headers := map[string]string{"Content-Type": "application/json"}
	psp := fmt.Sprintf("PSP%d", time.Now().UnixNano())
	merchantRef := fmt.Sprintf("ref_%d", time.Now().UnixNano())

	if result.HTTPStatus >= 400 {
		errResp := errorResponse{
			Status:       result.HTTPStatus,
			ErrorCode:    adyenErrorCode(result.Mode),
			Message:      adyenMessage(result.Mode),
			ErrorType:    "validation",
			PspReference: psp,
		}
		errBody, _ := json.Marshal(errResp)
		return result.HTTPStatus, errBody, headers
	}

	amount, currency := extractAmountCurrency(body, 5000, "USD")
	resp := paymentsResponse{
		PspReference:      psp,
		ResultCode:        adyenResultCode(result.Mode),
		MerchantReference: merchantRef,
		Amount:            Amount{Value: amount, Currency: currency},
		PaymentMethod:     paymentMethodEcho{Type: "scheme", Brand: "visa"},
		AdditionalData:    map[string]string{"acquirerCode": "TestAcquirer"},
	}
	// "Refused" at 200 is a valid Adyen shape — the acquirer said no without
	// throwing an HTTP error. Populate refusalReason so the consumer can
	// surface it without needing to look at resultCode specifically.
	if resp.ResultCode == "Refused" {
		resp.RefusalReason = adyenMessage(result.Mode)
		resp.RefusalReasonCode = adyenErrorCode(result.Mode)
	}
	out, _ := json.Marshal(resp)
	return 200, out, headers
}

func (a *Adapter) BuildWebhookPayload(result *engine.Result, chargeID string, amount int64, currency string, requestBody map[string]any) map[string]any {
	// additionalData echo: Adyen caches the card.summary, authCode, etc. here.
	addl := map[string]string{"authCode": "123456"}
	if requestBody != nil {
		if md, ok := requestBody["additionalData"].(map[string]any); ok {
			for k, v := range md {
				if s, ok := v.(string); ok {
					addl[k] = s
				}
			}
		}
	}

	success := "true"
	if result.HTTPStatus >= 400 {
		success = "false"
	}

	item := notificationRequestItem{
		AdditionalData:      addl,
		Amount:              Amount{Value: amount, Currency: currency},
		EventCode:           adyenEventCode(result.Mode),
		EventDate:           time.Now().UTC(),
		MerchantAccountCode: merchantAccountCode,
		MerchantReference:   fmt.Sprintf("ref_%d", time.Now().UnixNano()),
		PaymentMethod:       "scheme",
		PspReference:        chargeID,
		Success:             success,
		Operations:          []string{"CANCEL", "CAPTURE", "REFUND"},
	}
	if success == "false" {
		item.Reason = adyenMessage(result.Mode)
	}
	if result.Mode == engine.ModeSuccessThenReversed {
		item.OriginalReference = chargeID
	}

	notif := webhookNotification{
		Live:              "false",
		NotificationItems: []notificationItem{{NotificationRequestItem: item}},
	}

	// Round-trip through JSON so the caller sees a map[string]any payload
	// consistent with what the dispatcher expects, while we keep the strongly-
	// typed DTO at the adapter boundary.
	raw, _ := json.Marshal(notif)
	var out map[string]any
	_ = json.Unmarshal(raw, &out)
	if requestBody != nil {
		out["request_echo"] = requestBody
	}
	return out
}

// adyenResultCode maps the engine's failure-mode taxonomy onto Adyen's
// resultCode vocabulary (Authorised / Refused / Pending / Received / ...).
func adyenResultCode(mode engine.FailureMode) string {
	switch mode {
	case engine.ModeSuccess, engine.ModeWebhookMissing, engine.ModeWebhookDelayed,
		engine.ModeWebhookDuplicate, engine.ModeWebhookOutOfOrder,
		engine.ModeWebhookMalformed, engine.ModeAmountMismatch,
		engine.ModePartialSuccess, engine.ModeDoubleCharge,
		engine.ModeRedirectSuccess:
		return "Authorised"
	case engine.ModePendingThenFailed, engine.ModePendingThenSuccess,
		engine.ModeFailedThenSuccess, engine.ModeSuccessThenReversed:
		return "Received"
	default:
		return "Refused"
	}
}

// adyenEventCode maps the engine mode onto Adyen's webhook eventCode set.
// AUTHORISATION is the default; REFUND fires for the reversal/refund modes;
// CHARGEBACK isn't directly modeled (leaving room for a future mode).
func adyenEventCode(mode engine.FailureMode) string {
	switch mode {
	case engine.ModeSuccessThenReversed:
		return "CANCELLATION"
	case engine.ModeDoubleCharge:
		// Adyen sends an extra AUTHORISATION on duplicate capture paths.
		return "AUTHORISATION"
	default:
		return "AUTHORISATION"
	}
}

func adyenMessage(mode engine.FailureMode) string {
	m := map[engine.FailureMode]string{
		engine.ModeBankDeclineHard:   "Refused",
		engine.ModeBankDeclineSoft:   "Not enough balance",
		engine.ModeBankInvalidCVV:    "CVC Declined",
		engine.ModeBankDoNotHonour:   "Do Not Honour",
		engine.ModeBankTimeout:       "Issuer Unavailable",
		engine.ModeBankServerDown:    "Issuer Unavailable",
		engine.ModePGTimeout:         "Connection Timeout",
		engine.ModePGServerError:     "Internal Error",
		engine.ModePGRateLimited:     "Too many requests",
		engine.ModePGMaintenance:     "Service Unavailable",
		engine.ModeNetworkError:      "Network Error",
		engine.ModeRedirectAbandoned: "Cancelled",
		engine.ModeRedirectTimeout:   "Expired",
		engine.ModeRedirectFailed:    "Refused",
	}
	if v, ok := m[mode]; ok {
		return v
	}
	return "Refused"
}

// adyenErrorCode maps engine modes to Adyen's errorCode / refusalReasonCode
// vocabulary — numeric strings that merchants parse programmatically.
// See https://docs.adyen.com/development-resources/refusal-reasons
func adyenErrorCode(mode engine.FailureMode) string {
	m := map[engine.FailureMode]string{
		engine.ModeBankDeclineHard: "2",    // "Refused"
		engine.ModeBankDeclineSoft: "5",    // "Not enough balance"
		engine.ModeBankInvalidCVV:  "7",    // "CVC Declined"
		engine.ModeBankDoNotHonour: "4",    // "Do Not Honour"
		engine.ModeBankTimeout:     "9",    // "Issuer Unavailable"
		engine.ModeBankServerDown:  "9",
		engine.ModePGTimeout:       "26",   // "Connection Timeout"
		engine.ModePGServerError:   "10",   // "Not supported"
		engine.ModePGRateLimited:   "29",   // "Too Many Requests"
		engine.ModePGMaintenance:   "905",  // "Service Unavailable"
	}
	if v, ok := m[mode]; ok {
		return v
	}
	return "2"
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
