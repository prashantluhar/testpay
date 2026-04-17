package adapters

import (
	"fmt"
	"sort"
	"strings"

	"github.com/prashantluhar/testpay/internal/adapters/adyen"
	"github.com/prashantluhar/testpay/internal/adapters/agnostic"
	"github.com/prashantluhar/testpay/internal/adapters/espay"
	"github.com/prashantluhar/testpay/internal/adapters/instamojo"
	"github.com/prashantluhar/testpay/internal/adapters/komoju"
	"github.com/prashantluhar/testpay/internal/adapters/mastercard"
	"github.com/prashantluhar/testpay/internal/adapters/paynamics"
	"github.com/prashantluhar/testpay/internal/adapters/razorpay"
	"github.com/prashantluhar/testpay/internal/adapters/stripe"
	"github.com/prashantluhar/testpay/internal/adapters/tappay"
	"github.com/prashantluhar/testpay/internal/adapters/tillpay"
)

type Registry struct {
	adapters map[string]Adapter
}

func NewRegistry() *Registry {
	r := &Registry{adapters: make(map[string]Adapter)}
	ag := agnostic.New()
	// Only gateways with full DTO-backed responses + webhook shapes are
	// registered. Stub adapters (epay, omise, payletter) were removed
	// because they returned generic shapes that misrepresented what a
	// real integration with those PSPs would look like.
	r.adapters["stripe"] = stripe.New()
	r.adapters["razorpay"] = razorpay.New()
	r.adapters["adyen"] = adyen.New()
	r.adapters["mastercard"] = mastercard.New()
	r.adapters["komoju"] = komoju.New()
	r.adapters["instamojo"] = instamojo.New()
	r.adapters["tillpay"] = tillpay.New()
	r.adapters["tappay"] = tappay.New()
	r.adapters["paynamics"] = paynamics.New()
	r.adapters["espay"] = espay.New()
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
