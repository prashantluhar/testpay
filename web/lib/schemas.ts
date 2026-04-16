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
  Name: z.string().min(1),
  Description: z.string().default(''),
  Gateway: z.enum(['stripe', 'razorpay', 'agnostic']),
  Steps: z.array(scenarioStepSchema).min(1),
  WebhookDelayMs: z.number().int().min(0).default(0),
  IsDefault: z.boolean().default(false),
});
export type ScenarioInput = z.infer<typeof scenarioSchema>;
