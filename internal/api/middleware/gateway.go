package middleware

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const GatewayKey contextKey = "gateway"

// GatewayResolver infers the gateway name from the URL prefix and stores it in context.
func GatewayResolver(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/"), "/", 2)
		gateway := "agnostic"
		if len(parts) > 0 && parts[0] != "api" {
			gateway = parts[0]
		}
		ctx := context.WithValue(r.Context(), GatewayKey, gateway)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetGateway(r *http.Request) string {
	if g, ok := r.Context().Value(GatewayKey).(string); ok {
		return g
	}
	return "agnostic"
}
