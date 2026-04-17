DROP INDEX IF EXISTS idx_request_logs_scenario_id;
DROP INDEX IF EXISTS idx_request_logs_merchant_order_id;
ALTER TABLE request_logs
    DROP COLUMN IF EXISTS scenario_id,
    DROP COLUMN IF EXISTS merchant_order_id;
