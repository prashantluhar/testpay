-- Add per-session call counter so live mock traffic can walk multi-step
-- scenarios: each call advances the counter, mock.go uses
--   stepIndex = call_index % len(steps)
-- so scenarios with multiple steps play out over successive SDK calls.
ALTER TABLE sessions ADD COLUMN call_index INTEGER NOT NULL DEFAULT 0;
