package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const RequestIDHeader = "X-Request-ID"

type ctxKey string

const traceIDCtxKey ctxKey = "trace_id"

// RequestID extracts or generates a trace ID and stores it in both
// the response header and the request context.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}
		w.Header().Set(RequestIDHeader, id)
		ctx := context.WithValue(r.Context(), traceIDCtxKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// TraceID returns the trace ID stored in ctx, or "" if none.
func TraceID(ctx context.Context) string {
	if v, ok := ctx.Value(traceIDCtxKey).(string); ok {
		return v
	}
	return ""
}
