export interface Workspace {
  id: string;
  slug: string;
  api_key: string;
  webhook_urls: Record<string, string>;
  created_at: string;
}

export interface User {
  id: string;
  workspace_id: string;
  email: string;
  role: 'owner' | 'member';
  created_at: string;
}

export interface Step {
  event: string;
  outcome: string;
  code?: string;
}

export interface Scenario {
  id: string;
  workspace_id: string;
  name: string;
  description: string;
  gateway: 'stripe' | 'razorpay' | 'agnostic';
  steps: Step[];
  webhook_delay_ms: number;
  is_default: boolean;
  created_at: string;
}

export interface ScenarioRun {
  id: string;
  scenario_id: string;
  status: 'running' | 'completed' | 'failed';
  started_at: string;
  completed_at?: string;
}

export interface AttemptLog {
  status: number;
  duration_ms: number;
  response: string;
  attempted_at: string;
}

export interface RequestLog {
  id: string;
  workspace_id: string;
  scenario_run_id?: string;
  gateway: string;
  method: string;
  path: string;
  request_headers: Record<string, string>;
  request_body: Record<string, unknown>;
  response_headers: Record<string, string>;
  response_body: Record<string, unknown>;
  response_status: number;
  duration_ms: number;
  client_ip: string;
  created_at: string;
}

export interface WebhookLog {
  id: string;
  request_log_id: string;
  payload: Record<string, unknown>;
  target_url: string;
  delivery_status: 'pending' | 'delivered' | 'failed' | 'duplicate';
  attempts: number;
  attempt_logs: AttemptLog[];
  delivered_at?: string;
  created_at: string;
}

export interface AuthResponse {
  user: User | null;
  workspace: Workspace;
}

export type Mode = 'local' | 'hosted';

export const MODE: Mode =
  (process.env.NEXT_PUBLIC_TESTPAY_MODE as Mode) || 'local';
