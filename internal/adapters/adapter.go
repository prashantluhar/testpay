package adapters

import (
	"net/http"

	"github.com/prashantluhar/testpay/internal/engine"
)

// Adapter translates between a gateway's wire format and TestPay internals.
type Adapter interface {
	Name() string
	// BuildResponse turns an engine Result into an HTTP response body + status.
	BuildResponse(result *engine.Result, originalBody []byte) (statusCode int, body []byte, headers map[string]string)
	// BuildWebhookPayload returns the webhook payload for a given result.
	BuildWebhookPayload(result *engine.Result, chargeID string, amount int64, currency string) map[string]any
}

// GatewayRequest is the parsed, gateway-agnostic form of an incoming mock request.
type GatewayRequest struct {
	ChargeID string
	Amount   int64
	Currency string
	RawBody  []byte
	Headers  http.Header
}
