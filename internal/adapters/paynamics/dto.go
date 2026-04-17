// DTO shapes for Paynamics (PH / SEA) payment API surface.
// Derived from the production reference library (test-dto-/payments.go,
// structs: PaynamicsInitRespPayload, PaynamicsQueryRespPayload,
// PaynamicsRefundRespPayload, PaynamicsWebhookPayload, ...).
//
// Paynamics uses an MD5-signed JSON envelope with a response_code /
// response_advise / response_message triple — "GR001" is the canonical
// success code; "GR051" and "GR0xx" cover the failure taxonomy. The real
// gateway historically emitted form-encoded responses; modern integrations
// use JSON. The mock emits JSON exclusively with a fixed signature
// placeholder.
package paynamics

// initRespPayload mirrors PaynamicsInitRespPayload — returned when the
// merchant creates / initializes a transaction. redirect_url drives the
// hosted-payment-page flow; our mock stubs it with a deterministic path so
// tests can assert on shape without relying on randomness.
type initRespPayload struct {
	ResponseCode    string `json:"response_code"`
	ResponseAdvise  string `json:"response_advise"`
	ResponseMessage string `json:"response_message"`
	Signature       string `json:"signature"`
	ResponseID      string `json:"response_id"`
	MerchantID      string `json:"merchant_id"`
	RequestID       string `json:"request_id"`
	RedirectURL     string `json:"redirect_url,omitempty"`
	Timestamp       string `json:"timestamp"`
	Currency        string `json:"currency,omitempty"`
	TotalAmount     string `json:"total_amount,omitempty"`
}

// errorResponse is Paynamics's envelope when the transaction fails at init
// time — same top-level fields as a success minus the redirect_url, plus a
// descriptive response_advise that merchants show to end users.
type errorResponse struct {
	ResponseCode    string `json:"response_code"`
	ResponseAdvise  string `json:"response_advise"`
	ResponseMessage string `json:"response_message"`
	Signature       string `json:"signature"`
	ResponseID      string `json:"response_id"`
	MerchantID      string `json:"merchant_id"`
	RequestID       string `json:"request_id"`
	Timestamp       string `json:"timestamp"`
}

// webhookPayload mirrors PaynamicsWebhookPayload — Paynamics posts this to
// the merchant's notification_url once the transaction reaches a terminal
// state. The shape overlaps with initRespPayload (same response_code
// vocabulary, same signature placeholder) but carries merchant_id /
// request_id so merchants can correlate.
type webhookPayload struct {
	ResponseCode    string `json:"response_code"`
	ResponseAdvise  string `json:"response_advise"`
	ResponseMessage string `json:"response_message"`
	Signature       string `json:"signature"`
	ResponseID      string `json:"response_id"`
	MerchantID      string `json:"merchant_id"`
	RequestID       string `json:"request_id"`
	RedirectURL     string `json:"redirect_url,omitempty"`
	Timestamp       string `json:"timestamp"`
	TotalAmount     string `json:"total_amount,omitempty"`
	Currency        string `json:"currency,omitempty"`
}

// queryRespPayload mirrors PaynamicsQueryRespPayload — Paynamics's status
// endpoint shape. Included for completeness (the mock doesn't expose a query
// endpoint yet) so the DTO surface parallels the reference library.
type queryRespPayload struct {
	MerchantID                string `json:"Merchantid"`
	RequestID                 string `json:"request_id"`
	ResponseID                string `json:"response_id"`
	ResponseCode              string `json:"response_code"`
	ResponseAdvise            string `json:"response_advise"`
	ResponseMessage           string `json:"response_message"`
	Timestamp                 string `json:"timestamp"`
	Signature                 string `json:"signature"`
	ProcessorResponseAuthCode string `json:"processor_response_authcode,omitempty"`
	PayReference              string `json:"pay_reference,omitempty"`
}
