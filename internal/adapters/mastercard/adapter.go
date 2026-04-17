// Package mastercard mocks the Mastercard Payment Gateway Services (MPGS)
// surface area. Real-world shape: result (SUCCESS | FAILURE | PENDING) +
// response.gatewayCode (APPROVED | DECLINED | INVALID_CSC | TIMED_OUT
// | BLOCKED | EXPIRED_CARD | ...) + order.id + transaction.id.
package mastercard

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/prashantluhar/testpay/internal/engine"
)

const (
	merchantID    = "TESTMERCHANT"
	mpgsVersion   = "79"
	entryPointAPI = "https://test-gateway.mastercard.com/api/rest/version/79"
)

type Adapter struct{}

func New() *Adapter             { return &Adapter{} }
func (a *Adapter) Name() string { return "mastercard" }

func (a *Adapter) BuildResponse(result *engine.Result, body []byte) (int, []byte, map[string]string) {
	headers := map[string]string{"Content-Type": "application/json"}

	orderID := fmt.Sprintf("ORD%d", time.Now().UnixNano())
	txnID := fmt.Sprintf("TXN%d", time.Now().UnixNano())
	now := time.Now().UTC().Format(time.RFC3339)

	amount, currency := extractAmountCurrency(body, 5000, "USD")
	amountF := float64(amount)

	resultField, gwCode, gwRec, orderStatus := classify(result)

	resp := paymentResponse{
		Result:           resultField,
		Merchant:         merchantID,
		Version:          mpgsVersion,
		TimeOfRecord:     now,
		TimeOfLastUpdate: now,
		GatewayEntryPoint: entryPointAPI,
		Order: orderResponse{
			ID:                   orderID,
			Amount:               amountF,
			Currency:             currency,
			Status:               orderStatus,
			Reference:            orderID,
			CreationTime:         now,
			LastUpdatedTime:      now,
			AuthenticationStatus: "AUTHENTICATION_NOT_IN_SCOPE",
		},
		Response: responseBlock{
			GatewayCode:           gwCode,
			GatewayRecommendation: gwRec,
			AcquirerCode:          acquirerCodeFor(gwCode),
			AcquirerMessage:       acquirerMessageFor(result.Mode),
			CardSecurityCode: &cardSecurityCode{
				GatewayCode:  cscGatewayCodeFor(result.Mode),
				AcquirerCode: "M",
			},
		},
		Transaction: transactionResponse{
			ID:                   txnID,
			Type:                 "PAYMENT",
			Amount:               amountF,
			Currency:             currency,
			AuthorizationCode:    "831000",
			AuthenticationStatus: "AUTHENTICATION_NOT_IN_SCOPE",
			Source:               "INTERNET",
			Acquirer: acquirerBlock{
				ID:            "TESTACQ",
				MerchantID:    merchantID,
				TransactionID: fmt.Sprintf("ACQ%d", time.Now().UnixNano()),
				Date:          time.Now().UTC().Format("0102"),
				TimeZone:      "+0000",
			},
		},
		AuthorizationResponse: authorizationResponse{
			Stan:                  "123456",
			ResponseCode:          responseCodeFor(result.Mode),
			ProcessingCode:        "000000",
			TransactionIdentifier: txnID,
			FinancialNetworkCode:  "MCW",
		},
	}

	// On success the totalAuthorizedAmount reflects the authorised total.
	if resultField == resultSuccess {
		resp.Order.TotalAuthorizedAmount = amountF
		resp.Order.TotalCapturedAmount = amountF
	}

	// Populate the error block for failures / errors — gives merchants the
	// human explanation and a machine-readable validationType.
	if resultField == resultFailure || resultField == resultError {
		resp.Error = &errorBlock{
			Cause:          mastercardCause(result.Mode),
			Explanation:    mastercardMessage(result.Mode),
			Field:          mastercardErrorField(result.Mode),
			ValidationType: mastercardValidationType(result.Mode),
		}
	}

	out, _ := json.Marshal(resp)
	statusCode := result.HTTPStatus
	if statusCode < 400 {
		statusCode = 200
	}
	return statusCode, out, headers
}

