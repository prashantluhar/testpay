package adapters

import (
	"fmt"
	"strings"

	"github.com/prashantluhar/testpay/internal/adapters/agnostic"
	"github.com/prashantluhar/testpay/internal/adapters/razorpay"
	"github.com/prashantluhar/testpay/internal/adapters/stripe"
)

type Registry struct {
	adapters map[string]Adapter
}

func NewRegistry() *Registry {
	r := &Registry{adapters: make(map[string]Adapter)}
	ag := agnostic.New()
	r.adapters["stripe"] = stripe.New()
	r.adapters["razorpay"] = razorpay.New()
	r.adapters["agnostic"] = ag
	// URL-prefix alias: /v1/... routes to the agnostic adapter.
	r.adapters["v1"] = ag
	return r
}

// Resolve returns the adapter matching the URL prefix (e.g. /stripe/v1/... → stripe).
func (r *Registry) Resolve(urlPath string) (Adapter, error) {
	parts := strings.SplitN(strings.TrimPrefix(urlPath, "/"), "/", 2)
	if len(parts) == 0 {
		return nil, fmt.Errorf("cannot resolve gateway from path: %s", urlPath)
	}
	a, ok := r.adapters[parts[0]]
	if !ok {
		return nil, fmt.Errorf("unknown gateway: %s", parts[0])
	}
	return a, nil
}

// GatewayFromPath returns just the gateway name string.
func GatewayFromPath(urlPath string) string {
	parts := strings.SplitN(strings.TrimPrefix(urlPath, "/"), "/", 2)
	if len(parts) == 0 {
		return "agnostic"
	}
	return parts[0]
}
