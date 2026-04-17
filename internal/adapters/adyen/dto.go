// DTO shapes for Adyen's /payments response and webhook notifications.
// Derived from the production reference library (test-dto-/adyen.go).
// Kept scoped to the fields this mock emits; extend when a failure mode
// needs more surface.
package adyen

import "time"

// Amount — minor-units value + ISO currency. Shared by request, response,
// and webhook. Adyen sends int64 minor units in webhooks; the response
// uses the same shape.
type Amount struct {
	Currency string `json:"currency"`
	Value    int64  `json:"value"`
}

// paymentsResponse mirrors Adyen's POST /payments response. Success and
// refusal paths share most fields; refusal adds refusalReason +
// refusalReasonCode.
type paymentsResponse struct {
	PspReference       string            `json:"pspReference"`
	ResultCode         string            `json:"resultCode"`
	MerchantReference  string            `json:"merchantReference"`
	Amount             Amount            `json:"amount"`
	PaymentMethod      paymentMethodEcho `json:"paymentMethod,omitempty"`
	AdditionalData     map[string]string `json:"additionalData,omitempty"`
	RefusalReason      string            `json:"refusalReason,omitempty"`
	RefusalReasonCode  string            `json:"refusalReasonCode,omitempty"`
}

// paymentMethodEcho is the minimal shape Adyen echoes back on the response —
// { brand: "visa", type: "scheme" } for card, or { type: "googlepay" } etc.
type paymentMethodEcho struct {
	Brand string `json:"brand,omitempty"`
	Type  string `json:"type,omitempty"`
}

// errorResponse is Adyen's validation / auth error envelope. Distinct from
// paymentsResponse — the gateway returns HTTP 4xx with this body when the
// request itself is rejected before hitting the acquirer.
type errorResponse struct {
	Status       int    `json:"status"`
	ErrorCode    string `json:"errorCode"`
	Message      string `json:"message"`
	ErrorType    string `json:"errorType"`
	PspReference string `json:"pspReference"`
}

// webhookNotification is the top-level webhook envelope Adyen sends to
// merchants. Always exactly one `notificationItems` entry in our mock
// (Adyen batches in prod; we don't simulate that).
type webhookNotification struct {
	Live              string             `json:"live"`
	NotificationItems []notificationItem `json:"notificationItems"`
}

type notificationItem struct {
	NotificationRequestItem notificationRequestItem `json:"NotificationRequestItem"`
}

// notificationRequestItem is the rich webhook payload Adyen sends per
// settled transaction. The merchant-integration contract is:
//   - success: "true" | "false" (string, not bool)
//   - eventCode drives the consumer's state machine (AUTHORISATION,
//     REFUND, CANCELLATION, CAPTURE, CHARGEBACK, etc.)
//   - reason populated on refusal
//   - operations lists what the merchant can still do with this pspReference
type notificationRequestItem struct {
	AdditionalData      map[string]string `json:"additionalData,omitempty"`
	Amount              Amount            `json:"amount"`
	EventCode           string            `json:"eventCode"`
	EventDate           time.Time         `json:"eventDate"`
	MerchantAccountCode string            `json:"merchantAccountCode"`
	MerchantReference   string            `json:"merchantReference"`
	OriginalReference   string            `json:"originalReference,omitempty"`
	PaymentMethod       string            `json:"paymentMethod"`
	PspReference        string            `json:"pspReference"`
	Reason              string            `json:"reason,omitempty"`
	Success             string            `json:"success"`
	Operations          []string          `json:"operations,omitempty"`
}
