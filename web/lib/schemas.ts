import { z } from 'zod';
import { FAILURE_MODES, EVENT_TYPES } from './failure-modes';

export const loginSchema = z.object({
  email: z.string().email(),
  password: z.string().min(8, 'at least 8 characters'),
});
export type LoginInput = z.infer<typeof loginSchema>;

export const signupSchema = loginSchema;
export type SignupInput = LoginInput;

const outcomes = FAILURE_MODES.map((m) => m.value) as [string, ...string[]];
const events = EVENT_TYPES as unknown as [string, ...string[]];

export const scenarioStepSchema = z.object({
  event: z.enum(events),
  outcome: z.enum(outcomes),
  code: z.string().optional(),
});

export const scenarioSchema = z.object({
  name: z.string().min(1),
  description: z.string(),
  gateway: z.enum(['stripe', 'razorpay', 'agnostic']),
  steps: z.array(scenarioStepSchema).min(1),
  webhook_delay_ms: z.number().int().min(0),
  is_default: z.boolean(),
});
export type ScenarioInput = z.infer<typeof scenarioSchema>;