func (a *Adapter) BuildWebhookPayload(result *engine.Result, chargeID string, amount int64, currency string, requestBody map[string]any) map[string]any {
	now := time.Now().UTC().Format(time.RFC3339)
	amountF := float64(amount)

	resultField, gwCode, gwRec, orderStatus := classify(result)
	notifType := notificationTypeFor(result)

	notif := webhookNotification{
		NotificationID:     fmt.Sprintf("NOTIF%d", time.Now().UnixNano()),
		NotificationType:   notifType,
		TimeOfNotification: now,
		Result:             resultField,
		Merchant:           merchantID,
		Version:            mpgsVersion,
		Order: orderResponse{
			ID:                   chargeID,
			Amount:               amountF,
			Currency:             currency,
			Status:               orderStatus,
			Reference:            chargeID,
			CreationTime:         now,
			LastUpdatedTime:      now,
			AuthenticationStatus: "AUTHENTICATION_NOT_IN_SCOPE",
		},
		Response: responseBlock{
			GatewayCode:           gwCode,
			GatewayRecommendation: gwRec,
			AcquirerCode:          acquirerCodeFor(gwCode),
			AcquirerMessage:       acquirerMessageFor(result.Mode),
		},
		Transaction: transactionResponse{
			ID:                fmt.Sprintf("TXN%d", time.Now().UnixNano()),
			Type:              transactionTypeFor(result.Mode),
			Amount:            amountF,
			Currency:          currency,
			AuthorizationCode: "831000",
			Source:            "INTERNET",
			Acquirer: acquirerBlock{
				ID:         "TESTACQ",
				MerchantID: merchantID,
			},
		},
	}

	if resultField == resultSuccess {
		notif.Order.TotalAuthorizedAmount = amountF
		notif.Order.TotalCapturedAmount = amountF
	}

	raw, _ := json.Marshal(notif)
	var out map[string]any
	_ = json.Unmarshal(raw, &out)

	// MPGS's notificationUrl contract echoes the merchant's order.reference
	// back to them so webhook consumers can correlate. Carry through the
	// full request body under request_echo — consistent with the other
	// adapters in this repo (Adyen, Stripe, Agnostic).
	if requestBody != nil {
		out["request_echo"] = requestBody
	}
	return out
}

// classify collapses an engine.Result into the four MPGS-shaped knobs the
// response/webhook builders need: top-level `result`, `response.gatewayCode`,
// `response.gatewayRecommendation`, and `order.status`.
func classify(result *engine.Result) (resultField, gwCode, gwRec, orderStatus string) {
	switch {
	case result.HTTPStatus >= 500:
		// Server-side failures come back as result=ERROR in MPGS so
		// merchants can distinguish "something broke at the gateway"
		// from "the acquirer said no".
		return resultError, mastercardGatewayCode(result.Mode), "RESUBMIT", "FAILED"
	case result.HTTPStatus >= 400:
		return resultFailure, mastercardGatewayCode(result.Mode), "DO_NOT_PROCEED", "FAILED"
	case result.IsPending:
		return resultPending, "PENDING", "PROCEED", "PENDING"
	default:
		return resultSuccess, "APPROVED", "PROCEED", "CAPTURED"
	}
}

// mastercardGatewayCode maps the engine's failure taxonomy onto MPGS's
// response.gatewayCode vocabulary. The five canonical categories the spec
// asked for — APPROVED, DECLINED, PENDING, EXPIRED_CARD, BLOCKED — are
// all represented, alongside the more specific codes MPGS uses in
// production (INVALID_CSC, TIMED_OUT, SYSTEM_ERROR, ...).
func mastercardGatewayCode(mode engine.FailureMode) string {
	m := map[engine.FailureMode]string{
		engine.ModeSuccess:          "APPROVED",
		engine.ModeBankDeclineHard:  "DECLINED",
		engine.ModeBankDeclineSoft:  "DECLINED_DO_NOT_CONTACT",
		engine.ModeBankInvalidCVV:   "INVALID_CSC",
		engine.ModeBankDoNotHonour:  "DECLINED",
		engine.ModeBankTimeout:      "TIMED_OUT",
		engine.ModeBankServerDown:   "ACQUIRER_SYSTEM_ERROR",
		engine.ModePGServerError:    "SYSTEM_ERROR",
		engine.ModePGTimeout:        "TIMED_OUT",
		engine.ModePGRateLimited:    "BLOCKED",
		engine.ModePGMaintenance:    "SYSTEM_ERROR",
		engine.ModeNetworkError:     "UNSPECIFIED_FAILURE",
		engine.ModeRedirectSuccess:  "APPROVED",
		engine.ModeRedirectAbandoned: "CANCELLED",
		engine.ModeRedirectTimeout:  "EXPIRED_CARD",
		engine.ModeRedirectFailed:   "DECLINED",
		engine.ModePendingThenSuccess: "PENDING",
		engine.ModePendingThenFailed:  "PENDING",
	}
	if v, ok := m[mode]; ok {
		return v
	}
	return "DECLINED"
}

// mastercardMessage maps engine modes to the acquirerMessage / explanation
// the merchant sees — human strings pulled from the MPGS integration guide.
func mastercardMessage(mode engine.FailureMode) string {
	m := map[engine.FailureMode]string{
		engine.ModeBankDeclineHard:  "Declined",
		engine.ModeBankDeclineSoft:  "Insufficient Funds",
		engine.ModeBankInvalidCVV:   "Invalid Card Security Code",
		engine.ModeBankDoNotHonour:  "Do Not Honour",
		engine.ModeBankTimeout:      "Issuer Timeout",
		engine.ModeBankServerDown:   "Acquirer System Error",
		engine.ModePGTimeout:        "Gateway Timeout",
		engine.ModePGServerError:    "System Error",
		engine.ModePGRateLimited:    "Blocked by Gateway",
		engine.ModePGMaintenance:    "Service Unavailable",
		engine.ModeNetworkError:     "Network Error",
		engine.ModeRedirectAbandoned: "Cancelled by Payer",
		engine.ModeRedirectTimeout:  "Authentication Expired",
		engine.ModeRedirectFailed:   "Authentication Failed",
	}
	if v, ok := m[mode]; ok {
		return v
	}
	return "Transaction Declined"
}

