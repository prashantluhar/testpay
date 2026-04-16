package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger injects a contextualized zerolog logger (with trace_id) into the
// request context, then emits a single "request completed" log at the edge.
// Downstream code retrieves the logger via log.Ctx(r.Context()).
func Logger(env, service string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			traceID := TraceID(r.Context())
			ctxLogger := log.With().
				Str("trace_id", traceID).
				Str("env", env).
				Str("service", service).
				Logger()
			ctx := ctxLogger.WithContext(r.Context())

			rw := &statusWriter{ResponseWriter: w, status: 200}
			next.ServeHTTP(rw, r.WithContext(ctx))

			durNs := time.Since(start).Nanoseconds()
			if durNs <= 0 {
				// On some platforms (e.g. Windows) the monotonic clock
				// resolution can leave durNs at 0 for sub-millisecond
				// handlers. Report a small non-zero value so the field
				// is always present and meaningful.
				durNs = 1
			}
			durMs := float64(durNs) / 1e6
			zerolog.Ctx(ctx).Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("client_ip", r.RemoteAddr).
				Str("user_agent", r.UserAgent()).
				Int("status", rw.status).
				Float64("duration_ms", durMs).
				Msg("request completed")
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}
