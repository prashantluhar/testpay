package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// maxAttemptBodyBytes caps how much of each attempt's response body we capture.
// Deliberately tight because attempt bodies are persisted per-delivery in JSONB
// and the hosted free tier has a 0.5 GB Neon cap; HTML error pages from
// misconfigured endpoints can balloon this column fast without the limit.
const maxAttemptBodyBytes = 8 * 1024

type DispatchResult struct {
	StatusCode  int
	Attempts    int
	AttemptLogs []AttemptEntry
}

type AttemptEntry struct {
	Status      int
	DurationMs  int
	AttemptedAt time.Time
	Response    string // response body (capped at maxAttemptBodyBytes)
	Error       string // transport-level error; empty when an HTTP response was received
}

type Dispatcher struct {
	maxAttempts int
	baseDelay   time.Duration
	client      *http.Client
}

func NewDispatcher(maxAttempts int, baseDelay time.Duration) *Dispatcher {
	return &Dispatcher{
		maxAttempts: maxAttempts,
		baseDelay:   baseDelay,
		client:      &http.Client{Timeout: 10 * time.Second},
	}
}

func (d *Dispatcher) Dispatch(targetURL string, payload map[string]any) (*DispatchResult, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshalling webhook payload: %w", err)
	}

	result := &DispatchResult{}
	var lastStatus int

	for i := 0; i < d.maxAttempts; i++ {
		if i > 0 {
			time.Sleep(d.baseDelay * time.Duration(1<<uint(i-1))) // exponential backoff
		}

		start := time.Now()
		req, _ := http.NewRequest("POST", targetURL, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "TestPay-Webhook/1.0")

		resp, err := d.client.Do(req)
		durationMs := int(time.Since(start).Milliseconds())

		entry := AttemptEntry{DurationMs: durationMs, AttemptedAt: time.Now()}
		if err != nil {
			entry.Status = 0
			entry.Error = err.Error()
		} else {
			entry.Status = resp.StatusCode
			// Cap the body read so a massive error page doesn't bloat the JSONB
			// column. io.LimitReader + io.ReadAll combines cleanly.
			bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, maxAttemptBodyBytes))
			entry.Response = string(bodyBytes)
			resp.Body.Close()
			lastStatus = resp.StatusCode
		}

		result.AttemptLogs = append(result.AttemptLogs, entry)
		result.Attempts = i + 1

		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			result.StatusCode = resp.StatusCode
			return result, nil
		}
	}

	result.StatusCode = lastStatus
	return result, fmt.Errorf("webhook delivery failed after %d attempts", d.maxAttempts)
}