// mastercardCause maps engine modes to MPGS error.cause values — the
// machine-readable side of the error block.
func mastercardCause(mode engine.FailureMode) string {
	m := map[engine.FailureMode]string{
		engine.ModeBankDeclineHard:  "DECLINED",
		engine.ModeBankDeclineSoft:  "INSUFFICIENT_FUNDS",
		engine.ModeBankInvalidCVV:   "INVALID_REQUEST",
		engine.ModeBankDoNotHonour:  "DECLINED",
		engine.ModeBankTimeout:      "REQUEST_TIMED_OUT",
		engine.ModeBankServerDown:   "SERVER_BUSY",
		engine.ModePGTimeout:        "REQUEST_TIMED_OUT",
		engine.ModePGServerError:    "SERVER_FAILED",
		engine.ModePGRateLimited:    "REQUEST_REJECTED",
		engine.ModePGMaintenance:    "SERVER_BUSY",
		engine.ModeNetworkError:     "REQUEST_TIMED_OUT",
	}
	if v, ok := m[mode]; ok {
		return v
	}
	return "INVALID_REQUEST"
}

// mastercardErrorField — MPGS reports which request field caused the
// validation failure. For CVV-style failures it points at securityCode;
// for amount issues, order.amount; else empty.
func mastercardErrorField(mode engine.FailureMode) string {
	switch mode {
	case engine.ModeBankInvalidCVV:
		return "sourceOfFunds.provided.card.securityCode"
	case engine.ModeAmountMismatch:
		return "order.amount"
	}
	return ""
}

// mastercardValidationType — MPGS categorises errors as INVALID, MISSING,
// UNSUPPORTED, or (for business failures) omitted. The field is only
// populated for actual validation errors.
func mastercardValidationType(mode engine.FailureMode) string {
	switch mode {
	case engine.ModeBankInvalidCVV, engine.ModeAmountMismatch:
		return "INVALID"
	}
	return ""
}

// acquirerCodeFor — ISO 8583 style numeric response code. "00" for
// approvals, "05" for do-not-honour, etc. MPGS pipes the acquirer's code
// through unchanged.
func acquirerCodeFor(gatewayCode string) string {
	switch gatewayCode {
	case "APPROVED":
		return "00"
	case "DECLINED", "DECLINED_DO_NOT_CONTACT":
		return "05"
	case "INVALID_CSC":
		return "N7"
	case "TIMED_OUT":
		return "91"
	case "EXPIRED_CARD":
		return "54"
	case "BLOCKED":
		return "62"
	case "SYSTEM_ERROR", "ACQUIRER_SYSTEM_ERROR":
		return "96"
	case "CANCELLED":
		return "17"
	}
	return "05"
}

// acquirerMessageFor — acquirer-sourced human message (distinct from
// response.gatewayCode which is MPGS's normalised vocabulary).
func acquirerMessageFor(mode engine.FailureMode) string {
	if mode == engine.ModeSuccess {
		return "Approved"
	}
	return mastercardMessage(mode)
}

// cscGatewayCodeFor — MPGS returns a nested cardSecurityCode.gatewayCode
// on every card auth (MATCH, NO_MATCH, NOT_PROCESSED, ...).
func cscGatewayCodeFor(mode engine.FailureMode) string {
	switch mode {
	case engine.ModeBankInvalidCVV:
		return "NO_MATCH"
	case engine.ModeSuccess:
		return "MATCH"
	}
	return "NOT_PROCESSED"
}

// responseCodeFor — ISO 8583 authorisation response code. Populated into
// authorizationResponse.responseCode so merchants can parse it
// independently from the normalised gatewayCode.
func responseCodeFor(mode engine.FailureMode) string {
	switch mode {
	case engine.ModeSuccess:
		return "00"
	case engine.ModeBankDeclineSoft:
		return "51"
	case engine.ModeBankInvalidCVV:
		return "82"
	case engine.ModeBankDoNotHonour:
		return "05"
	case engine.ModeBankTimeout:
		return "91"
	}
	return "05"
}

// notificationTypeFor — MPGS webhooks carry a notificationType that drives
// the consumer's state machine. ORDER is the generic order-state change,
// TRANSACTION fires per transaction event (capture, void, refund).
func notificationTypeFor(result *engine.Result) string {
	switch result.Mode {
	case engine.ModeSuccessThenReversed:
		return "TRANSACTION"
	}
	if result.HTTPStatus >= 400 {
		return "TRANSACTION"
	}
	return "ORDER"
}

// transactionTypeFor — MPGS sends PAYMENT for the initial auth, REFUND
// for reversals/refunds, and VOID_AUTHORIZATION for cancellations.
func transactionTypeFor(mode engine.FailureMode) string {
	switch mode {
	case engine.ModeSuccessThenReversed:
		return "REFUND"
	}
	return "PAYMENT"
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
	// MPGS nests: { "order": { "amount": N, "currency": "USD" } }.
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
