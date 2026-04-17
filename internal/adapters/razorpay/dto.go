// DTO shapes for Razorpay's /payments response and webhook notifications.
// Derived from the production reference library (test-dto-/razorpay.go) and
// Razorpay's public API docs. Kept scoped to the fields this mock emits;
// extend when a new failure mode needs more surface.
package razorpay

// Razorpay payment status vocabulary. Razorpay's real lifecycle is
// created → authorized → captured, with failed as a terminal error state.
// The reference DTO only exposes created/attempted/paid — we add the
// full set needed for this mock.
const (
	statusCreated    = "created"
	statusAttempted  = "attempted"
	statusAuthorized = "authorized"
	statusCaptured   = "captured"
	statusFailed     = "failed"
)

// Razorpay webhook event names. Only the three payment lifecycle events
// are modeled — refund/dispute events are out of scope for this mock.
const (
	eventPaymentCaptured   = "payment.captured"
	eventPaymentAuthorized = "payment.authorized"
	eventPaymentFailed     = "payment.failed"
)

// paymentEntity mirrors Razorpay's payment object — both the POST /payments
// response body and the `payload.payment.entity` node inside a webhook.
// Error fields are populated only when status == "failed".
type paymentEntity struct {
	ID               string            `json:"id"`
	Entity           string            `json:"entity"`
	Amount           int64             `json:"amount"`
	Currency         string            `json:"currency"`
	Status           string            `json:"status"`
	Method           string            `json:"method,omitempty"`
	Notes            map[string]any    `json:"notes"`
	ErrorCode        string            `json:"error_code,omitempty"`
	ErrorDescription string            `json:"error_description,omitempty"`
	ErrorSource      string            `json:"error_source,omitempty"`
	ErrorStep        string            `json:"error_step,omitempty"`
	ErrorReason      string            `json:"error_reason,omitempty"`
}

// errorBody is Razorpay's error envelope for HTTP 4xx responses — distinct
// from an in-band `status: "failed"` payment. The gateway returns this when
// it rejects the request itself (auth, validation, rate limit).
// See https://razorpay.com/docs/api/errors/
type errorBody struct {
	Code        string            `json:"code"`
	Description string            `json:"description"`
	Source      string            `json:"source,omitempty"`
	Step        string            `json:"step,omitempty"`
	Reason      string            `json:"reason,omitempty"`
	Metadata    map[string]any    `json:"metadata,omitempty"`
	Field       string            `json:"field,omitempty"`
}

// errorResponse wraps errorBody in the top-level `{ "error": {...} }` shape
// Razorpay sends on 4xx.
type errorResponse struct {
	Error errorBody `json:"error"`
}

// webhookEnvelope is the top-level shape Razorpay POSTs to merchant webhook
// URLs. `contains` lists which entities are nested in `payload` — for
// payment events that's just ["payment"].
type webhookEnvelope struct {
	Entity    string         `json:"entity"`
	Event     string         `json:"event"`
	Contains  []string       `json:"contains"`
	Payload   webhookPayload `json:"payload"`
	CreatedAt int64          `json:"created_at"`
}

type webhookPayload struct {
	Payment webhookPaymentWrapper `json:"payment"`
}

// webhookPaymentWrapper wraps the payment entity under an extra `entity`
// key — Razorpay's convention so the same envelope can carry refunds,
// orders, etc. alongside in future events.
type webhookPaymentWrapper struct {
	Entity paymentEntity `json:"entity"`
}
