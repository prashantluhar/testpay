/**
 * MUST stay in sync with internal/engine/modes.go.
 * Update this list when modes.go changes.
 */
export const FAILURE_MODES = [
  { value: 'success', label: 'Success', group: 'Generic' },
  { value: 'bank_decline_hard', label: 'Bank — Decline (hard)', group: 'Bank' },
  { value: 'bank_decline_soft', label: 'Bank — Decline (soft)', group: 'Bank' },
  { value: 'bank_server_down', label: 'Bank — Server down', group: 'Bank' },
  { value: 'bank_timeout', label: 'Bank — Timeout', group: 'Bank' },
  { value: 'bank_invalid_cvv', label: 'Bank — Invalid CVV', group: 'Bank' },
  { value: 'bank_do_not_honour', label: 'Bank — Do not honour', group: 'Bank' },
  { value: 'pg_server_error', label: 'PG — 500/503', group: 'PG' },
  { value: 'pg_timeout', label: 'PG — Timeout', group: 'PG' },
  { value: 'pg_rate_limited', label: 'PG — Rate limited (429)', group: 'PG' },
  { value: 'pg_maintenance', label: 'PG — Maintenance', group: 'PG' },
  { value: 'network_error', label: 'Network error', group: 'PG' },
  { value: 'webhook_missing', label: 'Webhook — Missing', group: 'Webhook' },
  { value: 'webhook_delayed', label: 'Webhook — Delayed', group: 'Webhook' },
  { value: 'webhook_duplicate', label: 'Webhook — Duplicate', group: 'Webhook' },
  { value: 'webhook_out_of_order', label: 'Webhook — Out of order', group: 'Webhook' },
  { value: 'webhook_malformed', label: 'Webhook — Malformed', group: 'Webhook' },
  { value: 'redirect_success', label: 'Redirect — Success', group: 'Redirect' },
  { value: 'redirect_abandoned', label: 'Redirect — Abandoned', group: 'Redirect' },
  { value: 'redirect_timeout', label: 'Redirect — Timeout', group: 'Redirect' },
  { value: 'redirect_failed', label: 'Redirect — Failed', group: 'Redirect' },
  { value: 'double_charge', label: 'Charge — Double', group: 'Charge' },
  { value: 'amount_mismatch', label: 'Charge — Amount mismatch', group: 'Charge' },
  { value: 'partial_success', label: 'Charge — Partial success', group: 'Charge' },
  { value: 'pending_then_failed', label: 'Async — pending → failed', group: 'Async' },
  { value: 'pending_then_success', label: 'Async — pending → success', group: 'Async' },
  { value: 'failed_then_success', label: 'Async — failed → success', group: 'Async' },
  { value: 'success_then_reversed', label: 'Async — success → reversed', group: 'Async' },
] as const;

export type FailureMode = (typeof FAILURE_MODES)[number]['value'];
export const EVENT_TYPES = ['charge', 'refund', 'capture'] as const;
export type EventType = (typeof EVENT_TYPES)[number];
