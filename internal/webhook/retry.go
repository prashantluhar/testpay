package webhook

import (
	"context"
	"time"

	"github.com/prashantluhar/testpay/internal/store"
)

// DispatchAsync fires a webhook in a goroutine and updates the WebhookLog.
// delayMs: milliseconds to wait before dispatching (0 = immediate).
func DispatchAsync(
	ctx context.Context,
	d *Dispatcher,
	s store.Store,
	webhookLog *store.WebhookLog,
	delayMs int,
) {
	go func() {
		if delayMs > 0 {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}

		result, err := d.Dispatch(webhookLog.TargetURL, webhookLog.Payload)

		webhookLog.Attempts = result.Attempts
		for _, a := range result.AttemptLogs {
			webhookLog.AttemptLogs = append(webhookLog.AttemptLogs, store.AttemptLog{
				Status:      a.Status,
				DurationMs:  a.DurationMs,
				AttemptedAt: a.AttemptedAt.Format(time.RFC3339),
			})
		}

		if err == nil {
			now := time.Now()
			webhookLog.DeliveryStatus = "delivered"
			webhookLog.DeliveredAt = &now
		} else {
			webhookLog.DeliveryStatus = "failed"
		}

		s.UpdateWebhookLog(ctx, webhookLog)
	}()
}
