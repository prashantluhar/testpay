export interface Workspace {
  ID: string;
  Slug: string;
  APIKey: string;
  CreatedAt: string;
}

export interface User {
  ID: string;
  WorkspaceID: string;
  Email: string;
  Role: 'owner' | 'member';
  CreatedAt: string;
}

export interface Step {
  event: string;
  outcome: string;
  code?: string;
}

export interface Scenario {
  ID: string;
  WorkspaceID: string;
  Name: string;
  Description: string;
  Gateway: 'stripe' | 'razorpay' | 'agnostic';
  Steps: Step[];
  WebhookDelayMs: number;
  IsDefault: boolean;
  CreatedAt: string;
}

export interface ScenarioRun {
  ID: string;
  ScenarioID: string;
  Status: 'running' | 'completed' | 'failed';
  StartedAt: string;
  CompletedAt?: string;
}

export interface AttemptLog {
  status: number;
  duration_ms: number;
  response: string;
  attempted_at: string;
}

export interface RequestLog {
  ID: string;
  WorkspaceID: string;
  ScenarioRunID?: string;
  Gateway: string;
  Method: string;
  Path: string;
  RequestHeaders: Record<string, string>;
  RequestBody: Record<string, unknown>;
  ResponseHeaders: Record<string, string>;
  ResponseBody: Record<string, unknown>;
  ResponseStatus: number;
  DurationMs: number;
  ClientIP: string;
  CreatedAt: string;
}

export interface WebhookLog {
  ID: string;
  RequestLogID: string;
  Payload: Record<string, unknown>;
  TargetURL: string;
  DeliveryStatus: 'pending' | 'delivered' | 'failed' | 'duplicate';
  Attempts: number;
  AttemptLogs: AttemptLog[];
  DeliveredAt?: string;
  CreatedAt: string;
}

export interface AuthResponse {
  user: User | null;
  workspace: Workspace;
}

export type Mode = 'local' | 'hosted';

export const MODE: Mode =
  (process.env.NEXT_PUBLIC_TESTPAY_MODE as Mode) || 'local';
