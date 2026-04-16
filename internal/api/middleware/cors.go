package middleware

import (
	"net/http"
	"strings"
)

// CORS returns a middleware that sets permissive CORS headers for the allowed
// origins. Credentials are allowed (required for the httpOnly session cookie).
// Empty allowedOrigins means allow the request's Origin header (for local mode).
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	originSet := make(map[string]bool, len(allowedOrigins))
	allowAll := false
	for _, o := range allowedOrigins {
		if o == "*" {
			allowAll = true
			continue
		}
		originSet[strings.TrimRight(o, "/")] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			allow := ""
			switch {
			case len(allowedOrigins) == 0:
				// No config: echo the request origin (local dev).
				allow = origin
			case allowAll:
				allow = origin
			case originSet[origin]:
				allow = origin
			}

			if allow != "" {
				w.Header().Set("Access-Control-Allow-Origin", allow)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Vary", "Origin")
			}

			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set(
					"Access-Control-Allow-Headers",
					"Content-Type, Authorization, X-Request-ID",
				)
				w.Header().Set("Access-Control-Max-Age", "3600")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
