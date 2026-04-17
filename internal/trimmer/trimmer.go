// Package trimmer runs a background goroutine that periodically deletes
// request_logs (and cascading webhook_logs) older than the configured
// retention window. Keeps the free-tier Neon 0.5 GB storage cap from
// filling up, and bounds what pilots see in the Logs tab to the last
// N days of activity.
package trimmer

import (
	"context"
	"time"

	"github.com/prashantluhar/testpay/internal/store"
	"github.com/rs/zerolog/log"
)

// Run starts a trimmer loop on the given store. Blocks until ctx is cancelled.
// Call in a goroutine.
//
// runEvery: how often to sweep. 1 hour is a good default — trimming more
//
//	often wastes cycles, less often lets storage drift longer.
//
// retention: rows older than now()-retention are deleted. 10 days is the
//
//	default in cli/start.go.
//
// A single initial sweep fires ~30s after Run() so a freshly-booted server
// doesn't wait a full hour before its first cleanup.
func Run(ctx context.Context, s store.Store, runEvery, retention time.Duration) {
	log.Info().
		Dur("run_every", runEvery).
		Dur("retention", retention).
		Msg("trimmer started")

	firstTick := time.NewTimer(30 * time.Second)
	defer firstTick.Stop()

	ticker := time.NewTicker(runEvery)
	defer ticker.Stop()

	sweep := func() {
		cutoff := time.Now().Add(-retention)
		start := time.Now()
		deleted, err := s.TrimOldLogs(ctx, cutoff)
		if err != nil {
			log.Error().Err(err).Time("cutoff", cutoff).Msg("trimmer sweep failed")
			return
		}
		log.Info().
			Int64("deleted", deleted).
			Time("cutoff", cutoff).
			Dur("duration", time.Since(start)).
			Msg("trimmer sweep complete")
	}

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("trimmer stopped")
			return
		case <-firstTick.C:
			sweep()
		case <-ticker.C:
			sweep()
		}
	}
}
