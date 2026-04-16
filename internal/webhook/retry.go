package webhook

import (
	"context"
	"time"

	"github.com/prashantluhar/testpay/internal/store"
	"github.com/rs/zerolog"
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
		logger := zerolog.Ctx(ctx)

		logger.Info().
			Str("webhook_log_id", webhookLog.ID).
			Str("target_url", webhookLog.TargetURL).
			Int("delay_ms", delayMs).
			Msg("webhook dispatch scheduled")

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
			logger.Info().
				Str("webhook_log_id", webhookLog.ID).
				Str("target_url", webhookLog.TargetURL).
				Int("attempt_status", a.Status).
				Int("attempt_duration_ms", a.DurationMs).
				Msg("webhook attempt")
		}

		if err == nil {
			now := time.Now()
			webhookLog.DeliveryStatus = "delivered"
			webhookLog.DeliveredAt = &now
			logger.Info().
				Str("webhook_log_id", webhookLog.ID).
				Str("target_url", webhookLog.TargetURL).
				Int("attempts", result.Attempts).
				Int("status_code", result.StatusCode).
				Msg("webhook delivered")
		} else {
			webhookLog.DeliveryStatus = "failed"
			logger.Error().
				Err(err).
				Str("webhook_log_id", webhookLog.ID).
				Str("target_url", webhookLog.TargetURL).
				Int("attempts", result.Attempts).
				Msg("webhook failed")
		}

		if updErr := s.UpdateWebhookLog(ctx, webhookLog); updErr != nil {
			logger.Error().
				Err(updErr).
				Str("webhook_log_id", webhookLog.ID).
				Msg("failed to update webhook log")
		}
	}()
}
