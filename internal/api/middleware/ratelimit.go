package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// bucket is a token-bucket entry keyed by client IP.
type bucket struct {
	tokens     float64
	lastRefill time.Time
}

// RateLimiter enforces per-IP and global token-bucket caps. In-memory only;
// state is lost on restart, which is fine for free-tier abuse prevention
// (the server also sleeps, so long-term accumulation doesn't matter).
type RateLimiter struct {
	mu          sync.Mutex
	buckets     map[string]*bucket
	ratePerSec  float64
	burst       float64
	globalBkt   bucket
	globalRate  float64
	globalBurst float64
}

// NewRateLimiter returns a limiter that allows roughly perMinute requests per
// IP with a burst of burstSize, and caps all clients combined at globalPerMin.
// Either value <=0 disables that tier.
func NewRateLimiter(perMinute, burstSize, globalPerMin int) *RateLimiter {
	rl := &RateLimiter{
		buckets:    map[string]*bucket{},
		ratePerSec: float64(perMinute) / 60.0,
		burst:      float64(burstSize),
	}
	if globalPerMin > 0 {
		rl.globalRate = float64(globalPerMin) / 60.0
		rl.globalBurst = float64(globalPerMin)
		rl.globalBkt = bucket{tokens: rl.globalBurst, lastRefill: time.Now()}
	}
	return rl
}

func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()

	// Global bucket first — if the instance is at capacity, reject regardless of IP.
	if rl.globalRate > 0 {
		elapsed := now.Sub(rl.globalBkt.lastRefill).Seconds()
		rl.globalBkt.tokens += elapsed * rl.globalRate
		if rl.globalBkt.tokens > rl.globalBurst {
			rl.globalBkt.tokens = rl.globalBurst
		}
		rl.globalBkt.lastRefill = now
		if rl.globalBkt.tokens < 1 {
			return false
		}
	}

	if rl.ratePerSec > 0 {
		b, ok := rl.buckets[ip]
		if !ok {
			b = &bucket{tokens: rl.burst, lastRefill: now}
			rl.buckets[ip] = b
		}
		elapsed := now.Sub(b.lastRefill).Seconds()
		b.tokens += elapsed * rl.ratePerSec
		if b.tokens > rl.burst {
			b.tokens = rl.burst
		}
		b.lastRefill = now
		if b.tokens < 1 {
			return false
		}
		b.tokens--
	}
	if rl.globalRate > 0 {
		rl.globalBkt.tokens--
	}
	return true
}

// Middleware returns the HTTP middleware. Requests over the limit get 429.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.allow(clientIP(r)) {
			w.Header().Set("Retry-After", "60")
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// clientIP prefers X-Forwarded-For (Render terminates TLS at a proxy so the
// direct RemoteAddr is always the proxy IP). Falls back to RemoteAddr.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.Index(xff, ","); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil || host == "" {
		return r.RemoteAddr
	}
	return host
}
