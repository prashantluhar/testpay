// DTO shapes for Instamojo's (India) payment-request API.
// Derived from the production reference library (test-dto-/payments.go,
// structs: InstamojoInitRespPayload, InstamojoPaymentDetailsPayload,
// InstamojoErrorPayload, InstamojoRefundResponse, ...).
//
// Instamojo's API returns a top-level `{ success: bool, payment_request: {...} }`
// envelope. Failure responses flip `success` to false and ship a `message` —
// plus an optional `errors` map keyed on field name.
package instamojo

import "time"

// initResponse mirrors InstamojoInitRespPayload wrapped in the outer success
// envelope. payment_request is the nested object merchants key off of.
type initResponse struct {
	Success        bool           `json:"success"`
	PaymentRequest paymentRequest `json:"payment_request"`
}

// paymentRequest is the subset of InstamojoInitRespPayload this mock
// populates. `Amount` is string-encoded on Instamojo wire — a quirk merchants
// frequently trip over in integration.
type paymentRequest struct {
	ID                    string    `json:"id"`
	Phone                 string    `json:"phone,omitempty"`
	Email                 string    `json:"email,omitempty"`
	BuyerName             string    `json:"buyer_name,omitempty"`
	Amount                string    `json:"amount"`
	Purpose               string    `json:"purpose,omitempty"`
	Status                string    `json:"status"`
	SendSMS               bool      `json:"send_sms"`
	SendEmail             bool      `json:"send_email"`
	Longurl               string    `json:"longurl"`
	RedirectURL           string    `json:"redirect_url,omitempty"`
	Webhook               string    `json:"webhook,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
	ModifiedAt            time.Time `json:"modified_at"`
	AllowRepeatedPayments bool      `json:"allow_repeated_payments"`
}

// errorResponse mirrors InstamojoErrorPayload plus the field-level `errors`
// map Instamojo returns on validation failures.
type errorResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Errors  map[string][]string `json:"errors,omitempty"`
}

// webhookPayload is Instamojo's webhook shape — the fields of
// InstamojoPaymentDetailsPayload flattened into the form-encoded body
// Instamojo actually posts. We re-serialise as JSON for the mock.
type webhookPayload struct {
	PaymentID        string    `json:"payment_id"`
	PaymentRequestID string    `json:"payment_request_id"`
	Status           string    `json:"status"`
	Amount           string    `json:"amount"`
	Currency         string    `json:"currency"`
	Buyer            string    `json:"buyer"`
	BuyerName        string    `json:"buyer_name,omitempty"`
	BuyerPhone       string    `json:"buyer_phone,omitempty"`
	Purpose          string    `json:"purpose,omitempty"`
	Fees             string    `json:"fees"`
	InstrumentType   string    `json:"instrument_type,omitempty"`
	FailureReason    string    `json:"failure_reason,omitempty"`
	FailureMessage   string    `json:"failure_message,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}
