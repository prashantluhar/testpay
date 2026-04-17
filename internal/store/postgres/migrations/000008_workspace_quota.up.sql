-- Per-workspace daily request cap. NULL means unlimited (default for
-- authenticated users). The seeded "local" workspace (used by anonymous
-- mock traffic against the hosted demo) gets a hard cap so a runaway
-- SDK integration test can't drain the free-tier's Neon quota or
-- Render compute minutes.
ALTER TABLE workspaces
    ADD COLUMN max_daily_requests INTEGER;

-- Cap the anonymous demo workspace at 200 requests/day. Real workspaces
-- created via signup stay NULL (unlimited) until we decide otherwise.
UPDATE workspaces
SET max_daily_requests = 200
WHERE slug = 'local';
