package adapters

import (
	"fmt"
	"sort"
	"strings"

	"github.com/prashantluhar/testpay/internal/adapters/adyen"
	"github.com/prashantluhar/testpay/internal/adapters/agnostic"
	"github.com/prashantluhar/testpay/internal/adapters/mastercard"
	"github.com/prashantluhar/testpay/internal/adapters/omise"
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
	r.adapters["adyen"] = adyen.New()
	r.adapters["omise"] = omise.New()
	r.adapters["mastercard"] = mastercard.New()
	r.adapters["agnostic"] = ag
	// URL-prefix alias: /v1/... routes to the agnostic adapter.
	r.adapters["v1"] = ag
	return r
}

// KnownGateways returns the canonical gateway names (not URL aliases).
// Used by the dashboard to render per-gateway webhook configuration.
func (r *Registry) KnownGateways() []string {
	seen := map[string]bool{}
	out := []string{}
	for k, a := range r.adapters {
		// Skip URL-only aliases that point to the same adapter as a canonical name.
		if a.Name() != k {
			continue
		}
		if !seen[k] {
			seen[k] = true
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
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
